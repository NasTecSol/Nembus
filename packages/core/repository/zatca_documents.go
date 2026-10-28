package repository

import "context"

// ListZatcaDocumentsPage reads all saved document types and statuses in the
// request tenant, restricted to the authenticated organization when supplied.
func (q *Queries) ListZatcaDocumentsPage(ctx context.Context, organizationID int32, limit int32, offset int64) ([]ZatcaDocumentChain, int64, error) {
	var total int64
	if err := q.db.QueryRow(ctx, `/* CountZatcaDocuments */
 SELECT count(*) FROM zatca_document_chain WHERE ($1::int=0 OR organization_id=$1)`, organizationID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := q.db.Query(ctx, `/* ListZatcaDocumentsPage */
 SELECT id,entity_type,entity_id,device_config_id,organization_id,zatca_uuid,icv,pih,xml_hash,zatca_status,zatca_response,qr_code_base64,signed_xml,submitted_at,cleared_at,created_at,updated_at
 FROM zatca_document_chain WHERE ($1::int=0 OR organization_id=$1)
 ORDER BY id DESC LIMIT $2 OFFSET $3`, organizationID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	entries := make([]ZatcaDocumentChain, 0)
	for rows.Next() {
		var c ZatcaDocumentChain
		if err := rows.Scan(&c.ID, &c.EntityType, &c.EntityID, &c.DeviceConfigID, &c.OrganizationID, &c.ZatcaUuid, &c.Icv, &c.Pih, &c.XmlHash, &c.ZatcaStatus, &c.ZatcaResponse, &c.QrCodeBase64, &c.SignedXml, &c.SubmittedAt, &c.ClearedAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		entries = append(entries, c)
	}
	return entries, total, rows.Err()
}
