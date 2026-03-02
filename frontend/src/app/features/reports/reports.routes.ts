import { Routes } from "@angular/router";
import { StarterComponent } from "./components/starter/starter.component";
import { DailySalesComponent } from "./components/daily-sales/daily-sales.component";
import { MonthlySalesComponent } from "./components/monthly-sales/monthly-sales.component";
import { ProductPerformanceComponent } from "./components/product-performance/product-performance.component";
import { PurchaseSummaryComponent } from "./components/purchase-summary/purchase-summary.component";
import { SupplierAnalysisComponent } from "./components/supplier-analysis/supplier-analysis.component";
import { StockValuationComponent } from "./components/stock-valuation/stock-valuation.component";
import { InventoryTurnoverComponent } from "./components/inventory-turnover/inventory-turnover.component";
import { ProfitLossComponent } from "./components/profit-loss/profit-loss.component";
import { DiscountAnalysisComponent } from "./components/discount-analysis/discount-analysis.component";

export const REPORTS_ROUTES: Routes = [
  {
       path: "",
       redirectTo: "daily",
       pathMatch: "full",
     },
     {
       path: "daily",
       component: DailySalesComponent,
     },
      {
       path: "monthly",
       component: MonthlySalesComponent,
     },
      {
       path: "products",
       component: ProductPerformanceComponent,
     },
      {
       path: "summary",
       component: PurchaseSummaryComponent,
     },
       {
       path: "suppliers",
       component: SupplierAnalysisComponent,
     }, 
       {
       path: "valuation",
       component: StockValuationComponent,
     },
      {
       path: "turnover",
       component: InventoryTurnoverComponent,
     },
     {
       path: "pl",
       component: ProfitLossComponent,
     },
     {
       path: "discounts",
       component: DiscountAnalysisComponent,
     },
       {
    path: "sales",
    children: [
      {
        path: "",
        redirectTo: "daily",
        pathMatch: "full",
      },
      {
        path: "daily",
        loadComponent: () =>
          import("./components/daily-sales/daily-sales.component").then(
            (m) => m.DailySalesComponent
          ),
      },
      {
        path: "monthly",
        loadComponent: () =>
          import("./components/monthly-sales/monthly-sales.component").then(
            (m) => m.MonthlySalesComponent
          ),
      },
        {
        path: "products",
        loadComponent: () =>
          import("./components/product-performance/product-performance.component").then(
            (m) => m.ProductPerformanceComponent
          ),
      },
    
    ],
  },
   {
    path: "purchases",
    children: [
      {
        path: "",
        redirectTo: "summary",
        pathMatch: "full",
      },
      {
        path: "summary",
        loadComponent: () =>
          import("./components/purchase-summary/purchase-summary.component").then(
            (m) => m.PurchaseSummaryComponent
          ),
      },
      {
        path: "suppliers",
        loadComponent: () =>
          import("./components/supplier-analysis/supplier-analysis.component").then(
            (m) => m.SupplierAnalysisComponent
          ),
      },
   
    
    ],
  },
   {
    path: "inventory",
    children: [
      {
        path: "",
        redirectTo: "valuation",
        pathMatch: "full",
      },
      {
        path: "valuation",
        loadComponent: () =>
          import("./components/stock-valuation/stock-valuation.component").then(
            (m) => m.StockValuationComponent
          ),
      },
      {
        path: "turnover",
        loadComponent: () =>
          import("./components/inventory-turnover/inventory-turnover.component").then(
            (m) => m.InventoryTurnoverComponent
          ),
      },
   
    
    ],
  },
    {
    path: "financial",
    children: [
      {
        path: "",
        redirectTo: "pl",
        pathMatch: "full",
      },
      {
        path: "pl",
        loadComponent: () =>
          import("./components/profit-loss/profit-loss.component").then(
            (m) => m.ProfitLossComponent
          ),
      },
      {
        path: "discounts",
        loadComponent: () =>
          import("./components/discount-analysis/discount-analysis.component").then(
            (m) => m.DiscountAnalysisComponent
          ),
      },
   
    
    ],
  },
];
