#!/bin/bash

# Script to run the delivery service locally

cd "$(dirname "$0")/../src"

# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=delivery
export VERSION=$(cat version.json | grep -o '"version": "[^"]*"' | cut -d'"' -f4)
export LOG_LEVEL=debug

# Build and run
echo "Building and running delivery service..."
go run main.go
