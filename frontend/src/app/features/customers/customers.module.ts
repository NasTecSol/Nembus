import { NgModule } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterModule } from "@angular/router";
import { CUSTOMERS_ROUTES } from "./customers.routes";

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    RouterModule.forChild(CUSTOMERS_ROUTES),
  ],
})
export class CustomersModule {}
