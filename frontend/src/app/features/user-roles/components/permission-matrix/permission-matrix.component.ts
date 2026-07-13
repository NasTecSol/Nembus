import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../../../../environments/environment';
import { forkJoin } from 'rxjs';
import { ToastyService } from '../../../../core/services/toasty.service';

interface Submenu {
  id: number;
  label: string;
}

interface Menu {
  id: number;
  label: string;
  submenus?: Submenu[];
}

interface Module {
  id: number;
  name: string;
  code: string;
  menus: Menu[];
}

interface Role {
  id: number;        
  name: string;      
  code: string;      
  userCount: number;
  color: string;
}

interface PermissionState {
  [roleId: string]: {
    [moduleId: string]: {
      [menuId: string]: boolean;
    };
  };
}

interface PermissionMapping {
  code: string;
  permission_id: number;
  scope: string;
}

interface PermissionPayload {
  permission_id: number;
  scope: string;
  metadata: {
    notes: string;
  };
}

interface PermissionRequest {
  permissions: PermissionPayload[];
}

interface NavigationData {
  module_id: number;
  module_code: string;
  menus: Array<{
    menu_id: number;
    menu_code: string;
    menu_permissions: string[];
    submenus?: Array<{
      submenu_id: number;
      submenu_code: string;
      required_permissions: string[];
    }>;
  }>;
}

@Component({
  selector: 'app-permission-matrix',
  standalone: true,
  imports: [CommonModule, TranslateModule, FormsModule],
  templateUrl: './permission-matrix.component.html',
  styleUrl: './permission-matrix.component.scss'
})
export class PermissionMatrixComponent implements OnInit {
  searchTerm: string = '';
  selectedRole: number | 'all' = 'all';
  permissions: PermissionState = {};
  originalPermissions: PermissionState = {}; 
  hasChanges: boolean = false;
  expandedMenus: Set<string> = new Set();
  apiBaseUrl = environment.baseUrl;
  roles: Role[] = [];
  displayRoles: Role[] = [];
  modules: Module[] = [];
  filteredModules: Module[] = [];
  permissionMappings: PermissionMapping[] = [];
  roleOneNavigation: NavigationData[] = [];

  isSaving: boolean = false;

  constructor(private http: HttpClient, private toasty: ToastyService, private translate: TranslateService) { }

  ngOnInit(): void {
    this.updateFilters();
    this.fetchRoles();
  }
  fetchRoles(): void {
    this.http.get<any>(`${this.apiBaseUrl}/api/roles`).subscribe({
      next: (res) => {
        this.roles = res.data
          .sort((a: any, b: any) => a.id - b.id)
          .map((role: any, index: number) => ({
            id: role.id,
            name: role.name,
            code: role.code,
            userCount: 0,
            color: this.getRoleColor(index)
          }));
        this.displayRoles = [...this.roles];
        this.updateFilters();
        this.fetchReferenceData();
        this.displayRoles.forEach(role => {
          this.fetchNavigationAndCountByRoleCode(role);
        });
        this.displayRoles.forEach(role => {
          this.fetchNavigationForRole(role.id);
        });
      },
      error: (err) => console.error('Error fetching roles', err)
    });
  }

  fetchReferenceData(): void {
    // Fetch both permission mappings and navigation for role ID 1
    forkJoin({
      permissions: this.http.get<any>(`${this.apiBaseUrl}/api/roles/1/permissions`),
      navigation: this.http.get<any>(`${this.apiBaseUrl}/api/navigation/user/1`)
    }).subscribe({
      next: (result) => {
        // Store permission mappings
        this.permissionMappings = result.permissions.data.map((perm: any) => ({
          code: perm.code,
          permission_id: perm.permission_id,
          scope: perm.scope
        }));

        // Store role 1 navigation
        this.roleOneNavigation = result.navigation.data;

        console.log('Permission Mappings:', this.permissionMappings);
        console.log('Role 1 Navigation:', this.roleOneNavigation);
      },
      error: (err) => console.error('Error fetching reference data', err)
    });
  }

  fetchNavigationAndCountByRoleCode(role: any): void {
    const url = `${this.apiBaseUrl}/api/navigation/rolesWithUserCounts/${role.code}`;

    this.http.get<any>(url).subscribe({
      next: (res) => {
        const navigation = res.data.navigation;
        const userCount = res.data.user_count;

        role.userCount = userCount;

        if (!this.modules || this.modules.length === 0) {
          this.modules = navigation.map((mod: any) => ({
            id: mod.module_id,
            name: mod.module_name,
            code: mod.module_code,
            menus: mod.menus.map((menu: any) => ({
              id: menu.menu_id,
              label: menu.menu_name,
              submenus: menu.submenus?.map((sub: any) => ({
                id: sub.submenu_id,
                label: sub.submenu_name
              })) || []
            }))
          }));
        }

        this.initializePermissionsFromApi(role.id, navigation);

        console.log(`Role ${role.code} → users:`, userCount);
      },
      error: (err) => console.error(`Navigation error for ${role.code}`, err)
    });
  }

  initializePermissionsFromApi(roleId: number, apiModules: any[]): void {
    if (!this.permissions[roleId]) {
      this.permissions[roleId] = {};
    }
    if (!this.originalPermissions[roleId]) {
      this.originalPermissions[roleId] = {};
    }

    apiModules.forEach((mod: any) => {
      if (!this.permissions[roleId][mod.module_id]) {
        this.permissions[roleId][mod.module_id] = {};
      }
      if (!this.originalPermissions[roleId][mod.module_id]) {
        this.originalPermissions[roleId][mod.module_id] = {};
      }

      mod.menus.forEach((menu: any) => {
        if (menu.submenus && menu.submenus.length > 0) {
          menu.submenus.forEach((sub: any) => {
            const key = `${menu.menu_id}-${sub.submenu_id}`;
            const hasAccess = sub.has_access === true;
            this.permissions[roleId][mod.module_id][key] = hasAccess;
            this.originalPermissions[roleId][mod.module_id][key] = hasAccess;
          });
        } else {
          const hasAccess = menu.has_access === true;
          this.permissions[roleId][mod.module_id][menu.menu_id] = hasAccess;
          this.originalPermissions[roleId][mod.module_id][menu.menu_id] = hasAccess;
        }
      });
    });
  }

  fetchNavigationForRole(roleId: number): void {
    const url = `${this.apiBaseUrl}/api/navigation/user/${roleId}`;
    this.http.get<any>(url).subscribe({
      next: (res) => {
        const apiModules = res.data;

        const mappedModules: Module[] = apiModules.map((mod: any) => ({
          id: mod.module_id,
          name: mod.module_name,
          code: mod.module_code,
          menus: mod.menus.map((menu: any) => ({
            id: menu.menu_id,
            label: menu.menu_name,
            submenus: menu.submenus?.map((sub: any) => ({
              id: sub.submenu_id,
              label: sub.submenu_name
            })) || []
          }))
        }));

        this.initializePermissions(roleId, apiModules);
        this.mergeModules(mappedModules);
        this.updateFilters();

        console.log(`Mapped modules for role ${roleId}:`, mappedModules);
      },
      error: (err) => console.error('Navigation API error', err)
    });
  }

  initializePermissions(roleId: number, apiModules: any[]): void {
    if (!this.permissions[roleId]) this.permissions[roleId] = {};
    if (!this.originalPermissions[roleId]) this.originalPermissions[roleId] = {};

    apiModules.forEach(mod => {
      if (!this.permissions[roleId][mod.module_id]) this.permissions[roleId][mod.module_id] = {};
      if (!this.originalPermissions[roleId][mod.module_id]) this.originalPermissions[roleId][mod.module_id] = {};

      mod.menus.forEach((menu: any) => {
        if (menu.submenus && menu.submenus.length > 0) {
          menu.submenus.forEach((sub: any) => {
            const key = `${menu.menu_id}-${sub.submenu_id}`;
            const hasAccess = sub.has_access === true;
            this.permissions[roleId][mod.module_id][key] = hasAccess;
            this.originalPermissions[roleId][mod.module_id][key] = hasAccess;
          });
        } else {
          const hasAccess = menu.has_access === true;
          this.permissions[roleId][mod.module_id][menu.menu_id] = hasAccess;
          this.originalPermissions[roleId][mod.module_id][menu.menu_id] = hasAccess;
        }
      });
    });
  }

  mergeModules(newModules: Module[]): void {
    if (!this.modules.length) {
      this.modules = [...newModules];
    } else {
      newModules.forEach(newMod => {
        const existing = this.modules.find(m => m.id === newMod.id);
        if (!existing) {
          this.modules.push(newMod);
        }
      });
    }
  }

  getRoleColor(index: number): string {
    const colors = ['purple', 'blue', 'green', 'yellow', 'orange', 'pink'];
    return colors[index % colors.length];
  }

  updateFilters(): void {
    this.filteredModules = this.modules.filter(
      (module) =>
        module.name.toLowerCase().includes(this.searchTerm.toLowerCase()) ||
        module.menus.some((menu) =>
          menu.label.toLowerCase().includes(this.searchTerm.toLowerCase())
        )
    );

    this.displayRoles =
      this.selectedRole === 'all'
        ? this.roles
        : this.roles.filter((r) => r.id === this.selectedRole);
  }

  onSearchChange(): void {
    this.updateFilters();
  }

  onRoleFilterChange(): void {
    this.updateFilters();
  }

  toggleMenuExpand(moduleId: number, menuId: number): void {
    const key = `${moduleId}-${menuId}`;
    if (this.expandedMenus.has(key)) {
      this.expandedMenus.delete(key);
    } else {
      this.expandedMenus.add(key);
    }
  }

  isMenuExpanded(moduleId: number, menuId: number): boolean {
    return this.expandedMenus.has(`${moduleId}-${menuId}`);
  }

  togglePermission(roleId: number, moduleId: number, menuId: number, submenuId?: number): void {
    if (!this.permissions[roleId]) this.permissions[roleId] = {};
    if (!this.permissions[roleId][moduleId]) this.permissions[roleId][moduleId] = {};

    const key = submenuId ? `${menuId}-${submenuId}` : menuId;
    this.permissions[roleId][moduleId][key] = !this.permissions[roleId][moduleId][key];
    this.checkForChanges();
  }

  toggleMenu(roleId: number, moduleId: number, menuId: number, event: Event): void {
    const checked = (event.target as HTMLInputElement).checked;
    const module = this.modules.find(m => m.id === moduleId);
    const menu = module?.menus.find(m => m.id === menuId);

    if (!this.permissions[roleId]) this.permissions[roleId] = {};
    if (!this.permissions[roleId][moduleId]) this.permissions[roleId][moduleId] = {};

    if (menu?.submenus?.length) {
      menu.submenus.forEach(sub => {
        this.permissions[roleId][moduleId][`${menuId}-${sub.id}`] = checked;
      });
    } else {
      this.permissions[roleId][moduleId][menuId] = checked;
    }

    this.checkForChanges();
  }

  toggleModule(roleId: number, moduleId: number, event: Event): void {
    const checked = (event.target as HTMLInputElement).checked;
    const module = this.modules.find(m => m.id === moduleId);

    if (!this.permissions[roleId]) this.permissions[roleId] = {};
    if (!this.permissions[roleId][moduleId]) this.permissions[roleId][moduleId] = {};

    module?.menus.forEach(menu => {
      if (menu.submenus?.length) {
        menu.submenus.forEach(sub => {
          this.permissions[roleId][moduleId][`${menu.id}-${sub.id}`] = checked;
        });
      } else {
        this.permissions[roleId][moduleId][menu.id] = checked;
      }
    });

    this.checkForChanges();
  }

  checkForChanges(): void {
    this.hasChanges = JSON.stringify(this.permissions) !== JSON.stringify(this.originalPermissions);
  }

  isMenuChecked(roleId: number, moduleId: number, menuId: number): boolean {
    const module = this.modules.find(m => m.id === moduleId);
    const menu = module?.menus.find(m => m.id === menuId);

    if (!menu) return false;

    if (menu.submenus?.length) {
      return menu.submenus.every(sub => this.permissions[roleId]?.[moduleId]?.[`${menuId}-${sub.id}`]);
    } else {
      return this.permissions[roleId]?.[moduleId]?.[menuId] || false;
    }
  }

  isMenuIndeterminate(roleId: number, moduleId: number, menuId: number): boolean {
    const module = this.modules.find(m => m.id === moduleId);
    const menu = module?.menus.find(m => m.id === menuId);

    if (!menu?.submenus?.length) return false;

    const subStates = menu.submenus.map(sub => !!this.permissions[roleId]?.[moduleId]?.[`${menuId}-${sub.id}`]);
    return subStates.some(s => s) && !subStates.every(s => s);
  }

  isSubmenuChecked(roleId: number, moduleId: number, menuId: number, submenuId: number): boolean {
    return this.permissions[roleId]?.[moduleId]?.[`${menuId}-${submenuId}`] === true;
  }

  isModuleChecked(roleId: number, moduleId: number): boolean | 'indeterminate' {
    const module = this.modules.find(m => m.id === moduleId);
    if (!module) return false;

    const menuStates = module.menus.map(menu => this.isMenuChecked(roleId, moduleId, menu.id));
    if (menuStates.every(s => s)) return true;
    if (menuStates.some(s => s)) return 'indeterminate';
    return false;
  }

  isModuleIndeterminate(roleId: number, moduleId: number): boolean {
    return this.isModuleChecked(roleId, moduleId) === 'indeterminate';
  }

  getPermissionChanges(roleId: number): { toAdd: PermissionPayload[], toDelete: PermissionPayload[] } {
    const toAdd: PermissionPayload[] = [];
    const toDelete: PermissionPayload[] = [];

    const currentPerms = this.permissions[roleId] || {};
    const originalPerms = this.originalPermissions[roleId] || {};

    this.roleOneNavigation.forEach(navModule => {
      const moduleId = navModule.module_id;

      navModule.menus.forEach(menu => {
        if (menu.submenus?.length) {
          menu.submenus.forEach(submenu => {
            const key = `${menu.menu_id}-${submenu.submenu_id}`;
            const currentState = currentPerms[moduleId]?.[key] || false;
            const originalState = originalPerms[moduleId]?.[key] || false;

            submenu.required_permissions.forEach(permCode => {
              const mapping = this.permissionMappings.find(p => p.code === permCode);
              if (mapping) {
                const payload: PermissionPayload = {
                  permission_id: mapping.permission_id,
                  scope: permCode,
                  metadata: { notes: "Updated permission" }
                };
                if (currentState && !originalState) toAdd.push(payload);     
                if (!currentState && originalState) toDelete.push(payload); 
              }
            });
          });
        } else {
          const key = menu.menu_id;
          const currentState = currentPerms[moduleId]?.[key] || false;
          const originalState = originalPerms[moduleId]?.[key] || false;

          menu.menu_permissions.forEach(permCode => {
            const mapping = this.permissionMappings.find(p => p.code === permCode);
            if (mapping) {
              const payload: PermissionPayload = {
                permission_id: mapping.permission_id,
                scope: permCode,
                metadata: { notes: "Updated permission" }
              };
              if (currentState && !originalState) toAdd.push(payload);
              if (!currentState && originalState) toDelete.push(payload);
            }
          });
        }
      });
    });
    const unique = (arr: PermissionPayload[]) =>
      arr.filter((p, i, self) => i === self.findIndex(x => x.permission_id === p.permission_id));

    return { toAdd: unique(toAdd), toDelete: unique(toDelete) };
  }


  handleSave(): void {
    if (!this.hasChanges || this.isSaving) return;

    this.isSaving = true;
    const saveRequests: any[] = [];

    Object.keys(this.permissions).forEach(roleIdStr => {
      const roleId = parseInt(roleIdStr);
      if (roleId === 1) return;

      const { toAdd, toDelete } = this.getPermissionChanges(roleId);

      if (toAdd.length > 0) {
        console.log(`Role ${roleId} → POST payload:`, { permissions: toAdd });
        saveRequests.push(
          this.http.post(`${this.apiBaseUrl}/api/roles/${roleId}/permissions`, { permissions: toAdd })
        );
      }

      if (toDelete.length > 0) {
        const permissionIds = toDelete.map(p => p.permission_id);
        console.log(`Role ${roleId} → DELETE payload:`, { permission_ids: permissionIds });
        saveRequests.push(
          this.http.request('delete', `${this.apiBaseUrl}/api/roles/${roleId}/permissions`, { body: { permission_ids: permissionIds } })
        );
      }
    });

    if (saveRequests.length === 0) {
      this.isSaving = false;
      alert(this.translate.instant('PERMISSION_MATRIX.NO_CHANGES_TO_SAVE'));
      return;
    }

    forkJoin(saveRequests).subscribe({
      next: responses => {
        console.log('Permissions updated successfully', responses);
        this.originalPermissions = JSON.parse(JSON.stringify(this.permissions));
        this.hasChanges = false;
        this.isSaving = false;
        this.toasty.success(this.translate.instant('PERMISSION_MATRIX.PERMISSIONS_UPDATED'));
      },
      error: err => {
        console.error('Error saving permissions', err);
        this.isSaving = false;
        this.toasty.error(this.translate.instant('PERMISSION_MATRIX.ERROR_SAVING'));
      }
    });
  }



  getChangedPermissionsForRole(roleId: number): PermissionPayload[] {
    const changedPermissions: PermissionPayload[] = [];
    const currentPerms = this.permissions[roleId] || {};
    const originalPerms = this.originalPermissions[roleId] || {};

    // Iterate through all modules in role 1 navigation
    this.roleOneNavigation.forEach(navModule => {
      const moduleId = navModule.module_id;

      navModule.menus.forEach(menu => {
        // Check if menu has submenus
        if (menu.submenus && menu.submenus.length > 0) {
          menu.submenus.forEach(submenu => {
            const key = `${menu.menu_id}-${submenu.submenu_id}`;
            const currentState = currentPerms[moduleId]?.[key] || false;
            const originalState = originalPerms[moduleId]?.[key] || false;

            // If state changed
            if (currentState !== originalState) {
              // Find permission mappings for required permissions
              submenu.required_permissions.forEach(permCode => {
                const mapping = this.permissionMappings.find(p => p.code === permCode);
                if (mapping && currentState) {
                  // Only add if permission is now enabled
                  changedPermissions.push({
                    permission_id: mapping.permission_id,
                    scope: permCode,  // Send the code in scope field
                    metadata: {
                      notes: "Updated permission"
                    }
                  });
                }
              });
            }
          });
        } else {
          // Menu without submenus
          const key = menu.menu_id;
          const currentState = currentPerms[moduleId]?.[key] || false;
          const originalState = originalPerms[moduleId]?.[key] || false;

          if (currentState !== originalState) {
            // Find permission mappings for menu permissions
            menu.menu_permissions.forEach(permCode => {
              const mapping = this.permissionMappings.find(p => p.code === permCode);
              if (mapping && currentState) {
                changedPermissions.push({
                  permission_id: mapping.permission_id,
                  scope: permCode,  // Send the code in scope field
                  metadata: {
                    notes: "Updated permission"
                  }
                });
              }
            });
          }
        }
      });
    });

    // Remove duplicates based on permission_id
    const uniquePermissions = changedPermissions.filter((perm, index, self) =>
      index === self.findIndex(p => p.permission_id === perm.permission_id)
    );

    return uniquePermissions;
  }

  handleReset(): void {
    if (confirm(this.translate.instant('PERMISSION_MATRIX.ARE_YOU_SURE_RESET'))) {
      this.permissions = JSON.parse(JSON.stringify(this.originalPermissions));
      this.hasChanges = false;
    }
  }

  exportPermissions(): void {
    console.log('Exporting permissions...');
  }

  getTotalPermissions(): number {
    return this.modules.reduce((acc, m) => acc + m.menus.length, 0);
  }
}
