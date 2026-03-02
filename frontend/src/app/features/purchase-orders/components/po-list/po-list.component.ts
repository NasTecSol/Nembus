import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-po-list',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './po-list.component.html',
  styleUrl: './po-list.component.scss'
})
export class PoListComponent {
  submenuId = 0;
  submenuName = 'Purchase Order List';
  submenuCode = 'po_list';
}
