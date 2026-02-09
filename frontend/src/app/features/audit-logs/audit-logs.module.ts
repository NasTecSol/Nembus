import { NgModule } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterModule } from "@angular/router";
import { AUDIT_LOGS_ROUTES } from "./audit-logs.routes";

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    RouterModule.forChild(AUDIT_LOGS_ROUTES),
  ],
})
export class AuditLogsModule {}
