package repository

import (
	"context"
	"encoding/json"
)

// Atomic nested update preserves seller data, other ZATCA options, and all
// unrelated organization metadata. No certificate or document rows are changed.
func (q *Queries) SetOrganizationZatcaEnabled(ctx context.Context, id int32, enabled bool) (json.RawMessage, error) {
	var metadata json.RawMessage
	err := q.db.QueryRow(ctx, `/* SetOrganizationZatcaEnabled */ UPDATE organizations
 SET metadata=jsonb_set(COALESCE(NULLIF(metadata,'null'::jsonb),'{}'::jsonb),'{zatca}',
 COALESCE(NULLIF(metadata->'zatca','null'::jsonb),'{}'::jsonb)||jsonb_build_object('enabled',$2::boolean),true),updated_at=now()
 WHERE id=$1 RETURNING metadata`, id, enabled).Scan(&metadata)
	return metadata, err
}

// Filter eligibility before LIMIT so disabled organizations cannot starve
// enabled organizations' reporting jobs. The worker rechecks before dispatch.
func (q *Queries) ListEnabledPendingZatcaDocuments(ctx context.Context, limit int32) ([]ZatcaDocumentChain, error) {
	rows, err := q.db.Query(ctx, `/* ListEnabledPendingZatcaDocuments */
 SELECT c.id,c.entity_type,c.entity_id,c.device_config_id,c.organization_id,c.zatca_uuid,c.icv,c.pih,c.xml_hash,c.zatca_status,c.zatca_response,c.qr_code_base64,c.signed_xml,c.submitted_at,c.cleared_at,c.created_at,c.updated_at
 FROM zatca_document_chain c JOIN organizations o ON o.id=c.organization_id
 JOIN zatca_device_configs d ON d.id=c.device_config_id
 WHERE c.zatca_status IN ('pending','failed') AND o.metadata#>'{zatca,enabled}'='true'::jsonb
 AND c.entity_type IN ('sales_order_v2','pos_transaction','sales_return')
 AND d.is_active=true AND d.is_revoked=false AND COALESCE(d.production_csid,'')<>''
 AND d.production_csid IS DISTINCT FROM d.compliance_csid
 AND (d.csid_expiry IS NULL OR d.csid_expiry>now())
 ORDER BY c.created_at LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []ZatcaDocumentChain
	for rows.Next() {
		var c ZatcaDocumentChain
		if err := rows.Scan(&c.ID, &c.EntityType, &c.EntityID, &c.DeviceConfigID, &c.OrganizationID, &c.ZatcaUuid, &c.Icv, &c.Pih, &c.XmlHash, &c.ZatcaStatus, &c.ZatcaResponse, &c.QrCodeBase64, &c.SignedXml, &c.SubmittedAt, &c.ClearedAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, c)
	}
	return entries, rows.Err()
}
