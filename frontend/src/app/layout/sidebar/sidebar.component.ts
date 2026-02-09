import { CommonModule } from "@angular/common";
import { Component, OnInit } from "@angular/core";
import { Router, RouterLink, RouterLinkActive, NavigationEnd } from "@angular/router";
import { TranslateModule } from "@ngx-translate/core";
import { AuthService } from "../../core/services/auth.service";
import { filter } from "rxjs/operators";

interface MenuItem {
  id: string;
  name: string;
  path?: string;
  icon: string;
  expanded?: boolean;
  children?: MenuItem[];
}

interface NavigationSubmenu {
  has_access: boolean;
  submenu_id: number;
  submenu_code: string;
  submenu_icon: string;
  submenu_name: string;
  submenu_order: number;
  submenu_route: string;
  submenu_is_active: boolean;
  required_permissions: string[];
}

interface NavigationMenu {
  menu_id: number;
  menu_code: string;
  menu_icon: string;
  menu_name: string;
  menu_order: number;
  menu_route: string;
  menu_is_active: boolean;
  menu_permissions: string[];
  submenus: NavigationSubmenu[];
}

interface NavigationModule {
  module_id: number;
  module_name: string;
  module_code: string;
  module_description: string;
  module_icon: string;
  module_order: number;
  module_is_active: boolean;
  menus: NavigationMenu[];
}

interface NavigationResponse {
  statusCode: number;
  message: string;
  data: NavigationModule[];
}

@Component({
  selector: "app-sidebar",
  standalone: true,
  imports: [CommonModule, RouterLinkActive, RouterLink, TranslateModule],
  templateUrl: "./sidebar.component.html",
})
export class SidebarComponent implements OnInit {
  public menuItems: MenuItem[] = [];

  // Hover/floating submenu properties
  public submenuLeft: number = 0;
  public submenuTop: number = 0;
  public hoveredMenuId: string | null = null;
  public currentHoveredElement: HTMLElement | null = null;
  private submenuTimeout: any;
  private hoverTimeout: any;

  // Icon mapping from API icons to FontAwesome classes
  private iconMap: { [key: string]: string } = {
    'dashboard': 'fas fa-tachometer-alt',
    'building': 'fas fa-building',
    'briefcase': 'fas fa-briefcase',
    'users': 'fas fa-users',
    'store': 'fas fa-store',
    'shopping-cart': 'fas fa-shopping-cart',
    'user-check': 'fas fa-user-check',
    'package': 'fas fa-box',
    'box': 'fas fa-cube',
    'user-circle': 'fas fa-user-circle',
    'truck': 'fas fa-truck',
    'file-text': 'fas fa-file-alt',
    'shopping-bag': 'fas fa-shopping-bag',
    'bar-chart': 'fas fa-chart-bar',
    'settings': 'fas fa-cog',
    'home': 'fas fa-home',
    'trending-up': 'fas fa-chart-line',
    'layout': 'fas fa-th-large',
    'list': 'fas fa-list',
    'plus': 'fas fa-plus',
    'shield': 'fas fa-shield-alt',
    'grid': 'fas fa-th',
    'user-plus': 'fas fa-user-plus',
    'activity': 'fas fa-chart-area',
    'map-pin': 'fas fa-map-marker-alt',
    'credit-card': 'fas fa-credit-card',
    'monitor': 'fas fa-desktop',
    'calendar': 'fas fa-calendar-alt',
    'award': 'fas fa-trophy',
    'clock': 'fas fa-clock',
    'history': 'fas fa-history',
    'unlock': 'fas fa-unlock',
    'lock': 'fas fa-lock',
    'alert-triangle': 'fas fa-exclamation-triangle',
    'arrow-right': 'fas fa-arrow-right',
    'clipboard': 'fas fa-clipboard',
    'upload': 'fas fa-upload',
    'tag': 'fas fa-tag',
    'dollar-sign': 'fas fa-dollar-sign',
    'check-circle': 'fas fa-check-circle',
    'refresh-cw': 'fas fa-sync-alt',
    'percent': 'fas fa-percent',
    'eye': 'fas fa-eye',
    'menu': 'fas fa-bars',
    'x-circle': 'fas fa-times-circle'
  };

  constructor(private authService: AuthService, private router: Router) {}

  ngOnInit(): void {
    this.loadNavigationFromStorage();
    this.expandActiveMenuOnInit();

    // Listen to route changes to expand active menu
    this.router.events
      .pipe(filter(event => event instanceof NavigationEnd))
      .subscribe(() => {
        this.expandActiveMenu();
      });
  }

  private loadNavigationFromStorage(): void {
    try {
      const navigationData = sessionStorage.getItem('UI-navigations');

      if (navigationData) {
        const parsedData: NavigationResponse = JSON.parse(navigationData);
        this.menuItems = this.transformNavigationData(parsedData.data);
      } else {
        console.warn('No navigation data found in session storage');
      }
    } catch (error) {
      console.error('Error loading navigation from session storage:', error);
    }
  }

  private transformNavigationData(modules: NavigationModule[]): MenuItem[] {
    const transformedItems: MenuItem[] = [];

    const activeModules = modules
      .filter(module => module.module_is_active)
      .sort((a, b) => a.module_order - b.module_order);

    activeModules.forEach(module => {
      const activeMenus = module.menus
        .filter(menu => menu.menu_is_active)
        .sort((a, b) => a.menu_order - b.menu_order);

      activeMenus.forEach(menu => {
        const menuItem: MenuItem = {
          id: menu.menu_code,
          name: menu.menu_name,
          icon: this.mapIcon(menu.menu_icon),
          expanded: false
        };

        const accessibleSubmenus = menu.submenus.filter(
          submenu => submenu.submenu_is_active && submenu.has_access
        ).sort((a, b) => a.submenu_order - b.submenu_order);

        if (accessibleSubmenus.length > 0) {
          menuItem.children = accessibleSubmenus.map(submenu => ({
            id: submenu.submenu_code,
            name: submenu.submenu_name,
            path: submenu.submenu_route,
            icon: this.mapIcon(submenu.submenu_icon)
          }));
        } else {
          menuItem.path = menu.menu_route;
        }

        transformedItems.push(menuItem);
      });
    });

    return transformedItems;
  }

  private mapIcon(apiIcon: string): string {
    return this.iconMap[apiIcon] || `fas fa-${apiIcon}`;
  }

  private expandActiveMenuOnInit(): void {
    this.expandActiveMenu();
  }

private expandActiveMenu(): void {
  const currentUrl = this.router.url;

  // First, collapse all menus to ensure only one can be expanded
  this.menuItems.forEach(item => {
    if (item.children) {
      item.expanded = false;
    }
  });

  // Then, expand only the menu that contains the active child route
  this.menuItems.forEach(item => {
    if (item.children) {
      const hasActiveChild = item.children.some(child =>
        currentUrl.startsWith(child.path || '')
      );

      if (hasActiveChild) {
        item.expanded = true;
      }
    }
  });
}

  toggleMenu(itemId: string): void {
    const item = this.findMenuItem(itemId);
    if (item && item.children) {
      const wasExpanded = item.expanded || false;
      item.expanded = !wasExpanded;
      if (item.expanded) {
        this.closeOtherMenus(itemId);
        // Hide any floating submenu when expanding accordion
        this.hoveredMenuId = null;
        this.currentHoveredElement = null;
      }
    }
  }

  onMenuHover(itemId: string, event: MouseEvent): void {
    if (this.hoverTimeout) {
      clearTimeout(this.hoverTimeout);
    }
    if (this.submenuTimeout) {
      clearTimeout(this.submenuTimeout);
    }

    const item = this.findMenuItem(itemId);
    if (!item || item.expanded) {
      return; // Do not show floating submenu if already expanded
    }

    this.currentHoveredElement = event.currentTarget as HTMLElement;
    const rect = this.currentHoveredElement.getBoundingClientRect();

    this.submenuLeft = rect.right + 8;
    this.submenuTop = rect.top + 1; // Small vertical offset for alignment

    this.hoveredMenuId = itemId;
  }

  onMenuLeave(itemId: string): void {
    if (this.hoverTimeout) {
      clearTimeout(this.hoverTimeout);
    }
    this.submenuTimeout = setTimeout(() => {
      if (this.hoveredMenuId === itemId) {
        this.hoveredMenuId = null;
        this.currentHoveredElement = null;
      }
    }, 200);
  }

  keepSubmenuOpen(itemId: string): void {
    if (this.submenuTimeout) {
      clearTimeout(this.submenuTimeout);
    }
    this.hoveredMenuId = itemId;
  }

  closeSubmenu(): void {
    this.submenuTimeout = setTimeout(() => {
      this.hoveredMenuId = null;
      this.currentHoveredElement = null;
    }, 150);
  }

  onSidebarScroll(): void {
    if (this.hoveredMenuId !== null && this.currentHoveredElement) {
      const rect = this.currentHoveredElement.getBoundingClientRect();
      this.submenuLeft = rect.right + 8;
      this.submenuTop = rect.top + 8;
    }
  }

  isMenuActive(item: MenuItem): boolean {
    if (!item.children) return false;
    const currentUrl = this.router.url;
    return item.children.some(child =>
      currentUrl.startsWith(child.path || '')
    );
  }

  private findMenuItem(id: string): MenuItem | undefined {
    return this.menuItems.find(item => item.id === id);
  }

  private closeOtherMenus(exceptId: string): void {
    this.menuItems.forEach(item => {
      if (item.id !== exceptId && item.children) {
        item.expanded = false;
      }
    });
  }

  logout(): void {
    this.authService.logout();
    this.router.navigate(["/login"]);
  }
}