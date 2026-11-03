package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Dispatcher struct {
	pools   func(context.Context) ([]*pgxpool.Pool, error)
	pool    *pgxpool.Pool
	notifUC *usecase.NotificationUseCase
	wsHub   *WebSocketHub
}

// NewDispatcher polls every active tenant database. POS receipts are written
// to the tenant that received the sale, not the master database.
func NewDispatcher(pools func(context.Context) ([]*pgxpool.Pool, error), wsHub *WebSocketHub) *Dispatcher {
	return &Dispatcher{pools: pools, wsHub: wsHub}
}

func (d *Dispatcher) Start(ctx context.Context) {
	log.Println("[NotificationDispatcher] Background outbox dispatcher started.")
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	d.runOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			log.Println("[NotificationDispatcher] Dispatcher stopping on context cancellation.")
			return
		case <-ticker.C:
			d.runOnce(ctx)
		}
	}
}

func (d *Dispatcher) runOnce(ctx context.Context) {
	if d.pools == nil {
		return
	}
	pools, err := d.pools(ctx)
	if err != nil {
		log.Printf("[NotificationDispatcher] List tenant pools: %v", err)
		return
	}
	for _, pool := range pools {
		if pool == nil {
			continue
		}
		d.pool = pool
		d.notifUC = usecase.NewNotificationUseCase(repository.New(pool))
		d.processPendingEvents(ctx)
	}
}

func (d *Dispatcher) processPendingEvents(ctx context.Context) {
	events, err := d.notifUC.FetchPendingEvents(ctx, 25)
	if err != nil {
		return
	}

	for _, event := range events {
		d.processSingleEvent(ctx, event)
	}
}

func (d *Dispatcher) processSingleEvent(ctx context.Context, event repository.NotificationEvent) {
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Printf("[NotificationDispatcher] Error parsing payload for event %s: %v", event.ID, err)
		_, _ = d.notifUC.MarkEventFailed(ctx, event.ID)
		return
	}

	if event.EventTypeID == "pos.receipt" {
		if d.dispatchPosReceipt(ctx, event, payload) {
			rules, ruleErr := d.notifUC.ListRulesForEvent(ctx, event.OrganizationID, event.EventTypeID)
			if ruleErr == nil {
				for _, rule := range rules {
					d.dispatchByRule(ctx, event, rule, payload)
				}
			}
			_, _ = d.notifUC.MarkEventCompleted(ctx, event.ID)
		}
		return
	}

	rules, err := d.notifUC.ListRulesForEvent(ctx, event.OrganizationID, event.EventTypeID)
	if err != nil || len(rules) == 0 {
		// Default routing if no custom rules configured
		d.dispatchDefault(ctx, event, payload)
		_, _ = d.notifUC.MarkEventCompleted(ctx, event.ID)
		return
	}

	// Dispatch according to matched rules
	for _, rule := range rules {
		d.dispatchByRule(ctx, event, rule, payload)
	}

	_, _ = d.notifUC.MarkEventCompleted(ctx, event.ID)
}

func (d *Dispatcher) dispatchDefault(ctx context.Context, event repository.NotificationEvent, payload map[string]interface{}) {
	orgID := event.OrganizationID

	switch event.EventTypeID {
	case "po.submitted", "po.approved", "po.rejected":
		poNum := fmt.Sprintf("%v", payload["po_number"])
		status := fmt.Sprintf("%v", payload["status"])
		title := fmt.Sprintf("Purchase Order %s: %s", status, poNum)
		msg := fmt.Sprintf("Purchase order %s status updated to %s.", poNum, status)

		// Broadcast to staff in organization via WebSocket
		d.wsHub.BroadcastToOrg(orgID, map[string]interface{}{
			"type":       event.EventTypeID,
			"event_id":   event.ID,
			"title":      title,
			"message":    msg,
			"created_at": time.Now(),
		})

	case "grn.posted":
		grnNum := fmt.Sprintf("%v", payload["grn_number"])
		title := "Goods Receipt Note Posted"
		msg := fmt.Sprintf("GRN %s has been successfully posted to inventory.", grnNum)

		d.wsHub.BroadcastToOrg(orgID, map[string]interface{}{
			"type":       event.EventTypeID,
			"event_id":   event.ID,
			"title":      title,
			"message":    msg,
			"created_at": time.Now(),
		})

	case "inventory.low_stock":
		prodName := fmt.Sprintf("%v", payload["product_name"])
		storeName := fmt.Sprintf("%v", payload["store_name"])
		avail := fmt.Sprintf("%v", payload["quantity_available"])
		reorder := fmt.Sprintf("%v", payload["reorder_level"])
		title := fmt.Sprintf("Low Stock Alert: %s", prodName)
		msg := fmt.Sprintf("Item %s at store %s has only %s units remaining (reorder level: %s).", prodName, storeName, avail, reorder)

		d.wsHub.BroadcastToOrg(orgID, map[string]interface{}{
			"type":       event.EventTypeID,
			"event_id":   event.ID,
			"title":      title,
			"message":    msg,
			"created_at": time.Now(),
		})

	case "customer.created":
		custName, _ := payload["customer_name"].(string)
		custPhone, _ := payload["customer_phone"].(string)
		if custPhone != "" {
			d.sendWhatsAppMessage(ctx, orgID, event, custPhone, customerWelcomeMessage(custName))
		}
	}
}

func (d *Dispatcher) dispatchByRule(ctx context.Context, event repository.NotificationEvent, rule repository.NotificationRule, payload map[string]interface{}) {
	orgID := event.OrganizationID

	var channelList []string
	_ = json.Unmarshal(rule.Channels, &channelList)

	for _, channel := range channelList {
		if event.EventTypeID == "pos.receipt" && usecase.CanonicalChannelType(channel) == "whatsapp_cloud" {
			continue
		}
		switch channel {
		case "websocket":
			if rule.TargetRoleID.Valid {
				users, err := d.notifUC.GetInternalRecipientsByRole(ctx, orgID, rule.TargetRoleID.Int32)
				if err == nil {
					for _, u := range users {
						title := fmt.Sprintf("Notification: %s", event.EventTypeID)
						msg := fmt.Sprintf("Event %s triggered for %s #%s", event.EventTypeID, event.EntityType, event.EntityID)
						_, _ = d.notifUC.CreateInAppNotification(ctx, orgID, u.ID, &event.ID, title, msg, nil)
						d.wsHub.SendToUser(orgID, u.ID, map[string]interface{}{
							"type":       event.EventTypeID,
							"event_id":   event.ID,
							"title":      title,
							"message":    msg,
							"created_at": time.Now(),
						})
					}
				}
			} else {
				d.wsHub.BroadcastToOrg(orgID, map[string]interface{}{
					"type":     event.EventTypeID,
					"event_id": event.ID,
					"payload":  payload,
				})
			}

		case "whatsapp_cloud":
			recipientPhone, _ := payload["customer_phone"].(string)
			if recipientPhone == "" {
				recipientPhone, _ = payload["partner_phone"].(string)
			}
			if recipientPhone != "" {
				msgText := fmt.Sprintf("Update regarding %s %s: Notification event %s.", event.EntityType, event.EntityID, event.EventTypeID)
				if event.EventTypeID == "customer.created" {
					custName, _ := payload["customer_name"].(string)
					msgText = customerWelcomeMessage(custName)
				}
				d.sendWhatsAppMessage(ctx, orgID, event, recipientPhone, msgText)
			}

		case "email_ses":
			recipientEmail, _ := payload["customer_email"].(string)
			if recipientEmail == "" {
				recipientEmail, _ = payload["partner_email"].(string)
			}
			if recipientEmail != "" {
				d.sendEmailMessage(ctx, orgID, event, recipientEmail,
					fmt.Sprintf("Nembus Alert: %s", event.EventTypeID),
					fmt.Sprintf("<p>Notification for %s <b>#%s</b> has been processed.</p>", event.EntityType, event.EntityID))
			}
		}
	}
}

func (d *Dispatcher) sendWhatsAppMessage(ctx context.Context, orgID int32, event repository.NotificationEvent, phone, text string) {
	ch, err := d.notifUC.GetChannel(ctx, orgID, "whatsapp_cloud")
	if err != nil || (ch.IsActive.Valid && !ch.IsActive.Bool) {
		log.Printf("[NotificationDispatcher] WhatsApp channel inactive for org %d", orgID)
		return
	}

	var waCfg WhatsAppConfig
	if err := json.Unmarshal(ch.Config, &waCfg); err != nil {
		log.Printf("[NotificationDispatcher] Invalid WhatsApp config: %v", err)
		return
	}

	client := NewWhatsAppClient(waCfg)
	resp, sendErr := client.SendTextMessage(ctx, phone, text)

	status := "sent"
	var errMsg *string
	var providerID *string

	if sendErr != nil {
		status = "failed"
		s := sendErr.Error()
		errMsg = &s
	} else if resp != nil && len(resp.Messages) > 0 {
		providerID = &resp.Messages[0].ID
	}

	_, _ = d.notifUC.RecordDelivery(ctx, usecase.RecordDeliveryInput{
		EventID:           event.ID,
		ChannelType:       "whatsapp_cloud",
		Recipient:         phone,
		Status:            status,
		ProviderMessageID: providerID,
		ErrorMessage:      errMsg,
	})
}

func customerWelcomeMessage(name string) string {
	clean := strings.TrimSpace(name)
	if clean == "" {
		clean = "Valued Customer"
	}
	return fmt.Sprintf("Welcome to Nembus, %s! Thank you for joining us. We are delighted to serve you.", clean)
}

func (d *Dispatcher) sendEmailMessage(ctx context.Context, orgID int32, event repository.NotificationEvent, email, subject, htmlBody string) {
	ch, err := d.notifUC.GetChannel(ctx, orgID, "email_ses")
	if err != nil || (ch.IsActive.Valid && !ch.IsActive.Bool) {
		log.Printf("[NotificationDispatcher] AWS SES channel inactive for org %d", orgID)
		return
	}

	var sesCfg SESConfig
	if err := json.Unmarshal(ch.Config, &sesCfg); err != nil {
		log.Printf("[NotificationDispatcher] Invalid SES config: %v", err)
		return
	}

	client := NewSESClient(sesCfg)
	msgID, sendErr := client.SendEmail(ctx, email, subject, htmlBody)

	status := "sent"
	var errMsg *string
	var providerID *string

	if sendErr != nil {
		status = "failed"
		s := sendErr.Error()
		errMsg = &s
	} else {
		providerID = &msgID
	}

	_, _ = d.notifUC.RecordDelivery(ctx, usecase.RecordDeliveryInput{
		EventID:           event.ID,
		ChannelType:       "email_ses",
		Recipient:         email,
		Status:            status,
		ProviderMessageID: providerID,
		ErrorMessage:      errMsg,
	})
}
