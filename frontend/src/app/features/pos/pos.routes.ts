import { Routes } from "@angular/router";

export const POS_ROUTES: Routes = [
  {
    path: "",
    redirectTo: "transactions/list",
    pathMatch: "full",
  },
  {
    path: "transactions",
    children: [
      {
        path: "list",
        loadComponent: () =>
          import("./components/transaction-list/transaction-list.component").then(
            (m) => m.TransactionListComponent
          ),
      },
      {
        path: "new",
        loadComponent: () =>
          import("./components/process-sale/process-sale.component").then(
            (m) => m.ProcessSaleComponent
          ),
      },
      {
        path: "void",
        loadComponent: () =>
          import("./components/void-transaction/void-transaction.component").then(
            (m) => m.VoidTransactionComponent
          ),
      },
    ],
  },
  {
    path: "terminals",
    children: [
      {
        path: "",
        redirectTo: "list",
        pathMatch: "full",
      },
      {
        path: "list",
        loadComponent: () =>
          import("./components/terminal-list/terminal-list.component").then(
            (m) => m.TerminalListComponent
          ),
      },
      {
        path: "new",
        loadComponent: () =>
          import("./components/add-terminal/add-terminal.component").then(
            (m) => m.AddTerminalComponent
          ),
      },
    ],
  },
  {
    path: "reports",
    children: [
      {
        path: "daily",
        loadComponent: () =>
          import("./components/daily-sales/daily-sales.component").then(
            (m) => m.DailySalesComponent
          ),
      },
      {
        path: "cashier",
        loadComponent: () =>
          import("./components/cashier-performance/cashier-performance.component").then(
            (m) => m.CashierPerformanceComponent
          ),
      },
    ],
  },
];
