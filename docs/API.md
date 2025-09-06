# Crypto Checkout API

- [Crypto Checkout API](#crypto-checkout-api)
  - [Base URLs](#base-urls)
  - [Authentication](#authentication)
    - [API Key Authentication (Recommended for Server-to-Server)](#api-key-authentication-recommended-for-server-to-server)
    - [JWT Token Authentication (Recommended for Interactive Applications)](#jwt-token-authentication-recommended-for-interactive-applications)
    - [Permission Scopes](#permission-scopes)
  - [Core Endpoints](#core-endpoints)
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
  - [Webhook System](#webhook-system)
    - [Configure Webhook Endpoint](#configure-webhook-endpoint)
    - [Webhook Events](#webhook-events)
    - [Webhook Payload](#webhook-payload)
  - [Error Handling](#error-handling)
    - [Error Response Format](#error-response-format)
    - [HTTP Status Codes](#http-status-codes)
  - [Rate Limiting](#rate-limiting)
    - [Merchant/Admin API Limits](#merchantadmin-api-limits)
    - [Customer API Limits](#customer-api-limits)
    - [Rate Limit Headers](#rate-limit-headers)
  - [Integration Examples](#integration-examples)
    - [Merchant Integration Flow](#merchant-integration-flow)
    - [Payment Web App Data Flow](#payment-web-app-data-flow)
    - [Webhook Delivery Flow](#webhook-delivery-flow)
  - [Security Features](#security-features)
  - [Deployment Architecture](#deployment-architecture)
    - [Service Separation](#service-separation)
    - [System Communication Flow](#system-communication-flow)
    - [Benefits of Separate Web App](#benefits-of-separate-web-app)


## Base URLs

**Merchant/Admin API** (Authenticated):
```
https://api.crypto-checkout.com/api/v1
```

**Customer API** (Public):
```
https://api.crypto-checkout.com/api/v1/public
```

**Payment Web App** (HTML Pages):
```
https://pay.crypto-checkout.com
```

**Real-time Events**:
```
wss://events.crypto-checkout.com/api/v1
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
- `invoices:create` - Create new invoices
- `invoices:read` - Read invoice data
- `invoices:cancel` - Cancel invoices
- `invoices:refund` - Process refunds
- `analytics:read` - Access analytics data
- `webhooks:manage` - Configure webhooks
- `*` - Full access (API keys only)

## Core Endpoints

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
  "qr_code_url": "https://api.crypto-checkout.com/api/v1/public/invoice/inv_abc123/qr?size=256",
  "status": "created",
  "customer_url": "https://pay.crypto-checkout.com/invoice/inv_abc123",
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
GET https://pay.crypto-checkout.com/invoice/{invoice_id}
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
1. **Customer visits payment URL**: `https://pay.crypto-checkout.com/invoice/inv_abc123`
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
Host: api.crypto-checkout.com
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
Host: api.crypto-checkout.com
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
Host: api.crypto-checkout.com
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

## Webhook System

### Configure Webhook Endpoint
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

### Webhook Events
- `invoice.created` - New invoice created
- `invoice.expired` - Invoice expired without payment
- `invoice.cancelled` - Invoice manually cancelled
- `payment.detected` - New payment detected (0 confirmations)
- `payment.confirming` - Payment gaining confirmations
- `payment.confirmed` - Payment fully confirmed
- `invoice.paid` - Invoice fully paid and confirmed
- `invoice.refunded` - Invoice refunded
- `invoice.overpaid` - Invoice received overpayment

### Webhook Payload
```http
POST https://merchant.com/webhook
Content-Type: application/json
X-Crypto-Checkout-Signature: sha256=abc123...
X-Crypto-Checkout-Timestamp: 1642694400
X-Crypto-Checkout-Webhook-Id: whk_abc123
X-Crypto-Checkout-Event: invoice.paid
```

**Payload:**
```json
{
  "id": "evt_abc123",
  "event": "invoice.paid",
  "api_version": "v1",
  "created_at": "2025-01-15T10:18:00Z",
  "data": {
    "invoice_id": "inv_abc123",
    "status": "paid",
    "total_amount": 16.49,
    "total_received": 16.49,
    "currency": "USD",
    "crypto_currency": "USDT",
    "payment_id": "pay_xyz789",
    "tx_hash": "a7b2c3d4e5f6789012345678901234567890abcdef",
    "confirmations": 15,
    "paid_at": "2025-01-15T10:18:00Z",
    "metadata": {
      "customer_id": "cust_456",
      "order_id": "ord_789"
    }
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
    "documentation_url": "https://docs.crypto-checkout.com/errors#invalid-amount"
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
- **Invoice creation**: 1000/hour per API key
- **Bulk operations**: 100/hour per API key
- **Invoice/payment status checks**: 5000/hour per API key
- **List operations**: 2000/hour per API key
- **Analytics**: 500/hour per API key
- **Webhook management**: 100/hour per API key

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

## Integration Examples

### Merchant Integration Flow

```mermaid
sequenceDiagram
    participant M as Merchant
    participant API as Crypto Checkout API
    participant C as Customer
    participant W as Payment Web App
    
    M->>API: POST /api/v1/invoices
    API-->>M: Invoice with status "created"
    M->>C: Redirect to customer_url
    C->>W: GET /invoice/inv_abc123
    Note over W,API: Invoice status changes to "pending"
    W->>API: GET /api/v1/public/invoice/inv_abc123
    API-->>W: Invoice data (JSON)
    W-->>C: Payment page (HTML)
    
    W->>API: SSE /api/v1/public/invoice/inv_abc123/events
    C->>C: Send crypto payment
    API->>W: payment.detected event
    API->>W: invoice.paid event
    W->>C: Redirect to return_url
    API->>M: Webhook notification
```

### Payment Web App Data Flow

```mermaid
graph TB
    A[Customer visits payment URL] --> B[Web app loads]
    B --> C[Fetch invoice data from public API]
    C --> D[Invoice status: created → pending]
    D --> E[Display payment interface]
    E --> F[Establish SSE connection]
    F --> G[Customer sends payment]
    G --> H[Real-time status updates]
    H --> I{Payment complete?}
    I -->|Yes| J[Redirect to success URL]
    I -->|No| H
    I -->|Expired| K[Show expiration message]
```

### Webhook Delivery Flow

```mermaid
sequenceDiagram
    participant P as Payment System
    participant K as Kafka
    participant W as Webhook Service
    participant M as Merchant Endpoint
    
    P->>K: Publish invoice.paid event
    K->>W: Consume event
    W->>W: Create webhook payload
    W->>W: Sign with HMAC
    W->>M: POST webhook with signature
    M-->>W: 200 OK (or error)
    
    alt Delivery failed
        W->>W: Schedule retry (exponential backoff)
        W->>M: Retry delivery
    end
    
    alt Max retries exceeded
        W->>W: Mark as failed
        W->>W: Send alert to monitoring
    end
```

## Security Features

- **HMAC Webhook Signatures** - Verify webhook authenticity
- **Idempotency Keys** - Prevent duplicate invoice creation
- **IP Allowlisting** - Restrict webhook delivery to specific IPs
- **Scoped API Keys** - Granular permission control
- **Rate Limiting** - Protect against abuse
- **HTTPS Everywhere** - All communication encrypted
- **Audit Logging** - Complete activity tracking

## Deployment Architecture

### Service Separation

```mermaid
graph TB
    subgraph "Crypto Checkout System"
        subgraph "API Layer"
            MA[Merchant API<br/>api.crypto-checkout.com/api/v1]
            CA[Customer API<br/>api.crypto-checkout.com/api/v1/public]
        end
        
        subgraph "Frontend Layer"
            WA[Payment Web App<br/>pay.crypto-checkout.com]
        end
        
        subgraph "Infrastructure Layer"
            PG[(PostgreSQL<br/>Event Store + Read Models)]
            KF[Kafka<br/>Event Bus]
            RD[(Redis<br/>Cache + Sessions)]
            BS[Blockchain Scanner]
        end
    end
    
    MA --> PG
    MA --> KF
    MA --> RD
    
    CA --> PG
    CA --> RD
    
    WA --> CA
    
    BS --> KF
    KF --> PG
```

### System Communication Flow

```mermaid
graph LR
    M[Merchants] --> MA[Merchant API]
    C[Customers] --> WA[Payment Web App]
    WA --> CA[Customer API]
    
    MA --> ES[Event Store]
    CA --> RM[Read Models]
    
    ES --> EB[Event Bus]
    EB --> WH[Webhook Delivery]
    EB --> RM
    EB --> RT[Real-time Updates]
    
    BC[Blockchain] --> BS[Scanner]
    BS --> EB
```

### Benefits of Separate Web App
- **Independent scaling**: Payment pages can be cached/CDN-delivered
- **Technology flexibility**: Web app can use different stack (React, Vue, etc.)
- **Security isolation**: Payment UI separated from sensitive API operations  
- **Performance optimization**: Static assets, progressive loading
- **Custom branding**: Easy white-labeling for different merchants

This API provides enterprise-grade cryptocurrency payment processing with developer-friendly integration patterns and comprehensive security features.