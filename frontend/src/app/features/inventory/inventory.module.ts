import { NgModule } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterModule } from "@angular/router";
import { INVENTORY_ROUTES } from "./inventory.routes";

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    RouterModule.forChild(INVENTORY_ROUTES), // ✅ IMPORTANT
  ],
})
export class InventoryModule {}
