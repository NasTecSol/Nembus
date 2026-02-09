import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-low-stock',
  imports: [CommonModule, TranslateModule],
  templateUrl: './low-stock.component.html',
  styleUrl: './low-stock.component.scss'
})
export class LowStockComponent {
  submenuId = 35;
  submenuName = 'Low Stock Alert';
  submenuCode = 'low_stock';
}
