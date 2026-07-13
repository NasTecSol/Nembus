import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';
import { NavigationService, ModuleItem, CreateModulePayload } from '../../../../core/services/navigation.service';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../../../../environments/environment';

export interface NewModuleForm {
  name: string;
  code: string;
  description: string;
  icon: string;
  displayOrder: number;
  isActive: boolean;
  permissions: Record<string, boolean>;
}

// ── API payload / response shapes ──────────────────────────────────────────
interface CreatePermissionPayload {
  code: string;
  description: string;
  metadata: Record<string, unknown>;
  name: string;
}

interface CreatedPermission {
  id: number;
  name: string;
  code: string;
  description: string;
  metadata: Record<string, unknown>;
  created_at: string;
}

interface AssignPermissionPayload {
  metadata: Record<string, unknown>;
  permission_id: number;
}

@Component({
  selector: 'app-module-list',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslateModule],
  templateUrl: './module-list.component.html',
  styleUrl: './module-list.component.scss'
})
export class ModuleListComponent implements OnInit {

  private apiUrl = environment.baseUrl;

  // ── List view state ──────────────────────────────────────────────────────
  modules: ModuleItem[] = [];
  isLoading = true;
  errorMessage = '';

  // ── View toggle ──────────────────────────────────────────────────────────
  showCreateForm = false;

  // ── Save state ───────────────────────────────────────────────────────────
  isCreating = false;
  createError = '';
  /** Human-readable step label shown during the 3-step save flow */
  savingStep = '';

  // ── Icons ────────────────────────────────────────────────────────────────
  availableIcons = [
    { name: 'inventory_2', label: 'Inventory' },
    { name: 'bar_chart', label: 'Chart' },
    { name: 'groups', label: 'People' },
    { name: 'corporate_fare', label: 'Corporate' },
    { name: 'store', label: 'Store' },
    { name: 'account_balance', label: 'Finance' },
    { name: 'shopping_cart', label: 'Shopping' },
    { name: 'trending_up', label: 'Trending' },
    { name: 'settings', label: 'Settings' },
    { name: 'dashboard', label: 'Dashboard' },
    { name: 'point_of_sale', label: 'POS' },
    { name: 'local_atm', label: 'Cashier' },
    { name: 'manage_accounts', label: 'Users' },
    { name: 'security', label: 'Security' },
    { name: 'analytics', label: 'Analytics' },
    { name: 'sell', label: 'Sales' },
  ];

  // ── Permissions ───────────────────────────────────────────────────────────
  permissionKeys: { key: string; label: string }[] = [
    { key: 'view', label: ':view' },
    { key: 'manage', label: ':manage' },
    { key: 'delete', label: ':delete' },
    { key: 'export', label: ':export' },
    { key: 'add', label: ':add' },
    { key: 'list', label: ':list' },
    { key: 'configure', label: ':configure' },
  ];

  getPermission(key: string): boolean {
    return this.newModule.permissions[key] ?? false;
  }

  setPermission(key: string, value: boolean): void {
    this.newModule.permissions[key] = value;
  }

  getPermissionLabel(label: string): string {
    return this.newModule.name ? (this.newModule.name.trim().toLowerCase().replace(/\s+/g, '_') + label) : label;
  }

  getPermissionCode(key: string): string {
    const name = this.newModule.name ? this.newModule.name.trim().toLowerCase().replace(/\s+/g, '_') : '';
    return `${name}:${key}`;
  }

  get hasSelectedPermissions(): boolean {
    return this.permissionKeys.some(p => this.getPermission(p.key));
  }

  newModule: NewModuleForm = this.getEmptyForm();

  constructor(
    private navigationService: NavigationService,
    private http: HttpClient
  ) { }

  ngOnInit(): void {
    this.loadModules();
  }

  // ── List helpers ─────────────────────────────────────────────────────────
  loadModules(): void {
    this.isLoading = true;
    this.errorMessage = '';
    this.navigationService.getModules().subscribe({
      next: (data) => {
        this.modules = data.sort((a, b) => a.display_order - b.display_order);
        this.isLoading = false;
      },
      error: (err) => {
        console.error('Failed to load modules:', err);
        this.errorMessage = 'Failed to load modules. Please try again.';
        this.isLoading = false;
      }
    });
  }

  get activeModulesCount(): number {
    return this.modules.filter(m => m.is_active).length;
  }

  /**
   * Resolves any icon string (from DB or picker) to a valid Material Icons ligature.
   *
   * Strategy:
   *  1. Check against a whitelist of confirmed-valid Material icon names.
   *     If it matches, return as-is.
   *  2. Try the comprehensive alias/slug lookup map (handles kebab-case, human
   *     readable names, legacy values, etc.).
   *  3. Fall back to "widgets" so the UI never shows a broken glyph.
   */
  getMaterialIcon(iconName: string): string {
    if (!iconName) return 'widgets';

    // ── Step 1: whitelist of valid Material Icons names we know about ────────
    const knownMaterialIcons = new Set([
      'dashboard', 'home', 'settings', 'person', 'group', 'groups',
      'manage_accounts', 'admin_panel_settings', 'business', 'corporate_fare',
      'store', 'point_of_sale', 'local_atm', 'inventory_2', 'category',
      'people', 'local_shipping', 'bar_chart', 'analytics', 'trending_up',
      'account_balance', 'shopping_cart', 'sell', 'security', 'widgets',
      'info', 'palette', 'lock', 'add', 'add_circle', 'arrow_back',
      'more_vert', 'download', 'filter_list', 'swap_vert', 'refresh',
      'error_outline', 'layers_clear', 'chevron_left', 'chevron_right',
      'account_tree', 'sort', 'visibility', 'construction', 'handyman',
      'build', 'build_circle', 'engineering', 'miscellaneous_services',
      'receipt', 'receipt_long', 'payments', 'money', 'attach_money',
      'currency_exchange', 'storefront', 'inventory', 'warehouse',
      'package_2', 'deployed_code', 'hub', 'lan', 'dns', 'cloud',
      'supervisor_account', 'badge', 'contact_page', 'contacts',
      'person_add', 'person_search', 'support_agent',
    ]);

    if (knownMaterialIcons.has(iconName)) {
      return iconName;
    }

    // ── Step 2: comprehensive alias / slug lookup ────────────────────────────
    const iconMap: Record<string, string> = {
      // generic
      'dashboard': 'dashboard',
      'home': 'home',
      'settings': 'settings',
      'admin': 'admin_panel_settings',
      'widgets': 'widgets',
      'construction': 'construction',
      'build': 'build',
      'wrench': 'build',          // ← Tenant Management
      'tool': 'build',
      'tools': 'build',
      'gear': 'settings',
      'cog': 'settings',

      // users / accounts
      'users': 'manage_accounts',
      'user': 'manage_accounts',
      'user-management': 'manage_accounts',
      'manage-accounts': 'manage_accounts',
      'person': 'manage_accounts',
      'people-alt': 'manage_accounts',

      // organisations / tenants
      'tenants': 'business',
      'tenant': 'business',
      'tenant-management': 'business',
      'organizations': 'corporate_fare',
      'organisations': 'corporate_fare',
      'organization': 'corporate_fare',
      'organisation': 'corporate_fare',
      'organization-setup': 'corporate_fare',
      'building': 'business',       // ← Organization Setup
      'office': 'business',
      'corporate': 'corporate_fare',
      'company': 'business',

      // store
      'store': 'store',
      'stores': 'store',
      'store-management': 'store',
      'store-settings': 'settings',
      'shop': 'storefront',
      'storefront': 'storefront',

      // point of sale
      'pos': 'point_of_sale',
      'point-of-sale': 'point_of_sale',
      'pointofsale': 'point_of_sale',
      'terminal': 'point_of_sale',
      'cash-register': 'point_of_sale',

      // cashier
      'cashier': 'local_atm',
      'cashiers': 'local_atm',
      'cashier-operations': 'local_atm',
      'atm': 'local_atm',

      // inventory
      'inventory': 'inventory_2',
      'inventory-management': 'inventory_2',
      'inventory_management': 'inventory_2',
      'stock': 'inventory_2',
      'warehouse': 'warehouse',
      'package': 'inventory_2',   // ← Inventory Management
      'box': 'inventory_2',   // ← Inventory Management
      'cube': 'inventory_2',
      'archive': 'inventory_2',

      // products
      'products': 'category',
      'product': 'category',
      'product-catalog': 'category',
      'product_catalog': 'category',
      'catalog': 'category',
      'items': 'category',
      'tag': 'sell',
      'label': 'sell',

      // customers
      'customers': 'people',
      'customer': 'people',
      'customer-management': 'people',
      'customer_management': 'people',
      'clients': 'people',
      'client': 'people',
      'contact': 'contacts',
      'contacts': 'contacts',

      // suppliers
      'suppliers': 'local_shipping',
      'supplier': 'local_shipping',
      'supplier-management': 'local_shipping',
      'supplier_management': 'local_shipping',
      'vendor': 'local_shipping',
      'vendors': 'local_shipping',
      'procurement': 'local_shipping',
      'delivery': 'local_shipping',
      'truck': 'local_shipping',
      'shipping': 'local_shipping',

      // reports / analytics
      'reports': 'bar_chart',
      'report': 'bar_chart',
      'analytics': 'analytics',
      'chart': 'bar_chart',
      'bar-chart': 'bar_chart',
      'bar_chart': 'bar_chart',
      'trending-up': 'trending_up',
      'trending_up': 'trending_up',
      'graph': 'bar_chart',
      'statistics': 'bar_chart',
      'insights': 'analytics',

      // finance / hr
      'finance': 'account_balance',
      'accounting': 'account_balance',
      'bank': 'account_balance',
      'hr': 'groups',
      'human-resources': 'groups',
      'human_resources': 'groups',
      'payroll': 'payments',
      'salary': 'payments',
      'money': 'attach_money',
      'payments': 'payments',

      // misc commerce
      'purchase': 'shopping_cart',
      'purchases': 'shopping_cart',
      'shopping-cart': 'shopping_cart',
      'shopping_cart': 'shopping_cart',
      'cart': 'shopping_cart',
      'sales': 'sell',
      'sell': 'sell',
      'order': 'receipt_long',
      'orders': 'receipt_long',
      'receipt': 'receipt',

      // security / access
      'security': 'security',
      'permissions': 'lock',
      'roles': 'badge',
      'access': 'admin_panel_settings',
      'shield': 'security',

      // support / misc
      'support': 'support_agent',
      'help': 'help',
      'notification': 'notifications',
      'notifications': 'notifications',
    };

    // Try direct key, then lowercase, then kebab-converted, then underscore-converted
    const lower = iconName.toLowerCase();
    const kebab = lower.replace(/_/g, '-');
    const undersc = lower.replace(/-/g, '_');

    return (
      iconMap[iconName] ??
      iconMap[lower] ??
      iconMap[kebab] ??
      iconMap[undersc] ??
      'widgets'
    );
  }

  // ── View toggle ──────────────────────────────────────────────────────────
  openCreateForm(): void {
    this.newModule = this.getEmptyForm();
    this.createError = '';
    this.savingStep = '';
    this.showCreateForm = true;
  }

  cancelCreate(): void {
    this.showCreateForm = false;
    this.newModule = this.getEmptyForm();
    this.createError = '';
    this.savingStep = '';
  }

  // ── Main create flow: 3 sequential steps ─────────────────────────────────
  createModule(): void {
    if (!this.newModule.name.trim() || !this.newModule.code.trim()) {
      this.createError = 'Module Name and Code are required.';
      return;
    }

    // Collect which permission keys the user checked
    const selectedPermKeys = this.permissionKeys
      .filter(p => this.getPermission(p.key))
      .map(p => p.key);

    this.isCreating = true;
    this.createError = '';

    // ── STEP 1: Create the module ─────────────────────────────────────────
    this.savingStep = 'Step 1/3 — Creating module...';

    const modulePayload: CreateModulePayload = {
      code: this.newModule.code.trim().toUpperCase(),
      description: this.newModule.description.trim(),
      display_order: this.newModule.displayOrder,
      icon: this.newModule.icon,
      is_active: this.newModule.isActive,
      name: this.newModule.name.trim(),
    };

    this.navigationService.createModule(modulePayload).subscribe({
      next: (res: any) => {
        const moduleId: number = res?.data?.id ?? res?.id;
        if (!moduleId) {
          this.finishWithError('Module created but ID was not returned.');
          return;
        }

        // If no permissions were selected, skip steps 2 & 3
        if (selectedPermKeys.length === 0) {
          this.finishSuccess();
          return;
        }

        this.createPermissions(moduleId, selectedPermKeys);
      },
      error: (err) => {
        this.finishWithError(err?.error?.message || 'Failed to create module.');
      }
    });
  }

  // ── STEP 2: Create ALL permissions in a single batch request ─────────────
  private createPermissions(moduleId: number, selectedKeys: string[]): void {
    this.savingStep = 'Step 2/3 — Creating permissions...';

    const moduleName = this.newModule.name.trim().toLowerCase().replace(/\s+/g, '_');

    // Build a single array payload — one POST, all permissions at once
    const permPayloads: CreatePermissionPayload[] = selectedKeys.map(key => ({
      code: `${moduleName}:${key}`,
      name: `${this.newModule.name.trim()} ${key.charAt(0).toUpperCase() + key.slice(1)}`,
      description: `Can ${key} ${this.newModule.name.trim()}`,
      metadata: {},
    }));

    this.http
      .post<{ statusCode: number; message: string; data: CreatedPermission[] }>(
        `${this.apiUrl}/api/permissions`,
        permPayloads           // ← send the whole array in one request
      )
      .subscribe({
        next: (response) => {
          // The API returns data as an array of created permissions
          const permissionIds: number[] = (response?.data ?? [])
            .map((p) => p?.id)
            .filter((id): id is number => typeof id === 'number');

          if (permissionIds.length === 0) {
            this.finishWithError('Permissions created but no IDs were returned.');
            return;
          }

          this.assignPermissionsToModule(moduleId, permissionIds);
        },
        error: (err) => {
          this.finishWithError(err?.error?.message || 'Failed to create permissions.');
        }
      });
  }

  // ── STEP 3: Assign permissions to the module (send IDs as array) ──────────
  private assignPermissionsToModule(moduleId: number, permissionIds: number[]): void {
    this.savingStep = 'Step 3/3 — Assigning permissions to module...';

    // Send all permission IDs in a single request as an array
    const assignPayload: AssignPermissionPayload[] = permissionIds.map(id => ({
      metadata: {},
      permission_id: id,
    }));

    this.http
      .post(`${this.apiUrl}/api/permissions/module/${moduleId}`, assignPayload)
      .subscribe({
        next: () => {
          this.finishSuccess();
        },
        error: (err) => {
          this.finishWithError(err?.error?.message || 'Failed to assign permissions to module.');
        }
      });
  }

  // ── Flow finalizers ───────────────────────────────────────────────────────
  private finishSuccess(): void {
    this.isCreating = false;
    this.savingStep = '';
    this.showCreateForm = false;
    this.newModule = this.getEmptyForm();
    this.loadModules(); // refresh list
  }

  private finishWithError(message: string): void {
    this.isCreating = false;
    this.savingStep = '';
    this.createError = message;
  }

  // ── Form helpers ──────────────────────────────────────────────────────────
  private getEmptyForm(): NewModuleForm {
    return {
      name: '',
      code: '',
      description: '',
      icon: 'inventory_2',
      displayOrder: this.modules.length + 1,
      isActive: true,
      permissions: {
        view: true,
        manage: false,
        delete: false,
        export: false,
        add: false,
        list: false,
        configure: false,
      }
    };
  }

  get formProgress(): number {
    let filled = 0;
    if (this.newModule.name.trim()) filled++;
    if (this.newModule.code.trim()) filled++;
    if (this.newModule.description.trim()) filled++;
    if (this.newModule.icon) filled++;
    if (this.newModule.displayOrder > 0) filled++;
    filled++; // active status always has a value
    return Math.round((filled / 6) * 100);
  }

  get previewModules(): { name: string; icon: string; active: boolean; order: number }[] {
    const newEntry = {
      name: this.newModule.name || 'New Module',
      icon: this.newModule.icon || 'widgets',
      active: true,
      order: this.newModule.displayOrder,
    };
    const existing = this.modules
      .filter(m => m.display_order !== this.newModule.displayOrder)
      .slice(0, 3)
      .map(m => ({
        name: m.name,
        icon: this.getMaterialIcon(m.icon),
        active: false,
        order: m.display_order,
      }));
    return [...existing, newEntry].sort((a, b) => a.order - b.order).slice(0, 4);
  }
}
