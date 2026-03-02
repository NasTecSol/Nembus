import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';
import { FormsModule, ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { TenantService, Tenant } from '../../../../core/services/tenant.service';
import { ToastrService } from 'ngx-toastr';

@Component({
  selector: 'app-tenant-config',
  standalone: true,
  imports: [CommonModule, TranslateModule, FormsModule, ReactiveFormsModule],
  templateUrl: './tenant-config.component.html',
  styleUrl: './tenant-config.component.scss'
})
export class TenantConfigComponent implements OnInit {
  submenuId = 7;
  submenuName = 'Tenant Configuration';
  submenuCode = 'tenant_config';

  tenantForm: FormGroup;
  isSubmitting: boolean = false;
  isLoading: boolean = false;
  
  // Tenant selection
  allTenants: Tenant[] = [];
  selectedTenantId: string | null = null;
  currentTenant: Tenant | null = null;
  
  // Determines if tenant was pre-selected from route
  isPreSelected: boolean = false;

  // Settings options
  planOptions = [
    { value: 'basic', label: 'Basic' },
    { value: 'professional', label: 'Professional' },
    { value: 'enterprise', label: 'Enterprise' }
  ];

  featureOptions = [
    { value: 'pos', label: 'Point of Sale' },
    { value: 'inventory', label: 'Inventory Management' },
    { value: 'reports', label: 'Reports & Analytics' },
    { value: 'basic_analytics', label: 'Basic Analytics' },
    { value: 'advanced_analytics', label: 'Advanced Analytics' },
    { value: 'crm', label: 'Customer Relationship Management' },
    { value: 'multi_store', label: 'Multi-Store Support' }
  ];

  selectedFeatures: string[] = [];

  constructor(
    private fb: FormBuilder,
    private route: ActivatedRoute,
    private router: Router,
    private tenantService: TenantService,
    private toasty: ToastrService
  ) {
    this.tenantForm = this.fb.group({
  tenant_name: ['', [Validators.required, Validators.minLength(3)]],
  slug: ['Tenant-Code', [Validators.required, Validators.pattern('^[a-z0-9-]+$')]],
  db_conn_str: ['postgres://nembus_admin_user:Nembus_Client2023@postgres:5432/', Validators.required],
  plan: ['professional', Validators.required],
  max_users: [100, [Validators.required, Validators.min(1)]],
  is_active: [true]
});
  }

  async ngOnInit(): Promise<void> {
    // Check if tenant ID is in route params
    this.route.params.subscribe(async params => {
      const tenantId = params['id'];
      
      if (tenantId) {
        // Pre-selected from tenant list
        this.isPreSelected = true;
        this.selectedTenantId = tenantId;
        await this.loadTenantData(tenantId);
      } else {
        // No pre-selection, load all tenants for dropdown
        await this.loadAllTenants();
      }
    });
  }

  async loadAllTenants(): Promise<void> {
    this.isLoading = true;
    try {
      this.allTenants = await this.tenantService.getAllTenants();
      console.log('✅ Loaded tenants for selection:', this.allTenants);
    } catch (error) {
      console.error('❌ Error loading tenants:', error);
      this.toasty.error('Failed to load tenants');
    } finally {
      this.isLoading = false;
    }
  }

  async onTenantSelect(tenantId: string): Promise<void> {
    this.selectedTenantId = tenantId;
    await this.loadTenantData(tenantId);
  }

async loadTenantData(tenantId: string): Promise<void> {
  this.isLoading = true;
  try {
    // Find tenant from loaded list or fetch all if not available
    if (this.allTenants.length === 0) {
      await this.loadAllTenants();
    }

    this.currentTenant = this.allTenants.find(t => t.id === tenantId) || null;

    if (!this.currentTenant) {
      this.toasty.error('Tenant not found');
      return;
    }

    console.log('✅ Loading tenant data:', this.currentTenant);

    // Decode settings
    const settings = this.currentTenant.settings || {};
    this.selectedFeatures = settings.features || [];

    // Populate form with complete connection string
    this.tenantForm.patchValue({
      tenant_name: this.currentTenant.tenant_name,
      slug: this.currentTenant.slug,
      db_conn_str: this.currentTenant.db_conn_str, // ← Use complete connection string
      plan: settings.plan || 'professional',
      max_users: settings.max_users || 100,
      is_active: this.currentTenant.is_active
    });

  } catch (error) {
    console.error('❌ Error loading tenant data:', error);
    this.toasty.error('Failed to load tenant data');
  } finally {
    this.isLoading = false;
  }
}

parseDbConnString(connStr: string): any {
  // Parse: postgres://username:password@host:port/database?sslmode=disable
  try {
    const regex = /postgres:\/\/([^:]+):([^@]+)@([^:]+):(\d+)\/([^?]+)/;
    const match = connStr.match(regex);

    if (match) {
      return {
        username: match[1],
        password: match[2],
        host: match[3],
        port: match[4],
        database: match[5]
      };
    }
  } catch (error) {
    console.error('Error parsing connection string:', error);
  }

  return {
    username: 'nembus_admin_user',
    password: '',
    host: 'postgres',
    port: '5432',
    database: ''
  };
}

  toggleFeature(feature: string): void {
    const index = this.selectedFeatures.indexOf(feature);
    if (index > -1) {
      this.selectedFeatures.splice(index, 1);
    } else {
      this.selectedFeatures.push(feature);
    }
  }

  isFeatureSelected(feature: string): boolean {
    return this.selectedFeatures.includes(feature);
  }

  get isFormValid(): boolean {
    return this.tenantForm.valid && this.selectedFeatures.length > 0 && this.selectedTenantId !== null;
  }

async onSubmit(): Promise<void> {
  if (!this.isFormValid || this.isSubmitting || !this.selectedTenantId) {
    return;
  }

  this.isSubmitting = true;

  try {
    const formValue = this.tenantForm.value;

    // Build settings object
    const settings = {
      plan: formValue.plan,
      features: this.selectedFeatures,
      max_users: formValue.max_users
    };

    // Encode settings to base64
    const settingsBase64 = btoa(JSON.stringify(settings));

    // Prepare COMPLETE tenant payload (all fields required)
    const fullPayload = {
      tenant_name: formValue.tenant_name,
      slug: formValue.slug,
      db_conn_str: formValue.db_conn_str, // Use complete connection string from form
      is_active: formValue.is_active,
      settings: settingsBase64
    };

    console.log('🔵 Updating tenant:', this.selectedTenantId);
    console.log('🔵 Full update payload:', JSON.stringify(fullPayload, null, 2));

    // Call API with complete object
    await this.tenantService.updateTenant(this.selectedTenantId, fullPayload);

    this.toasty.success('Tenant configuration updated successfully');

    // Navigate back to tenant list
    this.router.navigate(['/admin/tenants']);

  } catch (error: any) {
    console.error('❌ Error updating tenant:', error);
    const errorMessage = error.error?.message || error.message || 'Failed to update tenant';
    this.toasty.error(errorMessage);
  } finally {
    this.isSubmitting = false;
  }
}

  onCancel(): void {
    this.router.navigate(['/admin/tenants']);
  }
}