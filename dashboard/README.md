# Fleet Dashboard

Live dashboard for monitoring autonomous vehicle fleet operations.

## Features

- **Real-time vehicle tracking** on OpenStreetMap
- **Job queue monitoring** with status updates
- **Fleet metrics** including battery levels and availability
- **Responsive design** for desktop and mobile

## Quick Start

```bash
# Install dependencies and build
make dev

# Open dashboard in browser
make open
```

## Prerequisites

Ensure the following services are running:
- Fleet Service on `http://localhost:8080`
- Job Service on `http://localhost:8081`

## Development

```bash
# Watch TypeScript files for changes
make watch

# Build once
make build

# Clean build artifacts
make clean
```

## Deployment

For S3 deployment, upload these files:
- `index.html`
- `style.css`
- `dist/app.js`

Configure S3 bucket for static website hosting and ensure CORS is enabled on your backend services.

## API Endpoints Used

- `GET /vehicles` - Fleet service vehicle data
- `GET /jobs` - Job service job data

Updates every 3 seconds automatically.
