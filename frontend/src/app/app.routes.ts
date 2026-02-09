import { Routes } from '@angular/router';
import { SetupComponent } from './features/setup/setup.component';
import { LoginComponent } from './features/login/login.component';
import { HomeComponent } from './features/home/home.component';

// export const routes: Routes = [
//     { path: 'setup', component: SetupComponent },
//     { path: 'login', component: LoginComponent },
//     { path: 'home', component: HomeComponent },
//     { path: '', redirectTo: '/login', pathMatch: 'full' }
// ];


export const routes: Routes = [
  {
    path: "",
    loadChildren: () =>
      import("./features/auth/auth-routes").then((m) => m.Auth_ROUTES),
  },
//   {
//     path: "",
//     loadChildren: () =>
//       import("./features/dashboard-layout/dashboard-layout.routes").then(
//         (m) => m.DASHBOARD_LAYOUT_ROUTES
//       ),
//   },
];