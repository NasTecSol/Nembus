import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-cashier-list',
  imports: [CommonModule, TranslateModule],
  templateUrl: './cashier-list.component.html',
  styleUrl: './cashier-list.component.scss'
})
export class CashierListComponent {
  submenuId = 28;
  submenuName = 'Cashier List';
  submenuCode = 'cashier_list';
}
