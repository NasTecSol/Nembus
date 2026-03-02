import { NgModule } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterModule } from "@angular/router";
import { CASHIERS_ROUTES } from "./cashiers.routes";

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    RouterModule.forChild(CASHIERS_ROUTES),
  ],
})
export class CashiersModule {}
