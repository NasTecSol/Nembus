import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';


interface Module {
  id: string;
  name: string;
  description: string;
  displayOrder: number;
  icon: string;
  isActive: boolean;
}

interface Menu {
  id: string;
  name: string;
  routePath: string;
  displayOrder: number;
  icon: string;
  isActive: boolean;
}

interface Submenu {
  id: string;
  name: string;
  routePath: string;
  displayOrder: number;
  icon: string;
  isActive: boolean;
  isPage: boolean;
  addPagination: boolean;
  addFilters: boolean;
}

type OverlayType = 'module' | 'menu' | 'submenu' | null;
type OverlayMode = 'add' | 'edit';


@Component({
  selector: 'app-home',
  imports: [CommonModule, FormsModule],
  templateUrl: './home.component.html',
  styleUrl: './home.component.scss'

})
export class HomeComponent {
  activeTab: 'permissions' | 'shifts' | 'employees' = 'permissions';
  activeSideMenu: string = 'menu-modules';
  
  overlayType: OverlayType = null;
  overlayMode: OverlayMode = 'add';
  
  // Form data for overlays
  moduleForm: Module = this.getEmptyModule();
  menuForm: Menu = this.getEmptyMenu();
  submenuForm: Submenu = this.getEmptySubmenu();
  
  // File upload
  selectedFile: File | null = null;
  selectedFileName: string = '';

  constructor() {}

  // Tab methods
  setActiveTab(tab: 'permissions' | 'shifts' | 'employees') {
    this.activeTab = tab;
  }

  // Side menu methods
  setActiveSideMenu(menu: string) {
    this.activeSideMenu = menu;
  }

  // Overlay methods
  openAddModuleOverlay() {
    this.overlayType = 'module';
    this.overlayMode = 'add';
    this.moduleForm = this.getEmptyModule();
    this.selectedFile = null;
    this.selectedFileName = '';
  }

  openEditModuleOverlay() {
    this.overlayType = 'module';
    this.overlayMode = 'edit';
    // In real scenario, load existing data
    this.selectedFile = null;
    this.selectedFileName = '';
  }

  openAddMenuOverlay() {
    this.overlayType = 'menu';
    this.overlayMode = 'add';
    this.menuForm = this.getEmptyMenu();
    this.selectedFile = null;
    this.selectedFileName = '';
  }

  openAddSubmenuOverlay() {
    this.overlayType = 'submenu';
    this.overlayMode = 'add';
    this.submenuForm = this.getEmptySubmenu();
    this.selectedFile = null;
    this.selectedFileName = '';
  }

  openEditSubmenuOverlay() {
    this.overlayType = 'submenu';
    this.overlayMode = 'edit';
    // In real scenario, load existing data
    this.selectedFile = null;
    this.selectedFileName = '';
  }

  closeOverlay() {
    this.overlayType = null;
    this.selectedFile = null;
    this.selectedFileName = '';
  }

  // File upload handler
  onFileSelected(event: any) {
    const file = event.target.files[0];
    if (file) {
      this.selectedFile = file;
      this.selectedFileName = file.name;
    }
  }

  triggerFileInput(inputId: string) {
    const fileInput = document.getElementById(inputId) as HTMLInputElement;
    if (fileInput) {
      fileInput.click();
    }
  }

  // Form submission methods
  submitModule() {
    console.log('Module submitted:', this.moduleForm);
    // Add your submission logic here
    this.closeOverlay();
  }

  submitMenu() {
    console.log('Menu submitted:', this.menuForm);
    // Add your submission logic here
    this.closeOverlay();
  }

  submitSubmenu() {
    console.log('Submenu submitted:', this.submenuForm);
    // Add your submission logic here
    this.closeOverlay();
  }

  saveChanges() {
    console.log('Changes saved');
    // Add your save logic here
    this.closeOverlay();
  }

  // Helper methods to get empty form objects
  private getEmptyModule(): Module {
    return {
      id: '',
      name: '',
      description: '',
      displayOrder: 0,
      icon: '',
      isActive: false
    };
  }

  private getEmptyMenu(): Menu {
    return {
      id: '',
      name: '',
      routePath: '',
      displayOrder: 0,
      icon: '',
      isActive: false
    };
  }

  private getEmptySubmenu(): Submenu {
    return {
      id: '',
      name: '',
      routePath: '',
      displayOrder: 0,
      icon: '',
      isActive: false,
      isPage: false,
      addPagination: false,
      addFilters: false
    };
  }

  // Get overlay title
  getOverlayTitle(): string {
    if (!this.overlayType) return '';
    
    const titles: Record<string, Record<string, string>> = {
      module: {
        add: 'Add Module',
        edit: 'Edit Module'
      },
      menu: {
        add: 'Add Menu',
        edit: 'Edit Menu'
      },
      submenu: {
        add: 'Add Submenu',
        edit: 'Edit Submenu'
      }
    };
    
    return titles[this.overlayType][this.overlayMode];
  }
}
