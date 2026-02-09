import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-open-session',
  imports: [CommonModule, TranslateModule],
  templateUrl: './open-session.component.html',
  styleUrl: './open-session.component.scss'
})
export class OpenSessionComponent {
  submenuId = 32;
  submenuName = 'Open Session';
  submenuCode = 'open_session';
}
