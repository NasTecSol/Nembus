import { NgModule } from "@angular/core";
import { CommonModule } from "@angular/common";
import { WHOLESALE_ROUTES } from "./wholesale.routes";
import { RouterModule } from "@angular/router";

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    RouterModule.forChild(WHOLESALE_ROUTES),
  ],
})
export class WholesaleModule {}
