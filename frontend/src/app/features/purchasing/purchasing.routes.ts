import { Routes } from "@angular/router";
import { ListComponent } from "./components/list/list.component";
import { CreateComponent } from "./components/create/create.component";
import { SupplierProductsComponent } from "./components/supplier-products/supplier-products.component";

export const PURCHASHING_RROUTES: Routes = [
  {
    path: "",
    component: ListComponent,
  },
  {
    path: "create",
    component: CreateComponent,
  },
  {
    path: "supplier-products",
    component: SupplierProductsComponent,
  },
];
