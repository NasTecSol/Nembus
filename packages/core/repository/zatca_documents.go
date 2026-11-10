package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type ZatcaDocumentFilters struct {
	StoreID       int32
	PosTerminalID int32
}

// ZatcaDocumentListEntry adds source location to the saved document without
// exposing the signing device's configuration or changing the chain ledger.
type ZatcaDocumentListEntry struct {
	ZatcaDocumentChain
	StoreID       pgtype.Int4 `json:"store_id"`
	PosTerminalID pgtype.Int4 `json:"pos_terminal_id"`
}

// Resolve location from the source, never from the signing device: a cloud
// device can sign documents for multiple stores and terminals. Text comparisons
// also keep malformed or deleted source references from causing cast errors.
const zatcaDocumentLocations = `WITH located_documents AS (
 SELECT d.*,
 CASE d.entity_type
   WHEN 'sales_order_v2' THEN so.store_id
   WHEN 'pos_transaction' THEN pt.store_id
   WHEN 'invoice' THEN i.store_id
   WHEN 'sales_return' THEN sr.store_id
 END AS store_id,
 CASE d.entity_type
   WHEN 'sales_order_v2' THEN so.pos_terminal_id
   WHEN 'pos_transaction' THEN COALESCE(pt.pos_terminal_id, ps.pos_terminal_id)
   WHEN 'invoice' THEN io.pos_terminal_id
   WHEN 'sales_return' THEN rs.pos_terminal_id
 END AS pos_terminal_id
 FROM zatca_document_chain d
 LEFT JOIN sales_orders_v2 so ON d.entity_type='sales_order_v2' AND so.id::text=d.entity_id AND so.organization_id=d.organization_id
 LEFT JOIN pos_transactions pt ON d.entity_type='pos_transaction' AND pt.id::text=d.entity_id
   AND EXISTS (SELECT 1 FROM stores s WHERE s.id=pt.store_id AND s.organization_id=d.organization_id)
 LEFT JOIN cashier_sessions ps ON ps.id=pt.cashier_session_id
   AND EXISTS (SELECT 1 FROM pos_terminals t WHERE t.id=ps.pos_terminal_id AND t.store_id=pt.store_id)
 LEFT JOIN invoices i ON d.entity_type='invoice' AND i.id::text=d.entity_id AND i.organization_id=d.organization_id
 LEFT JOIN sales_orders_v2 io ON io.id=i.sales_order_id AND io.organization_id=i.organization_id AND io.store_id=i.store_id
 LEFT JOIN sales_returns sr ON d.entity_type='sales_return' AND sr.id::text=d.entity_id
   AND EXISTS (SELECT 1 FROM stores s WHERE s.id=sr.store_id AND s.organization_id=d.organization_id)
 LEFT JOIN cashier_sessions rs ON rs.id=sr.cashier_session_id
   AND EXISTS (SELECT 1 FROM pos_terminals t WHERE t.id=rs.pos_terminal_id AND t.store_id=sr.store_id)
 WHERE d.organization_id=$1
)
`

const zatcaDocumentLocationFilter = ` WHERE ($2::int=0 OR store_id=$2) AND ($3::int=0 OR pos_terminal_id=$3)`

// ListZatcaDocumentsPage reads all saved document types and statuses in the
// request tenant and authenticated organization. Filters are applied before
// pagination and identically to the count. Zero filter IDs mean no filter.
func (q *Queries) ListZatcaDocumentsPage(ctx context.Context, organizationID int32, limit int32, offset int64, filters ZatcaDocumentFilters) ([]ZatcaDocumentListEntry, int64, error) {
	if organizationID <= 0 {
		return nil, 0, fmt.Errorf("organization is required")
	}
	var total int64
	if err := q.db.QueryRow(ctx, `/* CountZatcaDocuments */`+zatcaDocumentLocations+
		` SELECT count(*) FROM located_documents`+zatcaDocumentLocationFilter, organizationID, filters.StoreID, filters.PosTerminalID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := q.db.Query(ctx, `/* ListZatcaDocumentsPage */`+zatcaDocumentLocations+`
 SELECT id,entity_type,entity_id,device_config_id,organization_id,zatca_uuid,icv,pih,xml_hash,zatca_status,zatca_response,qr_code_base64,signed_xml,submitted_at,cleared_at,created_at,updated_at
 ,store_id,pos_terminal_id
 FROM located_documents`+zatcaDocumentLocationFilter+`
 ORDER BY id DESC LIMIT $4 OFFSET $5`, organizationID, filters.StoreID, filters.PosTerminalID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	entries := make([]ZatcaDocumentListEntry, 0)
	for rows.Next() {
		var c ZatcaDocumentListEntry
		if err := rows.Scan(&c.ID, &c.EntityType, &c.EntityID, &c.DeviceConfigID, &c.OrganizationID, &c.ZatcaUuid, &c.Icv, &c.Pih, &c.XmlHash, &c.ZatcaStatus, &c.ZatcaResponse, &c.QrCodeBase64, &c.SignedXml, &c.SubmittedAt, &c.ClearedAt, &c.CreatedAt, &c.UpdatedAt, &c.StoreID, &c.PosTerminalID); err != nil {
			return nil, 0, err
		}
		entries = append(entries, c)
	}
	return entries, total, rows.Err()
}
