# POS Machine Data Synchronization Architecture
## gRPC-based Bidirectional Sync System

---

## Table of Contents
1. [Overview](#overview)
2. [Sync Categories](#sync-categories)
3. [Architecture](#architecture)
4. [Data Flow](#data-flow)
5. [Implementation Guide](#implementation-guide)
6. [Conflict Resolution](#conflict-resolution)
7. [Performance Optimization](#performance-optimization)

---

## Overview

### System Architecture
```
┌─────────────────────────────────┐
│   Central Server (Cloud)        │
│   - Multi-tenant Database       │
│   - All Organizations           │
│   - All Stores                  │
│   - Analytics & Reporting       │
└──────────┬──────────────────────┘
           │ gRPC Sync
           │ (Bidirectional)
           │
┌──────────┴──────────────────────┐
│   POS Machine (Store)           │
│   - Single Store Database       │
│   - Same Schema                 │
│   - Offline-capable             │
│   - Local Processing            │
└─────────────────────────────────┘
```

### Sync Objectives
- ✅ **Bidirectional**: Server ↔ POS machine
- ✅ **Incremental**: Only changed data
- ✅ **Conflict-aware**: Handle concurrent updates
- ✅ **Offline-capable**: Queue when disconnected
- ✅ **Store-specific**: Filter by store_id
- ✅ **Efficient**: Minimize bandwidth

---

## Sync Categories

### Category 1: FULL SYNC (Server → POS)
**Complete table download** - Organization-wide master data

| Table | Description | Frequency |
|-------|-------------|-----------|
| modules | Menu structure | On change |
| menus | Navigation menus | On change |
| submenus | Sub-navigation | On change |
| permissions | Access control | On change |
| module_permissions | Permission mappings | On change |
| menu_permissions | Permission mappings | On change |
| submenu_permissions | Permission mappings | On change |
| roles | User roles | On change |
| role_permissions | Role assignments | On change |
| ui_settings | UI configurations | On change |
| role_ui_customizations | UI customizations | On change |
| organizations | Org info (single) | On change |
| product_categories | Category hierarchy | On change |
| brands | Product brands | On change |
| units_of_measure | UOM master | On change |
| price_lists | Pricing schemes | On change |
| tax_categories | Tax rates | On change |

**Characteristics:**
- No store_id filtering
- Shared across all stores
- Change-driven sync
- Relatively small data volume

---

### Category 2: STORE-SPECIFIC SYNC (Server → POS)
**Filtered by store_id** - Data specific to POS location

| Table | Filter | Frequency |
|-------|--------|-----------|
| stores | id = $store_id | On change |
| storage_locations | store_id = $store_id | On change |
| users | via user_store_access | On change |
| user_roles | via user_store_access | On change |
| user_store_access | store_id = $store_id | On change |
| cashiers | store_id = $store_id | On change |
| pos_terminals | store_id = $store_id | On change |
| products | org-wide + active | Daily |
| product_variants | org-wide + active | Daily |
| product_barcodes | org-wide | Daily |
| product_prices | org-wide + active | Daily |
| product_uom_conversions | org-wide | Daily |
| inventory_stock | store_id = $store_id | Hourly |
| product_batches | store_id = $store_id | Hourly |
| product_serial_numbers | store_id = $store_id | Hourly |
| customers | org-wide + active | Daily |
| suppliers | org-wide + active | Weekly |
| restaurant_tables | store_id = $store_id | On change |
| menu_categories | store_id = $store_id | On change |
| menu_items | store_id = $store_id | On change |
| menu_item_modifiers | store_id = $store_id | On change |
| recipes | org-wide | Weekly |
| recipe_ingredients | org-wide | Weekly |
| carts | store_id = $store_id | Real-time |
| sales_orders_v2 | store_id = $store_id | Real-time |
| invoices | store_id = $store_id | Real-time |

**Characteristics:**
- Filtered by store_id or organization_id
- Larger data volume
- Mixed frequency based on data type
- Some require real-time sync

---

### Category 3: BIDIRECTIONAL SYNC (POS → Server)
**Data created on POS** - Must sync back to central server

| Table | Direction | Priority | Frequency |
|-------|-----------|----------|-----------|
| cashier_sessions | POS → Server | High | Real-time |
| pos_transactions | POS → Server | Critical | Real-time |
| pos_transaction_lines | POS → Server | Critical | Real-time |
| pos_payments | POS → Server | Critical | Real-time |
| stock_movements | Both | High | Real-time |
| stock_counts | POS → Server | Medium | On completion |
| stock_count_lines | POS → Server | Medium | On completion |
| sales_orders | Both | High | Real-time |
| sales_order_lines | Both | High | Real-time |
| restaurant_orders | POS → Server | High | Real-time |
| restaurant_order_items | POS → Server | High | Real-time |
| waste_logs | POS → Server | Low | Daily |
| carts | Both | High | Real-time |
| cart_items | Both | High | Real-time |
| sales_orders_v2 | Both | Critical | Real-time |
| sales_order_lines_v2 | Both | Critical | Real-time |
| invoices | Both | Critical | Real-time |

**Characteristics:**
- Created on POS machine
- Must upload to server
- Conflict resolution required
- Queue when offline

---

### Category 4: NO SYNC (Excluded)
**Never synced to POS** - Server-only data

| Table | Reason |
|-------|--------|
| profit_loss_analytics | Reporting only |
| discount_analytics | Reporting only |
| sales_analytics | Reporting only |
| purchase_analytics | Reporting only |
| inventory_analytics | Reporting only |
| stores (other stores) | Not relevant |
| tenants | Infrastructure |

---

## Architecture

### gRPC Service Definition

```protobuf
syntax = "proto3";

package possync;

service POSSync {
    // Initial full sync
    rpc InitialSync(SyncRequest) returns (stream SyncResponse);
    
    // Incremental sync (delta)
    rpc IncrementalSync(SyncRequest) returns (stream SyncResponse);
    
    // Upload POS data to server
    rpc UploadPOSData(stream POSDataUpload) returns (UploadResponse);
    
    // Real-time event sync
    rpc SyncEvents(stream SyncEvent) returns (stream SyncEvent);
    
    // Conflict resolution
    rpc ResolveConflict(ConflictRequest) returns (ConflictResponse);
    
    // Health check
    rpc HealthCheck(HealthRequest) returns (HealthResponse);
}

message SyncRequest {
    int32 organization_id = 1;
    int32 store_id = 2;
    int32 terminal_id = 3;
    string last_sync_timestamp = 4; // ISO 8601
    repeated string tables = 5; // Specific tables to sync
    SyncMode mode = 6;
}

enum SyncMode {
    FULL = 0;
    INCREMENTAL = 1;
    TABLES_ONLY = 2;
}

message SyncResponse {
    string table_name = 1;
    int32 total_records = 2;
    int32 current_batch = 3;
    bytes data = 4; // JSON or protobuf serialized
    string checksum = 5;
    bool has_more = 6;
}

message POSDataUpload {
    string table_name = 1;
    bytes data = 2; // JSON or protobuf serialized
    string client_timestamp = 3;
    string device_id = 4;
}

message UploadResponse {
    bool success = 1;
    int32 records_processed = 2;
    int32 records_failed = 3;
    repeated ConflictInfo conflicts = 4;
}

message SyncEvent {
    string event_type = 1; // create, update, delete
    string table_name = 2;
    string record_id = 3;
    bytes data = 4;
    string timestamp = 5;
}

message ConflictInfo {
    string table_name = 1;
    string record_id = 2;
    string conflict_type = 3;
    bytes server_version = 4;
    bytes client_version = 5;
}

message ConflictRequest {
    string conflict_id = 1;
    string resolution_strategy = 2; // server_wins, client_wins, merge
    bytes merged_data = 3;
}

message ConflictResponse {
    bool success = 1;
    string message = 2;
}

message HealthRequest {
    int32 terminal_id = 1;
}

message HealthResponse {
    bool server_healthy = 1;
    bool database_healthy = 2;
    string server_version = 3;
    int64 server_time = 4;
}
```

---

## Data Flow

### 1. Initial Sync (POS Setup)

```
POS Machine                          Server
    │                                   │
    ├─── InitialSync Request ──────────>│
    │    (store_id, terminal_id)        │
    │                                   │
    │<─── FULL SYNC Tables ─────────────┤
    │    (modules, menus, roles, etc)   │
    │                                   │
    │<─── STORE-SPECIFIC Data ──────────┤
    │    (products, inventory, etc)     │
    │                                   │
    ├─── Acknowledge ──────────────────>│
    │                                   │
    └─── Ready for Operations           │
```

**Steps:**
1. POS sends InitialSync request
2. Server streams all Category 1 tables (FULL SYNC)
3. Server streams all Category 2 tables (STORE-SPECIFIC)
4. POS validates checksums
5. POS marks sync complete
6. POS ready for operations

---

### 2. Incremental Sync (Periodic)

```
POS Machine                          Server
    │                                   │
    ├─── IncrementalSync Request ─────>│
    │    (last_sync: 2024-01-15T10:00) │
    │                                   │
    │<─── Changed Records ──────────────┤
    │    (updated_at > last_sync)       │
    │                                   │
    ├─── UploadPOSData ────────────────>│
    │    (transactions, stock moves)    │
    │                                   │
    │<─── Upload Acknowledgment ────────┤
    │    (success, conflicts)           │
    │                                   │
    └─── Update last_sync_timestamp     │
```

**Steps:**
1. POS sends last successful sync timestamp
2. Server queries changes since that time
3. Server streams changed records
4. POS uploads locally created data
5. Server processes uploads
6. Server returns conflicts (if any)
7. POS updates sync timestamp

---

### 3. Real-time Sync (Event-driven)

```
POS Machine                          Server
    │                                   │
    ├─── SyncEvents Stream ───────────>│
    │    (continuous connection)        │
    │                                   │
    │    [Transaction Created]          │
    ├─── Event: pos_transaction ──────>│
    │                                   │
    │<─── Acknowledgment ───────────────┤
    │                                   │
    │    [Inventory Updated on Server]  │
    │<─── Event: inventory_stock ───────┤
    │                                   │
    ├─── Acknowledgment ───────────────>│
    │                                   │
```

**Use Cases:**
- POS transactions
- Cart updates
- Order status changes
- Stock movements
- Price changes

---

### 4. Offline Queue & Reconnection

```
POS Machine (Offline)               Server
    │                                   │
    │  [Network Disconnected]           X
    │                                   
    ├─ Queue locally:                   
    │  - Transactions                    
    │  - Stock movements                 
    │  - Orders                          
    │                                   
    │  [Network Reconnected]            │
    ├─── HealthCheck ──────────────────>│
    │                                   │
    │<─── Server Healthy ───────────────┤
    │                                   │
    ├─── UploadPOSData ────────────────>│
    │    (queued transactions)          │
    │                                   │
    │<─── Process & Validate ───────────┤
    │                                   │
    ├─── IncrementalSync ──────────────>│
    │    (get server changes)           │
    │                                   │
    └─── Sync Complete                  │
```

**Steps:**
1. Detect network disconnection
2. Enable offline mode
3. Queue all write operations
4. Continue POS operations
5. Detect reconnection
6. Health check server
7. Upload queued data
8. Download server changes
9. Resolve conflicts
10. Resume normal sync

---

## Implementation Guide

### Server-Side (Go Example)

```go
package main

import (
    "context"
    "database/sql"
    "time"
    
    pb "your_package/possync"
    "google.golang.org/grpc"
)

type POSSyncServer struct {
    pb.UnimplementedPOSSyncServer
    db *sql.DB
}

func (s *POSSyncServer) InitialSync(
    req *pb.SyncRequest,
    stream pb.POSSync_InitialSyncServer,
) error {
    ctx := stream.Context()
    
    // Validate request
    if err := s.validateRequest(req); err != nil {
        return err
    }
    
    // Log sync start
    syncLog := s.createSyncLog(req, "full", "download")
    
    // Sync Category 1: Full Sync Tables
    fullSyncTables := []string{
        "modules", "menus", "submenus",
        "permissions", "roles", "role_permissions",
        // ... etc
    }
    
    for _, tableName := range fullSyncTables {
        if err := s.syncTable(ctx, stream, tableName, req, false); err != nil {
            s.updateSyncLog(syncLog, "failed", err)
            return err
        }
    }
    
    // Sync Category 2: Store-Specific Tables
    storeSpecificTables := []string{
        "stores", "users", "cashiers", "pos_terminals",
        "products", "inventory_stock", "customers",
        // ... etc
    }
    
    for _, tableName := range storeSpecificTables {
        if err := s.syncTable(ctx, stream, tableName, req, true); err != nil {
            s.updateSyncLog(syncLog, "failed", err)
            return err
        }
    }
    
    // Complete sync
    s.updateSyncLog(syncLog, "completed", nil)
    return nil
}

func (s *POSSyncServer) syncTable(
    ctx context.Context,
    stream pb.POSSync_InitialSyncServer,
    tableName string,
    req *pb.SyncRequest,
    filterByStore bool,
) error {
    // Build query
    query := s.buildSyncQuery(tableName, req, filterByStore)
    
    // Execute query
    rows, err := s.db.QueryContext(ctx, query, req.StoreId, req.OrganizationId)
    if err != nil {
        return err
    }
    defer rows.Close()
    
    // Stream data in batches
    const batchSize = 100
    batch := make([]map[string]interface{}, 0, batchSize)
    
    for rows.Next() {
        record := s.scanRow(rows, tableName)
        batch = append(batch, record)
        
        if len(batch) >= batchSize {
            if err := s.sendBatch(stream, tableName, batch); err != nil {
                return err
            }
            batch = batch[:0] // Clear batch
        }
    }
    
    // Send remaining records
    if len(batch) > 0 {
        if err := s.sendBatch(stream, tableName, batch); err != nil {
            return err
        }
    }
    
    return nil
}

func (s *POSSyncServer) IncrementalSync(
    req *pb.SyncRequest,
    stream pb.POSSync_IncrementalSyncServer,
) error {
    ctx := stream.Context()
    
    // Parse last sync time
    lastSync, err := time.Parse(time.RFC3339, req.LastSyncTimestamp)
    if err != nil {
        return err
    }
    
    // Get changed records
    tables := s.getIncrementalSyncTables()
    
    for _, tableName := range tables {
        query := s.buildIncrementalQuery(tableName, req, lastSync)
        
        if err := s.syncChangedRecords(ctx, stream, query, tableName); err != nil {
            return err
        }
    }
    
    return nil
}

func (s *POSSyncServer) UploadPOSData(
    stream pb.POSSync_UploadPOSDataServer,
) error {
    ctx := stream.Context()
    
    recordsProcessed := 0
    recordsFailed := 0
    conflicts := make([]*pb.ConflictInfo, 0)
    
    for {
        upload, err := stream.Recv()
        if err == io.EOF {
            // Done receiving
            return stream.SendAndClose(&pb.UploadResponse{
                Success:          recordsFailed == 0,
                RecordsProcessed: int32(recordsProcessed),
                RecordsFailed:    int32(recordsFailed),
                Conflicts:        conflicts,
            })
        }
        if err != nil {
            return err
        }
        
        // Process uploaded data
        conflict, err := s.processUploadedRecord(ctx, upload)
        if err != nil {
            recordsFailed++
        } else {
            recordsProcessed++
        }
        
        if conflict != nil {
            conflicts = append(conflicts, conflict)
        }
    }
}

func (s *POSSyncServer) processUploadedRecord(
    ctx context.Context,
    upload *pb.POSDataUpload,
) (*pb.ConflictInfo, error) {
    // Parse data
    var data map[string]interface{}
    if err := json.Unmarshal(upload.Data, &data); err != nil {
        return nil, err
    }
    
    // Check for conflicts
    existingRecord, err := s.fetchExistingRecord(
        upload.TableName,
        data["id"],
    )
    if err != nil && err != sql.ErrNoRows {
        return nil, err
    }
    
    if existingRecord != nil {
        // Check timestamps
        serverTime := existingRecord["updated_at"].(time.Time)
        clientTime, _ := time.Parse(time.RFC3339, upload.ClientTimestamp)
        
        if serverTime.After(clientTime) {
            // Conflict detected
            return &pb.ConflictInfo{
                TableName:     upload.TableName,
                RecordId:      fmt.Sprintf("%v", data["id"]),
                ConflictType:  "version_mismatch",
                ServerVersion: s.serializeRecord(existingRecord),
                ClientVersion: upload.Data,
            }, nil
        }
    }
    
    // No conflict - insert/update record
    return nil, s.upsertRecord(ctx, upload.TableName, data)
}
```

### Client-Side (Go Example)

```go
package main

import (
    "context"
    "database/sql"
    "time"
    
    pb "your_package/possync"
    "google.golang.org/grpc"
)

type POSClient struct {
    db         *sql.DB
    grpcClient pb.POSSyncClient
    storeID    int32
    terminalID int32
    orgID      int32
}

func (c *POSClient) InitialSync() error {
    ctx := context.Background()
    
    req := &pb.SyncRequest{
        OrganizationId: c.orgID,
        StoreId:        c.storeID,
        TerminalId:     c.terminalID,
        Mode:           pb.SyncMode_FULL,
    }
    
    stream, err := c.grpcClient.InitialSync(ctx, req)
    if err != nil {
        return err
    }
    
    // Receive and process data
    for {
        resp, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        
        if err := c.processSync Response(resp); err != nil {
            return err
        }
    }
    
    // Update last sync time
    return c.updateLastSyncTime(time.Now())
}

func (c *POSClient) IncrementalSync() error {
    ctx := context.Background()
    
    lastSync, err := c.getLastSyncTime()
    if err != nil {
        return err
    }
    
    req := &pb.SyncRequest{
        OrganizationId:    c.orgID,
        StoreId:           c.storeID,
        TerminalId:        c.terminalID,
        LastSyncTimestamp: lastSync.Format(time.RFC3339),
        Mode:              pb.SyncMode_INCREMENTAL,
    }
    
    // Download changes from server
    stream, err := c.grpcClient.IncrementalSync(ctx, req)
    if err != nil {
        return err
    }
    
    for {
        resp, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        
        if err := c.processSyncResponse(resp); err != nil {
            return err
        }
    }
    
    // Upload local changes
    if err := c.uploadLocalChanges(lastSync); err != nil {
        return err
    }
    
    return c.updateLastSyncTime(time.Now())
}

func (c *POSClient) uploadLocalChanges(since time.Time) error {
    ctx := context.Background()
    
    stream, err := c.grpcClient.UploadPOSData(ctx)
    if err != nil {
        return err
    }
    
    // Get tables to upload
    uploadTables := []string{
        "pos_transactions",
        "pos_transaction_lines",
        "pos_payments",
        "stock_movements",
        "cashier_sessions",
    }
    
    for _, tableName := range uploadTables {
        records, err := c.getLocalChanges(tableName, since)
        if err != nil {
            return err
        }
        
        for _, record := range records {
            data, _ := json.Marshal(record)
            
            if err := stream.Send(&pb.POSDataUpload{
                TableName:       tableName,
                Data:            data,
                ClientTimestamp: time.Now().Format(time.RFC3339),
                DeviceId:        fmt.Sprintf("%d", c.terminalID),
            }); err != nil {
                return err
            }
        }
    }
    
    resp, err := stream.CloseAndRecv()
    if err != nil {
        return err
    }
    
    // Handle conflicts
    if len(resp.Conflicts) > 0 {
        return c.handleConflicts(resp.Conflicts)
    }
    
    return nil
}

func (c *POSClient) RunContinuousSync() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if err := c.IncrementalSync(); err != nil {
                log.Printf("Sync error: %v", err)
            }
        }
    }
}
```

---

## Conflict Resolution

### Conflict Types

1. **Version Mismatch**
   - Server updated after POS read
   - Different updated_at timestamps
   
2. **Delete Conflict**
   - Record deleted on server
   - Modified on POS
   
3. **Unique Violation**
   - Duplicate key (order number, etc)
   - Both created independently

### Resolution Strategies

#### 1. Server Wins (Default)
```sql
-- Discard POS changes, use server version
UPDATE {table}
SET column1 = server.column1,
    column2 = server.column2,
    updated_at = server.updated_at
WHERE id = {record_id};
```

#### 2. Client Wins
```sql
-- Keep POS changes, overwrite server
UPDATE {table}
SET column1 = client.column1,
    column2 = client.column2,
    updated_at = NOW()
WHERE id = {record_id};
```

#### 3. Merge Strategy
```go
func mergeRecords(server, client map[string]interface{}) map[string]interface{} {
    merged := make(map[string]interface{})
    
    // Copy all server fields
    for k, v := range server {
        merged[k] = v
    }
    
    // Override with non-null client fields
    for k, v := range client {
        if v != nil && v != "" {
            merged[k] = v
        }
    }
    
    // Set metadata
    merged["merged_at"] = time.Now()
    merged["merge_source"] = "auto_merge"
    
    return merged
}
```

#### 4. Manual Resolution
- Store conflict in `sync_conflicts` table
- Display in admin UI
- User chooses resolution
- Apply and mark resolved

---

## Performance Optimization

### 1. Batch Processing
```go
const BatchSize = 100

func syncInBatches(records []Record) error {
    for i := 0; i < len(records); i += BatchSize {
        end := i + BatchSize
        if end > len(records) {
            end = len(records)
        }
        
        batch := records[i:end]
        if err := processBatch(batch); err != nil {
            return err
        }
    }
    return nil
}
```

### 2. Compression
```go
import "compress/gzip"

func compressData(data []byte) ([]byte, error) {
    var buf bytes.Buffer
    writer := gzip.NewWriter(&buf)
    
    if _, err := writer.Write(data); err != nil {
        return nil, err
    }
    
    if err := writer.Close(); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}
```

### 3. Delta Sync
```sql
-- Only sync changed columns
SELECT 
    id,
    CASE WHEN updated_at > $last_sync THEN name ELSE NULL END as name,
    CASE WHEN updated_at > $last_sync THEN price ELSE NULL END as price,
    updated_at
FROM products
WHERE updated_at > $last_sync;
```

### 4. Parallel Sync
```go
func parallelSync(tables []string) error {
    var wg sync.WaitGroup
    errorsChan := make(chan error, len(tables))
    
    for _, table := range tables {
        wg.Add(1)
        go func(t string) {
            defer wg.Done()
            if err := syncTable(t); err != nil {
                errorsChan <- err
            }
        }(table)
    }
    
    wg.Wait()
    close(errorsChan)
    
    for err := range errorsChan {
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

### 5. Connection Pooling
```go
db, err := sql.Open("postgres", connString)
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

---

## Sync Schedule Recommendations

| Data Type | Frequency | Method |
|-----------|-----------|--------|
| POS Transactions | Real-time | Event stream |
| Inventory Updates | Real-time | Event stream |
| Orders | Real-time | Event stream |
| Master Data | On change | Push notification |
| Product Prices | Hourly | Incremental |
| Stock Levels | 15 minutes | Incremental |
| Customers | Daily | Incremental |
| Analytics | Never | N/A |

---

## Monitoring & Alerts

### Key Metrics
- Sync latency
- Failed sync attempts
- Conflict rate
- Queue size (offline mode)
- Network connectivity
- Database performance

### Alerting Rules
```yaml
alerts:
  - name: SyncFailure
    condition: failed_syncs > 3 in 1 hour
    severity: critical
    
  - name: HighConflictRate
    condition: conflicts > 10 in 1 hour
    severity: warning
    
  - name: SyncLatency
    condition: sync_duration > 5 minutes
    severity: warning
    
  - name: OfflineMode
    condition: offline_duration > 1 hour
    severity: critical
```

---

## Testing Strategy

### Unit Tests
- Query generation
- Data serialization
- Conflict detection
- Merge logic

### Integration Tests
- Full sync flow
- Incremental sync
- Conflict resolution
- Network interruption

### Load Tests
- 1000+ products sync
- 10000+ transactions upload
- Concurrent POS machines
- Network bandwidth limits

---

## Security Considerations

### Authentication
```go
func (s *POSSyncServer) authenticate(ctx context.Context) error {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return status.Errorf(codes.Unauthenticated, "missing metadata")
    }
    
    tokens := md["authorization"]
    if len(tokens) == 0 {
        return status.Errorf(codes.Unauthenticated, "missing token")
    }
    
    return validateToken(tokens[0])
}
```

### TLS/SSL
```go
creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
server := grpc.NewServer(grpc.Creds(creds))
```

### Data Validation
- Validate store_id access
- Verify organization ownership
- Sanitize SQL inputs
- Validate data types
- Check business rules

---

## Conclusion

This sync architecture provides:
- ✅ Efficient bidirectional sync
- ✅ Offline capability
- ✅ Conflict resolution
- ✅ Real-time updates
- ✅ Scalable design
- ✅ Security & validation

**Next Steps:**
1. Implement gRPC service
2. Create client libraries
3. Set up monitoring
4. Test with real data
5. Deploy and monitor
