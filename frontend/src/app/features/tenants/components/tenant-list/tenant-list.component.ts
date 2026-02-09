import { CommonModule } from "@angular/common";
import { Component, OnInit } from "@angular/core";
import { Router } from "@angular/router";
import { FormsModule } from "@angular/forms";
import { TranslateModule } from "@ngx-translate/core";
import { TenantService, Tenant } from "../../../../core/services/tenant.service";
import { ToastrService } from "ngx-toastr";
import { environment } from "../../../../../environments/environment";

@Component({
  selector: 'app-tenant-list',
  standalone: true,
  imports: [CommonModule, TranslateModule, FormsModule],
  templateUrl: './tenant-list.component.html',
  styleUrl: './tenant-list.component.scss'
})
export class TenantListComponent implements OnInit {
  tenants: Tenant[] = [];
  filteredTenants: Tenant[] = [];
  searchQuery: string = "";
  statusFilter: string = "all";
  showConfirmModal: boolean = false;
  selectedTenant: Tenant | null = null;
  selectedIds: number[] = [];
  dropdownOpenId: number | null = null;
  submenuId = 5;
  submenuName = 'Tenant List';
  submenuCode = 'tenant_list';
  isLoading: boolean = false;

  constructor(
    private router: Router,
    private tenantService: TenantService,
    private toasty: ToastrService
  ) {}

  async ngOnInit(): Promise<void> {
    await this.loadTenants();
  }

  async loadTenants(): Promise<void> {
    this.isLoading = true;
    try {
      // Fetch all tenants (active and inactive)
      this.tenants = await this.tenantService.getAllTenants();
      this.filteredTenants = [...this.tenants];
      console.log('✅ Loaded tenants:', this.tenants);
    } catch (error) {
      console.error('❌ Error loading tenants:', error);
      this.toasty.error('Failed to load tenants');
      
      // Fallback to mock data on error
      this.loadMockTenants();
    } finally {
      this.isLoading = false;
    }
  }

  loadMockTenants(): void {
    // Mock data fallback
    this.tenants = [
      {
        id: "550e8400-e29b-41d4-a716-446655440001",
        tenant_name: "Northstar Retail Group",
        slug: "northstar",
        db_conn_str: "postgresql://user:pass@localhost:5432/northstar_db",
        is_active: true,
        settings: {
          theme: "default",
          features: ["pos", "inventory", "reports"],
        },
        created_at: new Date("2024-01-15"),
        updated_at: new Date("2024-12-20"),
      },
      {
        id: "550e8400-e29b-41d4-a716-446655440002",
        tenant_name: "Urban Outfitters Local",
        slug: "urban-outfitters",
        db_conn_str: "postgresql://user:pass@localhost:5432/urban_db",
        is_active: true,
        settings: {
          theme: "modern",
          features: ["pos", "inventory"],
        },
        created_at: new Date("2024-02-10"),
        updated_at: new Date("2025-01-10"),
      },
      {
        id: "550e8400-e29b-41d4-a716-446655440003",
        tenant_name: "Mom & Pop Corner Store",
        slug: "mompop",
        db_conn_str: "postgresql://user:pass@localhost:5432/mompop_db",
        is_active: false,
        settings: {
          theme: "classic",
          features: ["pos"],
        },
        created_at: new Date("2023-11-05"),
        updated_at: new Date("2024-06-15"),
      },
    ];

    this.filteredTenants = [...this.tenants];
    this.toasty.warning('Using demo data. API not available.', 'Demo Mode');
  }

  filterTenants(): void {
    let filtered = [...this.tenants];

    // Filter by search query
    if (this.searchQuery.trim()) {
      const query = this.searchQuery.toLowerCase();
      filtered = filtered.filter(
        (tenant) =>
          tenant.tenant_name.toLowerCase().includes(query) ||
          tenant.slug.toLowerCase().includes(query)
      );
    }

    // Filter by status
    if (this.statusFilter !== "all") {
      const isActive = this.statusFilter === "active";
      filtered = filtered.filter((tenant) => tenant.is_active === isActive);
    }

    this.filteredTenants = filtered;
  }

  getTenantInitial(name: string): string {
    return name.charAt(0).toUpperCase();
  }

  formatDate(date: Date): string {
    const now = new Date();
    const diff = now.getTime() - new Date(date).getTime();
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));

    if (days === 0) return "Today";
    if (days === 1) return "Yesterday";
    if (days < 7) return `${days} days ago`;
    if (days < 30) return `${Math.floor(days / 7)} weeks ago`;
    if (days < 365) return `${Math.floor(days / 30)} months ago`;
    return `${Math.floor(days / 365)} years ago`;
  }

  getTotalTenants(): number {
    return this.tenants.length;
  }

  getActiveTenants(): number {
    return this.tenants.filter((t) => t.is_active).length;
  }

  getInactiveTenants(): number {
    return this.tenants.filter((t) => !t.is_active).length;
  }

  getTenantsThisMonth(): number {
    const now = new Date();
    const thisMonth = now.getMonth();
    const thisYear = now.getFullYear();

    return this.tenants.filter((t) => {
      const created = new Date(t.created_at);
      return (
        created.getMonth() === thisMonth && created.getFullYear() === thisYear
      );
    }).length;
  }

  addTenant(): void {
    // Navigate to add tenant page
    this.router.navigate(["/admin/tenants/new"]);
  }

  editConfiguration(tenantId: string): void {
    // Navigate to tenant configuration page
    this.router.navigate(["/admin/tenants/config", tenantId]);
  }

  toggleTenantStatus(tenant: Tenant): void {
    this.selectedTenant = tenant;
    this.showConfirmModal = true;
  }
  // Extract base domain from environment.baseUrl
  getBaseDomain(): string {
    try {
      const url = new URL(environment.baseUrl);
      return url.hostname; // Returns 'nembus.nashrms.com'
    } catch (error) {
      console.error('Error parsing baseUrl:', error);
      return 'nembus.nashrms.com'; // Fallback
    }
  }
    getTenantEnvironmentUrl(tenant: Tenant): string {
    const baseDomain = this.getBaseDomain();
    return `${tenant.tenant_name}.${baseDomain}`;
  }

async confirmStatusChange(): Promise<void> {
  if (this.selectedTenant) {
    const wasActive = this.selectedTenant.is_active;
    const actionText = wasActive ? 'deactivating' : 'activating';
    
    try {
      console.log(`🔵 ${actionText} tenant:`, this.selectedTenant.id);
      console.log('🔵 Current tenant data:', this.selectedTenant);

      // Encode settings back to base64 if needed
      let settingsToSend = this.selectedTenant.settings;
      if (typeof settingsToSend === 'object') {
        settingsToSend = btoa(JSON.stringify(settingsToSend));
      }

      // Send the complete tenant object with updated status
      const fullPayload = {
        tenant_name: this.selectedTenant.tenant_name,
        slug: this.selectedTenant.slug,
        db_conn_str: this.selectedTenant.db_conn_str,
        is_active: !wasActive, // Toggle the status
        settings: settingsToSend
      };

      console.log('🔵 Sending full payload:', fullPayload);

      await this.tenantService.updateTenant(this.selectedTenant.id, fullPayload);

      // Show success message
      const successText = wasActive ? 'deactivated' : 'activated';
      this.toasty.success(`Tenant ${successText} successfully`);

      // Reload tenants to get fresh data
      await this.loadTenants();

    } catch (error) {
      console.error(`❌ Error ${actionText} tenant:`, error);
      this.toasty.error(`Failed to ${actionText.replace('ing', 'e')} tenant`);
    } finally {
      this.showConfirmModal = false;
      this.selectedTenant = null;
      this.filterTenants();
    }
  }
}

  cancelStatusChange(): void {
    this.showConfirmModal = false;
    this.selectedTenant = null;
  }
}