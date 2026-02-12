package usecase

import (
	"context"
	"encoding/json"
	"strconv"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// CashierOutput is the response shape for cashier APIs
type CashierOutput struct {
	ID            int32            `json:"id"`
	UserID        int32            `json:"user_id"`
	StoreID       int32            `json:"store_id"`
	CashierCode   string           `json:"cashier_code"`
	DrawerLimit   pgtype.Numeric   `json:"drawer_limit"`
	DiscountLimit pgtype.Numeric   `json:"discount_limit"`
	IsActive      pgtype.Bool      `json:"is_active"`
	Metadata      json.RawMessage  `json:"metadata"`
	CreatedAt     pgtype.Timestamp `json:"created_at"`
}

func cashierToOutput(c repository.Cashier) CashierOutput {
	return CashierOutput{
		ID:            c.ID,
		UserID:        c.UserID,
		StoreID:       c.StoreID,
		CashierCode:   c.CashierCode,
		DrawerLimit:   c.DrawerLimit,
		DiscountLimit: c.DiscountLimit,
		IsActive:      c.IsActive,
		Metadata:      utils.BytesToJSONRawMessage(c.Metadata),
		CreatedAt:     c.CreatedAt,
	}
}

type CashierUseCase struct {
	repo *repository.Queries
}

// NewCashierUseCase creates a new cashier use case without repository
func NewCashierUseCase() *CashierUseCase {
	return &CashierUseCase{}
}

// SetRepository sets the repository for this request
func (uc *CashierUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// CreateCashier creates a new cashier
func (uc *CashierUseCase) CreateCashier(
	ctx context.Context,
	userID int32,
	storeID int32,
	cashierCode string,
	drawerLimit *pgtype.Numeric,
	discountLimit *pgtype.Numeric,
	isActive bool,
	metadata []byte,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if cashierCode == "" {
		return utils.NewResponse(utils.CodeBadReq, "cashier code cannot be empty", nil)
	}

	// Check if code already exists for this store
	exists, err := uc.repo.CashierCodeExists(ctx, repository.CashierCodeExistsParams{
		CashierCode: cashierCode,
		StoreID:     storeID,
	})
	if err == nil && exists {
		return utils.NewResponse(utils.CodeBadReq, "cashier code already exists for this store", nil)
	}

	var drawerLimitVal pgtype.Numeric
	if drawerLimit != nil {
		drawerLimitVal = *drawerLimit
	}

	var discountLimitVal pgtype.Numeric
	if discountLimit != nil {
		discountLimitVal = *discountLimit
	}

	var metaBytes []byte
	if metadata != nil {
		metaBytes = metadata
	} else {
		metaBytes = []byte("{}")
	}

	cashier, err := uc.repo.CreateCashier(ctx, repository.CreateCashierParams{
		UserID:        userID,
		StoreID:       storeID,
		CashierCode:   cashierCode,
		DrawerLimit:   drawerLimitVal,
		DiscountLimit: discountLimitVal,
		IsActive:      pgtype.Bool{Bool: isActive, Valid: true},
		Metadata:      metaBytes,
	})

	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "cashier created successfully", cashierToOutput(cashier))
}

// CreateCashierWithDefaults creates a new cashier with default values
func (uc *CashierUseCase) CreateCashierWithDefaults(
	ctx context.Context,
	userID int32,
	storeID int32,
	cashierCode string,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if cashierCode == "" {
		return utils.NewResponse(utils.CodeBadReq, "cashier code cannot be empty", nil)
	}

	// Check if code already exists for this store
	exists, err := uc.repo.CashierCodeExists(ctx, repository.CashierCodeExistsParams{
		CashierCode: cashierCode,
		StoreID:     storeID,
	})
	if err == nil && exists {
		return utils.NewResponse(utils.CodeBadReq, "cashier code already exists for this store", nil)
	}

	cashier, err := uc.repo.CreateCashierWithDefaults(ctx, repository.CreateCashierWithDefaultsParams{
		UserID:      userID,
		StoreID:     storeID,
		CashierCode: cashierCode,
	})

	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "cashier created successfully", cashierToOutput(cashier))
}

// GetCashierByID retrieves a cashier by ID
func (uc *CashierUseCase) GetCashierByID(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashierID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid cashier id", nil)
	}

	cashier, err := uc.repo.GetCashierByID(ctx, int32(cashierID))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cashier not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier fetched successfully", cashierToOutput(cashier))
}

// GetCashierByCode retrieves a cashier by code and store ID
func (uc *CashierUseCase) GetCashierByCode(ctx context.Context, code string, storeID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "cashier code is required", nil)
	}

	cashier, err := uc.repo.GetCashierByCode(ctx, repository.GetCashierByCodeParams{
		CashierCode: code,
		StoreID:     storeID,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cashier not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier fetched successfully", cashierToOutput(cashier))
}

// GetCashierByUserID retrieves a cashier by user ID and store ID
func (uc *CashierUseCase) GetCashierByUserID(ctx context.Context, userID int32, storeID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashier, err := uc.repo.GetCashierByUserID(ctx, repository.GetCashierByUserIDParams{
		UserID:  userID,
		StoreID: storeID,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cashier not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier fetched successfully", cashierToOutput(cashier))
}

// ListAllCashiers lists all cashiers
func (uc *CashierUseCase) ListAllCashiers(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashiers, err := uc.repo.ListAllCashiers(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]CashierOutput, len(cashiers))
	for i := range cashiers {
		out[i] = cashierToOutput(cashiers[i])
	}

	return utils.NewResponse(utils.CodeOK, "cashiers fetched successfully", out)
}

// ListActiveCashiers lists only active cashiers
func (uc *CashierUseCase) ListActiveCashiers(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashiers, err := uc.repo.ListActiveCashiers(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]CashierOutput, len(cashiers))
	for i := range cashiers {
		out[i] = cashierToOutput(cashiers[i])
	}

	return utils.NewResponse(utils.CodeOK, "active cashiers fetched successfully", out)
}

// ListCashiersByStore lists all cashiers for a specific store
func (uc *CashierUseCase) ListCashiersByStore(ctx context.Context, storeID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashiers, err := uc.repo.ListCashiersByStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]CashierOutput, len(cashiers))
	for i := range cashiers {
		out[i] = cashierToOutput(cashiers[i])
	}

	return utils.NewResponse(utils.CodeOK, "cashiers fetched successfully", out)
}

// ListActiveCashiersByStore lists active cashiers for a specific store
func (uc *CashierUseCase) ListActiveCashiersByStore(ctx context.Context, storeID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashiers, err := uc.repo.ListActiveCashiersByStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]CashierOutput, len(cashiers))
	for i := range cashiers {
		out[i] = cashierToOutput(cashiers[i])
	}

	return utils.NewResponse(utils.CodeOK, "active cashiers fetched successfully", out)
}

// ListCashiersWithPagination lists cashiers with pagination
func (uc *CashierUseCase) ListCashiersWithPagination(ctx context.Context, limit, offset int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashiers, err := uc.repo.ListCashiersWithPagination(ctx, repository.ListCashiersWithPaginationParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]CashierOutput, len(cashiers))
	for i := range cashiers {
		out[i] = cashierToOutput(cashiers[i])
	}

	return utils.NewResponse(utils.CodeOK, "cashiers fetched successfully", out)
}

// CountCashiers counts total number of cashiers
func (uc *CashierUseCase) CountCashiers(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	count, err := uc.repo.CountCashiers(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier count fetched successfully", count)
}

// CountActiveCashiers counts total number of active cashiers
func (uc *CashierUseCase) CountActiveCashiers(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	count, err := uc.repo.CountActiveCashiers(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "active cashier count fetched successfully", count)
}

// CountCashiersByStore counts cashiers in a specific store
func (uc *CashierUseCase) CountCashiersByStore(ctx context.Context, storeID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	count, err := uc.repo.CountCashiersByStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier count fetched successfully", count)
}

// UpdateCashier updates cashier information
func (uc *CashierUseCase) UpdateCashier(
	ctx context.Context,
	id string,
	userID *int32,
	storeID *int32,
	cashierCode *string,
	drawerLimit *pgtype.Numeric,
	discountLimit *pgtype.Numeric,
	isActive *bool,
	metadata []byte,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashierID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid cashier id", nil)
	}

	// Check if cashier exists
	exists, err := uc.repo.CashierExists(ctx, int32(cashierID))
	if err != nil || !exists {
		return utils.NewResponse(utils.CodeNotFound, "cashier not found", nil)
	}

	// Get current cashier to preserve values if not provided
	currentCashier, err := uc.repo.GetCashierByID(ctx, int32(cashierID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	// If code is being updated, check if it already exists
	if cashierCode != nil {
		exists, err := uc.repo.CashierCodeExistsExcludingID(ctx, repository.CashierCodeExistsExcludingIDParams{
			CashierCode: *cashierCode,
			StoreID:     currentCashier.StoreID,
			ID:          int32(cashierID),
		})
		if err == nil && exists {
			return utils.NewResponse(utils.CodeBadReq, "cashier code already exists for this store", nil)
		}
	}

	params := repository.UpdateCashierParams{
		ID: int32(cashierID),
	}

	if userID != nil {
		params.UserID = *userID
	} else {
		params.UserID = currentCashier.UserID
	}

	if storeID != nil {
		params.StoreID = *storeID
	} else {
		params.StoreID = currentCashier.StoreID
	}

	if cashierCode != nil {
		params.CashierCode = *cashierCode
	} else {
		params.CashierCode = currentCashier.CashierCode
	}

	if drawerLimit != nil {
		params.DrawerLimit = *drawerLimit
	} else {
		params.DrawerLimit = currentCashier.DrawerLimit
	}

	if discountLimit != nil {
		params.DiscountLimit = *discountLimit
	} else {
		params.DiscountLimit = currentCashier.DiscountLimit
	}

	if isActive != nil {
		params.IsActive = pgtype.Bool{Bool: *isActive, Valid: true}
	} else {
		params.IsActive = currentCashier.IsActive
	}

	if metadata != nil {
		params.Metadata = metadata
	} else {
		params.Metadata = currentCashier.Metadata
	}

	cashier, err := uc.repo.UpdateCashier(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier updated successfully", cashierToOutput(cashier))
}

// UpdateCashierLimits updates cashier drawer and discount limits
func (uc *CashierUseCase) UpdateCashierLimits(
	ctx context.Context,
	id string,
	drawerLimit pgtype.Numeric,
	discountLimit pgtype.Numeric,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashierID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid cashier id", nil)
	}

	cashier, err := uc.repo.UpdateCashierLimits(ctx, repository.UpdateCashierLimitsParams{
		ID:            int32(cashierID),
		DrawerLimit:   drawerLimit,
		DiscountLimit: discountLimit,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier limits updated successfully", cashierToOutput(cashier))
}

// UpdateCashierDrawerLimit updates only the drawer limit
func (uc *CashierUseCase) UpdateCashierDrawerLimit(ctx context.Context, id string, drawerLimit pgtype.Numeric) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashierID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid cashier id", nil)
	}

	cashier, err := uc.repo.UpdateCashierDrawerLimit(ctx, repository.UpdateCashierDrawerLimitParams{
		ID:          int32(cashierID),
		DrawerLimit: drawerLimit,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier drawer limit updated successfully", cashierToOutput(cashier))
}

// UpdateCashierDiscountLimit updates only the discount limit
func (uc *CashierUseCase) UpdateCashierDiscountLimit(ctx context.Context, id string, discountLimit pgtype.Numeric) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashierID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid cashier id", nil)
	}

	cashier, err := uc.repo.UpdateCashierDiscountLimit(ctx, repository.UpdateCashierDiscountLimitParams{
		ID:            int32(cashierID),
		DiscountLimit: discountLimit,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier discount limit updated successfully", cashierToOutput(cashier))
}

// UpdateCashierMetadata updates cashier metadata
func (uc *CashierUseCase) UpdateCashierMetadata(ctx context.Context, id string, metadata []byte) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashierID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid cashier id", nil)
	}

	cashier, err := uc.repo.UpdateCashierMetadata(ctx, repository.UpdateCashierMetadataParams{
		ID:       int32(cashierID),
		Metadata: metadata,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier metadata updated successfully", cashierToOutput(cashier))
}

// ActivateCashier activates a cashier
func (uc *CashierUseCase) ActivateCashier(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashierID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid cashier id", nil)
	}

	cashier, err := uc.repo.ActivateCashier(ctx, int32(cashierID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier activated successfully", cashierToOutput(cashier))
}

// DeactivateCashier deactivates a cashier
func (uc *CashierUseCase) DeactivateCashier(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashierID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid cashier id", nil)
	}

	cashier, err := uc.repo.DeactivateCashier(ctx, int32(cashierID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier deactivated successfully", cashierToOutput(cashier))
}

// DeleteCashier hard deletes a cashier
func (uc *CashierUseCase) DeleteCashier(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashierID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid cashier id", nil)
	}

	err = uc.repo.DeleteCashier(ctx, int32(cashierID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier deleted successfully", nil)
}

// SoftDeleteCashier soft deletes a cashier by deactivating it
func (uc *CashierUseCase) SoftDeleteCashier(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashierID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid cashier id", nil)
	}

	cashier, err := uc.repo.SoftDeleteCashier(ctx, int32(cashierID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier soft deleted successfully", cashierToOutput(cashier))
}

// CashierExists checks if a cashier exists by ID
func (uc *CashierUseCase) CashierExists(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashierID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid cashier id", nil)
	}

	exists, err := uc.repo.CashierExists(ctx, int32(cashierID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier existence checked", exists)
}

// CashierCodeExists checks if a cashier code already exists for a store
func (uc *CashierUseCase) CashierCodeExists(ctx context.Context, code string, storeID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "cashier code is required", nil)
	}

	exists, err := uc.repo.CashierCodeExists(ctx, repository.CashierCodeExistsParams{
		CashierCode: code,
		StoreID:     storeID,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier code existence checked", exists)
}

// GetCashierWithLimits gets cashier with limits and user details
func (uc *CashierUseCase) GetCashierWithLimits(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashierID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid cashier id", nil)
	}

	result, err := uc.repo.GetCashierWithLimits(ctx, int32(cashierID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "cashier with limits fetched successfully", result)
}

// ListActiveCashiersInStore lists active cashiers in a store with session info
func (uc *CashierUseCase) ListActiveCashiersInStore(ctx context.Context, storeID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	cashiers, err := uc.repo.ListActiveCashiersInStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "active cashiers fetched successfully", cashiers)
}
