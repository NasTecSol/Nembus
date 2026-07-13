import { Component, OnInit } from "@angular/core";
import { Router } from "@angular/router";
import { AuthService } from "../../../../core/services/auth.service";
import { FormsModule } from "@angular/forms";
import { CommonModule } from "@angular/common";
import { TranslateModule } from "@ngx-translate/core";
import { ToastyService } from "../../../../core/services/toasty.service";
import { HttpClient } from '@angular/common/http';
import { jwtDecode } from 'jwt-decode';
import { environment } from "../../../../../environments/environment";

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslateModule],
  templateUrl: './login.component.html',
  // styleUrl: './login.component.scss'
})
export class LoginComponent implements OnInit {
  public email = "";
  public password = "";
  public errorMsg = "";
  apiUrl = environment.baseUrl;
  private loginUrl = `${this.apiUrl}/api/auth/login`;
  constructor(
    private authService: AuthService,
    private router: Router,
    private toasty: ToastyService,
    private http: HttpClient
  ) { }
  ngOnInit(): void {

  }



  onLogin() {
    if (!this.email || !this.password) {
      this.errorMsg = 'Email and password are required';
      return;
    }
    const payload = {
      user_login: this.email,
      password: this.password
    };
    this.http.post<any>(this.loginUrl, payload).subscribe({
      next: (res) => {
        console.log('Login response:', res);
        const token = res?.data;
        if (!token) {
          console.error('Token not found in response', res);
          return;
        }
        const decoded: any = jwtDecode(token);
        console.log('Decoded token:', decoded);
        const userId = decoded?.user_id;
        if (!userId) {
          console.error('user_id not found in token');
          return;
        }
        localStorage.setItem('token', token);
        localStorage.setItem('decoded', JSON.stringify(decoded));
        const navUrl = `${this.apiUrl}/api/navigation/user/${userId}`;
        this.http.get<any>(navUrl).subscribe({
          next: (navRes) => {
            console.log('Navigation response:', navRes);
            sessionStorage.setItem('UI-navigations', JSON.stringify(navRes));
            this.toasty.success('Login is successful');
            const backofficeUrl = environment.pos ? '/backoffice/dashboard' : '/dashboard';
            this.router.navigateByUrl(backofficeUrl);
          },
          error: (navErr) => {
            console.error('Navigation API failed:', navErr);
            this.toasty.error('Failed to load user navigation');
          }
        });
      },
      error: (err) => {
        console.error('Login failed:', err);
        this.errorMsg = 'Invalid email or password';
        this.toasty.error('Invalid Credentials');
      }
    });
  }
}
