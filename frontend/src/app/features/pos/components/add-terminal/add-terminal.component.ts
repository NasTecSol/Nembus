import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-add-terminal',
  imports: [CommonModule, TranslateModule],
  templateUrl: './add-terminal.component.html',
  styleUrl: './add-terminal.component.scss'
})
export class AddTerminalComponent {
  submenuId = 25;
  submenuName = 'Add Terminal';
  submenuCode = 'add_terminal';
}
