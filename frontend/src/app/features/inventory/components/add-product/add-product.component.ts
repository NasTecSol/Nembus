import { CommonModule, Location } from "@angular/common";
import { Component, OnDestroy, OnInit } from "@angular/core";
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
import { AddProductService } from "../../../../core/services/add-product.service";
import { Subscription } from "rxjs";

@Component({
  selector: "app-add-product",
  standalone: true,
  imports: [
    FormsModule,
    CommonModule,
    TranslateModule,
    ItemDetailComponent,
    InventoryDetailComponent,
    TaxInfoComponent,
    WholesaleComponent,
    PriceComponent,
    PackagingComponent,
  ],
  templateUrl: "./add-product.component.html",
})
export class AddProductComponent implements OnInit, OnDestroy {
  private readonly tabsWithVariants: string[] = [
    "Item Details",
    "Tax Info",
    "Product Variants",
    "Inventory Details",
    "WholeSale & UoM",
    "Prices",
  ];
  private readonly tabsWithoutVariants: string[] = [
    "Item Details",
    "Tax Info",
    "Inventory Details",
    "WholeSale & UoM",
    "Prices",
  ];

  private hasVariantsSub?: Subscription;

  public roleTabs: string[] = [...this.tabsWithVariants];
  public roleActiveTab: string = "Item Details";
  tabTranslationMap: Record<string, string> = {
    "Item Details": "INVENTORY.ITEM_DETAILS",
    "Dimensions": "INVENTORY.DIMENSIONS",
    "Inventory Details": "INVENTORY.INVENTORY_DETAILS",
    "Tax Info": "INVENTORY.TAX_INFO",
    "WholeSale & UoM": "INVENTORY.WHOLESALE_UOM",
    "Product Variants": "INVENTORY.PRODUCT_VARIANTS",
    "Prices": "INVENTORY.PRICES",
  };

  constructor(
    private location: Location,
    private addProductService: AddProductService,
    private router: Router
  ) { }

  ngOnInit(): void {
    this.updateTabs(this.addProductService.getHasVariants());
    this.hasVariantsSub = this.addProductService
      .getHasVariants$()
      .subscribe((hasVariants) => this.updateTabs(hasVariants));
  }

  ngOnDestroy(): void {
    this.hasVariantsSub?.unsubscribe();
  }

  private updateTabs(hasVariants: boolean): void {
    this.roleTabs = hasVariants
      ? [...this.tabsWithVariants]
      : [...this.tabsWithoutVariants];

    if (!this.roleTabs.includes(this.roleActiveTab)) {
      this.roleActiveTab = this.roleTabs[0];
    }
  }

  setActiveTab(tab: string) {
    this.roleActiveTab = tab;
  }

  markCurrentTabCompleteAndAdvance() {
    const currentIndex = this.roleTabs.indexOf(this.roleActiveTab);
    const nextTab = this.roleTabs[currentIndex + 1];
    if (nextTab) {
      this.roleActiveTab = nextTab;
    }
  }

  onStepComplete(): void {
    if (this.roleActiveTab === "Prices") {
      void this.router.navigate(["/products/list"]);
      return;
    }

    this.markCurrentTabCompleteAndAdvance();
  }

  goBack() {
    this.location.back();
  }
}
