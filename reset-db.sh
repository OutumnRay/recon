#!/bin/bash

# Reset PostgreSQL database by dropping and recreating schema

echo "🔄 Stopping managing-portal..."
docker-compose stop managing-portal

echo "🗑️  Dropping and recreating database schema..."

# Check if we're using external PostgreSQL or Docker PostgreSQL
if docker ps -a | grep -q recontext-postgres; then
    # Docker PostgreSQL
    echo "Using Docker PostgreSQL..."
    docker exec recontext-postgres psql -U recontext -d recontext -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO recontext; GRANT ALL ON SCHEMA public TO public;"
else
    # External PostgreSQL (localhost or remote)
    echo "Using external PostgreSQL..."
    PGPASSWORD=recontext psql -h localhost -U recontext -d recontext -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO recontext; GRANT ALL ON SCHEMA public TO public;"
fi

if [ $? -eq 0 ]; then
    echo "✅ Database schema reset successfully"
else
    echo "❌ Failed to reset database schema"
    echo "ℹ️  Attempting manual reset through Docker..."

    # Try to find any PostgreSQL container
    PG_CONTAINER=$(docker ps -a | grep postgres | awk '{print $1}' | head -1)

    if [ -n "$PG_CONTAINER" ]; then
        docker exec $PG_CONTAINER psql -U recontext -d recontext -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO recontext; GRANT ALL ON SCHEMA public TO public;"
    fi
fi

echo "🚀 Starting managing-portal (will recreate schema)..."
docker-compose up -d managing-portal

echo "📋 Waiting for migrations to complete..."
sleep 5

echo "📊 Checking logs..."
docker-compose logs managing-portal | tail -20

echo ""
echo "✅ Database reset complete!"
echo "ℹ️  Check logs above to verify migrations completed successfully"
