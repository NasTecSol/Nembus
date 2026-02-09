import { NgModule } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterModule } from "@angular/router";
import { SALES_ORDERS_ROUTES } from "./sales-orders.routes";

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    RouterModule.forChild(SALES_ORDERS_ROUTES),
  ],
})
export class SalesOrdersModule {}
