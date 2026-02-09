import { Routes } from "@angular/router";

export const SALES_ORDERS_ROUTES: Routes = [
  {
    path: "",
    redirectTo: "list",
    pathMatch: "full",
  },
  {
    path: "list",
    loadComponent: () =>
      import("./components/so-list/so-list.component").then(
        (m) => m.SoListComponent
      ),
  },
  {
    path: "new",
    loadComponent: () =>
      import("./components/create-so/create-so.component").then(
        (m) => m.CreateSoComponent
      ),
  },
];
