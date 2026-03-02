import { CommonModule } from "@angular/common";
import { NgModule } from "@angular/core";
import { RouterModule } from "@angular/router";
import { REPORTS_ROUTES } from "./reports.routes";

@NgModule({
  declarations: [],
  imports: [CommonModule, RouterModule.forChild(REPORTS_ROUTES)],
})
export class ReportsModule {}
