package usecase

import (
	"context"
	"encoding/base64"
	"log"
	"time"

	"github.com/NasTecSol/nembus-core/biometrics"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

type BiometricUseCase struct {
	repo *repository.Queries
}

func NewBiometricUseCase() *BiometricUseCase {
	return &BiometricUseCase{}
}

func (uc *BiometricUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

type EnrollFingerInput struct {
	UserID           int32    `json:"user_id"`
	FingerIndex      int16    `json:"finger_index"`
	FMDBase64        string   `json:"fmd_base64"`
	ViewFMDsBase64   []string `json:"view_fmds_base64,omitempty"`
	QualityScore     int16    `json:"quality_score,omitempty"`
	ReaderSerial     string   `json:"reader_serial,omitempty"`
	EnrolledByUserID *int32   `json:"enrolled_by_user_id,omitempty"`
}

type EnrolledFingerSummary struct {
	ID               int32     `json:"id"`
	OrganizationID   int32     `json:"organization_id"`
	UserID           int32     `json:"user_id"`
	FingerIndex      int16     `json:"finger_index"`
	FingerName       string    `json:"finger_name"`
	FMDFormat        string    `json:"fmd_format"`
	QualityScore     int16     `json:"quality_score"`
	ReaderSerial     string    `json:"reader_serial,omitempty"`
	EnrolledByUserID *int32    `json:"enrolled_by_user_id,omitempty"`
	IsActive         bool      `json:"is_active"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// EnrollFinger validates and consolidates an ISO/IEC 19794-2:2005 FMD, runs a 1:N
// cross-user duplicate check across the organization, and upserts the template.
func (uc *BiometricUseCase) EnrollFinger(ctx context.Context, input EnrollFingerInput) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if input.UserID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "valid user_id is required", nil)
	}
	if input.FingerIndex < 1 || input.FingerIndex > 10 {
		return utils.NewResponse(utils.CodeBadReq, "finger_index must be between 1 and 10 (NIST finger position)", nil)
	}

	targetUser, err := uc.repo.GetUser(ctx, input.UserID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "target user not found", nil)
	}

	var probeFMD []byte
	if len(input.ViewFMDsBase64) > 0 {
		views := make([][]byte, 0, len(input.ViewFMDsBase64))
		for _, vb64 := range input.ViewFMDsBase64 {
			vb, decErr := biometrics.ValidateAndDecodeFMD(vb64)
			if decErr != nil {
				return utils.NewResponse(utils.CodeBadReq, "invalid view FMD in view_fmds_base64: "+decErr.Error(), nil)
			}
			views = append(views, vb)
		}
		consolidated, consErr := biometrics.ConsolidateEnrollmentFMD(views)
		if consErr != nil {
			return utils.NewResponse(utils.CodeBadReq, "failed to consolidate enrollment FMDs: "+consErr.Error(), nil)
		}
		probeFMD = consolidated
		input.FMDBase64 = base64.StdEncoding.EncodeToString(consolidated)
	} else {
		decoded, decErr := biometrics.ValidateAndDecodeFMD(input.FMDBase64)
		if decErr != nil {
			return utils.NewResponse(utils.CodeBadReq, decErr.Error(), nil)
		}
		probeFMD = decoded
		input.FMDBase64 = base64.StdEncoding.EncodeToString(decoded)
	}

	if !biometrics.IsEngineAvailable() {
		return utils.NewResponse(utils.CodeError, biometrics.ErrEngineUnavailable.Error(), nil)
	}

	// 1:N Organization-wide duplicate check (prevent shared fingerprint across users)
	orgTemplates, err := uc.repo.ListActiveBiometricTemplatesByOrg(ctx, targetUser.OrganizationID)
	if err == nil && len(orgTemplates) > 0 {
		candidates := make([][]byte, 0, len(orgTemplates))
		validRows := make([]repository.ListActiveBiometricTemplatesByOrgRow, 0, len(orgTemplates))
		for _, row := range orgTemplates {
			if b, decErr := biometrics.ValidateAndDecodeFMD(row.FmdBase64); decErr == nil {
				candidates = append(candidates, b)
				validRows = append(validRows, row)
			}
		}
		if len(candidates) > 0 {
			bestIdx, score, idErr := biometrics.Identify(probeFMD, candidates, biometrics.DefaultMaxFalseMatchScore)
			if idErr == nil && bestIdx >= 0 && bestIdx < len(validRows) {
				matchedRow := validRows[bestIdx]
				if matchedRow.UserID != input.UserID {
					log.Printf("[BIOMETRICS] Duplicate fingerprint blocked: user_id=%d matched existing user_id=%d (%s) score=%d",
						input.UserID, matchedRow.UserID, matchedRow.Username, score)
					return utils.NewResponse(utils.CodeConflict,
						"this fingerprint is already enrolled to another user ("+matchedRow.Username+")", nil)
				}
			}
		}
	}

	quality := input.QualityScore
	if quality < 1 || quality > 5 {
		quality = 1
	}

	var enrolledBy pgtype.Int4
	if input.EnrolledByUserID != nil && *input.EnrolledByUserID > 0 {
		enrolledBy = pgtype.Int4{Int32: *input.EnrolledByUserID, Valid: true}
	}
	var readerSerial pgtype.Text
	if input.ReaderSerial != "" {
		readerSerial = pgtype.Text{String: input.ReaderSerial, Valid: true}
	}

	saved, err := uc.repo.UpsertUserBiometricTemplate(ctx, repository.UpsertUserBiometricTemplateParams{
		OrganizationID:   targetUser.OrganizationID,
		UserID:           input.UserID,
		FingerIndex:      input.FingerIndex,
		FmdFormat:        "ISO_19794_2_2005",
		FmdBase64:        input.FMDBase64,
		QualityScore:     pgtype.Int2{Int16: quality, Valid: true},
		ReaderSerial:     readerSerial,
		EnrolledByUserID: enrolledBy,
		Column9:          []byte("{}"),
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to save biometric template: "+err.Error(), nil)
	}

	log.Printf("[BIOMETRICS] Enrolled finger %d (%s) for user_id=%d (%s) by actor=%v",
		input.FingerIndex, biometrics.FingerName(input.FingerIndex), input.UserID, targetUser.Username, input.EnrolledByUserID)

	return utils.NewResponse(utils.CodeOK, "biometric fingerprint enrolled successfully", toEnrolledFingerSummary(saved))
}

func (uc *BiometricUseCase) ListUserBiometrics(ctx context.Context, userID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if userID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "valid user_id is required", nil)
	}
	rows, err := uc.repo.ListUserBiometricTemplates(ctx, userID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list biometric templates: "+err.Error(), nil)
	}
	summaries := make([]EnrolledFingerSummary, 0, len(rows))
	for _, r := range rows {
		var enrolledBy *int32
		if r.EnrolledByUserID.Valid {
			v := r.EnrolledByUserID.Int32
			enrolledBy = &v
		}
		summaries = append(summaries, EnrolledFingerSummary{
			ID:               r.ID,
			OrganizationID:   r.OrganizationID,
			UserID:           r.UserID,
			FingerIndex:      r.FingerIndex,
			FingerName:       biometrics.FingerName(r.FingerIndex),
			FMDFormat:        r.FmdFormat,
			QualityScore:     r.QualityScore.Int16,
			ReaderSerial:     r.ReaderSerial.String,
			EnrolledByUserID: enrolledBy,
			IsActive:         r.IsActive,
			UpdatedAt:        r.UpdatedAt.Time,
		})
	}
	return utils.NewResponse(utils.CodeOK, "user biometrics fetched successfully", summaries)
}

func (uc *BiometricUseCase) DeleteUserBiometric(ctx context.Context, userID int32, fingerIndex int16) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if userID <= 0 || fingerIndex < 1 || fingerIndex > 10 {
		return utils.NewResponse(utils.CodeBadReq, "valid user_id and finger_index (1..10) are required", nil)
	}
	_, err := uc.repo.DeactivateUserBiometricTemplate(ctx, repository.DeactivateUserBiometricTemplateParams{
		UserID:      userID,
		FingerIndex: fingerIndex,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "enrolled finger template not found", nil)
	}
	log.Printf("[BIOMETRICS] Deactivated finger %d (%s) for user_id=%d",
		fingerIndex, biometrics.FingerName(fingerIndex), userID)
	return utils.NewResponse(utils.CodeOK, "biometric template removed", nil)
}

func toEnrolledFingerSummary(t repository.UserBiometricTemplate) EnrolledFingerSummary {
	var enrolledBy *int32
	if t.EnrolledByUserID.Valid {
		v := t.EnrolledByUserID.Int32
		enrolledBy = &v
	}
	return EnrolledFingerSummary{
		ID:               t.ID,
		OrganizationID:   t.OrganizationID,
		UserID:           t.UserID,
		FingerIndex:      t.FingerIndex,
		FingerName:       biometrics.FingerName(t.FingerIndex),
		FMDFormat:        t.FmdFormat,
		QualityScore:     t.QualityScore.Int16,
		ReaderSerial:     t.ReaderSerial.String,
		EnrolledByUserID: enrolledBy,
		IsActive:         t.IsActive,
		UpdatedAt:        t.UpdatedAt.Time,
	}
}

