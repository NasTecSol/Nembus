import { CommonModule, Location } from "@angular/common";
import { Component } from "@angular/core";
import { Router } from "@angular/router";
import { TranslateModule } from "@ngx-translate/core";
import { CheckIconComponent } from "../../../../shared/icons/check.component";

@Component({
  selector: "app-product-detail",
  imports: [CommonModule, TranslateModule, CheckIconComponent],
  templateUrl: "./product-detail.component.html",
})
export class ProductDetailComponent {
  public tabs: string[] = [
    "INVENTORY.ITEM_DETAILS",
    "INVENTORY.DIMENSIONS",
    "INVENTORY.INVENTORY_DETAILS",
    "INVENTORY.TAX_INFO",
    "INVENTORY.WHOLESALE_UOM",
    "INVENTORY.PRICING",
  ];

  public activeTab: string = this.tabs[0];
  public product: any;
  constructor(private router: Router, private location: Location) {
    // Try to get navigation state data here
    const nav = this.router.getCurrentNavigation();
    this.product = nav?.extras?.state?.["data"];
    console.log("Data from navigation state:", this.product);
  }
  setActiveTab(tab: string) {
    this.activeTab = tab;
  }
  goBack() {
    this.location.back();
  }
}
