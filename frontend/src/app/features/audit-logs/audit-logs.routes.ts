import { Routes } from "@angular/router";

export const AUDIT_LOGS_ROUTES: Routes = [
  {
    path: "",
    redirectTo: "view",
    pathMatch: "full",
  },
  {
    path: "view",
    loadComponent: () =>
      import("./components/view-audit-logs/view-audit-logs.component").then(
        (m) => m.ViewAuditLogsComponent
      ),
  },
];
