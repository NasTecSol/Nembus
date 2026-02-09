import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-transaction-list',
  imports: [CommonModule, TranslateModule],
  templateUrl: './transaction-list.component.html',
  styleUrl: './transaction-list.component.scss'
})
export class TransactionListComponent {
  submenuId = 21;
  submenuName = 'Transaction List';
  submenuCode = 'transaction_list';
}
