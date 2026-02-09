import { NgModule } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterModule } from "@angular/router";
import { ORGANIZATIONS_ROUTES } from "./organizations.routes";

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    RouterModule.forChild(ORGANIZATIONS_ROUTES),
  ],
})
export class OrganizationsModule {}
