import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-add-cashier',
  imports: [CommonModule, TranslateModule],
  templateUrl: './add-cashier.component.html',
  styleUrl: './add-cashier.component.scss'
})
export class AddCashierComponent {
  submenuId = 29;
  submenuName = 'Add Cashier';
  submenuCode = 'add_cashier';
}
