import { Routes } from "@angular/router";
import { SupplierListComponent } from "./components/supplier-list/supplier-list.component";
import { AddSupplierComponent } from "./components/add-supplier/add-supplier.component";

export const SUPPLIERS_ROUTES: Routes = [
    {
      path: "",
      redirectTo: "list",
      pathMatch: "full",
    },
    {
      path: "list",
      component: SupplierListComponent,
    },
    {
      path: "new",
      component: AddSupplierComponent,
    },
    {
    path: "suppliers",
    children: [
      {
        path: "",
        redirectTo: "list",
        pathMatch: "full",
      },
      {
        path: "list",
        loadComponent: () =>
          import("./components/supplier-list/supplier-list.component").then(
            (m) => m.SupplierListComponent
          ),
      },
      {
        path: "new",
        loadComponent: () =>
          import("./components/add-supplier/add-supplier.component").then(
            (m) => m.AddSupplierComponent
          ),
      },
    
    ],
  },
];
