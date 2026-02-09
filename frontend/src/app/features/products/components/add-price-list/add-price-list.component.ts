import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-add-price-list',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './add-price-list.component.html',
  styleUrl: './add-price-list.component.scss'
})
export class AddPriceListComponent {
  submenuId = 0;
  submenuName = 'Add Price List';
  submenuCode = 'add_price_list';
}
