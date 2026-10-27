package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	uc *usecase.NotificationUseCase
}

func NewNotificationHandler(uc *usecase.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{uc: uc}
}

func (h *NotificationHandler) setup(c *gin.Context) (int32, int32, bool) {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found"})
		return 0, 0, false
	}
	org, ok := middleware.GetOrganizationIDFromContext(c)
	if !ok || org <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "organization claim missing"})
		return 0, 0, false
	}
	uid, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user claim missing"})
		return 0, 0, false
	}
	n, err := strconv.ParseInt(uid, 10, 32)
	if err != nil || n <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user claim"})
		return 0, 0, false
	}
	h.uc.SetRepository(repo)
	return org, int32(n), true
}

// -----------------------------------------------------
// In-App Notification Center Endpoints
// -----------------------------------------------------

// NotificationInAppItem represents an in-app notification item
type NotificationInAppItem struct {
	ID             string  `json:"id" example:"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"`
	OrganizationID int32   `json:"organization_id" example:"1"`
	UserID         int32   `json:"user_id" example:"42"`
	EventID        *string `json:"event_id,omitempty" example:"c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22"`
	Title          string  `json:"title" example:"Purchase Order Approved"`
	Message        string  `json:"message" example:"PO #PO-2026-001 has been approved by the manager."`
	Severity       string  `json:"severity" example:"info"`
	ActionURL      *string `json:"action_url,omitempty" example:"/purchasing/orders/12"`
	IsRead         bool    `json:"is_read" example:"false"`
	ReadAt         *string `json:"read_at,omitempty" example:"2026-10-25T20:30:00Z"`
	CreatedAt      string  `json:"created_at" example:"2026-10-25T20:00:00Z"`
}

// NotificationChannelItem represents a notification channel configuration
type NotificationChannelItem struct {
	ID             int32           `json:"id" example:"1"`
	OrganizationID int32           `json:"organization_id" example:"1"`
	ChannelType    string          `json:"channel_type" example:"whatsapp"`
	IsEnabled      bool            `json:"is_enabled" example:"true"`
	Config         json.RawMessage `json:"config"`
	UpdatedAt      string          `json:"updated_at" example:"2026-10-25T20:00:00Z"`
}

// NotificationTemplateItem represents a notification template
type NotificationTemplateItem struct {
	ID               int32   `json:"id" example:"1"`
	OrganizationID   int32   `json:"organization_id" example:"1"`
	EventTypeID      string  `json:"event_type_id" example:"pos.receipt"`
	ChannelType      string  `json:"channel_type" example:"whatsapp"`
	Locale           string  `json:"locale" example:"en"`
	SubjectTemplate  *string `json:"subject_template,omitempty" example:"Receipt #{{.TransactionNumber}}"`
	BodyTemplate     string  `json:"body_template" example:"Hello {{.CustomerName}}, total: {{.TotalAmount}}"`
	MetaTemplateName *string `json:"meta_template_name,omitempty" example:"pos_digital_receipt"`
	IsActive         bool    `json:"is_active" example:"true"`
	CreatedAt        string  `json:"created_at" example:"2026-10-25T20:00:00Z"`
	UpdatedAt        string  `json:"updated_at" example:"2026-10-25T20:00:00Z"`
}

// NotificationRoutingRuleItem represents a routing rule
type NotificationRoutingRuleItem struct {
	ID             int32           `json:"id" example:"1"`
	OrganizationID int32           `json:"organization_id" example:"1"`
	EventTypeID    string          `json:"event_type_id" example:"po.approved"`
	AudienceType   string          `json:"audience_type" example:"role"`
	TargetRoleID   *int32          `json:"target_role_id,omitempty" example:"2"`
	Channels       json.RawMessage `json:"channels"`
	Conditions     json.RawMessage `json:"conditions,omitempty"`
	IsActive       bool            `json:"is_active" example:"true"`
	CreatedAt      string          `json:"created_at" example:"2026-10-25T20:00:00Z"`
	UpdatedAt      string          `json:"updated_at" example:"2026-10-25T20:00:00Z"`
}

// NotificationListResponse represents the paginated response for in-app notifications
type NotificationListResponse struct {
	Items  []NotificationInAppItem `json:"items"`
	Limit  int32                   `json:"limit" example:"20"`
	Offset int32                   `json:"offset" example:"0"`
}

// UnreadCountResponse represents the unread notification badge count
type UnreadCountResponse struct {
	UnreadCount int64 `json:"unread_count" example:"5"`
}

// ChannelsListResponse represents the list of configured notification channels
type ChannelsListResponse struct {
	Channels []NotificationChannelItem `json:"channels"`
}

// TemplatesListResponse represents the list of notification templates
type TemplatesListResponse struct {
	Templates []NotificationTemplateItem `json:"templates"`
}

// RulesListResponse represents the list of notification routing rules
type RulesListResponse struct {
	Rules []NotificationRoutingRuleItem `json:"rules"`
}

// ListUserNotifications returns the paginated notifications for the current user
// @Summary      List user notifications
// @Description  Returns paginated in-app notifications for the currently authenticated user
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        limit         query     int     false  "Number of items to return" default(20)
// @Param        offset        query     int     false  "Offset to skip" default(0)
// @Param        is_read       query     bool    false  "Filter by read status"
// @Success      200  {object}  NotificationListResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/notifications [get]
func (h *NotificationHandler) ListUserNotifications(c *gin.Context) {
	orgID, userID, ok := h.setup(c)
	if !ok {
		return
	}

	limit := int32(20)
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = int32(n)
		}
	}
	offset := int32(0)
	if o := c.Query("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = int32(n)
		}
	}

	var isRead *bool
	if r := c.Query("is_read"); r != "" {
		b, err := strconv.ParseBool(r)
		if err == nil {
			isRead = &b
		}
	}

	items, err := h.uc.ListUserNotifications(c.Request.Context(), orgID, userID, isRead, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "limit": limit, "offset": offset})
}

// GetUnreadCount returns the badge count of unread notifications for the user
// @Summary      Get unread notifications count
// @Description  Returns the total count of unread in-app notifications for the bell badge
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Success      200  {object}  UnreadCountResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/notifications/unread-count [get]
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	orgID, userID, ok := h.setup(c)
	if !ok {
		return
	}

	count, err := h.uc.GetUnreadCount(c.Request.Context(), orgID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// MarkAsRead marks a specific notification as read
// @Summary      Mark notification as read
// @Description  Marks a specific notification as read by UUID
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        id            path      string  true   "Notification UUID"
// @Success      200  {object}  NotificationInAppItem
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/notifications/{id}/read [post]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	_, userID, ok := h.setup(c)
	if !ok {
		return
	}

	notifID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}

	notif, err := h.uc.MarkAsRead(c.Request.Context(), notifID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notif)
}

// MarkAllAsRead marks all unread notifications for the user as read
// @Summary      Mark all notifications as read
// @Description  Marks all pending notifications as read for the current user
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Success      200  {object}  SuccessResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/notifications/read-all [post]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	orgID, userID, ok := h.setup(c)
	if !ok {
		return
	}

	if err := h.uc.MarkAllAsRead(c.Request.Context(), orgID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// -----------------------------------------------------
// Channel Configuration Endpoints
// -----------------------------------------------------

type UpsertChannelRequest struct {
	IsActive bool            `json:"is_active" example:"true"`
	Config   json.RawMessage `json:"config" binding:"required"`
}

// ListChannels lists all configured notification channels
// @Summary      List notification channels
// @Description  Returns configuration status for all supported channels (whatsapp, email, sms, websocket)
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Success      200  {object}  ChannelsListResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/notifications/channels [get]
func (h *NotificationHandler) ListChannels(c *gin.Context) {
	orgID, _, ok := h.setup(c)
	if !ok {
		return
	}

	channels, err := h.uc.ListChannels(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"channels": channels})
}

// GetChannel gets the configuration of a specific notification channel
// @Summary      Get channel configuration
// @Description  Returns credentials and settings for a specific channel type (e.g. whatsapp, email)
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        type          path      string  true   "Channel type (whatsapp, email, sms, websocket)"
// @Success      200  {object}  NotificationChannelItem
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/notifications/channels/{type} [get]
func (h *NotificationHandler) GetChannel(c *gin.Context) {
	orgID, _, ok := h.setup(c)
	if !ok {
		return
	}

	chType := c.Param("type")
	ch, err := h.uc.GetChannel(c.Request.Context(), orgID, chType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	c.JSON(http.StatusOK, ch)
}

// UpsertChannel creates or updates a notification channel configuration
// @Summary      Upsert channel configuration
// @Description  Updates active status and JSON credentials for a channel type (e.g. Meta WhatsApp token, AWS SES SMTP)
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string               true  "Tenant identifier"
// @Param        Authorization header    string               true  "Bearer token"
// @Param        type          path      string               true  "Channel type (whatsapp, email)"
// @Param        body          body      UpsertChannelRequest true  "Channel payload"
// @Success      200  {object}  NotificationChannelItem
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/notifications/channels/{type} [put]
func (h *NotificationHandler) UpsertChannel(c *gin.Context) {
	orgID, _, ok := h.setup(c)
	if !ok {
		return
	}

	chType := c.Param("type")
	var req UpsertChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ch, err := h.uc.UpsertChannel(c.Request.Context(), orgID, chType, req.IsActive, req.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ch)
}

// -----------------------------------------------------
// Template Endpoints
// -----------------------------------------------------

type UpsertTemplateRequest struct {
	EventTypeID      string  `json:"event_type_id" binding:"required" example:"pos.receipt"`
	ChannelType      string  `json:"channel_type" binding:"required" example:"whatsapp"`
	Locale           string  `json:"locale" example:"en"`
	SubjectTemplate  *string `json:"subject_template,omitempty" example:"Receipt #{{.TransactionNumber}}"`
	BodyTemplate     string  `json:"body_template" binding:"required" example:"Hello {{.CustomerName}}, total: {{.TotalAmount}}"`
	MetaTemplateName *string `json:"meta_template_name,omitempty" example:"pos_digital_receipt"`
}

// ListTemplates returns notification templates for an organization
// @Summary      List notification templates
// @Description  Returns active notification templates optionally filtered by event_type_id
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        event_type_id query     string  false  "Event type ID filter (e.g. pos.receipt, po.approved)"
// @Success      200  {object}  TemplatesListResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/notifications/templates [get]
func (h *NotificationHandler) ListTemplates(c *gin.Context) {
	orgID, _, ok := h.setup(c)
	if !ok {
		return
	}

	var eventType *string
	if et := c.Query("event_type_id"); et != "" {
		eventType = &et
	}

	templates, err := h.uc.ListTemplates(c.Request.Context(), orgID, eventType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// UpsertTemplate creates or updates a notification template
// @Summary      Upsert notification template
// @Description  Creates or updates subject and body templates with interpolation variables (e.g. {{.CustomerName}})
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                true  "Tenant identifier"
// @Param        Authorization header    string                true  "Bearer token"
// @Param        body          body      UpsertTemplateRequest true  "Template payload"
// @Success      200  {object}  NotificationTemplateItem
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/notifications/templates [post]
func (h *NotificationHandler) UpsertTemplate(c *gin.Context) {
	orgID, _, ok := h.setup(c)
	if !ok {
		return
	}

	var req UpsertTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl, err := h.uc.UpsertTemplate(c.Request.Context(), usecase.UpsertTemplateInput{
		OrganizationID:   orgID,
		EventTypeID:      req.EventTypeID,
		ChannelType:      req.ChannelType,
		Locale:           req.Locale,
		SubjectTemplate:  req.SubjectTemplate,
		BodyTemplate:     req.BodyTemplate,
		MetaTemplateName: req.MetaTemplateName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tmpl)
}

// DeleteTemplate deletes a notification template by ID
// @Summary      Delete notification template
// @Description  Removes a notification template by its numeric ID
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Template ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/notifications/templates/{id} [delete]
func (h *NotificationHandler) DeleteTemplate(c *gin.Context) {
	orgID, _, ok := h.setup(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
		return
	}

	if err := h.uc.DeleteTemplate(c.Request.Context(), orgID, int32(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// -----------------------------------------------------
// Routing Rule Endpoints
// -----------------------------------------------------

type CreateRuleRequest struct {
	EventTypeID  string          `json:"event_type_id" binding:"required" example:"po.approved"`
	AudienceType string          `json:"audience_type" binding:"required" example:"role"`
	TargetRoleID *int32          `json:"target_role_id,omitempty" example:"2"`
	Channels     json.RawMessage `json:"channels" binding:"required"`
	Conditions   json.RawMessage `json:"conditions,omitempty"`
	IsActive     bool            `json:"is_active" example:"true"`
}

type UpdateRuleRequest struct {
	AudienceType string          `json:"audience_type" binding:"required" example:"role"`
	TargetRoleID *int32          `json:"target_role_id,omitempty" example:"2"`
	Channels     json.RawMessage `json:"channels" binding:"required"`
	Conditions   json.RawMessage `json:"conditions,omitempty"`
	IsActive     bool            `json:"is_active" example:"true"`
}

// ListRules lists all notification routing rules
// @Summary      List notification rules
// @Description  Returns routing rules mapping events to audiences, channels, and roles
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200  {object}  RulesListResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/notifications/rules [get]
func (h *NotificationHandler) ListRules(c *gin.Context) {
	orgID, _, ok := h.setup(c)
	if !ok {
		return
	}

	rules, err := h.uc.ListRules(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rules": rules})
}

// CreateRule creates a new notification routing rule
// @Summary      Create notification rule
// @Description  Creates a routing rule for an event type with audience targeting and channels
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string            true  "Tenant identifier"
// @Param        Authorization header    string            true  "Bearer token"
// @Param        body          body      CreateRuleRequest true  "Rule payload"
// @Success      201  {object}  NotificationRoutingRuleItem
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/notifications/rules [post]
func (h *NotificationHandler) CreateRule(c *gin.Context) {
	orgID, _, ok := h.setup(c)
	if !ok {
		return
	}

	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule, err := h.uc.CreateRule(c.Request.Context(), usecase.CreateRuleInput{
		OrganizationID: orgID,
		EventTypeID:    req.EventTypeID,
		AudienceType:   req.AudienceType,
		TargetRoleID:   req.TargetRoleID,
		Channels:       req.Channels,
		Conditions:     req.Conditions,
		IsActive:       req.IsActive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// UpdateRule updates an existing notification routing rule
// @Summary      Update notification rule
// @Description  Updates rule audience, target role, channels, or active status
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string            true  "Tenant identifier"
// @Param        Authorization header    string            true  "Bearer token"
// @Param        id            path      int               true  "Rule ID"
// @Param        body          body      UpdateRuleRequest true  "Rule payload"
// @Success      200  {object}  NotificationRoutingRuleItem
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/notifications/rules/{id} [put]
func (h *NotificationHandler) UpdateRule(c *gin.Context) {
	orgID, _, ok := h.setup(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule id"})
		return
	}

	var req UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule, err := h.uc.UpdateRule(c.Request.Context(), usecase.UpdateRuleInput{
		ID:             int32(id),
		OrganizationID: orgID,
		AudienceType:   req.AudienceType,
		TargetRoleID:   req.TargetRoleID,
		Channels:       req.Channels,
		Conditions:     req.Conditions,
		IsActive:       req.IsActive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// DeleteRule deletes a notification routing rule by ID
// @Summary      Delete notification rule
// @Description  Deletes an existing notification routing rule
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Rule ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/notifications/rules/{id} [delete]
func (h *NotificationHandler) DeleteRule(c *gin.Context) {
	orgID, _, ok := h.setup(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule id"})
		return
	}

	if err := h.uc.DeleteRule(c.Request.Context(), orgID, int32(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// -----------------------------------------------------
// Route Registration
// -----------------------------------------------------

func (h *NotificationHandler) RegisterRoutes(rg *gin.RouterGroup) {
	// User Inbox & Bell
	notifs := rg.Group("/notifications")
	{
		notifs.GET("", h.ListUserNotifications)
		notifs.GET("/unread-count", h.GetUnreadCount)
		notifs.POST("/:id/read", h.MarkAsRead)
		notifs.POST("/read-all", h.MarkAllAsRead)

		// Settings & Channel Configuration
		notifs.GET("/channels", h.ListChannels)
		notifs.GET("/channels/:type", h.GetChannel)
		notifs.PUT("/channels/:type", h.UpsertChannel)

		// Templates
		notifs.GET("/templates", h.ListTemplates)
		notifs.POST("/templates", h.UpsertTemplate)
		notifs.DELETE("/templates/:id", h.DeleteTemplate)

		// Routing Rules
		notifs.GET("/rules", h.ListRules)
		notifs.POST("/rules", h.CreateRule)
		notifs.PUT("/rules/:id", h.UpdateRule)
		notifs.DELETE("/rules/:id", h.DeleteRule)
	}
}
