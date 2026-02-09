import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-process-sale',
  imports: [CommonModule, TranslateModule],
  templateUrl: './process-sale.component.html',
  styleUrl: './process-sale.component.scss'
})
export class ProcessSaleComponent {
  submenuId = 22;
  submenuName = 'Process Sale';
  submenuCode = 'process_sale';
}
