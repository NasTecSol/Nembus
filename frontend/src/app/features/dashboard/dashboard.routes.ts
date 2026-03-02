import { Routes } from "@angular/router";
import { HomeComponent } from "./components/home/home.component";
import { AdminDashboardComponent } from "./components/admin-dashboard/admin-dashboard.component";
import { StoreDashboardComponent } from "./components/store-dashboard/store-dashboard.component";
import { SalesAnalyticsComponent } from "./components/sales-analytics/sales-analytics.component";
import { InventoryAnalyticsComponent } from "./components/inventory-analytics/inventory-analytics.component";

export const DASHBOARD_ROUTES: Routes = [


  {
    path: "",
    redirectTo: "admin",
    pathMatch: "full",
  },
  {
    path: "admin",
    component: AdminDashboardComponent,
  },

  {
    path: "stores",
    component: StoreDashboardComponent,
  },
  {
    path: "analytics",
    children: [
      {
        path: "sales",
        component: SalesAnalyticsComponent,
      },
      {
        path: "inventory",
        component: InventoryAnalyticsComponent,
      },
    ],
  },
];
