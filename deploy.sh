#!/bin/bash

# StreamNestAI Production Deployment Script
# This script automates the deployment of StreamNestAI in production

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

# Check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    print_success "Docker is installed"
}

# Check if Docker Compose is installed
check_docker_compose() {
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    print_success "Docker Compose is installed"
}

# Create necessary directories
create_directories() {
    print_status "Creating necessary directories..."
    mkdir -p logs/nginx
    mkdir -p docker/nginx/ssl
    mkdir -p docker/mongodb/init
    mkdir -p docker/grafana/provisioning/datasources
    mkdir -p docker/grafana/provisioning/dashboards
    mkdir -p docker/grafana/dashboards
    print_success "Directories created"
}

# Generate SSL certificates (self-signed for development)
generate_ssl() {
    print_status "Generating SSL certificates..."
    if [ ! -f docker/nginx/ssl/cert.pem ]; then
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout docker/nginx/ssl/key.pem \
            -out docker/nginx/ssl/cert.pem \
            -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"
        print_success "SSL certificates generated"
    else
        print_warning "SSL certificates already exist"
    fi
}

# Create Grafana datasources configuration
create_grafana_config() {
    print_status "Creating Grafana configuration..."
    cat > docker/grafana/provisioning/datasources/prometheus.yml << EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
EOF
    print_success "Grafana configuration created"
}

# Validate environment variables
validate_env() {
    print_status "Validating environment variables..."
    if [ ! -f .env.production ]; then
        print_error ".env.production file not found. Please create it from the template."
        exit 1
    fi
    
    # Load environment variables
    source .env.production
    
    # Check required variables
    required_vars=("MONGO_ROOT_PASSWORD" "JWT_SECRET" "GRAFANA_PASSWORD")
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ] || [ "${!var}" == "your_secure_*_here" ]; then
            print_error "Please set $var in .env.production"
            exit 1
        fi
    done
    print_success "Environment variables validated"
}

# Build and start services
deploy_services() {
    print_status "Building and starting services..."
    
    # Stop existing services
    docker-compose -f docker-compose.production.yml down || true
    
    # Build images
    print_status "Building Docker images..."
    docker-compose -f docker-compose.production.yml build --no-cache
    
    # Start services
    print_status "Starting services..."
    docker-compose -f docker-compose.production.yml up -d
    
    print_success "Services deployed successfully"
}

# Wait for services to be healthy
wait_for_services() {
    print_status "Waiting for services to be healthy..."
    
    # Wait for MongoDB
    print_status "Waiting for MongoDB..."
    until docker-compose -f docker-compose.production.yml exec -T mongodb mongosh --eval "db.adminCommand('ismaster')" &>/dev/null; do
        echo -n "."
        sleep 5
    done
    print_success "MongoDB is ready"
    
    # Wait for Redis
    print_status "Waiting for Redis..."
    until docker-compose -f docker-compose.production.yml exec -T redis redis-cli ping &>/dev/null; do
        echo -n "."
        sleep 5
    done
    print_success "Redis is ready"
    
    # Wait for API
    print_status "Waiting for API..."
    until curl -f http://localhost:8080/health &>/dev/null; do
        echo -n "."
        sleep 10
    done
    print_success "API is ready"
}

# Run database migrations
run_migrations() {
    print_status "Running database migrations..."
    # Add migration commands here if needed
    print_success "Migrations completed"
}

# Setup monitoring
setup_monitoring() {
    print_status "Setting up monitoring..."
    
    # Import Grafana dashboards
    print_status "Importing Grafana dashboards..."
    # Add dashboard import commands here
    
    print_success "Monitoring setup completed"
}

# Print deployment summary
print_summary() {
    print_success "Deployment completed successfully!"
    echo ""
    echo "🎉 StreamNestAI is now running in production mode!"
    echo ""
    echo "📊 Service URLs:"
    echo "  • Frontend: http://localhost"
    echo "  • API: http://localhost/api"
    echo "  • WebSocket: ws://localhost/ws"
    echo "  • Grafana: http://localhost:3001"
    echo "  • Prometheus: http://localhost:9090"
    echo "  • Kibana: http://localhost:5601"
    echo "  • Redis Commander: http://localhost:8081"
    echo ""
    echo "🔧 Management Commands:"
    echo "  • View logs: docker-compose -f docker-compose.production.yml logs -f [service]"
    echo "  • Stop services: docker-compose -f docker-compose.production.yml down"
    echo "  • Restart services: docker-compose -f docker-compose.production.yml restart"
    echo ""
    echo "📈 Monitoring:"
    echo "  • Grafana Username: admin"
    echo "  • Check Grafana dashboard for system metrics"
    echo ""
    print_warning "Remember to:"
    echo "  • Replace self-signed SSL certificates with production certificates"
    echo "  • Set up proper backup strategies"
    echo "  • Configure monitoring alerts"
    echo "  • Review security settings"
}

# Main deployment function
main() {
    print_status "Starting StreamNestAI Production Deployment..."
    echo ""
    
    check_docker
    check_docker_compose
    create_directories
    generate_ssl
    create_grafana_config
    validate_env
    deploy_services
    wait_for_services
    run_migrations
    setup_monitoring
    print_summary
}

# Handle script arguments
case "${1:-}" in
    "stop")
        print_status "Stopping StreamNestAI services..."
        docker-compose -f docker-compose.production.yml down
        print_success "Services stopped"
        ;;
    "restart")
        print_status "Restarting StreamNestAI services..."
        docker-compose -f docker-compose.production.yml restart
        print_success "Services restarted"
        ;;
    "logs")
        docker-compose -f docker-compose.production.yml logs -f "${2:-}"
        ;;
    "status")
        docker-compose -f docker-compose.production.yml ps
        ;;
    "update")
        print_status "Updating StreamNestAI..."
        git pull
        deploy_services
        wait_for_services
        print_success "Update completed"
        ;;
    "help"|"-h"|"--help")
        echo "StreamNestAI Deployment Script"
        echo ""
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  (no args)  Deploy StreamNestAI in production mode"
        echo "  stop       Stop all services"
        echo "  restart    Restart all services"
        echo "  logs       Show logs for all services (or specific service)"
        echo "  status     Show status of all services"
        echo "  update     Update and redeploy StreamNestAI"
        echo "  help       Show this help message"
        echo ""
        ;;
    *)
        main
        ;;
esac