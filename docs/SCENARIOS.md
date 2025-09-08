# Crypto Checkout User Scenarios & Domain Operations

- [Crypto Checkout User Scenarios \& Domain Operations](#crypto-checkout-user-scenarios--domain-operations)
  - [Merchant Scenarios](#merchant-scenarios)
    - [1. Merchant Onboarding](#1-merchant-onboarding)
    - [2. Adjust Platform Fee Rate](#2-adjust-platform-fee-rate)
    - [3. View Settlement Dashboard](#3-view-settlement-dashboard)
  - [Customer Payment Scenarios](#customer-payment-scenarios)
    - [4. Successful Payment with Real-time Settlement](#4-successful-payment-with-real-time-settlement)
    - [5. Partial Payment Handling](#5-partial-payment-handling)
  - [Developer Integration Scenarios](#developer-integration-scenarios)
    - [6. E-commerce Integration](#6-e-commerce-integration)
    - [7. Settlement Reporting API](#7-settlement-reporting-api)
  - [Platform Administrator Scenarios](#platform-administrator-scenarios)
    - [8. Revenue Analytics Dashboard](#8-revenue-analytics-dashboard)
    - [9. Failed Settlement Recovery](#9-failed-settlement-recovery)
  - [Error \& Edge Case Scenarios](#error--edge-case-scenarios)
    - [10. Exchange Rate Expiration](#10-exchange-rate-expiration)
    - [11. Settlement Failure Cascade](#11-settlement-failure-cascade)
  - [Summary](#summary)

---

## Merchant Scenarios

### 1. Merchant Onboarding

**Scenario**: New VPN service provider registers for crypto payment processing

**Steps:**
1. **Merchant visits platform** and clicks "Sign Up"
2. **Fills registration form** with business details
3. **System creates merchant account** with default 1% platform fee
4. **Initial API key generated** automatically
5. **Welcome email sent** with getting started guide

**Domain Operations:**
```go
// 1. Create merchant account
merchantService := NewMerchantOnboardingService()
merchant, err := merchantService.CreateMerchant(CreateMerchantRequest{
    BusinessName: "Acme VPN Services",
    ContactEmail: "admin@acmevpn.com",
    Settings: MerchantSettings{
        DefaultCurrency:        "USD",
        DefaultCryptoCurrency:  "USDT", 
        InvoiceExpiryMinutes:   30,
        PlatformFeePercentage:  decimal.NewFromFloat(1.0), // Default 1%
        PaymentTolerance: PaymentTolerance{
            UnderpaymentThreshold: decimal.NewFromFloat(0.01),
            OverpaymentThreshold:  decimal.NewFromFloat(1.00),
            OverpaymentAction:     "credit_account",
        },
    },
})

// 2. Generate initial API key
apiKey, err := merchantService.GenerateInitialApiKey(merchant, []string{
    "invoices:create", "invoices:read", "webhooks:manage",
})

// 3. Send welcome notification
result := merchantService.SendWelcomeNotification(merchant, apiKey)

// Events emitted:
// - MerchantCreated{MerchantID, Settings}
// - ApiKeyGenerated{ApiKeyID, MerchantID, Permissions}
```

### 2. Adjust Platform Fee Rate

**Scenario**: Platform admin negotiates custom fee rate with high-volume merchant

**Steps:**
1. **Admin reviews merchant volume** (>$100k/month)
2. **Negotiates reduced rate** from 1% to 0.7%
3. **Updates merchant settings** in admin panel
4. **Merchant receives notification** of new rate
5. **New rate applies to future settlements**

**Domain Operations:**
```go
// Update merchant platform fee
merchantRepo := NewMerchantRepository()
merchant, err := merchantRepo.FindByID(merchantID)

// Validate new fee rate (admin can set below 0.1% minimum)
newSettings := merchant.Settings
newSettings.PlatformFeePercentage = decimal.NewFromFloat(0.7) // Custom rate

// Apply business rules validation
if newSettings.PlatformFeePercentage.GreaterThan(decimal.NewFromFloat(5.0)) {
    return ErrPlatformFeeTooHigh
}

// Update merchant
merchant.UpdateSettings(newSettings)
merchantRepo.Save(merchant)

// Events emitted:
// - SettingsUpdated{MerchantID, UpdatedSettings}
```

### 3. View Settlement Dashboard

**Scenario**: Merchant checks earnings and fee breakdown

**Steps:**
1. **Merchant logs into dashboard**
2. **Views settlements page** with fee breakdown
3. **Sees gross/fee/net amounts** for each payment
4. **Downloads settlement report** for accounting
5. **Reviews platform fee percentage**

**Domain Operations:**
```go
// Get merchant settlements with fee breakdown
settlementRepo := NewSettlementRepository()
settlements, err := settlementRepo.FindByMerchant(merchantID, 
    DateRange{Start: startDate, End: endDate})

// Each settlement shows:
type SettlementSummary struct {
    InvoiceID         string          `json:"invoice_id"`
    GrossAmount       money.Money     `json:"gross_amount"`     // $9.99
    PlatformFeeAmount money.Money     `json:"platform_fee"`    // $0.10
    NetAmount         money.Money     `json:"net_amount"`      // $9.89
    FeePercentage     decimal.Decimal `json:"fee_percentage"`  // 1.0%
    SettledAt         time.Time       `json:"settled_at"`
}

// Calculate totals for dashboard
totalGross := decimal.Zero
totalFees := decimal.Zero
totalNet := decimal.Zero

for _, settlement := range settlements {
    totalGross = totalGross.Add(settlement.GrossAmount.Amount())
    totalFees = totalFees.Add(settlement.PlatformFeeAmount.Amount())
    totalNet = totalNet.Add(settlement.NetAmount.Amount())
}
```

---

## Customer Payment Scenarios

### 4. Successful Payment with Real-time Settlement

**Scenario**: Customer pays $9.99 for VPN subscription, settlement processed immediately

**Steps:**
1. **Customer visits payment page** from merchant redirect
2. **Scans QR code** and sends $9.99 USDT
3. **Payment detected** on blockchain
4. **Payment confirmed** after sufficient confirmations
5. **Settlement created** automatically
6. **Platform deducts 1% fee** ($0.10)
7. **Merchant receives $9.89** immediately
8. **Customer redirected** to success page

**Domain Operations:**
```go
// 1. Payment confirmation triggers settlement
paymentService := NewPaymentService()
settlementService := NewSettlementService()

// When payment is confirmed
payment := &Payment{
    InvoiceID: invoiceID,
    Amount:    money.New(999, "USDT"), // $9.99
    Status:    PaymentStatusConfirmed,
}

// Get invoice and merchant for settlement
invoice, _ := invoiceRepo.FindByID(payment.InvoiceID)
merchant, _ := merchantRepo.FindByID(invoice.MerchantID)

// 2. Create settlement automatically
settlement, err := settlementService.CreateSettlement(
    invoice, payment, merchant)

// Settlement calculation:
// GrossAmount = $9.99 (customer payment)
// PlatformFeeAmount = $9.99 × 1% = $0.10
// NetAmount = $9.99 - $0.10 = $9.89

type Settlement struct {
    ID                SettlementID    
    InvoiceID         InvoiceID       
    MerchantID        MerchantID      
    PaymentID         PaymentID       
    GrossAmount       money.Money     // $9.99
    PlatformFeeAmount money.Money     // $0.10 (1% of $9.99)
    NetAmount         money.Money     // $9.89
    FeePercentage     decimal.Decimal // 1.0% (from merchant settings)
    Status            SettlementStatusCompleted
    SettledAt         time.Time       // Real-time
}

// 3. Process real-time payout to merchant
payoutResult, err := settlementService.ProcessPayout(settlement)

// Events emitted:
// - PaymentConfirmed{PaymentID, Amount}
// - InvoicePaid{InvoiceID, TotalReceived}
// - SettlementCreated{SettlementID, GrossAmount}
// - PlatformFeeCollected{SettlementID, PlatformFeeAmount, FeeRate}
// - SettlementCompleted{SettlementID, NetAmount, SettledAt}
```

### 5. Partial Payment Handling

**Scenario**: Customer sends $5.00 instead of $9.99, then completes payment

**Steps:**
1. **First payment detected** ($5.00 USDT)
2. **Invoice status updated** to "partial"
3. **Customer sees progress** (50% paid)
4. **Second payment sent** ($4.99 USDT)
5. **Total payment confirmed** ($9.99 total)
6. **Settlement processed** for full amount

**Domain Operations:**
```go
// 1. First partial payment
payment1 := &Payment{
    InvoiceID: invoiceID,
    Amount:    money.New(500, "USDT"), // $5.00
    Status:    PaymentStatusConfirmed,
}

// Update invoice status
invoice.Status = InvoiceStatusPartial
invoice.UpdatePaymentProgress()

// 2. Second payment completes the invoice
payment2 := &Payment{
    InvoiceID: invoiceID,
    Amount:    money.New(499, "USDT"), // $4.99
    Status:    PaymentStatusConfirmed,
}

// Check if invoice is now fully paid
totalReceived := payment1.Amount.Add(payment2.Amount) // $9.99
if totalReceived.GreaterThanOrEqual(invoice.Total) {
    invoice.Status = InvoiceStatusPaid
    invoice.PaidAt = time.Now()
    
    // Create settlement for total amount
    settlement := settlementService.CreateSettlement(
        invoice, payment2, merchant) // Uses total received amount
    
    // Settlement calculation on full $9.99:
    // GrossAmount = $9.99 (total received)
    // PlatformFeeAmount = $0.10 (1% fee)
    // NetAmount = $9.89 (merchant receives)
}

// Events emitted:
// - PaymentDetected{PaymentID: payment1.ID, Amount: $5.00}
// - InvoicePartiallyPaid{InvoiceID, PartialAmount: $5.00}
// - PaymentDetected{PaymentID: payment2.ID, Amount: $4.99}  
// - InvoicePaid{InvoiceID, TotalReceived: $9.99}
// - SettlementCreated{SettlementID, GrossAmount: $9.99}
// - SettlementCompleted{SettlementID, NetAmount: $9.89}
```

---

## Developer Integration Scenarios

### 6. E-commerce Integration

**Scenario**: VPN service integrates payment processing into their checkout flow

**Steps:**
1. **Customer completes VPN signup** on merchant site
2. **Merchant backend calls API** to create invoice
3. **Customer redirected** to payment page
4. **Payment processed** and settled
5. **Webhook delivered** to merchant
6. **VPN account activated** automatically

**Domain Operations:**
```go
// 1. Merchant creates invoice via API
invoiceService := NewInvoiceService()
exchangeRateService := NewExchangeRateService()

// Lock exchange rate for 30 minutes
exchangeRate, err := exchangeRateService.GetLockedRate(
    "USD", "USDT", 30*time.Minute)

// Create invoice
createRequest := CreateInvoiceRequest{
    MerchantID:  merchantID,
    Title:       "VPN Premium Subscription",
    Description: "Monthly VPN service with unlimited bandwidth",
    Items: []InvoiceItem{
        {
            Name:        "VPN Premium Plan",
            Description: "Monthly subscription",
            Quantity:    1,
            UnitPrice:   money.New(999, "USD"), // $9.99
        },
    },
    Currency:         "USD",
    CryptoCurrency:   "USDT",
    ExchangeRate:     exchangeRate,
    PaymentTolerance: merchant.Settings.PaymentTolerance,
    ExpiresAt:        time.Now().Add(30 * time.Minute),
    Metadata: map[string]string{
        "customer_id": "cust_123",
        "order_id":    "order_456",
    },
}

invoice, err := invoiceService.CreateInvoice(createRequest)

// 2. When payment confirmed, webhook delivered
webhookService := NewWebhookDeliveryService()
settlement, _ := settlementRepo.FindByInvoiceID(invoice.ID)

webhookPayload := WebhookPayload{
    Event: "invoice.paid",
    Data: WebhookData{
        InvoiceID:         invoice.ID,
        Status:           "paid",
        GrossAmount:       settlement.GrossAmount,      // $9.99
        PlatformFeeAmount: settlement.PlatformFeeAmount, // $0.10
        NetAmount:         settlement.NetAmount,        // $9.89
        SettlementID:     settlement.ID,
        Metadata:         invoice.Metadata,
    },
}

// Deliver webhook to merchant
deliveryResult := webhookService.DeliverWebhook(
    merchant.WebhookEndpoints[0], webhookPayload)

// Events emitted:
// - InvoiceCreated{InvoiceID, MerchantID, Amount}
// - WebhookDelivered{WebhookEndpointID, Event, Status}
```

### 7. Settlement Reporting API

**Scenario**: Merchant queries settlement data for accounting integration

**Steps:**
1. **Accounting system calls API** monthly
2. **Requests settlement data** with date range
3. **Receives detailed breakdown** of all transactions
4. **Processes fee information** for tax reporting
5. **Reconciles payments** with gross/net amounts

**Domain Operations:**
```go
// API endpoint: GET /api/v1/settlements
type GetSettlementsRequest struct {
    MerchantID string    `json:"merchant_id"`
    StartDate  time.Time `json:"start_date"`
    EndDate    time.Time `json:"end_date"`
    Limit      int       `json:"limit"`
    Cursor     string    `json:"cursor"`
}

type SettlementResponse struct {
    Settlements []SettlementDetail `json:"settlements"`
    Summary     SettlementSummary  `json:"summary"`
    Pagination  PaginationInfo     `json:"pagination"`
}

type SettlementDetail struct {
    ID                string          `json:"id"`
    InvoiceID         string          `json:"invoice_id"`
    GrossAmount       money.Money     `json:"gross_amount"`
    PlatformFeeAmount money.Money     `json:"platform_fee_amount"`
    NetAmount         money.Money     `json:"net_amount"`
    FeePercentage     decimal.Decimal `json:"fee_percentage"`
    Currency          string          `json:"currency"`
    SettledAt         time.Time       `json:"settled_at"`
    Metadata          map[string]string `json:"metadata"`
}

// Calculate summary for accounting
type SettlementSummary struct {
    TotalGrossAmount       money.Money     `json:"total_gross_amount"`     // $2,500.00
    TotalPlatformFees      money.Money     `json:"total_platform_fees"`   // $25.00 (1%)
    TotalNetAmount         money.Money     `json:"total_net_amount"`       // $2,475.00
    AverageFeePercentage   decimal.Decimal `json:"average_fee_percentage"` // 1.0%
    TransactionCount       int             `json:"transaction_count"`      // 250
}

// Repository query
settlementRepo := NewSettlementRepository()
settlements, err := settlementRepo.FindByMerchantAndDateRange(
    request.MerchantID, request.StartDate, request.EndDate)
```

---

## Platform Administrator Scenarios

### 8. Revenue Analytics Dashboard

**Scenario**: Platform admin monitors revenue and merchant performance

**Steps:**
1. **Admin accesses analytics** dashboard
2. **Views platform fee collection** by time period
3. **Analyzes merchant volume** and fee rates
4. **Identifies high-volume merchants** for rate negotiation
5. **Tracks settlement success rates**

**Domain Operations:**
```go
// Platform revenue analytics
type PlatformAnalytics struct {
    Period                 TimePeriod      `json:"period"`
    TotalGrossVolume       money.Money     `json:"total_gross_volume"`     // $1M
    TotalPlatformRevenue   money.Money     `json:"total_platform_revenue"` // $10K (1%)
    TotalMerchantPayouts   money.Money     `json:"total_merchant_payouts"` // $990K
    AverageFeeRate         decimal.Decimal `json:"average_fee_rate"`       // 1.0%
    TransactionCount       int             `json:"transaction_count"`      // 100,000
    ActiveMerchants        int             `json:"active_merchants"`       // 1,250
    SettlementSuccessRate  decimal.Decimal `json:"settlement_success_rate"` // 99.8%
}

// High-volume merchant analysis
type MerchantVolumeReport struct {
    MerchantID            string          `json:"merchant_id"`
    BusinessName          string          `json:"business_name"`
    MonthlyGrossVolume    money.Money     `json:"monthly_gross_volume"`    // $150K
    MonthlyPlatformRevenue money.Money    `json:"monthly_platform_revenue"` // $1.5K
    CurrentFeeRate        decimal.Decimal `json:"current_fee_rate"`        // 1.0%
    SuggestedFeeRate      decimal.Decimal `json:"suggested_fee_rate"`      // 0.8%
    TransactionCount      int             `json:"transaction_count"`       // 15,000
}

// Query high-volume merchants for rate negotiation
analyticsService := NewAnalyticsService()
highVolumeReport, err := analyticsService.GetHighVolumeMerchants(
    MinMonthlyVolume: money.New(100000, "USD"), // $100K+
    Period: LastMonth,
)
```

### 9. Failed Settlement Recovery

**Scenario**: Admin handles failed settlement and retry processing

**Steps:**
1. **Settlement fails** due to wallet issue
2. **Alert sent** to operations team
3. **Admin investigates** payout failure
4. **Updates merchant wallet** address
5. **Retries settlement** manually

**Domain Operations:**
```go
// Handle failed settlement
settlementService := NewSettlementService()
settlement, err := settlementRepo.FindByID(settlementID)

if settlement.Status == SettlementStatusFailed {
    // Update merchant payout details
    merchant, _ := merchantRepo.FindByID(settlement.MerchantID)
    merchant.UpdatePayoutAddress(newWalletAddress)
    merchantRepo.Save(merchant)
    
    // Retry settlement with updated details
    retryResult, err := settlementService.RetryFailedPayout(settlement)
    
    if retryResult.Success {
        settlement.Status = SettlementStatusCompleted
        settlement.SettledAt = time.Now()
        settlementRepo.Save(settlement)
        
        // Events emitted:
        // - SettlementCompleted{SettlementID, NetAmount, SettledAt}
    } else {
        // Log failure and alert operations
        settlement.Status = SettlementStatusFailed
        settlement.FailureReason = retryResult.Error
        
        // Events emitted:
        // - SettlementFailed{SettlementID, FailureReason}
    }
}
```

---

## Error & Edge Case Scenarios

### 10. Exchange Rate Expiration

**Scenario**: Exchange rate expires during payment process

**Steps:**
1. **Invoice created** with 30-minute rate lock
2. **Customer delays payment** beyond expiry
3. **Rate expires** while payment in transit
4. **System detects expired rate**
5. **Payment processed** with tolerance check

**Domain Operations:**
```go
// Check exchange rate validity during payment processing
exchangeRateService := NewExchangeRateService()
payment := &Payment{
    InvoiceID: invoiceID,
    Amount:    money.New(999, "USDT"),
}

invoice, _ := invoiceRepo.FindByID(payment.InvoiceID)

// Validate exchange rate hasn't expired
if !exchangeRateService.ValidateRate(invoice.ExchangeRate) {
    // Rate expired - calculate with current rate and tolerance
    currentRate, _ := exchangeRateService.GetCurrentRate("USD", "USDT")
    
    expectedAmount := exchangeRateService.ConvertAmount(
        invoice.Total, currentRate)
    
    tolerance := invoice.PaymentTolerance
    minAcceptable := expectedAmount.Subtract(tolerance.UnderpaymentThreshold)
    maxAcceptable := expectedAmount.Add(tolerance.OverpaymentThreshold)
    
    if payment.Amount.LessThan(minAcceptable) {
        return ErrUnderpayment
    }
    
    if payment.Amount.GreaterThan(maxAcceptable) {
        // Handle overpayment based on policy
        reconciliationService := NewPaymentReconciliationService()
        result := reconciliationService.HandleOverpayment(
            payment.Amount.Subtract(expectedAmount),
            tolerance.OverpaymentAction)
    }
}

// Events emitted:
// - ExchangeRateExpired{InvoiceID, ExpiredRate, CurrentRate}
// - PaymentReconciled{PaymentID, ExpectedAmount, ActualAmount}
```

### 11. Settlement Failure Cascade

**Scenario**: Multiple settlements fail due to platform wallet issue

**Steps:**
1. **Platform hot wallet** runs out of funds
2. **Multiple settlements fail** simultaneously
3. **Monitoring alerts** operations team
4. **Hot wallet refilled** from cold storage
5. **Failed settlements retried** in batch

**Domain Operations:**
```go
// Batch retry failed settlements
type FailedSettlementBatch struct {
    FailedSettlements []Settlement
    TotalAmount       money.Money
    FailureReason     string
    RetryAttempts     int
}

// Query all failed settlements
failedSettlements, err := settlementRepo.FindByStatus(
    SettlementStatusFailed)

// Group by failure reason
hotWalletFailures := filterByFailureReason(failedSettlements, 
    "insufficient_wallet_balance")

// After hot wallet refill, retry batch
batchRetryService := NewBatchRetryService()
for _, settlement := range hotWalletFailures {
    retryResult := settlementService.RetryFailedPayout(settlement)
    
    if retryResult.Success {
        settlement.Status = SettlementStatusCompleted
        settlement.SettledAt = time.Now()
        
        // Events emitted:
        // - SettlementCompleted{SettlementID, NetAmount}
    } else {
        settlement.RetryAttempts++
        if settlement.RetryAttempts >= MaxRetryAttempts {
            // Escalate to manual review
            // Events emitted:
            // - SettlementRequiresManualReview{SettlementID}
        }
    }
}

// Events emitted:
// - BatchSettlementRetryCompleted{BatchID, SuccessCount, FailureCount}
```

---

## Summary

This comprehensive set of user scenarios demonstrates how the Crypto Checkout platform operates as a payment processor, with automatic fee deduction and real-time settlement. Key characteristics:

**Payment Processor Model:**
- Platform automatically deducts configurable fees (default 1%)
- Real-time settlement to merchants
- Transparent fee breakdown in merchant dashboards
- Proper revenue tracking for platform operations

**Domain Operation Patterns:**
- Event-driven architecture with comprehensive audit trails
- Settlement aggregate manages fee calculation and payouts
- Real-time processing with robust error handling
- Merchant-specific fee configuration and reporting

**Business Logic Enforcement:**
- Platform fee validation (0.1% - 5.0% range)
- Real-time settlement processing
- Automatic fee calculation and deduction
- Comprehensive settlement reporting and analytics

The scenarios cover the complete lifecycle from merchant onboarding through payment processing, settlement, and revenue analytics, demonstrating how the domain model supports a robust payment processor platform.