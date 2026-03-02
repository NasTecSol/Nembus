import { CommonModule, Location } from "@angular/common";
import { Component } from "@angular/core";
import { AccordionComponent } from "../../../../shared/accordion/accordion.component";
import { TranslateModule } from "@ngx-translate/core";
import { Router } from "@angular/router";
import { AddModalComponent } from "../add-modal/add-modal.component";
import { CheckIconComponent } from "../../../../shared/icons/check.component";

@Component({
  selector: "app-add-promotion",
  imports: [
    AccordionComponent,
    CommonModule,
    TranslateModule,
    AddModalComponent,
    CheckIconComponent
  ],
  templateUrl: "./add-promotion.component.html",
})
export class AddPromotionComponent {
  public showModal: boolean = false;
  public selectedProduct: any = null;
public categories = [
  {
    name: "All Items",
    enabled: true,
    open: true,
  },
  {
    name: "Frozen Items",
    itemList: [
      "Frozen Peas",
      "Frozen Corn",
      "Frozen Chicken Nuggets",
      "Ice Cream",
      "Frozen French Fries",
    ],
    enabled: true,
    open: false,
  },
  {
    name: "Daily Products",
    itemList: [
      "Milk",
      "Butter",
      "Eggs",
      "Bread",
      "Yogurt",
      "Cheese",
    ],
    enabled: true,
    open: false,
  },
];


  products = [
    { name: "Pepsi 1litre" },
    { name: "7Up 1litre" },
    { name: "Coca Cola 1litre" },
  ];

  addProduct(product: any) {
    this.selectedProduct = product;
    this.showModal = true;
  }
  handlePromotionSubmit(event: { product: any; discount: string }) {
    console.log("Promotion submitted:", event);
    // Add your logic here to save it
    this.showModal = false;
  }
  constructor(private location: Location, private router: Router) {}

  goBack() {
    this.location.back();
  }
  submit() {
    this.router.navigate(["/promotions"]);
  }
}
