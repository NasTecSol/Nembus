import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-category-list',
  imports: [CommonModule, TranslateModule],
  templateUrl: './category-list.component.html',
  styleUrl: './category-list.component.scss'
})
export class CategoryListComponent {
  submenuId = 43;
  submenuName = 'Category List';
  submenuCode = 'category_list';
}
