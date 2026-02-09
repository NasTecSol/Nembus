import { NgModule } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterModule } from "@angular/router";
import { STORES_ROUTES } from "./stores.routes";

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    RouterModule.forChild(STORES_ROUTES),
  ],
})
export class StoresModule {}
