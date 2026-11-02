package repository

import "context"

// GetReportedSalesOrderQR returns the saved signed QR only after ZATCA has
// reported the document belonging to this organization and persisted order.
// The reporting API confirms acceptance but does not return a new QR.
func (q *Queries) GetReportedSalesOrderQR(ctx context.Context, organizationID int32, orderID string) (string, error) {
	var qr string
	err := q.db.QueryRow(ctx, `SELECT COALESCE((
 SELECT NULLIF(c.qr_code_base64,'')
 FROM zatca_document_chain c
 WHERE c.entity_type='sales_order_v2' AND c.entity_id=o.id::text
 AND c.organization_id=o.organization_id
 AND c.zatca_status='reported'
 AND c.zatca_response->>'reportingStatus'='REPORTED'
 ORDER BY c.id DESC LIMIT 1), '')
 FROM sales_orders_v2 o WHERE o.id=$1::uuid AND o.organization_id=$2`, orderID, organizationID).Scan(&qr)
	return qr, err
}
