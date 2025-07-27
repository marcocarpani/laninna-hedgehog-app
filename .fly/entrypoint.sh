#!/bin/bash
# 🦔 La Ninna - Fly.io Entrypoint Script

set -e

echo "🦔 Starting La Ninna Hedgehog Management System..."

# Generate Swagger documentation if not exists
if [ ! -f "./docs/docs.go" ]; then
    echo "📚 Generating Swagger documentation..."
    if command -v swag &> /dev/null; then
        swag init
    else
        echo "⚠️ Swagger CLI not found, skipping documentation generation"
    fi
fi

# Start the application
echo "🚀 Starting application..."
exec ./laninna-app