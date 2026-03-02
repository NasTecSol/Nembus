import { NgModule } from "@angular/core";
import { CommonModule } from "@angular/common";
import { USER_ROLES_ROUTES } from "./user-roles.routes";
import { RouterModule } from "@angular/router";

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    RouterModule.forChild(USER_ROLES_ROUTES),
  ],
})
export class UserRolesModule {}
