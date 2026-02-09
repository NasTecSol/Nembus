import { NgModule } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterModule } from "@angular/router";
import { POS_ROUTES } from "./pos.routes";

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    RouterModule.forChild(POS_ROUTES),
  ],
})
export class PosModule {}
