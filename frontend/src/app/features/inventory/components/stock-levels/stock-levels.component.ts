import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-stock-levels',
  imports: [CommonModule, TranslateModule],
  templateUrl: './stock-levels.component.html',
  styleUrl: './stock-levels.component.scss'
})
export class StockLevelsComponent {
  submenuId = 34;
  submenuName = 'Stock Levels';
  submenuCode = 'stock_levels';
}
