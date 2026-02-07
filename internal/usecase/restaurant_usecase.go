package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type RestaurantUseCase struct {
	repo *repository.Queries
}

func NewRestaurantUseCase() *RestaurantUseCase {
	return &RestaurantUseCase{}
}

func (uc *RestaurantUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// === Tables ===

func (uc *RestaurantUseCase) GetTable(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	table, err := uc.repo.GetRestaurantTable(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "table not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "table fetched successfully", table)
}

func (uc *RestaurantUseCase) ListTables(ctx context.Context, storeID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	tables, err := uc.repo.ListRestaurantTables(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "tables fetched successfully", tables)
}

func (uc *RestaurantUseCase) CreateTable(ctx context.Context, arg repository.CreateRestaurantTableParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	table, err := uc.repo.CreateRestaurantTable(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "table created successfully", table)
}

func (uc *RestaurantUseCase) UpdateTable(ctx context.Context, arg repository.UpdateRestaurantTableParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	table, err := uc.repo.UpdateRestaurantTable(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "table updated successfully", table)
}

func (uc *RestaurantUseCase) DeleteTable(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	err := uc.repo.DeleteRestaurantTable(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "table deleted successfully", nil)
}

// === Menu Categories ===

func (uc *RestaurantUseCase) GetMenuCategory(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	category, err := uc.repo.GetMenuCategory(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "category not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "category fetched successfully", category)
}

func (uc *RestaurantUseCase) ListMenuCategories(ctx context.Context, storeID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	categories, err := uc.repo.ListMenuCategories(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "menu categories fetched successfully", categories)
}

func (uc *RestaurantUseCase) CreateMenuCategory(ctx context.Context, arg repository.CreateMenuCategoryParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	category, err := uc.repo.CreateMenuCategory(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "category created successfully", category)
}

func (uc *RestaurantUseCase) UpdateMenuCategory(ctx context.Context, arg repository.UpdateMenuCategoryParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	category, err := uc.repo.UpdateMenuCategory(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "category updated successfully", category)
}

func (uc *RestaurantUseCase) DeleteMenuCategory(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	err := uc.repo.DeleteMenuCategory(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "category deleted successfully", nil)
}

// === Menu Items ===

func (uc *RestaurantUseCase) GetMenuItem(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	item, err := uc.repo.GetMenuItem(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "menu item not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "menu item fetched successfully", item)
}

func (uc *RestaurantUseCase) ListMenuItems(ctx context.Context, categoryID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	items, err := uc.repo.ListMenuItems(ctx, categoryID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "menu items fetched successfully", items)
}

func (uc *RestaurantUseCase) CreateMenuItem(ctx context.Context, arg repository.CreateMenuItemParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	item, err := uc.repo.CreateMenuItem(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "menu item created successfully", item)
}

func (uc *RestaurantUseCase) UpdateMenuItem(ctx context.Context, arg repository.UpdateMenuItemParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	item, err := uc.repo.UpdateMenuItem(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "menu item updated successfully", item)
}

func (uc *RestaurantUseCase) DeleteMenuItem(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	err := uc.repo.DeleteMenuItem(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "menu item deleted successfully", nil)
}

func (uc *RestaurantUseCase) GetFullMenu(ctx context.Context, storeID int32, categoryID *int32, includeUnavail bool) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	menu, err := uc.repo.ListRestaurantMenuView(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	var result []repository.VwRestaurantMenu
	for _, item := range menu {
		if categoryID != nil && item.CategoryID != *categoryID {
			continue
		}
		if !includeUnavail && !item.IsAvailable.Bool {
			continue
		}
		result = append(result, item)
	}

	return utils.NewResponse(utils.CodeOK, "menu fetched successfully", result)
}

// === Modifiers ===

func (uc *RestaurantUseCase) ListModifiers(ctx context.Context, itemID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	modifiers, err := uc.repo.ListMenuItemModifiers(ctx, itemID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "modifiers fetched successfully", modifiers)
}

func (uc *RestaurantUseCase) GetModifier(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	mod, err := uc.repo.GetMenuItemModifier(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "modifier not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "modifier fetched successfully", mod)
}

func (uc *RestaurantUseCase) CreateModifier(ctx context.Context, arg repository.CreateMenuItemModifierParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	mod, err := uc.repo.CreateMenuItemModifier(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "modifier created successfully", mod)
}

func (uc *RestaurantUseCase) UpdateModifier(ctx context.Context, arg repository.UpdateMenuItemModifierParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	mod, err := uc.repo.UpdateMenuItemModifier(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "modifier updated successfully", mod)
}

func (uc *RestaurantUseCase) DeleteModifier(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	err := uc.repo.DeleteMenuItemModifier(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "modifier deleted successfully", nil)
}

// === Orders ===

func (uc *RestaurantUseCase) CreateOrder(ctx context.Context, arg repository.CreateRestaurantOrderParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if arg.OrderNumber == "" {
		arg.OrderNumber = fmt.Sprintf("ORD-%d", time.Now().UnixNano())
	}
	order, err := uc.repo.CreateRestaurantOrder(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "order created successfully", order)
}

func (uc *RestaurantUseCase) UpdateOrder(ctx context.Context, arg repository.UpdateRestaurantOrderParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	order, err := uc.repo.UpdateRestaurantOrder(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "order updated successfully", order)
}

func (uc *RestaurantUseCase) DeleteOrder(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	err := uc.repo.DeleteRestaurantOrder(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "order deleted successfully", nil)
}

func (uc *RestaurantUseCase) GetOrder(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	order, err := uc.repo.GetRestaurantOrder(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "order not found", nil)
	}
	items, _ := uc.repo.ListRestaurantOrderItems(ctx, order.ID)

	result := map[string]interface{}{
		"order": order,
		"items": items,
	}
	return utils.NewResponse(utils.CodeOK, "order fetched successfully", result)
}

func (uc *RestaurantUseCase) UpdateOrderStatus(ctx context.Context, orderID int32, status string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	order, err := uc.repo.UpdateRestaurantOrderStatus(ctx, repository.UpdateRestaurantOrderStatusParams{
		ID:     orderID,
		Status: status,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "order status updated successfully", order)
}

func (uc *RestaurantUseCase) GetOrderItem(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	item, err := uc.repo.GetRestaurantOrderItem(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "order item not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "order item fetched successfully", item)
}

func (uc *RestaurantUseCase) UpdateOrderItem(ctx context.Context, arg repository.UpdateRestaurantOrderItemParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	item, err := uc.repo.UpdateRestaurantOrderItem(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "order item updated successfully", item)
}

func (uc *RestaurantUseCase) DeleteOrderItem(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	err := uc.repo.DeleteRestaurantOrderItem(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "order item deleted successfully", nil)
}

func (uc *RestaurantUseCase) SettleOrder(ctx context.Context, orderID int32, posTxnID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	order, err := uc.repo.GetRestaurantOrder(ctx, orderID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "order not found", nil)
	}

	updatedOrder, err := uc.repo.UpdateRestaurantOrder(ctx, repository.UpdateRestaurantOrderParams{
		ID:               order.ID,
		TableID:          order.TableID,
		CashierID:        order.CashierID,
		CashierSessionID: order.CashierSessionID,
		CustomerID:       order.CustomerID,
		Status:           "paid",
		Subtotal:         order.Subtotal,
		DiscountAmount:   order.DiscountAmount,
		TaxAmount:        order.TaxAmount,
		TotalAmount:      order.TotalAmount,
		AmountPaid:       order.TotalAmount,
		ChangeGiven:      pgtype.Numeric{Int: big.NewInt(0), Exp: 0, Valid: true},
		Notes:            order.Notes,
		PosTransactionID: pgtype.Int4{Int32: posTxnID, Valid: true},
		ConfirmedAt:      order.ConfirmedAt,
		ServedAt:         order.ServedAt,
		PaidAt:           pgtype.Timestamp{Time: time.Now(), Valid: true},
		Metadata:         order.Metadata,
	})

	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "order settled successfully", updatedOrder)
}

// === Online Orders ===

func (uc *RestaurantUseCase) CreateOnlineOrder(ctx context.Context, storeID int32, customerID *int32, items []repository.CreateRestaurantOrderItemParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	orderParams := repository.CreateRestaurantOrderParams{
		StoreID:     storeID,
		OrderNumber: fmt.Sprintf("WEB-%d", time.Now().Unix()),
		OrderSource: "online",
		Status:      "pending",
		OrderedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
	}
	if customerID != nil {
		orderParams.CustomerID = pgtype.Int4{Int32: *customerID, Valid: true}
	}

	order, err := uc.repo.CreateRestaurantOrder(ctx, orderParams)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	for _, item := range items {
		item.OrderID = order.ID
		_, _ = uc.repo.CreateRestaurantOrderItem(ctx, item)
	}

	return utils.NewResponse(utils.CodeCreated, "online order created successfully", order)
}

// === KDS ===

func (uc *RestaurantUseCase) GetKdsOrders(ctx context.Context, storeID int32, statuses []string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	orders, err := uc.repo.ListActiveRestaurantOrdersView(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	var result []repository.VwActiveRestaurantOrder
	statusMap := make(map[string]bool)
	for _, s := range statuses {
		statusMap[s] = true
	}

	for _, o := range orders {
		if statusMap[o.OrderStatus] {
			result = append(result, o)
		}
	}

	return utils.NewResponse(utils.CodeOK, "KDS orders fetched successfully", result)
}

// === Waste ===

func (uc *RestaurantUseCase) CreateWasteLog(ctx context.Context, arg repository.CreateWasteLogParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	log, err := uc.repo.CreateWasteLog(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "waste log created successfully", log)
}

func (uc *RestaurantUseCase) GetWasteLog(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	log, err := uc.repo.GetWasteLog(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "waste log not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "waste log fetched successfully", log)
}

func (uc *RestaurantUseCase) UpdateWasteLog(ctx context.Context, arg repository.UpdateWasteLogParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	log, err := uc.repo.UpdateWasteLog(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "waste log updated successfully", log)
}

func (uc *RestaurantUseCase) DeleteWasteLog(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	err := uc.repo.DeleteWasteLog(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "waste log deleted successfully", nil)
}

func (uc *RestaurantUseCase) CreateWasteLogStandalone(ctx context.Context, arg repository.CreateWasteLogParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	log, err := uc.repo.CreateWasteLog(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "waste log created successfully", log)
}

func (uc *RestaurantUseCase) GetWasteReport(ctx context.Context, storeID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	summary, err := uc.repo.GetWasteDailySummaryView(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "waste daily summary fetched", summary)
}

// === Recipes ===

func (uc *RestaurantUseCase) CreateRecipe(ctx context.Context, arg repository.CreateRecipeParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	recipe, err := uc.repo.CreateRecipe(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "recipe created successfully", recipe)
}

func (uc *RestaurantUseCase) UpdateRecipe(ctx context.Context, arg repository.UpdateRecipeParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	recipe, err := uc.repo.UpdateRecipe(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "recipe updated successfully", recipe)
}

func (uc *RestaurantUseCase) DeleteRecipe(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	err := uc.repo.DeleteRecipe(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "recipe deleted successfully", nil)
}

func (uc *RestaurantUseCase) ListRecipes(ctx context.Context, orgID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	recipes, err := uc.repo.ListRecipes(ctx, orgID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "recipes fetched successfully", recipes)
}

func (uc *RestaurantUseCase) GetRecipe(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	recipe, err := uc.repo.GetRecipe(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "recipe not found", nil)
	}
	ingredients, _ := uc.repo.ListRecipeIngredients(ctx, recipe.ID)
	bom, _ := uc.repo.ListRecipeBomView(ctx, recipe.ID)

	totalCost := big.NewFloat(0)
	for _, item := range bom {
		if !item.IsByproduct.Bool && !item.IsOptional.Bool {
			// LineCostEstimate is Numeric in SQLC, which we override to pgtype.Numeric in sqlc.yaml
			// Let's assume we can convert it to big.Float
			if item.LineCostEstimate.Valid {
				f, _ := item.LineCostEstimate.Float64Value()
				if f.Valid {
					totalCost.Add(totalCost, big.NewFloat(f.Float64))
				}
			}
		}
	}

	result := map[string]interface{}{
		"recipe":      recipe,
		"ingredients": ingredients,
		"bom":         bom,
		"total_cost":  totalCost.String(),
	}
	return utils.NewResponse(utils.CodeOK, "recipe fetched successfully", result)
}

func (uc *RestaurantUseCase) GetRecipeIngredient(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	ing, err := uc.repo.GetRecipeIngredient(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "ingredient not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "ingredient fetched successfully", ing)
}

func (uc *RestaurantUseCase) UpdateRecipeIngredient(ctx context.Context, arg repository.UpdateRecipeIngredientParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	ing, err := uc.repo.UpdateRecipeIngredient(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "ingredient updated successfully", ing)
}

func (uc *RestaurantUseCase) DeleteRecipeIngredient(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	err := uc.repo.DeleteRecipeIngredient(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "ingredient deleted successfully", nil)
}

func (uc *RestaurantUseCase) AddRecipeIngredient(ctx context.Context, arg repository.CreateRecipeIngredientParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	ing, err := uc.repo.CreateRecipeIngredient(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "ingredient added successfully", ing)
}

// === Kiosk ===

func (uc *RestaurantUseCase) CreateKioskSession(ctx context.Context, arg repository.CreateKioskSessionParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	session, err := uc.repo.CreateKioskSession(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "kiosk session created successfully", session)
}

func (uc *RestaurantUseCase) UpdateKioskSession(ctx context.Context, arg repository.UpdateKioskSessionParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	session, err := uc.repo.UpdateKioskSession(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "kiosk session updated successfully", session)
}

func (uc *RestaurantUseCase) DeleteKioskSession(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	err := uc.repo.DeleteKioskSession(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "kiosk session deleted successfully", nil)
}

func (uc *RestaurantUseCase) GetKioskSessionByID(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	session, err := uc.repo.GetKioskSession(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "kiosk session not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "kiosk session fetched successfully", session)
}

func (uc *RestaurantUseCase) GetKioskSession(ctx context.Context, token string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	session, err := uc.repo.GetKioskSessionByToken(ctx, token)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "kiosk session not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "kiosk session fetched successfully", session)
}

// Utility to parse JSON RawMessage
func bytesToMap(b []byte) map[string]interface{} {
	var m map[string]interface{}
	json.Unmarshal(b, &m)
	return m
}
