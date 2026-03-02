import { Routes } from "@angular/router";
import { StoreslistComponent } from "./components/storeslist/storeslist.component";
import { AddStoreComponent } from "./components/add-store/add-store.component";

export const STORES_ROUTES: Routes = [
  {
    path: "",
    redirectTo: "list",
    pathMatch: "full",
  },
  {
    path: "list",
    component: StoreslistComponent,
  },
  {
    path: "new",
    component: AddStoreComponent,
  },
  {
    path: "add-store",
    component: AddStoreComponent,
  },
  {
    path: "config",
    loadComponent: () =>
      import("./components/store-config/store-config.component").then(
        (m) => m.StoreConfigComponent
      ),
  },
  {
    path: "locations",
    children: [
      {
        path: "",
        redirectTo: "list",
        pathMatch: "full",
      },
      {
        path: "list",
        loadComponent: () =>
          import("./components/location-list/location-list.component").then(
            (m) => m.LocationListComponent
          ),
      },
      {
        path: "new",
        loadComponent: () =>
          import("./components/add-location/add-location.component").then(
            (m) => m.AddLocationComponent
          ),
      },
    ],
  },
];
