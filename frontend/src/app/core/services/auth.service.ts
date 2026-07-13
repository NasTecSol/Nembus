import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  private readonly mockEmail = 'test@gmail.com';
  private readonly mockPassword = '123456';

  setLoggedIn(status: boolean): void {
    if (!status) this.logout();
  }

  /**
   * Backwards compatible overloads:
   * - BOFC legacy mock login: login(email, password) -> boolean
   * - Real login used by POS/BOFC API flows: login(token, decodedToken) -> void
   */
  login(token: string, decodedToken: any): void;
  login(email: string, password: string): boolean;
  login(a: string, b: any): void | boolean {
    // legacy mock login path
    if (typeof b === 'string') {
      const email = a;
      const password = b;
      if (email === this.mockEmail && password === this.mockPassword) {
        localStorage.setItem('token', 'mock-jwt-token');
        localStorage.setItem('decoded', JSON.stringify({ user_id: 'mock' }));
        return true;
      }
      return false;
    }

    // token-based login path
    const token = a;
    const decodedToken = b;
    localStorage.setItem('token', token);
    localStorage.setItem('decoded', JSON.stringify(decodedToken));
  }

  /**
   * Logout and clear session data (POS + BOFC safe)
   */
  logout(): void {
    // Auth
    localStorage.removeItem('token');
    localStorage.removeItem('decoded');

    // Cashier session
    localStorage.removeItem('cashier_mode');
    localStorage.removeItem('cashier_id');

    // Store/terminal
    localStorage.removeItem('store_id');
    localStorage.removeItem('store_data');
    localStorage.removeItem('terminal_id');

    // Navigation
    sessionStorage.removeItem('UI-navigations');

    // NGXS persisted state (cart, cart_visible) — must be cleared so the
    // next user doesn't see the previous user's customer / items / cart ID
    sessionStorage.removeItem('cart');
    sessionStorage.removeItem('cart_visible');

    // Redirect URL
    localStorage.removeItem('redirect_url');
  }

  isLoggedIn(): boolean {
    return !!localStorage.getItem('token');
  }

  getToken(): string | null {
    return localStorage.getItem('token');
  }

  getDecodedToken(): any {
    const decoded = localStorage.getItem('decoded');
    return decoded ? JSON.parse(decoded) : null;
  }

  getUserPermissions(): any {
    const permissions = sessionStorage.getItem('UI-navigations');
    return permissions ? JSON.parse(permissions) : null;
  }

  isCashierMode(): boolean {
    return localStorage.getItem('cashier_mode') === 'true';
  }

  getCashierId(): string | null {
    return localStorage.getItem('cashier_id');
  }

  getStoreId(): string | null {
    return localStorage.getItem('store_id');
  }

  getTerminalId(): string | null {
    return localStorage.getItem('terminal_id');
  }

  getCurrentUser(): any {
    return this.getDecodedToken();
  }

  isAuthenticated(): boolean {
    const token = localStorage.getItem('token');
    const decoded = localStorage.getItem('decoded');
    return !!(token && decoded);
  }
}