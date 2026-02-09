import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-customer-list',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './customer-list.component.html',
  styleUrl: './customer-list.component.scss'
})
export class CustomerListComponent {
  submenuId = 0;
  submenuName = 'Customer List';
  submenuCode = 'customer_list';
}
