import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-product-import',
  imports: [CommonModule, TranslateModule],
  templateUrl: './product-import.component.html',
  styleUrl: './product-import.component.scss'
})
export class ProductImportComponent {
  submenuId = 42;
  submenuName = 'Product Import';
  submenuCode = 'product_import';
}
