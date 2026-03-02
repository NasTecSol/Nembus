import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-price-list-mgmt',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './price-list-mgmt.component.html',
  styleUrl: './price-list-mgmt.component.scss'
})
export class PriceListMgmtComponent {
  submenuId = 0;
  submenuName = 'Price List Management';
  submenuCode = 'price_list_mgmt';
}
