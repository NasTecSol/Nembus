package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type NotificationUseCase struct {
	repo *repository.Queries
}

func NewNotificationUseCase(repo *repository.Queries) *NotificationUseCase {
	return &NotificationUseCase{repo: repo}
}

func (uc *NotificationUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// -----------------------------------------------------
// DTO Inputs
// -----------------------------------------------------

type EnqueueNotificationEventInput struct {
	OrganizationID int32                  `json:"organization_id"`
	EventTypeID    string                 `json:"event_type_id"`
	EntityType     string                 `json:"entity_type"`
	EntityID       string                 `json:"entity_id"`
	IdempotencyKey string                 `json:"idempotency_key"`
	Payload        map[string]interface{} `json:"payload"`
}

type UpsertTemplateInput struct {
	OrganizationID   int32   `json:"organization_id"`
	EventTypeID      string  `json:"event_type_id"`
	ChannelType      string  `json:"channel_type"`
	Locale           string  `json:"locale"`
	SubjectTemplate  *string `json:"subject_template,omitempty"`
	BodyTemplate     string  `json:"body_template"`
	MetaTemplateName *string `json:"meta_template_name,omitempty"`
}

type CreateRuleInput struct {
	OrganizationID int32           `json:"organization_id"`
	EventTypeID    string          `json:"event_type_id"`
	AudienceType   string          `json:"audience_type"`
	TargetRoleID   *int32          `json:"target_role_id,omitempty"`
	Channels       json.RawMessage `json:"channels"`
	Conditions     json.RawMessage `json:"conditions,omitempty"`
	IsActive       bool            `json:"is_active"`
}

type UpdateRuleInput struct {
	ID             int32           `json:"id"`
	OrganizationID int32           `json:"organization_id"`
	AudienceType   string          `json:"audience_type"`
	TargetRoleID   *int32          `json:"target_role_id,omitempty"`
	Channels       json.RawMessage `json:"channels"`
	Conditions     json.RawMessage `json:"conditions,omitempty"`
	IsActive       bool            `json:"is_active"`
}

type RecordDeliveryInput struct {
	EventID           uuid.UUID `json:"event_id"`
	ChannelType       string    `json:"channel_type"`
	Recipient         string    `json:"recipient"`
	Status            string    `json:"status"` // 'sent', 'delivered', 'failed'
	ProviderMessageID *string   `json:"provider_message_id,omitempty"`
	ErrorMessage      *string   `json:"error_message,omitempty"`
}

// -----------------------------------------------------
// Outbox & Dispatcher Operations
// -----------------------------------------------------

func (uc *NotificationUseCase) EnqueueEvent(ctx context.Context, in EnqueueNotificationEventInput) (*repository.NotificationEvent, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}

	payloadBytes, err := json.Marshal(in.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	idempotencyKey := in.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey = fmt.Sprintf("%s:%s:%s", in.EventTypeID, in.EntityID, time.Now().UTC().Format(time.RFC3339Nano))
	}

	event, err := uc.repo.EnqueueNotificationEvent(ctx, repository.EnqueueNotificationEventParams{
		OrganizationID: in.OrganizationID,
		EventTypeID:    in.EventTypeID,
		EntityType:     in.EntityType,
		EntityID:       in.EntityID,
		IdempotencyKey: idempotencyKey,
		Payload:        payloadBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue event: %w", err)
	}
	return &event, nil
}

func (uc *NotificationUseCase) FetchPendingEvents(ctx context.Context, limit int32) ([]repository.NotificationEvent, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}
	if limit <= 0 {
		limit = 50
	}
	return uc.repo.GetPendingNotificationEventsForUpdate(ctx, limit)
}

func (uc *NotificationUseCase) MarkEventCompleted(ctx context.Context, eventID uuid.UUID) (*repository.NotificationEvent, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}
	ev, err := uc.repo.UpdateNotificationEventStatus(ctx, repository.UpdateNotificationEventStatusParams{
		ID:     eventID,
		Status: "dispatched",
	})
	if err != nil {
		return nil, err
	}
	return &ev, nil
}

func (uc *NotificationUseCase) MarkEventFailed(ctx context.Context, eventID uuid.UUID) (*repository.NotificationEvent, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}
	ev, err := uc.repo.MarkNotificationEventFailed(ctx, eventID)
	if err != nil {
		return nil, err
	}
	return &ev, nil
}

func (uc *NotificationUseCase) RecordDelivery(ctx context.Context, in RecordDeliveryInput) (*repository.NotificationDelivery, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}

	params := repository.CreateNotificationDeliveryParams{
		EventID:     in.EventID,
		ChannelType: in.ChannelType,
		Recipient:   in.Recipient,
		Status:      in.Status,
		SentAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	if in.ProviderMessageID != nil {
		params.ProviderMessageID = pgtype.Text{String: *in.ProviderMessageID, Valid: true}
	}
	if in.ErrorMessage != nil {
		params.ErrorMessage = pgtype.Text{String: *in.ErrorMessage, Valid: true}
	}

	del, err := uc.repo.CreateNotificationDelivery(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create delivery log: %w", err)
	}
	return &del, nil
}

// -----------------------------------------------------
// In-App User Notifications (Inbox & Bell)
// -----------------------------------------------------

func (uc *NotificationUseCase) CreateInAppNotification(ctx context.Context, orgID, userID int32, eventID *uuid.UUID, title, message string, actionURL *string) (*repository.UserNotification, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}

	params := repository.CreateUserNotificationParams{
		OrganizationID: orgID,
		UserID:         userID,
		Title:          title,
		Message:        message,
	}
	if eventID != nil {
		params.EventID = pgtype.UUID{Bytes: *eventID, Valid: true}
	}
	if actionURL != nil {
		params.ActionUrl = pgtype.Text{String: *actionURL, Valid: true}
	}

	notif, err := uc.repo.CreateUserNotification(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create user notification: %w", err)
	}
	return &notif, nil
}

func (uc *NotificationUseCase) ListUserNotifications(ctx context.Context, orgID, userID int32, isRead *bool, limit, offset int32) ([]repository.UserNotification, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}
	if limit <= 0 {
		limit = 20
	}

	params := repository.ListUserNotificationsParams{
		OrganizationID: orgID,
		UserID:         userID,
		LimitCount:     limit,
		OffsetCount:    offset,
	}
	if isRead != nil {
		params.IsRead = pgtype.Bool{Bool: *isRead, Valid: true}
	}

	return uc.repo.ListUserNotifications(ctx, params)
}

func (uc *NotificationUseCase) GetUnreadCount(ctx context.Context, orgID, userID int32) (int64, error) {
	if uc.repo == nil {
		return 0, errors.New("repository not configured")
	}
	return uc.repo.GetUnreadNotificationCount(ctx, repository.GetUnreadNotificationCountParams{
		OrganizationID: orgID,
		UserID:         userID,
	})
}

func (uc *NotificationUseCase) MarkAsRead(ctx context.Context, notifID uuid.UUID, userID int32) (*repository.UserNotification, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}
	notif, err := uc.repo.MarkUserNotificationAsRead(ctx, repository.MarkUserNotificationAsReadParams{
		ID:     notifID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &notif, nil
}

func (uc *NotificationUseCase) MarkAllAsRead(ctx context.Context, orgID, userID int32) error {
	if uc.repo == nil {
		return errors.New("repository not configured")
	}
	return uc.repo.MarkAllUserNotificationsAsRead(ctx, repository.MarkAllUserNotificationsAsReadParams{
		OrganizationID: orgID,
		UserID:         userID,
	})
}

// -----------------------------------------------------
// Channels Configuration (AWS SES, Meta WhatsApp)
// -----------------------------------------------------

func (uc *NotificationUseCase) GetChannel(ctx context.Context, orgID int32, channelType string) (*repository.NotificationChannel, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}
	ch, err := uc.repo.GetNotificationChannel(ctx, repository.GetNotificationChannelParams{
		OrganizationID: orgID,
		ChannelType:    channelType,
	})
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func (uc *NotificationUseCase) ListChannels(ctx context.Context, orgID int32) ([]repository.NotificationChannel, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}
	return uc.repo.ListNotificationChannels(ctx, orgID)
}

func (uc *NotificationUseCase) UpsertChannel(ctx context.Context, orgID int32, channelType string, isActive bool, config json.RawMessage) (*repository.NotificationChannel, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}
	if config == nil {
		config = json.RawMessage("{}")
	}
	ch, err := uc.repo.UpsertNotificationChannel(ctx, repository.UpsertNotificationChannelParams{
		OrganizationID: orgID,
		ChannelType:    channelType,
		IsActive:       pgtype.Bool{Bool: isActive, Valid: true},
		Config:         config,
	})
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

// -----------------------------------------------------
// Templates & Interpolation
// -----------------------------------------------------

func (uc *NotificationUseCase) GetTemplate(ctx context.Context, orgID int32, eventType, channelType, locale string) (*repository.NotificationTemplate, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}
	if locale == "" {
		locale = "en"
	}
	tmpl, err := uc.repo.GetNotificationTemplate(ctx, repository.GetNotificationTemplateParams{
		OrganizationID: orgID,
		EventTypeID:    eventType,
		ChannelType:    channelType,
		Locale:         locale,
	})
	if err != nil {
		return nil, err
	}
	return &tmpl, nil
}

func (uc *NotificationUseCase) ListTemplates(ctx context.Context, orgID int32, eventType *string) ([]repository.NotificationTemplate, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}
	params := repository.ListNotificationTemplatesParams{OrganizationID: orgID}
	if eventType != nil {
		params.EventTypeID = pgtype.Text{String: *eventType, Valid: true}
	}
	return uc.repo.ListNotificationTemplates(ctx, params)
}

func (uc *NotificationUseCase) UpsertTemplate(ctx context.Context, in UpsertTemplateInput) (*repository.NotificationTemplate, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}
	if in.Locale == "" {
		in.Locale = "en"
	}

	params := repository.UpsertNotificationTemplateParams{
		OrganizationID: in.OrganizationID,
		EventTypeID:    in.EventTypeID,
		ChannelType:    in.ChannelType,
		Locale:         in.Locale,
		BodyTemplate:   in.BodyTemplate,
	}
	if in.SubjectTemplate != nil {
		params.SubjectTemplate = pgtype.Text{String: *in.SubjectTemplate, Valid: true}
	}
	if in.MetaTemplateName != nil {
		params.MetaTemplateName = pgtype.Text{String: *in.MetaTemplateName, Valid: true}
	}

	tmpl, err := uc.repo.UpsertNotificationTemplate(ctx, params)
	if err != nil {
		return nil, err
	}
	return &tmpl, nil
}

func (uc *NotificationUseCase) DeleteTemplate(ctx context.Context, orgID, id int32) error {
	if uc.repo == nil {
		return errors.New("repository not configured")
	}
	_, err := uc.repo.DeleteNotificationTemplate(ctx, repository.DeleteNotificationTemplateParams{
		OrganizationID: orgID,
		ID:             id,
	})
	return err
}

var variableRegex = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)

// RenderTemplate interpolates variables from data map into subject & body templates (e.g. {{order_number}})
func (uc *NotificationUseCase) RenderTemplate(subjectTmpl, bodyTmpl string, data map[string]interface{}) (string, string) {
	replacer := func(match string) string {
		varName := strings.Trim(match, "{ }")
		if val, exists := data[varName]; exists {
			return fmt.Sprintf("%v", val)
		}
		return match
	}

	renderedSubject := variableRegex.ReplaceAllStringFunc(subjectTmpl, replacer)
	renderedBody := variableRegex.ReplaceAllStringFunc(bodyTmpl, replacer)

	return renderedSubject, renderedBody
}

// -----------------------------------------------------
// Routing Rules & Recipient Resolution
// -----------------------------------------------------

func (uc *NotificationUseCase) ListRules(ctx context.Context, orgID int32) ([]repository.NotificationRule, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}
	return uc.repo.ListAllNotificationRules(ctx, orgID)
}

func (uc *NotificationUseCase) ListRulesForEvent(ctx context.Context, orgID int32, eventType string) ([]repository.NotificationRule, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}
	return uc.repo.ListNotificationRulesByEvent(ctx, repository.ListNotificationRulesByEventParams{
		OrganizationID: orgID,
		EventTypeID:    eventType,
	})
}

func (uc *NotificationUseCase) CreateRule(ctx context.Context, in CreateRuleInput) (*repository.NotificationRule, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}
	if in.Channels == nil {
		in.Channels = json.RawMessage(`["websocket"]`)
	}
	if in.Conditions == nil {
		in.Conditions = json.RawMessage(`{}`)
	}

	params := repository.CreateNotificationRuleParams{
		OrganizationID: in.OrganizationID,
		EventTypeID:    in.EventTypeID,
		AudienceType:   in.AudienceType,
		Channels:       in.Channels,
		Conditions:     in.Conditions,
		IsActive:       pgtype.Bool{Bool: in.IsActive, Valid: true},
	}
	if in.TargetRoleID != nil {
		params.TargetRoleID = pgtype.Int4{Int32: *in.TargetRoleID, Valid: true}
	}

	rule, err := uc.repo.CreateNotificationRule(ctx, params)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (uc *NotificationUseCase) UpdateRule(ctx context.Context, in UpdateRuleInput) (*repository.NotificationRule, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}
	params := repository.UpdateNotificationRuleParams{
		ID:             in.ID,
		OrganizationID: in.OrganizationID,
		AudienceType:   in.AudienceType,
		Channels:       in.Channels,
		Conditions:     in.Conditions,
		IsActive:       pgtype.Bool{Bool: in.IsActive, Valid: true},
	}
	if in.TargetRoleID != nil {
		params.TargetRoleID = pgtype.Int4{Int32: *in.TargetRoleID, Valid: true}
	}

	rule, err := uc.repo.UpdateNotificationRule(ctx, params)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (uc *NotificationUseCase) DeleteRule(ctx context.Context, orgID, id int32) error {
	if uc.repo == nil {
		return errors.New("repository not configured")
	}
	_, err := uc.repo.DeleteNotificationRule(ctx, repository.DeleteNotificationRuleParams{
		ID:             id,
		OrganizationID: orgID,
	})
	return err
}

func (uc *NotificationUseCase) GetInternalRecipientsByRole(ctx context.Context, orgID, roleID int32) ([]repository.GetUsersByRoleRow, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not configured")
	}
	return uc.repo.GetUsersByRole(ctx, repository.GetUsersByRoleParams{
		OrganizationID: orgID,
		RoleID:         roleID,
	})
}
