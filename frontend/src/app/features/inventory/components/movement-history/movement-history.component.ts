import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-movement-history',
  imports: [CommonModule, TranslateModule],
  templateUrl: './movement-history.component.html',
  styleUrl: './movement-history.component.scss'
})
export class MovementHistoryComponent {
  submenuId = 36;
  submenuName = 'Movement History';
  submenuCode = 'movement_history';
}
