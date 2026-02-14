import { CanActivateChildFn } from '@angular/router';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from '../../../../core/services/auth.service';

/**
 * This guard prevents access to child routes if user is not logged in
 */
export const AuthGuard: CanActivateChildFn = () => {
    const router = inject(Router);
    const authService = inject(AuthService);

    const isLoggedIn = authService.isLoggedIn();

    if (!isLoggedIn) {
        router.navigate(['/login']);
    }

    return isLoggedIn;
};
