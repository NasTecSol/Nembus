import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-daily-sales',
  imports: [CommonModule, TranslateModule],
  templateUrl: './daily-sales.component.html',
  styleUrl: './daily-sales.component.scss'
})
export class DailySalesComponent {
  submenuId = 26;
  submenuName = 'Daily Sales Report';
  submenuCode = 'daily_sales';
}
