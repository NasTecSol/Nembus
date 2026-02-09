import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-session-history',
  imports: [CommonModule, TranslateModule],
  templateUrl: './session-history.component.html',
  styleUrl: './session-history.component.scss'
})
export class SessionHistoryComponent {
  submenuId = 31;
  submenuName = 'Session History';
  submenuCode = 'session_history';
}
