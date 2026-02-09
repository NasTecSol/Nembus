import { NgModule } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterModule } from "@angular/router";
import { TENANTS_ROUTES } from "./tenants.routes";

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    RouterModule.forChild(TENANTS_ROUTES),
  ],
})
export class TenantsModule {}
