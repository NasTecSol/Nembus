package usecase

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const ZatcaSalesOrderEntity = "sales_order_v2"

// OriginalIssuedInvoiceNumber keeps return references aligned with whichever
// document was actually issued: old POS invoices or new order invoices.
func (uc *ZatcaUseCase) OriginalIssuedInvoiceNumber(ctx context.Context, txn *repository.PosTransaction) (string, error) {
	entry, err := uc.IssuedTransactionDocument(ctx, txn)
	if errors.Is(err, pgx.ErrNoRows) && !txn.SalesOrderID.Valid {
		return txn.TransactionNumber, nil
	}
	if err != nil {
		return "", fmt.Errorf("original signed invoice not found: %w", err)
	}
	var snapshot struct {
		ID string `xml:"ID"`
	}
	if xml.Unmarshal([]byte(entry.SignedXml.String), &snapshot) != nil || snapshot.ID == "" {
		return "", fmt.Errorf("original invoice XML has no invoice number")
	}
	return snapshot.ID, nil
}

// IssuedTransactionDocument resolves the immutable source of a return.
func (uc *ZatcaUseCase) IssuedTransactionDocument(ctx context.Context, txn *repository.PosTransaction) (repository.ZatcaDocumentChain, error) {
	entry, err := uc.repo.GetChainEntryByEntity(ctx, repository.GetChainEntryByEntityParams{EntityType: "pos_transaction", EntityID: strconv.Itoa(int(txn.ID))})
	if errors.Is(err, pgx.ErrNoRows) && txn.SalesOrderID.Valid {
		entry, err = uc.repo.GetChainEntryByEntity(ctx, repository.GetChainEntryByEntityParams{EntityType: ZatcaSalesOrderEntity, EntityID: uuid.UUID(txn.SalesOrderID.Bytes).String()})
	}
	return entry, err
}

// EnsureSalesOrderDocument reads persisted orders and lines. Presence and valid
// totals determine readiness; no additional order/payment status is required.
// Existing documents are returned unchanged, including historical POS invoices.
func (uc *ZatcaUseCase) EnsureSalesOrderDocument(ctx context.Context, id uuid.UUID) (*repository.ZatcaDocumentChain, error) {
	if uc.cfg == nil {
		return nil, fmt.Errorf("ZATCA configuration unavailable")
	}
	if uc.repo == nil {
		return nil, fmt.Errorf("repository not set")
	}
	var result *repository.ZatcaDocumentChain
	err := uc.repo.InZatcaTransaction(ctx, func(repo *repository.Queries) error {
		if err := repo.LockZatcaOrder(ctx, id); err != nil {
			return err
		}
		order, err := repo.GetSalesOrderV2(ctx, id)
		if err != nil {
			return fmt.Errorf("sales order not found: %w", err)
		}
		org, err := repo.GetOrganization(ctx, order.OrganizationID)
		if err != nil {
			return err
		}
		enabled, err := OrganizationZatcaEnabled(&org)
		if err != nil {
			return err
		}
		if !enabled {
			return ErrZatcaDisabled
		}
		existing, err := repo.GetChainEntryByEntity(ctx, repository.GetChainEntryByEntityParams{EntityType: ZatcaSalesOrderEntity, EntityID: id.String()})
		if err == nil {
			result = &existing
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		oldTransaction, err := repo.GetLegacyZatcaOrderTransaction(ctx, id)
		if err == nil {
			old, err := repo.GetChainEntryByEntity(ctx, repository.GetChainEntryByEntityParams{EntityType: "pos_transaction", EntityID: strconv.Itoa(int(oldTransaction))})
			if err != nil {
				return err
			}
			result = &old
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		lines, err := repo.ListSalesOrderLinesV2(ctx, id)
		if err != nil {
			return err
		}
		if !order.StoreID.Valid {
			return fmt.Errorf("sales order requires store_id for ZATCA")
		}
		store, err := repo.GetStore(ctx, order.StoreID.Int32)
		if err != nil {
			return err
		}
		if store.OrganizationID != order.OrganizationID {
			return fmt.Errorf("order store belongs to a different organization")
		}
		selector := NewZatcaUseCase(uc.cfg)
		selector.SetRepository(repo)
		device, err := selector.SelectZatcaDevice(ctx, order.OrganizationID, order.PosTerminalID, order.Metadata)
		if err != nil {
			return err
		}

		if err := repo.LockZatcaDevice(ctx, device.ID); err != nil {
			return err
		}
		input, err := salesOrderInvoiceInput(&order, lines, &org, &store, uc.cfg.OrgVATID)
		if err != nil {
			return err
		}
		txUC := NewZatcaUseCase(uc.cfg)
		txUC.SetRepository(repo)
		input.ICV, input.PIH, err = txUC.GetNextICVAndPIH(ctx, device.ID)
		if err != nil {
			return err
		}
		input.ZatcaUUID = uuid.NewString()
		key, err := DecodePrivateKeyPEM(device.PrivateKeyPem)
		if err != nil {
			return err
		}
		signed, err := signZatcaDocument(input, key, &device)
		if err != nil {
			return err
		}
		entry, err := repo.InsertChainEntry(ctx, repository.InsertChainEntryParams{
			EntityType: ZatcaSalesOrderEntity, EntityID: id.String(), DeviceConfigID: device.ID, OrganizationID: order.OrganizationID,
			ZatcaUuid: uuid.MustParse(signed.ZatcaUUID), Icv: signed.ICV, Pih: signed.PIH, XmlHash: signed.XMLHash,
			ZatcaStatus:  repository.NullZatcaDocStatus{ZatcaDocStatus: repository.ZatcaDocStatusPending, Valid: true},
			QrCodeBase64: pgtype.Text{String: signed.QRCode, Valid: true}, SignedXml: pgtype.Text{String: string(signed.RawXML), Valid: true},
		})
		if err != nil {
			return err
		}
		result = &entry
		return nil
	})
	return result, err
}

func salesOrderInvoiceInput(order *repository.SalesOrdersV2, lines []repository.ListSalesOrderLinesV2Row, org *repository.Organization, store *repository.Store, defaultVAT string) (InvoiceXMLInput, error) {
	input := InvoiceXMLInput{InvoiceNumber: order.OrderNumber, IssueDate: time.Now().UTC(), InvoiceType: ZatcaTypeStandard, SubType: ZatcaSubTypeB2CSimplified, CurrencyCode: "SAR", DiscountAmount: "0.00", TaxCategoryID: "S", TaxPercent: "15.00"}
	if input.InvoiceNumber == "" {
		return input, fmt.Errorf("order_number is required")
	}
	if len(lines) == 0 {
		return input, fmt.Errorf("sales order requires at least one persisted line")
	}
	if org.CurrencyCode.Valid && org.CurrencyCode.String != "" && org.CurrencyCode.String != "SAR" {
		return input, fmt.Errorf("order ZATCA mapping currently supports SAR only")
	}
	for _, charge := range []struct {
		name  string
		value pgtype.Numeric
	}{{"shipping_amount", order.ShippingAmount}, {"adjustment_amount", order.AdjustmentAmount}, {"promotional_credits", order.PromotionalCredits}} {
		value, err := decimalValue(numericToString(charge.value))
		if err != nil {
			return input, err
		}
		if value.Sign() != 0 {
			return input, fmt.Errorf("nonzero %s requires an explicit ZATCA charge/allowance mapping", charge.name)
		}
	}
	if err := applySeller(&input, org, store, defaultVAT); err != nil {
		return input, err
	}
	netSum := new(big.Rat)
	for _, line := range lines {
		if line.SalesOrderID != order.ID || line.OrganizationID != order.OrganizationID {
			return input, fmt.Errorf("order line belongs to a different order or organization")
		}
		if line.TaxRate.Valid {
			rate, err := decimalValue(numericToString(line.TaxRate))
			if err != nil || rate.Cmp(big.NewRat(15, 1)) != 0 {
				return input, fmt.Errorf("order line %d: only standard 15 percent VAT is currently supported", line.LineNumber)
			}
		}
		qty, err := decimalValue(numericToString(line.QuantityOrdered))
		if err != nil {
			return input, err
		}
		price, err := decimalValue(numericToString(line.UnitPrice))
		if err != nil {
			return input, err
		}
		discount, err := decimalValue(numericToString(line.DiscountAmount))
		if err != nil {
			return input, err
		}
		net := new(big.Rat).Sub(new(big.Rat).Mul(qty, price), discount)
		gross, _ := addMoney(net.FloatString(2), numericToString(line.TaxAmount))
		stored, err := decimalValue(numericToString(line.LineTotal))
		if err != nil || stored.FloatString(2) != gross {
			return input, fmt.Errorf("order line %d line_total must equal quantity × price − discount + VAT", line.LineNumber)
		}
		input.Lines = append(input.Lines, InvoiceLineInput{LineNumber: int(line.LineNumber), ProductName: line.ProductName, Quantity: numericToString(line.QuantityOrdered), UnitCode: "PCE", UnitPrice: numericToString(line.UnitPrice), DiscountAmt: numericToString(line.DiscountAmount), LineTotal: net.FloatString(2), TaxAmount: numericToString(line.TaxAmount), TaxPercent: "15.00", TaxCategory: "S"})
		netSum.Add(netSum, net)
	}
	// Header subtotal/discount must reconcile to the final stored total. Allocate
	// any additional order discount across lines so VAT and QR totals agree.
	headerNet, err := subtractMoney(numericToString(order.Subtotal), numericToString(order.DiscountAmount))
	if err != nil {
		return input, err
	}
	targetNet, _ := decimalValue(headerNet)
	allowance := new(big.Rat).Sub(netSum, targetNet)
	if targetNet.Sign() < 0 || allowance.Sign() < 0 {
		return input, fmt.Errorf("order subtotal/discount does not reconcile with line amounts")
	}
	headerDiscount, _ := decimalValue(numericToString(order.DiscountAmount))
	if allowance.Cmp(headerDiscount) > 0 {
		return input, fmt.Errorf("unexplained difference between order and line totals")
	}
	remaining := new(big.Rat).Set(allowance)
	for i := range input.Lines {
		line := &input.Lines[i]
		net, _ := decimalValue(line.LineTotal)
		share := new(big.Rat)
		if allowance.Sign() > 0 {
			if i == len(input.Lines)-1 {
				share.Set(remaining)
			} else {
				share.Quo(new(big.Rat).Mul(allowance, net), netSum)
				share, _ = decimalValue(share.FloatString(2))
			}
			remaining.Sub(remaining, share)
			if share.Sign() < 0 || share.Cmp(net) > 0 {
				return input, fmt.Errorf("order discount cannot be allocated safely across lines")
			}
			line.DiscountAmt, _ = addMoney(line.DiscountAmt, share.FloatString(2))
			net.Sub(net, share)
			line.LineTotal = net.FloatString(2)
			line.TaxAmount = new(big.Rat).Mul(net, big.NewRat(15, 100)).FloatString(2)
		}
	}
	input.Subtotal = headerNet
	input.TaxAmount = numericToString(order.TaxAmount)
	input.TotalAmount = numericToString(order.TotalAmount)
	return input, validateInvoiceAmounts(input)
}
