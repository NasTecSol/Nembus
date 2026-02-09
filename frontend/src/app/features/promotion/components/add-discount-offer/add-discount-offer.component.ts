import { Location } from "@angular/common";
import { Component } from "@angular/core";
import { Router } from "@angular/router";
import { TranslateModule } from "@ngx-translate/core";

@Component({
  selector: "app-add-discount-offer",
  imports: [TranslateModule],
  templateUrl: "./add-discount-offer.component.html",
})
export class AddDiscountOfferComponent {
  constructor(private location: Location, private router: Router) {}
  goBack() {
    this.location.back();
  }
  submit() {
    this.router.navigate(["/promotions"]);
  }
}
