import { Routes } from "@angular/router";
import { environment } from "../../../environments/environment";
import { SignInComponent } from "./components/sign-in/sign-in";
import { PosSignInComponent } from "./components/sign-in/pos-sign-in";
import { AuthRedirectGuard } from "./guards/auth-redirect.guard";

export const Auth_ROUTES: Routes = [
  {
    path: "setup",
    loadComponent: () =>
      import("./components/setup/setup.component").then((m) => m.SetupComponent),
  },
  {
    path: "login",
    component: environment.pos ? PosSignInComponent : SignInComponent,
    canActivate: [AuthRedirectGuard],
  },
  {
    path: "",
    redirectTo: "login",
    pathMatch: "full",
  },
];
