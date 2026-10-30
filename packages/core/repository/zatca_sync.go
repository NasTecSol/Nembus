package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ZatcaSyncSnapshot always carries the owning device so database-local numeric
// IDs can be remapped on the receiver. Credentials must travel over HTTPS only.
type ZatcaSyncSnapshot struct {
	Device   json.RawMessage `json:"device"`
	Document json.RawMessage `json:"document,omitempty"`
}

type ZatcaRestorePage struct {
	Device    json.RawMessage   `json:"device"`
	Documents []json.RawMessage `json:"documents"`
	HeadICV   int64             `json:"head_icv"`
	NextICV   int64             `json:"next_icv"`
	HasMore   bool              `json:"has_more"`
}

// ApplyZatcaSnapshot persists a POS snapshot atomically. A source ID never
// overwrites an unrelated cloud device or document with the same integer ID.
func (q *Queries) ApplyZatcaSnapshot(ctx context.Context, storeID, callerOrgID int32, snapshot ZatcaSyncSnapshot) error {
	var source ZatcaDeviceConfig
	if err := json.Unmarshal(snapshot.Device, &source); err != nil {
		return fmt.Errorf("invalid ZATCA device payload")
	}
	if source.ID <= 0 || source.OrganizationID <= 0 || source.DeviceSerial == "" || source.DeviceType != "pos" || !source.StoreID.Valid || source.StoreID.Int32 != storeID || !source.PosTerminalID.Valid || !source.UpdatedAt.Valid {
		return fmt.Errorf("a POS device with matching store, terminal, organization and timestamp is required")
	}
	if callerOrgID > 0 && callerOrgID != source.OrganizationID {
		return fmt.Errorf("device organization does not match authentication")
	}
	if source.ZatcaEnv.String != "simulation" && source.ZatcaEnv.String != "production" && source.ZatcaEnv.String != "sandbox" {
		return fmt.Errorf("invalid device environment")
	}
	if source.PrivateKeyPem == "" || !source.ComplianceCsid.Valid {
		return fmt.Errorf("device signing credentials are missing")
	}
	return q.InZatcaTransaction(ctx, func(tx *Queries) error {
		var validScope bool
		if err := tx.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM stores s JOIN pos_terminals t ON t.store_id=s.id WHERE s.id=$1 AND s.organization_id=$2 AND t.id=$3)`, storeID, source.OrganizationID, source.PosTerminalID.Int32).Scan(&validScope); err != nil {
			return err
		}
		if !validScope {
			return fmt.Errorf("device store or terminal does not belong to this organization")
		}
		if _, err := tx.db.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, fmt.Sprintf("zatca-sync:%d:%s", source.OrganizationID, source.DeviceSerial)); err != nil {
			return err
		}
		existing, err := tx.GetZatcaDeviceConfigBySerial(ctx, GetZatcaDeviceConfigBySerialParams{OrganizationID: source.OrganizationID, DeviceSerial: source.DeviceSerial})
		var deviceID int32
		if errors.Is(err, pgx.ErrNoRows) {
			err = tx.db.QueryRow(ctx, `INSERT INTO zatca_device_configs(organization_id,store_id,pos_terminal_id,device_serial,device_type,csr_pem,private_key_pem,compliance_csid,production_csid,csid_expiry,zatca_env,is_active,is_revoked,revoked_at,revoked_reason,metadata,created_at,updated_at)
 SELECT organization_id,store_id,pos_terminal_id,device_serial,device_type,csr_pem,private_key_pem,compliance_csid,production_csid,csid_expiry,zatca_env,is_active,is_revoked,revoked_at,revoked_reason,
 COALESCE(metadata,'{}'::jsonb)||jsonb_build_object('_pos_source_updated_at',$2::text,'signing_owner','pos'),created_at,NOW()
 FROM json_populate_record(NULL::zatca_device_configs,$1::json) RETURNING id`, string(snapshot.Device), source.UpdatedAt.Time.Format(time.RFC3339Nano)).Scan(&deviceID)
		} else if err == nil {
			deviceID = existing.ID
			if strings.Join(strings.Fields(existing.PrivateKeyPem), "") != strings.Join(strings.Fields(source.PrivateKeyPem), "") || existing.ZatcaEnv != source.ZatcaEnv || existing.OrganizationID != source.OrganizationID || existing.StoreID != source.StoreID || existing.PosTerminalID != source.PosTerminalID {
				return fmt.Errorf("existing device identity, signing key, environment or location differs; credentials were not overwritten")
			}
			if existing.IsRevoked.Bool && !source.IsRevoked.Bool {
				return fmt.Errorf("cloud device is revoked and cannot be reactivated by POS sync")
			}
			var meta struct {
				Version time.Time `json:"_pos_source_updated_at"`
			}
			_ = json.Unmarshal(existing.Metadata, &meta)
			ready := existing.ProductionCsid.String != "" && existing.ProductionCsid.String != existing.ComplianceCsid.String
			incomingReady := source.ProductionCsid.String != "" && source.ProductionCsid.String != source.ComplianceCsid.String
			if !source.UpdatedAt.Time.Before(meta.Version) && (!ready || incomingReady) {
				_, err = tx.db.Exec(ctx, `UPDATE zatca_device_configs d SET csr_pem=i.csr_pem,compliance_csid=i.compliance_csid,production_csid=i.production_csid,csid_expiry=i.csid_expiry,is_active=i.is_active,is_revoked=i.is_revoked,revoked_at=i.revoked_at,revoked_reason=i.revoked_reason,
 metadata=COALESCE(i.metadata,'{}'::jsonb)||jsonb_build_object('_pos_source_updated_at',$3::text,'signing_owner','pos'),updated_at=NOW()
 FROM json_populate_record(NULL::zatca_device_configs,$2::json) i WHERE d.id=$1`, deviceID, string(snapshot.Device), source.UpdatedAt.Time.Format(time.RFC3339Nano))
			}
		}
		if err != nil {
			return fmt.Errorf("could not save ZATCA device snapshot: %w", err)
		}
		if len(snapshot.Document) == 0 || string(snapshot.Document) == "null" {
			return nil
		}
		var doc struct {
			DeviceID int32  `json:"device_config_id"`
			OrgID    int32  `json:"organization_id"`
			ICV      int64  `json:"icv"`
			UUID     string `json:"zatca_uuid"`
			PIH      string `json:"pih"`
			Hash     string `json:"xml_hash"`
		}
		if json.Unmarshal(snapshot.Document, &doc) != nil || doc.DeviceID != source.ID || doc.OrgID != source.OrganizationID || doc.ICV < 1 || doc.Hash == "" || doc.PIH == "" {
			return fmt.Errorf("invalid document identity or device binding")
		}
		if _, err := uuid.Parse(doc.UUID); err != nil {
			return fmt.Errorf("invalid document UUID")
		}
		if doc.ICV > 1 {
			var previousHash string
			if err := tx.db.QueryRow(ctx, `SELECT xml_hash FROM zatca_document_chain WHERE device_config_id=$1 AND icv=$2`, deviceID, doc.ICV-1).Scan(&previousHash); err != nil {
				return fmt.Errorf("sync the preceding device-chain entry first")
			}
			if previousHash != doc.PIH {
				return fmt.Errorf("document PIH does not match the preceding device-chain hash")
			}
		}
		var savedID int64
		err = tx.db.QueryRow(ctx, `INSERT INTO zatca_document_chain(entity_type,entity_id,device_config_id,organization_id,zatca_uuid,icv,pih,xml_hash,zatca_status,zatca_response,qr_code_base64,signed_xml,submitted_at,cleared_at,created_at,updated_at)
 SELECT entity_type,entity_id,$2,organization_id,zatca_uuid,icv,pih,xml_hash,zatca_status,zatca_response,qr_code_base64,signed_xml,submitted_at,cleared_at,created_at,NOW()
 FROM json_populate_record(NULL::zatca_document_chain,$1::json)
 ON CONFLICT(device_config_id,icv) DO UPDATE SET
 zatca_status=CASE WHEN zatca_document_chain.zatca_status IN ('reported','cleared','warning') AND EXCLUDED.zatca_status NOT IN ('reported','cleared','warning') THEN zatca_document_chain.zatca_status ELSE EXCLUDED.zatca_status END,
 zatca_response=CASE WHEN zatca_document_chain.zatca_status IN ('reported','cleared','warning') AND EXCLUDED.zatca_status NOT IN ('reported','cleared','warning') THEN zatca_document_chain.zatca_response ELSE EXCLUDED.zatca_response END,
 qr_code_base64=COALESCE(EXCLUDED.qr_code_base64,zatca_document_chain.qr_code_base64),submitted_at=COALESCE(EXCLUDED.submitted_at,zatca_document_chain.submitted_at),cleared_at=COALESCE(EXCLUDED.cleared_at,zatca_document_chain.cleared_at),updated_at=NOW()
 WHERE zatca_document_chain.zatca_uuid=EXCLUDED.zatca_uuid AND zatca_document_chain.xml_hash=EXCLUDED.xml_hash AND zatca_document_chain.pih=EXCLUDED.pih
 AND zatca_document_chain.entity_type=EXCLUDED.entity_type AND zatca_document_chain.entity_id=EXCLUDED.entity_id AND zatca_document_chain.signed_xml IS NOT DISTINCT FROM EXCLUDED.signed_xml
 RETURNING id`, string(snapshot.Document), deviceID).Scan(&savedID)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("a different signed document already occupies this device ICV; chain was not overwritten")
		}
		return err
	})
}

// ZatcaRecoveryPage returns one physical device's registration and a bounded,
// ICV-ordered chain page. The caller authenticates tenant/organization access.
func (q *Queries) ZatcaRecoveryPage(ctx context.Context, serial string, orgID int32, after, through int64, limit int32) (ZatcaRestorePage, error) {
	page := ZatcaRestorePage{Documents: make([]json.RawMessage, 0), NextICV: after}
	rows, err := q.db.Query(ctx, `SELECT id FROM zatca_device_configs WHERE device_type='pos' AND device_serial=$1 AND ($2::int=0 OR organization_id=$2) ORDER BY id LIMIT 2`, serial, orgID)
	if err != nil {
		return page, err
	}
	var ids []int32
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return page, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return page, err
	}
	if len(ids) == 0 {
		return page, pgx.ErrNoRows
	}
	if len(ids) > 1 {
		return page, fmt.Errorf("device serial belongs to multiple organizations; organization_id is required")
	}
	if err := q.db.QueryRow(ctx, `SELECT row_to_json(d),COALESCE((SELECT MAX(icv) FROM zatca_document_chain WHERE device_config_id=d.id),0) FROM zatca_device_configs d WHERE d.id=$1`, ids[0]).Scan(&page.Device, &page.HeadICV); err != nil {
		return page, err
	}
	var ownership struct {
		Metadata struct {
			Owner string `json:"signing_owner"`
		} `json:"metadata"`
	}
	if json.Unmarshal(page.Device, &ownership) != nil || ownership.Metadata.Owner != "pos" {
		return page, fmt.Errorf("device has not been backed up by its owning POS; sync that POS before recovery")
	}
	if through >= 0 && through < page.HeadICV {
		page.HeadICV = through
	}
	if after > page.HeadICV {
		return page, fmt.Errorf("restore cursor exceeds the device-chain head")
	}
	rows, err = q.db.Query(ctx, `SELECT row_to_json(c),icv FROM zatca_document_chain c WHERE device_config_id=$1 AND icv>$2 AND icv<=$3 ORDER BY icv LIMIT $4`, ids[0], after, page.HeadICV, limit+1)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		var raw json.RawMessage
		var icv int64
		if err := rows.Scan(&raw, &icv); err != nil {
			return page, err
		}
		if len(page.Documents) == int(limit) {
			page.HasMore = true
			break
		}
		page.Documents = append(page.Documents, raw)
		page.NextICV = icv
	}
	return page, rows.Err()
}

// RestoreZatcaDevice imports a bounded HTTPS recovery stream in one transaction.
// No partial device or chain becomes available for signing if a page fails.
func (q *Queries) RestoreZatcaDevice(ctx context.Context, serial string, orgID int32, fetch func(int64, int64) (ZatcaRestorePage, error)) error {
	page, err := fetch(0, -1)
	if err != nil {
		return err
	}
	var device struct {
		ID      int32  `json:"id"`
		OrgID   int32  `json:"organization_id"`
		StoreID int32  `json:"store_id"`
		Serial  string `json:"device_serial"`
	}
	if json.Unmarshal(page.Device, &device) != nil || device.ID <= 0 || device.Serial != serial || (orgID > 0 && device.OrgID != orgID) {
		return fmt.Errorf("recovery device identity does not match the requested device")
	}
	return q.InZatcaTransaction(ctx, func(tx *Queries) error {
		if _, err := tx.db.Exec(ctx, `SELECT set_config('nembus.sync_origin','cloud',true)`); err != nil {
			return err
		}
		if err := tx.ApplyZatcaSnapshot(ctx, device.StoreID, device.OrgID, ZatcaSyncSnapshot{Device: page.Device}); err != nil {
			return err
		}
		local, err := tx.GetZatcaDeviceConfigBySerial(ctx, GetZatcaDeviceConfigBySerialParams{OrganizationID: device.OrgID, DeviceSerial: serial})
		if err != nil {
			return err
		}
		var localHead int64
		if err := tx.db.QueryRow(ctx, `SELECT COALESCE(MAX(icv),0) FROM zatca_document_chain WHERE device_config_id=$1`, local.ID).Scan(&localHead); err != nil {
			return err
		}
		head := page.HeadICV
		if localHead > head {
			return fmt.Errorf("recovery would roll back the local invoice counter; local state was preserved")
		}
		for {
			for _, document := range page.Documents {
				if err := tx.ApplyZatcaSnapshot(ctx, device.StoreID, device.OrgID, ZatcaSyncSnapshot{Device: page.Device, Document: document}); err != nil {
					return fmt.Errorf("restore ZATCA chain: %w", err)
				}
			}
			if !page.HasMore {
				if page.NextICV != head {
					return fmt.Errorf("incomplete ZATCA recovery chain")
				}
				break
			}
			after := page.NextICV
			page, err = fetch(after, head)
			if err != nil {
				return err
			}
			var nextDevice struct {
				ID     int32  `json:"id"`
				Serial string `json:"device_serial"`
			}
			if json.Unmarshal(page.Device, &nextDevice) != nil || nextDevice.ID != device.ID || nextDevice.Serial != serial || page.HeadICV != head || page.NextICV < after || (page.HasMore && page.NextICV == after) {
				return fmt.Errorf("ZATCA device or cursor changed during recovery")
			}
		}
		_, err = tx.db.Exec(ctx, `INSERT INTO sync_queue(entity_type,entity_id,action,payload,status,synced_at)
 SELECT 'zatca_device_configs',d.id,'UPDATE',jsonb_build_object('id',d.id,'device_serial',d.device_serial,'updated_at',d.updated_at),'synced',NOW() FROM zatca_device_configs d WHERE d.id=$1
 AND NOT EXISTS(SELECT 1 FROM sync_queue q WHERE q.entity_type='zatca_device_configs' AND q.entity_id=d.id)`, local.ID)
		if err != nil {
			return err
		}
		_, err = tx.db.Exec(ctx, `INSERT INTO sync_queue(entity_type,entity_id,action,payload,status,synced_at)
 SELECT 'zatca_document_chain',c.id,'UPDATE',jsonb_build_object('id',c.id,'device_config_id',c.device_config_id,'icv',c.icv,'updated_at',c.updated_at),'synced',NOW() FROM zatca_document_chain c WHERE c.device_config_id=$1
 AND NOT EXISTS(SELECT 1 FROM sync_queue q WHERE q.entity_type='zatca_document_chain' AND q.entity_id=c.id)`, local.ID)
		return err
	})
}
