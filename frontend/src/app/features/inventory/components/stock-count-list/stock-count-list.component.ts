import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-stock-count-list',
  imports: [CommonModule, TranslateModule],
  templateUrl: './stock-count-list.component.html',
  styleUrl: './stock-count-list.component.scss'
})
export class StockCountListComponent {
  submenuId = 38;
  submenuName = 'Stock Count List';
  submenuCode = 'stock_count_list';
}
