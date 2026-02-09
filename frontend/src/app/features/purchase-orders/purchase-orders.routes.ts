import { Routes } from "@angular/router";

export const PURCHASE_ORDERS_ROUTES: Routes = [
  {
    path: "",
    redirectTo: "list",
    pathMatch: "full",
  },
  {
    path: "list",
    loadComponent: () =>
      import("./components/po-list/po-list.component").then(
        (m) => m.PoListComponent
      ),
  },
  {
    path: "new",
    loadComponent: () =>
      import("./components/create-po/create-po.component").then(
        (m) => m.CreatePoComponent
      ),
  },
  {
    path: "approve",
    loadComponent: () =>
      import("./components/approve-po/approve-po.component").then(
        (m) => m.ApprovePoComponent
      ),
  },
];
