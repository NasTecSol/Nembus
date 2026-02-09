import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';

@Component({
    selector: 'app-login',
    standalone: true,
    imports: [CommonModule, FormsModule],
    templateUrl: './login.component.html',
})
export class LoginComponent {
    username = '';
    password = '';

    constructor(private router: Router) { }

    login() {
        console.log('Login attempt:', this.username);
        // For now, any login redirects to home
        if (this.username && this.password) {
            this.router.navigate(['/home']);
        } else {
            alert('Please enter username and password');
        }
    }
}
