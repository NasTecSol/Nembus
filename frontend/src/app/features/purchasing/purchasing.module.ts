import { CommonModule } from "@angular/common";
import { NgModule } from "@angular/core";
import { RouterModule } from "@angular/router";
import { PURCHASHING_RROUTES } from "./purchasing.routes";

@NgModule({
  declarations: [],
  imports: [CommonModule, RouterModule.forChild(PURCHASHING_RROUTES)],
})
export class PurchasingModule {}
