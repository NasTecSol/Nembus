import { Routes } from "@angular/router";
import { environment } from "../environments/environment";

const SETUP_AUTH_ROUTES: Routes = [
  {
    path: "auth",
    loadChildren: () =>
      import("./features/auth/auth-routes").then((m) => m.Auth_ROUTES),
  },
];

export const routes: Routes = environment.pos
  ? [
      ...SETUP_AUTH_ROUTES,
      {
        path: "dashboard",
        loadChildren: () =>
          import(
            "./features/pos-dashboard-layout/pos-dashboard-layout.routes"
          ).then((m) => m.POS_DASHBOARD_LAYOUT_ROUTES),
      },
      {
        path: "backoffice",
        loadChildren: () =>
          import("./features/dashboard-layout/dashboard-layout.routes").then(
            (m) => m.DASHBOARD_LAYOUT_ROUTES
          ),
      },
      {
        path: "",
        pathMatch: "full",
        redirectTo: "auth/login",
      },
      {
        path: "**",
        redirectTo: "auth/login",
      },
    ]
  : [
      ...SETUP_AUTH_ROUTES,
      {
        path: "",
        loadChildren: () =>
          import("./features/dashboard-layout/dashboard-layout.routes").then(
            (m) => m.DASHBOARD_LAYOUT_ROUTES
          ),
      },
      {
        path: "**",
        redirectTo: "auth/login",
      },
    ];
