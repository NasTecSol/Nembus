import { Routes } from "@angular/router";
import { AuthGuard } from "../auth/components/guards/auth.guard";
import { DashboardStoreComponent } from "../dashboard-store/dashboard-store.component";
import { AddOrgComponent } from "../organizations/components/add-org/add-org.component";

export const DASHBOARD_LAYOUT_ROUTES: Routes = [
  {
    path: "",
    loadComponent: () =>
      import("./main-layout/main-layout.component").then(
        (m) => m.MainLayoutComponent
      ),
    canActivateChild: [AuthGuard],
    children: [
      {
        path: "dashboard",
        loadChildren: () =>
          import("../dashboard/dashboard.module").then(
            (m) => m.DashboardModule
          ),
      },
      {
        path: "users",
        loadChildren: () =>
          import("../user-roles/user-roles.module").then(
            (m) => m.UserRolesModule
          ),
      },
      {
        path: "inventory",
        loadChildren: () =>
          import("../inventory/inventory.module").then(
            (m) => m.InventoryModule
          ),
      },
      {
        path: "wholesale",
        loadChildren: () =>
          import("../wholesale/wholesale.module").then(
            (m) => m.WholesaleModule
          ),
      },
      {
        path: "stores",
        loadChildren: () =>
          import("../stores/stores.module").then((m) => m.StoresModule),
      },
      {
        path: "cash-collection",
        loadChildren: () =>
          import("../cash-collection/cash-collection.module").then(
            (m) => m.CashCollectionModule
          ),
      },
      {
        path: "purchasing",
        loadChildren: () =>
          import("../purchasing/purchasing.module").then(
            (m) => m.PurchasingModule
          ),
      },
      {
        path: "promotions",
        loadChildren: () =>
          import("../promotion/promotion.module").then(
            (m) => m.PromotionModule
          ),
      },
      {
        path: "reports",
        loadChildren: () =>
          import("../reports/reports.module").then((m) => m.ReportsModule),
      },
      {
        path: "suppliers",
        loadChildren: () =>
          import("../suppliers/suppliers.module").then((m) => m.SuppliersModule),
      },
      {
        path: "Ui-Modules",
        loadChildren: () =>
          import("../ui-modules/ui-modules.module").then((m) => m.UiModulesModule),
      },
      {
        path: "settings",
        loadChildren: () =>
          import("../settings/settings.module").then((m) => m.SettingsModule)
      },
      {
        path: "dashboard-stores",
        component: DashboardStoreComponent,
      },
      {
        path: "admin",
        children: [
          {
            path: "tenants",
            loadChildren: () =>
              import("../tenants/tenants.module").then((m) => m.TenantsModule),
          },
          {
            path: "organizations",
            loadChildren: () =>
              import("../organizations/organizations.module").then(
                (m) => m.OrganizationsModule
              ),
          },
          {
            path: 'organizations/:id/edit',
            component: AddOrgComponent
          },
          {
            path: "users",
            loadChildren: () =>
              import("../user-roles/user-roles.module").then(
                (m) => m.UserRolesModule
              ),
          },
          {
            path: "roles",
            loadChildren: () =>
              import("../user-roles/user-roles.module").then(
                (m) => m.UserRolesModule
              ),
          },
          {
            path: "ui-modules",
            loadChildren: () =>
              import("../ui-modules/ui-modules.module").then(
                (m) => m.UiModulesModule
              ),
          },
          {
            path: "settings",
            loadChildren: () =>
              import("../settings/settings.module").then(
                (m) => m.SettingsModule
              ),
          },
          {
            path: "audit-logs",
            loadChildren: () =>
              import("../audit-logs/audit-logs.module").then(
                (m) => m.AuditLogsModule
              ),
          },
        ],
      },
      {
        path: "pos",
        loadChildren: () =>
          import("../pos/pos.module").then((m) => m.PosModule),
      },
      {
        path: "cashiers",
        loadChildren: () =>
          import("../cashiers/cashiers.module").then((m) => m.CashiersModule),
      },
      {
        path: "products",
        loadChildren: () =>
          import("../products/products.module").then((m) => m.ProductsModule),
      },
      {
        path: "customers",
        loadChildren: () =>
          import("../customers/customers.module").then(
            (m) => m.CustomersModule
          ),
      },
      {
        path: "purchase-orders",
        loadChildren: () =>
          import("../purchase-orders/purchase-orders.module").then(
            (m) => m.PurchaseOrdersModule
          ),
      },
      {
        path: "sales-orders",
        loadChildren: () =>
          import("../sales-orders/sales-orders.module").then(
            (m) => m.SalesOrdersModule
          ),
      },
      {
        path: "admin",
        loadChildren: () =>
          import("../admin/admin.module").then(
            (m) => m.AdminModule
          ),
      },
    ],
  },
];
