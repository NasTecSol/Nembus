import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-add-category',
  imports: [CommonModule, TranslateModule],
  templateUrl: './add-category.component.html',
  styleUrl: './add-category.component.scss'
})
export class AddCategoryComponent {
  submenuId = 44;
  submenuName = 'Add Category';
  submenuCode = 'add_category';
}
