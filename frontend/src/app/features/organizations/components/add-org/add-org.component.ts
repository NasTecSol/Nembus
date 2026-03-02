import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';
import { ActivatedRoute, Router } from '@angular/router';
import { OrganizationService, Organization } from '../../../../core/services/organization.service';
import { ToastyService } from '../../../../core/services/toasty.service';

export interface OrgFormErrors {
  name?: string;
  code?: string;
  currency_code?: string;
}
@Component({
  selector: 'app-add-org',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslateModule],
  templateUrl: './add-org.component.html'
})
export class AddOrgComponent implements OnInit {
submenuId!: number;
submenuName!: string;
submenuCode!: string;

  // ── Mode flags ──
  isEditMode = false;
  isLoading = false;
  isSaving = false;

  // ── Form model ──
  form: Organization = {
    name: '',
    code: '',
    legal_name: null,
    tax_id: null,
    currency_code: '',
    fiscal_year_variant: null,
    is_active: true
  };

  // ── Inline validation errors ──
 errors: OrgFormErrors = {};

  // ── Dropdown data ──
  currencies = [
    { code: 'SAR', name: 'Saudi Riyal' },
    { code: 'USD', name: 'US Dollar' },
    { code: 'EUR', name: 'Euro' },
    { code: 'GBP', name: 'British Pound' },
    { code: 'AED', name: 'UAE Dirham' },
    { code: 'PKR', name: 'Pakistani Rupee' },
    { code: 'EGP', name: 'Egyptian Pound' },
    { code: 'KWD', name: 'Kuwaiti Dinar' },
    { code: 'BHD', name: 'Bahraini Dinar' },
    { code: 'QAR', name: 'Qatari Riyal' },
    { code: 'OMR', name: 'Omani Rial' },
    { code: 'JOD', name: 'Jordanian Dinar' }
  ];

  fiscalYearOptions = [
    { value: 'Jan-Dec', label: 'January – December' },
    { value: 'Apr-Mar', label: 'April – March' },
    { value: 'Jul-Jun', label: 'July – June' },
    { value: 'Oct-Sep', label: 'October – September' }
  ];

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private organizationService: OrganizationService,
    private toasty: ToastyService
  ) {}

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');

    if (id) {
      // ── Edit mode: fetch existing org ──
this.isEditMode = true;
  this.submenuId = 8;
  this.submenuName = 'Organization List';
  this.submenuCode = 'org_list';

  this.isLoading = true;  // ← ADD THIS LINE
  this.organizationService.getOrganization(+id).subscribe({
    next: (org) => {
      this.form = { ...org };
      this.isLoading = false;
    },
        
        error: (err) => {
          console.error('Failed to load organization:', err);
          this.isLoading = false;
          this.navigateBack();
        }
      });
    } else{
    this.isEditMode = false;
    this.submenuId = 9;
    this.submenuName = 'Add Organization';
    this.submenuCode = 'add_org';
    }
    
    // else → create mode, form stays as defaults
  }

  // ── Validation ──
  validate(): boolean {
    this.errors = {};

    if (!this.form.name?.trim()) {
      this.errors['name'] = 'Organization name is required';
    }
    if (!this.form.code?.trim()) {
      this.errors['code'] = 'Organization code is required';
    }
    if (!this.form.currency_code) {
      this.errors['currency_code'] = 'Please select a currency';
    }

    return Object.keys(this.errors).length === 0;
  }

  // ── Submit ──
  onSubmit(): void {
    if (!this.validate()) return;
    if (this.isSaving) return;

    this.isSaving = true;

    if (this.isEditMode) {
      // ── Update ──
      this.organizationService.updateOrganization(this.form.id!, this.form).subscribe({
    next: () => {
      this.toasty.success('Organization updated successfully');
      this.navigateBack();
    },
        error: (err) => {
          console.error('Failed to update organization:', err);
          this.isSaving = false;
        }
      });
    } else {
      // ── Create ──
      this.organizationService.createOrganization(this.form).subscribe({
    next: () => {
      this.toasty.success('Organization created successfully');
      this.navigateBack();
    },
        error: (err) => {
          console.error('Failed to create organization:', err);
          this.isSaving = false;
        }
      });
    }
  }

  // ── Navigation ──
  navigateBack(): void {
    this.router.navigate(['admin/organizations/list']);
  }
}