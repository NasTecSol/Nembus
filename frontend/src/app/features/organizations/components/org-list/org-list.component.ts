import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';
import { Router } from '@angular/router';
import { OrganizationService, Organization } from '../../../../core/services/organization.service';
import { ToastyService } from '../../../../core/services/toasty.service';

@Component({
  selector: 'app-org-list',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslateModule],
  templateUrl: './org-list.component.html'
})
export class OrgListComponent implements OnInit {
  submenuId = 8;
  submenuName = 'Organization List';
  submenuCode = 'org_list';
  confirmDeleteOrg: Organization | null = null;
  organizations: Organization[] = [];
  loading: boolean = true;
  searchQuery: string = '';
  filterStatus: string = 'all';
  actionMenuOpen: number | null = null;

  constructor(
    private router: Router,
    private organizationService: OrganizationService,
    private toasty: ToastyService
  ) {}

  ngOnInit(): void {
    this.fetchOrganizations();
  }

  fetchOrganizations(): void {
    this.loading = true;
    const params = this.filterStatus === 'active' ? { is_active: true } : 
                   this.filterStatus === 'inactive' ? { is_active: false } : 
                   {};

    this.organizationService.getOrganizations(params).subscribe({
      next: (data) => {
        this.organizations = data;
        this.loading = false;
      },
      error: (error) => {
        console.error('Failed to load organizations:', error);
        this.loading = false;
      }
    });
  }

get filteredOrganizations(): Organization[] {
  if (!Array.isArray(this.organizations)) return [];

  return this.organizations.filter(org => {
    const matchesSearch = 
      org.name.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
      org.code.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
      (org.legal_name?.toLowerCase().includes(this.searchQuery.toLowerCase()) ?? false);

    const matchesStatus = 
      this.filterStatus === 'all' ||
      (this.filterStatus === 'active' && org.is_active) ||
      (this.filterStatus === 'inactive' && !org.is_active);

    return matchesSearch && matchesStatus;
  });
}
openDeleteConfirm(org: Organization): void {
  this.confirmDeleteOrg = org;
  this.closeDropdown();
}
  toggleDropdown(id: number): void {
    this.actionMenuOpen = this.actionMenuOpen === id ? null : id;
  }
performDelete(): void {
  if (!this.confirmDeleteOrg) return;

  this.organizationService.deleteOrganization(this.confirmDeleteOrg.id!).subscribe({
    next: () => {
      this.organizations = this.organizations.filter(o => o.id !== this.confirmDeleteOrg?.id);
      this.toasty.success('Organization deleted successfully');
      this.confirmDeleteOrg = null;
    },
    error: (error) => {
      console.error('Failed to delete organization:', error);
      this.toasty.error('Failed to delete organization');
      this.confirmDeleteOrg = null;
    }
  });
}
  closeDropdown(): void {
    this.actionMenuOpen = null;
  }

  isDropdownOpen(id: number): boolean {
    return this.actionMenuOpen === id;
  }

  getInitial(name: string): string {
    return name.charAt(0).toUpperCase();
  }

  onExport(): void {
    this.organizationService.exportOrganizations('xlsx').subscribe({
      next: (blob) => {
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `organizations_${new Date().toISOString().split('T')[0]}.xlsx`;
        link.click();
        window.URL.revokeObjectURL(url);
      },
      error: (error) => {
        console.error('Failed to export organizations:', error);
      }
    });
  }

  navigateToAddOrganization(): void {
    this.router.navigate(['admin/organizations/new']);
  }

  navigateToEditConfiguration(org: Organization): void {
    this.closeDropdown();
    this.router.navigate(['/admin/organizations', org.id, 'edit']);
  }

deleteOrganization(org: Organization): void {
  this.openDeleteConfirm(org);
}

  getFiscalYearDisplay(variant: string | null): string {
    const variants: { [key: string]: string } = {
      'Jan-Dec': 'January - December',
      'Apr-Mar': 'April - March',
      'Jul-Jun': 'July - June',
      'Oct-Sep': 'October - September'
    };
    return variant ? variants[variant] || variant : 'Default';
  }
}