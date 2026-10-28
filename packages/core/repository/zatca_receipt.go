package repository

import "context"

// GetZatcaReceiptQR resolves a saved signed POS document by sale or order ID.
// No QR data is reconstructed from mutable receipt totals or the current clock.
func (q *Queries) GetZatcaReceiptQR(ctx context.Context, transactionID int32, orderID string) (string, error) {
	var qr string
	err := q.db.QueryRow(ctx, `/* GetZatcaReceiptQR */
 SELECT COALESCE(NULLIF(c.qr_code_base64,''), NULLIF(o.qr_code_base64,''), t.metadata->>'qr_code_base64', '')
 FROM pos_transactions t
 LEFT JOIN zatca_document_chain c ON c.entity_type='pos_transaction' AND c.entity_id=t.id::text
 LEFT JOIN zatca_document_chain o ON o.entity_type='sales_order_v2' AND o.entity_id=t.sales_order_id::text
 WHERE ($1::int > 0 AND t.id=$1) OR ($1::int=0 AND t.sales_order_id::text=$2)
 ORDER BY t.id DESC LIMIT 1`, transactionID, orderID).Scan(&qr)
	return qr, err
}

func (q *Queries) GetZatcaSalesOrderQR(ctx context.Context, orderID string) (string, error) {
	var qr string
	err := q.db.QueryRow(ctx, `/* GetZatcaSalesOrderQR */
 SELECT COALESCE(
 (SELECT c.qr_code_base64 FROM zatca_document_chain c WHERE c.entity_type='pos_transaction'
  AND c.entity_id IN (SELECT t.id::text FROM pos_transactions t WHERE t.sales_order_id=o.id) ORDER BY c.id LIMIT 1),
 (SELECT c.qr_code_base64 FROM zatca_document_chain c WHERE c.entity_type='sales_order_v2' AND c.entity_id=o.id::text ORDER BY c.id LIMIT 1),'')
 FROM sales_orders_v2 o WHERE o.id::text=$1`, orderID).Scan(&qr)
	return qr, err
}
