import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../../../../environments/environment';
import { forkJoin } from 'rxjs';
import { ToastyService } from '../../../../core/services/toasty.service';

interface Module {
  id: string;
  name: string;
  backendId: number;
  menus: Menu[];
}

interface Menu {
  id: string;
  label: string;
  backendId: number;
  menu_permissions: string[];
  submenus: Submenu[];
}

interface Submenu {
  id: string;
  label: string;
  backendId: number;
  hasAccess: boolean;
  required_permissions: string[];
}



interface RoleTemplate {
  id: number;
  name: string;
  code: string;
  description: string;
  userCount: number;
  icon: string;
}




@Component({
  selector: 'app-add-role',
  standalone: true,
  imports: [CommonModule, TranslateModule, FormsModule],
  templateUrl: './add-role.component.html',
  styleUrl: './add-role.component.scss'
})
export class AddRoleComponent implements OnInit {
  roleName: string = '';
  roleDescription: string = '';
  permissionScope: 'all' | 'own_store' | 'specific' = 'all';
  expandedModules: Set<string> = new Set();
  expandedMenus: Set<string> = new Set();
  activeTab: 'visual' | 'list' = 'visual';
  apiBaseUrl = environment.baseUrl;
  modules: Module[] = [];
  permissions = new Set<string>();
  roles: any;

  roleTemplates: RoleTemplate[] = [];
  selectedTemplate: number | null = null;
  stores: any[] = [];
  selectedStoreIds: number[] = [];



  constructor(private http: HttpClient,private toasty: ToastyService,) { }

  ngOnInit(): void {
    this.fetchAllRoles();
    this.fetchStores();

  }

  fetchStores() {
    this.http.get<any>(`${this.apiBaseUrl}/api/stores`)
      .subscribe(res => {
        this.stores = res.data.map((s: any) => ({
          id: s.id,
          name: s.name
        }));

        console.log('STORES LOADED:', this.stores);
      });
  }
  onScopeChange(scope: string) {
    this.permissionScope = scope as any;

    if (scope !== 'specific') {
      this.selectedStoreIds = [];
    }

    this.logPermissionData();
  }
  onStoreToggle(storeId: number, event: any) {
    if (event.target.checked) {
      this.selectedStoreIds.push(storeId);
    } else {
      this.selectedStoreIds = this.selectedStoreIds.filter(id => id !== storeId);
    }

    this.logPermissionData();
  }
  logPermissionData() {
    let payload: any;

    if (this.permissionScope === 'all') {
      payload = 'all';
    } else if (this.permissionScope === 'own_store') {
      payload = 'own';
    } else {
      payload = this.selectedStoreIds;
    }

    console.log('Permission Scope Data:', payload);
  }
  fetchAllRoles() {
    this.http.get<any>(`${this.apiBaseUrl}/api/roles`)
      .subscribe(res => {
        this.roleTemplates = res.data.map((role: any) => ({
          id: Number(role.id),                // 👈 adjust if backend uses different key
          name: role.name,
          description: role.description,
          code: role.code,
          userCount: 0
        })).sort((a: RoleTemplate, b: RoleTemplate) => a.id - b.id);
        if (this.roleTemplates.length > 0) {
          const firstRoleId = this.roleTemplates[0].id;
          this.handleTemplateSelect(firstRoleId);
        }
        this.loadUserCountsPerRole();

      });
  }
  loadUserCountsPerRole() {
    this.roleTemplates.forEach(role => {
      this.http.get<any>(
        `${this.apiBaseUrl}/api/navigation/rolesWithUserCounts/${role.code}`
      ).subscribe(res => {
        role.userCount = res.data?.user_count || 0;
      });
    });
  }
  handleTemplateSelect(roleId: number) {
    this.selectedTemplate = roleId;

    // clear old data
    this.modules = [];
    this.permissions.clear();
    this.expandedModules.clear();
    this.expandedMenus.clear();

    // load navigation for selected role
    this.loadNavigation(roleId);
  }
  loadNavigation(roleId: number) {
    this.http.get<any>(`${this.apiBaseUrl}/api/navigation/user/${roleId}`)
      .subscribe(res => {
        this.modules = this.mapApiToModules(res.data);

        this.permissions.clear();
        this.initCheckedPermissions(res.data);
      });
  }

  mapApiToModules(apiData: any[]): Module[] {
    return apiData.map(module => ({
      id: module.module_code,
      name: module.module_name,
      backendId: module.module_id,
      menus: module.menus.map((menu: any) => ({
        id: menu.menu_code,
        label: menu.menu_name,
        backendId: menu.menu_id,
        menu_permissions: menu.menu_permissions || [],
        submenus: menu.submenus?.map((sub: any) => ({
          id: sub.submenu_code,
          label: sub.submenu_name,
          backendId: sub.submenu_id,
          hasAccess: sub.has_access,
          required_permissions: sub.required_permissions || []
        })) || []
      }))
    }));
  }
  initCheckedPermissions(apiData: any[]) {
    apiData.forEach(module => {
      module.menus.forEach((menu: any) => {
        menu.submenus?.forEach((sub: any) => {
          if (sub.has_access) {
            const key = `${module.module_code}-${menu.menu_code}-${sub.submenu_code}`;
            this.permissions.add(key);
          }
        });
      });
    });
  }

  toggleModule(moduleId: string): void {
    if (this.expandedModules.has(moduleId)) {
      this.expandedModules.delete(moduleId);
    } else {
      this.expandedModules.add(moduleId);
    }
  }

  toggleMenu(menuId: string): void {
    if (this.expandedMenus.has(menuId)) {
      this.expandedMenus.delete(menuId);
    } else {
      this.expandedMenus.add(menuId);
    }
  }

  isModuleExpanded(moduleId: string): boolean {
    return this.expandedModules.has(moduleId);
  }

  isMenuExpanded(menuId: string): boolean {
    return this.expandedMenus.has(menuId);
  }

  togglePermission(
    key: string,
    hasChildren: boolean,
    childKeys: string[] = []
  ) {
    if (this.permissions.has(key)) {
      this.permissions.delete(key);
      if (hasChildren && childKeys.length) {
        childKeys.forEach(child => this.permissions.delete(child));
      }
    } else {
      this.permissions.add(key);
      if (hasChildren && childKeys.length) {
        childKeys.forEach(child => this.permissions.add(child));
      }
    }
    console.log('PERMISSIONS SET:', Array.from(this.permissions));
  }
  buildSelectedPermissions(): string[] {
    const selectedPermissions: string[] = [];
    this.modules.forEach(module => {
      const moduleKey = module.id;
      const moduleChildKeys = this.getChildIds(module);
      const moduleChecked = this.isAllChecked(moduleChildKeys);
      module.menus.forEach(menu => {
        const menuKey = `${module.id}-${menu.id}`;
        const menuChildKeys = this.getChildIds(module, menu.id);
        const menuChecked = this.isAllChecked(menuChildKeys);
        if (moduleChecked || menuChecked) {
          if (menu.menu_permissions?.length) {
            selectedPermissions.push(...menu.menu_permissions);
          }
          menu.submenus?.forEach(sub => {
            selectedPermissions.push(...sub.required_permissions);
          });
        }
        else {
          menu.submenus?.forEach(sub => {
            const subKey = `${module.id}-${menu.id}-${sub.id}`;
            if (this.permissions.has(subKey)) {
              selectedPermissions.push(...sub.required_permissions);
            }
          });
        }
      });
    });

    return Array.from(new Set(selectedPermissions));
  }
  getChildIds(module: Module, menuId?: string): string[] {
    const ids: string[] = [];
    if (menuId) {
      const menu = module.menus.find(m => m.id === menuId);
      if (menu) {
        ids.push(`${module.id}-${menu.id}`);
        menu.submenus?.forEach(sub => ids.push(`${module.id}-${menu.id}-${sub.id}`));
      }
    } else {
      module.menus.forEach(menu => {
        ids.push(`${module.id}-${menu.id}`);
        menu.submenus?.forEach(sub => ids.push(`${module.id}-${menu.id}-${sub.id}`));
      });
    }
    return ids;
  }

  isChecked(id: string): boolean {
    return this.permissions.has(id);
  }

  isAllChecked(childIds: string[]): boolean {
    return childIds.every(id => this.permissions.has(id));
  }

  isIndeterminate(module: Module, menuId?: string): boolean {
    const childIds = this.getChildIds(module, menuId);
    const checkedCount = childIds.filter(id => this.permissions.has(id)).length;
    return checkedCount > 0 && checkedCount < childIds.length;
  }

  generateRoleCode(name: string): string {
    return name
      .trim()
      .toLowerCase()
      .replace(/\s+/g, '_');
  }

  handleSave(): void {
    const selectedPermissions = this.buildSelectedPermissions();
    console.log('FINAL SELECTED PERMISSIONS (REAL):', selectedPermissions);
    const roleCode = this.generateRoleCode(this.roleName);
    let metadata: any = {
      scope: this.permissionScope
    };

    if (this.permissionScope === 'specific') {
      metadata.ids = this.selectedStoreIds;
    }
    const rolePayload = {
      code: roleCode,
      name: this.roleName,
      description: this.roleDescription,
      is_active: true,
      is_system_role: false,
      metadata: metadata
    };
    console.log('ROLE PAYLOAD:', rolePayload);
    this.http.post<any>(`${this.apiBaseUrl}/api/roles`, rolePayload)
      .subscribe({
        next: (res) => {
          const roleId = res.data?.id;
          console.log('ROLE CREATED, ID =', roleId);
          if (!roleId) {
            console.error('No roleId returned from API');
            return;
          }
          this.fetchRolePermissionsAndSave(roleId, selectedPermissions);
        },
        error: (err) => {
          this.toasty.error('Role creation failed');
          
        }
      });
  }

  fetchRolePermissionsAndSave(roleId: number, selectedPermissions: string[]) {
    this.http.get<any>(`${this.apiBaseUrl}/api/roles/1/permissions`)
      .subscribe({
        next: (res) => {
          const apiPermissions = res.data;
          const matched = apiPermissions.filter((p: any) =>
            selectedPermissions.includes(p.code)
          );
          console.log('MATCHED PERMISSIONS:', matched);
          const finalPayload = {
            permissions: matched.map((p: any, index: number) => ({
              permission_id: p.permission_id,
              scope: p.code,
              metadata: {
                notes: `permission ${index + 1}`
              }
            }))
          };
          console.log('FINAL PERMISSION PAYLOAD:', finalPayload);
          this.saveRolePermissions(roleId, finalPayload);
        },
        error: (err) => {
          console.error('Fetching role permissions failed:', err);
        }
      });
  }

  saveRolePermissions(roleId: number, payload: any) {
    this.http.post(
      `${this.apiBaseUrl}/api/roles/${roleId}/permissions`,
      payload
    ).subscribe({
      next: (res) => {
        console.log('PERMISSIONS SAVED:', res);
        this.toasty.success('Role + permissions created successfully!');
      },
      error: (err) => {
        console.error('Saving permissions failed:', err);
        this.toasty.error('Permissions save failed');
      }
    });
  }
  handleCloneExisting(): void {
    console.log('Clone existing role');
  }

  getSelectedCount(): number {
    return this.permissions.size;
  }

  getTotalCount(): number {
    return this.modules.reduce(
      (acc, m) => acc + m.menus.reduce((mAcc, menu) => mAcc + (menu.submenus?.length || 0), 0),
      0
    );
  }

  getProgressPercentage(): number {
    const total = this.getTotalCount();
    return total > 0 ? (this.getSelectedCount() / total) * 100 : 0;
  }

  setActiveTab(tab: 'visual' | 'list'): void {
    this.activeTab = tab;
  }
}
