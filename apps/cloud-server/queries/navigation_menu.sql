-- =====================================================
-- COMPLETE NAVIGATION & UI STRUCTURE QUERIES
-- =====================================================
-- These queries fetch all modules, menus, submenus with permissions,
-- UI settings, and role customizations for logged-in users
-- =====================================================

-- =====================================================
-- QUERY 1: GET COMPLETE NAVIGATION BY USER ID
-- =====================================================
-- Returns the complete navigation structure for a specific user
-- including modules, menus, submenus with permissions and UI settings
-- =====================================================

-- name: GetCompleteNavigationByUserID :many
WITH user_permissions AS (
    -- Get all permissions for the user through their roles
    SELECT DISTINCT p.id, p.code, p.name
    FROM user_roles ur
    JOIN role_permissions rp ON ur.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.id
    WHERE ur.user_id = sqlc.arg(user_id)
),
user_roles_list AS (
    -- Get all roles for the user
    SELECT r.id, r.code, r.name
    FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id
    WHERE ur.user_id = sqlc.arg(user_id)
)
SELECT 
    -- Module Information
    mod.id as module_id,
    mod.name as module_name,
    mod.code as module_code,
    mod.description as module_description,
    mod.icon as module_icon,
    mod.display_order as module_order,
    mod.is_active as module_is_active,
    
    -- Aggregated Menus with Submenus
    (
        SELECT jsonb_agg(
            jsonb_build_object(
                'menu_id', menu_data.id,
                'menu_name', menu_data.name,
                'menu_code', menu_data.code,
                'menu_route', menu_data.route_path,
                'menu_icon', menu_data.icon,
                'menu_order', menu_data.display_order,
                'menu_is_active', menu_data.is_active,
                
                -- Menu Permissions
                'menu_permissions', menu_data.menu_permissions,
                
                -- Submenus Array
                'submenus', menu_data.submenus
            )
            ORDER BY menu_data.display_order
        )
        FROM (
            SELECT DISTINCT
                m.id,
                m.name,
                m.code,
                m.route_path,
                m.icon,
                m.display_order,
                m.is_active,
                
                -- Menu Permissions
                (
                    SELECT jsonb_agg(DISTINCT p.code)
                    FROM menu_permissions mp
                    JOIN permissions p ON mp.permission_id = p.id
                    WHERE mp.menu_id = m.id
                ) as menu_permissions,
                
                -- Submenus
                COALESCE((
                    SELECT jsonb_agg(
                        jsonb_build_object(
                            'submenu_id', s.id,
                            'submenu_name', s.name,
                            'submenu_code', s.code,
                            'submenu_route', s.route_path,
                            'submenu_icon', s.icon,
                            'submenu_order', s.display_order,
                            'submenu_is_active', s.is_active,
                            
                            -- Submenu Permissions
                            'required_permissions', (
                                SELECT jsonb_agg(DISTINCT p.code)
                                FROM submenu_permissions sp
                                JOIN permissions p ON sp.permission_id = p.id
                                WHERE sp.submenu_id = s.id
                            ),
                            
                            -- Check if user has access to this submenu
                            'has_access', EXISTS(
                                SELECT 1 
                                FROM submenu_permissions sp
                                WHERE sp.submenu_id = s.id 
                                    AND sp.permission_id IN (SELECT id FROM user_permissions)
                            ),
                            
                            -- UI Settings (default + role-specific customizations)
                            'ui_settings', COALESCE(
                                -- First try to get role-specific customizations
                                (
                                    SELECT ruc.customization_data
                                    FROM role_ui_customizations ruc
                                    WHERE ruc.submenu_id = s.id 
                                        AND ruc.role_id IN (SELECT id FROM user_roles_list)
                                    ORDER BY ruc.created_at DESC
                                    LIMIT 1
                                ),
                                -- If no role customization, get default UI settings
                                (
                                    SELECT jsonb_object_agg(ui.setting_key, ui.setting_value)
                                    FROM ui_settings ui
                                    WHERE ui.submenu_id = s.id
                                ),
                                -- Default empty object if no settings
                                '{}'::jsonb
                            ),
                            
                            -- Default UI Settings (always include for reference)
                            'default_ui_settings', COALESCE(
                                (
                                    SELECT jsonb_object_agg(ui.setting_key, ui.setting_value)
                                    FROM ui_settings ui
                                    WHERE ui.submenu_id = s.id
                                ),
                                '{}'::jsonb
                            ),
                            
                            -- Role Customizations (if any)
                            'role_customizations', COALESCE(
                                (
                                    SELECT jsonb_agg(
                                        jsonb_build_object(
                                            'role_code', r.code,
                                            'role_name', r.name,
                                            'customization_data', ruc.customization_data
                                        )
                                    )
                                    FROM role_ui_customizations ruc
                                    JOIN roles r ON ruc.role_id = r.id
                                    WHERE ruc.submenu_id = s.id 
                                        AND ruc.role_id IN (SELECT id FROM user_roles_list)
                                ),
                                '[]'::jsonb
                            )
                        )
                        ORDER BY s.display_order
                    )
                    FROM submenus s
                    WHERE s.menu_id = m.id 
                        AND s.is_active = true
                        -- Only include submenus the user has access to
                        AND EXISTS(
                            SELECT 1 
                            FROM submenu_permissions sp
                            WHERE sp.submenu_id = s.id 
                                AND sp.permission_id IN (SELECT id FROM user_permissions)
                        )
                ), '[]'::jsonb) as submenus
            FROM menus m
            WHERE m.module_id = mod.id
                AND m.is_active = true
                AND EXISTS(
                    SELECT 1 
                    FROM submenus s2
                    JOIN submenu_permissions sp2 ON s2.id = sp2.submenu_id
                    WHERE s2.menu_id = m.id
                        AND sp2.permission_id IN (SELECT id FROM user_permissions)
                )
        ) menu_data
    ) as menus

FROM modules mod
WHERE mod.is_active = true
    -- Only include modules where user has at least one permission
    AND EXISTS(
        SELECT 1 
        FROM menus m2
        JOIN submenus s2 ON m2.id = s2.menu_id
        JOIN submenu_permissions sp2 ON s2.id = sp2.submenu_id
        WHERE m2.module_id = mod.id
            AND sp2.permission_id IN (SELECT id FROM user_permissions)
    )
ORDER BY mod.display_order;


-- =====================================================
-- QUERY 2: GET COMPLETE NAVIGATION BY ROLE CODE
-- =====================================================
-- Returns navigation structure for a specific role
-- Useful for role management and preview
-- =====================================================

-- name: GetCompleteNavigationByRoleCode :many
WITH role_permissions AS (
    -- Get all permissions for the role
    SELECT DISTINCT p.id, p.code, p.name
    FROM role_permissions rp
    JOIN permissions p ON rp.permission_id = p.id
    JOIN roles r ON rp.role_id = r.id
    WHERE r.code = sqlc.arg(role_code)
)
SELECT 
    -- Module Information
    mod.id as module_id,
    mod.name as module_name,
    mod.code as module_code,
    mod.description as module_description,
    mod.icon as module_icon,
    mod.display_order as module_order,
    
    -- Aggregated Menus
    (
        SELECT jsonb_agg(
            jsonb_build_object(
                'menu_id', menu_data.id,
                'menu_name', menu_data.name,
                'menu_code', menu_data.code,
                'menu_route', menu_data.route_path,
                'menu_icon', menu_data.icon,
                'menu_order', menu_data.display_order,
                'submenus', menu_data.submenus
            )
            ORDER BY menu_data.display_order
        )
        FROM (
            SELECT DISTINCT
                m.id,
                m.name,
                m.code,
                m.route_path,
                m.icon,
                m.display_order,
                
                -- Submenus
                COALESCE((
                    SELECT jsonb_agg(
                        jsonb_build_object(
                            'submenu_id', s.id,
                            'submenu_name', s.name,
                            'submenu_code', s.code,
                            'submenu_route', s.route_path,
                            'submenu_icon', s.icon,
                            'submenu_order', s.display_order,
                            'required_permissions', (
                                SELECT jsonb_agg(DISTINCT p.code)
                                FROM submenu_permissions sp
                                JOIN permissions p ON sp.permission_id = p.id
                                WHERE sp.submenu_id = s.id
                            ),
                            'has_access', EXISTS(
                                SELECT 1 
                                FROM submenu_permissions sp
                                WHERE sp.submenu_id = s.id 
                                    AND sp.permission_id IN (SELECT id FROM role_permissions)
                            ),
                            'ui_settings', COALESCE(
                                (
                                    SELECT ruc.customization_data
                                    FROM role_ui_customizations ruc
                                    JOIN roles r ON ruc.role_id = r.id
                                    WHERE ruc.submenu_id = s.id 
                                        AND r.code = sqlc.arg(role_code)
                                    LIMIT 1
                                ),
                                (
                                    SELECT jsonb_object_agg(ui.setting_key, ui.setting_value)
                                    FROM ui_settings ui
                                    WHERE ui.submenu_id = s.id
                                ),
                                '{}'::jsonb
                            )
                        )
                        ORDER BY s.display_order
                    )
                    FROM submenus s
                    WHERE s.menu_id = m.id 
                        AND s.is_active = true
                        AND EXISTS(
                            SELECT 1 
                            FROM submenu_permissions sp
                            WHERE sp.submenu_id = s.id 
                                AND sp.permission_id IN (SELECT id FROM role_permissions)
                        )
                ), '[]'::jsonb) as submenus
            FROM menus m
            WHERE m.module_id = mod.id
                AND m.is_active = true
                AND EXISTS(
                    SELECT 1 
                    FROM submenus s2
                    JOIN submenu_permissions sp2 ON s2.id = sp2.submenu_id
                    WHERE s2.menu_id = m.id
                        AND sp2.permission_id IN (SELECT id FROM role_permissions)
                )
        ) menu_data
    ) as menus

FROM modules mod
WHERE mod.is_active = true
    AND EXISTS(
        SELECT 1 
        FROM menus m2
        JOIN submenus s2 ON m2.id = s2.menu_id
        JOIN submenu_permissions sp2 ON s2.id = sp2.submenu_id
        WHERE m2.module_id = mod.id
            AND sp2.permission_id IN (SELECT id FROM role_permissions)
    )
ORDER BY mod.display_order;


-- =====================================================
-- QUERY 3: GET USER PROFILE WITH ROLES AND PERMISSIONS
-- =====================================================
-- Returns complete user information including all roles,
-- permissions, and store access
-- =====================================================

-- name: GetUserProfileWithRolesAndPermissions :one
WITH user_data AS (
    SELECT 
        u.id as user_id,
        u.username,
        u.email,
        u.first_name,
        u.last_name,
        u.employee_code,
        u.is_active,
        u.organization_id,
        o.name as organization_name,
        o.code as organization_code,
        u.created_at,
        u.updated_at
    FROM users u
    JOIN organizations o ON u.organization_id = o.id
    WHERE u.id = sqlc.arg(user_id)
)
SELECT 
    -- User Basic Information
    ud.*,
    
    -- User Roles
    (
        SELECT jsonb_agg(
            jsonb_build_object(
                'role_id', r.id,
                'role_code', r.code,
                'role_name', r.name,
                'role_description', r.description,
                'is_system_role', r.is_system_role,
                'assigned_at', ur.assigned_at
            )
            ORDER BY ur.assigned_at
        )
        FROM user_roles ur
        JOIN roles r ON ur.role_id = r.id
        WHERE ur.user_id = ud.user_id
    ) as roles,
    
    -- User Permissions (aggregated from all roles)
    (
        SELECT jsonb_agg(DISTINCT 
            jsonb_build_object(
                'permission_id', p.id,
                'permission_code', p.code,
                'permission_name', p.name,
                'permission_description', p.description
            )
            ORDER BY jsonb_build_object(
                'permission_id', p.id,
                'permission_code', p.code,
                'permission_name', p.name,
                'permission_description', p.description
            )
        )
        FROM user_roles ur
        JOIN role_permissions rp ON ur.role_id = rp.role_id
        JOIN permissions p ON rp.permission_id = p.id
        WHERE ur.user_id = ud.user_id
    ) as permissions,
    
    -- Permission Codes (simple array for quick checking)
    (
        SELECT jsonb_agg(DISTINCT p.code ORDER BY p.code)
        FROM user_roles ur
        JOIN role_permissions rp ON ur.role_id = rp.role_id
        JOIN permissions p ON rp.permission_id = p.id
        WHERE ur.user_id = ud.user_id
    ) as permission_codes,
    
    -- Store Access
    (
        SELECT jsonb_agg(
            jsonb_build_object(
                'store_id', s.id,
                'store_code', s.code,
                'store_name', s.name,
                'store_type', s.store_type,
                'is_warehouse', s.is_warehouse,
                'is_pos_enabled', s.is_pos_enabled,
                'is_primary', usa.is_primary,
                'granted_at', usa.granted_at
            )
            ORDER BY usa.is_primary DESC, s.name
        )
        FROM user_store_access usa
        JOIN stores s ON usa.store_id = s.id
        WHERE usa.user_id = ud.user_id
    ) as store_access,
    
    -- Primary Store
    (
        SELECT jsonb_build_object(
            'store_id', s.id,
            'store_code', s.code,
            'store_name', s.name,
            'store_type', s.store_type
        )
        FROM user_store_access usa
        JOIN stores s ON usa.store_id = s.id
        WHERE usa.user_id = ud.user_id AND usa.is_primary = true
        LIMIT 1
    ) as primary_store,
    
    -- Cashier Information (if user is a cashier)
    (
        SELECT jsonb_build_object(
            'cashier_id', c.id,
            'cashier_code', c.cashier_code,
            'store_id', c.store_id,
            'store_name', s.name,
            'drawer_limit', c.drawer_limit,
            'discount_limit', c.discount_limit,
            'is_active', c.is_active
        )
        FROM cashiers c
        JOIN stores s ON c.store_id = s.id
        WHERE c.user_id = ud.user_id
        LIMIT 1
    ) as cashier_info

FROM user_data ud;


-- =====================================================
-- QUERY 4: SIMPLIFIED NAVIGATION FOR FRONTEND (RECOMMENDED)
-- =====================================================
-- Optimized version that returns a clean, frontend-ready structure
-- This is the RECOMMENDED query for your Angular application
-- =====================================================

-- name: GetSimplifiedNavigationByUserID :many
WITH user_permissions AS (
    SELECT DISTINCT p.id, p.code
    FROM user_roles ur
    JOIN role_permissions rp ON ur.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.id
    WHERE ur.user_id = sqlc.arg(user_id)
),
user_roles_list AS (
    SELECT r.id, r.code
    FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id
    WHERE ur.user_id = sqlc.arg(user_id)
)
SELECT 
    mod.id,
    mod.code,
    mod.name,
    mod.icon,
    mod.display_order as "order",
    
    (
        SELECT jsonb_agg(
            jsonb_build_object(
                'id', m.id,
                'code', m.code,
                'name', m.name,
                'route', m.route_path,
                'icon', m.icon,
                'order', m.display_order,
                'submenus', (
                    SELECT jsonb_agg(
                        jsonb_build_object(
                            'id', s.id,
                            'code', s.code,
                            'name', s.name,
                            'route', s.route_path,
                            'icon', s.icon,
                            'order', s.display_order,
                            'settings', COALESCE(
                                (
                                    SELECT customization_data
                                    FROM role_ui_customizations
                                    WHERE submenu_id = s.id 
                                        AND role_id IN (SELECT id FROM user_roles_list)
                                    LIMIT 1
                                ),
                                (
                                    SELECT jsonb_object_agg(setting_key, setting_value)
                                    FROM ui_settings
                                    WHERE submenu_id = s.id
                                ),
                                '{}'::jsonb
                            )
                        )
                        ORDER BY s.display_order
                    )
                    FROM submenus s
                    WHERE s.menu_id = m.id 
                        AND s.is_active = true
                        AND EXISTS(
                            SELECT 1 FROM submenu_permissions sp
                            WHERE sp.submenu_id = s.id 
                                AND sp.permission_id IN (SELECT id FROM user_permissions)
                        )
                )
            )
            ORDER BY m.display_order
        )
        FROM menus m
        WHERE m.module_id = mod.id 
            AND m.is_active = true
            AND EXISTS(
                SELECT 1 
                FROM submenus s2
                JOIN submenu_permissions sp2 ON s2.id = sp2.submenu_id
                WHERE s2.menu_id = m.id
                    AND sp2.permission_id IN (SELECT id FROM user_permissions)
            )
    ) as menus

FROM modules mod
WHERE mod.is_active = true
    AND EXISTS(
        SELECT 1 
        FROM menus m2
        JOIN submenus s2 ON m2.id = s2.menu_id
        JOIN submenu_permissions sp2 ON s2.id = sp2.submenu_id
        WHERE m2.module_id = mod.id
            AND sp2.permission_id IN (SELECT id FROM user_permissions)
    )
ORDER BY mod.display_order;


-- =====================================================
-- QUERY 5: GET SPECIFIC SUBMENU UI SETTINGS
-- =====================================================
-- Returns UI settings for a specific submenu, including role customizations
-- Useful for lazy-loading submenu configurations
-- =====================================================

-- name: GetSubmenuUISettings :one
WITH user_roles_list AS (
    SELECT r.id, r.code, r.name
    FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id
    WHERE ur.user_id = sqlc.arg(user_id)
)
SELECT 
    s.id as submenu_id,
    s.name as submenu_name,
    s.code as submenu_code,
    s.route_path,
    
    -- Default UI Settings
    COALESCE(
        (
            SELECT jsonb_object_agg(ui.setting_key, ui.setting_value)
            FROM ui_settings ui
            WHERE ui.submenu_id = s.id
        ),
        '{}'::jsonb
    ) as default_settings,
    
    -- Role-specific Customizations (merged with defaults)
    COALESCE(
        (
            SELECT ruc.customization_data
            FROM role_ui_customizations ruc
            WHERE ruc.submenu_id = s.id 
                AND ruc.role_id IN (SELECT id FROM user_roles_list)
            ORDER BY ruc.created_at DESC
            LIMIT 1
        ),
        (
            SELECT jsonb_object_agg(ui.setting_key, ui.setting_value)
            FROM ui_settings ui
            WHERE ui.submenu_id = s.id
        ),
        '{}'::jsonb
    ) as active_settings,
    
    -- All available role customizations
    (
        SELECT jsonb_agg(
            jsonb_build_object(
                'role_code', r.code,
                'role_name', r.name,
                'customization_data', ruc.customization_data
            )
        )
        FROM role_ui_customizations ruc
        JOIN roles r ON ruc.role_id = r.id
        WHERE ruc.submenu_id = s.id 
            AND ruc.role_id IN (SELECT id FROM user_roles_list)
    ) as all_role_customizations

FROM submenus s
WHERE s.code = sqlc.arg(submenu_code)
LIMIT 1;


-- =====================================================
-- QUERY 6: CHECK USER PERMISSION
-- =====================================================
-- Quick check if user has a specific permission
-- Returns boolean result
-- =====================================================

-- name: CheckUserPermission :one
SELECT EXISTS(
    SELECT 1
    FROM user_roles ur
    JOIN role_permissions rp ON ur.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.id
    WHERE ur.user_id = sqlc.arg(user_id)
        AND p.code = sqlc.arg(permission_code)
) as has_permission;


-- =====================================================
-- QUERY 7: GET ALL ACCESSIBLE ROUTES FOR USER
-- =====================================================
-- Returns a flat list of all routes the user can access
-- Useful for route guards in Angular
-- =====================================================

-- name: GetAllAccessibleRoutesByUserID :many
WITH user_permissions AS (
    SELECT DISTINCT p.id
    FROM user_roles ur
    JOIN role_permissions rp ON ur.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.id
    WHERE ur.user_id = sqlc.arg(user_id)
)
SELECT DISTINCT
    s.route_path,
    s.code as submenu_code,
    s.name as submenu_name,
    m.code as menu_code,
    m.name as menu_name,
    mod.code as module_code,
    mod.name as module_name,
    array_agg(DISTINCT p.code) as required_permissions
FROM submenus s
JOIN menus m ON s.menu_id = m.id
JOIN modules mod ON m.module_id = mod.id
JOIN submenu_permissions sp ON s.id = sp.submenu_id
JOIN permissions p ON sp.permission_id = p.id
WHERE s.is_active = true
    AND m.is_active = true
    AND mod.is_active = true
    AND EXISTS(
        SELECT 1 
        FROM submenu_permissions sp2
        WHERE sp2.submenu_id = s.id 
            AND sp2.permission_id IN (SELECT id FROM user_permissions)
    )
GROUP BY s.id, s.route_path, s.code, s.name, m.code, m.name, mod.code, mod.name
ORDER BY mod.display_order, m.display_order, s.display_order;


-- =====================================================
-- EXAMPLE USAGE WITH ACTUAL VALUES
-- =====================================================

-- Example 1: Get navigation for user ID 1 (Super Admin)
-- Replace :user_id with 1

-- Example 2: Get navigation for user ID 3 (Store Manager)
-- Replace :user_id with 3

-- Example 3: Get navigation for role 'cashier'
-- Replace :role_code with 'cashier'

-- Example 4: Get UI settings for 'user_list' submenu for user ID 1
-- Replace :user_id with 1 and :submenu_code with 'user_list'

-- Example 5: Check if user 4 has 'pos:void_transactions' permission
-- Replace :user_id with 4 and :permission_code with 'pos:void_transactions'


-- =====================================================
-- SAMPLE OUTPUT STRUCTURE (JSON FORMAT)
-- =====================================================

/*
QUERY 4 (RECOMMENDED) OUTPUT STRUCTURE:

[
  {
    "id": 1,
    "code": "dashboard",
    "name": "Dashboard",
    "icon": "dashboard",
    "order": 1,
    "menus": [
      {
        "id": 1,
        "code": "overview",
        "name": "Overview",
        "route": "/dashboard/overview",
        "icon": "home",
        "order": 1,
        "submenus": [
          {
            "id": 1,
            "code": "admin_dashboard",
            "name": "Admin Dashboard",
            "route": "/dashboard/admin",
            "icon": "layout",
            "order": 1,
            "settings": {
              "widgets": ["sales_overview", "inventory_status"],
              "layout": "grid",
              "refresh_interval": 30
            }
          }
        ]
      }
    ]
  },
  {
    "id": 6,
    "code": "pos",
    "name": "Point of Sale",
    "icon": "shopping-cart",
    "order": 6,
    "menus": [
      {
        "id": 9,
        "code": "pos_transactions",
        "name": "POS Transactions",
        "route": "/pos/transactions",
        "icon": "credit-card",
        "order": 1,
        "submenus": [
          {
            "id": 21,
            "code": "transaction_list",
            "name": "Transaction List",
            "route": "/pos/transactions/list",
            "icon": "list",
            "order": 1,
            "settings": {
              "table_columns": {
                "columns": ["transaction_number", "cashier_name", "total_amount"]
              }
            }
          }
        ]
      }
    ]
  }
]
*/

-- =====================================================
-- PERFORMANCE OPTIMIZATION INDEXES (If not already created)
-- =====================================================

-- These indexes will significantly improve query performance
-- Only create if they don't already exist

CREATE INDEX IF NOT EXISTS idx_user_roles_composite ON user_roles(user_id, role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_composite ON role_permissions(role_id, permission_id);
CREATE INDEX IF NOT EXISTS idx_submenu_permissions_composite ON submenu_permissions(submenu_id, permission_id);
CREATE INDEX IF NOT EXISTS idx_menu_permissions_composite ON menu_permissions(menu_id, permission_id);
CREATE INDEX IF NOT EXISTS idx_module_permissions_composite ON module_permissions(module_id, permission_id);
CREATE INDEX IF NOT EXISTS idx_ui_settings_submenu ON ui_settings(submenu_id);
CREATE INDEX IF NOT EXISTS idx_role_ui_customizations_composite ON role_ui_customizations(role_id, submenu_id);

-- =====================================================
-- END OF NAVIGATION QUERIES
-- =====================================================