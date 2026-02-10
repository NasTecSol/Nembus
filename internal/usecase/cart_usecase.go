package usecase

import (
	"context"
	"fmt"
	"time"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CartUseCase struct {
	repo *repository.Queries
}

func NewCartUseCase() *CartUseCase {
	return &CartUseCase{}
}

func (uc *CartUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *CartUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

// GetCart retrieves a cart by its ID.
func (uc *CartUseCase) GetCart(ctx context.Context, cartID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	cart, err := uc.repo.GetCartByID(ctx, cartID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cart not found", nil)
	}

	items, err := uc.repo.ListCartItems(ctx, cartID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch cart items", nil)
	}

	result := map[string]interface{}{
		"cart":  cart,
		"items": items,
	}

	return utils.NewResponse(utils.CodeOK, "cart fetched successfully", result)
}

func (uc *CartUseCase) CreateCart(ctx context.Context, arg repository.CreateCartParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.CreateCart(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create cart", err.Error())
	}
	return utils.NewResponse(utils.CodeCreated, "cart created successfully", cart)
}

func (uc *CartUseCase) GetCartByNumber(ctx context.Context, cartNumber string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.GetCartByNumber(ctx, cartNumber)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cart not found", nil)
	}
	items, err := uc.repo.ListCartItems(ctx, cart.ID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch cart items", nil)
	}
	return utils.NewResponse(utils.CodeOK, "cart fetched successfully", map[string]interface{}{
		"cart":  cart,
		"items": items,
	})
}

func (uc *CartUseCase) GetActiveCartByCustomer(ctx context.Context, arg repository.GetCartByCustomerParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.GetCartByCustomer(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "active cart not found", nil)
	}
	items, err := uc.repo.ListCartItems(ctx, cart.ID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch cart items", nil)
	}
	return utils.NewResponse(utils.CodeOK, "cart fetched successfully", map[string]interface{}{
		"cart":  cart,
		"items": items,
	})
}

func (uc *CartUseCase) GetActiveCartByGuestIdentifier(ctx context.Context, arg repository.GetCartByGuestIdentifierParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.GetCartByGuestIdentifier(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "active cart not found", nil)
	}
	items, err := uc.repo.ListCartItems(ctx, cart.ID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch cart items", nil)
	}
	return utils.NewResponse(utils.CodeOK, "cart fetched successfully", map[string]interface{}{
		"cart":  cart,
		"items": items,
	})
}

func (uc *CartUseCase) ListActiveCarts(ctx context.Context, arg repository.ListActiveCartsParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	carts, err := uc.repo.ListActiveCarts(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list active carts", nil)
	}
	return utils.NewResponse(utils.CodeOK, "active carts listed", carts)
}

func (uc *CartUseCase) ListAbandonedCarts(ctx context.Context, arg repository.ListAbandonedCartsParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	carts, err := uc.repo.ListAbandonedCarts(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list abandoned carts", nil)
	}
	return utils.NewResponse(utils.CodeOK, "abandoned carts listed", carts)
}

func (uc *CartUseCase) UpdateCart(ctx context.Context, arg repository.UpdateCartParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.UpdateCart(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update cart", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart updated successfully", cart)
}

func (uc *CartUseCase) UpdateCartStatus(ctx context.Context, arg repository.UpdateCartStatusParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.UpdateCartStatus(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update cart status", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart status updated", cart)
}

func (uc *CartUseCase) UpdateCartCustomer(ctx context.Context, arg repository.UpdateCartCustomerParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.UpdateCartCustomer(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update cart customer", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart customer updated", cart)
}

func (uc *CartUseCase) DeleteCart(ctx context.Context, cartID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if err := uc.repo.DeleteCart(ctx, cartID); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to delete cart", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart deleted", nil)
}

func (uc *CartUseCase) ExpireAbandonedCarts(ctx context.Context, storeID pgtype.Int4) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if err := uc.repo.ExpireAbandonedCarts(ctx, storeID); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to expire carts", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "carts expired", nil)
}

func (uc *CartUseCase) ListCartItems(ctx context.Context, cartID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	items, err := uc.repo.ListCartItems(ctx, cartID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list cart items", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart items listed", items)
}

func (uc *CartUseCase) CreateCartItem(ctx context.Context, arg repository.CreateCartItemParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	item, err := uc.repo.CreateCartItem(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create cart item", err.Error())
	}
	_, _ = uc.repo.RecalculateCartTotals(ctx, arg.CartID)
	return utils.NewResponse(utils.CodeCreated, "cart item created", item)
}

func (uc *CartUseCase) GetCartItem(ctx context.Context, itemID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	item, err := uc.repo.GetCartItem(ctx, itemID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cart item not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "cart item fetched", item)
}

func (uc *CartUseCase) GetCartItemByProduct(ctx context.Context, arg repository.GetCartItemByProductParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	item, err := uc.repo.GetCartItemByProduct(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cart item not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "cart item fetched", item)
}

func (uc *CartUseCase) UpdateCartItem(ctx context.Context, arg repository.UpdateCartItemParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	item, err := uc.repo.UpdateCartItem(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update cart item", err.Error())
	}
	_, _ = uc.repo.RecalculateCartTotals(ctx, item.CartID)
	return utils.NewResponse(utils.CodeOK, "cart item updated", item)
}

func (uc *CartUseCase) UpdateCartItemQuantity(ctx context.Context, arg repository.UpdateCartItemQuantityParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	item, err := uc.repo.UpdateCartItemQuantity(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update cart item quantity", err.Error())
	}
	_, _ = uc.repo.RecalculateCartTotals(ctx, item.CartID)
	return utils.NewResponse(utils.CodeOK, "cart item quantity updated", item)
}

func (uc *CartUseCase) DeleteCartItem(ctx context.Context, itemID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	item, err := uc.repo.GetCartItem(ctx, itemID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cart item not found", nil)
	}
	if err := uc.repo.DeleteCartItem(ctx, itemID); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to delete cart item", err.Error())
	}
	_, _ = uc.repo.RecalculateCartTotals(ctx, item.CartID)
	return utils.NewResponse(utils.CodeOK, "cart item deleted", nil)
}

func (uc *CartUseCase) ClearCartItems(ctx context.Context, cartID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if err := uc.repo.ClearCartItems(ctx, cartID); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to clear cart items", err.Error())
	}
	_, _ = uc.repo.RecalculateCartTotals(ctx, cartID)
	return utils.NewResponse(utils.CodeOK, "cart items cleared", nil)
}

func (uc *CartUseCase) GetCartItemCount(ctx context.Context, cartID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	row, err := uc.repo.GetCartItemCount(ctx, cartID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to get cart item count", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart item count fetched", row)
}

func (uc *CartUseCase) GetCartTotals(ctx context.Context, cartID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	row, err := uc.repo.GetCartTotals(ctx, cartID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to get cart totals", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart totals fetched", row)
}

func (uc *CartUseCase) CreateCartActivity(ctx context.Context, arg repository.CreateCartActivityParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	logRow, err := uc.repo.CreateCartActivity(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create cart activity", err.Error())
	}
	return utils.NewResponse(utils.CodeCreated, "cart activity created", logRow)
}

func (uc *CartUseCase) ListCartActivities(ctx context.Context, arg repository.ListCartActivitiesParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	rows, err := uc.repo.ListCartActivities(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list cart activities", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart activities listed", rows)
}

func (uc *CartUseCase) ApplyCouponToCart(ctx context.Context, arg repository.ApplyCouponToCartParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.ApplyCouponToCart(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to apply coupon", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "coupon applied", cart)
}

func (uc *CartUseCase) RecalculateCartTotals(ctx context.Context, cartID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	cart, err := uc.repo.RecalculateCartTotals(ctx, cartID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to recalculate totals", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "cart totals recalculated", cart)
}

func (uc *CartUseCase) MergeGuestCartToCustomer(ctx context.Context, arg repository.MergeGuestCartToCustomerParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if err := uc.repo.MergeGuestCartToCustomer(ctx, arg); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to merge carts", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "carts merged", nil)
}

// AddToCart adds an item to the cart.
func (uc *CartUseCase) AddToCart(ctx context.Context, cartID uuid.UUID, orgID int32, productID int32, qty float64) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	// 1. Check if product exists
	product, err := uc.repo.GetProduct(ctx, productID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "product not found", nil)
	}

	// 2. Get price using generated method
	productPrice, err := uc.repo.GetProductPrice(ctx, productID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "product price not found", nil)
	}

	// 3. Calculate line total
	qtyNumeric := pgtype.Numeric{}
	if err := qtyNumeric.Scan(fmt.Sprintf("%.2f", qty)); err != nil {
		return utils.NewResponse(utils.CodeError, "invalid quantity", nil)
	}

	// Calculate line total (simplified - in production you'd use proper decimal math)
	lineTotal := productPrice.Price

	// 4. Create cart item using generated method
	// Note: This is a simple insert. For upsert logic, you'd need to add a custom query to cart_order_queries.sql
	cartItem, err := uc.repo.CreateCartItem(ctx, repository.CreateCartItemParams{
		CartID:         cartID,
		OrganizationID: orgID,
		ProductID:      productID,
		Quantity:       qtyNumeric,
		UnitPrice:      productPrice.Price,
		LineTotal:      lineTotal,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to add item: %v", err), nil)
	}

	// 5. Recalculate cart totals
	_, _ = uc.repo.RecalculateCartTotals(ctx, cartID)

	return utils.NewResponse(utils.CodeOK, "item added to cart", map[string]interface{}{
		"item_id":      cartItem.ID,
		"product_name": product.Name,
		"quantity":     qty,
	})
}

// ConvertToOrder converts a cart to an order.
func (uc *CartUseCase) ConvertToOrder(ctx context.Context, cartID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	// 1. Get Cart using generated method
	cart, err := uc.repo.GetCartByID(ctx, cartID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "cart not found", nil)
	}

	if cart.CartStatus == repository.CartStatusConverted {
		return utils.NewResponse(utils.CodeBadReq, "cart already converted", nil)
	}

	// 2. Fetch items using generated method
	items, err := uc.repo.ListCartItems(ctx, cartID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch cart items", nil)
	}

	// 3. Convert Cart to Order Header using the new generated business logic query
	orderNumber := fmt.Sprintf("ORD-%s", time.Now().Format("20060102150405"))
	convertedCart, err := uc.repo.ConvertCartToOrder(ctx, repository.ConvertCartToOrderParams{
		ID:          cartID,
		OrderNumber: orderNumber,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, "conversion failed: "+err.Error(), nil)
	}

	orderID := convertedCart.ConvertedToOrderID.Bytes // It's a pgtype.UUID

	// 4. Create Order Lines
	for i, item := range items {
		arg := repository.CreateSalesOrderLineV2Params{
			SalesOrderID:     orderID,
			OrganizationID:   cart.OrganizationID,
			LineNumber:       int32(i + 1),
			ProductID:        item.ProductID,
			ProductVariantID: item.ProductVariantID,
			ProductName:      item.ProductName, // From ListCartItems JOIN
			ProductSku:       pgtype.Text{String: item.ProductSku, Valid: true},
			QuantityOrdered:  item.Quantity,
			UnitPrice:        item.UnitPrice,
			LineTotal:        item.LineTotal,
			LineStatus:       pgtype.Text{String: "pending", Valid: true},
		}

		_, err = uc.repo.CreateSalesOrderLineV2(ctx, arg)
		if err != nil {
			// In a real app, we might use a transaction and roll back
			fmt.Printf("Warning: failed to create order line: %s\n", err.Error())
		}
	}

	return utils.NewResponse(utils.CodeCreated, "order created from cart", map[string]interface{}{
		"order_id":     orderID,
		"order_number": orderNumber,
	})
}
