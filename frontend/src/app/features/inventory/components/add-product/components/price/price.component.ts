import { Component } from '@angular/core';
import { CheckIconComponent } from "../../../../../../shared/icons/check.component";
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'add-price',
  imports: [CheckIconComponent,TranslateModule],
  templateUrl: './price.component.html',
})
export class PriceComponent {

}
