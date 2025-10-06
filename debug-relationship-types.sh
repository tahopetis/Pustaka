#!/bin/bash

echo "=== Debugging Relationship Types Issue ==="

# Check if Docker services are running
echo "1. Checking Docker services..."
docker-compose ps

# Start Docker services if not running
if ! docker-compose ps | grep -q "Up"; then
    echo "Starting Docker services..."
    make docker-up
    sleep 10
fi

# Check if API is running
echo "2. Checking API health..."
curl -s http://localhost:8080/health || echo "API not responding"

# Check if frontend is running
echo "3. Checking frontend..."
curl -s http://localhost:3000 | head -n 1 || echo "Frontend not responding"

# Start API if not running
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "Starting API server..."
    cd /home/syam/dev/Pustaka
    make run &
    API_PID=$!
    sleep 5
fi

# Start frontend if not running
if ! curl -s http://localhost:3000 > /dev/null; then
    echo "Starting frontend..."
    cd /home/syam/dev/Pustaka/web
    npm run dev &
    FRONTEND_PID=$!
    sleep 10
fi

# Wait for services to be ready
echo "4. Waiting for services to be ready..."
sleep 10

# Check API endpoints directly
echo "5. Testing API endpoints..."

# Create admin user if needed
echo "Creating admin user..."
curl -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"Admin@123"}' \
    -c /tmp/cookies.txt || echo "Login failed"

# Get auth token
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"Admin@123"}' \
    -s | jq -r '.access_token')

echo "Token: $TOKEN"

# Test relationship types API
echo "Testing relationship types API..."
curl -X GET http://localhost:8080/api/v1/relationship-types \
    -H "Authorization: Bearer $TOKEN" \
    -s | jq '.' || echo "API call failed"

# Test relationship types stats API
echo "Testing relationship types stats API..."
curl -X GET http://localhost:8080/api/v1/relationship-types/stats \
    -H "Authorization: Bearer $TOKEN" \
    -s | jq '.' || echo "Stats API call failed"

# Check database for relationship types
echo "6. Checking database..."
docker exec pustaka-postgres psql -U pustaka -d pustaka -c "SELECT COUNT(*) as total_types FROM relationship_types;"

docker exec pustaka-postgres psql -U pustaka -d pustaka -c "SELECT name, is_active, is_system FROM relationship_types LIMIT 10;"

# Run Playwright test
echo "7. Running Playwright test..."
cd /home/syam/dev/Pustaka
npx playwright test tests/e2e/debug-relationship-types-issue.spec.ts --headed

echo "=== Debugging Complete ==="

# Cleanup background processes
if [ ! -z "$API_PID" ]; then
    kill $API_PID 2>/dev/null
fi
if [ ! -z "$FRONTEND_PID" ]; then
    kill $FRONTEND_PID 2>/dev/null
fi