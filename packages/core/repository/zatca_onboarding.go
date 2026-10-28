package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// CompleteZatcaOnboarding saves the credential pair and stage atomically. The
// compliance token guard prevents a concurrent re-onboarding from receiving an
// unrelated reporting certificate. Preserve unrelated metadata and secrets.
func (q *Queries) CompleteZatcaOnboarding(ctx context.Context, id int32, complianceToken, productionToken string, expiry time.Time, metadata json.RawMessage, environment string) error {
	result, err := q.db.Exec(ctx, `/* CompleteZatcaOnboarding */
 UPDATE zatca_device_configs SET production_csid=$3, csid_expiry=$4,
 metadata=COALESCE(metadata, '{}'::jsonb) || $5::jsonb, updated_at=now()
 WHERE id=$1 AND compliance_csid=$2 AND is_active=true AND is_revoked=false
 AND COALESCE(production_csid, '') IN ('', $2) AND zatca_env=$6`, id, complianceToken, productionToken, expiry, metadata, environment)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("device changed during onboarding; reload its configuration before retrying")
	}
	return nil
}

func (q *Queries) SaveRenewedZatcaCredentials(ctx context.Context, id int32, currentToken, newToken string, expiry time.Time, metadata json.RawMessage, environment string) error {
	result, err := q.db.Exec(ctx, `/* SaveRenewedZatcaCredentials */ UPDATE zatca_device_configs
 SET production_csid=$3,csid_expiry=$4,metadata=COALESCE(metadata,'{}'::jsonb)||$5::jsonb,updated_at=now()
 WHERE id=$1 AND production_csid=$2 AND is_active=true AND is_revoked=false AND zatca_env=$6`, id, currentToken, newToken, expiry, metadata, environment)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("device changed during credential renewal")
	}
	return nil
}
