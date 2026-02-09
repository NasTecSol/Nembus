import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-cashier-performance',
  imports: [CommonModule, TranslateModule],
  templateUrl: './cashier-performance.component.html',
  styleUrl: './cashier-performance.component.scss'
})
export class CashierPerformanceComponent {
  submenuId = 27;
  submenuName = 'Cashier Performance';
  submenuCode = 'cashier_performance';
}
