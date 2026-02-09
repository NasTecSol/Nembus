import { Routes } from "@angular/router";

export const PRODUCTS_ROUTES: Routes = [
  {
    path: "",
    redirectTo: "list",
    pathMatch: "full",
  },
  {
    path: "list",
    loadComponent: () =>
      import("../inventory/components/itemslist/itemslist.component").then(
        (m) => m.ItemslistComponent
      ),
  },
  {
    path: "new",
    loadComponent: () =>
      import("../inventory/components/add-product/add-product.component").then(
        (m) => m.AddProductComponent
      ),
  },
  {
    path: "import",
    loadComponent: () =>
      import("./components/product-import/product-import.component").then(
        (m) => m.ProductImportComponent
      ),
  },
  {
    path: "categories",
    children: [
      {
        path: "",
        redirectTo: "list",
        pathMatch: "full",
      },
      {
        path: "list",
        loadComponent: () =>
          import("./components/category-list/category-list.component").then(
            (m) => m.CategoryListComponent
          ),
      },
      {
        path: "new",
        loadComponent: () =>
          import("./components/add-category/add-category.component").then(
            (m) => m.AddCategoryComponent
          ),
      },
    ],
  },
  {
    path: "brands",
    children: [
      {
        path: "",
        redirectTo: "list",
        pathMatch: "full",
      },
      {
        path: "list",
        loadComponent: () =>
          import("./components/brand-list/brand-list.component").then(
            (m) => m.BrandListComponent
          ),
      },
      {
        path: "new",
        loadComponent: () =>
          import("./components/add-brand/add-brand.component").then(
            (m) => m.AddBrandComponent
          ),
      },
    ],
  },
  {
    path: "price-lists",
    children: [
      {
        path: "",
        redirectTo: "list",
        pathMatch: "full",
      },
      {
        path: "list",
        loadComponent: () =>
          import("./components/price-list-mgmt/price-list-mgmt.component").then(
            (m) => m.PriceListMgmtComponent
          ),
      },
      {
        path: "new",
        loadComponent: () =>
          import("./components/add-price-list/add-price-list.component").then(
            (m) => m.AddPriceListComponent
          ),
      },
    ],
  },
];
