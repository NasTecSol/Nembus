import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-add-brand',
  imports: [CommonModule, TranslateModule],
  templateUrl: './add-brand.component.html',
  styleUrl: './add-brand.component.scss'
})
export class AddBrandComponent {
  submenuId = 46;
  submenuName = 'Add Brand';
  submenuCode = 'add_brand';
}
