import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-create-po',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './create-po.component.html',
  styleUrl: './create-po.component.scss'
})
export class CreatePoComponent {
  submenuId = 0;
  submenuName = 'Create Purchase Order';
  submenuCode = 'create_po';
}
