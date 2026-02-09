import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-close-session',
  imports: [CommonModule, TranslateModule],
  templateUrl: './close-session.component.html',
  styleUrl: './close-session.component.scss'
})
export class CloseSessionComponent {
  submenuId = 33;
  submenuName = 'Close Session';
  submenuCode = 'close_session';
}
