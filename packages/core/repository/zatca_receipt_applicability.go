package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ReceiptZatcaState is resolved from the persisted sale, not browser-provided
// terminal labels or QR data. A reported document remains printable after its
// device registration is disabled or removed.
type ReceiptZatcaState struct {
	QRCode        string
	Reported      bool
	RegisteredPOS bool
}

func (q *Queries) GetReceiptZatcaState(ctx context.Context, organizationID int32, salesOrderID, receiptNumber string) (ReceiptZatcaState, error) {
	var state ReceiptZatcaState
	reference := strings.TrimSpace(receiptNumber)
	var orderID any
	if explicit := strings.TrimSpace(salesOrderID); explicit != "" {
		parsed, err := uuid.Parse(explicit)
		if err != nil {
			return state, fmt.Errorf("invalid receipt sales_order_id: %w", err)
		}
		orderID = parsed
	} else if parsed, err := uuid.Parse(reference); err == nil {
		orderID = parsed
	} else if reference == "" {
		return state, fmt.Errorf("receipt requires a persisted sale identifier")
	}
	err := q.db.QueryRow(ctx, `/* GetReceiptZatcaState */
 WITH sale_candidates AS (
   SELECT o.id AS order_id, t.id::bigint AS transaction_id, o.organization_id,
          COALESCE(t.store_id,o.store_id) AS store_id,
          CASE WHEN t.id IS NOT NULL THEN t.pos_terminal_id ELSE o.pos_terminal_id END AS pos_terminal_id,
          0 AS priority
   FROM sales_orders_v2 o
   LEFT JOIN LATERAL (
       SELECT t.id,t.store_id,t.pos_terminal_id
       FROM pos_transactions t JOIN stores st ON st.id=t.store_id
       WHERE t.sales_order_id=o.id AND st.organization_id=o.organization_id
       ORDER BY t.id DESC LIMIT 1
   ) t ON true
   WHERE o.organization_id=$1 AND
     (($2::uuid IS NOT NULL AND o.id=$2::uuid) OR
      ($2::uuid IS NULL AND o.order_number=$3))
   UNION ALL
   SELECT t.sales_order_id, t.id::bigint, s.organization_id, t.store_id, t.pos_terminal_id, 1
   FROM pos_transactions t JOIN stores s ON s.id=t.store_id
   WHERE s.organization_id=$1 AND $2::uuid IS NULL AND t.transaction_number=$3
 ), sale AS (
   SELECT * FROM sale_candidates ORDER BY priority LIMIT 1
 ), saved_document AS (
   SELECT c.qr_code_base64, c.id AS sort_id, 0 AS source_priority
   FROM zatca_document_chain c JOIN sale s ON s.organization_id=c.organization_id
   WHERE NULLIF(c.qr_code_base64,'') IS NOT NULL
     AND ((c.entity_type='sales_order_v2' AND c.entity_id=s.order_id::text)
       OR (c.entity_type='pos_transaction' AND
           (c.entity_id=s.transaction_id::text OR EXISTS (
              SELECT 1 FROM pos_transactions t
              WHERE t.sales_order_id=s.order_id AND t.id::text=c.entity_id
                AND t.store_id=s.store_id AND t.pos_terminal_id=s.pos_terminal_id
           ))))
   UNION ALL
   -- POS checkout also persists the signed order QR in transaction metadata.
   -- That snapshot can be available before the document-chain sync reaches
   -- the database serving receipt printing.
   SELECT NULLIF(t.metadata->>'qr_code_base64',''), t.id::bigint, 1
   FROM pos_transactions t JOIN sale s ON t.id::bigint=s.transaction_id
   WHERE NULLIF(t.metadata->>'qr_code_base64','') IS NOT NULL
   UNION ALL
   SELECT NULLIF(t.metadata->>'qr_code_base64',''), t.id::bigint, 1
   FROM pos_transactions t JOIN sale s ON t.sales_order_id=s.order_id
   WHERE NULLIF(t.metadata->>'qr_code_base64','') IS NOT NULL
     AND t.store_id=s.store_id AND t.pos_terminal_id=s.pos_terminal_id
   ORDER BY source_priority, sort_id DESC LIMIT 1
 ), reported_document AS (
   SELECT c.qr_code_base64
   FROM zatca_document_chain c JOIN sale s ON s.organization_id=c.organization_id
   WHERE c.zatca_status='reported' AND c.zatca_response->>'reportingStatus'='REPORTED'
     AND ((c.entity_type='sales_order_v2' AND c.entity_id=s.order_id::text)
       OR (c.entity_type='pos_transaction' AND
           (c.entity_id=s.transaction_id::text OR EXISTS (
              SELECT 1 FROM pos_transactions t
              WHERE t.sales_order_id=s.order_id AND t.id::text=c.entity_id
                AND t.store_id=s.store_id AND t.pos_terminal_id=s.pos_terminal_id
           ))))
   ORDER BY c.id DESC LIMIT 1
 )
 SELECT COALESCE((SELECT qr_code_base64 FROM reported_document),
                 (SELECT qr_code_base64 FROM saved_document),''),
        EXISTS(SELECT 1 FROM reported_document),
        EXISTS(SELECT 1 FROM zatca_device_configs d
               WHERE d.organization_id=s.organization_id AND d.store_id=s.store_id
                 AND d.pos_terminal_id=s.pos_terminal_id AND d.device_type='pos'
                 AND d.is_active=true AND d.is_revoked=false)
 FROM sale s`, organizationID, orderID, reference).Scan(&state.QRCode, &state.Reported, &state.RegisteredPOS)
	return state, err
}
