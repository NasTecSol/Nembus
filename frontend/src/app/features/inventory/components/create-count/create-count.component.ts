import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-create-count',
  imports: [CommonModule, TranslateModule],
  templateUrl: './create-count.component.html',
  styleUrl: './create-count.component.scss'
})
export class CreateCountComponent {
  submenuId = 39;
  submenuName = 'Create Count';
  submenuCode = 'create_count';
}
