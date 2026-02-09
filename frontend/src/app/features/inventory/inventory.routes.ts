import { Routes } from "@angular/router";
import { InventoryHomeComponent } from "./components/inventory-home/inventory-home.component";
import { AddProductComponent } from "./components/add-product/add-product.component";
import { ProductDetailComponent } from "./components/product-detail/product-detail.component";

export const INVENTORY_ROUTES: Routes = [
  {
    path: "",
    component: InventoryHomeComponent,
  },
  {
    path: "add-product",
    component: AddProductComponent,
  },
  {
    path: "product-detail/:id",
    component: ProductDetailComponent,
  },
  {
    path: "overview",
    children: [
      {
        path: "levels",
        loadComponent: () =>
          import("./components/stock-levels/stock-levels.component").then(
            (m) => m.StockLevelsComponent
          ),
      },
      {
        path: "low-stock",
        loadComponent: () =>
          import("./components/low-stock/low-stock.component").then(
            (m) => m.LowStockComponent
          ),
      },
    ],
  },
  {
    path: "movements",
    children: [
      {
        path: "history",
        loadComponent: () =>
          import("./components/movement-history/movement-history.component").then(
            (m) => m.MovementHistoryComponent
          ),
      },
      {
        path: "new",
        loadComponent: () =>
          import("./components/record-movement/record-movement.component").then(
            (m) => m.RecordMovementComponent
          ),
      },
    ],
  },
  {
    path: "counts",
    children: [
      {
        path: "",
        redirectTo: "list",
        pathMatch: "full",
      },
      {
        path: "list",
        loadComponent: () =>
          import("./components/stock-count-list/stock-count-list.component").then(
            (m) => m.StockCountListComponent
          ),
      },
      {
        path: "new",
        loadComponent: () =>
          import("./components/create-count/create-count.component").then(
            (m) => m.CreateCountComponent
          ),
      },
    ],
  },
];
