import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-void-transaction',
  imports: [CommonModule, TranslateModule],
  templateUrl: './void-transaction.component.html',
  styleUrl: './void-transaction.component.scss'
})
export class VoidTransactionComponent {
  submenuId = 23;
  submenuName = 'Void Transaction';
  submenuCode = 'void_transaction';
}
