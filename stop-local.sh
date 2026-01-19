#!/bin/bash

# StreamNestAI Local Development Stop Script
# This script stops all running services

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

# Function to stop process by PID file
stop_service() {
    local pid_file=$1
    local service_name=$2
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if kill -0 "$pid" 2>/dev/null; then
            print_status "Stopping $service_name (PID: $pid)..."
            kill "$pid"
            
            # Wait for process to stop
            local count=0
            while kill -0 "$pid" 2>/dev/null && [ $count -lt 10 ]; do
                sleep 1
                ((count++))
            done
            
            # Force kill if still running
            if kill -0 "$pid" 2>/dev/null; then
                print_warning "Force stopping $service_name..."
                kill -9 "$pid"
            fi
            
            print_success "$service_name stopped"
        else
            print_warning "$service_name was not running"
        fi
        rm -f "$pid_file"
    else
        print_warning "No PID file found for $service_name"
    fi
}

# Function to kill processes by port
kill_by_port() {
    local port=$1
    local service_name=$2
    
    local pid=$(lsof -ti:$port 2>/dev/null || true)
    if [ ! -z "$pid" ]; then
        print_status "Killing $service_name process on port $port (PID: $pid)..."
        kill -9 "$pid" 2>/dev/null || true
        print_success "$service_name on port $port killed"
    fi
}

# Main stop function
main() {
    print_status "Stopping StreamNestAI Local Development Environment..."
    echo
    
    # Stop services using PID files
    stop_service ".backend.pid" "Backend"
    stop_service ".frontend.pid" "Frontend"
    stop_service ".cloudflare.pid" "Cloudflare Worker"
    
    echo
    
    # Kill any remaining processes on ports
    print_status "Checking for remaining processes on ports..."
    
    kill_by_port "8080" "Backend"
    kill_by_port "5173" "Frontend"
    kill_by_port "8787" "Cloudflare Worker"
    kill_by_port "9090" "Metrics"
    
    echo
    
    # Clean up any remaining Go processes
    print_status "Cleaning up any remaining Go processes..."
    pkill -f "go run main.go" 2>/dev/null || true
    pkill -f "streamnest-server" 2>/dev/null || true
    
    # Clean up any remaining Node processes
    print_status "Cleaning up any remaining Node processes..."
    pkill -f "vite" 2>/dev/null || true
    pkill -f "wrangler dev" 2>/dev/null || true
    
    echo
    print_success "🛑 All StreamNestAI services have been stopped"
    echo
    
    # Show final status
    print_status "Final port check:"
    
    for port in 8080 5173 8787 9090; do
        local pid=$(lsof -ti:$port 2>/dev/null || true)
        if [ ! -z "$pid" ]; then
            print_warning "Port $port is still in use by PID $pid"
        else
            print_success "Port $port is free"
        fi
    done
    
    echo
    print_success "Cleanup completed!"
}

# Run main function
main "$@"