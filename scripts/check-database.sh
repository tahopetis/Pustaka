#!/bin/bash

echo "=== Checking Relationship Types Database ==="

# Check if Docker is running and container is accessible
if ! docker ps | grep -q "porta-postgres"; then
    echo "PostgreSQL container is not running. Starting Docker services..."
    make docker-up
    sleep 10
fi

# Check relationship types table
echo "1. Total relationship types:"
docker exec porta-postgres psql -U porta -d porta -c "SELECT COUNT(*) as total FROM relationship_types;"

echo "2. System vs Non-system types:"
docker exec porta-postgres psql -U porta -d porta -c "SELECT is_system, COUNT(*) as count FROM relationship_types GROUP BY is_system;"

echo "3. Active vs Inactive types:"
docker exec porta-postgres psql -U porta -d porta -c "SELECT is_active, COUNT(*) as count FROM relationship_types GROUP BY is_active;"

echo "4. Sample relationship types data:"
docker exec porta-postgres psql -U porta -d porta -c "SELECT id, name, display_name, is_active, is_system, category FROM relationship_types LIMIT 10;"

echo "5. Categories:"
docker exec porta-postgres psql -U porta -d porta -c "SELECT category, COUNT(*) as count FROM relationship_types WHERE category IS NOT NULL GROUP BY category;"

echo "=== Database Check Complete ==="