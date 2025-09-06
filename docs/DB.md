# Crypto Checkout Database Schema Design

- [Crypto Checkout Database Schema Design](#crypto-checkout-database-schema-design)
  - [Architecture Principles](#architecture-principles)
  - [Database Schema](#database-schema)
  - [Detailed Index Strategy](#detailed-index-strategy)
    - [Event Store Indexes](#event-store-indexes)
    - [Invoice Read Model Indexes](#invoice-read-model-indexes)
    - [Analytics Indexes](#analytics-indexes)
  - [Partitioning Strategy](#partitioning-strategy)
    - [Time-Based Partitioning](#time-based-partitioning)
    - [Partition Implementation](#partition-implementation)
  - [Sharding Strategy](#sharding-strategy)
    - [Horizontal Sharding by Merchant](#horizontal-sharding-by-merchant)
    - [Sharding Configuration](#sharding-configuration)
    - [Shard Key Strategy](#shard-key-strategy)
  - [Read Replica Strategy](#read-replica-strategy)
    - [Read/Write Separation](#readwrite-separation)
    - [Replica Configuration](#replica-configuration)
  - [Performance Optimization](#performance-optimization)
    - [Query Performance Targets](#query-performance-targets)
    - [Caching Strategy](#caching-strategy)
    - [Connection Pooling](#connection-pooling)
  - [Monitoring and Maintenance](#monitoring-and-maintenance)
    - [Key Metrics to Monitor](#key-metrics-to-monitor)
    - [Automated Maintenance](#automated-maintenance)

## Architecture Principles

- **Event Sourcing**: Immutable event log as source of truth
- **CQRS**: Separate read/write models optimized for their use cases
- **Eventual Consistency**: Read models updated asynchronously from events
- **Horizontal Scaling**: Sharding strategies for high-volume tables
- **Performance**: Strategic indexing and partitioning

---

## Database Schema

```mermaid
erDiagram
    %% Event Store (Write Side)
    events {
        uuid id PK
        uuid aggregate_id
        varchar aggregate_type
        varchar event_type
        integer event_version
        jsonb event_data
        jsonb metadata
        timestamptz created_at
        bigserial sequence_number
    }
    
    aggregate_snapshots {
        uuid aggregate_id PK
        varchar aggregate_type
        integer version
        jsonb data
        timestamptz created_at
    }

    %% Merchant Aggregate Read Models
    merchants {
        uuid id PK
        varchar business_name
        varchar contact_email UK
        varchar status
        varchar plan_type
        jsonb settings
        timestamptz created_at
        timestamptz updated_at
    }
    
    api_keys {
        uuid id PK
        uuid merchant_id FK
        varchar key_hash UK
        varchar key_type
        text_array permissions
        varchar status
        varchar name
        timestamptz last_used_at
        timestamptz expires_at
        timestamptz created_at
    }
    
    webhook_endpoints {
        uuid id PK
        uuid merchant_id FK
        varchar url
        text_array events
        varchar secret
        varchar status
        integer max_retries
        varchar retry_backoff
        integer timeout_seconds
        text_array allowed_ips
        timestamptz created_at
    }

    %% Invoice Aggregate Read Models
    invoices {
        uuid id PK
        uuid merchant_id FK
        uuid customer_id FK
        varchar title
        text description
        jsonb items
        decimal subtotal
        decimal tax
        decimal total
        varchar currency
        varchar crypto_currency
        decimal crypto_amount
        varchar payment_address
        varchar status
        jsonb exchange_rate
        jsonb payment_tolerance
        timestamptz expires_at
        timestamptz created_at
        timestamptz updated_at
        timestamptz paid_at
    }
    
    payments {
        uuid id PK
        uuid invoice_id FK
        varchar tx_hash UK
        decimal amount
        varchar from_address
        varchar to_address
        varchar status
        integer confirmations
        integer required_confirmations
        bigint block_number
        varchar block_hash
        decimal network_fee
        timestamptz detected_at
        timestamptz confirmed_at
        timestamptz created_at
    }
    
    audit_entries {
        uuid id PK
        uuid invoice_id FK
        varchar event
        varchar actor
        timestamptz timestamp
        varchar ip_address
        varchar user_agent
        varchar request_id
        jsonb data
    }

    %% Customer Aggregate Read Models
    customers {
        uuid id PK
        uuid merchant_id FK
        varchar email
        varchar status
        jsonb metadata
        timestamptz created_at
        timestamptz updated_at
        timestamptz last_seen_at
    }
    
    payment_history {
        uuid customer_id FK
        uuid invoice_id FK
        decimal amount
        varchar status
        timestamptz paid_at
        timestamptz created_at
    }

    %% System Configuration Read Models
    system_configuration {
        uuid id PK
        jsonb exchange_rate_providers
        jsonb blockchain_networks
        jsonb payment_settings
        jsonb security_settings
        jsonb feature_flags
        timestamptz updated_at
    }

    %% Analytics Read Models
    merchant_analytics {
        uuid merchant_id PK
        varchar period
        jsonb metrics
        jsonb time_series_data
        timestamptz generated_at
    }
    
    daily_metrics {
        uuid merchant_id FK
        date metric_date
        integer invoices_created
        integer invoices_paid
        decimal amount_created
        decimal amount_paid
        decimal conversion_rate
    }

    %% Notification Templates Read Models
    notification_templates {
        uuid id PK
        uuid merchant_id FK
        varchar event_type
        varchar subject
        text body_template
        boolean is_active
        timestamptz created_at
    }
    
    notification_deliveries {
        uuid id PK
        uuid template_id FK
        uuid invoice_id FK
        varchar recipient
        varchar status
        varchar delivery_method
        integer attempts
        timestamptz delivered_at
        timestamptz created_at
    }

    %% Webhook Delivery Tracking
    webhook_deliveries {
        uuid id PK
        uuid webhook_endpoint_id FK
        uuid invoice_id FK
        varchar event_type
        varchar status
        integer attempts
        integer response_code
        text response_body
        timestamptz delivered_at
        timestamptz created_at
    }

    %% Relationships
    merchants ||--o{ api_keys : owns
    merchants ||--o{ webhook_endpoints : configures
    merchants ||--o{ invoices : creates
    merchants ||--o{ customers : serves
    merchants ||--|| merchant_analytics : generates
    merchants ||--o{ notification_templates : defines
    merchants ||--o{ daily_metrics : tracks
    
    customers ||--o{ invoices : pays
    customers ||--o{ payment_history : accumulates
    
    invoices ||--o{ payments : receives
    invoices ||--o{ audit_entries : logs
    invoices ||--o{ webhook_deliveries : triggers
    invoices ||--o{ notification_deliveries : sends
    
    webhook_endpoints ||--o{ webhook_deliveries : delivers
    notification_templates ||--o{ notification_deliveries : generates
```

---

## Detailed Index Strategy

### Event Store Indexes

| Index Name                    | Type   | Columns                         | Purpose                  | Performance Impact             |
| ----------------------------- | ------ | ------------------------------- | ------------------------ | ------------------------------ |
| `idx_events_aggregate_lookup` | B-Tree | `(aggregate_id, event_version)` | Aggregate reconstruction | Critical for write performance |
| `idx_events_type_timeline`    | B-Tree | `(aggregate_type, created_at)`  | Event replay by type     | Projection rebuilding          |
| `idx_events_sequence`         | B-Tree | `(sequence_number)`             | Global ordering          | Kafka offset tracking          |
| `idx_events_created_at`       | B-Tree | `(created_at)`                  | Time-based queries       | Analytics and reporting        |
| `idx_events_metadata_gin`     | GIN    | `(metadata)`                    | Event metadata search    | Debugging and auditing         |

```sql
-- Critical Event Store Indexes
CREATE INDEX CONCURRENTLY idx_events_aggregate_lookup 
    ON events (aggregate_id, event_version);

CREATE INDEX CONCURRENTLY idx_events_type_timeline 
    ON events (aggregate_type, created_at DESC);

CREATE INDEX CONCURRENTLY idx_events_sequence 
    ON events (sequence_number);

CREATE INDEX CONCURRENTLY idx_events_metadata_gin 
    ON events USING gin (metadata);
```

### Invoice Read Model Indexes

| Index Name                     | Type   | Columns                                                     | Purpose                 | Query Pattern                                     |
| ------------------------------ | ------ | ----------------------------------------------------------- | ----------------------- | ------------------------------------------------- |
| `idx_invoices_merchant_status` | B-Tree | `(merchant_id, status, created_at)`                         | Merchant dashboard      | `WHERE merchant_id = ? AND status = ?`            |
| `idx_invoices_expires_pending` | B-Tree | `(expires_at, status)`                                      | Expiration cleanup      | `WHERE expires_at < NOW() AND status = 'pending'` |
| `idx_invoices_payment_address` | B-Tree | `(payment_address)`                                         | Payment detection       | `WHERE payment_address = ?`                       |
| `idx_invoices_text_search`     | GIN    | `(to_tsvector('english', title \|\| ' ' \|\| description))` | Full-text search        | `WHERE search_vector @@ plainto_tsquery(?)`       |
| `idx_payments_tx_hash`         | B-Tree | `(tx_hash)`                                                 | Blockchain tracking     | `WHERE tx_hash = ?`                               |
| `idx_payments_status_pending`  | B-Tree | `(status, detected_at)`                                     | Confirmation monitoring | `WHERE status IN ('detected', 'confirming')`      |

```sql
-- Invoice Performance Indexes
CREATE INDEX CONCURRENTLY idx_invoices_merchant_status 
    ON invoices (merchant_id, status, created_at DESC);

CREATE INDEX CONCURRENTLY idx_invoices_expires_pending 
    ON invoices (expires_at) 
    WHERE status IN ('pending', 'partial');

CREATE INDEX CONCURRENTLY idx_invoices_payment_address 
    ON invoices (payment_address);

-- Full-text search
CREATE INDEX CONCURRENTLY idx_invoices_text_search 
    ON invoices USING gin (to_tsvector('english', title || ' ' || coalesce(description, '')));

-- Payment tracking indexes
CREATE UNIQUE INDEX CONCURRENTLY idx_payments_tx_hash 
    ON payments (tx_hash);

CREATE INDEX CONCURRENTLY idx_payments_status_pending 
    ON payments (status, detected_at) 
    WHERE status IN ('detected', 'confirming');
```

### Analytics Indexes

| Index Name                        | Type   | Columns                      | Purpose             | Aggregation Type         |
| --------------------------------- | ------ | ---------------------------- | ------------------- | ------------------------ |
| `idx_daily_metrics_merchant_date` | B-Tree | `(merchant_id, metric_date)` | Merchant timeseries | Revenue trends           |
| `idx_daily_metrics_date_range`    | B-Tree | `(metric_date, merchant_id)` | Platform analytics  | Cross-merchant reporting |
| `idx_audit_entries_timeline`      | B-Tree | `(invoice_id, timestamp)`    | Audit trail         | Compliance queries       |
| `idx_webhook_deliveries_status`   | B-Tree | `(status, created_at)`       | Failure monitoring  | Retry processing         |

```sql
-- Analytics Performance Indexes
CREATE INDEX CONCURRENTLY idx_daily_metrics_merchant_date 
    ON daily_metrics (merchant_id, metric_date DESC);

CREATE INDEX CONCURRENTLY idx_daily_metrics_date_range 
    ON daily_metrics (metric_date DESC, merchant_id);

-- Audit and monitoring indexes
CREATE INDEX CONCURRENTLY idx_audit_entries_timeline 
    ON audit_entries (invoice_id, timestamp DESC);

CREATE INDEX CONCURRENTLY idx_webhook_deliveries_status 
    ON webhook_deliveries (status, created_at) 
    WHERE status IN ('failed', 'pending');
```

---

## Partitioning Strategy

### Time-Based Partitioning

```mermaid
graph TB
    subgraph "Events Table Partitioning"
        events[events] --> events_2025_01[events_2025_01]
        events --> events_2025_02[events_2025_02]
        events --> events_2025_03[events_2025_03]
        events --> events_2025_04[events_2025_04]
    end
    
    subgraph "Audit Entries Partitioning"
        audit_entries[audit_entries] --> audit_2025_01[audit_2025_01]
        audit_entries --> audit_2025_02[audit_2025_02]
        audit_entries --> audit_2025_03[audit_2025_03]
    end
    
    subgraph "Daily Metrics Partitioning"
        daily_metrics[daily_metrics] --> metrics_2025_q1[metrics_2025_q1]
        daily_metrics --> metrics_2025_q2[metrics_2025_q2]
        daily_metrics --> metrics_2025_q3[metrics_2025_q3]
    end
```

### Partition Implementation

```sql
-- Events table partitioning by month
CREATE TABLE events (
    id UUID DEFAULT gen_random_uuid(),
    aggregate_id UUID NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_version INTEGER NOT NULL,
    event_data JSONB NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sequence_number BIGSERIAL,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Monthly partitions
CREATE TABLE events_2025_01 PARTITION OF events
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE events_2025_02 PARTITION OF events
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

-- Quarterly partitions for metrics
CREATE TABLE daily_metrics (
    merchant_id UUID NOT NULL,
    metric_date DATE NOT NULL,
    invoices_created INTEGER DEFAULT 0,
    invoices_paid INTEGER DEFAULT 0,
    amount_created DECIMAL(15,2) DEFAULT 0,
    amount_paid DECIMAL(15,2) DEFAULT 0,
    conversion_rate DECIMAL(5,2) DEFAULT 0,
    PRIMARY KEY (merchant_id, metric_date)
) PARTITION BY RANGE (metric_date);

CREATE TABLE daily_metrics_2025_q1 PARTITION OF daily_metrics
    FOR VALUES FROM ('2025-01-01') TO ('2025-04-01');
```

---

## Sharding Strategy

### Horizontal Sharding by Merchant

```mermaid
graph TB
    subgraph "Application Layer"
        router[Shard Router]
        app1[App Server 1]
        app2[App Server 2]
        app3[App Server 3]
    end
    
    subgraph "Database Shards"
        shard1[Shard 1<br/>Merchants 0-999]
        shard2[Shard 2<br/>Merchants 1000-1999]
        shard3[Shard 3<br/>Merchants 2000-2999]
        shard4[Shard 4<br/>Merchants 3000+]
    end
    
    router --> shard1
    router --> shard2
    router --> shard3
    router --> shard4
    
    app1 --> router
    app2 --> router
    app3 --> router
```

### Sharding Configuration

| Shard       | Merchant Range     | Expected Load            | Hardware Profile   |
| ----------- | ------------------ | ------------------------ | ------------------ |
| **Shard 1** | 0-999 (Enterprise) | High volume, low latency | 32GB RAM, NVMe SSD |
| **Shard 2** | 1000-1999 (Pro)    | Medium volume            | 16GB RAM, SSD      |
| **Shard 3** | 2000-2999 (Free)   | Low volume, high count   | 8GB RAM, SSD       |
| **Shard 4** | 3000+ (Overflow)   | Variable                 | Auto-scaling       |

### Shard Key Strategy

```go
type ShardRouter struct {
    shards map[int]*sql.DB
}

func (sr *ShardRouter) GetShard(merchantID uuid.UUID) *sql.DB {
    // Hash-based sharding for even distribution
    hash := fnv.New32a()
    hash.Write(merchantID[:])
    shardID := int(hash.Sum32()) % len(sr.shards)
    return sr.shards[shardID]
}

func (sr *ShardRouter) GetShardForRange(start, end time.Time) []*sql.DB {
    // Cross-shard queries for analytics
    return getAllShards()
}
```

---

## Read Replica Strategy

### Read/Write Separation

```mermaid
graph LR
    subgraph "Write Operations"
        write_app[Write API] --> primary_db[(Primary DB<br/>Event Store)]
        primary_db --> kafka[Kafka Events]
    end
    
    subgraph "Read Operations"
        read_app[Read API] --> read_replica[(Read Replica<br/>Read Models)]
        dashboard[Dashboard] --> analytics_replica[(Analytics Replica)]
        customer_ui[Customer UI] --> customer_replica[(Customer Replica)]
    end
    
    kafka --> projection_service[Projection Service]
    projection_service --> read_replica
    projection_service --> analytics_replica
    projection_service --> customer_replica
```

### Replica Configuration

| Replica Type     | Purpose               | Lag Tolerance | Caching Strategy |
| ---------------- | --------------------- | ------------- | ---------------- |
| **Primary Read** | General queries       | <100ms        | Redis 1min TTL   |
| **Analytics**    | Reporting, dashboards | <5 seconds    | Redis 5min TTL   |
| **Customer**     | Payment pages         | <500ms        | Redis 30sec TTL  |
| **Archive**      | Historical data       | <1 hour       | No cache         |

---

## Performance Optimization

### Query Performance Targets

| Query Type           | Target Latency | Example          | Optimization        |
| -------------------- | -------------- | ---------------- | ------------------- |
| **Point Lookups**    | <10ms          | Invoice by ID    | Primary key indexes |
| **Range Queries**    | <100ms         | Invoices by date | Composite indexes   |
| **Full-Text Search** | <500ms         | Invoice search   | GIN indexes         |
| **Analytics**        | <2000ms        | Revenue reports  | Materialized views  |
| **Cross-Shard**      | <5000ms        | Platform metrics | Async aggregation   |

### Caching Strategy

```mermaid
graph TB
    subgraph "Cache Layers"
        redis_l1[Redis L1<br/>30sec TTL<br/>Hot Data]
        redis_l2[Redis L2<br/>5min TTL<br/>Warm Data]
        db_cache[DB Query Cache<br/>Shared Buffers]
    end
    
    subgraph "Cache Usage"
        invoice_lookup[Invoice Lookups] --> redis_l1
        merchant_data[Merchant Data] --> redis_l2
        analytics[Analytics] --> redis_l2
        payments[Payment Status] --> redis_l1
    end
```

### Connection Pooling

```go
type DatabaseConfig struct {
    MaxOpenConns    int           // 25 per shard
    MaxIdleConns    int           // 5 per shard  
    ConnMaxLifetime time.Duration // 5 minutes
    ConnMaxIdleTime time.Duration // 2 minutes
}

// Shard-aware connection pooling
type ShardedConnectionPool struct {
    pools map[string]*sql.DB
    config DatabaseConfig
}
```

---

## Monitoring and Maintenance

### Key Metrics to Monitor

| Metric Category       | Metric                | Alert Threshold | Action               |
| --------------------- | --------------------- | --------------- | -------------------- |
| **Query Performance** | Average query time    | >100ms          | Index optimization   |
| **Connection Pool**   | Active connections    | >80% pool size  | Scale connections    |
| **Partition Size**    | Partition row count   | >50M rows       | Create new partition |
| **Shard Balance**     | Shard size difference | >30% variation  | Rebalance data       |
| **Replication Lag**   | Primary-replica delay | >1 second       | Check network/load   |

### Automated Maintenance

```sql
-- Automated partition creation
CREATE OR REPLACE FUNCTION create_monthly_partitions()
RETURNS void AS $$
DECLARE
    start_date date;
    end_date date;
    partition_name text;
BEGIN
    start_date := date_trunc('month', CURRENT_DATE + interval '1 month');
    end_date := start_date + interval '1 month';
    partition_name := 'events_' || to_char(start_date, 'YYYY_MM');
    
    EXECUTE format('CREATE TABLE %I PARTITION OF events 
                    FOR VALUES FROM (%L) TO (%L)',
                   partition_name, start_date, end_date);
END;
$$ LANGUAGE plpgsql;

-- Schedule monthly partition creation
SELECT cron.schedule('create-partitions', '0 0 25 * *', 'SELECT create_monthly_partitions();');
```