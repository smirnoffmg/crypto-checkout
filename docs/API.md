# Crypto Checkout API v1

- [Crypto Checkout API v1](#crypto-checkout-api-v1)
  - [Base URLs](#base-urls)
  - [Authentication](#authentication)
    - [API Key Authentication (Server-to-Server)](#api-key-authentication-server-to-server)
    - [JWT Token Authentication (Interactive Applications)](#jwt-token-authentication-interactive-applications)
    - [Permission Scopes](#permission-scopes)
  - [Multi-Tier Rate Limiting](#multi-tier-rate-limiting)
    - [Rate Limit Tiers](#rate-limit-tiers)
    - [Rate Limit Headers](#rate-limit-headers)
  - [Merchant Management](#merchant-management)
    - [Create Merchant Account](#create-merchant-account)
    - [Get Merchant Details](#get-merchant-details)
  - [Invoice Management](#invoice-management)
    - [Create Invoice](#create-invoice)
    - [Get Invoice (Merchant View)](#get-invoice-merchant-view)
    - [List Invoices](#list-invoices)
  - [Customer API (Public) \& Payment Web App](#customer-api-public--payment-web-app)
    - [Payment Web App Architecture](#payment-web-app-architecture)
    - [View Invoice (Customer)](#view-invoice-customer)
    - [Real-time Payment Updates (Server-Sent Events)](#real-time-payment-updates-server-sent-events)
    - [Get QR Code](#get-qr-code)
  - [Settlement API](#settlement-api)
    - [Get Settlement Details](#get-settlement-details)
    - [List Settlements](#list-settlements)
  - [Analytics \& Reporting](#analytics--reporting)
    - [Get Analytics Dashboard](#get-analytics-dashboard)
  - [Webhook Management](#webhook-management)
    - [Create Webhook Endpoint](#create-webhook-endpoint)
    - [Webhook Event Payloads](#webhook-event-payloads)
  - [Error Handling](#error-handling)
    - [Error Response Format](#error-response-format)
    - [HTTP Status Codes](#http-status-codes)
  - [Integration Examples](#integration-examples)
    - [Complete Payment Flow](#complete-payment-flow)
    - [Settlement Reconciliation Flow](#settlement-reconciliation-flow)

## Base URLs

**Merchant/Admin API** (Authenticated):
```
https://api.cryptocheckout.com/api/v1
```

**Customer API** (Public):
```
https://api.cryptocheckout.com/api/v1/public
```

**Payment Web App** (HTML Pages):
```
https://pay.cryptocheckout.com
```

**Real-time Events**:
```
wss://events.cryptocheckout.com/api/v1
```

## Authentication

### API Key Authentication (Server-to-Server)
```http
Authorization: Bearer sk_live_abc123...
```

**API Key Types:**
- `sk_live_*` - Production keys with full permissions
- `sk_test_*` - Sandbox keys for testing
- `sk_live_invoices_*` - Production keys limited to invoice operations
- `sk_live_analytics_*` - Production keys limited to analytics

### JWT Token Authentication (Interactive Applications)
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
- `settlements:read` - Access settlement data
- `*` - Full access (API keys only)

---

## Multi-Tier Rate Limiting

### Rate Limit Tiers

| Tier                | Scope                | Limits                                                                                                                      | Purpose                   |
| ------------------- | -------------------- | --------------------------------------------------------------------------------------------------------------------------- | ------------------------- |
| **API Key Tier**    | Per merchant API key | Invoice creation: 1000/hour<br/>Status checks: 5000/hour<br/>Settlement queries: 2000/hour<br/>Webhook management: 100/hour | Business logic protection |
| **IP Address Tier** | Per client IP        | 100 requests/minute<br/>Burst: 200 requests/5min                                                                            | Abuse prevention          |
| **Global Tier**     | System-wide          | 1000 requests/second<br/>Peak: 2000 requests/second                                                                         | Infrastructure protection |

### Rate Limit Headers
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 950
X-RateLimit-Reset: 1642694400
X-RateLimit-Window: 3600
X-RateLimit-Policy: sliding-window
X-RateLimit-Tier: api-key
Retry-After: 60
```

---

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
    "platform_fee_percentage": 1.0,
    "payment_tolerance": {
      "underpayment_threshold": 0.01,
      "overpayment_threshold": 1.00,
      "overpayment_action": "credit_account"
    },
    "confirmation_settings": {
      "merchant_override": null,
      "use_amount_based_defaults": true
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
    "platform_fee_percentage": 1.0,
    "payment_tolerance": {
      "underpayment_threshold": 0.01,
      "overpayment_threshold": 1.00,
      "overpayment_action": "credit_account"
    },
    "confirmation_settings": {
      "merchant_override": null,
      "use_amount_based_defaults": true
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

**Response includes settlement metrics:**
```json
{
  "id": "mer_abc123",
  "business_name": "Acme VPN Services",
  "contact_email": "admin@acmevpn.com",
  "status": "active",
  "plan_type": "pro",
  "settings": {
    "platform_fee_percentage": 1.0,
    "confirmation_settings": {
      "merchant_override": null,
      "use_amount_based_defaults": true
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
  "settlement_summary": {
    "total_gross_volume": 125000.50,
    "total_platform_fees": 1250.00,
    "total_net_payouts": 123750.50,
    "settlement_success_rate": 99.8
  },
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:00:00Z"
}
```

---

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
  "confirmation_settings": {
    "merchant_override": 12
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

**Response (Merchant View with Settlement Preview):**
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
  "qr_code_url": "https://api.cryptocheckout.com/api/v1/public/invoice/inv_abc123/qr?size=256",
  "status": "pending",
  "customer_url": "https://pay.cryptocheckout.com/invoice/inv_abc123",
  "expires_at": "2025-01-15T10:30:00Z",
  "created_at": "2025-01-15T10:00:00Z",
  "exchange_rate": {
    "rate": 1.0001,
    "source": "coinbase_pro",
    "locked_at": "2025-01-15T10:00:00Z",
    "expires_at": "2025-01-15T10:30:00Z"
  },
  "confirmation_settings": {
    "required_confirmations": 12,
    "merchant_override": 12,
    "amount_based_default": 12
  },
  "settlement_preview": {
    "gross_amount": 16.49,
    "platform_fee_amount": 0.16,
    "platform_fee_percentage": 1.0,
    "net_amount": 16.33
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

**Response includes complete administrative data:**
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
  "settlement": {
    "id": "set_456",
    "gross_amount": 16.49,
    "platform_fee_amount": 0.16,
    "platform_fee_percentage": 1.0,
    "net_amount": 16.33,
    "status": "completed",
    "settled_at": "2025-01-15T10:18:30Z"
  },
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
  ]
}
```

### List Invoices
```http
GET /api/v1/invoices?status=pending&limit=50&cursor=eyJpZCI6Imludl8xMjMifQ&created_after=2025-01-01T00:00:00Z
Authorization: Bearer sk_live_abc123...
```

**Query Parameters:**
- `status` - Filter by status (`pending`, `partial`, `confirming`, `paid`, `expired`, `cancelled`)
- `limit` - Results per page (max 100, default 20)
- `cursor` - Pagination cursor
- `created_after` - ISO 8601 datetime filter
- `created_before` - ISO 8601 datetime filter
- `amount_gte` - Minimum amount filter
- `amount_lte` - Maximum amount filter
- `currency` - Filter by currency
- `search` - Text search in title, description, metadata

---

## Customer API (Public) & Payment Web App

### Payment Web App Architecture
The payment web app serves HTML payment pages with:
- Invoice data fetched from public API
- QR codes and payment instructions
- Real-time payment status updates via Server-Sent Events
- Success/failure redirects

### View Invoice (Customer)
```http
GET /api/v1/public/invoice/{invoice_id}
Host: api.cryptocheckout.com
```

**Response (public data only - no fees shown):**
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
Host: api.cryptocheckout.com
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
Host: api.cryptocheckout.com
```

**Query Parameters:**
- `size` - QR code size in pixels (128, 256, 512, 1024)
- `format` - Image format (`png`, `svg`, `pdf`)
- `style` - QR code style (`classic`, `modern`, `rounded`)
- `logo` - Include merchant logo (`true`, `false`)

---

## Settlement API

### Get Settlement Details
```http
GET /api/v1/settlements/{settlement_id}
Authorization: Bearer sk_live_abc123...
```

**Response (complete fee breakdown):**
```json
{
  "id": "set_456",
  "invoice_id": "inv_abc123",
  "merchant_id": "mer_abc123",
  "gross_amount": 16.49,
  "platform_fee_amount": 0.16,
  "platform_fee_percentage": 1.0,
  "net_amount": 16.33,
  "currency": "USDT",
  "status": "completed",
  "settled_at": "2025-01-15T10:18:30Z",
  "payout_details": {
    "payout_tx_hash": "def789...",
    "payout_network_fee": 0.05,
    "net_received": 16.28
  },
  "created_at": "2025-01-15T10:18:00Z"
}
```

### List Settlements
```http
GET /api/v1/settlements?start_date=2025-01-01&end_date=2025-01-31&limit=50
Authorization: Bearer sk_live_abc123...
```

**Query Parameters:**
- `start_date` - ISO 8601 date filter
- `end_date` - ISO 8601 date filter
- `status` - Filter by status (`pending`, `completed`, `failed`)
- `limit` - Results per page (max 100, default 20)
- `cursor` - Pagination cursor

**Response (with summary):**
```json
{
  "settlements": [
    {
      "id": "set_456",
      "invoice_id": "inv_abc123",
      "gross_amount": 16.49,
      "platform_fee_amount": 0.16,
      "net_amount": 16.33,
      "status": "completed",
      "settled_at": "2025-01-15T10:18:30Z"
    }
  ],
  "summary": {
    "total_gross_amount": 12500.00,
    "total_platform_fees": 125.00,
    "total_net_amount": 12375.00,
    "average_fee_percentage": 1.0,
    "settlement_count": 250
  },
  "pagination": {
    "total": 250,
    "limit": 50,
    "has_more": true,
    "next_cursor": "eyJpZCI6InNldF80NTYifQ"
  }
}
```

---

## Analytics & Reporting

### Get Analytics Dashboard
```http
GET /api/v1/analytics?period=30d&group_by=day&timezone=UTC
Authorization: Bearer sk_live_abc123...
```

**Response (with settlement metrics):**
```json
{
  "period": "30d",
  "timezone": "UTC",
  "metrics": {
    "total_invoices": 1250,
    "total_gross_amount": 45678.90,
    "total_platform_fees": 456.79,
    "total_net_payouts": 45222.11,
    "paid_invoices": 1100,
    "conversion_rate": 88.0,
    "average_gross_amount": 36.54,
    "average_net_amount": 36.17,
    "effective_fee_rate": 1.0,
    "average_payment_time": 425,
    "average_settlement_time": 28
  },
  "time_series": [
    {
      "date": "2025-01-01",
      "invoices_created": 42,
      "invoices_paid": 38,
      "gross_amount": 1534.50,
      "platform_fees": 15.35,
      "net_payouts": 1519.15,
      "conversion_rate": 90.5
    }
  ],
  "payment_methods": {
    "USDT": {
      "count": 1100,
      "gross_amount": 45678.90,
      "platform_fees": 456.79,
      "net_amount": 45222.11
    }
  }
}
```

---

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
  "events": ["invoice.paid", "settlement.completed", "invoice.expired"],
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
  "events": ["invoice.paid", "settlement.completed", "invoice.expired"],
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

### Webhook Event Payloads

**Settlement Completed Event:**
```json
{
  "id": "evt_123",
  "type": "settlement.completed",
  "created": "2025-01-15T10:18:30Z",
  "data": {
    "settlement": {
      "id": "set_456",
      "invoice_id": "inv_abc123",
      "gross_amount": 16.49,
      "platform_fee_amount": 0.16,
      "net_amount": 16.33,
      "settled_at": "2025-01-15T10:18:30Z"
    },
    "invoice": {
      "id": "inv_abc123",
      "title": "VPN Service Order",
      "status": "paid",
      "metadata": {
        "order_id": "ord_789"
      }
    }
  }
}
```

---

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
    "documentation_url": "https://docs.cryptocheckout.com/errors#invalid-amount"
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

---

## Integration Examples

### Complete Payment Flow

```mermaid
sequenceDiagram
    participant M as Merchant
    participant API as Crypto Checkout API
    participant C as Customer
    participant W as Payment Web App
    participant WH as Webhook
    
    M->>API: POST /api/v1/invoices
    API-->>M: Invoice with customer_url and settlement_preview
    M->>C: Redirect to customer_url
    
    C->>W: GET /invoice/inv_abc123
    W->>API: GET /api/v1/public/invoice/inv_abc123
    API-->>W: Invoice data (no fees shown)
    W-->>C: Payment page with QR code
    
    W->>API: SSE /api/v1/public/invoice/inv_abc123/events
    C->>C: Send crypto payment
    
    API->>W: payment.detected event
    API->>WH: POST webhook (payment.detected)
    
    API->>W: invoice.paid event
    API->>WH: POST webhook (settlement.completed)
    W->>C: Redirect to return_url
```

### Settlement Reconciliation Flow

```mermaid
sequenceDiagram
    participant Payment as Payment Confirmed
    participant Settlement as Settlement Service
    participant Merchant as Merchant Webhook
    participant Analytics as Analytics Service
    
    Payment->>Settlement: Create Settlement
    Settlement->>Settlement: Calculate Platform Fee
    Settlement->>Settlement: Calculate Net Amount
    Settlement->>Merchant: settlement.completed webhook
    Settlement->>Analytics: Update Revenue Metrics
    
    Note over Settlement: Gross: $16.49<br/>Fee: $0.16 (1%)<br/>Net: $16.33
```

This API specification provides complete payment processor functionality with transparent fee handling, near-real-time settlement processing, and comprehensive reporting capabilities while maintaining clean separation between customer and merchant experiences.