import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-store-dashboard',
  imports: [CommonModule, TranslateModule],
  templateUrl: './store-dashboard.component.html',
  styleUrl: './store-dashboard.component.scss'
})
export class StoreDashboardComponent {
  submenuId = 2;
  submenuName = 'Store Dashboard';
  submenuCode = 'store_dashboard';
}
