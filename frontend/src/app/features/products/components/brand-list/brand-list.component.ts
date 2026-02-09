import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-brand-list',
  imports: [CommonModule, TranslateModule],
  templateUrl: './brand-list.component.html',
  styleUrl: './brand-list.component.scss'
})
export class BrandListComponent {
  submenuId = 45;
  submenuName = 'Brand List';
  submenuCode = 'brand_list';
}
