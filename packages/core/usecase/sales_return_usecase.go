package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SalesReturnUseCase struct {
	repo    *repository.Queries
	pool    *pgxpool.Pool
	zatcaUC *ZatcaUseCase
}

func NewSalesReturnUseCase() *SalesReturnUseCase {
	return &SalesReturnUseCase{}
}

func (uc *SalesReturnUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *SalesReturnUseCase) SetPool(pool *pgxpool.Pool) {
	uc.pool = pool
}

func (uc *SalesReturnUseCase) SetZatcaUseCase(zuc *ZatcaUseCase) {
	uc.zatcaUC = zuc
}

// GenerateReturnNumber creates a collision-safe identifier for sales_returns.return_number.
// It combines a timestamp prefix (down to second) with 12 hex characters of cryptographic entropy (29 chars total),
// guaranteeing uniqueness within the 50-character column limit.
func GenerateReturnNumber(t time.Time) (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		// Fallback to nanosecond entropy if crypto/rand is unavailable
		return fmt.Sprintf("RET-%s-%012x", t.Format("20060102150405"), time.Now().UnixNano()&0xFFFFFFFFFFFF), nil
	}
	return fmt.Sprintf("RET-%s-%s", t.Format("20060102150405"), hex.EncodeToString(b)), nil
}

func numericToRat(n pgtype.Numeric) *big.Rat {
	if !n.Valid || n.Int == nil {
		return new(big.Rat)
	}
	r := new(big.Rat).SetInt(n.Int)
	if n.Exp > 0 {
		mul := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp)), nil)
		r.Mul(r, new(big.Rat).SetInt(mul))
	} else if n.Exp < 0 {
		denom := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-n.Exp)), nil)
		r.Quo(r, new(big.Rat).SetInt(denom))
	}
	return r
}

func parseQuantity(n pgtype.Numeric, raw string) *big.Rat {
	if n.Valid && n.Int != nil {
		return numericToRat(n)
	}
	raw = strings.TrimSpace(raw)
	if raw != "" {
		if r, ok := new(big.Rat).SetString(raw); ok {
			return r
		}
	}
	return new(big.Rat)
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

	activeRepo := uc.repo
	var tx pgx.Tx
	if pool := resolvePoolFromContext(ctx, uc.pool); pool != nil {
		startedTx, txErr := pool.Begin(ctx)
		if txErr != nil {
			return utils.NewResponse(utils.CodeError, "failed to begin sales return transaction: "+txErr.Error(), nil)
		}
		tx = startedTx
		defer func() {
			if tx != nil {
				_ = tx.Rollback(ctx)
			}
		}()
		activeRepo = uc.repo.WithTx(tx)
	}

	// 1. Validate all lines and group quantities by OriginalLineID before any DB mutations
	type preparedLine struct {
		index     int
		input     ProcessReturnLineInput
		qty       pgtype.Numeric
		uPrice    pgtype.Numeric
		refAmount pgtype.Numeric
		qtyRat    *big.Rat
	}

	preparedLines := make([]preparedLine, len(in.Lines))
	var originalLineIDs []int32
	seenOrigIDs := make(map[int32]bool)
	requestedPerOrigLine := make(map[int32]*big.Rat)

	for i, l := range in.Lines {
		qty, _ := uc.repo.ParseNumeric(ctx, l.Quantity)
		uPrice, _ := uc.repo.ParseNumeric(ctx, l.UnitPrice)
		refAmount, _ := uc.repo.ParseNumeric(ctx, l.RefundAmount)

		qtyRat := parseQuantity(qty, l.Quantity)
		if qtyRat.Sign() <= 0 {
			return utils.NewResponse(utils.CodeError, fmt.Sprintf("invalid return quantity for line %d: must be greater than zero", i+1), nil)
		}

		preparedLines[i] = preparedLine{
			index:     i,
			input:     l,
			qty:       qty,
			uPrice:    uPrice,
			refAmount: refAmount,
			qtyRat:    qtyRat,
		}

		if l.OriginalLineID != nil {
			origID := *l.OriginalLineID
			if !seenOrigIDs[origID] {
				seenOrigIDs[origID] = true
				originalLineIDs = append(originalLineIDs, origID)
			}
			if curr, ok := requestedPerOrigLine[origID]; ok {
				requestedPerOrigLine[origID] = new(big.Rat).Add(curr, qtyRat)
			} else {
				requestedPerOrigLine[origID] = new(big.Rat).Set(qtyRat)
			}
		}
	}

	// 2. Lock original POS transaction lines in ascending ID order to prevent concurrent return deadlocks
	sort.Slice(originalLineIDs, func(i, j int) bool { return originalLineIDs[i] < originalLineIDs[j] })
	for _, origID := range originalLineIDs {
		origLine, origErr := activeRepo.LockPosTransactionLineForUpdate(ctx, origID)
		if origErr != nil {
			return utils.NewResponse(utils.CodeError, fmt.Sprintf("original transaction line %d not found: %v", origID, origErr), nil)
		}

		prevReturned, prevErr := activeRepo.GetReturnedQuantityByLineID(ctx, pgtype.Int4{Int32: origID, Valid: true})
		if prevErr != nil {
			return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to get returned quantity for original line %d: %v", origID, prevErr), nil)
		}

		origRat := numericToRat(origLine.Quantity)
		prevRat := numericToRat(prevReturned)
		reqTotalRat := requestedPerOrigLine[origID]
		totalAfterReturn := new(big.Rat).Add(prevRat, reqTotalRat)

		if totalAfterReturn.Cmp(origRat) > 0 {
			availRat := new(big.Rat).Sub(origRat, prevRat)
			reqFloat, _ := reqTotalRat.Float64()
			availFloat, _ := availRat.Float64()
			origFloat, _ := origRat.Float64()
			prevFloat, _ := prevRat.Float64()
			return utils.NewResponse(utils.CodeError, fmt.Sprintf("return quantity %.3f exceeds available returnable quantity %.3f (original: %.3f, previously returned: %.3f)", reqFloat, availFloat, origFloat, prevFloat), nil)
		}
	}

	// 3. Generate collision-safe return number (RET-YYYYMMDDHHMMSS-<entropy>)
	returnNumber, err := GenerateReturnNumber(time.Now())
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to generate return number: "+err.Error(), nil)
	}

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

	salesReturn, err := activeRepo.CreateSalesReturn(ctx, headerParams)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create return header: "+err.Error(), nil)
	}

	// 4. Create Return Lines and apply inventory adjustments inside transaction
	createdReturnLines := make([]repository.SalesReturnLine, 0, len(preparedLines))
	for _, pl := range preparedLines {
		i := pl.index
		l := pl.input
		qty := pl.qty
		uPrice := pl.uPrice
		refAmount := pl.refAmount

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

		createdLine, err := activeRepo.CreateSalesReturnLine(ctx, lineParams)
		if err != nil {
			return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to create return line %d: %s", i+1, err.Error()), nil)
		}
		createdReturnLines = append(createdReturnLines, createdLine)

		// F2.3: Restock inventory when ReturnToStock is true and condition is good or empty
		isGoodCondition := l.Condition == "" || strings.ToLower(l.Condition) == "good"
		if l.ReturnToStock && isGoodCondition {
			_, adjErr := activeRepo.AdjustInventoryStockByProductAndStore(ctx, repository.AdjustInventoryStockByProductAndStoreParams{
				ProductID:         l.ProductID,
				ProductVariantID:  lineParams.ProductVariantID,
				StoreID:           in.StoreID,
				QuantityOnHand:    qty,
				QuantityAvailable: qty,
			})
			if adjErr != nil {
				return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to adjust inventory stock on return: %v", adjErr), nil)
			}

			// Determine unit cost and total cost for stock movement audit
			unitCost := uPrice
			if l.CostPrice != "" {
				if parsedCost, err := activeRepo.ParseNumeric(ctx, l.CostPrice); err == nil && parsedCost.Valid {
					unitCost = parsedCost
				}
			} else {
				if prod, pErr := activeRepo.GetProduct(ctx, l.ProductID); pErr == nil && prod.CostPrice.Valid {
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
				if parsedTot, err := activeRepo.ParseNumeric(ctx, fmt.Sprintf("%.2f", costVal*qtyVal)); err == nil && parsedTot.Valid {
					totalCost = parsedTot
				}
			}

			_, movErr := activeRepo.CreateStockMovement(ctx, repository.CreateStockMovementParams{
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
				return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to create stock movement on return: %v", movErr), nil)
			}
		}
	}

	// 5. Update original POS transaction status if referenced
	if in.OriginalTransactionID != nil {
		newStatus := "partially_refunded"
		origTxn, err := activeRepo.GetPosTransaction(ctx, *in.OriginalTransactionID)
		if err == nil && origTxn.TotalAmount.Valid && totalRefund.Valid {
			if origTxn.TotalAmount.Int != nil && totalRefund.Int != nil && origTxn.TotalAmount.Int.Cmp(totalRefund.Int) <= 0 {
				newStatus = "refunded"
			}
		}
		err = activeRepo.UpdatePOSTransactionStatus(ctx, repository.UpdatePOSTransactionStatusParams{
			ID:     *in.OriginalTransactionID,
			Status: pgtype.Text{String: newStatus, Valid: true},
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to update pos_transaction status: %v", err), nil)
		}
	}

	// 6. Adjust drawer expected_balance for cash/drawer refunds
	if in.SessionID != nil && totalRefund.Valid && (in.RefundMethod == "cash" || in.RefundMethod == "") {
		negRefund := totalRefund
		if negRefund.Int != nil {
			negRefund = pgtype.Numeric{Int: new(big.Int).Neg(negRefund.Int), Exp: negRefund.Exp, Valid: true}
		}

		err = activeRepo.UpdateSessionExpectedBalance(ctx, repository.UpdateSessionExpectedBalanceParams{
			ID:              *in.SessionID,
			ExpectedBalance: negRefund,
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to update session balance for refund: %v", err), nil)
		}
	}

	// 7. Commit database transaction
	if tx != nil {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return utils.NewResponse(utils.CodeError, "failed to commit sales return transaction: "+commitErr.Error(), nil)
		}
		tx = nil
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
