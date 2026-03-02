import { Routes } from "@angular/router";
import { ListComponent } from "./components/list/list.component";
import { DetailComponent } from "./components/detail/detail.component";


export const CASH_COLLECTION_ROUTES: Routes = [
  {
    path: "",
    component: ListComponent,
  },
  {
    path: "cash-detail/:id",
    component: DetailComponent,
  },
];
