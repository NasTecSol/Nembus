import { Component } from "@angular/core";
import { Router } from "@angular/router";
import { countries } from "../../../../utils/country-codes";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TranslateModule } from "@ngx-translate/core";
import { CheckIconComponent } from "../../../../shared/icons/check.component";
@Component({
  selector: "app-itemslist",
  imports: [CommonModule, FormsModule, TranslateModule, CheckIconComponent],
  templateUrl: "./itemslist.component.html",
})
export class ItemslistComponent {
  public tabs: string[] = [
    "Item Details",
    "Dimensions",
    "Inventory Details",
    "Tax Info",
    "WholeSale & UoM",
    "Pricing",
  ];

public tabTranslationMap:any = {
  "Item Details": "ITEM_DETAILS",
  "Dimensions": "DIMENSIONS",
  "Inventory Details": "INVENTORY_DETAILS",
  "Tax Info": "TAX_INFO",
  "WholeSale & UoM": "WHOLESALE_UOM",
  "Pricing": "PRICING",
};

  public activeTab: string = "Item Details";
  public items: any[] = [
    { itemNumber: "0326", name: "Qadsiya", category: "Tabuk" },
    { itemNumber: "0327", name: "Al Noor", category: "Riyadh" },
    { itemNumber: "0328", name: "Al Salam", category: "Jeddah" },
    { itemNumber: "0329", name: "Haram", category: "Makkah" },
    { itemNumber: "0330", name: "Iman", category: "Medina" },
    { itemNumber: "0331", name: "Falah", category: "Dammam" },
    { itemNumber: "0332", name: "Safa", category: "Hail" },
    { itemNumber: "0333", name: "Ameen", category: "Abha" },
    { itemNumber: "0334", name: "Burhan", category: "Jazan" },
    { itemNumber: "0335", name: "Nour", category: "Khobar" },
  ];

  setActiveTab(tab: string) {
    this.activeTab = tab;
  }

  public selectedStore: any = this.items[0];

  selectStore(store: any) {
    this.selectedStore = store;
  }
}
