import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-record-movement',
  imports: [CommonModule, TranslateModule],
  templateUrl: './record-movement.component.html',
  styleUrl: './record-movement.component.scss'
})
export class RecordMovementComponent {
  submenuId = 37;
  submenuName = 'Record Movement';
  submenuCode = 'record_movement';
}
