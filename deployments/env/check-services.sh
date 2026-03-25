#!/bin/bash

echo "Checking if docker-compose is available..."
if ! command -v docker-compose &> /dev/null; then
    echo "Error: docker-compose is not installed"
    exit 1
fi

echo "Validating docker-compose configuration..."
docker-compose -f deployments/env/docker-compose.yml config > /dev/null
if [ $? -eq 0 ]; then
    echo "✓ Docker Compose configuration is valid"
else
    echo "✗ Docker Compose configuration has errors"
    exit 1
fi

echo "Development environment is ready."
echo ""
echo "Commands to manage the environment:"
echo "  Start services: docker-compose -f deployments/env/docker-compose.yml up"
echo "  Stop services:  docker-compose -f deployments/env/docker-compose.yml down"
echo "  View logs:      docker-compose -f deployments/env/docker-compose.yml logs -f"
echo ""
echo "Or using make commands:"
echo "  Start services: make dev-env-up"
echo "  Stop services:  make dev-env-down"
echo "  View logs:      make dev-env-logs"