import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-view-audit-logs',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './view-audit-logs.component.html',
  styleUrl: './view-audit-logs.component.scss'
})
export class ViewAuditLogsComponent {
  submenuId = 0;
  submenuName = 'View Audit Logs';
  submenuCode = 'view_audit_logs';
}
