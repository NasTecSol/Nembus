package handler

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OrderZatcaReporter is supplied by the POS gateway after local checkout.
type OrderZatcaReporter interface {
	SubmitSalesOrder(context.Context, uuid.UUID) (*repository.ZatcaDocumentChain, error)
}

func autoReportSalesOrder(ctx context.Context, reporter OrderZatcaReporter, id uuid.UUID) gin.H {
	result := gin.H{"sales_order_id": id.String(), "reported": false}
	if reporter == nil {
		result["error"] = "automatic ZATCA reporting requires the POS gateway"
		return result
	}
	entry, err := reporter.SubmitSalesOrder(ctx, id)
	if errors.Is(err, usecase.ErrZatcaDisabled) {
		result["enabled"] = false
		return result
	}
	if entry != nil {
		result["generated"] = true
		result["qr_code"] = entry.QrCodeBase64.String
		result["zatca_uuid"] = entry.ZatcaUuid.String()
		if entry.ZatcaStatus.Valid {
			result["document_status"] = entry.ZatcaStatus.ZatcaDocStatus
		}
		if len(entry.ZatcaResponse) > 0 {
			result["zatca_response"] = json.RawMessage(entry.ZatcaResponse)
		}
	}
	if err != nil {
		result["error"] = err.Error()
		return result
	}
	result["reported"] = true
	result["reporting_status"] = "REPORTED"
	return result
}
