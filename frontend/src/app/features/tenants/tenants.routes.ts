import { Routes } from "@angular/router";
import { TenantListComponent } from "./components/tenant-list/tenant-list.component";
import { AddTenantComponent } from "./components/add-tenant/add-tenant.component";
import { TenantConfigComponent } from "./components/tenant-config/tenant-config.component";

export const TENANTS_ROUTES: Routes = [
  {
    path: "",
    redirectTo: "list",
    pathMatch: "full",
  },
  {
    path: "list",
    component: TenantListComponent,
  },
  {
    path: "new",
    component: AddTenantComponent,
  },
    {
    path: "config",
    component: TenantConfigComponent,
  },
  {
    path: "config/:id",
    component: TenantConfigComponent,
  },
];
