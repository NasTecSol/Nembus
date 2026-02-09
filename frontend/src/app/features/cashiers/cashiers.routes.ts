import { Routes } from "@angular/router";

export const CASHIERS_ROUTES: Routes = [
  {
    path: "",
    redirectTo: "list",
    pathMatch: "full",
  },
  {
    path: "list",
    loadComponent: () =>
      import("./components/cashier-list/cashier-list.component").then(
        (m) => m.CashierListComponent
      ),
  },
  {
    path: "new",
    loadComponent: () =>
      import("./components/add-cashier/add-cashier.component").then(
        (m) => m.AddCashierComponent
      ),
  },
  {
    path: "sessions",
    children: [
      {
        path: "active",
        loadComponent: () =>
          import("./components/active-sessions/active-sessions.component").then(
            (m) => m.ActiveSessionsComponent
          ),
      },
      {
        path: "history",
        loadComponent: () =>
          import("./components/session-history/session-history.component").then(
            (m) => m.SessionHistoryComponent
          ),
      },
      {
        path: "open",
        loadComponent: () =>
          import("./components/open-session/open-session.component").then(
            (m) => m.OpenSessionComponent
          ),
      },
      {
        path: "close",
        loadComponent: () =>
          import("./components/close-session/close-session.component").then(
            (m) => m.CloseSessionComponent
          ),
      },
    ],
  },
];
