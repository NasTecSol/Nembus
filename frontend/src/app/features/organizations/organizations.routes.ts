import { Routes } from "@angular/router";
import { OrgListComponent } from "./components/org-list/org-list.component";
import { AddOrgComponent } from "./components/add-org/add-org.component";

export const ORGANIZATIONS_ROUTES: Routes = [
  {
    path: "",
    redirectTo: "list",
    pathMatch: "full",
  },
  {
    path: "list",
    component: OrgListComponent,
  },
  {
    path: "new",
    component: AddOrgComponent,
  },
];
