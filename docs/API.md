# Crypto Checkout API v1

- [Crypto Checkout API v1](#crypto-checkout-api-v1)
  - [Base URLs](#base-urls)
  - [Authentication](#authentication)
    - [API Key Authentication (Recommended for Server-to-Server)](#api-key-authentication-recommended-for-server-to-server)
    - [JWT Token Authentication (Recommended for Interactive Applications)](#jwt-token-authentication-recommended-for-interactive-applications)
    - [Permission Scopes](#permission-scopes)
  - [Merchant Management](#merchant-management)
    - [Create Merchant Account](#create-merchant-account)
    - [Get Merchant Details](#get-merchant-details)
    - [Update Merchant Settings](#update-merchant-settings)
  - [API Key Management](#api-key-management)
    - [Create API Key](#create-api-key)
    - [List API Keys](#list-api-keys)
    - [Update API Key](#update-api-key)
    - [Revoke API Key](#revoke-api-key)
  - [Webhook Management](#webhook-management)
    - [Create Webhook Endpoint](#create-webhook-endpoint)
    - [List Webhook Endpoints](#list-webhook-endpoints)
    - [Update Webhook Endpoint](#update-webhook-endpoint)
    - [Delete Webhook Endpoint](#delete-webhook-endpoint)
    - [Test Webhook Endpoint](#test-webhook-endpoint)
    - [Get Webhook Deliveries](#get-webhook-deliveries)
  - [Invoice Management](#invoice-management)
    - [Create Invoice](#create-invoice)
    - [Get Invoice (Merchant View)](#get-invoice-merchant-view)
    - [List Invoices](#list-invoices)
    - [Cancel Invoice](#cancel-invoice)
  - [Customer API (Public) \& Payment Web App](#customer-api-public--payment-web-app)
    - [Payment Web App (HTML Interface)](#payment-web-app-html-interface)
      - [Payment Web App Architecture](#payment-web-app-architecture)
      - [Payment Flow](#payment-flow)
    - [Customer API Endpoints (JSON Data)](#customer-api-endpoints-json-data)
      - [View Invoice (Customer)](#view-invoice-customer)
    - [Real-time Payment Updates (Server-Sent Events)](#real-time-payment-updates-server-sent-events)
    - [Get QR Code](#get-qr-code)
  - [Analytics \& Reporting](#analytics--reporting)
    - [Get Analytics Dashboard](#get-analytics-dashboard)
  - [Error Handling](#error-handling)
    - [Error Response Format](#error-response-format)
    - [HTTP Status Codes](#http-status-codes)
  - [Rate Limiting](#rate-limiting)
    - [Merchant/Admin API Limits](#merchantadmin-api-limits)
    - [Customer API Limits](#customer-api-limits)
    - [Rate Limit Headers](#rate-limit-headers)
  - [Security Features](#security-features)
  - [Integration Flow Examples](#integration-flow-examples)
    - [Complete Merchant Onboarding Flow](#complete-merchant-onboarding-flow)
    - [Payment Processing Flow](#payment-processing-flow)

## Base URLs

**Merchant/Admin API** (Authenticated):
```
https://api.thecryptocheckout.com/api/v1
```

**Customer API** (Public):
```
https://api.thecryptocheckout.com/api/v1/public
```

**Payment Web App** (HTML Pages):
```
https://pay.thecryptocheckout.com
```

**Real-time Events**:
```
wss://events.thecryptocheckout.com/api/v1
```

## Authentication

### API Key Authentication (Recommended for Server-to-Server)
```http
Authorization: Bearer sk_live_abc123...
```

**API Key Types:**
- `sk_live_*` - Production keys with full permissions
- `sk_test_*` - Sandbox keys for testing
- `sk_live_invoices_*` - Production keys limited to invoice operations
- `sk_live_analytics_*` - Production keys limited to analytics

### JWT Token Authentication (Recommended for Interactive Applications)
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Get JWT Token:**
```http
POST /api/v1/auth/token
Content-Type: application/json

{
  "grant_type": "api_key",
  "api_key": "sk_live_abc123...",
  "scope": ["invoices:create", "invoices:read", "analytics:read"],
  "expires_in": 3600
}
```

### Permission Scopes
- `merchants:read` - Read merchant data
- `merchants:write` - Update merchant settings
- `api_keys:manage` - Create and manage API keys
- `invoices:create` - Create new invoices
- `invoices:read` - Read invoice data
- `invoices:cancel` - Cancel invoices
- `invoices:refund` - Process refunds
- `analytics:read` - Access analytics data
- `webhooks:manage` - Configure webhooks
- `*` - Full access (API keys only)

## Merchant Management

### Create Merchant Account
```http
POST /api/v1/merchants
Content-Type: application/json
```

**Request:**
```json
{
  "business_name": "Acme VPN Services",
  "contact_email": "admin@acmevpn.com",
  "plan_type": "pro",
  "settings": {
    "default_currency": "USD",
    "default_crypto_currency": "USDT",
    "invoice_expiry_minutes": 30,
    "payment_tolerance": {
      "underpayment_threshold": 0.01,
      "overpayment_threshold": 1.00,
      "overpayment_action": "credit_account"
    }
  }
}
```

**Response:**
```json
{
  "id": "mer_abc123",
  "business_name": "Acme VPN Services",
  "contact_email": "admin@acmevpn.com",
  "status": "active",
  "plan_type": "pro",
  "settings": {
    "default_currency": "USD",
    "default_crypto_currency": "USDT",
    "invoice_expiry_minutes": 30,
    "payment_tolerance": {
      "underpayment_threshold": 0.01,
      "overpayment_threshold": 1.00,
      "overpayment_action": "credit_account"
    }
  },
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:00:00Z"
}
```

### Get Merchant Details
```http
GET /api/v1/merchants/me
Authorization: Bearer sk_live_abc123...
```

**Response:**
```json
{
  "id": "mer_abc123",
  "business_name": "Acme VPN Services",
  "contact_email": "admin@acmevpn.com",
  "status": "active",
  "plan_type": "pro",
  "settings": {
    "default_currency": "USD",
    "default_crypto_currency": "USDT",
    "invoice_expiry_minutes": 30,
    "payment_tolerance": {
      "underpayment_threshold": 0.01,
      "overpayment_threshold": 1.00,
      "overpayment_action": "credit_account"
    }
  },
  "plan_limits": {
    "max_api_keys": 10,
    "max_webhook_endpoints": 5,
    "monthly_invoice_limit": 10000
  },
  "usage": {
    "api_keys_count": 3,
    "webhook_endpoints_count": 2,
    "monthly_invoices": 1250
  },
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:00:00Z"
}
```

### Update Merchant Settings
```http
PATCH /api/v1/merchants/me
Authorization: Bearer sk_live_abc123...
Content-Type: application/json
```

**Request:**
```json
{
  "business_name": "Acme VPN Services LLC",
  "settings": {
    "default_currency": "EUR",
    "invoice_expiry_minutes": 45
  }
}
```

## API Key Management

### Create API Key
```http
POST /api/v1/api-keys
Authorization: Bearer sk_live_abc123...
Content-Type: application/json
```

**Request:**
```json
{
  "name": "Production API Key",
  "key_type": "live",
  "permissions": ["invoices:create", "invoices:read", "analytics:read"],
  "expires_at": "2025-12-31T23:59:59Z"
}
```

**Response:**
```json
{
  "id": "key_xyz789",
  "name": "Production API Key",
  "key": "sk_live_def456...",
  "key_type": "live",
  "permissions": ["invoices:create", "invoices:read", "analytics:read"],
  "status": "active",
  "expires_at": "2025-12-31T23:59:59Z",
  "created_at": "2025-01-15T10:00:00Z",
  "last_used_at": null
}
```

### List API Keys
```http
GET /api/v1/api-keys?status=active&limit=20
Authorization: Bearer sk_live_abc123...
```

**Query Parameters:**
- `status` - Filter by status (`active`, `revoked`, `expired`)
- `key_type` - Filter by type (`live`, `test`)
- `limit` - Results per page (max 100, default 20)

**Response:**
```json
{
  "api_keys": [
    {
      "id": "key_xyz789",
      "name": "Production API Key",
      "key_prefix": "sk_live_def456",
      "key_type": "live",
      "permissions": ["invoices:create", "invoices:read", "analytics:read"],
      "status": "active",
      "expires_at": "2025-12-31T23:59:59Z",
      "created_at": "2025-01-15T10:00:00Z",
      "last_used_at": "2025-01-15T14:30:00Z"
    }
  ],
  "pagination": {
    "total": 3,
    "limit": 20,
    "has_more": false
  }
}
```

### Update API Key
```http
PATCH /api/v1/api-keys/{key_id}
Authorization: Bearer sk_live_abc123...
Content-Type: application/json
```

**Request:**
```json
{
  "name": "Updated Production Key",
  "permissions": ["invoices:create", "invoices:read", "invoices:cancel", "analytics:read"]
}
```

### Revoke API Key
```http
DELETE /api/v1/api-keys/{key_id}
Authorization: Bearer sk_live_abc123...
```

**Response:**
```json
{
  "id": "key_xyz789",
  "status": "revoked",
  "revoked_at": "2025-01-15T15:00:00Z"
}
```

## Webhook Management

### Create Webhook Endpoint
```http
POST /api/v1/webhook-endpoints
Authorization: Bearer sk_live_abc123...
Content-Type: application/json
```

**Request:**
```json
{
  "url": "https://merchant.com/webhook",
  "events": ["invoice.paid", "payment.detected", "invoice.expired"],
  "secret": "whsec_abc123...",
  "allowed_ips": ["192.168.1.100", "10.0.0.0/8"],
  "max_retries": 5,
  "retry_backoff": "exponential",
  "timeout": 30,
  "enabled": true
}
```

**Response:**
```json
{
  "id": "whe_def456",
  "url": "https://merchant.com/webhook",
  "events": ["invoice.paid", "payment.detected", "invoice.expired"],
  "secret": "whsec_***123",
  "status": "active",
  "allowed_ips": ["192.168.1.100", "10.0.0.0/8"],
  "max_retries": 5,
  "retry_backoff": "exponential",
  "timeout": 30,
  "enabled": true,
  "created_at": "2025-01-15T10:00:00Z"
}
```

### List Webhook Endpoints
```http
GET /api/v1/webhook-endpoints?status=active
Authorization: Bearer sk_live_abc123...
```

**Response:**
```json
{
  "webhook_endpoints": [
    {
      "id": "whe_def456",
      "url": "https://merchant.com/webhook",
      "events": ["invoice.paid", "payment.detected", "invoice.expired"],
      "status": "active",
      "enabled": true,
      "last_delivery_at": "2025-01-15T14:30:00Z",
      "success_rate": 98.5,
      "created_at": "2025-01-15T10:00:00Z"
    }
  ]
}
```

### Update Webhook Endpoint
```http
PATCH /api/v1/webhook-endpoints/{endpoint_id}
Authorization: Bearer sk_live_abc123...
Content-Type: application/json
```

**Request:**
```json
{
  "events": ["invoice.paid", "invoice.cancelled"],
  "enabled": true,
  "max_retries": 3
}
```

### Delete Webhook Endpoint
```http
DELETE /api/v1/webhook-endpoints/{endpoint_id}
Authorization: Bearer sk_live_abc123...
```

### Test Webhook Endpoint
```http
POST /api/v1/webhook-endpoints/{endpoint_id}/test
Authorization: Bearer sk_live_abc123...
```

**Response:**
```json
{
  "test_id": "test_abc123",
  "status": "delivered",
  "response_code": 200,
  "response_time_ms": 150,
  "delivered_at": "2025-01-15T15:00:00Z"
}
```

### Get Webhook Deliveries
```http
GET /api/v1/webhook-endpoints/{endpoint_id}/deliveries?limit=50
Authorization: Bearer sk_live_abc123...
```

**Response:**
```json
{
  "deliveries": [
    {
      "id": "whd_456",
      "event_type": "invoice.paid",
      "invoice_id": "inv_abc123",
      "status": "delivered",
      "attempts": 1,
      "response_code": 200,
      "delivered_at": "2025-01-15T10:18:30Z",
      "created_at": "2025-01-15T10:18:00Z"
    }
  ]
}
```

## Invoice Management

### Create Invoice
```http
POST /api/v1/invoices
Authorization: Bearer sk_live_abc123...
Idempotency-Key: order_789_attempt_1
Content-Type: application/json
```

**Request:**
```json
{
  "title": "VPN Service Order",
  "description": "Monthly VPN subscription with premium features",
  "items": [
    {
      "name": "VPN Premium Plan",
      "description": "Monthly subscription with unlimited bandwidth",
      "quantity": 1,
      "unit_price": 9.99
    },
    {
      "name": "Additional Static IP",
      "quantity": 2,
      "unit_price": 2.50
    }
  ],
  "tax": 1.50,
  "currency": "USD",
  "crypto_currency": "USDT",
  "price_lock_duration": 1800,
  "expires_in": 1800,
  "payment_tolerance": {
    "underpayment_threshold": 0.01,
    "overpayment_threshold": 1.00,
    "overpayment_action": "credit_account"
  },
  "webhook_url": "https://merchant.com/webhook",
  "return_url": "https://merchant.com/success",
  "cancel_url": "https://merchant.com/cancel",
  "metadata": {
    "customer_id": "cust_456",
    "order_id": "ord_789",
    "customer_email": "customer@example.com"
  }
}
```

**Response:**
```json
{
  "id": "inv_abc123",
  "title": "VPN Service Order",
  "description": "Monthly VPN subscription with premium features",
  "items": [...],
  "subtotal": 14.99,
  "tax": 1.50,
  "total": 16.49,
  "currency": "USD",
  "crypto_currency": "USDT",
  "usdt_amount": 16.49,
  "address": "TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN",
  "qr_code_url": "https://api.thecryptocheckout.com/api/v1/public/invoice/inv_abc123/qr?size=256",
  "status": "created",
  "customer_url": "https://pay.thecryptocheckout.com/invoice/inv_abc123",
  "expires_at": "2025-01-15T10:30:00Z",
  "created_at": "2025-01-15T10:00:00Z",
  "exchange_rate": {
    "rate": 1.0001,
    "source": "coinbase_pro",
    "locked_at": "2025-01-15T10:00:00Z",
    "expires_at": "2025-01-15T10:30:00Z"
  },
  "payment_tolerance": {
    "underpayment_threshold": 0.01,
    "overpayment_threshold": 1.00,
    "overpayment_action": "credit_account"
  },
  "metadata": {
    "customer_id": "cust_456",
    "order_id": "ord_789",
    "customer_email": "customer@example.com"
  }
}
```

### Get Invoice (Merchant View)
```http
GET /api/v1/invoices/{invoice_id}
Authorization: Bearer sk_live_abc123...
```

**Response includes comprehensive administrative data:**
```json
{
  "id": "inv_abc123",
  "title": "VPN Service Order",
  "status": "paid",
  "total": 16.49,
  "currency": "USD",
  "usdt_amount": 16.49,
  "address": "TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN",
  "expires_at": "2025-01-15T10:30:00Z",
  "paid_at": "2025-01-15T10:18:00Z",
  "payments": [
    {
      "id": "pay_xyz789",
      "amount": 16.49,
      "tx_hash": "a7b2c3d4e5f6789012345678901234567890abcdef",
      "status": "confirmed",
      "confirmations": 15,
      "required_confirmations": 12,
      "block_number": 58234567,
      "from_address": "TMuA6YqfCeX8EhbfYEg5y7S4DqzSJireY9",
      "detected_at": "2025-01-15T10:15:00Z",
      "confirmed_at": "2025-01-15T10:18:00Z",
      "network_fee": 0.12
    }
  ],
  "audit_log": [
    {
      "id": "audit_123",
      "event": "invoice.created",
      "timestamp": "2025-01-15T10:00:00Z",
      "actor": "api_key:sk_live_***123",
      "ip_address": "192.168.1.100",
      "user_agent": "MyApp/1.0",
      "request_id": "req_abc123"
    }
  ],
  "webhook_deliveries": [
    {
      "id": "whd_456",
      "event": "invoice.paid",
      "url": "https://merchant.com/webhook",
      "status": "delivered",
      "attempts": 1,
      "response_code": 200,
      "delivered_at": "2025-01-15T10:18:30Z"
    }
  ]
}
```

### List Invoices
```http
GET /api/v1/invoices?status=pending&limit=50&cursor=eyJpZCI6Imludl8xMjMifQ&created_after=2025-01-01T00:00:00Z
Authorization: Bearer sk_live_abc123...
```

**Query Parameters:**
- `status` - Filter by status (`created`, `pending`, `paid`, `expired`, `cancelled`)
- `limit` - Results per page (max 100, default 20)
- `cursor` - Pagination cursor
- `created_after` - ISO 8601 datetime filter
- `created_before` - ISO 8601 datetime filter
- `amount_gte` - Minimum amount filter
- `amount_lte` - Maximum amount filter
- `currency` - Filter by currency
- `search` - Text search in title, description, metadata

### Cancel Invoice
```http
POST /api/v1/invoices/{invoice_id}/cancel
Authorization: Bearer sk_live_abc123...
Content-Type: application/json
```

**Request:**
```json
{
  "reason": "Customer requested cancellation",
  "refund_partial_payments": true
}
```

## Customer API (Public) & Payment Web App

### Payment Web App (HTML Interface)
```http
GET https://pay.thecryptocheckout.com/invoice/{invoice_id}
Accept: text/html
```

**Returns HTML payment page with embedded JavaScript for real-time updates.**

#### Payment Web App Architecture
The payment web app is a separate frontend application that:
- Serves HTML payment pages to customers
- Fetches invoice data from the public API
- Displays QR codes and payment instructions
- Provides real-time payment status updates
- Handles success/failure redirects

#### Payment Flow
1. **Customer visits payment URL**: `https://pay.thecryptocheckout.com/invoice/inv_abc123`
2. **Web app fetches invoice data**: `GET /api/v1/public/invoice/inv_abc123` 
3. **Displays payment interface**: QR code, amount, address, timer
4. **Establishes real-time connection**: Server-Sent Events for status updates
5. **Customer sends payment**: Via crypto wallet
6. **Real-time status updates**: Payment detected → confirming → confirmed
7. **Redirect on completion**: Success/cancel URLs from invoice

### Customer API Endpoints (JSON Data)

#### View Invoice (Customer)
```http
GET /api/v1/public/invoice/{invoice_id}
Host: api.thecryptocheckout.com
```

**Response (public data only):**
```json
{
  "id": "inv_abc123",
  "title": "VPN Service Order",
  "description": "Monthly VPN subscription with premium features",
  "items": [...],
  "subtotal": 14.99,
  "tax": 1.50,
  "total": 16.49,
  "currency": "USD",
  "crypto_currency": "USDT",
  "usdt_amount": 16.49,
  "address": "TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN",
  "status": "pending",
  "expires_at": "2025-01-15T10:30:00Z",
  "payments": [
    {
      "amount": 10.00,
      "status": "confirmed",
      "confirmations": 15,
      "confirmed_at": "2025-01-15T10:15:00Z"
    }
  ],
  "payment_progress": {
    "received": 10.00,
    "required": 16.49,
    "remaining": 6.49,
    "percent": 60.64
  },
  "return_url": "https://merchant.com/success",
  "cancel_url": "https://merchant.com/cancel",
  "time_remaining": 900
}
```

### Real-time Payment Updates (Server-Sent Events)
```http
GET /api/v1/public/invoice/{invoice_id}/events
Host: api.thecryptocheckout.com
Accept: text/event-stream
```

**Event Stream:**
```
data: {"event": "payment.detected", "payment": {"amount": 10.00, "status": "detected"}, "payment_progress": {"received": 10.00, "required": 16.49, "percent": 60.64}}

data: {"event": "payment.confirmed", "payment": {"amount": 10.00, "status": "confirmed", "confirmations": 12}}

data: {"event": "invoice.paid", "status": "paid", "paid_at": "2025-01-15T10:18:00Z"}
```

### Get QR Code
```http
GET /api/v1/public/invoice/{invoice_id}/qr?size=256&format=png&style=modern
Host: api.thecryptocheckout.com
```

**Query Parameters:**
- `size` - QR code size in pixels (128, 256, 512, 1024)
- `format` - Image format (`png`, `svg`, `pdf`)
- `style` - QR code style (`classic`, `modern`, `rounded`)
- `logo` - Include merchant logo (`true`, `false`)

## Analytics & Reporting

### Get Analytics Dashboard
```http
GET /api/v1/analytics?period=30d&group_by=day&timezone=UTC
Authorization: Bearer sk_live_abc123...
```

**Response:**
```json
{
  "period": "30d",
  "timezone": "UTC",
  "metrics": {
    "total_invoices": 1250,
    "total_amount": 45678.90,
    "paid_invoices": 1100,
    "paid_amount": 42350.75,
    "conversion_rate": 88.0,
    "average_amount": 36.54,
    "average_payment_time": 425
  },
  "time_series": [
    {
      "date": "2025-01-01",
      "invoices_created": 42,
      "invoices_paid": 38,
      "amount_created": 1534.50,
      "amount_paid": 1388.75,
      "conversion_rate": 90.5
    }
  ],
  "payment_methods": {
    "USDT": {"count": 800, "amount": 28500.00},
    "BTC": {"count": 200, "amount": 9850.75},
    "ETH": {"count": 100, "amount": 4000.00}
  }
}
```

## Error Handling

### Error Response Format
```json
{
  "error": {
    "type": "validation_error",
    "code": "INVALID_AMOUNT",
    "message": "Invoice amount must be between $0.50 and $10,000.00",
    "field": "items[0].unit_price",
    "constraints": {
      "min_amount": 0.50,
      "max_amount": 10000.00,
      "currency": "USD"
    },
    "suggestions": [
      "Adjust the unit price to be within the allowed range",
      "Split large orders into multiple invoices",
      "Contact support for higher limits"
    ],
    "documentation_url": "https://docs.thecryptocheckout.com/errors#invalid-amount"
  },
  "request_id": "req_abc123",
  "trace_id": "trace_xyz789",
  "timestamp": "2025-01-15T10:00:00Z"
}
```

### HTTP Status Codes
- `200` - Success
- `201` - Created
- `202` - Accepted (async processing)
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict (e.g., already cancelled)
- `422` - Validation Error
- `429` - Rate Limited
- `500` - Internal Error
- `502` - Blockchain Network Error
- `503` - Service Unavailable

## Rate Limiting

### Merchant/Admin API Limits
- **Merchant operations**: 100/hour per API key
- **API key management**: 50/hour per API key
- **Webhook management**: 100/hour per API key
- **Invoice creation**: 1000/hour per API key
- **Invoice/payment status checks**: 5000/hour per API key
- **List operations**: 2000/hour per API key
- **Analytics**: 500/hour per API key

### Customer API Limits
- **Invoice views**: 60/minute per IP
- **Real-time events**: 10 concurrent connections per invoice
- **QR requests**: 100/minute per IP

### Rate Limit Headers
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 950
X-RateLimit-Reset: 1642694400
X-RateLimit-Window: 3600
X-RateLimit-Policy: sliding-window
Retry-After: 60
```

## Security Features

- **HMAC Webhook Signatures** - Verify webhook authenticity
- **Idempotency Keys** - Prevent duplicate invoice creation
- **IP Allowlisting** - Restrict webhook delivery to specific IPs
- **Scoped API Keys** - Granular permission control
- **Rate Limiting** - Protect against abuse
- **HTTPS Everywhere** - All communication encrypted
- **Audit Logging** - Complete activity tracking

## Integration Flow Examples

### Complete Merchant Onboarding Flow

```mermaid
sequenceDiagram
    participant M as New Merchant
    participant API as Crypto Checkout API
    participant W as Webhook Endpoint
    
    M->>API: POST /api/v1/merchants (Create account)
    API-->>M: Merchant created with initial API key
    
    M->>API: POST /api/v1/api-keys (Create production key)
    API-->>M: Production API key
    
    M->>API: POST /api/v1/webhook-endpoints (Configure webhook)
    API-->>M: Webhook endpoint created
    
    M->>API: POST /api/v1/webhook-endpoints/{id}/test
    API->>W: Test webhook delivery
    W-->>API: 200 OK
    API-->>M: Test successful
    
    M->>API: POST /api/v1/invoices (Create first invoice)
    API-->>M: Invoice created
```

### Payment Processing Flow

```mermaid
sequenceDiagram
    participant M as Merchant
    participant API as Crypto Checkout API
    participant C as Customer
    participant W as Payment Web App
    participant WH as Webhook
    
    M->>API: POST /api/v1/invoices
    API-->>M: Invoice with customer_url
    M->>C: Redirect to customer_url
    
    C->>W: GET /invoice/inv_abc123
    W->>API: GET /api/v1/public/invoice/inv_abc123
    API-->>W: Invoice data
    W-->>C: Payment page with QR code
    
    W->>API: SSE /api/v1/public/invoice/inv_abc123/events
    C->>C: Send crypto payment
    
    API->>W: payment.detected event
    API->>WH: POST webhook (payment.detected)
    
    API->>W: invoice.paid event
    API->>WH: POST webhook (invoice.paid)
    W->>C: Redirect to return_url
```

This comprehensive API provides full merchant lifecycle management, from account creation through payment processing, with enterprise-grade security and developer-friendly integration patterns.