import { CommonModule, Location } from "@angular/common";
import { Component, OnInit } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { Router, ActivatedRoute } from "@angular/router";
import { TranslateModule } from "@ngx-translate/core";

@Component({
  selector: "app-detail",
  imports: [CommonModule, FormsModule, TranslateModule],
  templateUrl: "./detail.component.html",
})
export class DetailComponent {
  public employeeData: any;
  public notes = [
    { quantity: 33, denomination: 500 },
    { quantity: 33, denomination: 100 },
    { quantity: 33, denomination: 50 },
    { quantity: 33, denomination: 20 },
    { quantity: 33, denomination: 10 },
    { quantity: 33, denomination: 5 },
    { quantity: 33, denomination: 1 },
  ];
  public creditCards = [
  { cardType: 'Visa', amount: 12000 },
  { cardType: 'Mada', amount: 7000 },
  { cardType: 'Mastercard', amount: 5000 },
];


  constructor(
    private router: Router,
    private route: ActivatedRoute,
    private location: Location
  ) {
    // Try to get navigation state data here
    const nav = this.router.getCurrentNavigation();
    this.employeeData = nav?.extras?.state?.["data"];
    console.log("Data from navigation state:", this.employeeData);
  }
  goBack() {
    this.location.back();
  }
}
