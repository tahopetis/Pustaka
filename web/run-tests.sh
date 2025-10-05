#!/bin/bash

# Playwright test runner script for Pustaka graph functionality
# This script sets up the environment and runs the tests

set -e

echo "🎭 Pustaka Graph Functionality Tests"
echo "=================================="

# Check if docker services are running
echo "📋 Checking if Docker services are running..."
if ! docker ps | grep -q "pustaka"; then
    echo "⚠️  Docker services not running. Starting services..."
    cd /home/syam/dev/pustaka
    docker compose up -d
    echo "⏳ Waiting for services to be ready..."
    sleep 30
else
    echo "✅ Docker services are running"
fi

# Check if frontend is accessible
echo "🌐 Checking if frontend is accessible..."
if curl -s http://localhost:3000 > /dev/null; then
    echo "✅ Frontend is accessible at http://localhost:3000"
else
    echo "❌ Frontend is not accessible. Please check docker services."
    exit 1
fi

# Check if API is accessible
echo "🔌 Checking if API is accessible..."
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ API is accessible at http://localhost:8080"
else
    echo "❌ API is not accessible. Please check docker services."
    exit 1
fi

# Navigate to web directory
cd /home/syam/dev/pustaka/web

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "📦 Installing dependencies..."
    npm install
fi

# Install Playwright browsers if needed
if ! npx playwright --version > /dev/null 2>&1; then
    echo "🎭 Installing Playwright..."
    npx playwright install
fi

echo "🧪 Running Playwright tests..."

# Run tests with different options based on arguments
if [ "$1" = "headed" ]; then
    echo "🖥️  Running tests in headed mode..."
    npx playwright test --headed
elif [ "$1" = "debug" ]; then
    echo "🐛 Running tests in debug mode..."
    npx playwright test --debug
else
    echo "🔬 Running tests in headless mode..."
    npx playwright test
fi

echo "📊 Test results available in: playwright-report/"
echo "🌐 To view results: npx playwright show-report"

echo "✅ Test execution completed!"