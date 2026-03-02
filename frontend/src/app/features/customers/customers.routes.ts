import { Routes } from "@angular/router";

export const CUSTOMERS_ROUTES: Routes = [
  {
    path: "",
    redirectTo: "list",
    pathMatch: "full",
  },
  {
    path: "list",
    loadComponent: () =>
      import("./components/customer-list/customer-list.component").then(
        (m) => m.CustomerListComponent
      ),
  },
  {
    path: "new",
    loadComponent: () =>
      import("./components/add-customer/add-customer.component").then(
        (m) => m.AddCustomerComponent
      ),
  },
  {
    path: "history",
    loadComponent: () =>
      import("./components/customer-history/customer-history.component").then(
        (m) => m.CustomerHistoryComponent
      ),
  },
];
