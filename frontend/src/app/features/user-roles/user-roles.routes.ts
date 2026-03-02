import { Routes } from "@angular/router";
import { UserListComponent } from "./components/user-list/user-list.component";
import { AddUserComponent } from "./components/add-user/add-user.component";
import { PermissionMatrixComponent } from "./components/permission-matrix/permission-matrix.component";
import { AddRoleComponent } from "./components/add-role/add-role.component";

export const USER_ROLES_ROUTES: Routes = [
  {
    path: "",
    redirectTo: "list",
    pathMatch: "full",
  },
  {
    path: "list",
    component: UserListComponent,
  },
  {
    path: "new",
    component: AddUserComponent,
  },
  // {
  //   path: "create",
  //   component: AddUserComponent,
  // },
    {
    path: "new",
    component: AddRoleComponent,
  },
   {
    path: "permissions",
    component: PermissionMatrixComponent,
  },
  {
    path: "activity",
    loadComponent: () =>
      import("./components/user-activity/user-activity.component").then(
        (m) => m.UserActivityComponent
      ),
  },
  {
    path: "roles",
    children: [
      {
        path: "",
        redirectTo: "list",
        pathMatch: "full",
      },
      {
        path: "list",
        loadComponent: () =>
          import("./components/role-list/role-list.component").then(
            (m) => m.RoleListComponent
          ),
      },
      {
        path: "new",
        loadComponent: () =>
          import("./components/add-role/add-role.component").then(
            (m) => m.AddRoleComponent
          ),
      },
      {
        path: "permissions",
        loadComponent: () =>
          import("./components/permission-matrix/permission-matrix.component").then(
            (m) => m.PermissionMatrixComponent
          ),
      },
    ],
  },
   {
    path: "users",
    children: [
      {
        path: "",
        redirectTo: "new",
        pathMatch: "full",
      },
      {
        path: "new",
        loadComponent: () =>
          import("./components/add-user/add-user.component").then(
            (m) => m.AddUserComponent
          ),
      }
    ],
  },
];
