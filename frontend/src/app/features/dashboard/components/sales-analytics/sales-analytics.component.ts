import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-sales-analytics',
  imports: [CommonModule, TranslateModule],
  templateUrl: './sales-analytics.component.html',
  styleUrl: './sales-analytics.component.scss'
})
export class SalesAnalyticsComponent {
  submenuId = 3;
  submenuName = 'Sales Analytics';
  submenuCode = 'sales_analytics';
}
