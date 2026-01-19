#!/bin/bash

# StreamNestAI Local Development Startup Script
# This script starts all services for local development

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if a service is running
check_service() {
    local url=$1
    local service_name=$2
    
    if curl -s "$url" > /dev/null 2>&1; then
        print_success "$service_name is running"
        return 0
    else
        print_warning "$service_name is not responding"
        return 1
    fi
}

# Function to wait for service
wait_for_service() {
    local url=$1
    local service_name=$2
    local max_attempts=30
    local attempt=1
    
    print_status "Waiting for $service_name to start..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$url" > /dev/null 2>&1; then
            print_success "$service_name is ready!"
            return 0
        fi
        
        echo -n "."
        sleep 2
        ((attempt++))
    done
    
    print_error "$service_name failed to start within expected time"
    return 1
}

# Main startup function
main() {
    print_status "Starting StreamNestAI Local Development Environment..."
    echo
    
    # Check prerequisites
    print_status "Checking prerequisites..."
    
    if ! command -v node &> /dev/null; then
        print_error "Node.js is not installed"
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        exit 1
    fi
    
    if ! command -v mongod &> /dev/null; then
        print_warning "MongoDB might not be running. Please ensure MongoDB is started."
    fi
    
    if ! command -v redis-server &> /dev/null; then
        print_warning "Redis might not be running. Please ensure Redis is started."
    fi
    
    print_success "Prerequisites check completed"
    echo
    
    # Setup environment files if they don't exist
    if [ ! -f "Client/StreamNestAIClient/.env" ]; then
        print_status "Setting up frontend environment file..."
        cp Client/StreamNestAIClient/.env.local Client/StreamNestAIClient/.env
        print_success "Frontend environment file created"
    fi
    
    if [ ! -f "Server/StreamNestAIServer/.env" ]; then
        print_status "Setting up backend environment file..."
        cp Server/StreamNestAIServer/.env.local Server/StreamNestAIServer/.env
        print_success "Backend environment file created"
    fi
    
    # Install dependencies if needed
    if [ ! -d "Client/StreamNestAIClient/node_modules" ]; then
        print_status "Installing frontend dependencies..."
        cd Client/StreamNestAIClient && npm install && cd ../..
        print_success "Frontend dependencies installed"
    fi
    
    if [ ! -d "MCP-and-CDN/node_modules" ]; then
        print_status "Installing Cloudflare Worker dependencies..."
        cd MCP-and-CDN && npm install && cd ..
        print_success "Cloudflare Worker dependencies installed"
    fi
    
    echo
    print_status "Starting services..."
    echo
    
    # Start Backend
    print_status "Starting Backend server..."
    cd Server/StreamNestAIServer
    go run main.go > ../../logs/backend.log 2>&1 &
    BACKEND_PID=$!
    cd ../..
    echo $BACKEND_PID > .backend.pid
    print_success "Backend started (PID: $BACKEND_PID)"
    
    # Wait for Backend to be ready
    wait_for_service "http://localhost:8080/health" "Backend"
    echo
    
    # Start Cloudflare Worker
    print_status "Starting Cloudflare Worker..."
    cd MCP-and-CDN
    npx wrangler dev --config wrangler.local.toml > ../logs/cloudflare.log 2>&1 &
    CLOUDFLARE_PID=$!
    cd ..
    echo $CLOUDFLARE_PID > .cloudflare.pid
    print_success "Cloudflare Worker started (PID: $CLOUDFLARE_PID)"
    
    # Wait for Cloudflare Worker to be ready
    sleep 5
    wait_for_service "http://localhost:8787" "Cloudflare Worker"
    echo
    
    # Start Frontend
    print_status "Starting Frontend..."
    cd Client/StreamNestAIClient
    npm run dev > ../../logs/frontend.log 2>&1 &
    FRONTEND_PID=$!
    cd ../..
    echo $FRONTEND_PID > .frontend.pid
    print_success "Frontend started (PID: $FRONTEND_PID)"
    
    # Wait for Frontend to be ready
    sleep 5
    wait_for_service "http://localhost:5173" "Frontend"
    echo
    
    # Final status check
    print_status "Performing final service health check..."
    echo
    
    check_service "http://localhost:8080/health" "Backend API"
    check_service "http://localhost:5173" "Frontend"
    check_service "http://localhost:8787" "Cloudflare Worker"
    
    echo
    print_success "🎉 StreamNestAI Local Development Environment is ready!"
    echo
    echo "📍 Service URLs:"
    echo "   • Frontend:        http://localhost:5173"
    echo "   • Backend API:     http://localhost:8080"
    echo "   • Backend Health:  http://localhost:8080/health"
    echo "   • Cloudflare:      http://localhost:8787"
    echo "   • Metrics:         http://localhost:8080/metrics"
    echo
    echo "📝 Logs:"
    echo "   • Backend:   logs/backend.log"
    echo "   • Frontend:  logs/frontend.log"
    echo "   • Cloudflare: logs/cloudflare.log"
    echo
    echo "🛑 To stop all services, run: ./stop-local.sh"
    echo
    
    # Keep script running or exit based on argument
    if [ "$1" != "--detach" ]; then
        print_status "Monitoring services... Press Ctrl+C to stop"
        
        # Monitor services
        trap 'print_status "Stopping services..."; ./stop-local.sh; exit 0' INT
        
        while true; do
            sleep 10
            if ! check_service "http://localhost:8080/health" "Backend" > /dev/null 2>&1; then
                print_error "Backend service stopped unexpectedly"
                break
            fi
        done
    else
        print_success "Services started in detached mode"
    fi
}

# Create logs directory if it doesn't exist
mkdir -p logs

# Run main function
main "$@"