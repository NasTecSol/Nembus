import { NgModule } from "@angular/core";
import { CommonModule } from "@angular/common";
import { CASH_COLLECTION_ROUTES } from "./cash-collection.routes";
import { RouterModule } from "@angular/router";

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    RouterModule.forChild(CASH_COLLECTION_ROUTES),
  ],
})
export class CashCollectionModule {}
