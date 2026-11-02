package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type SalesReturnUseCase struct {
	repo    *repository.Queries
	zatcaUC *ZatcaUseCase
}

func NewSalesReturnUseCase() *SalesReturnUseCase {
	return &SalesReturnUseCase{}
}

func (uc *SalesReturnUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *SalesReturnUseCase) SetZatcaUseCase(zuc *ZatcaUseCase) {
	uc.zatcaUC = zuc
}

func (uc *SalesReturnUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

type ProcessReturnInput struct {
	StoreID               int32
	CashierID             *int32
	SessionID             *int32
	OriginalTransactionID *int32
	CustomerID            *int32
	ReturnReason          string
	Subtotal              string
	TaxAmount             string
	TotalRefundAmount     string
	RefundMethod          string
	RefundReference       string
	Lines                 []ProcessReturnLineInput
	// AuthorizedByUserID is set when a supervisor override was used to allow this return.
	// It is stored in the sales_return metadata for audit trail purposes.
	AuthorizedByUserID *int32
}

type ProcessReturnLineInput struct {
	ProductID        int32
	ProductVariantID *int32
	OriginalLineID   *int32
	Quantity         string
	UnitPrice        string
	RefundAmount     string
	ReturnToStock    bool
	SerialNumber     *string
	BatchNumber      *string
	Condition        string
	CostPrice        string
}

func (uc *SalesReturnUseCase) ProcessSalesReturn(ctx context.Context, in ProcessReturnInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	// 1. Prepare Return Header
	returnNumber := fmt.Sprintf("RET-%d", time.Now().Unix())

	subtotal, _ := uc.repo.ParseNumeric(ctx, in.Subtotal)
	taxAmount, _ := uc.repo.ParseNumeric(ctx, in.TaxAmount)
	totalRefund, _ := uc.repo.ParseNumeric(ctx, in.TotalRefundAmount)

	// Build metadata — include supervisor audit info when an override was used
	var metaBytes []byte
	if in.AuthorizedByUserID != nil {
		type returnMeta struct {
			AuthorizedByUserID int32 `json:"authorized_by_user_id"`
		}
		if b, err := json.Marshal(returnMeta{AuthorizedByUserID: *in.AuthorizedByUserID}); err == nil {
			metaBytes = b
		}
	}

	headerParams := repository.CreateSalesReturnParams{
		ReturnNumber:      returnNumber,
		StoreID:           in.StoreID,
		ReturnDate:        pgtype.Timestamp{Time: time.Now(), Valid: true},
		ReturnReason:      pgtype.Text{String: in.ReturnReason, Valid: in.ReturnReason != ""},
		Status:            pgtype.Text{String: "completed", Valid: true},
		Subtotal:          subtotal,
		TaxAmount:         taxAmount,
		TotalRefundAmount: totalRefund,
		RefundMethod:      pgtype.Text{String: in.RefundMethod, Valid: in.RefundMethod != ""},
		RefundReference:   pgtype.Text{String: in.RefundReference, Valid: in.RefundReference != ""},
		Metadata:          metaBytes,
	}

	if in.CashierID != nil {
		headerParams.CashierID = pgtype.Int4{Int32: *in.CashierID, Valid: true}
	}
	if in.SessionID != nil {
		headerParams.CashierSessionID = pgtype.Int4{Int32: *in.SessionID, Valid: true}
	}
	if in.OriginalTransactionID != nil {
		headerParams.OriginalTransactionID = pgtype.Int4{Int32: *in.OriginalTransactionID, Valid: true}
	}
	if in.CustomerID != nil {
		headerParams.CustomerID = pgtype.Int4{Int32: *in.CustomerID, Valid: true}
	}

	salesReturn, err := uc.repo.CreateSalesReturn(ctx, headerParams)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create return header: "+err.Error(), nil)
	}

	// 2. Prepare and Create Return Lines
	createdReturnLines := make([]repository.SalesReturnLine, 0, len(in.Lines))
	for i, l := range in.Lines {
		qty, _ := uc.repo.ParseNumeric(ctx, l.Quantity)
		uPrice, _ := uc.repo.ParseNumeric(ctx, l.UnitPrice)
		refAmount, _ := uc.repo.ParseNumeric(ctx, l.RefundAmount)

		// Validation: Check if returning more than available quantity on original line
		if l.OriginalLineID != nil {
			prevReturned, prevErr := uc.repo.GetReturnedQuantityByLineID(ctx, pgtype.Int4{Int32: *l.OriginalLineID, Valid: true})
			if prevErr == nil && prevReturned.Valid && qty.Valid {
				var prevQtyFloat float64
				if vPrev, vErr := prevReturned.Value(); vErr == nil && vPrev != nil {
					if s, ok := vPrev.(string); ok {
						prevQtyFloat, _ = strconv.ParseFloat(s, 64)
					}
				}
				var reqQtyFloat float64
				if vReq, vErr := qty.Value(); vErr == nil && vReq != nil {
					if s, ok := vReq.(string); ok {
						reqQtyFloat, _ = strconv.ParseFloat(s, 64)
					}
				}
				if reqQtyFloat <= 0 {
					return utils.NewResponse(utils.CodeError, fmt.Sprintf("invalid return quantity %.3f for line %d: must be greater than zero", reqQtyFloat, i+1), nil)
				}
				origLine, origErr := uc.repo.GetPosTransactionLineByID(ctx, *l.OriginalLineID)
				if origErr == nil && origLine.Quantity.Valid {
					var origQtyFloat float64
					if vOrig, vErr := origLine.Quantity.Value(); vErr == nil && vOrig != nil {
						if s, ok := vOrig.(string); ok {
							origQtyFloat, _ = strconv.ParseFloat(s, 64)
						}
					}
					if origQtyFloat > 0 && (reqQtyFloat+prevQtyFloat) > origQtyFloat {
						return utils.NewResponse(utils.CodeError, fmt.Sprintf("return quantity %.3f exceeds available returnable quantity %.3f (original: %.3f, previously returned: %.3f)", reqQtyFloat, origQtyFloat-prevQtyFloat, origQtyFloat, prevQtyFloat), nil)
					}
				}
			}
		}

		lineParams := repository.CreateSalesReturnLineParams{
			ReturnID:      salesReturn.ID,
			ProductID:     l.ProductID,
			Quantity:      qty,
			UnitPrice:     uPrice,
			RefundAmount:  refAmount,
			ReturnToStock: pgtype.Bool{Bool: l.ReturnToStock, Valid: true},
			Condition:     pgtype.Text{String: l.Condition, Valid: l.Condition != ""},
			LineNumber:    pgtype.Int4{Int32: int32(i + 1), Valid: true},
		}

		if l.ProductVariantID != nil {
			lineParams.ProductVariantID = pgtype.Int4{Int32: *l.ProductVariantID, Valid: true}
		}
		if l.OriginalLineID != nil {
			lineParams.OriginalLineID = pgtype.Int4{Int32: *l.OriginalLineID, Valid: true}
		}
		if l.SerialNumber != nil {
			lineParams.SerialNumber = pgtype.Text{String: *l.SerialNumber, Valid: true}
		}
		if l.BatchNumber != nil {
			lineParams.BatchNumber = pgtype.Text{String: *l.BatchNumber, Valid: true}
		}

		createdLine, err := uc.repo.CreateSalesReturnLine(ctx, lineParams)
		if err != nil {
			return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to create return line %d: %s", i+1, err.Error()), nil)
		}
		createdReturnLines = append(createdReturnLines, createdLine)

		// F2.3: Restock inventory when ReturnToStock is true and condition is good or empty
		isGoodCondition := l.Condition == "" || strings.ToLower(l.Condition) == "good"
		if l.ReturnToStock && isGoodCondition {
			_, adjErr := uc.repo.AdjustInventoryStockByProductAndStore(ctx, repository.AdjustInventoryStockByProductAndStoreParams{
				ProductID:         l.ProductID,
				ProductVariantID:  lineParams.ProductVariantID,
				StoreID:           in.StoreID,
				QuantityOnHand:    qty,
				QuantityAvailable: qty,
			})
			if adjErr != nil {
				fmt.Printf("Warning: failed to adjust inventory stock on return: %v\n", adjErr)
			}

			// Determine unit cost and total cost for stock movement audit
			unitCost := uPrice
			if l.CostPrice != "" {
				if parsedCost, err := uc.repo.ParseNumeric(ctx, l.CostPrice); err == nil && parsedCost.Valid {
					unitCost = parsedCost
				}
			} else {
				if prod, pErr := uc.repo.GetProduct(ctx, l.ProductID); pErr == nil && prod.CostPrice.Valid {
					unitCost = prod.CostPrice
				}
			}

			totalCost := refAmount
			var costVal, qtyVal float64
			if cV, err := unitCost.Value(); err == nil && cV != nil {
				if s, ok := cV.(string); ok {
					costVal, _ = strconv.ParseFloat(s, 64)
				}
			}
			if qV, err := qty.Value(); err == nil && qV != nil {
				if s, ok := qV.(string); ok {
					qtyVal, _ = strconv.ParseFloat(s, 64)
				}
			}
			if costVal > 0 && qtyVal > 0 {
				if parsedTot, err := uc.repo.ParseNumeric(ctx, fmt.Sprintf("%.2f", costVal*qtyVal)); err == nil && parsedTot.Valid {
					totalCost = parsedTot
				}
			}

			_, movErr := uc.repo.CreateStockMovement(ctx, repository.CreateStockMovementParams{
				MovementType:     "return",
				ReferenceType:    pgtype.Text{String: "sales_return", Valid: true},
				ReferenceID:      pgtype.Int4{Int32: salesReturn.ID, Valid: true},
				ProductID:        l.ProductID,
				ProductVariantID: lineParams.ProductVariantID,
				ToStoreID:        pgtype.Int4{Int32: in.StoreID, Valid: true},
				Quantity:         qty,
				BatchNumber:      lineParams.BatchNumber,
				SerialNumber:     lineParams.SerialNumber,
				MovementDate:     pgtype.Timestamp{Time: time.Now(), Valid: true},
				Status:           pgtype.Text{String: "completed", Valid: true},
				CostPerUnit:      unitCost,
				TotalValue:       totalCost,
				Metadata:         json.RawMessage([]byte("{}")),
			})
			if movErr != nil {
				fmt.Printf("Warning: failed to create stock movement on return: %v\n", movErr)
			}
		}
	}

	// 3. Update original POS transaction status if referenced
	if in.OriginalTransactionID != nil {
		newStatus := "partially_refunded"
		origTxn, err := uc.repo.GetPosTransaction(ctx, *in.OriginalTransactionID)
		if err == nil && origTxn.TotalAmount.Valid && totalRefund.Valid {
			if origTxn.TotalAmount.Int != nil && totalRefund.Int != nil && origTxn.TotalAmount.Int.Cmp(totalRefund.Int) <= 0 {
				newStatus = "refunded"
			}
		}
		err = uc.repo.UpdatePOSTransactionStatus(ctx, repository.UpdatePOSTransactionStatusParams{
			ID:     *in.OriginalTransactionID,
			Status: pgtype.Text{String: newStatus, Valid: true},
		})
		if err != nil {
			fmt.Printf("Warning: failed to update pos_transaction status: %s\n", err.Error())
		}
	}

	// 4. Adjust drawer expected_balance for cash/drawer refunds
	if in.SessionID != nil && totalRefund.Valid && (in.RefundMethod == "cash" || in.RefundMethod == "") {
		negRefund := totalRefund
		if negRefund.Int != nil {
			negRefund = pgtype.Numeric{Int: new(big.Int).Neg(negRefund.Int), Exp: negRefund.Exp, Valid: true}
		}

		err = uc.repo.UpdateSessionExpectedBalance(ctx, repository.UpdateSessionExpectedBalanceParams{
			ID:              *in.SessionID,
			ExpectedBalance: negRefund,
		})
		if err != nil {
			fmt.Printf("Error: failed to update session balance for refund: %s\n", err.Error())
		}
	}

	// 5. F2.4: ZATCA Phase 2 Credit Note Signing (381)
	var qrCode string
	var zatcaUUID string
	if uc.zatcaUC != nil && uc.zatcaUC.GetConfig() != nil {
		store, storeErr := uc.repo.GetStore(ctx, in.StoreID)
		if storeErr == nil {
			org, orgErr := uc.repo.GetOrganization(ctx, store.OrganizationID)
			enabled, enableErr := OrganizationZatcaEnabled(&org)
			if orgErr == nil && enableErr == nil && enabled {
				var device repository.ZatcaDeviceConfig
				var devErr error
				var origInvoiceNumber string
				var originalDeviceID int32
				if in.OriginalTransactionID != nil {
					origTxn, err := uc.repo.GetPosTransaction(ctx, *in.OriginalTransactionID)
					if err == nil {
						origInvoiceNumber, _ = uc.zatcaUC.OriginalIssuedInvoiceNumber(ctx, &origTxn)
						if doc, err := uc.zatcaUC.IssuedTransactionDocument(ctx, &origTxn); err == nil {
							originalDeviceID = doc.DeviceConfigID
						}
					}
				}
				if originalDeviceID > 0 {
					device, devErr = uc.repo.GetZatcaDeviceConfig(ctx, originalDeviceID)
				} else {
					var terminal pgtype.Int4
					if in.SessionID != nil {
						if session, err := uc.repo.GetSessionByID(ctx, *in.SessionID); err == nil {
							terminal = pgtype.Int4{Int32: session.PosTerminalID, Valid: true}
						}
					}
					device, devErr = uc.zatcaUC.SelectZatcaDevice(ctx, store.OrganizationID, terminal, salesReturn.Metadata)
				}
				if devErr == nil {
					if device.OrganizationID != store.OrganizationID || !device.IsActive.Bool || device.IsRevoked.Bool {
						devErr = fmt.Errorf("original ZATCA device is unavailable")
					}
					if _, err := ZatcaDeviceEnvironment(&device); err != nil {
						devErr = err
					}
				}
				if devErr == nil {

					// ZATCA Credit Note (381) strictly requires BillingReference to the original tax invoice
					if origInvoiceNumber != "" {
						signedDoc, _, signErr := uc.zatcaUC.SignCreditNote(ctx, &salesReturn, createdReturnLines, origInvoiceNumber, &org, &device)
						if signErr == nil && signedDoc != nil {
							qrCode = signedDoc.QRCode
							zatcaUUID = signedDoc.ZatcaUUID

							var metaMap map[string]interface{}
							if len(salesReturn.Metadata) > 0 {
								_ = json.Unmarshal(salesReturn.Metadata, &metaMap)
							}
							if metaMap == nil {
								metaMap = make(map[string]interface{})
							}
							metaMap["qr_code_base64"] = signedDoc.QRCode
							metaMap["icv"] = signedDoc.ICV
							metaMap["zatca_uuid"] = signedDoc.ZatcaUUID
							if updatedMeta, mErr := json.Marshal(metaMap); mErr == nil {
								salesReturn.Metadata = updatedMeta
								if uErr := uc.repo.UpdateSalesReturnMetadata(ctx, repository.UpdateSalesReturnMetadataParams{
									ID:       salesReturn.ID,
									Metadata: updatedMeta,
								}); uErr != nil {
									fmt.Printf("Warning: failed to persist sales_return metadata: %v\n", uErr)
								}
							}
						}
					}
				}
			}
		}
	}

	return utils.NewResponse(utils.CodeCreated, "sales return processed successfully", ProcessSalesReturnOutput{
		SalesReturn: salesReturn,
		Lines:       createdReturnLines,
		QRCode:      qrCode,
		ZatcaUUID:   zatcaUUID,
	})
}

// ProcessSalesReturnOutput represents the result of processing a sales return.
type ProcessSalesReturnOutput struct {
	repository.SalesReturn
	Lines     []repository.SalesReturnLine `json:"lines"`
	QRCode    string                       `json:"qr_code,omitempty"`
	ZatcaUUID string                       `json:"zatca_uuid,omitempty"`
}

type SalesReturnResult = ProcessSalesReturnOutput
