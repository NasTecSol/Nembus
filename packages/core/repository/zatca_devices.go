package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

// ZatcaRegisteredDevice deliberately contains no certificates, keys, or metadata.
type ZatcaRegisteredDevice struct {
	ID                  int32              `json:"id"`
	OrganizationID      int32              `json:"organization_id"`
	StoreID             int32              `json:"store_id"`
	StoreName           string             `json:"store_name"`
	PosTerminalID       pgtype.Int4        `json:"pos_terminal_id"`
	TerminalName        pgtype.Text        `json:"terminal_name"`
	TerminalCode        pgtype.Text        `json:"terminal_code"`
	DeviceSerial        string             `json:"device_serial"`
	DeviceType          string             `json:"device_type"`
	ZatcaEnv            pgtype.Text        `json:"zatca_env"`
	IsActive            bool               `json:"is_active"`
	IsRevoked           bool               `json:"is_revoked"`
	OrganizationEnabled bool               `json:"organization_zatca_enabled"`
	ReportingReady      bool               `json:"reporting_ready"`
	CsidExpiry          pgtype.Timestamptz `json:"csid_expiry"`
	CreatedAt           pgtype.Timestamp   `json:"created_at"`
	UpdatedAt           pgtype.Timestamp   `json:"updated_at"`
}

// ListRegisteredZatcaDevices returns saved registrations, including incomplete,
// inactive and revoked ones, scoped to the authenticated organization and store.
// A zero terminal ID means all terminals in the requested store.
func (q *Queries) ListRegisteredZatcaDevices(ctx context.Context, orgID, storeID, terminalID int32) ([]ZatcaRegisteredDevice, error) {
	if orgID <= 0 || storeID <= 0 || terminalID < 0 {
		return nil, fmt.Errorf("valid organization and store are required")
	}
	rows, err := q.db.Query(ctx, `/* ListRegisteredZatcaDevices */
 SELECT d.id, d.organization_id, d.store_id, s.name,
 d.pos_terminal_id, t.terminal_name, t.terminal_code,
 d.device_serial, d.device_type, d.zatca_env,
 COALESCE(d.is_active,false), COALESCE(d.is_revoked,false),
 COALESCE(o.metadata #> '{zatca,enabled}' = 'true'::jsonb,false) AS organization_zatca_enabled,
 COALESCE(o.metadata #> '{zatca,enabled}' = 'true'::jsonb
   AND d.is_active AND NOT COALESCE(d.is_revoked,false)
   AND COALESCE(d.production_csid,'') <> ''
   AND d.production_csid IS DISTINCT FROM d.compliance_csid
   AND COALESCE(d.metadata->>'production_secret','') <> ''
   AND (d.csid_expiry IS NULL OR d.csid_expiry > now()),false) AS reporting_ready,
 d.csid_expiry, d.created_at, d.updated_at
 FROM zatca_device_configs d
 JOIN stores s ON s.id=d.store_id AND s.organization_id=d.organization_id
 JOIN organizations o ON o.id=d.organization_id
 LEFT JOIN pos_terminals t ON t.id=d.pos_terminal_id AND t.store_id=d.store_id
 WHERE d.organization_id=$1 AND d.store_id=$2
   AND ($3::int=0 OR d.pos_terminal_id=$3)
 ORDER BY d.pos_terminal_id NULLS LAST, d.id DESC`, orgID, storeID, terminalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	devices := make([]ZatcaRegisteredDevice, 0)
	for rows.Next() {
		var d ZatcaRegisteredDevice
		if err := rows.Scan(&d.ID, &d.OrganizationID, &d.StoreID, &d.StoreName,
			&d.PosTerminalID, &d.TerminalName, &d.TerminalCode, &d.DeviceSerial,
			&d.DeviceType, &d.ZatcaEnv, &d.IsActive, &d.IsRevoked,
			&d.OrganizationEnabled, &d.ReportingReady, &d.CsidExpiry, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}
