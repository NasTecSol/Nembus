package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skip2/go-qrcode"
)

const maxReceiptNotReady = 8

type receiptLine struct {
	Name  string
	Qty   string
	Total string
}

type saleReceipt struct {
	StoreName string
	Number    string
	Date      string
	Tax       string
	Total     string
	Currency  string
	Phone     string
	OptIn     bool
	QR        string
	Lines     []receiptLine
}

// dispatchPosReceipt returns true when the event should be marked dispatched.
// A false result leaves it pending so a later sync can supply lines or the QR.
func (d *Dispatcher) dispatchPosReceipt(ctx context.Context, event repository.NotificationEvent, payload map[string]interface{}) bool {
	if d.pool == nil || d.notifUC == nil {
		return true
	}
	receipt, err := loadSaleReceipt(ctx, d.pool, event, payload)
	if errors.Is(err, errReceiptVoided) {
		d.recordDelivery(ctx, event, "whatsapp_cloud", "", "skipped", nil, err.Error())
		return true
	}
	if err != nil {
		if receiptRetries(event) < maxReceiptNotReady {
			d.deferReceipt(ctx, event.ID)
			return false
		}
		d.recordDelivery(ctx, event, "whatsapp_cloud", "", "failed", nil, err.Error())
		return true
	}

	rules, err := d.notifUC.ListRulesForEvent(ctx, event.OrganizationID, event.EventTypeID)
	if err != nil {
		d.recordDelivery(ctx, event, "whatsapp_cloud", "", "failed", nil, err.Error())
		return true
	}
	if !whatsappPolicyAllows(rules, receipt.OptIn) {
		reason := "no active WhatsApp rule for pos.receipt"
		if len(whatsAppRules(rules)) > 0 {
			reason = "customer has not opted in to WhatsApp"
		}
		d.recordDelivery(ctx, event, "whatsapp_cloud", receipt.Phone, "skipped", nil, reason)
		return true
	}
	if strings.TrimSpace(receipt.Phone) == "" {
		d.recordDelivery(ctx, event, "whatsapp_cloud", "", "skipped", nil, "customer has no WhatsApp contact")
		return true
	}

	ch, err := d.notifUC.GetChannel(ctx, event.OrganizationID, "whatsapp_cloud")
	if err != nil || (ch.IsActive.Valid && !ch.IsActive.Bool) {
		d.recordDelivery(ctx, event, "whatsapp_cloud", receipt.Phone, "skipped", nil, "WhatsApp channel is inactive")
		return true
	}

	tmpl := d.receiptTemplate(ctx, event.OrganizationID)
	if tmpl == nil || !tmpl.MetaTemplateName.Valid || strings.TrimSpace(tmpl.MetaTemplateName.String) == "" {
		d.recordDelivery(ctx, event, "whatsapp_cloud", receipt.Phone, "failed", nil, "approve a WhatsApp utility template (meta_template_name) for pos.receipt")
		return true
	}

	missingQR := strings.TrimSpace(receipt.QR) == ""
	templateName := strings.ToLower(strings.TrimSpace(tmpl.MetaTemplateName.String))
	isTextOnlyTemplate := strings.Contains(templateName, "receipt_txt") || strings.Contains(templateName, "_txt")

	if !isTextOnlyTemplate && (missingQR || len(receipt.Lines) == 0) && receiptRetries(event) < maxReceiptNotReady {
		d.deferReceipt(ctx, event.ID)
		return false
	}

	var waCfg WhatsAppConfig
	if err := json.Unmarshal(ch.Config, &waCfg); err != nil {
		d.recordDelivery(ctx, event, "whatsapp_cloud", receipt.Phone, "failed", nil, "invalid WhatsApp config")
		return true
	}
	client := NewWhatsAppClient(waCfg)
	headerID := ""
	note := ""
	if isTextOnlyTemplate {
		headerID = ""
	} else if !missingQR {
		png, renderErr := RenderZATCAQR(receipt.QR)
		if renderErr != nil {
			note = "ZATCA QR could not be rendered: " + renderErr.Error()
		} else if id, upErr := client.UploadMedia(ctx, "zatca-qr.png", "image/png", png); upErr != nil {
			note = "ZATCA QR upload failed: " + upErr.Error()
		} else {
			headerID = id
		}
	} else {
		note = "ZATCA QR was not available; template sent without the image header"
	}

	language := tmpl.Locale
	if language == "" {
		language = "en"
	}
	params := receiptParametersForTemplate(tmpl, receipt)
	resp, sendErr := client.SendTemplateMessageWithHeader(ctx, receipt.Phone, strings.TrimSpace(tmpl.MetaTemplateName.String), language, params, headerID)
	status := "sent"
	var providerID *string
	if sendErr != nil {
		status = "failed"
		note = joinNotes(note, sendErr.Error())
	} else if resp != nil && len(resp.Messages) > 0 {
		providerID = &resp.Messages[0].ID
	}
	d.recordDelivery(ctx, event, "whatsapp_cloud", receipt.Phone, status, providerID, note)
	return true
}

func (d *Dispatcher) receiptTemplate(ctx context.Context, orgID int32) *repository.NotificationTemplate {
	for _, channel := range []string{"whatsapp_cloud", "whatsapp"} {
		for _, locale := range []string{"en_US", "en", "ar"} {
			tmpl, err := d.notifUC.GetTemplate(ctx, orgID, "pos.receipt", channel, locale)
			if err == nil && tmpl != nil && tmpl.MetaTemplateName.Valid && strings.TrimSpace(tmpl.MetaTemplateName.String) != "" {
				return tmpl
			}
		}
	}
	return nil
}

func (d *Dispatcher) deferReceipt(ctx context.Context, eventID uuid.UUID) {
	if d.pool == nil {
		return
	}
	_, _ = d.pool.Exec(ctx, `
		UPDATE notification_events
		SET retry_count = COALESCE(retry_count, 0) + 1
		WHERE id = $1 AND status = 'pending'
	`, eventID)
}

func (d *Dispatcher) recordDelivery(ctx context.Context, event repository.NotificationEvent, channel, recipient, status string, providerID *string, note string) {
	if recipient == "" {
		recipient = "-"
	}
	var errMsg *string
	if strings.TrimSpace(note) != "" {
		errMsg = &note
	}
	_, _ = d.notifUC.RecordDelivery(ctx, usecase.RecordDeliveryInput{
		EventID:           event.ID,
		ChannelType:       channel,
		Recipient:         recipient,
		Status:            status,
		ProviderMessageID: providerID,
		ErrorMessage:      errMsg,
	})
}

func receiptRetries(event repository.NotificationEvent) int32 {
	if event.RetryCount.Valid {
		return event.RetryCount.Int32
	}
	return 0
}

func whatsAppRules(rules []repository.NotificationRule) []repository.NotificationRule {
	var matched []repository.NotificationRule
	for _, rule := range rules {
		if rule.IsActive.Valid && !rule.IsActive.Bool {
			continue
		}
		if channelListIncludesWhatsApp(rule.Channels) {
			matched = append(matched, rule)
		}
	}
	return matched
}

func whatsappPolicyAllows(rules []repository.NotificationRule, optIn bool) bool {
	for _, rule := range whatsAppRules(rules) {
		if ruleRequiresWhatsAppOptIn(rule.Conditions) && !optIn {
			continue
		}
		return true
	}
	return false
}

func channelListIncludesWhatsApp(raw json.RawMessage) bool {
	var list []string
	if err := json.Unmarshal(raw, &list); err != nil {
		return false
	}
	for _, channel := range list {
		if usecase.CanonicalChannelType(channel) == "whatsapp_cloud" {
			return true
		}
	}
	return false
}

func ruleRequiresWhatsAppOptIn(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var cond map[string]interface{}
	if err := json.Unmarshal(raw, &cond); err != nil {
		return false
	}
	flag, ok := cond["require_whatsapp_opt_in"].(bool)
	return ok && flag
}

func resolveCustomerPhone(metadataPhone, orderPhone, customerPhone string) string {
	for _, phone := range []string{metadataPhone, orderPhone, customerPhone} {
		if strings.TrimSpace(phone) != "" {
			return strings.TrimSpace(phone)
		}
	}
	return ""
}

func receiptParameters(receipt saleReceipt) []string {
	items := summarizeItems(receipt.Lines)
	store := receipt.StoreName
	if strings.TrimSpace(store) == "" {
		store = "Store"
	}
	currency := receipt.Currency
	if currency == "" {
		currency = "SAR"
	}
	total := fmt.Sprintf("%s %s", moneyText(receipt.Total), currency)
	return []string{store, receipt.Number, receipt.Date, items, total}
}

func receiptParametersForTemplate(tmpl *repository.NotificationTemplate, receipt saleReceipt) []string {
	currency := receipt.Currency
	if currency == "" {
		currency = "SAR"
	}
	total := fmt.Sprintf("%s %s", moneyText(receipt.Total), currency)

	if tmpl != nil {
		name := strings.ToLower(strings.TrimSpace(tmpl.MetaTemplateName.String))
		body := tmpl.BodyTemplate
		// If template is nembus_receipt_txt or has 3 parameters: {{1}}=Receipt, {{2}}=Date, {{3}}=Total
		if strings.Contains(name, "receipt_txt") || (!strings.Contains(body, "{{4}}") && strings.Contains(body, "{{3}}")) {
			dateStr := strings.TrimSpace(receipt.Date)
			if dateStr == "" {
				dateStr = time.Now().Format("2006-01-02")
			}
			return []string{receipt.Number, dateStr, total}
		}
	}

	return receiptParameters(receipt)
}

func summarizeItems(lines []receiptLine) string {
	if len(lines) == 0 {
		return "Receipt items unavailable"
	}
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		name := strings.TrimSpace(line.Name)
		if name == "" {
			name = "Item"
		}
		parts = append(parts, fmt.Sprintf("%s x%s %s", name, strings.TrimSpace(line.Qty), moneyText(line.Total)))
	}
	text := strings.Join(parts, "; ")
	if len(text) > 900 {
		return text[:897] + "..."
	}
	return text
}

func moneyText(raw string) string {
	f, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		if strings.TrimSpace(raw) == "" {
			return "0.00"
		}
		return strings.TrimSpace(raw)
	}
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func joinNotes(left, right string) string {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == "" {
		return right
	}
	if right == "" {
		return left
	}
	return left + "; " + right
}

// RenderZATCAQR draws the TLV payload as a QR image. The stored value is the
// base64 TLV string, which is what a customer receipt QR encodes.
func RenderZATCAQR(payload string) ([]byte, error) {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return nil, errors.New("empty zatca qr")
	}
	return qrcode.Encode(payload, qrcode.Medium, 256)
}

var errReceiptVoided = errors.New("transaction voided")

func loadSaleReceipt(ctx context.Context, pool *pgxpool.Pool, event repository.NotificationEvent, payload map[string]interface{}) (saleReceipt, error) {
	txID := jsonNumber(payload["transaction_id"])
	if txID == 0 {
		txID, _ = strconv.ParseInt(event.EntityID, 10, 64)
	}
	if txID == 0 || pool == nil {
		return saleReceipt{}, errors.New("receipt transaction id missing")
	}

	var receipt saleReceipt
	var status, orderPhone, customerPhone, salesOrderID, metaPhone, metaQR, metaCurrency string
	var meta []byte
	var voidedAt *string
	err := pool.QueryRow(ctx, `
		SELECT
			pt.transaction_number,
			COALESCE(s.name, ''),
			COALESCE(pt.status, 'completed'),
			pt.voided_at::text,
			COALESCE(pt.tax_amount, 0)::text,
			COALESCE(pt.total_amount, 0)::text,
			COALESCE(pt.transaction_date::text, ''),
			COALESCE(pt.metadata, '{}'::jsonb),
			COALESCE(pt.sales_order_id::text, ''),
			COALESCE(so.customer_phone, ''),
			COALESCE(cu.phone, '')
		FROM pos_transactions pt
		JOIN stores s ON s.id = pt.store_id
		LEFT JOIN sales_orders_v2 so ON so.id = pt.sales_order_id
		LEFT JOIN customers cu ON cu.id = pt.customer_id
		WHERE pt.id = $1
	`, txID).Scan(
		&receipt.Number, &receipt.StoreName, &status, &voidedAt,
		&receipt.Tax, &receipt.Total, &receipt.Date, &meta, &salesOrderID,
		&orderPhone, &customerPhone,
	)
	if err != nil {
		return saleReceipt{}, err
	}
	if voidedAt != nil && strings.TrimSpace(*voidedAt) != "" {
		return saleReceipt{}, errReceiptVoided
	}
	if status != "" && !strings.EqualFold(status, "completed") {
		return saleReceipt{}, errReceiptVoided
	}
	var metaMap map[string]interface{}
	if json.Unmarshal(meta, &metaMap) == nil {
		metaPhone, _ = metaMap["customer_phone"].(string)
		metaQR, _ = metaMap["qr_code_base64"].(string)
		if currency, ok := metaMap["currency"].(string); ok {
			metaCurrency = currency
		} else if currency, ok := metaMap["currency_code"].(string); ok {
			metaCurrency = currency
		}
		if opted, ok := metaMap["whatsapp_opt_in"].(bool); ok {
			receipt.OptIn = opted
		}
	}
	receipt.Phone = resolveCustomerPhone(metaPhone, orderPhone, customerPhone)
	receipt.Currency = metaCurrency
	if receipt.Currency == "" {
		receipt.Currency = "SAR"
	}
	receipt.QR = strings.TrimSpace(metaQR)
	if len(receipt.Date) >= 10 {
		receipt.Date = receipt.Date[:10]
	}
	if receipt.QR == "" {
		receipt.QR = lookupZATCAQR(ctx, pool, salesOrderID, strconv.FormatInt(txID, 10))
	}

	rows, err := pool.Query(ctx, `
		SELECT COALESCE(NULLIF(ptl.metadata->>'product_name', ''), p.name, 'Item'),
		       ptl.quantity::text,
		       COALESCE(ptl.line_total, ptl.subtotal, 0)::text
		FROM pos_transaction_lines ptl
		LEFT JOIN products p ON p.id = ptl.product_id
		WHERE ptl.transaction_id = $1
		ORDER BY COALESCE(ptl.line_number, 0), ptl.id
	`, txID)
	if err != nil {
		return receipt, err
	}
	defer rows.Close()
	for rows.Next() {
		var line receiptLine
		if err := rows.Scan(&line.Name, &line.Qty, &line.Total); err != nil {
			return receipt, err
		}
		receipt.Lines = append(receipt.Lines, line)
	}
	return receipt, rows.Err()
}

func lookupZATCAQR(ctx context.Context, pool *pgxpool.Pool, salesOrderID, transactionID string) string {
	var qr string
	err := pool.QueryRow(ctx, `
		SELECT qr_code_base64
		FROM zatca_document_chain
		WHERE COALESCE(qr_code_base64, '') <> ''
		  AND entity_id IN ($1, $2)
		ORDER BY id DESC
		LIMIT 1
	`, salesOrderID, transactionID).Scan(&qr)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return ""
		}
		return ""
	}
	return strings.TrimSpace(qr)
}

func jsonNumber(value interface{}) int64 {
	switch n := value.(type) {
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	case json.Number:
		parsed, _ := n.Int64()
		return parsed
	case string:
		parsed, _ := strconv.ParseInt(n, 10, 64)
		return parsed
	default:
		return 0
	}
}
