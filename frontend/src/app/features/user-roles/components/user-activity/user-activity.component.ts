import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-user-activity',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './user-activity.component.html',
  styleUrl: './user-activity.component.scss'
})
export class UserActivityComponent {
  submenuId = 12;
  submenuName = 'User Activity';
  submenuCode = 'user_activity';
}
