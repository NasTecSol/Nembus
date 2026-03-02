import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-add-customer',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './add-customer.component.html',
  styleUrl: './add-customer.component.scss'
})
export class AddCustomerComponent {
  submenuId = 0;
  submenuName = 'Add Customer';
  submenuCode = 'add_customer';
}
