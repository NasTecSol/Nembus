package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/NasTecSol/nembus-core/biometrics"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// syncedBiometricRow mirrors JSON payloads from sync_queue / StreamPull.
type syncedBiometricRow struct {
	ID               int32     `json:"id"`
	OrganizationID   int32     `json:"organization_id"`
	UserID           int32     `json:"user_id"`
	FingerIndex      int16     `json:"finger_index"`
	FMDFormat        string    `json:"fmd_format"`
	FMDBase64        string    `json:"fmd_base64"`
	QualityScore     *int16    `json:"quality_score"`
	ReaderSerial     *string   `json:"reader_serial"`
	EnrolledByUserID *int32    `json:"enrolled_by_user_id"`
	IsActive         bool      `json:"is_active"`
	Metadata         []byte    `json:"metadata"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ApplyUserBiometricTemplateFromSync upserts a template pushed from POS with org-wide duplicate
// detection and last-write-wins on updated_at.
func ApplyUserBiometricTemplateFromSync(ctx context.Context, pool *pgxpool.Pool, payload []byte) error {
	if pool == nil {
		return fmt.Errorf("database pool is nil")
	}
	var row syncedBiometricRow
	if err := json.Unmarshal(payload, &row); err != nil {
		return fmt.Errorf("invalid user_biometric_templates payload: %w", err)
	}
	if row.UserID <= 0 || row.FingerIndex < 1 || row.FingerIndex > 10 {
		return fmt.Errorf("invalid user_id or finger_index in biometric sync payload")
	}
	if row.OrganizationID <= 0 {
		return fmt.Errorf("organization_id is required in biometric sync payload")
	}

	probeFMD, err := biometrics.ValidateAndDecodeFMD(row.FMDBase64)
	if err != nil {
		return err
	}
	if !biometrics.IsEngineAvailable() {
		return biometrics.ErrEngineUnavailable
	}

	q := repository.New(pool)
	orgTemplates, err := q.ListActiveBiometricTemplatesByOrg(ctx, row.OrganizationID)
	if err == nil && len(orgTemplates) > 0 {
		candidates := make([][]byte, 0, len(orgTemplates))
		validRows := make([]repository.ListActiveBiometricTemplatesByOrgRow, 0, len(orgTemplates))
		for _, t := range orgTemplates {
			if t.UserID == row.UserID && t.FingerIndex == row.FingerIndex {
				continue
			}
			if b, decErr := biometrics.ValidateAndDecodeFMD(t.FmdBase64); decErr == nil {
				candidates = append(candidates, b)
				validRows = append(validRows, t)
			}
		}
		if len(candidates) > 0 {
			bestIdx, _, idErr := biometrics.Identify(probeFMD, candidates, biometrics.DefaultMaxFalseMatchScore)
			if idErr == nil && bestIdx >= 0 && bestIdx < len(validRows) {
				matched := validRows[bestIdx]
				if matched.UserID != row.UserID {
					return fmt.Errorf("%w: matched user_id=%d (%s)",
						biometrics.ErrDuplicateBiometricOtherUser, matched.UserID, matched.Username)
				}
			}
		}
	}

	fmdFormat := row.FMDFormat
	if fmdFormat == "" {
		fmdFormat = "ISO_19794_2_2005"
	}
	quality := int16(1)
	if row.QualityScore != nil && *row.QualityScore >= 1 && *row.QualityScore <= 5 {
		quality = *row.QualityScore
	}
	meta := row.Metadata
	if len(meta) == 0 {
		meta = []byte("{}")
	}
	updatedAt := row.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	createdAt := row.CreatedAt
	if createdAt.IsZero() {
		createdAt = updatedAt
	}

	var enrolledBy any
	if row.EnrolledByUserID != nil && *row.EnrolledByUserID > 0 {
		enrolledBy = *row.EnrolledByUserID
	}
	var readerSerial any
	if row.ReaderSerial != nil && *row.ReaderSerial != "" {
		readerSerial = *row.ReaderSerial
	}

	const upsertSQL = `
INSERT INTO user_biometric_templates (
	organization_id, user_id, finger_index, fmd_format, fmd_base64,
	quality_score, reader_serial, enrolled_by_user_id, is_active, metadata,
	created_at, updated_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11, $12
)
ON CONFLICT (user_id, finger_index) DO UPDATE SET
	organization_id = EXCLUDED.organization_id,
	fmd_format = EXCLUDED.fmd_format,
	fmd_base64 = EXCLUDED.fmd_base64,
	quality_score = EXCLUDED.quality_score,
	reader_serial = EXCLUDED.reader_serial,
	enrolled_by_user_id = EXCLUDED.enrolled_by_user_id,
	is_active = EXCLUDED.is_active,
	metadata = EXCLUDED.metadata,
	updated_at = EXCLUDED.updated_at
WHERE EXCLUDED.updated_at >= user_biometric_templates.updated_at;
`
	tag, err := pool.Exec(ctx, upsertSQL,
		row.OrganizationID, row.UserID, row.FingerIndex, fmdFormat, row.FMDBase64,
		quality, readerSerial, enrolledBy, row.IsActive, string(meta),
		createdAt, updatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert user_biometric_templates: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil
	}
	return nil
}
