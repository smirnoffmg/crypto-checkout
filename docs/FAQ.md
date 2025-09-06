# Crypto Checkout - Frequently Asked Questions

- [Crypto Checkout - Frequently Asked Questions](#crypto-checkout---frequently-asked-questions)
  - [For Customers (Paying invoices)](#for-customers-paying-invoices)
    - [How do I pay with cryptocurrency?](#how-do-i-pay-with-cryptocurrency)
    - [What cryptocurrencies are accepted?](#what-cryptocurrencies-are-accepted)
    - [What happens if I send the wrong amount?](#what-happens-if-i-send-the-wrong-amount)
    - [How long does payment confirmation take?](#how-long-does-payment-confirmation-take)
    - [Can I get a refund if I overpaid?](#can-i-get-a-refund-if-i-overpaid)
    - [What if my payment doesn't appear after sending?](#what-if-my-payment-doesnt-appear-after-sending)
    - [Is it safe to pay? How is my privacy protected?](#is-it-safe-to-pay-how-is-my-privacy-protected)
    - [Do I need to create an account to pay?](#do-i-need-to-create-an-account-to-pay)
    - [Can I pay from any wallet or exchange?](#can-i-pay-from-any-wallet-or-exchange)
    - [What happens if the invoice expires while I'm paying?](#what-happens-if-the-invoice-expires-while-im-paying)
  - [For Merchants (Integrating the system)](#for-merchants-integrating-the-system)
    - [How do I integrate this with my website or application?](#how-do-i-integrate-this-with-my-website-or-application)
    - [What are the transaction fees?](#what-are-the-transaction-fees)
    - [How do I receive payment notifications (webhooks)?](#how-do-i-receive-payment-notifications-webhooks)
    - [Can I customize the invoice appearance?](#can-i-customize-the-invoice-appearance)
    - [How do I handle partial payments?](#how-do-i-handle-partial-payments)
    - [What happens to the crypto after customers pay?](#what-happens-to-the-crypto-after-customers-pay)
    - [How do I issue refunds?](#how-do-i-issue-refunds)
    - [Can I set custom expiration times for invoices?](#can-i-set-custom-expiration-times-for-invoices)
    - [Is there a test environment available?](#is-there-a-test-environment-available)
    - [What compliance/regulatory requirements should I know about?](#what-complianceregulatory-requirements-should-i-know-about)
  - [For Developers/Administrators](#for-developersadministrators)
    - [What are the minimum system requirements?](#what-are-the-minimum-system-requirements)
    - [How do I deploy and configure the system?](#how-do-i-deploy-and-configure-the-system)
    - [What databases are supported?](#what-databases-are-supported)
    - [How do I enable Kafka audit logging?](#how-do-i-enable-kafka-audit-logging)
    - [How do I monitor system health and performance?](#how-do-i-monitor-system-health-and-performance)
    - [What backup strategy should I implement?](#what-backup-strategy-should-i-implement)
    - [How do I secure the private keys?](#how-do-i-secure-the-private-keys)
    - [Can I run this behind a load balancer?](#can-i-run-this-behind-a-load-balancer)
    - [What are the rate limits for the API?](#what-are-the-rate-limits-for-the-api)
    - [How do I upgrade to newer versions?](#how-do-i-upgrade-to-newer-versions)
  - [Technical/Security Questions](#technicalsecurity-questions)
    - [How are payment addresses generated?](#how-are-payment-addresses-generated)
    - [What happens if the blockchain RPC goes down?](#what-happens-if-the-blockchain-rpc-goes-down)
    - [How do you prevent double-spending attacks?](#how-do-you-prevent-double-spending-attacks)
    - [Is the system PCI DSS compliant?](#is-the-system-pci-dss-compliant)
    - [How do you handle blockchain reorganizations?](#how-do-you-handle-blockchain-reorganizations)
    - [What encryption is used for sensitive data?](#what-encryption-is-used-for-sensitive-data)
    - [How are private keys managed and secured?](#how-are-private-keys-managed-and-secured)
    - [Can I audit the transaction history?](#can-i-audit-the-transaction-history)

## For Customers (Paying invoices)

### How do I pay with cryptocurrency?
1. Open the invoice link you received
2. Scan the QR code with your crypto wallet or copy the payment address
3. Send the exact USDT amount shown on the invoice
4. Wait for blockchain confirmation (usually 1-3 minutes)
5. The invoice will automatically update to "Paid" status

### What cryptocurrencies are accepted?
Currently, we only accept **USDT (Tether)** on the **Tron network (TRC-20)**. This ensures fast transactions with minimal fees (typically under $1).

### What happens if I send the wrong amount?
- **Underpayment**: Invoice remains unpaid. Contact the merchant for partial payment handling
- **Overpayment**: You'll receive credit for the full amount sent. Contact the merchant for refund of the excess
- **Wrong token**: Funds may be lost. Only send TRC-20 USDT to the provided address

### How long does payment confirmation take?
- **Typical time**: 1-3 minutes after sending
- **Network congestion**: Up to 10 minutes in rare cases  
- **Confirmation blocks**: Payment is confirmed after 1 block (about 15 seconds on Tron)

### Can I get a refund if I overpaid?
Yes, but refunds are handled by the merchant, not the payment system. Contact the merchant directly with your transaction hash and overpayment details.

### What if my payment doesn't appear after sending?
1. Check your transaction hash on [Tronscan](https://tronscan.org)
2. Verify you sent TRC-20 USDT (not other networks)
3. Confirm you sent to the exact address shown
4. Wait up to 10 minutes for network confirmation
5. If still not showing, contact the merchant with your transaction hash

### Is it safe to pay? How is my privacy protected?
- **Address privacy**: Each invoice gets a unique payment address
- **No personal data**: No registration or personal information required
- **Blockchain security**: Payments are secured by Tron blockchain cryptography
- **No logs**: The system follows a no-logs policy for maximum privacy

### Do I need to create an account to pay?
No. Simply open the invoice link and pay directly. No registration, email, or personal information required.

### Can I pay from any wallet or exchange?
- **✅ Recommended**: Any personal wallet (TronLink, Trust Wallet, etc.)
- **⚠️ Caution**: Some exchanges don't support direct payments to external addresses
- **❌ Avoid**: Using exchange internal transfers (they may not appear)

### What happens if the invoice expires while I'm paying?
If you send payment after expiration, your funds are still safe. Contact the merchant immediately with your transaction hash - they can manually verify and process your payment.

---

## For Merchants (Integrating the system)

### How do I integrate this with my website or application?
Use our REST API:
```bash
# Create invoice
POST /api/invoices
{
  "title": "Order #1234",
  "items": [{"name": "Product", "quantity": 1, "unit_price": 99.99}],
  "currency": "USD"
}

# Check status
GET /api/invoices/{id}/status
```
Redirect customers to the returned invoice URL.

### What are the transaction fees?
- **System fees**: None - we don't charge processing fees
- **Network fees**: ~$0.50-$1.00 paid by the customer to Tron network
- **Your costs**: Only infrastructure costs (servers, domains)

### How do I receive payment notifications (webhooks)?
Configure webhook endpoints in your settings:
```json
POST https://your-site.com/webhooks/payment
{
  "invoice_id": "inv_123",
  "status": "paid",
  "amount": "99.99",
  "tx_hash": "0x..."
}
```

### Can I customize the invoice appearance?
Currently, invoices use a standard template. Custom branding and themes are planned for future releases.

### How do I handle partial payments?
The system detects partial payments but marks invoices as "underpaid." You'll need to handle these manually through your business logic.

### What happens to the crypto after customers pay?
Funds are automatically swept from payment addresses to your configured master wallet within 10 minutes of confirmation.

### How do I issue refunds?
Refunds must be processed manually from your master wallet. The system provides transaction hashes and amounts for your records.

### Can I set custom expiration times for invoices?
Yes, configure default expiration in your settings (default: 30 minutes). Future API versions will support per-invoice expiration times.

### Is there a test environment available?
Yes, configure the system to use Tron testnet (Shasta) for testing. Use testnet USDT for integration testing.

### What compliance/regulatory requirements should I know about?
- **KYC/AML**: Not required by the system, but check your local regulations
- **Tax reporting**: You're responsible for reporting crypto income
- **Data privacy**: System is GDPR-compliant with minimal data collection
- **Financial licensing**: Consult local authorities about payment processor licensing

---

## For Developers/Administrators

### What are the minimum system requirements?
- **CPU**: 1 vCPU (2+ recommended)
- **RAM**: 1GB (2GB+ recommended) 
- **Storage**: 10GB SSD
- **Network**: Stable internet for blockchain RPC calls
- **OS**: Linux (Ubuntu 22.04+ recommended)

### How do I deploy and configure the system?
```bash
# Docker deployment (recommended)
docker run -d \
  --name crypto-checkout \
  -p 8080:8080 \
  -e DATABASE_URL="postgres://..." \
  -e TRON_RPC_URL="https://api.trongrid.io" \
  -e WALLET_SEED="your-master-seed" \
  crypto-checkout:latest

# Or compile from source
go build -o crypto-checkout ./cmd/server
./crypto-checkout
```

### What databases are supported?
- **PostgreSQL** (recommended for production)
- **SQLite** (suitable for development/small deployments)
- **MySQL** (supported but not recommended)

### How do I enable Kafka audit logging?
Add Kafka configuration to your config file:
```yaml
kafka:
  enabled: true
  brokers: ["localhost:9092"]
  topic: "payment-events"
```

### How do I monitor system health and performance?
- **Health endpoint**: `GET /health`
- **Metrics**: Prometheus metrics at `/metrics`
- **Logs**: Structured JSON logs to stdout
- **Monitoring**: Integrate with Grafana + Prometheus

### What backup strategy should I implement?
- **Database**: Daily automated backups with point-in-time recovery
- **Wallet seed**: Secure offline backup (paper wallet recommended)
- **Configuration**: Version control your config files
- **Logs**: Rotate and archive logs regularly

### How do I secure the private keys?
- **Master seed**: Store encrypted in environment variables
- **Auto-sweep**: Keys are derived on-demand and immediately discarded  
- **HSM integration**: Hardware Security Module support planned
- **Key rotation**: Not supported yet - planned for future releases

### Can I run this behind a load balancer?
Yes, the system is stateless and supports horizontal scaling:
- Use sticky sessions for invoice pages (optional)
- Shared database and Redis cache
- Load balance API endpoints normally

### What are the rate limits for the API?
Default rate limits:
- **Invoice creation**: 100 requests/hour per IP
- **Status checks**: 1000 requests/hour per IP
- **Webhook retries**: 10 attempts with exponential backoff

### How do I upgrade to newer versions?
```bash
# Docker deployment
docker pull crypto-checkout:latest
docker stop crypto-checkout
docker rm crypto-checkout
# Run with new image

# Database migrations run automatically on startup
```

---

## Technical/Security Questions

### How are payment addresses generated?
Using BIP44 HD (Hierarchical Deterministic) wallet derivation:
- **Path**: `m/44'/195'/0'/0/index` (195 = Tron coin type)
- **Unique**: Each invoice gets a fresh address
- **Deterministic**: Can be recreated from master seed

### What happens if the blockchain RPC goes down?
- **Primary**: System uses Trongrid API
- **Fallback**: Automatic failover to backup RPC endpoints
- **Circuit breaker**: Prevents cascading failures
- **Recovery**: Automatic retry with exponential backoff

### How do you prevent double-spending attacks?
- **Blockchain security**: Tron's consensus mechanism prevents double-spending
- **Confirmation blocks**: Wait for 1+ block confirmations
- **Unique addresses**: Each invoice uses a different address
- **Amount matching**: Only exact payment amounts are accepted

### Is the system PCI DSS compliant?
PCI DSS doesn't apply to cryptocurrency payments. However, we follow similar security practices:
- **Encryption**: All sensitive data encrypted at rest and in transit
- **Access control**: Role-based permissions
- **Monitoring**: Comprehensive audit logging
- **Network security**: TLS 1.3 for all communications

### How do you handle blockchain reorganizations?
- **Deep confirmations**: Wait for sufficient block confirmations
- **Reorg detection**: Monitor for chain reorganizations
- **Payment reversal**: Automatically handle reversed transactions
- **Alert system**: Notify administrators of reorg events

### What encryption is used for sensitive data?
- **Database**: AES-256 encryption at rest
- **Transit**: TLS 1.3 for all API communications  
- **Wallet seeds**: Encrypted with system key
- **Logs**: Sensitive data is masked/redacted

### How are private keys managed and secured?
- **Derivation**: Keys derived on-demand from master seed
- **Memory**: Private keys never stored in persistent memory
- **Auto-sweep**: Keys used once then immediately discarded
- **Master seed**: Encrypted and stored securely

### Can I audit the transaction history?
Yes, multiple audit options:
- **Database logs**: All invoice and payment events
- **Kafka audit log**: Immutable event stream (if enabled)
- **Blockchain**: All transactions are publicly verifiable on Tron
- **Export**: API endpoints for extracting audit data