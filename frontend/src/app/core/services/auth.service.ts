import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  private readonly mockEmail = 'test@gmail.com';
  private readonly mockPassword = '123456';

  login(email: string, password: string): boolean {
    if (email === this.mockEmail && password === this.mockPassword) {
      localStorage.setItem('token', 'mock-jwt-token'); 
      return true;
    }
    return false;
  }

  logout() {
    localStorage.clear();
    sessionStorage.clear();
  }

  isLoggedIn(): boolean {
    return !!localStorage.getItem('token');
  }
}
