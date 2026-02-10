package usecase

import (
	"context"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/google/uuid"
)

type OrderUseCase struct {
	repo *repository.Queries
}

func NewOrderUseCase() *OrderUseCase {
	return &OrderUseCase{}
}

func (uc *OrderUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *OrderUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

// GetOrder retrieves an order by its ID.
func (uc *OrderUseCase) GetOrder(ctx context.Context, orderID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	order, err := uc.repo.GetSalesOrderV2(ctx, orderID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "order not found", nil)
	}

	lines, err := uc.repo.ListSalesOrderLinesV2(ctx, orderID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch order lines", nil)
	}

	result := map[string]interface{}{
		"order": order,
		"lines": lines,
	}

	return utils.NewResponse(utils.CodeOK, "order fetched successfully", result)
}

func (uc *OrderUseCase) GetOrderByNumber(ctx context.Context, orderNumber string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	order, err := uc.repo.GetSalesOrderV2ByNumber(ctx, orderNumber)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "order not found", nil)
	}
	lines, err := uc.repo.ListSalesOrderLinesV2(ctx, order.ID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch order lines", nil)
	}
	return utils.NewResponse(utils.CodeOK, "order fetched successfully", map[string]interface{}{
		"order": order,
		"lines": lines,
	})
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, arg repository.CreateSalesOrderV2Params) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	order, err := uc.repo.CreateSalesOrderV2(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create order", err.Error())
	}
	return utils.NewResponse(utils.CodeCreated, "order created successfully", order)
}

func (uc *OrderUseCase) UpdateOrder(ctx context.Context, arg repository.UpdateSalesOrderV2Params) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	order, err := uc.repo.UpdateSalesOrderV2(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update order", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "order updated successfully", order)
}

func (uc *OrderUseCase) UpdateOrderStatus(ctx context.Context, arg repository.UpdateOrderStatusParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	order, err := uc.repo.UpdateOrderStatus(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update order status", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "order status updated", order)
}

func (uc *OrderUseCase) UpdateOrderPaymentStatus(ctx context.Context, arg repository.UpdateOrderPaymentStatusParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	order, err := uc.repo.UpdateOrderPaymentStatus(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update payment status", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "payment status updated", order)
}

func (uc *OrderUseCase) UpdateOrderFulfillmentStatus(ctx context.Context, arg repository.UpdateOrderFulfillmentStatusParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	order, err := uc.repo.UpdateOrderFulfillmentStatus(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update fulfillment status", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "fulfillment status updated", order)
}

func (uc *OrderUseCase) UpdateOrderTotals(ctx context.Context, arg repository.UpdateOrderTotalsParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	order, err := uc.repo.UpdateOrderTotals(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update totals", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "order totals updated", order)
}

func (uc *OrderUseCase) UpdateOrderDelivery(ctx context.Context, arg repository.UpdateOrderDeliveryParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	order, err := uc.repo.UpdateOrderDelivery(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update delivery", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "order delivery updated", order)
}

func (uc *OrderUseCase) AssignOrder(ctx context.Context, arg repository.AssignOrderParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	order, err := uc.repo.AssignOrder(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to assign order", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "order assigned", order)
}

func (uc *OrderUseCase) CancelOrder(ctx context.Context, orderID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	order, err := uc.repo.CancelOrder(ctx, orderID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to cancel order", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "order cancelled", order)
}

func (uc *OrderUseCase) DeleteOrder(ctx context.Context, orderID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if err := uc.repo.DeleteSalesOrderV2(ctx, orderID); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to delete order", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "order deleted", nil)
}

func (uc *OrderUseCase) CreateOrderLine(ctx context.Context, arg repository.CreateSalesOrderLineV2Params) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	line, err := uc.repo.CreateSalesOrderLineV2(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create order line", err.Error())
	}
	return utils.NewResponse(utils.CodeCreated, "order line created", line)
}

func (uc *OrderUseCase) GetOrderLine(ctx context.Context, lineID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	line, err := uc.repo.GetSalesOrderLineV2(ctx, lineID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "order line not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "order line fetched", line)
}

func (uc *OrderUseCase) ListOrderLines(ctx context.Context, orderID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	lines, err := uc.repo.ListSalesOrderLinesV2(ctx, orderID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list order lines", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "order lines listed", lines)
}

func (uc *OrderUseCase) UpdateOrderLine(ctx context.Context, arg repository.UpdateSalesOrderLineV2Params) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	line, err := uc.repo.UpdateSalesOrderLineV2(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update order line", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "order line updated", line)
}

func (uc *OrderUseCase) UpdateOrderLineFulfillment(ctx context.Context, arg repository.UpdateOrderLineFulfillmentParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	line, err := uc.repo.UpdateOrderLineFulfillment(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update line fulfillment", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "line fulfillment updated", line)
}

func (uc *OrderUseCase) UpdateOrderLineStatus(ctx context.Context, arg repository.UpdateOrderLineStatusParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	line, err := uc.repo.UpdateOrderLineStatus(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update line status", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "line status updated", line)
}

func (uc *OrderUseCase) DeleteOrderLine(ctx context.Context, lineID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if err := uc.repo.DeleteSalesOrderLineV2(ctx, lineID); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to delete order line", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "order line deleted", nil)
}

func (uc *OrderUseCase) GetOrderLineTotals(ctx context.Context, orderID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	row, err := uc.repo.GetOrderLineTotals(ctx, orderID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to get line totals", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "order line totals fetched", row)
}

func (uc *OrderUseCase) GetOrderLineMargin(ctx context.Context, orderID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	row, err := uc.repo.GetOrderLineMargin(ctx, orderID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to get margin", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "order line margin fetched", row)
}

func (uc *OrderUseCase) CreateOrderStatusHistory(ctx context.Context, arg repository.CreateOrderStatusHistoryParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	row, err := uc.repo.CreateOrderStatusHistory(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create status history", err.Error())
	}
	return utils.NewResponse(utils.CodeCreated, "status history created", row)
}

func (uc *OrderUseCase) ListOrderStatusHistory(ctx context.Context, orderID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	rows, err := uc.repo.ListOrderStatusHistory(ctx, orderID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list status history", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "status history listed", rows)
}

func (uc *OrderUseCase) CreateOrderFulfillment(ctx context.Context, arg repository.CreateOrderFulfillmentParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	row, err := uc.repo.CreateOrderFulfillment(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create fulfillment", err.Error())
	}
	return utils.NewResponse(utils.CodeCreated, "fulfillment created", row)
}

func (uc *OrderUseCase) GetOrderFulfillment(ctx context.Context, fulfillmentID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	row, err := uc.repo.GetOrderFulfillment(ctx, fulfillmentID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "fulfillment not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "fulfillment fetched", row)
}

func (uc *OrderUseCase) GetOrderFulfillmentByNumber(ctx context.Context, fulfillmentNumber string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	row, err := uc.repo.GetOrderFulfillmentByNumber(ctx, fulfillmentNumber)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "fulfillment not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "fulfillment fetched", row)
}

func (uc *OrderUseCase) ListOrderFulfillments(ctx context.Context, orderID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	rows, err := uc.repo.ListOrderFulfillments(ctx, orderID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list fulfillments", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "fulfillments listed", rows)
}

func (uc *OrderUseCase) UpdateOrderFulfillment(ctx context.Context, arg repository.UpdateOrderFulfillmentParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	row, err := uc.repo.UpdateOrderFulfillment(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update fulfillment", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "fulfillment updated", row)
}

func (uc *OrderUseCase) UpdateFulfillmentShipment(ctx context.Context, arg repository.UpdateFulfillmentShipmentParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	row, err := uc.repo.UpdateFulfillmentShipment(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update shipment", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "shipment updated", row)
}

func (uc *OrderUseCase) UpdateFulfillmentPickPack(ctx context.Context, arg repository.UpdateFulfillmentPickPackParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	row, err := uc.repo.UpdateFulfillmentPickPack(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update pick/pack", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "pick/pack updated", row)
}

func (uc *OrderUseCase) DeleteOrderFulfillment(ctx context.Context, fulfillmentID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if err := uc.repo.DeleteOrderFulfillment(ctx, fulfillmentID); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to delete fulfillment", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "fulfillment deleted", nil)
}

func (uc *OrderUseCase) CreateOrderFulfillmentItem(ctx context.Context, arg repository.CreateOrderFulfillmentItemParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	row, err := uc.repo.CreateOrderFulfillmentItem(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create fulfillment item", err.Error())
	}
	return utils.NewResponse(utils.CodeCreated, "fulfillment item created", row)
}

func (uc *OrderUseCase) ListOrderFulfillmentItems(ctx context.Context, fulfillmentID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	rows, err := uc.repo.ListOrderFulfillmentItems(ctx, fulfillmentID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list fulfillment items", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "fulfillment items listed", rows)
}

func (uc *OrderUseCase) DeleteOrderFulfillmentItem(ctx context.Context, itemID uuid.UUID) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if err := uc.repo.DeleteOrderFulfillmentItem(ctx, itemID); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to delete fulfillment item", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "fulfillment item deleted", nil)
}

// ListOrders lists orders with filters.
func (uc *OrderUseCase) ListOrders(ctx context.Context, orgID int32, status string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	arg := repository.ListSalesOrdersV2Params{
		OrganizationID: orgID,
		Limit:          50,
		Offset:         0,
	}

	if status != "" {
		arg.Column3 = repository.OrderStatusV2(status)
	}

	orders, err := uc.repo.ListSalesOrdersV2(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list orders", nil)
	}

	return utils.NewResponse(utils.CodeOK, "orders listed successfully", orders)
}
