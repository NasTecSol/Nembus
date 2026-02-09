import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-create-so',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './create-so.component.html',
  styleUrl: './create-so.component.scss'
})
export class CreateSoComponent {
  submenuId = 0;
  submenuName = 'Create Sales Order';
  submenuCode = 'create_so';
}
