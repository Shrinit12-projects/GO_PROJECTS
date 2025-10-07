#!/bin/bash

# E-Learning Analytics Platform Setup Script
set -e

echo "Setting up E-Learning Analytics Platform..."

# Navigate to project directory
cd "$(dirname "$0")"

# Ensure .env file exists for auth service
if [ ! -f "./services/auth/.env" ]; then
    echo "Creating .env file for auth service..."
    cat > ./services/auth/.env << EOF
PORT=8080
DB_URI=mongodb://mongodb:27017/elearn
REDIS_URL=redis:6379
REDIS_PASSWORD=redispassword
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_ACCESS_TTL_SECONDS=900
JWT_REFRESH_TTL_SECONDS=604800
READ_TIMEOUT_SECONDS=15
WRITE_TIMEOUT_SECONDS=15
IDLE_TIMEOUT_SECONDS=60
EOF
    echo ".env file created"
else
    echo ".env file already exists"
fi

# Build and run docker compose
echo "Building and starting services..."
docker-compose up --build

echo "Setup complete!"