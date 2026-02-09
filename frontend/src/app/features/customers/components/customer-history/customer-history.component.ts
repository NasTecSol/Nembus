import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-customer-history',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './customer-history.component.html',
  styleUrl: './customer-history.component.scss'
})
export class CustomerHistoryComponent {
  submenuId = 0;
  submenuName = 'Customer History';
  submenuCode = 'customer_history';
}
