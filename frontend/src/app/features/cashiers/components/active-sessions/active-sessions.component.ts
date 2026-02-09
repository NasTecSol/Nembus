import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-active-sessions',
  imports: [CommonModule, TranslateModule],
  templateUrl: './active-sessions.component.html',
  styleUrl: './active-sessions.component.scss'
})
export class ActiveSessionsComponent {
  submenuId = 30;
  submenuName = 'Active Sessions';
  submenuCode = 'active_sessions';
}
