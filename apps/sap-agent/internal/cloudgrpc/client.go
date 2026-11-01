package cloudgrpc

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/sapagentpb"
	"github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap/contracts"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const maxRetries = 3

type Client struct {
	cfg        *config.AgentConfig
	conn       *grpc.ClientConn
	rpc        sapagentpb.SapAgentServiceClient
	grpcTarget string
	mu         sync.Mutex
}

func NewClient(cfg *config.AgentConfig) (*Client, error) {
	target := cfg.Cloud.GRPCAddr
	if target == "" {
		target = ExtractGRPCTarget(cfg.Cloud.BaseURL)
	}
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial grpc %s: %w", target, err)
	}
	return &Client{
		cfg:        cfg,
		conn:       conn,
		rpc:        sapagentpb.NewSapAgentServiceClient(conn),
		grpcTarget: target,
	}, nil
}

func ExtractGRPCTarget(cloudURL string) string {
	if cloudURL == "" {
		return "localhost:50051"
	}
	if strings.Contains(cloudURL, ":") && !strings.Contains(cloudURL, "://") {
		return cloudURL
	}
	if !strings.Contains(cloudURL, "://") {
		cloudURL = "http://" + cloudURL
	}
	parsed, err := url.Parse(cloudURL)
	if err != nil {
		return "localhost:50051"
	}
	host := parsed.Hostname()
	if host == "" {
		host = "localhost"
	}
	return host + ":50051"
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) authContext(ctx context.Context) context.Context {
	token := c.cfg.MigratorToken
	if token == "" {
		token = c.cfg.Cloud.APIKey
	}
	md := metadata.Pairs("x-tenant-id", c.cfg.TenantSlug)
	if token != "" {
		md.Append("authorization", "Bearer "+token)
		md.Append("x-api-key", token)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

func (c *Client) TestConnection(ctx context.Context) error {
	_, err := c.ResolveTenant(ctx, c.cfg.TenantSlug)
	return err
}

func (c *Client) ResolveTenant(ctx context.Context, slug string) (*TenantResolveResult, error) {
	if slug == "" {
		slug = c.cfg.TenantSlug
	}
	resp, err := c.rpc.ResolveTenant(c.authContext(ctx), &sapagentpb.TenantResolveRequest{TenantSlug: slug})
	if err != nil {
		return nil, err
	}
	return &TenantResolveResult{
		TenantSlug:     resp.GetTenantSlug(),
		TenantName:     resp.GetTenantName(),
		OrganizationID: resp.GetOrganizationId(),
		DefaultStoreID: resp.GetDefaultStoreId(),
	}, nil
}

func (c *Client) SendBatchWithRetry(ctx context.Context, payload *contracts.MigrationBatchPayload) (*contracts.MigrationBatchResponse, error) {
	requestID := uuid.New().String()
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			wait := time.Second * time.Duration(1<<(attempt-1))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}
		resp, err := c.sendMigrationOnce(ctx, payload, requestID)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("grpc migration failed after %d attempts: %w", maxRetries, lastErr)
}

func (c *Client) sendMigrationOnce(ctx context.Context, payload *contracts.MigrationBatchPayload, requestID string) (*contracts.MigrationBatchResponse, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var compressed bytes.Buffer
	gz := gzip.NewWriter(&compressed)
	if _, err := gz.Write(jsonData); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	orgID := int32(payload.OrganizationID)
	if orgID <= 0 {
		orgID = int32(c.cfg.Cloud.OrganizationID)
	}

	stream, err := c.rpc.StreamMigrationIngest(c.authContext(ctx))
	if err != nil {
		return nil, err
	}
	if err := stream.Send(&sapagentpb.MigrationChunk{
		TenantSlug: c.cfg.TenantSlug, OrganizationId: orgID,
		BatchId: payload.BatchID, RunId: payload.RunID, Domain: string(payload.Domain),
		SequenceNumber: int32(payload.SequenceNumber), IsLastBatch: payload.IsLastBatch,
		PayloadGzip: compressed.Bytes(), RequestId: requestID,
	}); err != nil {
		return nil, err
	}
	if err := stream.CloseSend(); err != nil {
		return nil, err
	}
	ack, err := stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}
	return &contracts.MigrationBatchResponse{
		Success: ack.GetSuccess(), BatchID: ack.GetBatchId(),
		Domain: contracts.DomainType(ack.GetDomain()),
		RecordsStaged: int(ack.GetRecordsStaged()), RecordsMerged: int(ack.GetRecordsMerged()),
		RecordsFailed: int(ack.GetRecordsFailed()), RecordsSkipped: int(ack.GetRecordsSkipped()),
		ErrorMessage: ack.GetErrorMessage(), Errors: ack.GetErrors(),
	}, nil
}

func (c *Client) masterData(ctx context.Context, operation string, payload any) (int32, []byte, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, err
	}
	stream, err := c.rpc.StreamMasterDataUpsert(c.authContext(ctx))
	if err != nil {
		return 0, nil, err
	}
	if err := stream.Send(&sapagentpb.MasterDataUpsert{
		TenantSlug: c.cfg.TenantSlug, OrganizationId: int32(c.cfg.Cloud.OrganizationID),
		Operation: operation, PayloadJson: payloadJSON,
	}); err != nil {
		return 0, nil, err
	}
	if err := stream.CloseSend(); err != nil {
		return 0, nil, err
	}
	ack, err := stream.Recv()
	if err != nil && err != io.EOF {
		return 0, nil, err
	}
	if ack != nil && !ack.GetSuccess() {
		return ack.GetEntityId(), ack.GetResultJson(), fmt.Errorf("%s", ack.GetErrorMessage())
	}
	if ack != nil {
		return ack.GetEntityId(), ack.GetResultJson(), nil
	}
	return 0, nil, nil
}

func (c *Client) UpsertCategory(ctx context.Context, code, name string) (int32, error) {
	id, _, err := c.masterData(ctx, "upsert_category", map[string]string{"code": code, "name": name})
	return id, err
}

func (c *Client) UpsertUoM(ctx context.Context, code, name string) (int32, error) {
	id, _, err := c.masterData(ctx, "upsert_uom", map[string]string{"code": code, "name": name})
	return id, err
}

func (c *Client) UpsertProduct(ctx context.Context, sku, name, foreignName string, categoryID, uomID, taxCategoryID *int32, isWeighted bool, costPrice float64, skipCostUpdate bool) (int32, error) {
	id, _, err := c.masterData(ctx, "upsert_product", map[string]any{
		"sku": sku, "name": name, "foreign_name": foreignName,
		"category_id": categoryID, "uom_id": uomID, "tax_category_id": taxCategoryID,
		"is_weighted": isWeighted, "cost_price": costPrice, "skip_cost_update": skipCostUpdate,
	})
	return id, err
}

func (c *Client) LookupProductBySKU(ctx context.Context, sku string) (int32, float64, bool, error) {
	_, result, err := c.masterData(ctx, "lookup_product_by_sku", map[string]string{"sku": sku})
	if err != nil {
		return 0, 0, false, err
	}
	var parsed struct {
		Found     bool    `json:"found"`
		ProductID int32   `json:"product_id"`
		Cost      float64 `json:"cost"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		return 0, 0, false, err
	}
	if !parsed.Found {
		return 0, 0, false, nil
	}
	return parsed.ProductID, parsed.Cost, true, nil
}

func (c *Client) UpsertBarcode(ctx context.Context, productID int32, barcode string, isPrimary bool) error {
	_, _, err := c.masterData(ctx, "upsert_barcode", map[string]any{
		"product_id": productID, "barcode": barcode, "is_primary": isPrimary,
	})
	return err
}

func (c *Client) UpsertProductPrice(ctx context.Context, productID, priceListID int32, uomID *int32, price float64) error {
	_, _, err := c.masterData(ctx, "upsert_product_price", map[string]any{
		"product_id": productID, "price_list_id": priceListID, "uom_id": uomID, "price": price,
	})
	return err
}

func (c *Client) streamOutbox(ctx context.Context, limit int32) (sapagentpb.SapAgentService_StreamSapOutboxClient, error) {
	return c.rpc.StreamSapOutbox(c.authContext(ctx), &sapagentpb.SapOutboxRequest{
		TenantSlug: c.cfg.TenantSlug, Limit: limit,
	})
}

func (c *Client) FetchUnpostedPOSTransactions(ctx context.Context, limit int) ([]POSTransaction, error) {
	stream, err := c.streamOutbox(ctx, int32(limit))
	if err != nil {
		return nil, err
	}
	var out []POSTransaction
	for {
		ev, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if ev.GetEntityType() != "pos_transaction" {
			continue
		}
		var t POSTransaction
		if err := json.Unmarshal(ev.GetPayloadJson(), &t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

func (c *Client) findOutboxEvent(ctx context.Context, entityType string, entityID int64) (*sapagentpb.SapOutboxEvent, error) {
	stream, err := c.streamOutbox(ctx, 500)
	if err != nil {
		return nil, err
	}
	for {
		ev, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if ev.GetEntityType() == entityType && ev.GetEntityId() == entityID {
			return ev, nil
		}
	}
	return nil, fmt.Errorf("%s %d not found in outbox", entityType, entityID)
}

func (c *Client) FetchTransactionLines(ctx context.Context, transactionID int64) ([]POSTransactionLine, error) {
	ev, err := c.findOutboxEvent(ctx, "pos_transaction", transactionID)
	if err != nil {
		return nil, err
	}
	var lines []POSTransactionLine
	if len(ev.GetLinesJson()) > 0 {
		if err := json.Unmarshal(ev.GetLinesJson(), &lines); err != nil {
			return nil, err
		}
	}
	return lines, nil
}

func (c *Client) FetchTransactionPayments(ctx context.Context, transactionID int64) ([]POSPayment, error) {
	ev, err := c.findOutboxEvent(ctx, "pos_transaction", transactionID)
	if err != nil {
		return nil, err
	}
	var payments []POSPayment
	if len(ev.GetPaymentsJson()) > 0 {
		if err := json.Unmarshal(ev.GetPaymentsJson(), &payments); err != nil {
			return nil, err
		}
	}
	return payments, nil
}

func (c *Client) ackOutbox(ctx context.Context, entityType string, entityID int64, sapDocEntry, sapDocNum int) error {
	_, err := c.rpc.AckSapOutbox(c.authContext(ctx), &sapagentpb.AckSapOutboxRequest{
		TenantSlug: c.cfg.TenantSlug, OrganizationId: int32(c.cfg.Cloud.OrganizationID),
		Items: []*sapagentpb.SapOutboxAckItem{{
			EntityType: entityType, EntityId: entityID,
			SapDocEntry: int32(sapDocEntry), SapDocNum: int32(sapDocNum),
		}},
	})
	return err
}

func (c *Client) MarkTransactionPostedToSAP(ctx context.Context, transactionID int64, sapDocEntry, sapDocNum int) error {
	return c.ackOutbox(ctx, "pos_transaction", transactionID, sapDocEntry, sapDocNum)
}

func (c *Client) MarkSalesReturnPostedToSAP(ctx context.Context, returnID int64, sapDocEntry, sapDocNum int) error {
	return c.ackOutbox(ctx, "sales_return", returnID, sapDocEntry, sapDocNum)
}

func (c *Client) FetchUnpostedSalesReturns(ctx context.Context, limit int) ([]SalesReturn, error) {
	stream, err := c.streamOutbox(ctx, int32(limit))
	if err != nil {
		return nil, err
	}
	var out []SalesReturn
	for {
		ev, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if ev.GetEntityType() != "sales_return" {
			continue
		}
		var r SalesReturn
		if err := json.Unmarshal(ev.GetPayloadJson(), &r); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func (c *Client) FetchSalesReturnLines(ctx context.Context, returnID int64) ([]SalesReturnLine, error) {
	ev, err := c.findOutboxEvent(ctx, "sales_return", returnID)
	if err != nil {
		return nil, err
	}
	var lines []SalesReturnLine
	if err := json.Unmarshal(ev.GetLinesJson(), &lines); err != nil {
		return nil, err
	}
	return lines, nil
}

func (c *Client) FetchCustomerSAPCode(ctx context.Context, customerID int32) (string, error) {
	_, result, err := c.masterData(ctx, "fetch_customer_sap_code", map[string]int32{"customer_id": customerID})
	if err != nil {
		return "", err
	}
	var parsed struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(result, &parsed)
	return parsed.Code, nil
}

func (c *Client) FetchOriginalSAPDocEntry(ctx context.Context, originalTxID int32) (int, error) {
	id, result, err := c.masterData(ctx, "fetch_original_sap_doc_entry", map[string]int32{"original_tx_id": originalTxID})
	if err != nil {
		return 0, err
	}
	if len(result) > 0 {
		var parsed struct {
			DocEntry int `json:"doc_entry"`
		}
		_ = json.Unmarshal(result, &parsed)
		return parsed.DocEntry, nil
	}
	return int(id), nil
}

func (c *Client) HoldTransactionForStockReview(ctx context.Context, txID int64, txNum string, storeID int32, whCode string, totalAmount float64, shortageItems []ShortageItem, summary string) error {
	shortageJSON, _ := json.Marshal(shortageItems)
	_, err := c.rpc.HoldSapOutbox(c.authContext(ctx), &sapagentpb.HoldSapOutboxRequest{
		TenantSlug: c.cfg.TenantSlug, OrganizationId: int32(c.cfg.Cloud.OrganizationID),
		EntityType: "pos_transaction", EntityId: txID, TransactionNumber: txNum,
		StoreId: storeID, WarehouseCode: whCode, TotalAmount: totalAmount,
		Summary: summary, ShortageItemsJson: shortageJSON,
	})
	return err
}

func (c *Client) GetActiveProductPrice(ctx context.Context, productID, priceListID int32) (float64, bool, error) {
	_, result, err := c.masterData(ctx, "get_active_product_price", map[string]int32{
		"product_id": productID, "price_list_id": priceListID,
	})
	if err != nil {
		return 0, false, nil
	}
	var parsed struct {
		Price float64 `json:"price"`
	}
	if len(result) == 0 {
		return 0, false, nil
	}
	_ = json.Unmarshal(result, &parsed)
	return parsed.Price, true, nil
}

func (c *Client) GetProductCostPrice(ctx context.Context, productID int32) (float64, error) {
	_, result, err := c.masterData(ctx, "get_product_cost_price", map[string]int32{"product_id": productID})
	if err != nil {
		return 0, err
	}
	var parsed struct {
		Cost float64 `json:"cost"`
	}
	_ = json.Unmarshal(result, &parsed)
	return parsed.Cost, nil
}

func (c *Client) GetApprovalPolicyRules(ctx context.Context, domain, action string) (map[string]any, bool, error) {
	_, result, err := c.masterData(ctx, "get_approval_policy_rules", map[string]string{
		"domain": domain,
		"action": action,
	})
	if err != nil {
		return nil, false, err
	}
	var parsed struct {
		RequiresApproval bool           `json:"requires_approval"`
		Rules            map[string]any `json:"rules"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		return nil, false, err
	}
	if parsed.Rules == nil {
		parsed.Rules = map[string]any{}
	}
	return parsed.Rules, parsed.RequiresApproval, nil
}

func (c *Client) SubmitPriceApprovalRequest(ctx context.Context, productID int32, sku string, priceListID int32, uomID *int32, oldPrice, newPrice, oldCost, newCost *float64, reason string) error {
	_, _, err := c.masterData(ctx, "submit_price_approval_request", map[string]any{
		"product_id": productID, "product_sku": sku, "price_list_id": priceListID,
		"uom_id": uomID, "old_price": oldPrice, "new_price": newPrice,
		"old_cost_price": oldCost, "new_cost_price": newCost, "reason": reason,
	})
	return err
}

func (c *Client) GetReconciliationSummary(ctx context.Context, runID string, orgID int) ([]byte, error) {
	resp, err := c.rpc.GetReconciliationSummary(c.authContext(ctx), &sapagentpb.ReconciliationRequest{
		TenantSlug: c.cfg.TenantSlug, OrganizationId: int32(orgID), RunId: runID,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(resp)
}
