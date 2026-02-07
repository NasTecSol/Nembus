package handler

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
	OrganizationID      int32  `json:"organization_id" binding:"required"`
	RecipeCode          string `json:"recipe_code" binding:"required"`
	RecipeName          string `json:"recipe_name" binding:"required"`
	Description         string `json:"description"`
	FinishedProductID   *int32 `json:"finished_product_id"`
	YieldQuantity       string `json:"yield_quantity"`
	YieldUomID          *int32 `json:"yield_uom_id"`
	PreparationSteps    string `json:"preparation_steps"`
	PreparationTimeMin  int32  `json:"preparation_time_min"`
	CookingTimeMin      int32  `json:"cooking_time_min"`
	IsActive            bool   `json:"is_active"`
	Metadata            string `json:"metadata"`
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
	StoreID    int32                                `json:"store_id" binding:"required"`
	CustomerID *int32                               `json:"customer_id"`
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
