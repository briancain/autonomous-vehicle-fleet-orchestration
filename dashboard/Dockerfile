# Use AWS ECR Public Node.js base image
FROM public.ecr.aws/docker/library/node:18-alpine

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./
COPY tsconfig.json ./

# Install production dependencies only
RUN npm ci --production

# Copy source code and pre-built files
COPY . .

# Create non-root user
RUN adduser -D -s /bin/sh dashboard

# Change ownership to non-root user
RUN chown -R dashboard:dashboard /app

# Switch to non-root user
USER dashboard

# Expose port
EXPOSE 3000

# Run the server
CMD ["node", "server.js"]
