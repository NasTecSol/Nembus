import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../../../../environments/environment';
import { TranslateModule } from "@ngx-translate/core";
export interface Menu {
  id: number;
  module_id: number;
  parent_menu_id: number | null;
  name: string;
  code: string;
  route_path: string;
  icon: string;
  display_order: number;
  is_active: boolean;
  metadata: any;
  created_at: string;
  updated_at: string;
}

export interface Module {
  id: number;
  name: string;
  code: string;
  description: string;
  icon: string;
  is_active: boolean;
  display_order: number;
  metadata: any;
}

// ── Permission from API ──────────────────────────────────────────────────────
export interface Permission {
  id: number;
  name: string;
  code: string;
  description: string;
  metadata: any;
  created_at: string;
}

export interface MenuForm {
  name: string;
  code: string;
  route_path: string;
  icon: string;
  display_order: number;
  is_active: boolean;
  module_id: number | null;
  parent_menu_id: number | string | null;
  parent_menu_route: string;
  hide_no_perm: boolean;
  show_in_sidebar: boolean;
  /** Selected permission IDs (not names) */
  selectedPermissionIds: number[];
}

export interface CreateMenuPayload {
  code: string;
  display_order: number;
  icon: string;
  is_active: boolean;
  metadata: string;
  module_id: number;
  name: string;
  parent_menu_id: number | null;
  route_path: string;
}

export interface CreateSubmenuPayload {
  code: string;
  display_order: number;
  icon: string;
  is_active: boolean;
  menu_id: number;
  metadata: string;
  name: string;
  parent_submenu_id: null;
  route_path: string;
}

@Component({
  selector: 'app-menue-manage',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslateModule],
  templateUrl: './menue-manage.component.html',
  styleUrl: './menue-manage.component.scss'
})
export class MenueManageComponent implements OnInit {

  apiUrl = environment.baseUrl;

  // View state
  currentView: 'list' | 'add-menu' | 'add-submenu' | 'edit' = 'list';

  // Type selector modal
  showTypeSelector = false;
  selectedType: 'menu' | 'submenu' | null = null;

  // Menus data
  allMenus: Menu[] = [];
  filteredMenus: Menu[] = [];
  isLoading = false;
  errorMsg = '';

  // Modules data
  allModules: Module[] = [];
  modulesLoading = false;

  // ── Permissions state ────────────────────────────────────────────────────
  allPermissions: Permission[] = [];
  permissionsLoading = false;
  permissionsError = '';
  permissionSearch = '';

  // Saving state
  isSaving = false;
  saveError = '';
  saveSuccess = false;

  // Filters
  searchQuery = '';
  statusFilter = 'all';

  // Expand/collapse
  expandedIds = new Set<number>();

  // Edit state
  editingMenu: Menu | null = null;

  // Form model
  form: MenuForm = this.emptyForm();

  // Icon picker
  iconSearch = '';
  iconDefs: Record<string, string> = {
    home: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6',
    dashboard: 'M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z',
    settings: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z',
    people: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z',
    inventory: 'M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4',
    shipping: 'M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4',
    warehouse: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4',
    store: 'M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z',
    receipt: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2',
    money: 'M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
    chart: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z',
    analytics: 'M16 8v8m-4-5v5m-4-2v2m-2 4h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z',
    document: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
    folder: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z',
    build: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z',
    security: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z',
    lock: 'M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z',
    key: 'M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z',
    admin: 'M5.121 17.804A13.937 13.937 0 0112 16c2.5 0 4.847.655 6.879 1.804M15 10a3 3 0 11-6 0 3 3 0 016 0zm6 2a9 9 0 11-18 0 9 9 0 0118 0z',
    group: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z',
    business: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4',
    location: 'M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z M15 11a3 3 0 11-6 0 3 3 0 016 0z',
    map: 'M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7',
    schedule: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z',
    notifications: 'M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9',
    email: 'M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z',
    cloud: 'M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z',
    code: 'M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4',
    cart: 'M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z',
    payments: 'M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z',
    bank: 'M3 6l3 1m0 0l-3 9a5.002 5.002 0 006.001 0M6 7l3 9M6 7l6-2m6 2l3-1m-3 1l-3 9a5.002 5.002 0 006.001 0M18 7l3 9m-3-9l-6-2m0-2v2m0 16V5m0 16H9m3 0h3',
    trending: 'M13 7h8m0 0v8m0-8l-8 8-4-4-6 6',
    restaurant: 'M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z',
    pos: 'M9 7H6a2 2 0 00-2 2v9a2 2 0 002 2h9a2 2 0 002-2v-3M9 7V5a2 2 0 012-2h6l2 2v8a2 2 0 01-2 2h-3M9 7h3a2 2 0 012 2v3',
    category: 'M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z',
    layers: 'M12 6.042A8.967 8.967 0 006 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 016 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 016-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0018 18a8.967 8.967 0 00-6 2.292m0-14.25v14.25',
    menu: 'M4 6h16M4 12h16M4 18h16',
    support: 'M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192l-3.536 3.536M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z',
    report: 'M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
    users: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z',
    building: 'M8 14v3m4-3v3m4-3v3M3 21h18M3 10h18M3 7l9-4 9 4M4 10h16v11H4V10z',
    pie: 'M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z M20.488 9H15V3.512A9.025 9.025 0 0120.488 9z',
    link: 'M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1',
    flag: 'M3 21v-4m0 0V5a2 2 0 012-2h6.5l1 1H21l-3 6 3 6h-8.5l-1-1H5a2 2 0 00-2 2zm9-13.5V9',
    star: 'M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z',
    box: 'M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4',
    filter: 'M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z',
    search: 'M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z',
  };

  constructor(private http: HttpClient) { }

  ngOnInit(): void {
    this.fetchMenus();
    this.fetchModules();
    this.fetchPermissions();
  }

  // ── Fetch all permissions from API ────────────────────────────────────────
  fetchPermissions(): void {
    this.permissionsLoading = true;
    this.permissionsError = '';
    this.http.get<{ statusCode: number; message: string; data: Permission[] }>(
      `${this.apiUrl}/api/permissions`
    ).subscribe({
      next: (res) => {
        this.allPermissions = (res.data || []).sort((a, b) => a.name.localeCompare(b.name));
        this.permissionsLoading = false;
      },
      error: (err) => {
        this.permissionsError = 'Failed to load permissions.';
        this.permissionsLoading = false;
        console.error('Permissions fetch error:', err);
      }
    });
  }

  /** Permissions filtered by the search box inside the card */
  get filteredPermissions(): Permission[] {
    const q = this.permissionSearch.toLowerCase().trim();
    if (!q) return this.allPermissions;
    return this.allPermissions.filter(
      p => p.name.toLowerCase().includes(q) || p.code.toLowerCase().includes(q)
    );
  }

  isPermissionSelected(id: number): boolean {
    return this.form.selectedPermissionIds.includes(id);
  }

  togglePermission(id: number): void {
    const idx = this.form.selectedPermissionIds.indexOf(id);
    if (idx === -1) {
      this.form.selectedPermissionIds.push(id);
    } else {
      this.form.selectedPermissionIds.splice(idx, 1);
    }
  }

  selectAllPermissions(): void {
    this.form.selectedPermissionIds = this.allPermissions.map(p => p.id);
  }

  clearAllPermissions(): void {
    this.form.selectedPermissionIds = [];
  }

  // ── Fetch menus & modules ─────────────────────────────────────────────────
  fetchMenus(): void {
    this.isLoading = true;
    this.errorMsg = '';
    this.http.get<{ statusCode: number; message: string; data: Menu[] }>(
      `${this.apiUrl}/api/menus`
    ).subscribe({
      next: (res) => {
        this.allMenus = res.data || [];
        this.applyFilters();
        this.isLoading = false;
      },
      error: (err) => {
        this.errorMsg = 'Failed to load menus. Please try again.';
        this.isLoading = false;
        console.error('Menu fetch error:', err);
      }
    });
  }

  fetchModules(): void {
    this.modulesLoading = true;
    this.http.get<{ statusCode: number; message: string; data: Module[] }>(
      `${this.apiUrl}/api/modules`
    ).subscribe({
      next: (res) => {
        this.allModules = (res.data || []).filter(m => m.is_active);
        this.modulesLoading = false;
      },
      error: (err) => {
        console.error('Modules fetch error:', err);
        this.modulesLoading = false;
      }
    });
  }

  // ── Save dispatcher ───────────────────────────────────────────────────────
  saveMenu(): void {
    const isSubmenu = this.currentView === 'add-submenu';
    isSubmenu ? this.saveSubmenu() : this.saveParentMenu();
  }

  // ── Save Menu → then assign permissions via /api/permissions/menu/{id} ────
  private saveParentMenu(): void {
    if (!this.form.name || !this.form.code || !this.form.route_path) {
      this.saveError = 'Please fill in all required fields: Name, Code, Route Path.';
      return;
    }
    if (!this.form.module_id) {
      this.saveError = 'Please select a Parent Module.';
      return;
    }

    const payload: CreateMenuPayload = {
      code: this.form.code.trim(),
      display_order: Number(this.form.display_order),
      icon: this.form.icon,
      is_active: this.form.is_active,
      metadata: JSON.stringify({ color: 'blue' }),
      module_id: Number(this.form.module_id),
      name: this.form.name.trim(),
      parent_menu_id: null,
      route_path: this.form.route_path.trim()
    };

    this.isSaving = true;
    this.saveError = '';
    this.saveSuccess = false;

    this.http.post<{ statusCode: number; message: string; data: Menu }>(
      `${this.apiUrl}/api/menus`, payload
    ).subscribe({
      next: (res) => {
        const menuId = res?.data?.id;
        if (!menuId) {
          this.isSaving = false;
          this.saveError = 'Menu created but ID was not returned.';
          return;
        }

        // Update local list
        if (res.data) {
          this.allMenus.push(res.data);
          this.applyFilters();
        }

        // If no permissions selected, finish immediately
        if (this.form.selectedPermissionIds.length === 0) {
          this.finishSaveSuccess();
          return;
        }

        // Assign selected permissions to the menu
        this.assignPermissionsToMenu(menuId, this.form.selectedPermissionIds);
      },
      error: (err) => {
        this.isSaving = false;
        this.saveError = err?.error?.message || 'Failed to save menu. Please try again.';
        console.error('Save menu error:', err);
      }
    });
  }

  // ── Assign permissions to a Menu ──────────────────────────────────────────
  private assignPermissionsToMenu(menuId: number, permissionIds: number[]): void {
    const payload = permissionIds.map(id => ({
      metadata: {},
      permission_id: id
    }));

    this.http.post(
      `${this.apiUrl}/api/permissions/menu/${menuId}`,
      payload
    ).subscribe({
      next: () => {
        this.finishSaveSuccess();
      },
      error: (err) => {
        // Menu was saved — just log permission error, don't block success
        console.error('Permission assignment error:', err);
        this.finishSaveSuccess();
      }
    });
  }

  // ── Save Submenu → then assign permissions via /api/permissions/submenu/{id}
  private saveSubmenu(): void {
    if (!this.form.name || !this.form.code || !this.form.route_path) {
      this.saveError = 'Please fill in all required fields: Name, Code, Route Path.';
      return;
    }
    if (!this.form.parent_menu_id) {
      this.saveError = 'Please select a Parent Menu.';
      return;
    }

    const payload: CreateSubmenuPayload = {
      code: this.form.code.trim(),
      display_order: Number(this.form.display_order),
      icon: this.form.icon,
      is_active: this.form.is_active,
      menu_id: Number(this.form.parent_menu_id),
      metadata: JSON.stringify({ color: 'blue' }),
      name: this.form.name.trim(),
      parent_submenu_id: null,
      route_path: this.form.route_path.trim()
    };

    this.isSaving = true;
    this.saveError = '';
    this.saveSuccess = false;

    this.http.post<{ statusCode: number; message: string; data: any }>(
      `${this.apiUrl}/api/submenus`, payload
    ).subscribe({
      next: (res) => {
        const submenuId = res?.data?.id;

        if (!submenuId || this.form.selectedPermissionIds.length === 0) {
          this.finishSaveSuccess();
          return;
        }

        // Assign selected permissions to the submenu
        this.assignPermissionsToSubmenu(submenuId, this.form.selectedPermissionIds);
      },
      error: (err) => {
        this.isSaving = false;
        this.saveError = err?.error?.message || 'Failed to save submenu. Please try again.';
        console.error('Save submenu error:', err);
      }
    });
  }

  // ── Assign permissions to a Submenu ──────────────────────────────────────
  private assignPermissionsToSubmenu(submenuId: number, permissionIds: number[]): void {
    const payload = permissionIds.map(id => ({
      metadata: {},
      permission_id: id
    }));

    this.http.post(
      `${this.apiUrl}/api/permissions/submenu/${submenuId}`,
      payload
    ).subscribe({
      next: () => {
        this.finishSaveSuccess();
      },
      error: (err) => {
        console.error('Submenu permission assignment error:', err);
        this.finishSaveSuccess();
      }
    });
  }

  private finishSaveSuccess(): void {
    this.isSaving = false;
    this.saveSuccess = true;
    setTimeout(() => this.backToList(), 900);
  }

  // ── Parent menu change handler ────────────────────────────────────────────
  onParentMenuChange(menuId: string | number | null): void {
    if (!menuId) {
      this.form.parent_menu_route = '';
      return;
    }
    const selected = this.allMenus.find(m => m.id === Number(menuId));
    if (selected) {
      this.form.parent_menu_route = selected.route_path;
      this.form.route_path = selected.route_path;
    }
  }

  // ── Filters ───────────────────────────────────────────────────────────────
  applyFilters(): void {
    const q = this.searchQuery.toLowerCase().trim();
    this.filteredMenus = this.allMenus.filter(m => {
      const matchSearch = !q
        || m.name.toLowerCase().includes(q)
        || m.code.toLowerCase().includes(q)
        || m.route_path.toLowerCase().includes(q);
      const matchStatus = this.statusFilter === 'all'
        || (this.statusFilter === 'active' && m.is_active)
        || (this.statusFilter === 'inactive' && !m.is_active);
      return matchSearch && matchStatus;
    });
  }

  resetFilters(): void {
    this.searchQuery = '';
    this.statusFilter = 'all';
    this.applyFilters();
  }

  filteredParentMenus(): Menu[] {
    return this.filteredMenus.filter(m => m.parent_menu_id === null);
  }

  getChildren(parentId: number): Menu[] {
    return this.filteredMenus.filter(m => m.parent_menu_id === parentId);
  }

  toggleExpand(id: number): void {
    this.expandedIds.has(id) ? this.expandedIds.delete(id) : this.expandedIds.add(id);
  }

  get activeCount(): number { return this.allMenus.filter(m => m.is_active).length; }
  get parentCount(): number { return this.allMenus.filter(m => m.parent_menu_id === null).length; }
  get submenuCount(): number { return this.allMenus.filter(m => m.parent_menu_id !== null).length; }
  get parentMenus(): Menu[] { return this.allMenus.filter(m => m.parent_menu_id === null); }

  // ── View navigation ───────────────────────────────────────────────────────
  openTypeSelector(): void {
    this.selectedType = null;
    this.showTypeSelector = true;
  }

  closeTypeSelector(): void {
    this.showTypeSelector = false;
    this.selectedType = null;
  }

  selectType(type: 'menu' | 'submenu'): void {
    this.selectedType = type;
  }

  confirmTypeAndNavigate(): void {
    if (!this.selectedType) return;
    this.form = this.emptyForm();
    this.permissionSearch = '';
    this.saveError = '';
    this.saveSuccess = false;
    this.currentView = this.selectedType === 'menu' ? 'add-menu' : 'add-submenu';
    this.showTypeSelector = false;
    this.selectedType = null;
  }

  openEdit(menu: Menu): void {
    this.editingMenu = menu;
    this.saveError = '';
    this.saveSuccess = false;
    this.permissionSearch = '';
    this.form = {
      name: menu.name,
      code: menu.code,
      route_path: menu.route_path,
      icon: menu.icon,
      display_order: menu.display_order,
      is_active: menu.is_active,
      module_id: menu.module_id,
      parent_menu_id: menu.parent_menu_id ?? null,
      parent_menu_route: menu.parent_menu_id
        ? (this.allMenus.find(m => m.id === menu.parent_menu_id)?.route_path || '')
        : '',
      hide_no_perm: false,
      show_in_sidebar: true,
      selectedPermissionIds: []
    };
    this.currentView = 'edit';
  }

  backToList(): void {
    this.currentView = 'list';
    this.editingMenu = null;
    this.form = this.emptyForm();
    this.permissionSearch = '';
    this.saveError = '';
    this.saveSuccess = false;
    this.isSaving = false;
  }

  // ── Icon helpers ──────────────────────────────────────────────────────────
  filteredIconKeys(): string[] {
    const q = this.iconSearch.toLowerCase().trim();
    const keys = Object.keys(this.iconDefs);
    return q ? keys.filter(k => k.includes(q)) : keys;
  }

  getIconSvgPath(key: string): string {
    return this.iconDefs[key] ?? this.iconDefs['menu'];
  }

  getBreadcrumb(): string {
    const map: Record<string, string> = {
      'add-menu': 'Add Menu',
      'add-submenu': 'Add Submenu',
      'edit': 'Edit Item',
      'list': ''
    };
    return map[this.currentView] ?? '';
  }

  private emptyForm(): MenuForm {
    return {
      name: '',
      code: '',
      route_path: '',
      icon: 'menu',
      display_order: 1,
      is_active: true,
      module_id: null,
      parent_menu_id: null,
      parent_menu_route: '',
      hide_no_perm: true,
      show_in_sidebar: true,
      selectedPermissionIds: []
    };
  }
}