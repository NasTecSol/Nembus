package sync

import (
	"NEMBUS/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type TenantCloner struct {
	ctx        context.Context
	masterRepo *repository.Queries
	cloudURL   string
	authToken  string // JWT bearer token for cloud API auth
	orgIDMap   map[int32]int32
	storeIDMap map[int32]int32
}

func NewTenantCloner(ctx context.Context, repo *repository.Queries, cloudURL string, authToken string) *TenantCloner {
	return &TenantCloner{
		ctx:        ctx,
		masterRepo: repo,
		cloudURL:   cloudURL,
		authToken:  authToken,
		orgIDMap:   make(map[int32]int32),
		storeIDMap: make(map[int32]int32),
	}
}

func (c *TenantCloner) CloneMasterData(slug string) error {
	// 1. Fetch Organizations
	if err := c.cloneOrganizations(slug); err != nil {
		return fmt.Errorf("failed to clone organizations: %w", err)
	}

	// 2. Fetch Modules, Menus, Submenus
	if err := c.cloneUIMetadata(slug); err != nil {
		return fmt.Errorf("failed to clone UI metadata: %w", err)
	}

	// 3. Fetch Roles and Permissions
	if err := c.cloneAccessControl(slug); err != nil {
		return fmt.Errorf("failed to clone access control: %w", err)
	}

	// 4. Fetch Users (for this tenant/org)
	if err := c.cloneUsers(slug); err != nil {
		return fmt.Errorf("failed to clone users: %w", err)
	}

	// 5. Fetch Stores and Locations
	if err := c.cloneStores(slug); err != nil {
		return fmt.Errorf("failed to clone stores: %w", err)
	}

	return nil
}

func (c *TenantCloner) cloneOrganizations(slug string) error {
	var organizations []repository.Organization
	if err := c.fetchCloudData("/api/organizations", slug, &organizations); err != nil {
		return err
	}

	for _, org := range organizations {
		params := repository.CreateOrganizationParams{
			Name:              org.Name,
			Code:              org.Code,
			LegalName:         org.LegalName,
			TaxID:             org.TaxID,
			CurrencyCode:      org.CurrencyCode,
			FiscalYearVariant: org.FiscalYearVariant,
			IsActive:          org.IsActive,
			Metadata:          org.Metadata,
		}
		newOrg, err := c.masterRepo.CreateOrganization(c.ctx, params)
		if err != nil {
			fmt.Printf("Warning: failed to clone org %s: %v\n", org.Code, err)
			continue
		}
		c.orgIDMap[org.ID] = newOrg.ID
	}
	return nil
}

func (c *TenantCloner) fetchCloudData(endpoint string, slug string, target interface{}) error {
	url := fmt.Sprintf("%s%s", c.cloudURL, endpoint)
	req, err := http.NewRequestWithContext(c.ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	if slug != "" {
		req.Header.Set("x-tenant-id", slug)
	}
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cloud API returned status %d for %s", resp.StatusCode, endpoint)
	}

	var result repository.Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	dataJson, err := json.Marshal(result.Data)
	if err != nil {
		return err
	}

	return json.Unmarshal(dataJson, target)
}

func (c *TenantCloner) cloneUIMetadata(slug string) error {
	var modules []repository.Module
	if err := c.fetchCloudData("/api/modules", slug, &modules); err != nil {
		return err
	}

	for _, mod := range modules {
		modParams := repository.CreateModuleParams{
			Name:         mod.Name,
			Code:         mod.Code,
			Description:  mod.Description,
			Icon:         mod.Icon,
			IsActive:     mod.IsActive,
			DisplayOrder: mod.DisplayOrder,
			Metadata:     mod.Metadata,
		}
		newMod, err := c.masterRepo.CreateModule(c.ctx, modParams)
		if err != nil {
			fmt.Printf("Warning: failed to clone module %s: %v\n", mod.Code, err)
			continue
		}

		// Fetch menus for this module
		var menus []repository.Menu
		endpoint := fmt.Sprintf("/api/modules/%d/menus", mod.ID)
		if err := c.fetchCloudData(endpoint, slug, &menus); err != nil {
			continue
		}

		for _, menu := range menus {
			menuParams := repository.CreateMenuParams{
				ModuleID:     newMod.ID,
				Name:         menu.Name,
				Code:         menu.Code,
				RoutePath:    menu.RoutePath,
				Icon:         menu.Icon,
				DisplayOrder: menu.DisplayOrder,
				IsActive:     menu.IsActive,
				Metadata:     menu.Metadata,
			}
			newMenu, err := c.masterRepo.CreateMenu(c.ctx, menuParams)
			if err != nil {
				continue
			}

			// Fetch submenus
			var submenus []repository.Submenu
			smEndpoint := fmt.Sprintf("/api/menus/%d/submenus", menu.ID)
			if err := c.fetchCloudData(smEndpoint, slug, &submenus); err != nil {
				continue
			}

			for _, sm := range submenus {
				smParams := repository.CreateSubmenuParams{
					MenuID:       newMenu.ID,
					Name:         sm.Name,
					Code:         sm.Code,
					RoutePath:    sm.RoutePath,
					Icon:         sm.Icon,
					DisplayOrder: sm.DisplayOrder,
					IsActive:     sm.IsActive,
					Metadata:     sm.Metadata,
				}
				c.masterRepo.CreateSubmenu(c.ctx, smParams)
			}
		}
	}
	return nil
}

func (c *TenantCloner) cloneAccessControl(slug string) error {
	// 1. Clone Permissions
	var permissions []repository.Permission
	if err := c.fetchCloudData("/api/permissions", slug, &permissions); err != nil {
		return err
	}
	// Note: CreatePermission query might be needed
	// Skip detailed implementation for now to keep focus on flow
	return nil
}

func (c *TenantCloner) cloneUsers(slug string) error {
	var users []repository.User
	if err := c.fetchCloudData("/api/users", slug, &users); err != nil {
		return err
	}

	for _, user := range users {
		localOrgID, ok := c.orgIDMap[user.OrganizationID]
		if !ok {
			continue
		}

		params := repository.CreateUserParams{
			OrganizationID: localOrgID,
			Username:       user.Username,
			Email:          user.Email,
			PasswordHash:   user.PasswordHash,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			EmployeeCode:   user.EmployeeCode,
			IsActive:       user.IsActive,
			Metadata:       user.Metadata,
		}
		c.masterRepo.CreateUser(c.ctx, params)
	}
	return nil
}

func (c *TenantCloner) cloneStores(slug string) error {
	var stores []repository.Store
	if err := c.fetchCloudData("/api/stores", slug, &stores); err != nil {
		return err
	}

	for _, store := range stores {
		localOrgID, ok := c.orgIDMap[store.OrganizationID]
		if !ok {
			continue
		}

		params := repository.CreateStoreParams{
			OrganizationID: localOrgID,
			Name:           store.Name,
			Code:           store.Code,
			StoreType:      store.StoreType,
			IsWarehouse:    store.IsWarehouse,
			IsPosEnabled:   store.IsPosEnabled,
			Timezone:       store.Timezone,
			IsActive:       store.IsActive,
			Metadata:       store.Metadata,
		}
		c.masterRepo.CreateStore(c.ctx, params)
	}
	return nil
}
