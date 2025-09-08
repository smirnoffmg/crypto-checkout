#!/bin/bash

# API and Log Analysis Test
# This script tests API endpoints and analyzes Docker logs

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
KAFKA_BROKER="localhost:9092"
KAFKA_TOPIC="crypto-checkout.domain-events"
API_BASE_URL="http://localhost:8080"
TEST_TIMEOUT=30

# Test data
MERCHANT_ID="test-merchant-$(date +%s)"
INVOICE_AMOUNT="100.50"
CURRENCY="USD"
CRYPTO_CURRENCY="USDT"
DESCRIPTION="Kafka Integration Test Invoice"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up temporary files..."
    if [ ! -z "$CONSUMED_MESSAGES_FILE" ]; then
        rm -f "$CONSUMED_MESSAGES_FILE"
    fi
    # Note: NOT stopping containers - they should remain running for further testing
    log_info "Cleanup completed - containers remain running"
}

# Set up cleanup trap
trap cleanup EXIT

# Check if required tools are installed
check_dependencies() {
    log_info "Checking dependencies..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    # Check if docker compose is available (newer versions)
    if ! docker compose version &> /dev/null; then
        log_error "Docker Compose is not available"
        exit 1
    fi
    
    if ! command -v curl &> /dev/null; then
        log_error "curl is not installed"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_error "jq is not installed"
        exit 1
    fi
    
    if ! command -v openssl &> /dev/null; then
        log_error "openssl is not installed"
        exit 1
    fi
    
    log_success "All dependencies are available"
}

# Check if services are running
check_services_running() {
    log_info "Checking if services are running..."
    
    # Check if containers are running
    local containers=$(docker compose --env-file env.dev ps --format "table {{.Name}}\t{{.Status}}" | grep -E "(crypto-checkout|postgres|kafka|redis|zookeeper)" | grep -v "NAME")
    
    if [ -z "$containers" ]; then
        log_error "No services are running. Please start them first with: make up"
        exit 1
    fi
    
    log_success "Services are running:"
    echo "$containers"
    
    # Quick health check
    log_info "Performing quick health checks..."
    
    # Check application health
    if curl -s $API_BASE_URL/health > /dev/null 2>&1; then
        log_success "Application is responding"
    else
        log_error "Application is not responding"
        exit 1
    fi
    
    # Check PostgreSQL
    if docker compose --env-file env.dev exec -T postgres pg_isready -U ${CRYPTO_CHECKOUT_DATABASE_USER:-crypto_user} -d ${CRYPTO_CHECKOUT_DATABASE_DBNAME:-crypto_checkout} -h localhost > /dev/null 2>&1; then
        log_success "PostgreSQL is ready"
    else
        log_error "PostgreSQL is not ready"
        exit 1
    fi
    
    # Check Kafka
    if docker compose --env-file env.dev exec -T kafka kafka-broker-api-versions --bootstrap-server localhost:9092 > /dev/null 2>&1; then
        log_success "Kafka is ready"
    else
        log_error "Kafka is not ready"
        exit 1
    fi
}

# Start Kafka consumer to capture events
start_kafka_consumer() {
    log_info "Starting Kafka consumer to capture events..."
    
    # Create a temporary file to store consumed messages
    CONSUMED_MESSAGES_FILE=$(mktemp)
    
    # Start Kafka consumer in the background
    docker compose --env-file env.dev exec -T kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic "$KAFKA_TOPIC" --from-beginning > "$CONSUMED_MESSAGES_FILE" &
    KAFKA_CONSUMER_PID=$!
    
    # Give the consumer time to start
    sleep 3
    
    log_success "Kafka consumer started"
    log_info "Consumer output file: $CONSUMED_MESSAGES_FILE"
    echo "$CONSUMED_MESSAGES_FILE"
}

# Create an invoice
create_invoice() {
    log_info "Creating invoice..."
    
    local invoice_data=$(cat <<EOF
{
    "merchant_id": "$MERCHANT_ID",
    "amount": "$INVOICE_AMOUNT",
    "currency": "$CURRENCY",
    "crypto_currency": "$CRYPTO_CURRENCY",
    "description": "$DESCRIPTION"
}
EOF
)
    
    local response=$(curl -s -X POST "$API_BASE_URL/api/v1/invoices" \
        -H "Content-Type: application/json" \
        -d "$invoice_data")
    
    if [ $? -ne 0 ]; then
        log_error "Failed to create invoice"
        exit 1
    fi
    
    log_info "Invoice creation response: $response"
    
    # Extract invoice ID from response
    local invoice_id=$(echo "$response" | jq -r '.id')
    
    if [ "$invoice_id" = "null" ] || [ -z "$invoice_id" ]; then
        log_error "Failed to extract invoice ID from response: $response"
        exit 1
    fi
    
    log_success "Invoice created with ID: $invoice_id"
    echo "$invoice_id"
}

# Wait for events to be consumed
wait_for_events() {
    local messages_file="$1"
    local expected_event_type="$2"
    local timeout_seconds="$3"
    
    log_info "Waiting for event: $expected_event_type"
    
    local start_time=$(date +%s)
    while [ $(($(date +%s) - start_time)) -lt $timeout_seconds ]; do
        if grep -q "$expected_event_type" "$messages_file" 2>/dev/null; then
            log_success "Event found: $expected_event_type"
            return 0
        fi
        sleep 1
    done
    
    log_error "Timeout waiting for event: $expected_event_type"
    return 1
}

# Verify invoice created event
verify_invoice_created_event() {
    local messages_file="$1"
    local invoice_id="$2"
    
    log_info "Verifying invoice created event..."
    
    # Wait for the event
    if ! wait_for_events "$messages_file" "invoice.created" 10; then
        return 1
    fi
    
    # Check if the event contains the correct invoice ID
    if grep -q "$invoice_id" "$messages_file"; then
        log_success "Invoice created event verified for ID: $invoice_id"
        return 0
    else
        log_error "Invoice created event not found for ID: $invoice_id"
        return 1
    fi
}

# Process a payment
process_payment() {
    local invoice_id="$1"
    
    log_info "Processing payment for invoice: $invoice_id"
    
    # Get the invoice to get the payment address
    local invoice_response=$(curl -s "$API_BASE_URL/api/v1/invoices/$invoice_id")
    local payment_address=$(echo "$invoice_response" | jq -r '.payment_address')
    
    if [ "$payment_address" = "null" ] || [ -z "$payment_address" ]; then
        log_error "Failed to get payment address from invoice: $invoice_response"
        return 1
    fi
    
    log_info "Payment address: $payment_address"
    
    # Create a mock payment
    local payment_data=$(cat <<EOF
{
    "invoice_id": "$invoice_id",
    "amount": "$INVOICE_AMOUNT",
    "currency": "$CRYPTO_CURRENCY",
    "transaction_hash": "0x$(openssl rand -hex 32)",
    "from_address": "0x$(openssl rand -hex 20)",
    "to_address": "$payment_address",
    "confirmations": 1,
    "block_number": 12345
}
EOF
)
    
    # Process the payment
    local response=$(curl -s -X POST "$API_BASE_URL/api/v1/payments" \
        -H "Content-Type: application/json" \
        -d "$payment_data")
    
    if [ $? -ne 0 ]; then
        log_error "Failed to process payment"
        return 1
    fi
    
    log_success "Payment processed successfully"
    return 0
}

# Verify payment events
verify_payment_events() {
    local messages_file="$1"
    
    log_info "Verifying payment events..."
    
    # Wait for payment detected event
    if ! wait_for_events "$messages_file" "payment.detected" 10; then
        return 1
    fi
    
    # Wait for payment confirmed event
    if ! wait_for_events "$messages_file" "payment.confirmed" 10; then
        return 1
    fi
    
    # Wait for invoice paid event
    if ! wait_for_events "$messages_file" "invoice.paid" 10; then
        return 1
    fi
    
    log_success "All payment events verified"
    return 0
}

# Display consumed events
display_events() {
    local messages_file="$1"
    
    log_info "Consumed events:"
    echo "----------------------------------------"
    if [ -f "$messages_file" ] && [ -s "$messages_file" ]; then
        # Try to parse as JSON and extract event types
        while IFS= read -r line; do
            if echo "$line" | jq -e '.event_type' > /dev/null 2>&1; then
                local event_type=$(echo "$line" | jq -r '.event_type')
                local aggregate_id=$(echo "$line" | jq -r '.aggregate_id')
                echo "$event_type - $aggregate_id"
            else
                echo "$line"
            fi
        done < "$messages_file"
    else
        echo "No events consumed"
    fi
    echo "----------------------------------------"
}

# Test health endpoint
test_health_endpoint() {
    log_info "Testing health endpoint..."
    
    local response=$(curl -s "$API_BASE_URL/health")
    if [ $? -ne 0 ]; then
        log_error "Health endpoint test failed"
        return 1
    fi
    
    local status=$(echo "$response" | jq -r '.status')
    if [ "$status" = "healthy" ]; then
        log_success "Health endpoint is working"
        return 0
    else
        log_error "Health endpoint returned unhealthy status: $response"
        return 1
    fi
}

# Test API endpoints
test_api_endpoints() {
    log_info "Testing API endpoints..."
    
    # Test health endpoint
    log_info "Testing /health endpoint..."
    local health_response=$(curl -s "$API_BASE_URL/health")
    if [ $? -eq 0 ]; then
        log_success "Health endpoint responded: $health_response"
    else
        log_error "Health endpoint failed"
        return 1
    fi
    
    # Test API documentation
    log_info "Testing /swagger endpoint..."
    local swagger_code=$(curl -s -w "%{http_code}" -o /dev/null "$API_BASE_URL/swagger/index.html")
    if [ "$swagger_code" = "200" ]; then
        log_success "Swagger documentation accessible"
    else
        log_warning "Swagger documentation not accessible (HTTP $swagger_code)"
    fi
    
    # Test a simple API endpoint
    log_info "Testing API v1 root..."
    local api_code=$(curl -s -w "%{http_code}" -o /dev/null "$API_BASE_URL/api/v1/")
    log_info "API v1 root returned HTTP $api_code"
}

# Analyze Docker logs
analyze_logs() {
    log_info "Analyzing Docker logs..."
    echo "========================================"
    
    log_info "Application logs (last 15 lines):"
    echo "----------------------------------------"
    docker compose --env-file env.dev logs crypto-checkout --tail=15
    echo "----------------------------------------"
    
    log_info "Kafka logs (last 10 lines):"
    echo "----------------------------------------"
    docker compose --env-file env.dev logs kafka --tail=10
    echo "----------------------------------------"
    
    log_info "PostgreSQL logs (last 10 lines):"
    echo "----------------------------------------"
    docker compose --env-file env.dev logs postgres --tail=10
    echo "----------------------------------------"
    
    log_info "Redis logs (last 5 lines):"
    echo "----------------------------------------"
    docker compose --env-file env.dev logs redis --tail=5
    echo "========================================"
}

# Main test flow
main() {
    log_info "Starting API and Log Analysis Test"
    echo
    
    # Check dependencies
    check_dependencies
    
    # Check if services are already running
    check_services_running
    
    # Test API endpoints
    test_api_endpoints
    
    # Analyze logs
    analyze_logs
    
    log_success "API and Log Analysis Test completed!"
    
    # Note: cleanup is handled by the EXIT trap
    # Containers remain running for further testing
}

# Run the main function
main "$@"
