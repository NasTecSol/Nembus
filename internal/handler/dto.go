package handler

import "encoding/json"

// UserResponse represents a user in API responses
type UserResponse struct {
	ID             int32  `json:"id" example:"1"`
	OrganizationID int32  `json:"organization_id" example:"1"`
	Username       string `json:"username" example:"johndoe"`
	Email          string `json:"email" example:"john@example.com"`
	FirstName      string `json:"first_name" example:"John"`
	LastName       string `json:"last_name" example:"Doe"`
	EmployeeCode   string `json:"employee_code,omitempty" example:"EMP001"`
	IsActive       bool   `json:"is_active" example:"true"`
	CreatedAt      string `json:"created_at" example:"2026-01-24T21:43:00Z"`
	UpdatedAt      string `json:"updated_at" example:"2026-01-24T21:43:00Z"`
}

// LoginRequest represents login request body
type LoginRequest struct {
	UserLogin string `json:"user_login" binding:"required" example:"johndoe"`
	Password  string `json:"password" binding:"required" example:"securepassword123"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	Type  string `json:"type" example:"Bearer"`
}

// CreateUserRequest represents user creation request
type CreateUserRequest struct {
	FirstName    string  `json:"first_name" binding:"required" example:"John"`
	LastName     string  `json:"last_name" example:"Doe"`
	Username     string  `json:"username" binding:"required" example:"johndoe"`
	Email        string  `json:"email" binding:"required" example:"john@example.com"`
	IsActive     bool    `json:"is_active" example:"true"`
	Password     *string `json:"password,omitempty" example:"securepassword123"`
	EmployeeCode *string `json:"employee_code,omitempty" example:"EMP001"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid request"`
	Details string `json:"details,omitempty" example:"Additional error details"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message" example:"User created successfully"`
}

// OrganizationResponse represents an organization in API responses
type OrganizationResponse struct {
	ID                int32  `json:"id" example:"1"`
	Name              string `json:"name" example:"Acme Corporation"`
	Code              string `json:"code" example:"ACME"`
	LegalName         string `json:"legal_name,omitempty" example:"Acme Corporation Inc."`
	TaxID             string `json:"tax_id,omitempty" example:"TAX123456"`
	CurrencyCode      string `json:"currency_code" example:"USD"`
	FiscalYearVariant string `json:"fiscal_year_variant,omitempty" example:"FY"`
	IsActive          bool   `json:"is_active" example:"true"`
	CreatedAt         string `json:"created_at" example:"2026-01-24T21:43:00Z"`
	UpdatedAt         string `json:"updated_at" example:"2026-01-24T21:43:00Z"`
}

// CreateOrganizationRequest represents organization creation request
type CreateOrganizationRequest struct {
	Name              string  `json:"name" binding:"required" example:"Acme Corporation"`
	Code              string  `json:"code" binding:"required" example:"ACME"`
	LegalName         *string `json:"legal_name,omitempty" example:"Acme Corporation Inc."`
	TaxID             *string `json:"tax_id,omitempty" example:"TAX123456"`
	CurrencyCode      *string `json:"currency_code,omitempty" example:"USD"`
	FiscalYearVariant *string `json:"fiscal_year_variant,omitempty" example:"FY"`
	IsActive          bool    `json:"is_active" example:"true"`
}

// UpdateOrganizationRequest represents organization update request
type UpdateOrganizationRequest struct {
	Name              *string `json:"name,omitempty" example:"Acme Corporation"`
	LegalName         *string `json:"legal_name,omitempty" example:"Acme Corporation Inc."`
	TaxID             *string `json:"tax_id,omitempty" example:"TAX123456"`
	CurrencyCode      *string `json:"currency_code,omitempty" example:"USD"`
	FiscalYearVariant *string `json:"fiscal_year_variant,omitempty" example:"FY"`
	IsActive          *bool   `json:"is_active,omitempty" example:"true"`
}

// ModuleResponse represents a module in API responses
type ModuleResponse struct {
	ID           int32  `json:"id" example:"1"`
	Name         string `json:"name" example:"Sales"`
	Code         string `json:"code" example:"SALES"`
	Description  string `json:"description,omitempty" example:"Sales management module"`
	Icon         string `json:"icon,omitempty" example:"sales-icon"`
	IsActive     bool   `json:"is_active" example:"true"`
	DisplayOrder int32  `json:"display_order" example:"1"`
	CreatedAt    string `json:"created_at" example:"2026-01-24T21:43:00Z"`
	UpdatedAt    string `json:"updated_at" example:"2026-01-24T21:43:00Z"`
}

// CreateModuleRequest represents module creation request
type CreateModuleRequest struct {
	Name         string  `json:"name" binding:"required" example:"Sales"`
	Code         string  `json:"code" binding:"required" example:"SALES"`
	Description  *string `json:"description,omitempty" example:"Sales management module"`
	Icon         *string `json:"icon,omitempty" example:"sales-icon"`
	IsActive     bool    `json:"is_active" example:"true"`
	DisplayOrder int32   `json:"display_order" example:"1"`
}

// UpdateModuleRequest represents module update request
type UpdateModuleRequest struct {
	Name         *string `json:"name,omitempty" example:"Sales"`
	Description  *string `json:"description,omitempty" example:"Updated description"`
	Icon         *string `json:"icon,omitempty" example:"sales-icon"`
	IsActive     *bool   `json:"is_active,omitempty" example:"true"`
	DisplayOrder *int32  `json:"display_order,omitempty" example:"1"`
}

// RoleResponse represents a role in API responses
type RoleResponse struct {
	ID           int32  `json:"id" example:"1"`
	Name         string `json:"name" example:"Admin"`
	Code         string `json:"code" example:"ADMIN"`
	Description  string `json:"description,omitempty" example:"Administrator role with full access"`
	IsSystemRole bool   `json:"is_system_role" example:"false"`
	IsActive     bool   `json:"is_active" example:"true"`
	CreatedAt    string `json:"created_at" example:"2026-01-24T21:43:00Z"`
	UpdatedAt    string `json:"updated_at" example:"2026-01-24T21:43:00Z"`
}

// CreateRoleRequest represents role creation request
type CreateRoleRequest struct {
	Name         string  `json:"name" binding:"required" example:"Admin"`
	Code         string  `json:"code" binding:"required" example:"ADMIN"`
	Description  *string `json:"description,omitempty" example:"Administrator role with full access"`
	IsSystemRole bool    `json:"is_system_role" example:"false"`
	IsActive     bool    `json:"is_active" example:"true"`
}

// UpdateRoleRequest represents role update request
type UpdateRoleRequest struct {
	Name        string  `json:"name" binding:"required" example:"Admin"`
	Description *string `json:"description,omitempty" example:"Updated description"`
	IsActive    bool    `json:"is_active" example:"true"`
}

// AssignPermissionItem represents one permission assignment
type AssignPermissionItem struct {
	PermissionID int32  `json:"permission_id" example:"1"`
	Scope        string `json:"scope,omitempty" example:"read,write"`
	Metadata     string `json:"metadata,omitempty" example:"{\"level\":\"admin\"}"`
}

// AssignPermissionToRoleRequest represents assigning multiple permissions to a role
type AssignPermissionToRoleRequest struct {
	Permissions []AssignPermissionItem `json:"permissions"`
}

// RemovePermissionFromRoleRequest represents request body for bulk permission removal
type RemovePermissionFromRoleRequest struct {
	PermissionIDs []int32 `json:"permission_ids"`
}

// RoleNavigationResponse represents the response for GetNavigationByRoleCodeWithUserCounts
type RoleNavigationResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Data       struct {
		Navigation interface{} `json:"navigation"` // JSON structure returned by navigation use case
		UserCount  int         `json:"user_count"` // Number of users assigned to this role
		// Users []User `json:"users,omitempty"` // optional, if you want to include full user list
	} `json:"data"`
}

/// MenuResponse represents a menu in API responses
type MenuResponse struct {
	ID           int32   `json:"id" example:"1"`
	ModuleID     int32   `json:"module_id" example:"1"`
	ParentMenuID *int32  `json:"parent_menu_id,omitempty" example:"2"`
	Name         string  `json:"name" example:"Dashboard"`
	Code         string  `json:"code" example:"DASHBOARD"`
	RoutePath    *string `json:"route_path,omitempty" example:"/dashboard"`
	Icon         *string `json:"icon,omitempty" example:"dashboard-icon"`
	DisplayOrder *int32  `json:"display_order,omitempty" example:"1"`
	IsActive     bool    `json:"is_active" example:"true"`
	Metadata     string  `json:"metadata,omitempty" example:"{\"color\":\"blue\"}"`

	CreatedAt string `json:"created_at" example:"2026-01-24T21:43:00Z"`
	UpdatedAt string `json:"updated_at" example:"2026-01-24T21:43:00Z"`
}

// CreateMenuRequest represents menu creation request
type CreateMenuRequest struct {
	ModuleID     int32   `json:"module_id" binding:"required" example:"1"`
	ParentMenuID *int32  `json:"parent_menu_id,omitempty" example:"2"`
	Name         string  `json:"name" binding:"required" example:"Dashboard"`
	Code         string  `json:"code" binding:"required" example:"DASHBOARD"`
	RoutePath    *string `json:"route_path,omitempty" example:"/dashboard"`
	Icon         *string `json:"icon,omitempty" example:"dashboard-icon"`
	DisplayOrder *int32  `json:"display_order,omitempty" example:"1"`
	IsActive     bool    `json:"is_active" example:"true"`
	Metadata     string  `json:"metadata,omitempty" example:"{\"color\":\"blue\"}"`
}

// UpdateMenuRequest represents menu update request
type UpdateMenuRequest struct {
	ParentMenuID *int32  `json:"parent_menu_id,omitempty" example:"2"`
	Name         string  `json:"name" example:"Dashboard"`
	RoutePath    *string `json:"route_path,omitempty" example:"/dashboard"`
	Icon         *string `json:"icon,omitempty" example:"dashboard-icon"`
	DisplayOrder *int32  `json:"display_order,omitempty" example:"1"`
	IsActive     bool    `json:"is_active" example:"true"`
	Metadata     string  `json:"metadata,omitempty" example:"{\"color\":\"blue\"}"`
}

// ToggleMenuActiveRequest represents request body to toggle menu active status
type ToggleMenuActiveRequest struct {
	IsActive bool `json:"is_active" example:"true"`
}

// ListMenusResponse represents a list of menus
type ListMenusResponse struct {
	Menus []MenuResponse `json:"menus"`
}

// GetMenuByCodeResponse represents response for GetMenuByCode
type GetMenuByCodeResponse struct {
	Menu MenuResponse `json:"menu"`
}

// ListMenusByParentResponse represents response for listing menus by parent
type ListMenusByParentResponse struct {
	ParentID int32          `json:"parent_id" example:"2"`
	Menus    []MenuResponse `json:"menus"`
}

// ListMenusByModuleResponse represents response for listing menus by module
type ListMenusByModuleResponse struct {
	ModuleID int32          `json:"module_id" example:"1"`
	Menus    []MenuResponse `json:"menus"`
}

// ListActiveMenusByModuleResponse represents response for listing active menus by module
type ListActiveMenusByModuleResponse struct {
	ModuleID int32          `json:"module_id" example:"1"`
	Menus    []MenuResponse `json:"menus"`
}

// SubmenuResponse represents a submenu in API responses
type SubmenuResponse struct {
	ID              int32   `json:"id" example:"1"`
	MenuID          int32   `json:"menu_id" example:"1"`
	ParentSubmenuID *int32  `json:"parent_submenu_id,omitempty" example:"2"`
	Name            string  `json:"name" example:"User Management"`
	Code            string  `json:"code" example:"USER_MANAGEMENT"`
	RoutePath       *string `json:"route_path,omitempty" example:"/users"`
	Icon            *string `json:"icon,omitempty" example:"user-icon"`
	DisplayOrder    *int32  `json:"display_order,omitempty" example:"1"`
	IsActive        bool    `json:"is_active" example:"true"`
	Metadata        string  `json:"metadata,omitempty" example:"{\"color\":\"blue\"}"`

	CreatedAt string `json:"created_at" example:"2026-01-24T21:43:00Z"`
	UpdatedAt string `json:"updated_at" example:"2026-01-24T21:43:00Z"`
}

// CreateSubmenuRequest represents submenu creation request
type CreateSubmenuRequest struct {
	MenuID          int32   `json:"menu_id" binding:"required" example:"1"`
	ParentSubmenuID *int32  `json:"parent_submenu_id,omitempty" example:"2"`
	Name            string  `json:"name" binding:"required" example:"User Management"`
	Code            string  `json:"code" binding:"required" example:"USER_MANAGEMENT"`
	RoutePath       *string `json:"route_path,omitempty" example:"/users"`
	Icon            *string `json:"icon,omitempty" example:"user-icon"`
	DisplayOrder    *int32  `json:"display_order,omitempty" example:"1"`
	IsActive        bool    `json:"is_active" example:"true"`
	Metadata        string  `json:"metadata,omitempty" example:"{\"color\":\"blue\"}"`
}

// UpdateSubmenuRequest represents submenu update request
type UpdateSubmenuRequest struct {
	ParentSubmenuID *int32  `json:"parent_submenu_id,omitempty" example:"2"`
	Name            string  `json:"name" example:"User Management"`
	RoutePath       *string `json:"route_path,omitempty" example:"/users"`
	Icon            *string `json:"icon,omitempty" example:"user-icon"`
	DisplayOrder    *int32  `json:"display_order,omitempty" example:"1"`
	IsActive        bool    `json:"is_active" example:"true"`
	Metadata        string  `json:"metadata,omitempty" example:"{\"color\":\"blue\"}"`
}

// ToggleSubmenuActiveRequest represents request body to toggle submenu active status
type ToggleSubmenuActiveRequest struct {
	IsActive bool `json:"is_active" example:"true"`
}

// ListSubmenusResponse represents a list of submenus
type ListSubmenusResponse struct {
	Submenus []SubmenuResponse `json:"submenus"`
}

// GetSubmenuByCodeResponse represents response for GetSubmenuByCode
type GetSubmenuByCodeResponse struct {
	Submenu SubmenuResponse `json:"submenu"`
}

// ListSubmenusByParentResponse represents response for listing submenus by parent
type ListSubmenusByParentResponse struct {
	ParentID int32             `json:"parent_id" example:"2"`
	Submenus []SubmenuResponse `json:"submenus"`
}

// ListSubmenusByMenuResponse represents response for listing submenus by menu
type ListSubmenusByMenuResponse struct {
	MenuID   int32             `json:"menu_id" example:"1"`
	Submenus []SubmenuResponse `json:"submenus"`
}

// ListActiveSubmenusByMenuResponse represents response for listing active submenus by menu
type ListActiveSubmenusByMenuResponse struct {
	MenuID   int32             `json:"menu_id" example:"1"`
	Submenus []SubmenuResponse `json:"submenus"`
}

// =====================================================
// POS module
// =====================================================

// CreatePosProductRequest represents POS "add product" request.
type CreatePosProductRequest struct {
	OrganizationID       int32   `json:"organization_id" binding:"required"`
	SKU                  string  `json:"sku" binding:"required"`
	Name                 string  `json:"name" binding:"required"`
	Description          *string `json:"description,omitempty"`
	CategoryID           *int32  `json:"category_id,omitempty"`
	BrandID              *int32  `json:"brand_id,omitempty"`
	BaseUomID            *int32  `json:"base_uom_id,omitempty"`
	ProductType          *string `json:"product_type,omitempty"`
	TaxCategoryID        *int32  `json:"tax_category_id,omitempty"`
	IsSerialized         *bool   `json:"is_serialized,omitempty"`
	IsBatchManaged       *bool   `json:"is_batch_managed,omitempty"`
	IsActive             *bool   `json:"is_active,omitempty"`
	IsSellable           *bool   `json:"is_sellable,omitempty"`
	IsPurchasable        *bool   `json:"is_purchasable,omitempty"`
	AllowDecimalQuantity *bool   `json:"allow_decimal_quantity,omitempty"`
	TrackInventory       *bool   `json:"track_inventory,omitempty"`
	Barcode              *string `json:"barcode,omitempty"`
	RetailPrice          *string `json:"retail_price,omitempty"` // decimal as string, e.g. "12.50"
}

type AddProductRequest struct {
	OrganizationID       int32   `json:"organization_id" binding:"required"`
	SKU                  string  `json:"sku" binding:"required"`
	Name                 string  `json:"name" binding:"required"`
	Description          *string `json:"description"`
	CategoryID           *int32  `json:"category_id"`
	BrandID              *int32  `json:"brand_id"`
	BaseUomID            *int32  `json:"base_uom_id"`
	ProductType          *string `json:"product_type"`
	TaxCategoryID        *int32  `json:"tax_category_id"`
	IsSerialized         *bool   `json:"is_serialized"`
	IsBatchManaged       *bool   `json:"is_batch_managed"`
	IsActive             *bool   `json:"is_active"`
	IsSellable           *bool   `json:"is_sellable"`
	IsPurchasable        *bool   `json:"is_purchasable"`
	AllowDecimalQuantity *bool   `json:"allow_decimal_quantity"`
	TrackInventory       *bool   `json:"track_inventory"`
	Barcode              *string `json:"barcode"`
	RetailPrice          *string `json:"retail_price"`
}

// CreatePOSTerminalRequest is the request body for creating a POS terminal.
type CreatePOSTerminalRequest struct {
	StoreID      int32   `json:"store_id" binding:"required"`
	TerminalCode string  `json:"terminal_code" binding:"required"`
	TerminalName *string `json:"terminal_name,omitempty"`
	DeviceID     *string `json:"device_id,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

// UpdatePOSTerminalRequest is the request body for updating a POS terminal.
type UpdatePOSTerminalRequest struct {
	TerminalName *string `json:"terminal_name,omitempty"`
	DeviceID     *string `json:"device_id,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

// TogglePOSTerminalActiveRequest is the request body for toggling terminal active state.
type TogglePOSTerminalActiveRequest struct {
	IsActive bool `json:"is_active"`
}

// CreateStorageLocationRequest is the request body for creating a storage location.
type CreateStorageLocationRequest struct {
	StoreID          int32   `json:"store_id" binding:"required"`
	Code             string  `json:"code" binding:"required"`
	Name             string  `json:"name" binding:"required"`
	LocationType     *string `json:"location_type,omitempty"`
	ParentLocationID *int32  `json:"parent_location_id,omitempty"`
	IsActive         *bool   `json:"is_active,omitempty"`
}

// UpdateStorageLocationRequest is the request body for updating a storage location.
type UpdateStorageLocationRequest struct {
	Name             *string `json:"name,omitempty"`
	LocationType     *string `json:"location_type,omitempty"`
	ParentLocationID *int32  `json:"parent_location_id,omitempty"`
	IsActive         *bool   `json:"is_active,omitempty"`
}

// ToggleStorageLocationActiveRequest is the request body for toggling storage location active state.
type ToggleStorageLocationActiveRequest struct {
	IsActive bool `json:"is_active"`
}

type CreateTenantRequest struct {
	TenantName string                 `json:"tenant_name" example:"Acme Corporation"`
	Slug       string                 `json:"slug" example:"acme"`
	DbConnStr  string                 `json:"db_conn_str" example:"postgres://user:pass@localhost:5432/acme_db"`
	IsActive   bool                   `json:"is_active" example:"true"`
	Settings   map[string]interface{} `json:"settings"`
}

type UpdateTenantRequest struct {
	TenantName *string                `json:"tenant_name,omitempty" example:"Acme Corp"`
	Slug       *string                `json:"slug,omitempty" example:"acme-updated"`
	DbConnStr  *string                `json:"db_conn_str,omitempty" example:"postgres://user:pass@localhost:5432/new_db"`
	IsActive   *bool                  `json:"is_active,omitempty" example:"false"`
	Settings   map[string]interface{} `json:"settings"`
}

type TenantResponse struct {
	ID         string                 `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	TenantName string                 `json:"tenant_name" example:"Acme Corporation"`
	Slug       string                 `json:"slug" example:"acme"`
	DbConnStr  string                 `json:"db_conn_str" example:"postgres://user:pass@localhost:5432/acme_db"`
	IsActive   bool                   `json:"is_active" example:"true"`
	Settings   map[string]interface{} `json:"settings"`
	CreatedAt  string                 `json:"created_at" example:"2025-01-01T10:00:00Z"`
	UpdatedAt  string                 `json:"updated_at" example:"2025-01-01T10:00:00Z"`
}

type AssignRoleToUserRequest struct {
	RoleID   int32                  `json:"role_id" binding:"required" example:"1"`
	StoreID  *int32                 `json:"store_id,omitempty" example:"10"` // optional
	Metadata map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

// CreateStoreRequest represents request body for creating a store
type CreateStoreRequest struct {
	Name          string                 `json:"name" example:"Main Warehouse"`
	Code          string                 `json:"code" example:"MAIN_WH"`
	StoreType     *string                `json:"store_type,omitempty" example:"warehouse"`
	ParentStoreID *int32                 `json:"parent_store_id,omitempty"` // no example`
	IsWarehouse   bool                   `json:"is_warehouse" example:"true"`
	IsPOSEnabled  bool                   `json:"is_pos_enabled" example:"false"`
	Timezone      *string                `json:"timezone,omitempty" example:"Asia/Karachi"`
	IsActive      bool                   `json:"is_active" example:"true"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

// UpdateStoreRequest represents request body for updating a store
type UpdateStoreRequest struct {
	Name         *string                `json:"name,omitempty"`
	StoreType    *string                `json:"store_type,omitempty"`
	IsWarehouse  *bool                  `json:"is_warehouse,omitempty"`
	IsPOSEnabled *bool                  `json:"is_pos_enabled,omitempty"`
	Timezone     *string                `json:"timezone,omitempty"`
	IsActive     *bool                  `json:"is_active,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

// StoreResponse represents a store object in responses
type StoreResponse struct {
	ID             int32                  `json:"id" example:"1"`
	OrganizationID int32                  `json:"organization_id" example:"1"`
	ParentStoreID  *int32                 `json:"parent_store_id,omitempty"` // no example
	Name           string                 `json:"name" example:"Main Warehouse"`
	Code           string                 `json:"code" example:"MAIN_WH"`
	StoreType      string                 `json:"store_type" example:"warehouse"`
	IsWarehouse    bool                   `json:"is_warehouse" example:"true"`
	IsPOSEnabled   bool                   `json:"is_pos_enabled" example:"false"`
	Timezone       string                 `json:"timezone" example:"Asia/Karachi"`
	IsActive       bool                   `json:"is_active" example:"true"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
	CreatedAt      string                 `json:"created_at" example:"2026-02-04T10:15:30Z"`
	UpdatedAt      string                 `json:"updated_at" example:"2026-02-04T10:15:30Z"`
}

// UpdateUserRequest defines the request body for updating a user
type UpdateUserRequest struct {
	Email        *string                `json:"email,omitempty"`
	FirstName    *string                `json:"first_name,omitempty"`
	LastName     *string                `json:"last_name,omitempty"`
	EmployeeCode *string                `json:"employee_code,omitempty"`
	IsActive     *bool                  `json:"is_active,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateUserPasswordRequest defines the request body for updating a user's password
type UpdateUserPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
}

// GrantStoreAccessRequest defines the request body for granting store access to a user
type GrantStoreAccessRequest struct {
	StoreID   int32                  `json:"store_id" binding:"required"` // ID of the store to grant access
	IsPrimary bool                   `json:"is_primary"`                  // Whether this store should be the primary store
	Metadata  map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

type RevokeRoleRequest struct {
	RoleID int32 `json:"role_id" binding:"required" example:"1"`
}

type RevokeStoreAccessRequest struct {
	StoreID int32 `json:"store_id" binding:"required" example:"10"`
}

// =====================================================
// Restaurant Module
// =====================================================

type CreateRestaurantTableRequest struct {
	StoreID     int32  `json:"store_id" binding:"required"`
	TableNumber string `json:"table_number" binding:"required"`
	TableName   string `json:"table_name"`
	Section     string `json:"section"`
	Capacity    int32  `json:"capacity"`
	IsActive    bool   `json:"is_active"`
	Metadata    string `json:"metadata"`
}

type CreateMenuCategoryRequest struct {
	StoreID          int32  `json:"store_id" binding:"required"`
	ParentCategoryID *int32 `json:"parent_category_id"`
	Name             string `json:"name" binding:"required"`
	Code             string `json:"code" binding:"required"`
	Description      string `json:"description"`
	CategoryLevel    int32  `json:"category_level"`
	DisplayOrder     int32  `json:"display_order"`
	Icon             string `json:"icon"`
	ImageUrl         string `json:"image_url"`
	IsActive         bool   `json:"is_active"`
	Metadata         string `json:"metadata"`
}

type CreateMenuItemRequest struct {
	StoreID            int32  `json:"store_id" binding:"required"`
	MenuCategoryID     int32  `json:"menu_category_id" binding:"required"`
	ProductID          *int32 `json:"product_id"`
	RecipeID           *int32 `json:"recipe_id"`
	Name               string `json:"name" binding:"required"`
	ShortName          string `json:"short_name"`
	Description        string `json:"description"`
	ImageUrl           string `json:"image_url"`
	BasePrice          string `json:"base_price" binding:"required"`
	PreparationTimeMin int32  `json:"preparation_time_min"`
	TaxCategoryID      *int32 `json:"tax_category_id"`
	IsAvailable        bool   `json:"is_available"`
	IsActive           bool   `json:"is_active"`
	DisplayOrder       int32  `json:"display_order"`
	Metadata           string `json:"metadata"`
}

type CreateRecipeRequest struct {
	OrganizationID     int32  `json:"organization_id" binding:"required"`
	RecipeCode         string `json:"recipe_code" binding:"required"`
	RecipeName         string `json:"recipe_name" binding:"required"`
	Description        string `json:"description"`
	FinishedProductID  *int32 `json:"finished_product_id"`
	YieldQuantity      string `json:"yield_quantity"`
	YieldUomID         *int32 `json:"yield_uom_id"`
	PreparationSteps   string `json:"preparation_steps"`
	PreparationTimeMin int32  `json:"preparation_time_min"`
	CookingTimeMin     int32  `json:"cooking_time_min"`
	IsActive           bool   `json:"is_active"`
	Metadata           string `json:"metadata"`
}

type CreateRecipeIngredientRequest struct {
	RecipeID         int32  `json:"recipe_id"`
	ProductID        int32  `json:"product_id" binding:"required"`
	ProductVariantID *int32 `json:"product_variant_id"`
	Quantity         string `json:"quantity" binding:"required"`
	UomID            *int32 `json:"uom_id"`
	IsOptional       bool   `json:"is_optional"`
	IsByproduct      bool   `json:"is_byproduct"`
	LineNumber       int32  `json:"line_number"`
	Metadata         string `json:"metadata"`
}

type CreateRestaurantOrderRequest struct {
	StoreID          int32  `json:"store_id" binding:"required"`
	TableID          *int32 `json:"table_id"`
	CashierID        *int32 `json:"cashier_id"`
	CashierSessionID *int32 `json:"cashier_session_id"`
	CustomerID       *int32 `json:"customer_id"`
	OrderNumber      string `json:"order_number"`
	OrderSource      string `json:"order_source"`
	Status           string `json:"status"`
	Subtotal         string `json:"subtotal"`
	DiscountAmount   string `json:"discount_amount"`
	TaxAmount        string `json:"tax_amount"`
	TotalAmount      string `json:"total_amount"`
	AmountPaid       string `json:"amount_paid"`
	ChangeGiven      string `json:"change_given"`
	Notes            string `json:"notes"`
	PosTransactionID *int32 `json:"pos_transaction_id"`
	Metadata         string `json:"metadata"`
}

type CreateMenuItemModifierRequest struct {
	MenuItemID      int32  `json:"menu_item_id" binding:"required"`
	ModifierName    string `json:"modifier_name" binding:"required"`
	ModifierType    string `json:"modifier_type"`
	PriceAdjustment string `json:"price_adjustment"`
	IsActive        bool   `json:"is_active"`
	DisplayOrder    int32  `json:"display_order"`
	Metadata        string `json:"metadata"`
}

type CreateWasteLogRequest struct {
	StoreID     int32  `json:"store_id" binding:"required"`
	ProductID   *int32 `json:"product_id"`
	MenuItemID  *int32 `json:"menu_item_id"`
	RecipeID    *int32 `json:"recipe_id"`
	WasteSource string `json:"waste_source"`
	Quantity    string `json:"quantity" binding:"required"`
	UomID       *int32 `json:"uom_id"`
	UnitCost    string `json:"unit_cost"`
	TotalCost   string `json:"total_cost"`
	Reason      string `json:"reason"`
	LoggedBy    *int32 `json:"logged_by"`
	OrderID     *int32 `json:"order_id"`
	WastedAt    string `json:"wasted_at"`
	Metadata    string `json:"metadata"`
}

type CreateKioskSessionRequest struct {
	PosTerminalID int32  `json:"pos_terminal_id" binding:"required"`
	StoreID       int32  `json:"store_id" binding:"required"`
	SessionToken  string `json:"session_token" binding:"required"`
	Status        string `json:"status"`
	OpenedAt      string `json:"opened_at"`
	Metadata      string `json:"metadata"`
}

type CreateOnlineOrderRequest struct {
	StoreID    int32                              `json:"store_id" binding:"required"`
	CustomerID *int32                             `json:"customer_id"`
	Items      []CreateRestaurantOrderItemRequest `json:"items" binding:"required"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type SettleOrderRequest struct {
	PosTransactionID int32 `json:"pos_transaction_id" binding:"required"`
}

type CreateRestaurantOrderItemRequest struct {
	MenuItemID        int32  `json:"menu_item_id" binding:"required"`
	Quantity          string `json:"quantity" binding:"required"`
	UnitPrice         string `json:"unit_price" binding:"required"`
	ModifiersSnapshot string `json:"modifiers_snapshot"`
	ModifiersTotal    string `json:"modifiers_total"`
	DiscountAmount    string `json:"discount_amount"`
	TaxAmount         string `json:"tax_amount"`
	Subtotal          string `json:"subtotal" binding:"required"`
	LineNumber        int32  `json:"line_number"`
	Notes             string `json:"notes"`
	Status            string `json:"status"`
	Metadata          string `json:"metadata"`
}

// =====================================================
// Cart + Order (Enhanced schema) DTOs
// =====================================================

// CreateCartRequest represents request body for creating a cart.
type CreateCartRequest struct {
	CartNumber      string                 `json:"cart_number" binding:"required" example:"CART-20260210-0001"`
	OrganizationID  int32                  `json:"organization_id" binding:"required" example:"1"`
	StoreID         *int32                 `json:"store_id,omitempty" example:"10"`
	CustomerID      *int32                 `json:"customer_id,omitempty" example:"123"`
	GuestIdentifier *string                `json:"guest_identifier,omitempty" example:"guest-device-abc"`
	GuestEmail      *string                `json:"guest_email,omitempty" example:"guest@example.com"`
	GuestPhone      *string                `json:"guest_phone,omitempty" example:"+15551234567"`
	CartStatus      string                 `json:"cart_status" binding:"required" example:"active"`
	CartType        string                 `json:"cart_type" binding:"required" example:"standard"`
	Channel         *string                `json:"channel,omitempty" example:"web"`
	DeviceInfo      map[string]interface{} `json:"device_info,omitempty" swaggertype:"object"`
	CreatedByUserID *int32                 `json:"created_by_user_id,omitempty" example:"1"`
	CashierID       *int32                 `json:"cashier_id,omitempty" example:"1"`
	PosTerminalID   *int32                 `json:"pos_terminal_id,omitempty" example:"1"`
	ShippingAddress map[string]interface{} `json:"shipping_address,omitempty" swaggertype:"object"`
	BillingAddress  map[string]interface{} `json:"billing_address,omitempty" swaggertype:"object"`
	ShippingMethod  *string                `json:"shipping_method,omitempty" example:"standard"`
	CouponCode      *string                `json:"coupon_code,omitempty" example:"WELCOME10"`
	DiscountCode    *string                `json:"discount_code,omitempty" example:"DISC10"`
	ExpiresAt       *string                `json:"expires_at,omitempty" example:"2026-02-12T10:00:00Z"`
	Notes           *string                `json:"notes,omitempty" example:"Customer requested gift wrap"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

// UpdateCartRequest represents request body for updating a cart header totals and metadata.
// Decimal fields are accepted as strings to preserve precision, e.g. "12.50".
type UpdateCartRequest struct {
	Subtotal         string                 `json:"subtotal" binding:"required" example:"100.00"`
	DiscountAmount   string                 `json:"discount_amount" binding:"required" example:"10.00"`
	TaxAmount        string                 `json:"tax_amount" binding:"required" example:"5.00"`
	ShippingAmount   string                 `json:"shipping_amount" binding:"required" example:"0.00"`
	TotalAmount      string                 `json:"total_amount" binding:"required" example:"95.00"`
	CouponCode       *string                `json:"coupon_code,omitempty" example:"WELCOME10"`
	DiscountCode     *string                `json:"discount_code,omitempty" example:"DISC10"`
	PromotionalCreds *string                `json:"promotional_credits,omitempty" example:"0.00"`
	ShippingAddress  map[string]interface{} `json:"shipping_address,omitempty" swaggertype:"object"`
	BillingAddress   map[string]interface{} `json:"billing_address,omitempty" swaggertype:"object"`
	ShippingMethod   *string                `json:"shipping_method,omitempty" example:"standard"`
	Notes            *string                `json:"notes,omitempty" example:"Updated by cashier"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

type UpdateCartStatusRequest struct {
	CartStatus         string  `json:"cart_status" binding:"required" example:"active"`
	ConvertedToOrderID *string `json:"converted_to_order_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	ConvertedAtISO8601 *string `json:"converted_at,omitempty" example:"2026-02-10T10:00:00Z"`
}

type UpdateCartCustomerRequest struct {
	CustomerID int32 `json:"customer_id" binding:"required" example:"123"`
}

type AddToCartRequest struct {
	OrganizationID int32   `json:"organization_id" binding:"required" example:"1"`
	ProductID      int32   `json:"product_id" binding:"required" example:"1001"`
	Quantity       float64 `json:"quantity" binding:"required" example:"2"`
}

// CartItemUpsertRequest is for directly creating cart item via SQLC CreateCartItem.
// Decimal fields are strings for precision.
type CreateCartItemRequest struct {
	OrganizationID       int32                  `json:"organization_id" binding:"required" example:"1"`
	ProductID            int32                  `json:"product_id" binding:"required" example:"1001"`
	ProductVariantID     *int32                 `json:"product_variant_id,omitempty" example:"2001"`
	Quantity             string                 `json:"quantity" binding:"required" example:"2.00"`
	UomID                *int32                 `json:"uom_id,omitempty" example:"1"`
	UnitPrice            string                 `json:"unit_price" binding:"required" example:"50.00"`
	DiscountAmount       *string                `json:"discount_amount,omitempty" example:"0.00"`
	TaxAmount            *string                `json:"tax_amount,omitempty" example:"0.00"`
	LineTotal            string                 `json:"line_total" binding:"required" example:"100.00"`
	PriceListID          *int32                 `json:"price_list_id,omitempty" example:"1"`
	TaxCategoryID        *int32                 `json:"tax_category_id,omitempty" example:"1"`
	BatchNumber          *string                `json:"batch_number,omitempty" example:"BATCH-001"`
	SerialNumber         *string                `json:"serial_number,omitempty" example:"SN-001"`
	CustomizationDetails map[string]interface{} `json:"customization_details,omitempty" swaggertype:"object"`
	Notes                *string                `json:"notes,omitempty" example:"No onions"`
	Metadata             map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

type UpdateCartItemRequest struct {
	Quantity       string                 `json:"quantity" binding:"required" example:"2.00"`
	UnitPrice      string                 `json:"unit_price" binding:"required" example:"50.00"`
	DiscountAmount string                 `json:"discount_amount" binding:"required" example:"0.00"`
	TaxAmount      string                 `json:"tax_amount" binding:"required" example:"0.00"`
	LineTotal      string                 `json:"line_total" binding:"required" example:"100.00"`
	Notes          *string                `json:"notes,omitempty" example:"updated"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

type UpdateCartItemQuantityRequest struct {
	DeltaQuantity string `json:"delta_quantity" binding:"required" example:"1.00"`
}

type CreateCartActivityRequest struct {
	OrganizationID    int32                  `json:"organization_id" binding:"required" example:"1"`
	ActivityType      string                 `json:"activity_type" binding:"required" example:"item_added"`
	Description       *string                `json:"description,omitempty" example:"Added product 1001"`
	PerformedByUserID *int32                 `json:"performed_by_user_id,omitempty" example:"1"`
	IpAddress         *string                `json:"ip_address,omitempty" example:"127.0.0.1"`
	UserAgent         *string                `json:"user_agent,omitempty" example:"Mozilla/5.0"`
	OldValue          map[string]interface{} `json:"old_value,omitempty" swaggertype:"object"`
	NewValue          map[string]interface{} `json:"new_value,omitempty" swaggertype:"object"`
}

type ApplyCouponRequest struct {
	CouponCode     string `json:"coupon_code" binding:"required" example:"WELCOME10"`
	DiscountAmount string `json:"discount_amount" binding:"required" example:"10.00"`
}

type MergeGuestCartRequest struct {
	TargetCartID string `json:"target_cart_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// Order DTOs
type CreateSalesOrderV2Request struct {
	OrderNumber          string                 `json:"order_number" binding:"required" example:"ORD-20260210-0001"`
	OrganizationID       int32                  `json:"organization_id" binding:"required" example:"1"`
	StoreID              *int32                 `json:"store_id,omitempty" example:"10"`
	CustomerID           *int32                 `json:"customer_id,omitempty" example:"123"`
	CustomerName         *string                `json:"customer_name,omitempty" example:"John Doe"`
	CustomerEmail        *string                `json:"customer_email,omitempty" example:"john@example.com"`
	CustomerPhone        *string                `json:"customer_phone,omitempty" example:"+15551234567"`
	OrderType            string                 `json:"order_type" binding:"required" example:"standard"`
	OrderStatus          string                 `json:"order_status" binding:"required" example:"pending"`
	PaymentStatus        string                 `json:"payment_status" binding:"required" example:"unpaid"`
	FulfillmentStatus    string                 `json:"fulfillment_status" binding:"required" example:"unfulfilled"`
	SalesChannel         *string                `json:"sales_channel,omitempty" example:"web"`
	OrderSource          *string                `json:"order_source,omitempty" example:"checkout"`
	ReferralSource       *string                `json:"referral_source,omitempty" example:"campaign-1"`
	SourceCartID         *string                `json:"source_cart_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreatedByUserID      *int32                 `json:"created_by_user_id,omitempty" example:"1"`
	AssignedToUserID     *int32                 `json:"assigned_to_user_id,omitempty" example:"2"`
	OrderDate            *string                `json:"order_date,omitempty" example:"2026-02-10T10:00:00Z"`
	ExpectedDeliveryDate *string                `json:"expected_delivery_date,omitempty" example:"2026-02-12"`
	ShippingAddress      map[string]interface{} `json:"shipping_address,omitempty" swaggertype:"object"`
	BillingAddress       map[string]interface{} `json:"billing_address,omitempty" swaggertype:"object"`
	ShippingMethod       *string                `json:"shipping_method,omitempty" example:"standard"`
	PaymentMethod        *string                `json:"payment_method,omitempty" example:"cash"`
	PaymentTerms         *string                `json:"payment_terms,omitempty" example:"net_30"`
	PaymentDueDate       *string                `json:"payment_due_date,omitempty" example:"2026-03-11"`
	PosTerminalID        *int32                 `json:"pos_terminal_id,omitempty" example:"1"`
	CashierID            *int32                 `json:"cashier_id,omitempty" example:"1"`
	IsGift               *bool                  `json:"is_gift,omitempty" example:"false"`
	GiftMessage          *string                `json:"gift_message,omitempty" example:"Happy Birthday"`
	SpecialInstructions  *string                `json:"special_instructions,omitempty" example:"Leave at door"`
	InternalNotes        *string                `json:"internal_notes,omitempty" example:"VIP"`
	Tags                 []string               `json:"tags,omitempty"`
	Priority             *string                `json:"priority,omitempty" example:"normal"`
	Metadata             map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

type UpdateSalesOrderV2Request struct {
	CustomerID           *int32                 `json:"customer_id,omitempty" example:"123"`
	CustomerName         *string                `json:"customer_name,omitempty" example:"John Doe"`
	CustomerEmail        *string                `json:"customer_email,omitempty" example:"john@example.com"`
	CustomerPhone        *string                `json:"customer_phone,omitempty" example:"+15551234567"`
	ExpectedDeliveryDate *string                `json:"expected_delivery_date,omitempty" example:"2026-02-12"`
	ShippingAddress      map[string]interface{} `json:"shipping_address,omitempty" swaggertype:"object"`
	BillingAddress       map[string]interface{} `json:"billing_address,omitempty" swaggertype:"object"`
	ShippingMethod       *string                `json:"shipping_method,omitempty" example:"standard"`
	PaymentMethod        *string                `json:"payment_method,omitempty" example:"cash"`
	SpecialInstructions  *string                `json:"special_instructions,omitempty"`
	InternalNotes        *string                `json:"internal_notes,omitempty"`
	Tags                 []string               `json:"tags,omitempty"`
	Priority             *string                `json:"priority,omitempty" example:"normal"`
	Metadata             map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

type UpdateOrderStatusRequest struct {
	OrderStatus string `json:"order_status" binding:"required" example:"confirmed"`
}

type UpdateOrderPaymentStatusRequest struct {
	PaymentStatus string `json:"payment_status" binding:"required" example:"paid"`
	PaidAmount    string `json:"paid_amount" binding:"required" example:"95.00"`
}

type UpdateOrderFulfillmentStatusRequest struct {
	FulfillmentStatus string `json:"fulfillment_status" binding:"required" example:"fulfilled"`
}

type UpdateOrderTotalsRequest struct {
	Subtotal         string `json:"subtotal" binding:"required" example:"100.00"`
	DiscountAmount   string `json:"discount_amount" binding:"required" example:"10.00"`
	TaxAmount        string `json:"tax_amount" binding:"required" example:"5.00"`
	ShippingAmount   string `json:"shipping_amount" binding:"required" example:"0.00"`
	AdjustmentAmount string `json:"adjustment_amount" binding:"required" example:"0.00"`
	TotalAmount      string `json:"total_amount" binding:"required" example:"95.00"`
}

type UpdateOrderDeliveryRequest struct {
	ShippingCarrier    *string `json:"shipping_carrier,omitempty" example:"DHL"`
	TrackingNumber     *string `json:"tracking_number,omitempty" example:"TRACK123"`
	TrackingUrl        *string `json:"tracking_url,omitempty" example:"https://tracking.example.com/TRACK123"`
	ActualDeliveryDate *string `json:"actual_delivery_date,omitempty" example:"2026-02-15"`
}

type AssignOrderRequest struct {
	AssignedToUserID int32 `json:"assigned_to_user_id" binding:"required" example:"2"`
}

type CreateSalesOrderLineV2Request struct {
	OrganizationID     int32                  `json:"organization_id" binding:"required" example:"1"`
	LineNumber         int32                  `json:"line_number" binding:"required" example:"1"`
	ProductID          int32                  `json:"product_id" binding:"required" example:"1001"`
	ProductVariantID   *int32                 `json:"product_variant_id,omitempty" example:"2001"`
	ProductName        string                 `json:"product_name" binding:"required" example:"Burger"`
	ProductSku         *string                `json:"product_sku,omitempty" example:"SKU-001"`
	QuantityOrdered    string                 `json:"quantity_ordered" binding:"required" example:"2.00"`
	UomID              *int32                 `json:"uom_id,omitempty" example:"1"`
	UnitPrice          string                 `json:"unit_price" binding:"required" example:"50.00"`
	DiscountAmount     *string                `json:"discount_amount,omitempty" example:"0.00"`
	DiscountPercentage *string                `json:"discount_percentage,omitempty" example:"0.00"`
	TaxAmount          *string                `json:"tax_amount,omitempty" example:"0.00"`
	LineTotal          string                 `json:"line_total" binding:"required" example:"100.00"`
	TaxCategoryID      *int32                 `json:"tax_category_id,omitempty"`
	TaxRate            *string                `json:"tax_rate,omitempty" example:"0.00"`
	BatchNumber        *string                `json:"batch_number,omitempty"`
	SerialNumbers      []string               `json:"serial_numbers,omitempty"`
	ExpiryDate         *string                `json:"expiry_date,omitempty" example:"2026-12-31"`
	LineStatus         *string                `json:"line_status,omitempty" example:"pending"`
	Customization      map[string]interface{} `json:"customization_details,omitempty" swaggertype:"object"`
	UnitCost           *string                `json:"unit_cost,omitempty" example:"0.00"`
	Notes              *string                `json:"notes,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

type UpdateSalesOrderLineV2Request struct {
	QuantityOrdered    string                 `json:"quantity_ordered" binding:"required" example:"2.00"`
	UnitPrice          string                 `json:"unit_price" binding:"required" example:"50.00"`
	DiscountAmount     string                 `json:"discount_amount" binding:"required" example:"0.00"`
	DiscountPercentage string                 `json:"discount_percentage" binding:"required" example:"0.00"`
	TaxAmount          string                 `json:"tax_amount" binding:"required" example:"0.00"`
	LineTotal          string                 `json:"line_total" binding:"required" example:"100.00"`
	Notes              *string                `json:"notes,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

type UpdateOrderLineFulfillmentRequest struct {
	QuantityFulfilled string `json:"quantity_fulfilled" binding:"required" example:"1.00"`
}

type UpdateOrderLineStatusRequest struct {
	LineStatus string `json:"line_status" binding:"required" example:"fulfilled"`
}

type CreateOrderStatusHistoryRequest struct {
	OrganizationID  int32   `json:"organization_id" binding:"required" example:"1"`
	FromStatus      *string `json:"from_status,omitempty" example:"pending"`
	ToStatus        string  `json:"to_status" binding:"required" example:"confirmed"`
	Reason          *string `json:"reason,omitempty" example:"customer_confirmed"`
	Notes           *string `json:"notes,omitempty"`
	ChangedByUserID *int32  `json:"changed_by_user_id,omitempty" example:"1"`
}

type CreateOrderFulfillmentRequest struct {
	OrganizationID        int32                  `json:"organization_id" binding:"required" example:"1"`
	FulfillmentNumber     string                 `json:"fulfillment_number" binding:"required" example:"FUL-0001"`
	FulfillmentStatus     *string                `json:"fulfillment_status,omitempty" example:"created"`
	ShipmentStatus        *string                `json:"shipment_status,omitempty" example:"pending"`
	FulfillmentStoreID    *int32                 `json:"fulfillment_store_id,omitempty" example:"10"`
	ShippingCarrier       *string                `json:"shipping_carrier,omitempty" example:"DHL"`
	ShippingMethod        *string                `json:"shipping_method,omitempty" example:"standard"`
	TrackingNumber        *string                `json:"tracking_number,omitempty" example:"TRACK123"`
	TrackingUrl           *string                `json:"tracking_url,omitempty" example:"https://tracking.example.com/TRACK123"`
	EstimatedDeliveryDate *string                `json:"estimated_delivery_date,omitempty" example:"2026-02-15"`
	Notes                 *string                `json:"notes,omitempty"`
	Metadata              map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

type UpdateOrderFulfillmentRequest struct {
	FulfillmentStatus     *string                `json:"fulfillment_status,omitempty"`
	ShipmentStatus        *string                `json:"shipment_status,omitempty"`
	ShippingCarrier       *string                `json:"shipping_carrier,omitempty"`
	TrackingNumber        *string                `json:"tracking_number,omitempty"`
	TrackingUrl           *string                `json:"tracking_url,omitempty"`
	EstimatedDeliveryDate *string                `json:"estimated_delivery_date,omitempty" example:"2026-02-15"`
	ActualDeliveryDate    *string                `json:"actual_delivery_date,omitempty" example:"2026-02-16"`
	Notes                 *string                `json:"notes,omitempty"`
	Metadata              map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

type UpdateFulfillmentShipmentRequest struct {
	ShipmentStatus string `json:"shipment_status" binding:"required" example:"shipped"`
}

type UpdateFulfillmentPickPackRequest struct {
	PickedAt       *string `json:"picked_at,omitempty" example:"2026-02-10T10:00:00Z"`
	PackedAt       *string `json:"packed_at,omitempty" example:"2026-02-10T12:00:00Z"`
	PickedByUserID *int32  `json:"picked_by_user_id,omitempty" example:"1"`
	PackedByUserID *int32  `json:"packed_by_user_id,omitempty" example:"2"`
}

type CreateOrderFulfillmentItemRequest struct {
	OrderLineID       string   `json:"order_line_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	OrganizationID    int32    `json:"organization_id" binding:"required" example:"1"`
	QuantityFulfilled string   `json:"quantity_fulfilled" binding:"required" example:"1.00"`
	BatchNumber       *string  `json:"batch_number,omitempty"`
	SerialNumbers     []string `json:"serial_numbers,omitempty"`
}

// =====================================================
// Cashier Session Module
// =====================================================

type OpenCashierSessionRequest struct {
	CashierID      int32  `json:"cashier_id" binding:"required" example:"1"`
	PosTerminalID  int32  `json:"pos_terminal_id" binding:"required" example:"1"`
	SessionNumber  string `json:"session_number" binding:"required" example:"SES-20261026-001"`
	OpeningBalance string `json:"opening_balance" binding:"required" example:"100.00"`
}

type CloseCashierSessionRequest struct {
	ClosingBalance  string `json:"closing_balance" binding:"required" example:"500.00"`
	ExpectedBalance string `json:"expected_balance" binding:"required" example:"500.00"`
	Variance        string `json:"variance" binding:"required" example:"0.00"`
	ClosingNote     string `json:"closing_note,omitempty" example:"All good"`
	ClosedBy        int64  `json:"closed_by" binding:"required" example:"1"`
}

// =====================================================
// Product Barcode Module
// =====================================================

type CreateProductBarcodeRequest struct {
	ProductID        int32                  `json:"product_id" binding:"required" example:"1"`
	ProductVariantID *int32                 `json:"product_variant_id,omitempty" example:"1"`
	Barcode          string                 `json:"barcode" binding:"required" example:"1234567890123"`
	BarcodeType      *string                `json:"barcode_type,omitempty" example:"EAN13"`
	IsPrimary        *bool                  `json:"is_primary,omitempty" example:"true"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

type UpdateProductBarcodeRequest struct {
	BarcodeType *string                `json:"barcode_type,omitempty" example:"UPC"`
	IsPrimary   *bool                  `json:"is_primary,omitempty" example:"false"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

type SetPrimaryBarcodeRequest struct {
	BarcodeID int32 `json:"barcode_id" binding:"required" example:"1"`
}

// =====================================================
// Brand Module
// =====================================================

// CreateBrandRequest represents request body for creating a brand
type CreateBrandRequest struct {
	Name     string                 `json:"name" binding:"required" example:"Nike"`
	Code     string                 `json:"code" binding:"required" example:"NIKE"`
	IsActive bool                   `json:"is_active" example:"true"`
	Metadata map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

// CreateBrandWithDefaultsRequest represents request body for creating a brand with defaults
type CreateBrandWithDefaultsRequest struct {
	Name string `json:"name" binding:"required" example:"Nike"`
	Code string `json:"code" binding:"required" example:"NIKE"`
}

// UpdateBrandRequest represents request body for updating a brand
type UpdateBrandRequest struct {
	Name     *string                `json:"name,omitempty" example:"Nike Updated"`
	Code     *string                `json:"code,omitempty" example:"NIKE_V2"`
	IsActive *bool                  `json:"is_active,omitempty" example:"true"`
	Metadata map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

// UpdateBrandNameRequest represents request body for updating brand name
type UpdateBrandNameRequest struct {
	Name string `json:"name" binding:"required" example:"Nike Updated"`
}

// UpdateBrandCodeRequest represents request body for updating brand code
type UpdateBrandCodeRequest struct {
	Code string `json:"code" binding:"required" example:"NIKE_V2"`
}

// UpdateBrandMetadataRequest represents request body for updating brand metadata
type UpdateBrandMetadataRequest struct {
	Metadata map[string]interface{} `json:"metadata" binding:"required" swaggertype:"object"`
}

// BulkBrandIDsRequest represents request body for bulk brand operations
type BulkBrandIDsRequest struct {
	IDs []int32 `json:"ids" binding:"required" example:"1,2,3"`
}

// BrandResponse represents a brand in API responses
type BrandResponse struct {
	ID        int32           `json:"id" example:"1"`
	Name      string          `json:"name" example:"Nike"`
	Code      string          `json:"code" example:"NIKE"`
	IsActive  bool            `json:"is_active" example:"true"`
	Metadata  json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
	CreatedAt string          `json:"created_at" example:"2026-01-24T21:43:00Z"`
	UpdatedAt string          `json:"updated_at" example:"2026-01-24T21:43:00Z"`
}

// BrandWithProductCountResponse represents a brand with product count
type BrandWithProductCountResponse struct {
	ID           int32  `json:"id" example:"1"`
	Name         string `json:"name" example:"Nike"`
	Code         string `json:"code" example:"NIKE"`
	IsActive     bool   `json:"is_active" example:"true"`
	ProductCount int64  `json:"product_count" example:"150"`
	CreatedAt    string `json:"created_at" example:"2026-01-24T21:43:00Z"`
	UpdatedAt    string `json:"updated_at" example:"2026-01-24T21:43:00Z"`
}

// BrandWithStatsResponse represents a brand with statistics
type BrandWithStatsResponse struct {
	ID            int32  `json:"id" example:"1"`
	Name          string `json:"name" example:"Nike"`
	Code          string `json:"code" example:"NIKE"`
	IsActive      bool   `json:"is_active" example:"true"`
	ProductCount  int64  `json:"product_count" example:"150"`
	CategoryCount int64  `json:"category_count" example:"5"`
	CreatedAt     string `json:"created_at" example:"2026-01-24T21:43:00Z"`
	UpdatedAt     string `json:"updated_at" example:"2026-01-24T21:43:00Z"`
}

// CountResponse represents a count response
type CountResponse struct {
	Count int64 `json:"count" example:"100"`
}

// ExistsResponse represents an existence check response
type ExistsResponse struct {
	Exists bool `json:"exists" example:"true"`
}

// MetadataResponse represents a metadata response
type MetadataResponse struct {
	ID            int32  `json:"id" example:"1"`
	Name          string `json:"name" example:"Nike"`
	Code          string `json:"code" example:"NIKE"`
	MetadataValue string `json:"metadata_value" example:"{\"country\":\"USA\"}"`
}

// =====================================================
// Cashier Module
// =====================================================

// CreateCashierRequest represents request body for creating a cashier
type CreateCashierRequest struct {
	UserID        int32                  `json:"user_id" binding:"required" example:"10"`
	StoreID       int32                  `json:"store_id" binding:"required" example:"5"`
	CashierCode   string                 `json:"cashier_code" binding:"required" example:"CASH001"`
	DrawerLimit   *string                `json:"drawer_limit,omitempty" example:"1000.00"`
	DiscountLimit *string                `json:"discount_limit,omitempty" example:"50.00"`
	IsActive      bool                   `json:"is_active" example:"true"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

// CreateCashierWithDefaultsRequest represents request body for creating a cashier with defaults
type CreateCashierWithDefaultsRequest struct {
	UserID      int32  `json:"user_id" binding:"required" example:"10"`
	StoreID     int32  `json:"store_id" binding:"required" example:"5"`
	CashierCode string `json:"cashier_code" binding:"required" example:"CASH001"`
}

// UpdateCashierRequest represents request body for updating a cashier
type UpdateCashierRequest struct {
	UserID        *int32                 `json:"user_id,omitempty" example:"10"`
	StoreID       *int32                 `json:"store_id,omitempty" example:"5"`
	CashierCode   *string                `json:"cashier_code,omitempty" example:"CASH002"`
	DrawerLimit   *string                `json:"drawer_limit,omitempty" example:"1500.00"`
	DiscountLimit *string                `json:"discount_limit,omitempty" example:"75.00"`
	IsActive      *bool                  `json:"is_active,omitempty" example:"true"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

// UpdateCashierLimitsRequest represents request body for updating cashier limits
type UpdateCashierLimitsRequest struct {
	DrawerLimit   string `json:"drawer_limit" binding:"required" example:"1000.00"`
	DiscountLimit string `json:"discount_limit" binding:"required" example:"50.00"`
}

// UpdateCashierDrawerLimitRequest represents request body for updating drawer limit
type UpdateCashierDrawerLimitRequest struct {
	DrawerLimit string `json:"drawer_limit" binding:"required" example:"1000.00"`
}

// UpdateCashierDiscountLimitRequest represents request body for updating discount limit
type UpdateCashierDiscountLimitRequest struct {
	DiscountLimit string `json:"discount_limit" binding:"required" example:"50.00"`
}

// UpdateCashierMetadataRequest represents request body for updating cashier metadata
type UpdateCashierMetadataRequest struct {
	Metadata map[string]interface{} `json:"metadata" binding:"required" swaggertype:"object"`
}

// CashierResponse represents a cashier in API responses
type CashierResponse struct {
	ID            int32           `json:"id" example:"1"`
	UserID        int32           `json:"user_id" example:"10"`
	StoreID       int32           `json:"store_id" example:"5"`
	CashierCode   string          `json:"cashier_code" example:"CASH001"`
	DrawerLimit   string          `json:"drawer_limit,omitempty" example:"1000.00"`
	DiscountLimit string          `json:"discount_limit,omitempty" example:"50.00"`
	IsActive      bool            `json:"is_active" example:"true"`
	Metadata      json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
	CreatedAt     string          `json:"created_at" example:"2026-01-24T21:43:00Z"`
}

// CashierWithLimitsResponse represents a cashier with limits and user details
type CashierWithLimitsResponse struct {
	ID            int32  `json:"id" example:"1"`
	CashierCode   string `json:"cashier_code" example:"CASH001"`
	DrawerLimit   string `json:"drawer_limit,omitempty" example:"1000.00"`
	DiscountLimit string `json:"discount_limit,omitempty" example:"50.00"`
	IsActive      bool   `json:"is_active" example:"true"`
	FirstName     string `json:"first_name" example:"John"`
	LastName      string `json:"last_name" example:"Doe"`
	Email         string `json:"email" example:"john.doe@example.com"`
	StoreName     string `json:"store_name" example:"Main Store"`
}

// CashierWithSessionsResponse represents a cashier with session information
type CashierWithSessionsResponse struct {
	ID             int32  `json:"id" example:"1"`
	CashierCode    string `json:"cashier_code" example:"CASH001"`
	FullName       string `json:"full_name" example:"John Doe"`
	DiscountLimit  string `json:"discount_limit,omitempty" example:"50.00"`
	DrawerLimit    string `json:"drawer_limit,omitempty" example:"1000.00"`
	ActiveSessions int64  `json:"active_sessions" example:"1"`
}
