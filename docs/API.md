# Crypto Checkout API

- [Crypto Checkout API](#crypto-checkout-api)
  - [Base URLs](#base-urls)
  - [API Versioning](#api-versioning)
  - [Authentication](#authentication)
    - [Merchant/Admin API Authentication](#merchantadmin-api-authentication)
    - [Customer Endpoints (No Auth)](#customer-endpoints-no-auth)
  - [Merchant/Admin API Endpoints](#merchantadmin-api-endpoints)
    - [Create Invoice](#create-invoice)
    - [Get Invoice (Merchant/Admin)](#get-invoice-merchantadmin)
    - [Cancel Invoice](#cancel-invoice)
    - [List Invoices](#list-invoices)
    - [Get Analytics](#get-analytics)
  - [Customer API Endpoints (Public)](#customer-api-endpoints-public)
    - [View Invoice (Customer)](#view-invoice-customer)
    - [Check Payment Status (Customer)](#check-payment-status-customer)
    - [Get QR Code](#get-qr-code)
  - [Immutable Invoice Design](#immutable-invoice-design)
  - [Status Values](#status-values)
    - [Invoice Status](#invoice-status)
    - [Payment Status](#payment-status)
  - [Webhooks](#webhooks)
    - [Webhook Events](#webhook-events)
    - [Webhook Payload](#webhook-payload)
    - [Webhook Security](#webhook-security)
  - [Error Handling](#error-handling)
    - [Error Response Format](#error-response-format)
    - [HTTP Status Codes](#http-status-codes)
  - [Rate Limiting](#rate-limiting)
    - [Merchant/Admin API Limits](#merchantadmin-api-limits)
    - [Customer API Limits](#customer-api-limits)
    - [Rate Limit Headers](#rate-limit-headers)
  - [Integration Examples](#integration-examples)
    - [Creating Invoice](#creating-invoice)
    - [Customer Payment Flow](#customer-payment-flow)
    - [Frontend Implementation Pattern](#frontend-implementation-pattern)
    - [Admin Dashboard Integration](#admin-dashboard-integration)

RESTful API for creating and managing cryptocurrency payment invoices. Serves as foundation for merchant integrations, admin interfaces, and customer payment flows.

## Base URLs

**Merchant/Admin API** (Authenticated):
```
https://api.crypto-checkout.com/v1
```

**Customer API** (Public):
```
https://checkout.crypto-checkout.com
```

## API Versioning

- **Current version**: `v1`
- **Versioning strategy**: URL path versioning (`/v1/`, `/v2/`)
- **Backward compatibility**: Maintained for 12 months minimum
- **Deprecation**: 6 months advance notice via `Sunset` header

## Authentication

### Merchant/Admin API Authentication
All authenticated endpoints require API key authentication:

```http
Authorization: Bearer sk_live_abc123...
```

**API Key Types:**
- `sk_live_*` - Production keys
- `sk_test_*` - Sandbox keys

**Security Best Practices:**
- Use HTTPS only
- Rotate keys every 90 days
- Store keys securely (environment variables)
- Different keys per environment

### Customer Endpoints (No Auth)
Customer-facing endpoints are public and require no authentication.

---

## Merchant/Admin API Endpoints

### Create Invoice
```http
POST /v1/invoices
Authorization: Bearer sk_live_abc123...
```

**Request:**
```json
{
  "title": "VPN Service Order",
  "description": "Monthly VPN subscription",
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
  "expires_in": 1800,
  "metadata": {
    "customer_id": "cust_456",
    "order_id": "ord_789"
  }
}
```

**Response:**
```json
{
  "id": "inv_abc123",
  "title": "VPN Service Order",
  "description": "Monthly VPN subscription",
  "items": [...],
  "subtotal": 14.99,
  "tax": 1.50,
  "total": 16.49,
  "currency": "USD",
  "usdt_amount": 16.49,
  "address": "TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN",
  "status": "pending",
  "customer_url": "https://checkout.crypto-checkout.com/invoice/inv_abc123",
  "expires_at": "2025-01-15T10:30:00Z",
  "created_at": "2025-01-15T10:00:00Z",
  "metadata": {
    "customer_id": "cust_456",
    "order_id": "ord_789"
  }
}
```

### Get Invoice (Merchant/Admin)
```http
GET /v1/invoices/{invoice_id}
Authorization: Bearer sk_live_abc123...
```

**Response includes administrative data:**
```json
{
  "id": "inv_abc123",
  "title": "VPN Service Order",
  "description": "Monthly VPN subscription",
  "items": [...],
  "subtotal": 14.99,
  "tax": 1.50,
  "total": 16.49,
  "currency": "USD",
  "usdt_amount": 16.49,
  "address": "TQn9Y2khEsLMWn1aXKURNC62XLFPqpTUcN",
  "status": "paid",
  "expires_at": "2025-01-15T10:30:00Z",
  "created_at": "2025-01-15T10:00:00Z",
  "metadata": {
    "customer_id": "cust_456",
    "order_id": "ord_789"
  },
  "payments": [
    {
      "id": "pay_xyz789",
      "amount": 11.49,
      "tx_hash": "a7b2c3d4e5f6789012345678901234567890abcdef",
      "status": "confirmed",
      "confirmations": 15,
      "block_number": 58234567,
      "from_address": "TMuA6YqfCeX8EhbfYEg5y7S4DqzSJireY9",
      "detected_at": "2025-01-15T10:15:00Z",
      "confirmed_at": "2025-01-15T10:18:00Z"
    }
  ],
  "audit_log": [
    {
      "event": "invoice.created",
      "timestamp": "2025-01-15T10:00:00Z",
      "user_agent": "MyApp/1.0"
    },
    {
      "event": "payment.detected", 
      "timestamp": "2025-01-15T10:15:00Z",
      "data": {"payment_id": "pay_xyz789"}
    }
  ],
  "webhook_deliveries": [
    {
      "event": "invoice.paid",
      "url": "https://merchant.com/webhook",
      "status": "delivered",
      "attempts": 1,
      "delivered_at": "2025-01-15T10:18:30Z"
    }
  ]
}
```

### Cancel Invoice
```http
POST /v1/invoices/{invoice_id}/cancel
Authorization: Bearer sk_live_abc123...
```

**Response:**
```json
{
  "id": "inv_abc123",
  "status": "cancelled",
  "cancelled_at": "2025-01-15T11:00:00Z"
}
```

### List Invoices
```http
GET /v1/invoices?status=pending&limit=50&offset=0&created_after=2025-01-01
Authorization: Bearer sk_live_abc123...
```

**Query Parameters:**
- `status` - Filter by status
- `limit` - Results per page (max 100)
- `offset` - Pagination offset
- `created_after` - ISO 8601 date
- `created_before` - ISO 8601 date
- `amount_gte` - Minimum amount filter
- `amount_lte` - Maximum amount filter

**Response:**
```json
{
  "invoices": [...],
  "pagination": {
    "total": 150,
    "limit": 50,
    "offset": 0,
    "has_more": true
  }
}
```

### Get Analytics
```http
GET /v1/analytics?period=30d&group_by=day
Authorization: Bearer sk_live_abc123...
```

**Response:**
```json
{
  "period": "30d",
  "metrics": {
    "total_invoices": 1250,
    "total_amount": 45678.90,
    "paid_invoices": 1100,
    "paid_amount": 42350.75,
    "conversion_rate": 88.0,
    "average_amount": 36.54
  },
  "time_series": [
    {
      "date": "2025-01-01",
      "invoices": 42,
      "amount": 1534.50,
      "paid": 38
    }
  ]
}
```

---

## Customer API Endpoints (Public)

### View Invoice (Customer)
```http
GET /invoice/{invoice_id}
```

**Response (public data only):**
```json
{
  "id": "inv_abc123",
  "title": "VPN Service Order",
  "description": "Monthly VPN subscription",
  "items": [
    {
      "name": "VPN Premium Plan",
      "description": "Monthly subscription with unlimited bandwidth",
      "quantity": 1,
      "unit_price": 9.99,
      "total": 9.99
    }
  ],
  "subtotal": 14.99,
  "tax": 1.50,
  "total": 16.49,
  "currency": "USD",
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
    "percent": 60.64
  }
}
```

### Check Payment Status (Customer)
```http
GET /invoice/{invoice_id}/status
```

**Response:**
```json
{
  "status": "partial",
  "payment_progress": {
    "received": 10.00,
    "required": 16.49,
    "percent": 60.64
  },
  "last_payment_at": "2025-01-15T10:15:00Z",
  "expires_at": "2025-01-15T10:30:00Z"
}
```

### Get QR Code
```http
GET /invoice/{invoice_id}/qr?size=256&format=png
```

**Query Parameters:**
- `size` - QR code size in pixels (128, 256, 512)
- `format` - Image format (png, svg)

**Response:** Image with QR code containing payment information

---

## Immutable Invoice Design

**Invoices are immutable once created**. No update operations are allowed to maintain audit integrity and prevent payment confusion.

**Allowed State Changes:**
- `pending → partial` (automatic on payment detection)
- `pending → confirming` (automatic on full payment)
- `partial → confirming` (automatic on completion)
- `confirming → paid` (automatic on confirmation)
- `pending → expired` (automatic on timeout)
- `pending/partial → cancelled` (manual via API)
- `paid → refunded` (manual process, tracked separately)

**If invoice details need changes:**
1. Cancel existing invoice
2. Create new invoice with correct details
3. Direct customer to new invoice URL

---

## Status Values

### Invoice Status
| Status       | Description                     | Customer Visible | Final |
| ------------ | ------------------------------- | ---------------- | ----- |
| `pending`    | Waiting for payment             | ✅                | ❌     |
| `partial`    | Partial payment received        | ✅                | ❌     |
| `confirming` | Payment detected, confirming    | ✅                | ❌     |
| `paid`       | Payment confirmed and complete  | ✅                | ✅     |
| `expired`    | Invoice expired without payment | ✅                | ✅     |
| `cancelled`  | Manually cancelled              | ✅                | ✅     |
| `refunded`   | Payment refunded                | ✅                | ✅     |

### Payment Status
| Status       | Description                  | Confirmations |
| ------------ | ---------------------------- | ------------- |
| `detected`   | Transaction found in mempool | 0             |
| `confirming` | Gaining confirmations        | 1-11          |
| `confirmed`  | Sufficient confirmations     | 12+           |
| `failed`     | Transaction failed           | N/A           |
| `orphaned`   | Block orphaned (temporary)   | N/A           |

---

## Webhooks

Configure webhook endpoints in your merchant dashboard for real-time event notifications.

### Webhook Events
- `invoice.created` - New invoice created
- `invoice.expired` - Invoice expired without payment
- `invoice.cancelled` - Invoice manually cancelled
- `payment.detected` - New payment detected (0 confirmations)
- `payment.confirmed` - Payment confirmed (sufficient confirmations)
- `invoice.paid` - Invoice fully paid and confirmed
- `invoice.refunded` - Invoice refunded

### Webhook Payload
```http
POST https://your-website.com/webhooks/crypto-checkout
Content-Type: application/json
X-Crypto-Checkout-Signature: sha256=abc123...
```

```json
{
  "id": "evt_abc123",
  "event": "invoice.paid",
  "created_at": "2025-01-15T10:18:00Z",
  "data": {
    "invoice_id": "inv_abc123",
    "status": "paid",
    "total_received": 16.49,
    "payment_id": "pay_xyz789"
  }
}
```

### Webhook Security
Verify webhooks using HMAC-SHA256 signature in the `X-Crypto-Checkout-Signature` header.

---

## Error Handling

### Error Response Format
```json
{
  "error": {
    "type": "validation_error",
    "code": "INVALID_AMOUNT", 
    "message": "Invoice amount must be greater than 0",
    "field": "items[0].unit_price",
    "documentation_url": "https://docs.crypto-checkout.com/errors#invalid-amount"
  },
  "request_id": "req_abc123"
}
```

### HTTP Status Codes
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict (e.g., already cancelled)
- `422` - Validation Error
- `429` - Rate Limited
- `500` - Internal Error

---

## Rate Limiting

### Merchant/Admin API Limits
- **Invoice creation**: 100/hour per API key
- **Invoice/payment status checks**: 1000/hour per API key
- **List operations**: 500/hour per API key
- **Payment searches**: 200/hour per API key
- **Refund operations**: 50/hour per API key
- **Analytics**: 100/hour per API key

### Customer API Limits
- **Invoice views**: 10/minute per IP
- **Status checks**: 60/minute per invoice
- **QR requests**: 20/minute per IP

### Rate Limit Headers
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642694400
Retry-After: 60
```

---

## Integration Examples

### Creating Invoice
```http
POST /v1/invoices
Authorization: Bearer sk_live_abc123...
Content-Type: application/json

{
  "title": "Premium Service",
  "items": [{"name": "Service", "quantity": 1, "unit_price": 99.99}],
  "currency": "USD"
}
```

### Customer Payment Flow
1. Customer receives invoice URL: `checkout.crypto-checkout.com/invoice/inv_abc123`
2. Customer visits URL → Frontend calls `GET /invoice/inv_abc123` for invoice details
3. Frontend displays payment information (amount, address, QR code)
4. Customer sends crypto payment to provided address
5. Frontend polls same endpoint every 10-15 seconds for payment updates
6. System detects payment → response shows updated `payments` array and `payment_progress`
7. System confirms payment → invoice `status` changes to `paid`
8. Merchant receives webhook notification

### Frontend Implementation Pattern
```javascript
// Initial page load
const invoice = await fetch('/invoice/inv_abc123').then(r => r.json());

// Status polling for real-time updates
const pollStatus = setInterval(async () => {
  const updated = await fetch('/invoice/inv_abc123').then(r => r.json());
  
  if (updated.status === 'paid') {
    clearInterval(pollStatus);
    showSuccessMessage();
  } else {
    updateProgressBar(updated.payment_progress);
    updatePaymentHistory(updated.payments);
  }
}, 10000); // Poll every 10 seconds
```

### Admin Dashboard Integration
- List all invoices with filtering/pagination
- View detailed invoice information including payments and audit logs
- Monitor webhook delivery status
- Access analytics and metrics
- Cancel invoices when necessary

This API design provides a solid foundation for building merchant integrations, admin interfaces, and customer payment experiences while maintaining data immutability and security.