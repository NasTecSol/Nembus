import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-approve-po',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './approve-po.component.html',
  styleUrl: './approve-po.component.scss'
})
export class ApprovePoComponent {
  submenuId = 0;
  submenuName = 'Approve Purchase Order';
  submenuCode = 'approve_po';
}
