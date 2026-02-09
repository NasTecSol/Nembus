import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-inventory-analytics',
  imports: [CommonModule, TranslateModule],
  templateUrl: './inventory-analytics.component.html',
  styleUrl: './inventory-analytics.component.scss'
})
export class InventoryAnalyticsComponent {
  submenuId = 4;
  submenuName = 'Inventory Analytics';
  submenuCode = 'inventory_analytics';
}
