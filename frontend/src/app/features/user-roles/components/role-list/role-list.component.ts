import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-role-list',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './role-list.component.html',
  styleUrl: './role-list.component.scss'
})
export class RoleListComponent {
  submenuId = 13;
  submenuName = 'Role List';
  submenuCode = 'role_list';
}
