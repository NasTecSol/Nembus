import { CommonModule } from "@angular/common";
import { Component } from "@angular/core";
import { Router } from "@angular/router";
import { TranslateModule } from "@ngx-translate/core";
import { ItemsgridComponent } from "../itemsgrid/itemsgrid.component";
import { ItemslistComponent } from "../itemslist/itemslist.component";
import { BillsComponent } from "../bills/bills.component";

@Component({
  selector: "app-inventory-home",
  standalone: true,
  imports: [
    CommonModule,
    TranslateModule,
    ItemsgridComponent,
    ItemslistComponent,
    BillsComponent,
  ],
  templateUrl: "./inventory-home.component.html",
})
export class InventoryHomeComponent {
  constructor(private router: Router) {}

  public activeTab: "items" | "bills" = "items";
  public activeItemTab: "grid" | "list" = "grid";

  setActiveTab(tab: "items" | "bills") {
    this.activeTab = tab;
  }

  setActiveItemTab(tab: "grid" | "list") {
    this.activeItemTab = tab;
  }

  addProduct() {
    this.router.navigate(["/inventory/add-product"]);
  }
}
