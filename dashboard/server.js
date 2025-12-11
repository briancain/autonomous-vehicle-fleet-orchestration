const http = require('http');
const https = require('https');
const fs = require('fs');
const path = require('path');
const url = require('url');

const port = 3000;

const server = http.createServer((req, res) => {
    const parsedUrl = url.parse(req.url, true);
    const pathname = parsedUrl.pathname;

    // Health check endpoint
    if (pathname === '/health') {
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ status: 'healthy' }));
        return;
    }

    // Proxy API requests to avoid CORS
    if (pathname.startsWith('/api/fleet/')) {
        const fleetPath = pathname.replace('/api/fleet', '');
        const fleetServiceUrl = process.env.FLEET_SERVICE_URL || 'http://localhost:8080';
        proxyRequest(req, res, `${fleetServiceUrl}${fleetPath}`);
        return;
    }

    if (pathname.startsWith('/api/jobs/')) {
        const jobPath = pathname.replace('/api/jobs', '');
        const jobServiceUrl = process.env.JOB_SERVICE_URL || 'http://localhost:8081';
        proxyRequest(req, res, `${jobServiceUrl}${jobPath}`);
        return;
    }

    // Serve static files
    let filePath = '.' + pathname;
    if (filePath === './') {
        filePath = './index.html';
    }

    const extname = String(path.extname(filePath)).toLowerCase();
    const mimeTypes = {
        '.html': 'text/html',
        '.js': 'text/javascript',
        '.css': 'text/css',
        '.json': 'application/json',
        '.png': 'image/png',
        '.jpg': 'image/jpg',
        '.gif': 'image/gif',
        '.svg': 'image/svg+xml',
        '.wav': 'audio/wav',
        '.mp4': 'video/mp4',
        '.woff': 'application/font-woff',
        '.ttf': 'application/font-ttf',
        '.eot': 'application/vnd.ms-fontobject',
        '.otf': 'application/font-otf',
        '.wasm': 'application/wasm'
    };

    const contentType = mimeTypes[extname] || 'application/octet-stream';

    fs.readFile(filePath, (error, content) => {
        if (error) {
            if (error.code === 'ENOENT') {
                res.writeHead(404, { 'Content-Type': 'text/html' });
                res.end('<h1>404 Not Found</h1>', 'utf-8');
            } else {
                res.writeHead(500);
                res.end(`Server Error: ${error.code}`, 'utf-8');
            }
        } else {
            res.writeHead(200, { 'Content-Type': contentType });
            res.end(content, 'utf-8');
        }
    });
});

function proxyRequest(req, res, targetUrl) {
    const options = {
        method: req.method,
        headers: req.headers
    };

    const isHttps = targetUrl.startsWith('https:');
    const client = isHttps ? https : http;

    const proxyReq = client.request(targetUrl, options, (proxyRes) => {
        res.writeHead(proxyRes.statusCode, proxyRes.headers);
        proxyRes.pipe(res);
    });

    proxyReq.on('error', (err) => {
        console.error('Proxy error:', err);
        res.writeHead(500);
        res.end('Proxy Error');
    });

    req.pipe(proxyReq);
}

server.listen(port, () => {
    console.log(`Dashboard server running at http://localhost:${port}/`);
});
