package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// PromotionUseCase handles promotion/coupon administration.
type PromotionUseCase struct {
	repo *repository.Queries
}

func NewPromotionUseCase() *PromotionUseCase {
	return &PromotionUseCase{}
}

func (uc *PromotionUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *PromotionUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

// ─── Admin CRUD ────────────────────────────────────────────────────────────────

// CreatePromotion creates a new promotion/coupon.
func (uc *PromotionUseCase) CreatePromotion(ctx context.Context, arg repository.CreatePromotionParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	promotion, err := uc.repo.CreatePromotion(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create promotion: "+err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "promotion created successfully", promotion)
}

// GetPromotion retrieves a promotion by its ID.
func (uc *PromotionUseCase) GetPromotion(ctx context.Context, id int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	promotion, err := uc.repo.GetPromotion(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "promotion not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "promotion fetched successfully", promotion)
}

// GetPromotionByCode retrieves a promotion by its internal code (not coupon_code).
func (uc *PromotionUseCase) GetPromotionByCode(ctx context.Context, code string, orgID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	promotion, err := uc.repo.GetPromotionByCode(ctx, repository.GetPromotionByCodeParams{
		Code:           code,
		OrganizationID: orgID,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "promotion not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "promotion fetched successfully", promotion)
}

// ListActivePromotions lists all active promotions for an organisation.
func (uc *PromotionUseCase) ListActivePromotions(ctx context.Context, orgID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	promotions, err := uc.repo.ListActivePromotions(ctx, orgID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list promotions", nil)
	}
	return utils.NewResponse(utils.CodeOK, "promotions listed", promotions)
}

// ListAllPromotions paginates all promotions for an organisation (active and inactive).
func (uc *PromotionUseCase) ListAllPromotions(ctx context.Context, arg repository.ListAllPromotionsParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	promotions, err := uc.repo.ListAllPromotions(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list promotions", nil)
	}
	return utils.NewResponse(utils.CodeOK, "promotions listed", promotions)
}

// UpdatePromotion updates an existing promotion.
func (uc *PromotionUseCase) UpdatePromotion(ctx context.Context, arg repository.UpdatePromotionParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	promotion, err := uc.repo.UpdatePromotion(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update promotion: "+err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "promotion updated", promotion)
}

// UpdatePromotionStatus toggles a promotion active/inactive.
func (uc *PromotionUseCase) UpdatePromotionStatus(ctx context.Context, arg repository.UpdatePromotionStatusParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	promotion, err := uc.repo.UpdatePromotionStatus(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update promotion status: "+err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "promotion status updated", promotion)
}

// DeletePromotion deletes a promotion by ID.
func (uc *PromotionUseCase) DeletePromotion(ctx context.Context, id int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if err := uc.repo.DeletePromotion(ctx, id); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to delete promotion: "+err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "promotion deleted", nil)
}

// ─── Coupon Application Flow ───────────────────────────────────────────────────

// ApplyCouponInput holds the inputs for applying a coupon to a cart.
type ApplyCouponInput struct {
	CartID         uuid.UUID `json:"cart_id"`
	CouponCode     string    `json:"coupon_code"`
	OrganizationID int32     `json:"organization_id"`
}

// buyXGetYMetadata is the expected shape of action_metadata for buy_x_get_y promotions.
//
//	{"buy_id": 101, "buy_qty": 2, "get_id": 205, "get_qty": 1, "discount_percent": 100}
type buyXGetYMetadata struct {
	BuyID           int32   `json:"buy_id"`
	BuyQty          int32   `json:"buy_qty"`
	GetID           int32   `json:"get_id"`
	GetQty          int32   `json:"get_qty"`
	DiscountPercent float64 `json:"discount_percent"`
}

// happyHourSchedule is the expected shape of schedule_json for happy_hour promotions.
//
//	{"days": ["Monday","Tuesday"], "start_time": "20:00", "end_time": "22:00"}
type happyHourSchedule struct {
	Days      []string `json:"days"`
	StartTime string   `json:"start_time"` // "HH:MM" 24-hour
	EndTime   string   `json:"end_time"`   // "HH:MM" 24-hour
}

// pointsMultiplierMetadata is the expected shape of action_metadata for points_multiplier.
//
//	{"multiplier": 2}
type pointsMultiplierMetadata struct {
	Multiplier float64 `json:"multiplier"`
}

// ApplyCoupon validates and applies a coupon code to a cart.
//
// Flow:
//  1. Fetch the cart and verify it is active.
//  2. Look up the promotion by coupon_code — SQL checks: is_active, dates, usage_limit.
//  3. Validate min_order_amount and min_quantity constraints.
//  4. Calculate the discount based on promotion_type.
//  5. Apply discount: whole-cart via ApplyCouponToCart; per-item via
//     ApplyDiscountToCartItem (for product/category-targeted promotions).
//  6. Atomically increment the promotion usage counter.
//  7. Recalculate cart totals.
func (uc *PromotionUseCase) ApplyCoupon(ctx context.Context, in ApplyCouponInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	// 1. Fetch cart
	cart, err := uc.repo.GetCartByID(ctx, in.CartID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cart not found", nil)
	}
	if cart.CartStatus != repository.CartStatusActive {
		return utils.NewResponse(utils.CodeBadReq, "coupon can only be applied to an active cart", nil)
	}

	// 2. Fetch and validate promotion (SQL handles date + usage checks)
	promo, err := uc.repo.GetActivePromotionByCouponCode(ctx, repository.GetActivePromotionByCouponCodeParams{
		CouponCode:     pgtype.Text{String: in.CouponCode, Valid: true},
		OrganizationID: in.OrganizationID,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "coupon code is invalid, expired, or has reached its usage limit", nil)
	}

	// 2a. Check store restriction
	if !isStoreAllowed(promo.StoreIds, cart.StoreID) {
		return utils.NewResponse(utils.CodeBadReq, "coupon is not valid for this store", nil)
	}

	// 3. Fetch cart totals for constraint checks
	totals, err := uc.repo.GetCartTotals(ctx, in.CartID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch cart totals", nil)
	}

	subtotalFloat := interfaceToFloat(totals.Subtotal)

	// 4a. Check min_order_amount
	if promo.MinOrderAmount.Valid {
		minAmount := numericToFloat(promo.MinOrderAmount)
		if subtotalFloat < minAmount {
			return utils.NewResponse(utils.CodeBadReq,
				fmt.Sprintf("cart subtotal (%.2f) is below the minimum required (%.2f)", subtotalFloat, minAmount),
				nil)
		}
	}

	// 4b. Check min_quantity
	if promo.MinQuantity.Valid {
		itemCount, err := uc.repo.GetCartItemCount(ctx, in.CartID)
		if err != nil {
			return utils.NewResponse(utils.CodeError, "failed to fetch cart item count", nil)
		}
		minQty := numericToFloat(promo.MinQuantity)
		totalQty := interfaceToFloat(itemCount.TotalQuantity)
		if totalQty < minQty {
			return utils.NewResponse(utils.CodeBadReq,
				fmt.Sprintf("cart quantity (%.2f) is below the minimum required (%.2f)", totalQty, minQty),
				nil)
		}
	}

	// 5. Fetch cart items once (needed by several promotion types)
	items, err := uc.repo.ListCartItems(ctx, in.CartID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch cart items", nil)
	}

	// 6. Calculate discount and mark items
	var discountFloat float64
	extraResponse := map[string]interface{}{}

	isProductTargeted := promo.AppliesTo.Valid &&
		(promo.AppliesTo.String == "product" || promo.AppliesTo.String == "category") &&
		(len(promo.TargetProductIds) > 0 || len(promo.TargetCategoryIds) > 0)
	discountValueFloat := numericToFloat(promo.DiscountValue)
	switch promo.PromotionType {

	// ── 1. Simple Minimum-Spend (fixed or percentage) ──────────────────────────
	case "percentage_discount":
		if isProductTargeted {
			targetProductSet := make(map[int32]bool, len(promo.TargetProductIds))
			for _, pid := range promo.TargetProductIds {
				targetProductSet[pid] = true
			}
			discountFloat = uc.applyProductSetDiscount(ctx, items, targetProductSet, discountValueFloat, promo.Code)
			if discountFloat == 0 {
				return utils.NewResponse(utils.CodeBadReq, "coupon is not applicable to any products in your cart", nil)
			}
			isProductTargeted = false
		} else {
			// Applies a % off the subtotal (optionally constrained by min_order_amount already checked)
			discountFloat = subtotalFloat * (discountValueFloat / 100)
		}

	case "fixed_discount":
		if isProductTargeted {
			targetProductSet := make(map[int32]bool, len(promo.TargetProductIds))
			for _, pid := range promo.TargetProductIds {
				targetProductSet[pid] = true
			}
			var eligibleSubtotal float64
			var matchingItems []repository.ListCartItemsRow
			for _, item := range items {
				if targetProductSet[item.ProductID] {
					eligibleSubtotal += numericToFloat(item.UnitPrice) * numericToFloat(item.Quantity)
					matchingItems = append(matchingItems, item)
				}
			}
			if len(matchingItems) == 0 || eligibleSubtotal == 0 {
				return utils.NewResponse(utils.CodeBadReq, "coupon is not applicable to any products in your cart", nil)
			}
			discountFloat = discountValueFloat
			if discountFloat > eligibleSubtotal {
				discountFloat = eligibleSubtotal
			}
			for _, item := range matchingItems {
				itemLineTotal := numericToFloat(item.UnitPrice) * numericToFloat(item.Quantity)
				lineDiscount := discountFloat * (itemLineTotal / eligibleSubtotal)
				perItemNumeric := pgtype.Numeric{}
				_ = perItemNumeric.Scan(fmt.Sprintf("%.2f", lineDiscount))
				_, _ = uc.repo.ApplyDiscountToCartItem(ctx, repository.ApplyDiscountToCartItemParams{
					ID:             item.ID,
					DiscountAmount: perItemNumeric,
					Column3:        promo.Code,
				})
			}
			isProductTargeted = false
		} else {
			// Flat SAR amount off the order (example: "Spend 500 SAR, Get 50 SAR Off")
			discountFloat = discountValueFloat
			if discountFloat > subtotalFloat {
				discountFloat = subtotalFloat // clamp — never negative total
			}
		}

	// ── 2. Buy X Get Y ─────────────────────────────────────────────────────────
	case "buy_x_get_y":
		// action_metadata: {"buy_id":101,"buy_qty":2,"get_id":205,"get_qty":1,"discount_percent":100}
		var meta buyXGetYMetadata
		if err := json.Unmarshal(promo.ActionMetadata, &meta); err != nil {
			return utils.NewResponse(utils.CodeError, "invalid buy_x_get_y action_metadata", nil)
		}

		// Count how many buy-item units are in the cart
		var buyCount float64
		for _, it := range items {
			if it.ProductID == meta.BuyID {
				buyCount += numericToFloat(it.Quantity)
			}
		}

		if buyCount < float64(meta.BuyQty) {
			return utils.NewResponse(utils.CodeBadReq,
				fmt.Sprintf("buy_x_get_y requires %d unit(s) of product %d in cart (found %.0f)",
					meta.BuyQty, meta.BuyID, buyCount),
				nil)
		}

		// Apply discount to get-item line items
		for _, it := range items {
			if it.ProductID == meta.GetID {
				unitPrice := numericToFloat(it.UnitPrice)
				itemDiscount := unitPrice * (meta.DiscountPercent / 100)
				discountFloat += itemDiscount * float64(meta.GetQty)

				perItemNumeric := pgtype.Numeric{}
				_ = perItemNumeric.Scan(fmt.Sprintf("%.2f", itemDiscount))
				_, _ = uc.repo.ApplyDiscountToCartItem(ctx, repository.ApplyDiscountToCartItemParams{
					ID:             it.ID,
					DiscountAmount: perItemNumeric,
					Column3:        promo.Code,
				})
			}
		}
		// discount already applied per-item; cart header gets the total
		isProductTargeted = false // skip the generic per-item block below

	// ── 3. Happy Hour / Time-Based Discounts ───────────────────────────────────
	case "happy_hour":
		// schedule_json: {"days":["Monday","Tuesday"],"start_time":"20:00","end_time":"22:00"}
		if len(promo.ScheduleJson) > 0 {
			var schedule happyHourSchedule
			if err := json.Unmarshal(promo.ScheduleJson, &schedule); err != nil {
				return utils.NewResponse(utils.CodeError, "invalid happy_hour schedule_json", nil)
			}
			if !isHappyHourActive(schedule) {
				return utils.NewResponse(utils.CodeBadReq,
					"this promotion is only valid during happy hour", nil)
			}
		}

		// Apply % discount to category items — but since CartItem has no CategoryID field,
		// category-targeted happy_hour promotions use TargetProductIds to list affected products.
		if isProductTargeted && (len(promo.TargetCategoryIds) > 0 || len(promo.TargetProductIds) > 0) {
			targetProductSet := make(map[int32]bool)
			// Use TargetProductIds if specified; otherwise caller should pre-populate from category lookup
			for _, pid := range promo.TargetProductIds {
				targetProductSet[pid] = true
			}
			categoryDiscount := uc.applyProductSetDiscount(ctx, items, targetProductSet, discountValueFloat, promo.Code)
			discountFloat = categoryDiscount
			isProductTargeted = false
		} else {
			// No category filter: whole-order % discount
			discountFloat = subtotalFloat * (discountValueFloat / 100)
		}

	// ── 4. Loyalty Points Multiplier ───────────────────────────────────────────
	case "points_multiplier":
		// No monetary discount — external loyalty service handles reward
		discountFloat = 0
		var meta pointsMultiplierMetadata
		if len(promo.ActionMetadata) > 0 {
			_ = json.Unmarshal(promo.ActionMetadata, &meta)
		}
		if meta.Multiplier == 0 {
			meta.Multiplier = 1
		}
		extraResponse["points_multiplier"] = meta.Multiplier
		extraResponse["loyalty_note"] = fmt.Sprintf(
			"loyalty points will be multiplied by %.1fx when the order is finalised", meta.Multiplier)

	// ── 5. Free Item ───────────────────────────────────────────────────────────
	case "free_item":
		// discount_value holds the flat monetary value of the free item
		discountFloat = discountValueFloat

	// ── 6. Bundle Price ────────────────────────────────────────────────────────
	case "bundle_price":
		// discount_value is the target bundle total price; discount = subtotal − bundle_price
		if discountValueFloat < subtotalFloat {
			discountFloat = subtotalFloat - discountValueFloat
		}

	default:
		discountFloat = 0
	}

	// 6a. For product-targeted promotions (generic path, not handled inline above)
	if isProductTargeted {
		targetProductSet := make(map[int32]bool, len(promo.TargetProductIds))
		for _, pid := range promo.TargetProductIds {
			targetProductSet[pid] = true
		}

		var matchingCount int
		for _, item := range items {
			if targetProductSet[item.ProductID] {
				matchingCount++
			}
		}
		if matchingCount == 0 {
			return utils.NewResponse(utils.CodeBadReq, "coupon is not applicable to any products in your cart", nil)
		}
		if discountFloat > 0 {
			perItemDiscount := discountFloat / float64(matchingCount)
			perItemNumeric := pgtype.Numeric{}
			_ = perItemNumeric.Scan(fmt.Sprintf("%.2f", perItemDiscount))

			for _, item := range items {
				if targetProductSet[item.ProductID] {
					_, _ = uc.repo.ApplyDiscountToCartItem(ctx, repository.ApplyDiscountToCartItemParams{
						ID:             item.ID,
						DiscountAmount: perItemNumeric,
						Column3:        promo.Code,
					})
				}
			}
		}
	}

	discountNumeric := pgtype.Numeric{}
	if err := discountNumeric.Scan(fmt.Sprintf("%.2f", discountFloat)); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to compute discount amount", nil)
	}

	// 7. Apply coupon to cart header
	_, err = uc.repo.ApplyCouponToCart(ctx, repository.ApplyCouponToCartParams{
		ID:             in.CartID,
		CouponCode:     pgtype.Text{String: in.CouponCode, Valid: true},
		DiscountAmount: discountNumeric,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to apply coupon to cart: "+err.Error(), nil)
	}

	// 8. Increment promotion usage counter atomically
	_, _ = uc.repo.IncrementPromotionUsage(ctx, promo.ID)

	// 9. Recalculate cart totals
	updatedCart, err := uc.repo.RecalculateCartTotals(ctx, in.CartID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to recalculate cart totals", nil)
	}

	responseData := map[string]interface{}{
		"cart":           updatedCart,
		"discount":       discountFloat,
		"promotion_name": promo.Name,
		"promotion_type": promo.PromotionType,
	}
	for k, v := range extraResponse {
		responseData[k] = v
	}

	return utils.NewResponse(utils.CodeOK, "coupon applied successfully", responseData)
}

// ValidateCoupon checks whether a coupon is valid for a cart without applying it.
// Returns the promotion details and calculated discount amount.
func (uc *PromotionUseCase) ValidateCoupon(ctx context.Context, in ApplyCouponInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	cart, err := uc.repo.GetCartByID(ctx, in.CartID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cart not found", nil)
	}
	if cart.CartStatus != repository.CartStatusActive {
		return utils.NewResponse(utils.CodeBadReq, "cart is not active", nil)
	}

	promo, err := uc.repo.GetActivePromotionByCouponCode(ctx, repository.GetActivePromotionByCouponCodeParams{
		CouponCode:     pgtype.Text{String: in.CouponCode, Valid: true},
		OrganizationID: in.OrganizationID,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "coupon code is invalid, expired, or has reached its usage limit", nil)
	}

	totals, err := uc.repo.GetCartTotals(ctx, in.CartID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch cart totals", nil)
	}

	subtotalFloat := interfaceToFloat(totals.Subtotal)
	var validationErrors []string

	if !isStoreAllowed(promo.StoreIds, cart.StoreID) {
		validationErrors = append(validationErrors, "coupon is not valid for this store")
	}

	if promo.MinOrderAmount.Valid {
		minAmount := numericToFloat(promo.MinOrderAmount)
		if subtotalFloat < minAmount {
			validationErrors = append(validationErrors,
				fmt.Sprintf("cart subtotal (%.2f) is below the minimum required (%.2f)", subtotalFloat, minAmount))
		}
	}

	if promo.MinQuantity.Valid {
		itemCount, err := uc.repo.GetCartItemCount(ctx, in.CartID)
		if err == nil {
			minQty := numericToFloat(promo.MinQuantity)
			totalQty := interfaceToFloat(itemCount.TotalQuantity)
			if totalQty < minQty {
				validationErrors = append(validationErrors,
					fmt.Sprintf("cart quantity (%.2f) is below the minimum required (%.2f)", totalQty, minQty))
			}
		}
	}

	// happy_hour schedule check
	if promo.PromotionType == "happy_hour" && len(promo.ScheduleJson) > 0 {
		var schedule happyHourSchedule
		if err := json.Unmarshal(promo.ScheduleJson, &schedule); err == nil {
			if !isHappyHourActive(schedule) {
				validationErrors = append(validationErrors, "promotion is only valid during happy hour")
			}
		}
	}

	return utils.NewResponse(utils.CodeOK, "coupon validation result", map[string]interface{}{
		"valid":          len(validationErrors) == 0,
		"errors":         validationErrors,
		"promotion_name": promo.Name,
		"promotion_type": promo.PromotionType,
		"discount_value": numericToFloat(promo.DiscountValue),
		"min_order":      numericToFloat(promo.MinOrderAmount),
	})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// applyProductSetDiscount applies a percentage discount to all cart items whose
// ProductID is in targetProductSet. Returns the total monetary discount applied.
func (uc *PromotionUseCase) applyProductSetDiscount(
	ctx context.Context,
	items []repository.ListCartItemsRow,
	targetProductSet map[int32]bool,
	discountPercent float64,
	promoCode string,
) float64 {
	var totalDiscount float64
	for _, item := range items {
		if targetProductSet[item.ProductID] {
			unitPrice := numericToFloat(item.UnitPrice)
			qty := numericToFloat(item.Quantity)
			lineDiscount := unitPrice * qty * (discountPercent / 100)
			totalDiscount += lineDiscount

			perItemNumeric := pgtype.Numeric{}
			_ = perItemNumeric.Scan(fmt.Sprintf("%.2f", lineDiscount))
			_, _ = uc.repo.ApplyDiscountToCartItem(ctx, repository.ApplyDiscountToCartItemParams{
				ID:             item.ID,
				DiscountAmount: perItemNumeric,
				Column3:        promoCode,
			})
		}
	}
	return totalDiscount
}

// isHappyHourActive returns true when the current system time falls within the
// schedule's days and time window. Times are compared in the server's local timezone.
func isHappyHourActive(schedule happyHourSchedule) bool {
	now := time.Now()
	currentDay := now.Weekday().String() // e.g. "Monday"

	dayMatches := false
	for _, d := range schedule.Days {
		if strings.EqualFold(d, currentDay) {
			dayMatches = true
			break
		}
	}
	if !dayMatches {
		return false
	}

	// Parse HH:MM times and build comparable time values for today
	parseHHMM := func(hhmm string) (time.Time, bool) {
		parts := strings.SplitN(hhmm, ":", 2)
		if len(parts) != 2 {
			return time.Time{}, false
		}
		var h, m int
		_, err1 := fmt.Sscanf(parts[0], "%d", &h)
		_, err2 := fmt.Sscanf(parts[1], "%d", &m)
		if err1 != nil || err2 != nil {
			return time.Time{}, false
		}
		return time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, now.Location()), true
	}

	start, okS := parseHHMM(schedule.StartTime)
	end, okE := parseHHMM(schedule.EndTime)
	if !okS || !okE {
		return false // invalid schedule; fail safe (don't apply)
	}

	return !now.Before(start) && now.Before(end)
}

// numericToFloat decodes a pgtype.Numeric to float64. Returns 0 on NULL/error.
func numericToFloat(n pgtype.Numeric) float64 {
	val, err := n.Value()
	if err != nil || val == nil {
		return 0
	}
	var f float64
	fmt.Sscanf(fmt.Sprintf("%v", val), "%f", &f)
	return f
}

// interfaceToFloat decodes a generic interface{} (from COALESCE SUM columns) to float64.
func interfaceToFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	var f float64
	fmt.Sscanf(fmt.Sprintf("%v", v), "%f", &f)
	return f
}

// isStoreAllowed returns true if storeIDs is empty or contains cartStoreID.
func isStoreAllowed(storeIDs []int32, cartStoreID pgtype.Int4) bool {
	if len(storeIDs) == 0 {
		return true
	}
	if !cartStoreID.Valid {
		return false
	}
	for _, id := range storeIDs {
		if id == cartStoreID.Int32 {
			return true
		}
	}
	return false
}
