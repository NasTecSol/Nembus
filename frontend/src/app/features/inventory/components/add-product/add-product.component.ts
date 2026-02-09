import { CommonModule,Location } from "@angular/common";
import { Component } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { ItemDetailComponent } from "./components/item-detail/item-detail.component";
import { DimensionComponent } from "./components/dimension/dimension.component";
import { InventoryDetailComponent } from "./components/inventory-detail/inventory-detail.component";
import { TaxInfoComponent } from "./components/tax-info/tax-info.component";
import { WholesaleComponent } from "./components/wholesale/wholesale.component";
import { PriceComponent } from "./components/price/price.component";
import { Router } from "@angular/router";
import { TranslateModule } from "@ngx-translate/core";
import { PackagingComponent } from "./components/packaging/packaging.component";

@Component({
  selector: "app-add-product",
  imports: [
    FormsModule,
    CommonModule,
    TranslateModule,
    ItemDetailComponent,
    DimensionComponent,
    InventoryDetailComponent,
    TaxInfoComponent,
    WholesaleComponent,
    PriceComponent,
    PackagingComponent
],
  templateUrl: "./add-product.component.html",
})
export class AddProductComponent {
  public roleTabs: string[] = [
    "Item Details",
    "Dimensions",
    "Inventory Details",
    "Tax Info",
    "WholeSale & UoM",
    "Packaging",
    "Prices",
  ];
  public roleActiveTab: string = "Item Details";
  tabTranslationMap:any = {
    'Item Details': 'INVENTORY.ITEM_DETAILS',
    'Dimensions': 'INVENTORY.DIMENSIONS',
    'Inventory Details': 'INVENTORY.INVENTORY_DETAILS',
    'Tax Info': 'INVENTORY.TAX_INFO',
    'WholeSale & UoM': 'INVENTORY.WHOLESALE_UOM',
    'Packaging': 'INVENTORY.PACKAGING',
    'Prices': 'INVENTORY.PRICES',
  };

  constructor(private location:Location){
  }

  setActiveTab(tab: string) {
    this.roleActiveTab = tab;
  }
  goBack(){
    this.location.back();
  }
}
