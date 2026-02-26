package usecase

import (
	"context"
	"encoding/json"
	"strconv"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// LoyaltyRuleOutput is the response shape for loyalty rule APIs
type LoyaltyRuleOutput struct {
	ID                   int32            `json:"id"`
	OrganizationID       int32            `json:"organization_id"`
	RuleName             string           `json:"rule_name"`
	PointsEarningRate    pgtype.Numeric   `json:"points_earning_rate"`
	PointsRedemptionRate pgtype.Numeric   `json:"points_redemption_rate"`
	MinPointsToRedeem    pgtype.Numeric   `json:"min_points_to_redeem"`
	MaxPointsPerTxn      pgtype.Numeric   `json:"max_points_per_txn"`
	MaxRedemptionPercent pgtype.Numeric   `json:"max_redemption_percent"`
	EligibleProductTypes []string         `json:"eligible_product_types"`
	ExpiryDays           pgtype.Int4      `json:"expiry_days"`
	IsActive             pgtype.Bool      `json:"is_active"`
	ValidFrom            pgtype.Date      `json:"valid_from"`
	ValidTo              pgtype.Date      `json:"valid_to"`
	Metadata             json.RawMessage  `json:"metadata"`
	CreatedAt            pgtype.Timestamp `json:"created_at"`
	UpdatedAt            pgtype.Timestamp `json:"updated_at"`
}

func loyaltyRuleToOutput(r repository.LoyaltyRedemptionRule) LoyaltyRuleOutput {
	return LoyaltyRuleOutput{
		ID:                   r.ID,
		OrganizationID:       r.OrganizationID,
		RuleName:             r.RuleName,
		PointsEarningRate:    r.PointsEarningRate,
		PointsRedemptionRate: r.PointsRedemptionRate,
		MinPointsToRedeem:    r.MinPointsToRedeem,
		MaxPointsPerTxn:      r.MaxPointsPerTxn,
		MaxRedemptionPercent: r.MaxRedemptionPercent,
		EligibleProductTypes: r.EligibleProductTypes,
		ExpiryDays:           r.ExpiryDays,
		IsActive:             r.IsActive,
		ValidFrom:            r.ValidFrom,
		ValidTo:              r.ValidTo,
		Metadata:             utils.BytesToJSONRawMessage(r.Metadata),
		CreatedAt:            r.CreatedAt,
		UpdatedAt:            r.UpdatedAt,
	}
}

// LoyaltyUseCase handles business logic for loyalty redemption rules
type LoyaltyUseCase struct {
	repo *repository.Queries
}

func NewLoyaltyUseCase() *LoyaltyUseCase {
	return &LoyaltyUseCase{}
}

func (uc *LoyaltyUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// ---- Create ----

type CreateLoyaltyRuleInput struct {
	OrganizationID       int32
	RuleName             string
	PointsEarningRate    *pgtype.Numeric
	PointsRedemptionRate *pgtype.Numeric
	MinPointsToRedeem    *pgtype.Numeric
	MaxPointsPerTxn      *pgtype.Numeric
	MaxRedemptionPercent *pgtype.Numeric
	EligibleProductTypes []string
	ExpiryDays           *int32
	IsActive             *bool
	ValidFrom            *string // "2006-01-02"
	ValidTo              *string
	Metadata             []byte
}

func (uc *LoyaltyUseCase) CreateLoyaltyRule(ctx context.Context, in CreateLoyaltyRuleInput) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if in.OrganizationID == 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}
	if in.RuleName == "" {
		return utils.NewResponse(utils.CodeBadReq, "rule_name is required", nil)
	}

	params := repository.CreateLoyaltyRuleParams{
		OrganizationID:       in.OrganizationID,
		RuleName:             in.RuleName,
		EligibleProductTypes: in.EligibleProductTypes,
	}
	if in.PointsEarningRate != nil {
		params.PointsEarningRate = *in.PointsEarningRate
	}
	if in.PointsRedemptionRate != nil {
		params.PointsRedemptionRate = *in.PointsRedemptionRate
	}
	if in.MinPointsToRedeem != nil {
		params.MinPointsToRedeem = *in.MinPointsToRedeem
	}
	if in.MaxPointsPerTxn != nil {
		params.MaxPointsPerTxn = *in.MaxPointsPerTxn
	}
	if in.MaxRedemptionPercent != nil {
		params.MaxRedemptionPercent = *in.MaxRedemptionPercent
	}
	if in.ExpiryDays != nil {
		params.ExpiryDays = pgtype.Int4{Int32: *in.ExpiryDays, Valid: true}
	}
	if in.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *in.IsActive, Valid: true}
	} else {
		params.IsActive = pgtype.Bool{Bool: true, Valid: true}
	}
	if in.ValidFrom != nil && *in.ValidFrom != "" {
		var d pgtype.Date
		if err := d.Scan(*in.ValidFrom); err == nil {
			params.ValidFrom = d
		}
	}
	if in.ValidTo != nil && *in.ValidTo != "" {
		var d pgtype.Date
		if err := d.Scan(*in.ValidTo); err == nil {
			params.ValidTo = d
		}
	}
	if in.Metadata != nil {
		params.Metadata = in.Metadata
	} else {
		params.Metadata = []byte("{}")
	}

	rule, err := uc.repo.CreateLoyaltyRule(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "loyalty rule created successfully", loyaltyRuleToOutput(rule))
}

// ---- Get ----

func (uc *LoyaltyUseCase) GetLoyaltyRule(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	ruleID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid rule id", nil)
	}
	rule, err := uc.repo.GetLoyaltyRule(ctx, int32(ruleID))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "loyalty rule not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "loyalty rule fetched successfully", loyaltyRuleToOutput(rule))
}

// ---- GetActive ----

func (uc *LoyaltyUseCase) GetActiveLoyaltyRule(ctx context.Context, organizationID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if organizationID == 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}
	rule, err := uc.repo.GetActiveLoyaltyRule(ctx, organizationID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "no active loyalty rule found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "active loyalty rule fetched successfully", loyaltyRuleToOutput(rule))
}

// ---- List ----

func (uc *LoyaltyUseCase) ListLoyaltyRules(ctx context.Context, organizationID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if organizationID == 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}
	rules, err := uc.repo.ListLoyaltyRules(ctx, organizationID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]LoyaltyRuleOutput, len(rules))
	for i, r := range rules {
		out[i] = loyaltyRuleToOutput(r)
	}
	return utils.NewResponse(utils.CodeOK, "loyalty rules fetched successfully", out)
}

// ---- Update ----

type UpdateLoyaltyRuleInput struct {
	RuleName             string
	PointsEarningRate    *pgtype.Numeric
	PointsRedemptionRate *pgtype.Numeric
	MinPointsToRedeem    *pgtype.Numeric
	MaxPointsPerTxn      *pgtype.Numeric
	MaxRedemptionPercent *pgtype.Numeric
	IsActive             *bool
	ValidFrom            *string
	ValidTo              *string
	Metadata             []byte
}

func (uc *LoyaltyUseCase) UpdateLoyaltyRule(ctx context.Context, id string, in UpdateLoyaltyRuleInput) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	ruleID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid rule id", nil)
	}

	current, err := uc.repo.GetLoyaltyRule(ctx, int32(ruleID))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "loyalty rule not found", nil)
	}

	params := repository.UpdateLoyaltyRuleParams{
		ID:                   int32(ruleID),
		RuleName:             current.RuleName,
		PointsEarningRate:    current.PointsEarningRate,
		PointsRedemptionRate: current.PointsRedemptionRate,
		MinPointsToRedeem:    current.MinPointsToRedeem,
		MaxPointsPerTxn:      current.MaxPointsPerTxn,
		MaxRedemptionPercent: current.MaxRedemptionPercent,
		IsActive:             current.IsActive,
		ValidFrom:            current.ValidFrom,
		ValidTo:              current.ValidTo,
		Metadata:             current.Metadata,
	}
	if in.RuleName != "" {
		params.RuleName = in.RuleName
	}
	if in.PointsEarningRate != nil {
		params.PointsEarningRate = *in.PointsEarningRate
	}
	if in.PointsRedemptionRate != nil {
		params.PointsRedemptionRate = *in.PointsRedemptionRate
	}
	if in.MinPointsToRedeem != nil {
		params.MinPointsToRedeem = *in.MinPointsToRedeem
	}
	if in.MaxPointsPerTxn != nil {
		params.MaxPointsPerTxn = *in.MaxPointsPerTxn
	}
	if in.MaxRedemptionPercent != nil {
		params.MaxRedemptionPercent = *in.MaxRedemptionPercent
	}
	if in.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *in.IsActive, Valid: true}
	}
	if in.ValidFrom != nil && *in.ValidFrom != "" {
		var d pgtype.Date
		if err := d.Scan(*in.ValidFrom); err == nil {
			params.ValidFrom = d
		}
	}
	if in.ValidTo != nil && *in.ValidTo != "" {
		var d pgtype.Date
		if err := d.Scan(*in.ValidTo); err == nil {
			params.ValidTo = d
		}
	}
	if in.Metadata != nil {
		params.Metadata = in.Metadata
	}

	rule, err := uc.repo.UpdateLoyaltyRule(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "loyalty rule updated successfully", loyaltyRuleToOutput(rule))
}

// ---- Toggle Active ----

func (uc *LoyaltyUseCase) ToggleLoyaltyRuleActive(ctx context.Context, id string, isActive bool) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	ruleID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid rule id", nil)
	}
	rule, err := uc.repo.ToggleLoyaltyRuleActive(ctx, repository.ToggleLoyaltyRuleActiveParams{
		ID:       int32(ruleID),
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "loyalty rule status updated successfully", loyaltyRuleToOutput(rule))
}

// ---- Delete ----

func (uc *LoyaltyUseCase) DeleteLoyaltyRule(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	ruleID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid rule id", nil)
	}
	if err := uc.repo.DeleteLoyaltyRule(ctx, int32(ruleID)); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "loyalty rule deleted successfully", nil)
}

// ---- Customer loyalty points ----

func (uc *LoyaltyUseCase) AdjustCustomerLoyaltyPoints(ctx context.Context, customerID string, points pgtype.Numeric) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	id, err := strconv.ParseInt(customerID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid customer id", nil)
	}
	customer, err := uc.repo.AdjustCustomerLoyaltyPoints(ctx, repository.AdjustCustomerLoyaltyPointsParams{
		ID:            int32(id),
		LoyaltyPoints: points,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "loyalty points adjusted successfully", customerToOutput(customer))
}

func (uc *LoyaltyUseCase) GetCustomerLoyaltyBalance(ctx context.Context, customerID string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	id, err := strconv.ParseInt(customerID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid customer id", nil)
	}
	data, err := uc.repo.GetCustomerLoyaltyBalance(ctx, int32(id))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "customer not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "loyalty balance fetched successfully", data)
}
