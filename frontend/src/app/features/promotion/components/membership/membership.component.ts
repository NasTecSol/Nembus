import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-membership',
  imports: [CommonModule,TranslateModule],
  templateUrl: './membership.component.html',
})
export class MembershipComponent {
public discountOffers = [
    { number: 123, code: 101, name: "50%" },
    { number: 124, code: 102, name: "30%" },
    { number: 125, code: 103, name: "10%" },
    { number: 126, code: 104, name: "25%" },
    { number: 127, code: 105, name: "15%" },
    { number: 128, code: 106, name: "5%" },
    { number: 129, code: 107, name: "40%" },
    { number: 130, code: 108, name: "60%" },
    { number: 131, code: 109, name: "20%" },
    { number: 132, code: 110, name: "35%" },
    { number: 130, code: 108, name: "60%" },
    { number: 131, code: 109, name: "20%" },
    { number: 132, code: 110, name: "35%" },
  ];
  public activeRow = this.discountOffers[0];

  setActiveRow(row: any) {
    this.activeRow = row;
  }
}
