import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-so-list',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './so-list.component.html',
  styleUrl: './so-list.component.scss'
})
export class SoListComponent {
  submenuId = 0;
  submenuName = 'Sales Order List';
  submenuCode = 'so_list';
}
