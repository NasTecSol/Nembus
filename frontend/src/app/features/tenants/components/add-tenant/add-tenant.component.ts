import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';
import { FormsModule, ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { TenantService } from '../../../../core/services/tenant.service';
import { ToastrService } from 'ngx-toastr';

@Component({
  selector: 'app-add-tenant',
  standalone: true,
  imports: [CommonModule, TranslateModule, FormsModule, ReactiveFormsModule],
  templateUrl: './add-tenant.component.html',
  styleUrl: './add-tenant.component.scss'
})
export class AddTenantComponent {
  submenuId = 6;
  submenuName = 'Add Tenant';
  submenuCode = 'add_tenant';

  tenantForm: FormGroup;
  isSubmitting: boolean = false;

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
    private router: Router,
    private tenantService: TenantService,
    private toasty: ToastrService
  ) {
    this.tenantForm = this.fb.group({
  tenant_name: ['', [Validators.required, Validators.minLength(3)]],
  slug: ['', [Validators.required, Validators.pattern('^[a-z0-9-]+$')]],
  db_conn_str: ['postgres://nembus_admin_user:Nembus_Client2023@postgres:5432/', Validators.required], // ← Renamed
  plan: ['professional', Validators.required],
  max_users: [100, [Validators.required, Validators.min(1)]],
  is_active: [true]
});

    // Auto-generate slug from tenant name
this.tenantForm.get('tenant_name')?.valueChanges.subscribe(value => {
    if (value) {
      const slug = value.toLowerCase()
        .replace(/[^a-z0-9\s-]/g, '')
        .replace(/\s+/g, '-')
        .replace(/-+/g, '-')
        .trim();
      this.tenantForm.patchValue({ slug }, { emitEvent: false });
      
      // Auto-update db_conn_str with new database name
      const baseConnStr = 'postgres://nembus_admin_user:Nembus_Client2023@postgres:5432/';
      this.tenantForm.patchValue({ 
        db_conn_str: `${baseConnStr}${slug}?sslmode=disable` 
      }, { emitEvent: false });
    }
  });
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
    return this.tenantForm.valid && this.selectedFeatures.length > 0;
  }

// Updated onSubmit() in AddTenantComponent - send settings as plain object (no base64)
async onSubmit(): Promise<void> {
  if (!this.isFormValid || this.isSubmitting) {
    return;
  }

  this.isSubmitting = true;

  try {
    const formValue = this.tenantForm.value;

    // Build settings object (plain JSON - no encoding)
    const settings = {
      plan: formValue.plan,
      features: this.selectedFeatures,
      max_users: formValue.max_users
    };

    // Prepare payload - send settings directly as object
    const payload = {
      tenant_name: formValue.tenant_name,
      slug: formValue.slug,
      db_conn_str: formValue.db_conn_str,  // ← snake_case to match your desired/example payload
      is_active: formValue.is_active,
      settings: settings  // ← Plain object, no base64
    };

    console.log('🔵 Settings object (plain):', settings);
    console.log('🔵 Final payload:', JSON.stringify(payload, null, 2));

    // Call API
    const response = await this.tenantService.createTenant(payload);
    
    console.log('✅ Tenant created successfully:', response);
    this.toasty.success('Tenant created successfully');
    
    // Navigate back to tenant list
    this.router.navigate(['/admin/tenants']);

  } catch (error: any) {
    console.error('❌ Error creating tenant:', error);
    console.error('❌ Error details:', error.error);
    const errorMessage = error.error?.message || error.message || 'Failed to create tenant';
    this.toasty.error(errorMessage);
  } finally {
    this.isSubmitting = false;
  }
}
  onCancel(): void {
    this.router.navigate(['/admin/tenants']);
  }
}