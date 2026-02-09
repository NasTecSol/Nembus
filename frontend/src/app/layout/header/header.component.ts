import { CommonModule } from "@angular/common";
import { Component,ElementRef, HostListener, ViewChild } from "@angular/core";
import { TranslateService, TranslateModule } from "@ngx-translate/core";
import { AuthService } from "../../core/services/auth.service";
import { Router } from "@angular/router";

@Component({
  selector: "app-header",
  standalone: true,
  imports: [TranslateModule, CommonModule],
  templateUrl: "./header.component.html",
})
export class HeaderComponent {
  @ViewChild('searchInput') searchInput!: ElementRef;
  public currentLang: string = "en";
  public languageDropdownOpen: boolean = false;
  public userProfileOpen: boolean = false;
    // Search state
  searchExpanded = false;
  storeDropdownOpen = false;
  selectedStore = 'Downtown Store';
  stores = [
    'Downtown Store',
    'Uptown Store',
    'West Side Store',
    'East Side Store',
    'Mall Store'
  ];
  constructor(
    private translate: TranslateService,
    private router: Router,
    private authService: AuthService,
     private eRef: ElementRef 
  ) {
    const savedLang = localStorage.getItem('preferredLanguage') || 'en';
    this.currentLang = savedLang;
    this.translate.setDefaultLang(savedLang);
    this.translate.use(savedLang);

    // Optional: Set direction for RTL languages
    document.documentElement.dir = savedLang === 'ar' ? 'rtl' : 'ltr';
  }

  toggleLanguageDropdown() {
    this.languageDropdownOpen = !this.languageDropdownOpen;
  }

  // Search methods
  toggleSearch() {
    this.searchExpanded = true;
    // Focus input after expansion
    setTimeout(() => {
      this.searchInput?.nativeElement.focus();
    }, 100);
  }

  closeSearch() {
    this.searchExpanded = false;
  }

  onSearchBlur() {
    // Delay to allow click events to fire
    setTimeout(() => {
      const searchValue = this.searchInput?.nativeElement.value;
      if (!searchValue || searchValue.trim() === '') {
        this.searchExpanded = false;
      }
    }, 200);
  }
  // Store dropdown
  toggleStoreDropdown() {
    this.storeDropdownOpen = !this.storeDropdownOpen;
    // Close other dropdowns
    this.languageDropdownOpen = false;
    this.userProfileOpen = false;
  }

  selectStore(store: string) {
    this.selectedStore = store;
    this.storeDropdownOpen = false;
    // Add your store selection logic here
    console.log('Selected store:', store);
  }

  // User profile dropdown
  toggleUserProfile() {
    this.userProfileOpen = !this.userProfileOpen;
    // Close other dropdowns
    this.languageDropdownOpen = false;
    this.storeDropdownOpen = false;
  }

  navigateToProfile() {
    this.userProfileOpen = false;
    this.router.navigate(['/profile']);
  }

  navigateToSettings() {
    this.userProfileOpen = false;
    this.router.navigate(['/settings']);
  }
  logout() {
    this.authService.logout();
    this.router.navigate(['/login']);
  }

  changeLang(lang: string) {
    this.currentLang = lang;
    this.translate.use(lang);
    localStorage.setItem('preferredLanguage', lang);
    this.languageDropdownOpen = false;

    // Optional: Set direction for RTL languages
    document.documentElement.dir = lang === 'ar' ? 'rtl' : 'ltr';
  }

  // 👇 Close dropdowns if click is outside the component
  @HostListener('document:click', ['$event'])
  onClickOutside(event: MouseEvent) {
    if (!this.eRef.nativeElement.contains(event.target)) {
      this.languageDropdownOpen = false;
      this.userProfileOpen = false;
    }
  }
}
