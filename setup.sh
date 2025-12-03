#!/bin/bash

# Union Products SSO - Setup Script
# This script helps you set up and run the SSO service

set -e

echo "🚀 Union Products SSO Setup"
echo "================================"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi

echo "✓ Go version: $(go version)"
echo ""

# Check if PostgreSQL is running
if command -v psql &> /dev/null; then
    echo "✓ PostgreSQL is installed"
else
    echo "⚠️  PostgreSQL not found in PATH (you can still use Docker)"
fi
echo ""

# Ask user for setup method
echo "How would you like to run the SSO service?"
echo "1) Docker Compose (recommended)"
echo "2) Local setup (requires PostgreSQL)"
echo ""
read -p "Enter your choice (1 or 2): " choice

case $choice in
    1)
        echo ""
        echo "📦 Setting up with Docker Compose..."
        echo ""
        
        if ! command -v docker-compose &> /dev/null && ! command -v docker &> /dev/null; then
            echo "❌ Docker is not installed. Please install Docker and Docker Compose."
            exit 1
        fi
        
        echo "✓ Docker is installed"
        echo ""
        
        # Check if .env exists
        if [ ! -f .env ]; then
            echo "📝 Creating .env file from .env.example..."
            cp .env.example .env
            echo "✓ .env file created"
            echo "⚠️  Please edit .env file with your configuration"
            echo ""
        fi
        
        echo "🐳 Starting Docker containers..."
        docker-compose up -d
        
        echo ""
        echo "✅ SSO service is starting!"
        echo ""
        echo "📊 View logs:"
        echo "   docker-compose logs -f sso-server"
        echo ""
        echo "🔍 Check status:"
        echo "   docker-compose ps"
        echo ""
        echo "🛑 Stop services:"
        echo "   docker-compose down"
        echo ""
        echo "🌐 Service URL: http://localhost:8080"
        echo "💚 Health check: http://localhost:8080/health"
        ;;
        
    2)
        echo ""
        echo "🔧 Setting up locally..."
        echo ""
        
        # Check if .env exists
        if [ ! -f .env ]; then
            echo "📝 Creating .env file from .env.example..."
            cp .env.example .env
            echo "✓ .env file created"
            echo ""
        fi
        
        # Load environment variables
        export $(cat .env | grep -v '^#' | xargs)
        
        # Check database connection
        echo "🔍 Checking database connection..."
        if psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" &> /dev/null; then
            echo "✓ Database connection successful"
        else
            echo "❌ Cannot connect to database"
            echo ""
            echo "Please ensure PostgreSQL is running and credentials are correct."
            echo "Database: $DB_NAME"
            echo "User: $DB_USER"
            echo "Host: $DB_HOST"
            echo ""
            read -p "Would you like to create the database? (y/n): " create_db
            if [ "$create_db" = "y" ]; then
                createdb -h "$DB_HOST" -U "$DB_USER" "$DB_NAME" || echo "Failed to create database"
            else
                exit 1
            fi
        fi
        
        # Run migrations
        echo ""
        echo "📊 Running database migrations..."
        PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -f database/migrations/001_initial_schema.sql
        echo "✓ Migrations completed"
        
        # Download Go dependencies
        echo ""
        echo "📦 Downloading Go dependencies..."
        go mod download
        echo "✓ Dependencies downloaded"
        
        # Build the application
        echo ""
        echo "🔨 Building application..."
        go build -o bin/sso-server ./cmd/server
        echo "✓ Build successful"
        
        echo ""
        echo "✅ Setup complete!"
        echo ""
        echo "🚀 Start the server:"
        echo "   ./bin/sso-server"
        echo "   or"
        echo "   make run"
        echo ""
        echo "🌐 Service URL: http://localhost:8080"
        echo "💚 Health check: http://localhost:8080/health"
        ;;
        
    *)
        echo "Invalid choice. Exiting."
        exit 1
        ;;
esac

echo ""
echo "================================"
echo "📚 Documentation:"
echo "   README.md - Project documentation"
echo "   API.md - API reference"
echo "   sdk/typescript/README.md - TypeScript SDK guide"
echo ""
echo "🎉 Happy coding!"
