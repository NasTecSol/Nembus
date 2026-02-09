import { NgModule } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterModule } from "@angular/router";
import { PURCHASE_ORDERS_ROUTES } from "./purchase-orders.routes";

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    RouterModule.forChild(PURCHASE_ORDERS_ROUTES),
  ],
})
export class PurchaseOrdersModule {}
