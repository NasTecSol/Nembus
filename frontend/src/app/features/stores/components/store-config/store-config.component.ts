import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-store-config',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './store-config.component.html',
  styleUrl: './store-config.component.scss'
})
export class StoreConfigComponent {
  submenuId = 18;
  submenuName = 'Store Configuration';
  submenuCode = 'store_config';
}
