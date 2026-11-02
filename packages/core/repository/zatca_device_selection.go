package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
)

func (q *Queries) CountActiveZatcaDevices(ctx context.Context, organizationID int32, terminalID pgtype.Int4, deviceType string) (int64, error) {
	var count int64
	err := q.db.QueryRow(ctx, `/* CountActiveZatcaDevices */ SELECT COUNT(*) FROM zatca_device_configs
 WHERE organization_id=$1 AND device_type=$3 AND is_active=true AND is_revoked=false
 AND ($3='cloud' OR pos_terminal_id=$2)`, organizationID, terminalID, deviceType).Scan(&count)
	return count, err
}
