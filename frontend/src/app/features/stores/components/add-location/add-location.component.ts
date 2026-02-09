import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-add-location',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './add-location.component.html',
  styleUrl: './add-location.component.scss'
})
export class AddLocationComponent {
  submenuId = 20;
  submenuName = 'Add Location';
  submenuCode = 'add_location';
}
