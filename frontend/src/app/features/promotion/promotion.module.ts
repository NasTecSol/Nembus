import { NgModule } from "@angular/core";
import { RouterModule } from "@angular/router";
import { PROMOTION_ROUTES } from "./promotion.routes";

@NgModule({
  declarations: [],
  imports: [RouterModule.forChild(PROMOTION_ROUTES)],
})
export class PromotionModule {}
