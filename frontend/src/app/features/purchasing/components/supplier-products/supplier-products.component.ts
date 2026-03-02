import { CommonModule, Location } from "@angular/common";
import { Component } from "@angular/core";
import { TranslateModule } from "@ngx-translate/core";

@Component({
  selector: "app-supplier-products",
  imports: [CommonModule, TranslateModule],
  templateUrl: "./supplier-products.component.html",
})
export class SupplierProductsComponent {
  public productList: any[] = [
    {
      productId: "232325698",
      productName: "Knorr Noodles",
      uom: "Pack",
      quantity: 10,
      pricePerUnit: 120,
      stockAvailable: 50,
      discount: "5%",
      total: 1140.0,
      totalWithTax: 1197.0,
      vatGroup: "5%",
    },
    {
      productId: "232325699",
      productName: "Nestlé Milkpak",
      uom: "Liter",
      quantity: 8,
      pricePerUnit: 180,
      stockAvailable: 'Out Of Stock',
      discount: "3%",
      total: 1396.8,
      totalWithTax: 1466.64,
      vatGroup: "5%",
    },
    {
      productId: "232325700",
      productName: "Shan Biryani Masala",
      uom: "Box",
      quantity: 15,
      pricePerUnit: 90,
      stockAvailable: 100,
      discount: "10%",
      total: 1215.0,
      totalWithTax: 1275.75,
      vatGroup: "5%",
    },
    {
      productId: "232325701",
      productName: "Lays Chips",
      uom: "Pack",
      quantity: 20,
      pricePerUnit: 50,
      stockAvailable: 200,
      discount: "2%",
      total: 980.0,
      totalWithTax: 1029.0,
      vatGroup: "5%",
    },
    {
      productId: "232325702",
      productName: "Mitchell's Jam",
      uom: "Jar",
      quantity: 6,
      pricePerUnit: 250,
      stockAvailable: 30,
      discount: "5%",
      total: 1425.0,
      totalWithTax: 1496.25,
      vatGroup: "5%",
    },
  ];

  constructor(private location: Location) {}

  goBack() {
    this.location.back();
  }
}
