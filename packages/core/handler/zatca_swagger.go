package handler

// Swagger response models describe the public JSON, including pgtype values
// that marshal as strings/null rather than their internal Go struct fields.
type ZatcaAPIError struct {
	StatusCode int         `json:"statusCode" example:"409"`
	Message    string      `json:"message" example:"ZATCA is disabled for this organization"`
	Data       interface{} `json:"data" extensions:"x-nullable"`
}

type ZatcaOrganizationSettingRequest struct {
	Enabled *bool `json:"enabled" binding:"required" example:"true"`
}
type ZatcaOrganizationSettingData struct {
	OrganizationID int32  `json:"organization_id" example:"1"`
	Enabled        bool   `json:"enabled" example:"true"`
	Source         string `json:"source" example:"organizations.metadata.zatca.enabled"`
}
type ZatcaOrganizationSettingResponse struct {
	StatusCode int                          `json:"statusCode" example:"200"`
	Message    string                       `json:"message" example:"Organization ZATCA setting updated"`
	Data       ZatcaOrganizationSettingData `json:"data"`
}

type ZatcaStatusData struct {
	Enabled              *bool  `json:"enabled" extensions:"x-nullable" example:"true"`
	EnabledSource        string `json:"enabled_source" example:"organization_metadata"`
	EnvironmentSource    string `json:"environment_source" example:"device_config"`
	OrganizationID       *int32 `json:"organization_id,omitempty" example:"1"`
	OrganizationRequired bool   `json:"organization_required,omitempty"`
	DeviceID             *int32 `json:"device_id,omitempty" example:"3"`
	Env                  string `json:"env,omitempty" enums:"sandbox,simulation,production" example:"simulation"`
	BaseURL              string `json:"base_url,omitempty" example:"https://gw-fatoora.zatca.gov.sa/e-invoicing/simulation"`
	VatID                string `json:"vat_id,omitempty" example:"310123456700003"`
	IsActive             *bool  `json:"is_active,omitempty"`
	ReportingReady       *bool  `json:"reporting_ready,omitempty"`
}
type ZatcaStatusAPIResponse struct {
	StatusCode int             `json:"statusCode" example:"200"`
	Message    string          `json:"message" example:"ZATCA status"`
	Data       ZatcaStatusData `json:"data"`
}

type ZatcaValidationMessage struct {
	Type     string `json:"type" example:"INFO"`
	Code     string `json:"code" example:"XSD_ZATCA_VALID"`
	Category string `json:"category" example:"XSD validation"`
	Message  string `json:"message"`
	Status   string `json:"status" example:"PASS"`
}
type ZatcaValidationResults struct {
	InfoMessages    []ZatcaValidationMessage `json:"infoMessages"`
	WarningMessages []ZatcaValidationMessage `json:"warningMessages"`
	ErrorMessages   []ZatcaValidationMessage `json:"errorMessages"`
	Status          string                   `json:"status" enums:"PASS,WARNING,ERROR" example:"PASS"`
}
type ZatcaGatewayData struct {
	ValidationResults ZatcaValidationResults `json:"validationResults"`
	ReportingStatus   *string                `json:"reportingStatus" extensions:"x-nullable" example:"REPORTED"`
	ClearanceStatus   *string                `json:"clearanceStatus,omitempty" extensions:"x-nullable"`
	QRSellerStatus    *string                `json:"qrSellertStatus,omitempty" extensions:"x-nullable"`
	QRBuyerStatus     *string                `json:"qrBuyertStatus,omitempty" extensions:"x-nullable"`
}
type ZatcaCheckAPIResponse struct {
	StatusCode int              `json:"statusCode" example:"200"`
	Message    string           `json:"message" example:"Compliance check response; this call does not report the invoice"`
	Data       ZatcaGatewayData `json:"data"`
}

type ZatcaDocumentStatus struct {
	Status string `json:"zatca_doc_status" example:"reported"`
	Valid  bool   `json:"valid" example:"true"`
}
type ZatcaDocumentData struct {
	ID             int64               `json:"id" example:"10"`
	EntityType     string              `json:"entity_type" example:"sales_order_v2"`
	EntityID       string              `json:"entity_id" example:"00000000-0000-4000-8000-000000000001"`
	DeviceConfigID int32               `json:"device_config_id" example:"3"`
	OrganizationID int32               `json:"organization_id" example:"1"`
	ZatcaUUID      string              `json:"zatca_uuid" format:"uuid"`
	ICV            int64               `json:"icv" example:"1"`
	PIH            string              `json:"pih"`
	XMLHash        string              `json:"xml_hash"`
	ZatcaStatus    ZatcaDocumentStatus `json:"zatca_status"`
	ZatcaResponse  interface{}         `json:"zatca_response"`
	QRCodeBase64   *string             `json:"qr_code_base64" extensions:"x-nullable"`
	SignedXML      *string             `json:"signed_xml" extensions:"x-nullable"`
	SubmittedAt    *string             `json:"submitted_at" format:"date-time" extensions:"x-nullable"`
	ClearedAt      *string             `json:"cleared_at" format:"date-time" extensions:"x-nullable"`
	CreatedAt      *string             `json:"created_at" format:"date-time" extensions:"x-nullable"`
	UpdatedAt      *string             `json:"updated_at" format:"date-time" extensions:"x-nullable"`
}
type ZatcaDocumentAPIResponse struct {
	StatusCode int               `json:"statusCode" example:"200"`
	Message    string            `json:"message"`
	Data       ZatcaDocumentData `json:"data"`
}

type ZatcaOnboardData struct {
	ID                 int32  `json:"id" example:"3"`
	DeviceSerial       string `json:"device_serial" example:"POS-01-SIM"`
	DeviceType         string `json:"device_type" enums:"pos,cloud"`
	ComplianceCSID     string `json:"compliance_csid"`
	ProductionCSID     string `json:"production_csid"`
	CSIDExpiry         string `json:"csid_expiry" format:"date-time"`
	ZatcaEnv           string `json:"zatca_env" enums:"sandbox,simulation,production" example:"simulation"`
	IsActive           bool   `json:"is_active" example:"true"`
	OnboardingStage    string `json:"onboarding_stage" example:"production"`
	ReportingReady     bool   `json:"reporting_ready" example:"true"`
	CompletionEndpoint string `json:"completion_endpoint" example:"/api/zatca/devices/3/complete-onboarding"`
}
type ZatcaOnboardAPIResponse struct {
	StatusCode int              `json:"statusCode" example:"200"`
	Message    string           `json:"message" example:"Device onboarded with ZATCA successfully"`
	Data       ZatcaOnboardData `json:"data"`
}
type ZatcaComplianceStep struct {
	DocumentType     string           `json:"document_type" example:"simplified-invoice"`
	Response         ZatcaGatewayData `json:"response"`
	AlreadyCompleted bool             `json:"already_completed,omitempty"`
}
type ZatcaCompletionData struct {
	DeviceID         int32                 `json:"device_id" example:"3"`
	ZatcaEnv         string                `json:"zatca_env" example:"simulation"`
	OnboardingStage  string                `json:"onboarding_stage" example:"production"`
	ReportingReady   bool                  `json:"reporting_ready" example:"true"`
	ComplianceChecks []ZatcaComplianceStep `json:"compliance_checks"`
}
type ZatcaCompletionAPIResponse struct {
	StatusCode int                 `json:"statusCode" example:"200"`
	Message    string              `json:"message"`
	Data       ZatcaCompletionData `json:"data"`
}
type ZatcaReceiptQRData struct {
	QRCodeBase64 string `json:"qr_code_base64"`
}
type ZatcaReceiptQRAPIResponse struct {
	StatusCode int                `json:"statusCode" example:"200"`
	Message    string             `json:"message" example:"Saved receipt QR"`
	Data       ZatcaReceiptQRData `json:"data"`
}

// Used by legacy configuration/sync responses whose data can be an array/object.
type ZatcaGenericAPIResponse struct {
	StatusCode int         `json:"statusCode" example:"200"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
}
