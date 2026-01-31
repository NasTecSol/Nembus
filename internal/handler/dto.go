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

// AssignPermissionToRoleRequest represents permission assignment to a role
type AssignPermissionToRoleRequest struct {
	PermissionID int32   `json:"permission_id" binding:"required" example:"1"`
	Scope        *string `json:"scope,omitempty" example:"read,write"`
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
