#!/bin/bash

# Build script for React frontend

set -e

echo "Building React frontend..."

# Navigate to React app directory
cd web/react-app

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo "Installing npm dependencies..."
    npm install
fi

# Build the React app
echo "Building React app for production..."
npm run build

echo "React build completed successfully!"
echo "Built files are in web/react-app/build/"

# Go back to project root
cd ../..

echo "Frontend build process completed."