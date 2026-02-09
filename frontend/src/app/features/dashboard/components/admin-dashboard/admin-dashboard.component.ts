import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-admin-dashboard',
  imports: [CommonModule, TranslateModule],
  templateUrl: './admin-dashboard.component.html',
  styleUrl: './admin-dashboard.component.scss'
})
export class AdminDashboardComponent {
  submenuId = 1;
  submenuName = 'Admin Dashboard';
  submenuCode = 'admin_dashboard';
}
