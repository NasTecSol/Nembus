import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-location-list',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  templateUrl: './location-list.component.html',
  styleUrl: './location-list.component.scss'
})
export class LocationListComponent {
  submenuId = 19;
  submenuName = 'Location List';
  submenuCode = 'location_list';
}
