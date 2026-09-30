package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// StockCountLineInput represents payload for a stock count line item.
type StockCountLineInput struct {
	ProductID         int32                  `json:"product_id"`
	ProductVariantID  *int32                 `json:"product_variant_id,omitempty"`
	StorageLocationID *int32                 `json:"storage_location_id,omitempty"`
	ExpectedQuantity  *float64               `json:"expected_quantity,omitempty"`
	SystemQuantity    *float64               `json:"system_quantity,omitempty"`
	CountedQuantity   float64                `json:"counted_quantity"`
	Variance          *float64               `json:"variance,omitempty"`
	VarianceValue     *float64               `json:"variance_value,omitempty"`
	CountedAt         *string                `json:"counted_at,omitempty"`
	UomID             *int32                 `json:"uom_id,omitempty"`
	BatchNumber       *string                `json:"batch_number,omitempty"`
	SerialNumber      *string                `json:"serial_number,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// BulkStockCountLineItem represents a line item in a bulk update operation.
type BulkStockCountLineItem struct {
	ID              int32                  `json:"id"`
	CountedQuantity float64                `json:"counted_quantity"`
	Variance        *float64               `json:"variance,omitempty"`
	VarianceValue   *float64               `json:"variance_value,omitempty"`
	CountedAt       *string                `json:"counted_at,omitempty"`
	BatchNumber     *string                `json:"batch_number,omitempty"`
	SerialNumber    *string                `json:"serial_number,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// CreateStockCountInput represents payload for creating a stock count.
type CreateStockCountInput struct {
	CountNumber       *string                `json:"count_number,omitempty"`
	StoreID           int32                  `json:"store_id"`
	StorageLocationID *int32                 `json:"storage_location_id,omitempty"`
	CountType         *string                `json:"count_type,omitempty"`
	Status            *string                `json:"status,omitempty"`
	ScheduledDate     *string                `json:"scheduled_date,omitempty"`
	CountedBy         *int32                 `json:"counted_by,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	Lines             []StockCountLineInput  `json:"lines,omitempty"`
}

// UpdateStockCountInput represents payload for updating a stock count header.
type UpdateStockCountInput struct {
	StorageLocationID *int32                 `json:"storage_location_id,omitempty"`
	CountType         *string                `json:"count_type,omitempty"`
	Status            *string                `json:"status,omitempty"`
	ScheduledDate     *string                `json:"scheduled_date,omitempty"`
	CountedBy         *int32                 `json:"counted_by,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateStockCountLineInput represents payload for updating an individual stock count line.
type UpdateStockCountLineInput struct {
	CountedQuantity float64                `json:"counted_quantity"`
	Variance        *float64               `json:"variance,omitempty"`
	VarianceValue   *float64               `json:"variance_value,omitempty"`
	CountedAt       *string                `json:"counted_at,omitempty"`
	BatchNumber     *string                `json:"batch_number,omitempty"`
	SerialNumber    *string                `json:"serial_number,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ApproveStockCountInput represents payload for approving a stock count.
type ApproveStockCountInput struct {
	ApprovedBy int32 `json:"approved_by"`
}

// StockCountFilter represents query parameters for listing stock counts.
type StockCountFilter struct {
	StoreID   *int32  `json:"store_id,omitempty"`
	Status    *string `json:"status,omitempty"`
	CountType *string `json:"count_type,omitempty"`
	FromDate  *string `json:"from_date,omitempty"`
	ToDate    *string `json:"to_date,omitempty"`
	Search    *string `json:"search,omitempty"`
	Page      int32   `json:"page"`
	Limit     int32   `json:"limit"`
}

// StockCountLineOutput represents line item details.
type StockCountLineOutput struct {
	ID                  int32            `json:"id"`
	StockCountID        int32            `json:"stock_count_id"`
	ProductID           int32            `json:"product_id"`
	ProductName         pgtype.Text      `json:"product_name"`
	ProductSKU          pgtype.Text      `json:"product_sku"`
	ProductVariantID    pgtype.Int4      `json:"product_variant_id"`
	VariantName         pgtype.Text      `json:"variant_name"`
	VariantSKU          pgtype.Text      `json:"variant_sku"`
	StorageLocationID   pgtype.Int4      `json:"storage_location_id"`
	StorageLocationName pgtype.Text      `json:"storage_location_name"`
	ExpectedQuantity    pgtype.Numeric   `json:"expected_quantity"`
	SystemQuantity      pgtype.Numeric   `json:"system_quantity"`
	CountedQuantity     pgtype.Numeric   `json:"counted_quantity"`
	Variance            pgtype.Numeric   `json:"variance"`
	VarianceValue       pgtype.Numeric   `json:"variance_value"`
	CountedAt           pgtype.Timestamp `json:"counted_at"`
	UomID               pgtype.Int4      `json:"uom_id"`
	UomName             pgtype.Text      `json:"uom_name"`
	BatchNumber         pgtype.Text      `json:"batch_number"`
	SerialNumber        pgtype.Text      `json:"serial_number"`
	Metadata            json.RawMessage  `json:"metadata"`
	CreatedAt           pgtype.Timestamp `json:"created_at"`
	UpdatedAt           pgtype.Timestamp `json:"updated_at"`
}

// StockCountSummaryOutput represents summary variance calculations.
type StockCountSummaryOutput struct {
	TotalLines         int64          `json:"total_lines"`
	LinesWithVariance  int64          `json:"lines_with_variance"`
	TotalVarianceValue pgtype.Numeric `json:"total_variance_value"`
	PositiveVariance   pgtype.Numeric `json:"positive_variance"`
	NegativeVariance   pgtype.Numeric `json:"negative_variance"`
}

// StockCountOutput represents stock count with details and lines.
type StockCountOutput struct {
	ID                  int32                    `json:"id"`
	CountNumber         string                   `json:"count_number"`
	StoreID             int32                    `json:"store_id"`
	StoreName           pgtype.Text              `json:"store_name"`
	StorageLocationID   pgtype.Int4              `json:"storage_location_id"`
	StorageLocationName pgtype.Text              `json:"storage_location_name"`
	CountType           pgtype.Text              `json:"count_type"`
	Status              pgtype.Text              `json:"status"`
	ScheduledDate       pgtype.Date              `json:"scheduled_date"`
	StartedAt           pgtype.Timestamp         `json:"started_at"`
	CompletedAt         pgtype.Timestamp         `json:"completed_at"`
	CountedBy           pgtype.Int4              `json:"counted_by"`
	CountedByName       pgtype.Text              `json:"counted_by_name"`
	ApprovedBy          pgtype.Int4              `json:"approved_by"`
	ApprovedByName      pgtype.Text              `json:"approved_by_name"`
	Metadata            json.RawMessage          `json:"metadata"`
	CreatedAt           pgtype.Timestamp         `json:"created_at"`
	UpdatedAt           pgtype.Timestamp         `json:"updated_at"`
	TotalLines          int64                    `json:"total_lines"`
	LinesWithVariance   int64                    `json:"lines_with_variance"`
	TotalVarianceValue  pgtype.Numeric           `json:"total_variance_value"`
	Summary             *StockCountSummaryOutput `json:"summary,omitempty"`
	Lines               []StockCountLineOutput   `json:"lines,omitempty"`
}

// StockCountListResponse represents paginated stock counts list response.
type StockCountListResponse struct {
	Data       []StockCountOutput `json:"data"`
	TotalCount int64              `json:"total_count"`
	Page       int32              `json:"page"`
	Limit      int32              `json:"limit"`
	TotalPages int32              `json:"total_pages"`
}

// StockCountsUseCase handles stock count business logic.
type StockCountsUseCase struct {
	repo *repository.Queries
}

// NewStockCountsUseCase creates a new StockCountsUseCase instance.
func NewStockCountsUseCase() *StockCountsUseCase {
	return &StockCountsUseCase{}
}

// SetRepository sets repository queries.
func (uc *StockCountsUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *StockCountsUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

// CreateStockCount creates a new stock count header with optional line items.
func (uc *StockCountsUseCase) CreateStockCount(ctx context.Context, input CreateStockCountInput) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}

	if input.StoreID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "store_id is required", nil)
	}

	// Generate count_number if not provided
	countNumber := ""
	if input.CountNumber != nil && strings.TrimSpace(*input.CountNumber) != "" {
		countNumber = strings.TrimSpace(*input.CountNumber)
	} else {
		countNumber = fmt.Sprintf("SC-%s-%04d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)
	}

	// Check count_number uniqueness
	if existing, err := uc.repo.GetStockCountByNumber(ctx, countNumber); err == nil && existing.ID > 0 {
		countNumber = fmt.Sprintf("SC-%s-%d", time.Now().Format("20060102"), time.Now().Unix())
	}

	var statusText pgtype.Text
	if input.Status != nil && *input.Status != "" {
		statusText = pgtype.Text{String: *input.Status, Valid: true}
	} else {
		statusText = pgtype.Text{String: "planned", Valid: true}
	}

	var scheduledDate pgtype.Date
	if input.ScheduledDate != nil && *input.ScheduledDate != "" {
		if t, err := time.Parse("2006-01-02", *input.ScheduledDate); err == nil {
			scheduledDate = pgtype.Date{Time: t, Valid: true}
		}
	}

	metadataBytes := []byte("{}")
	if input.Metadata != nil {
		if b, err := json.Marshal(input.Metadata); err == nil {
			metadataBytes = b
		}
	}

	createParams := repository.CreateStockCountParams{
		CountNumber:       countNumber,
		StoreID:           input.StoreID,
		StorageLocationID: utils.Int32ToPgInt4(input.StorageLocationID),
		CountType:         utils.StringToPgText(input.CountType),
		Status:            statusText,
		ScheduledDate:     scheduledDate,
		CountedBy:         utils.Int32ToPgInt4(input.CountedBy),
		Metadata:          metadataBytes,
	}

	sc, err := uc.repo.CreateStockCount(ctx, createParams)
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to create stock count: %v", err), nil)
	}

	// Insert line items if provided
	for _, line := range input.Lines {
		if line.ProductID <= 0 {
			continue
		}

		// Calculate variance if not explicitly provided
		expectedQty := 0.0
		if line.ExpectedQuantity != nil {
			expectedQty = *line.ExpectedQuantity
		} else if line.SystemQuantity != nil {
			expectedQty = *line.SystemQuantity
		}

		variance := line.CountedQuantity - expectedQty
		if line.Variance != nil {
			variance = *line.Variance
		}

		varianceVal := 0.0
		if line.VarianceValue != nil {
			varianceVal = *line.VarianceValue
		}

		var countedAt pgtype.Timestamp
		if line.CountedAt != nil && *line.CountedAt != "" {
			if t, err := time.Parse(time.RFC3339, *line.CountedAt); err == nil {
				countedAt = pgtype.Timestamp{Time: t, Valid: true}
			} else if t, err := time.Parse("2006-01-02", *line.CountedAt); err == nil {
				countedAt = pgtype.Timestamp{Time: t, Valid: true}
			}
		} else {
			countedAt = pgtype.Timestamp{Time: time.Now(), Valid: true}
		}

		lineMetaBytes := []byte("{}")
		if line.Metadata != nil {
			if b, err := json.Marshal(line.Metadata); err == nil {
				lineMetaBytes = b
			}
		}

		lineParams := repository.CreateStockCountLineParams{
			StockCountID:      sc.ID,
			ProductID:         line.ProductID,
			ProductVariantID:  utils.Int32ToPgInt4(line.ProductVariantID),
			StorageLocationID: utils.Int32ToPgInt4(line.StorageLocationID),
			ExpectedQuantity:  utils.Float64ToPgNumeric(expectedQty),
			SystemQuantity:    utils.Float64ToPgNumeric(expectedQty),
			CountedQuantity:   utils.Float64ToPgNumeric(line.CountedQuantity),
			Variance:          utils.Float64ToPgNumeric(variance),
			VarianceValue:     utils.Float64ToPgNumeric(varianceVal),
			CountedAt:         countedAt,
			UomID:             utils.Int32ToPgInt4(line.UomID),
			BatchNumber:       utils.StringToPgText(line.BatchNumber),
			SerialNumber:      utils.StringToPgText(line.SerialNumber),
			Metadata:          lineMetaBytes,
		}

		if _, err := uc.repo.CreateStockCountLine(ctx, lineParams); err != nil {
			return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to create stock count line for product %d: %v", line.ProductID, err), nil)
		}
	}

	return uc.GetStockCount(ctx, sc.ID)
}

// GetStockCount retrieves a stock count by ID with full details, lines, and summary.
func (uc *StockCountsUseCase) GetStockCount(ctx context.Context, id int32) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}

	sc, err := uc.repo.GetStockCountWithDetails(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "stock count not found", nil)
	}

	linesRows, err := uc.repo.ListStockCountLinesWithDetails(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to fetch stock count lines: %v", err), nil)
	}

	lines := make([]StockCountLineOutput, 0, len(linesRows))
	for _, l := range linesRows {
		lines = append(lines, StockCountLineOutput{
			ID:                  l.ID,
			StockCountID:        l.StockCountID,
			ProductID:           l.ProductID,
			ProductName:         pgtype.Text{String: l.ProductName, Valid: true},
			ProductSKU:          pgtype.Text{String: l.ProductSku, Valid: true},
			ProductVariantID:    l.ProductVariantID,
			VariantName:         l.VariantName,
			VariantSKU:          l.VariantSku,
			StorageLocationID:   l.StorageLocationID,
			StorageLocationName: l.StorageLocationName,
			ExpectedQuantity:    l.ExpectedQuantity,
			SystemQuantity:      l.SystemQuantity,
			CountedQuantity:     l.CountedQuantity,
			Variance:            l.Variance,
			VarianceValue:       l.VarianceValue,
			CountedAt:           l.CountedAt,
			UomID:               l.UomID,
			UomName:             l.UomName,
			BatchNumber:         l.BatchNumber,
			SerialNumber:        l.SerialNumber,
			Metadata:            utils.BytesToJSONRawMessage(l.Metadata),
			CreatedAt:           l.CreatedAt,
			UpdatedAt:           l.UpdatedAt,
		})
	}

	summaryRow, err := uc.repo.GetStockCountSummary(ctx, id)
	var summary *StockCountSummaryOutput
	if err == nil {
		summary = &StockCountSummaryOutput{
			TotalLines:         summaryRow.TotalLines,
			LinesWithVariance:  summaryRow.LinesWithVariance,
			TotalVarianceValue: summaryRow.TotalVarianceValue,
			PositiveVariance:   summaryRow.PositiveVariance,
			NegativeVariance:   summaryRow.NegativeVariance,
		}
	}

	out := StockCountOutput{
		ID:                  sc.ID,
		CountNumber:         sc.CountNumber,
		StoreID:             sc.StoreID,
		StoreName:           pgtype.Text{String: sc.StoreName, Valid: true},
		StorageLocationID:   sc.StorageLocationID,
		StorageLocationName: sc.StorageLocationName,
		CountType:           sc.CountType,
		Status:              sc.Status,
		ScheduledDate:       sc.ScheduledDate,
		StartedAt:           sc.StartedAt,
		CompletedAt:         sc.CompletedAt,
		CountedBy:           sc.CountedBy,
		CountedByName:       sc.CountedByName,
		ApprovedBy:          sc.ApprovedBy,
		ApprovedByName:      sc.ApprovedByName,
		Metadata:            utils.BytesToJSONRawMessage(sc.Metadata),
		CreatedAt:           sc.CreatedAt,
		UpdatedAt:           sc.UpdatedAt,
		Summary:             summary,
		Lines:               lines,
	}

	if summary != nil {
		out.TotalLines = summary.TotalLines
		out.LinesWithVariance = summary.LinesWithVariance
		out.TotalVarianceValue = summary.TotalVarianceValue
	}

	return utils.NewResponse(utils.CodeOK, "stock count retrieved successfully", out)
}

// GetStockCountByNumber retrieves a stock count by count_number.
func (uc *StockCountsUseCase) GetStockCountByNumber(ctx context.Context, countNumber string) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}

	sc, err := uc.repo.GetStockCountByNumber(ctx, countNumber)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "stock count not found", nil)
	}

	return uc.GetStockCount(ctx, sc.ID)
}

// ListStockCounts lists stock counts with filtering, search, and pagination.
func (uc *StockCountsUseCase) ListStockCounts(ctx context.Context, filter StockCountFilter) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}

	rows, err := uc.repo.ListStockCountsWithDetails(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to list stock counts: %v", err), nil)
	}

	var filtered []StockCountOutput
	searchLower := ""
	if filter.Search != nil && *filter.Search != "" {
		searchLower = strings.ToLower(strings.TrimSpace(*filter.Search))
	}

	for _, r := range rows {
		// Filter by store
		if filter.StoreID != nil && *filter.StoreID > 0 && r.StoreID != *filter.StoreID {
			continue
		}

		// Filter by status
		if filter.Status != nil && *filter.Status != "" {
			if !r.Status.Valid || !strings.EqualFold(r.Status.String, *filter.Status) {
				continue
			}
		}

		// Filter by count_type
		if filter.CountType != nil && *filter.CountType != "" {
			if !r.CountType.Valid || !strings.EqualFold(r.CountType.String, *filter.CountType) {
				continue
			}
		}

		// Filter by search (count_number, store_name, counted_by_name)
		if searchLower != "" {
			match := strings.Contains(strings.ToLower(r.CountNumber), searchLower) ||
				(r.StoreName != "" && strings.Contains(strings.ToLower(r.StoreName), searchLower)) ||
				(r.CountedByName.Valid && strings.Contains(strings.ToLower(r.CountedByName.String), searchLower))
			if !match {
				continue
			}
		}

		// Filter by date range (scheduled_date or created_at)
		if filter.FromDate != nil && *filter.FromDate != "" {
			if fromT, err := time.Parse("2006-01-02", *filter.FromDate); err == nil {
				if r.ScheduledDate.Valid && r.ScheduledDate.Time.Before(fromT) {
					continue
				} else if !r.ScheduledDate.Valid && r.CreatedAt.Valid && r.CreatedAt.Time.Before(fromT) {
					continue
				}
			}
		}

		if filter.ToDate != nil && *filter.ToDate != "" {
			if toT, err := time.Parse("2006-01-02", *filter.ToDate); err == nil {
				toTEnd := toT.Add(24 * time.Hour)
				if r.ScheduledDate.Valid && r.ScheduledDate.Time.After(toTEnd) {
					continue
				} else if !r.ScheduledDate.Valid && r.CreatedAt.Valid && r.CreatedAt.Time.After(toTEnd) {
					continue
				}
			}
		}

		filtered = append(filtered, StockCountOutput{
			ID:                  r.ID,
			CountNumber:         r.CountNumber,
			StoreID:             r.StoreID,
			StoreName:           pgtype.Text{String: r.StoreName, Valid: true},
			StorageLocationID:   r.StorageLocationID,
			StorageLocationName: r.StorageLocationName,
			CountType:           r.CountType,
			Status:              r.Status,
			ScheduledDate:       r.ScheduledDate,
			StartedAt:           r.StartedAt,
			CompletedAt:         r.CompletedAt,
			CountedBy:           r.CountedBy,
			CountedByName:       r.CountedByName,
			ApprovedBy:          r.ApprovedBy,
			ApprovedByName:      r.ApprovedByName,
			Metadata:            utils.BytesToJSONRawMessage(r.Metadata),
			CreatedAt:           r.CreatedAt,
			UpdatedAt:           r.UpdatedAt,
			TotalLines:          r.TotalLines,
			LinesWithVariance:   r.LinesWithVariance,
			TotalVarianceValue:  r.TotalVarianceValue,
		})
	}

	totalCount := int64(len(filtered))
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 20
	}

	start := (page - 1) * limit
	end := start + limit

	var paginated []StockCountOutput
	if start < int32(totalCount) {
		if end > int32(totalCount) {
			end = int32(totalCount)
		}
		paginated = filtered[start:end]
	} else {
		paginated = []StockCountOutput{}
	}

	totalPages := int32(math.Ceil(float64(totalCount) / float64(limit)))

	return utils.NewResponse(utils.CodeOK, "stock counts listed successfully", StockCountListResponse{
		Data:       paginated,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// UpdateStockCount updates a stock count header.
func (uc *StockCountsUseCase) UpdateStockCount(ctx context.Context, id int32, input UpdateStockCountInput) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}

	current, err := uc.repo.GetStockCount(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "stock count not found", nil)
	}

	// Prevent updates if already reconciled/cancelled
	if current.Status.Valid && (current.Status.String == "reconciled" || current.Status.String == "cancelled") {
		return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("cannot update stock count in '%s' status", current.Status.String), nil)
	}

	var scheduledDate pgtype.Date
	if input.ScheduledDate != nil && *input.ScheduledDate != "" {
		if t, err := time.Parse("2006-01-02", *input.ScheduledDate); err == nil {
			scheduledDate = pgtype.Date{Time: t, Valid: true}
		}
	} else {
		scheduledDate = current.ScheduledDate
	}

	var metadataBytes []byte
	if input.Metadata != nil {
		if b, err := json.Marshal(input.Metadata); err == nil {
			metadataBytes = b
		}
	}

	updateParams := repository.UpdateStockCountParams{
		ID:                id,
		StorageLocationID: utils.Int32ToPgInt4(input.StorageLocationID),
		CountType:         utils.StringToPgText(input.CountType),
		Status:            utils.StringToPgText(input.Status),
		ScheduledDate:     scheduledDate,
		CountedBy:         utils.Int32ToPgInt4(input.CountedBy),
		Metadata:          metadataBytes,
	}

	if _, err := uc.repo.UpdateStockCount(ctx, updateParams); err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to update stock count: %v", err), nil)
	}

	return uc.GetStockCount(ctx, id)
}

// DeleteStockCount deletes a stock count and its lines.
func (uc *StockCountsUseCase) DeleteStockCount(ctx context.Context, id int32) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}

	current, err := uc.repo.GetStockCount(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "stock count not found", nil)
	}

	if current.Status.Valid && (current.Status.String == "reconciled" || current.Status.String == "completed") {
		return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("cannot delete stock count in '%s' status", current.Status.String), nil)
	}

	if err := uc.repo.DeleteStockCount(ctx, id); err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to delete stock count: %v", err), nil)
	}

	return utils.NewResponse(utils.CodeOK, "stock count deleted successfully", map[string]interface{}{"id": id})
}

// StartStockCount transitions status to 'in_progress' and records started_at.
func (uc *StockCountsUseCase) StartStockCount(ctx context.Context, id int32) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}

	current, err := uc.repo.GetStockCount(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "stock count not found", nil)
	}

	if current.Status.Valid && current.Status.String != "planned" {
		return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("cannot start stock count from status '%s'", current.Status.String), nil)
	}

	if _, err := uc.repo.StartStockCount(ctx, id); err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to start stock count: %v", err), nil)
	}

	return uc.GetStockCount(ctx, id)
}

// CompleteStockCount transitions status to 'completed' and records completed_at.
func (uc *StockCountsUseCase) CompleteStockCount(ctx context.Context, id int32) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}

	current, err := uc.repo.GetStockCount(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "stock count not found", nil)
	}

	if current.Status.Valid && current.Status.String != "in_progress" {
		return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("cannot complete stock count from status '%s'", current.Status.String), nil)
	}

	if _, err := uc.repo.CompleteStockCount(ctx, id); err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to complete stock count: %v", err), nil)
	}

	return uc.GetStockCount(ctx, id)
}

// ApproveStockCount approves a completed stock count.
func (uc *StockCountsUseCase) ApproveStockCount(ctx context.Context, id int32, approvedBy int32) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}

	current, err := uc.repo.GetStockCount(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "stock count not found", nil)
	}

	if current.Status.Valid && current.Status.String != "completed" {
		return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("cannot approve stock count with status '%s' (must be 'completed')", current.Status.String), nil)
	}

	var approvedByPg pgtype.Int4
	if approvedBy > 0 {
		approvedByPg = pgtype.Int4{Int32: approvedBy, Valid: true}
	}

	if _, err := uc.repo.ApproveStockCount(ctx, repository.ApproveStockCountParams{
		ID:         id,
		ApprovedBy: approvedByPg,
	}); err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to approve stock count: %v", err), nil)
	}

	return uc.GetStockCount(ctx, id)
}

// ReconcileStockCount reconciles approved stock count variances into inventory_stock and creates stock movements.
func (uc *StockCountsUseCase) ReconcileStockCount(ctx context.Context, id int32) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}

	current, err := uc.repo.GetStockCount(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "stock count not found", nil)
	}

	if !current.Status.Valid || current.Status.String != "approved" {
		return utils.NewResponse(utils.CodeBadReq, "stock count must be in 'approved' status to reconcile", nil)
	}

	lines, err := uc.repo.ListStockCountLines(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to list lines: %v", err), nil)
	}

	// Update inventory_stock for each line
	for _, l := range lines {
		// Lookup existing inventory stock record
		existingStock, err := uc.repo.GetInventoryStockByProductAndStore(ctx, repository.GetInventoryStockByProductAndStoreParams{
			ProductID:         l.ProductID,
			ProductVariantID:  l.ProductVariantID,
			StoreID:           current.StoreID,
			StorageLocationID: l.StorageLocationID,
		})
		if err == nil {
			_, _ = uc.repo.UpdateInventoryStock(ctx, repository.UpdateInventoryStockParams{
				ID:                existingStock.ID,
				QuantityOnHand:    l.CountedQuantity,
				QuantityAllocated: existingStock.QuantityAllocated,
				QuantityAvailable: l.CountedQuantity,
				QuantityOnOrder:   existingStock.QuantityOnOrder,
				QuantityInTransit: existingStock.QuantityInTransit,
				ReorderLevel:      existingStock.ReorderLevel,
				ReorderQuantity:   existingStock.ReorderQuantity,
				MaxStockLevel:     existingStock.MaxStockLevel,
				LastCountedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
				Metadata:          existingStock.Metadata,
			})
		} else {
			_, _ = uc.repo.CreateInventoryStock(ctx, repository.CreateInventoryStockParams{
				ProductID:         l.ProductID,
				ProductVariantID:  l.ProductVariantID,
				StoreID:           current.StoreID,
				StorageLocationID: l.StorageLocationID,
				QuantityOnHand:    l.CountedQuantity,
				QuantityAllocated: pgtype.Numeric{},
				QuantityAvailable: l.CountedQuantity,
				QuantityOnOrder:   pgtype.Numeric{},
				QuantityInTransit: pgtype.Numeric{},
				ReorderLevel:      pgtype.Numeric{},
				ReorderQuantity:   pgtype.Numeric{},
				MaxStockLevel:     pgtype.Numeric{},
				Metadata:          []byte("{}"),
			})
		}

		// Record stock movement if variance is not 0
		varianceVal, _ := l.Variance.Float64Value()
		if varianceVal.Valid && varianceVal.Float64 != 0 {
			_, _ = uc.repo.CreateStockMovement(ctx, repository.CreateStockMovementParams{
				MovementType:     "stock_adjustment",
				ReferenceType:    pgtype.Text{String: "stock_count", Valid: true},
				ReferenceID:      pgtype.Int4{Int32: id, Valid: true},
				ProductID:        l.ProductID,
				ProductVariantID: l.ProductVariantID,
				FromStoreID:      pgtype.Int4{Int32: current.StoreID, Valid: true},
				ToStoreID:        pgtype.Int4{Int32: current.StoreID, Valid: true},
				FromLocationID:   l.StorageLocationID,
				ToLocationID:     l.StorageLocationID,
				Quantity:         l.Variance,
				UomID:            l.UomID,
				BatchNumber:      l.BatchNumber,
				SerialNumber:     l.SerialNumber,
				MovementDate:     pgtype.Timestamp{Time: time.Now(), Valid: true},
				PostedBy:         current.ApprovedBy,
				Status:           pgtype.Text{String: "completed", Valid: true},
				CostPerUnit:      pgtype.Numeric{},
				TotalValue:       l.VarianceValue,
				Metadata:         []byte("{}"),
			})
		}
	}

	if _, err := uc.repo.ReconcileStockCountStatus(ctx, id); err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to update reconciled status: %v", err), nil)
	}

	return uc.GetStockCount(ctx, id)
}

// GetStockCountSummary calculates variance metrics for a stock count.
func (uc *StockCountsUseCase) GetStockCountSummary(ctx context.Context, id int32) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}

	summary, err := uc.repo.GetStockCountSummary(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to calculate summary: %v", err), nil)
	}

	return utils.NewResponse(utils.CodeOK, "stock count summary calculated successfully", StockCountSummaryOutput{
		TotalLines:         summary.TotalLines,
		LinesWithVariance:  summary.LinesWithVariance,
		TotalVarianceValue: summary.TotalVarianceValue,
		PositiveVariance:   summary.PositiveVariance,
		NegativeVariance:   summary.NegativeVariance,
	})
}

// AddStockCountLine adds a line item to a stock count.
func (uc *StockCountsUseCase) AddStockCountLine(ctx context.Context, countID int32, input StockCountLineInput) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}

	if countID <= 0 || input.ProductID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "stock_count_id and product_id are required", nil)
	}

	expectedQty := 0.0
	if input.ExpectedQuantity != nil {
		expectedQty = *input.ExpectedQuantity
	} else if input.SystemQuantity != nil {
		expectedQty = *input.SystemQuantity
	}

	variance := input.CountedQuantity - expectedQty
	if input.Variance != nil {
		variance = *input.Variance
	}

	varianceVal := 0.0
	if input.VarianceValue != nil {
		varianceVal = *input.VarianceValue
	}

	var countedAt pgtype.Timestamp
	if input.CountedAt != nil && *input.CountedAt != "" {
		if t, err := time.Parse(time.RFC3339, *input.CountedAt); err == nil {
			countedAt = pgtype.Timestamp{Time: t, Valid: true}
		} else if t, err := time.Parse("2006-01-02", *input.CountedAt); err == nil {
			countedAt = pgtype.Timestamp{Time: t, Valid: true}
		}
	} else {
		countedAt = pgtype.Timestamp{Time: time.Now(), Valid: true}
	}

	lineMetaBytes := []byte("{}")
	if input.Metadata != nil {
		if b, err := json.Marshal(input.Metadata); err == nil {
			lineMetaBytes = b
		}
	}

	lineParams := repository.CreateStockCountLineParams{
		StockCountID:      countID,
		ProductID:         input.ProductID,
		ProductVariantID:  utils.Int32ToPgInt4(input.ProductVariantID),
		StorageLocationID: utils.Int32ToPgInt4(input.StorageLocationID),
		ExpectedQuantity:  utils.Float64ToPgNumeric(expectedQty),
		SystemQuantity:    utils.Float64ToPgNumeric(expectedQty),
		CountedQuantity:   utils.Float64ToPgNumeric(input.CountedQuantity),
		Variance:          utils.Float64ToPgNumeric(variance),
		VarianceValue:     utils.Float64ToPgNumeric(varianceVal),
		CountedAt:         countedAt,
		UomID:             utils.Int32ToPgInt4(input.UomID),
		BatchNumber:       utils.StringToPgText(input.BatchNumber),
		SerialNumber:      utils.StringToPgText(input.SerialNumber),
		Metadata:          lineMetaBytes,
	}

	createdLine, err := uc.repo.CreateStockCountLine(ctx, lineParams)
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to create stock count line: %v", err), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "stock count line created successfully", createdLine)
}

// ListStockCountLines lists line items for a stock count.
func (uc *StockCountsUseCase) ListStockCountLines(ctx context.Context, countID int32) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}

	linesRows, err := uc.repo.ListStockCountLinesWithDetails(ctx, countID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to list stock count lines: %v", err), nil)
	}

	lines := make([]StockCountLineOutput, 0, len(linesRows))
	for _, l := range linesRows {
		lines = append(lines, StockCountLineOutput{
			ID:                  l.ID,
			StockCountID:        l.StockCountID,
			ProductID:           l.ProductID,
			ProductName:         pgtype.Text{String: l.ProductName, Valid: true},
			ProductSKU:          pgtype.Text{String: l.ProductSku, Valid: true},
			ProductVariantID:    l.ProductVariantID,
			VariantName:         l.VariantName,
			VariantSKU:          l.VariantSku,
			StorageLocationID:   l.StorageLocationID,
			StorageLocationName: l.StorageLocationName,
			ExpectedQuantity:    l.ExpectedQuantity,
			SystemQuantity:      l.SystemQuantity,
			CountedQuantity:     l.CountedQuantity,
			Variance:            l.Variance,
			VarianceValue:       l.VarianceValue,
			CountedAt:           l.CountedAt,
			UomID:               l.UomID,
			UomName:             l.UomName,
			BatchNumber:         l.BatchNumber,
			SerialNumber:        l.SerialNumber,
			Metadata:            utils.BytesToJSONRawMessage(l.Metadata),
			CreatedAt:           l.CreatedAt,
			UpdatedAt:           l.UpdatedAt,
		})
	}

	return utils.NewResponse(utils.CodeOK, "stock count lines listed successfully", lines)
}

// GetStockCountLine retrieves a single line item by ID.
func (uc *StockCountsUseCase) GetStockCountLine(ctx context.Context, lineID int32) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}

	l, err := uc.repo.GetStockCountLine(ctx, lineID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "stock count line not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "stock count line retrieved successfully", l)
}

// UpdateStockCountLine updates an individual stock count line.
func (uc *StockCountsUseCase) UpdateStockCountLine(ctx context.Context, lineID int32, input UpdateStockCountLineInput) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}

	currentLine, err := uc.repo.GetStockCountLine(ctx, lineID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "stock count line not found", nil)
	}

	sysQtyVal, _ := currentLine.SystemQuantity.Float64Value()
	sysQty := sysQtyVal.Float64

	variance := input.CountedQuantity - sysQty
	if input.Variance != nil {
		variance = *input.Variance
	}

	varianceVal := 0.0
	if input.VarianceValue != nil {
		varianceVal = *input.VarianceValue
	}

	var countedAt pgtype.Timestamp
	if input.CountedAt != nil && *input.CountedAt != "" {
		if t, err := time.Parse(time.RFC3339, *input.CountedAt); err == nil {
			countedAt = pgtype.Timestamp{Time: t, Valid: true}
		} else if t, err := time.Parse("2006-01-02", *input.CountedAt); err == nil {
			countedAt = pgtype.Timestamp{Time: t, Valid: true}
		}
	} else {
		countedAt = pgtype.Timestamp{Time: time.Now(), Valid: true}
	}

	var metadataBytes []byte
	if input.Metadata != nil {
		if b, err := json.Marshal(input.Metadata); err == nil {
			metadataBytes = b
		}
	}

	updated, err := uc.repo.UpdateStockCountLine(ctx, repository.UpdateStockCountLineParams{
		ID:              lineID,
		CountedQuantity: utils.Float64ToPgNumeric(input.CountedQuantity),
		Variance:        utils.Float64ToPgNumeric(variance),
		VarianceValue:   utils.Float64ToPgNumeric(varianceVal),
		CountedAt:       countedAt,
		BatchNumber:     utils.StringToPgText(input.BatchNumber),
		SerialNumber:    utils.StringToPgText(input.SerialNumber),
		Metadata:        metadataBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to update stock count line: %v", err), nil)
	}

	return utils.NewResponse(utils.CodeOK, "stock count line updated successfully", updated)
}

// DeleteStockCountLine deletes a line item by ID.
func (uc *StockCountsUseCase) DeleteStockCountLine(ctx context.Context, lineID int32) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}

	if err := uc.repo.DeleteStockCountLine(ctx, lineID); err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to delete stock count line: %v", err), nil)
	}

	return utils.NewResponse(utils.CodeOK, "stock count line deleted successfully", map[string]interface{}{"id": lineID})
}

// BulkUpdateStockCountLines updates multiple lines in a single call.
func (uc *StockCountsUseCase) BulkUpdateStockCountLines(ctx context.Context, countID int32, lines []BulkStockCountLineItem) *repository.Response {
	if errResp := uc.repoOrErr(); errResp != nil {
		return errResp
	}

	if countID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "stock_count_id is required", nil)
	}

	updatedCount := 0
	for _, l := range lines {
		if l.ID <= 0 {
			continue
		}

		currentLine, err := uc.repo.GetStockCountLine(ctx, l.ID)
		if err != nil {
			continue
		}

		sysQtyVal, _ := currentLine.SystemQuantity.Float64Value()
		sysQty := sysQtyVal.Float64

		variance := l.CountedQuantity - sysQty
		if l.Variance != nil {
			variance = *l.Variance
		}

		varianceVal := 0.0
		if l.VarianceValue != nil {
			varianceVal = *l.VarianceValue
		}

		var countedAt pgtype.Timestamp
		if l.CountedAt != nil && *l.CountedAt != "" {
			if t, err := time.Parse(time.RFC3339, *l.CountedAt); err == nil {
				countedAt = pgtype.Timestamp{Time: t, Valid: true}
			} else if t, err := time.Parse("2006-01-02", *l.CountedAt); err == nil {
				countedAt = pgtype.Timestamp{Time: t, Valid: true}
			}
		} else {
			countedAt = pgtype.Timestamp{Time: time.Now(), Valid: true}
		}

		var metadataBytes []byte
		if l.Metadata != nil {
			if b, err := json.Marshal(l.Metadata); err == nil {
				metadataBytes = b
			}
		}

		_, err = uc.repo.UpdateStockCountLine(ctx, repository.UpdateStockCountLineParams{
			ID:              l.ID,
			CountedQuantity: utils.Float64ToPgNumeric(l.CountedQuantity),
			Variance:        utils.Float64ToPgNumeric(variance),
			VarianceValue:   utils.Float64ToPgNumeric(varianceVal),
			CountedAt:       countedAt,
			BatchNumber:     utils.StringToPgText(l.BatchNumber),
			SerialNumber:    utils.StringToPgText(l.SerialNumber),
			Metadata:        metadataBytes,
		})
		if err == nil {
			updatedCount++
		}
	}

	return uc.GetStockCount(ctx, countID)
}
