import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-terminal-list',
  imports: [CommonModule, TranslateModule],
  templateUrl: './terminal-list.component.html',
  styleUrl: './terminal-list.component.scss'
})
export class TerminalListComponent {
  submenuId = 24;
  submenuName = 'Terminal List';
  submenuCode = 'terminal_list';
}
