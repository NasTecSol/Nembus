import { Injectable } from "@angular/core";
import { CanActivate, Router } from "@angular/router";
import { AuthService } from "../../../../core/services/auth.service";

/**
 * This guard redirects already logged-in users from login/register pages
 */
@Injectable({ providedIn: "root" })
export class AuthRedirectGuard implements CanActivate {
    constructor(private authService: AuthService, private router: Router) { }

    canActivate(): boolean {
        const isLoggedIn = this.authService.isLoggedIn();

        if (isLoggedIn) {
            this.router.navigate(["/dashboard"]);
            return false;
        }

        return true;
    }
}
