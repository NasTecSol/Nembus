import { CommonModule } from "@angular/common";
import { Component } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { Router } from "@angular/router";
import { TranslateModule } from "@ngx-translate/core";

@Component({
  selector: "app-wholesale-list",
  imports: [FormsModule, CommonModule, TranslateModule],
  templateUrl: "./wholesale-list.component.html",
})
export class WholesaleListComponent {
  public wholesaleProducts: any[] = [
    {
      sno: 1,
      itemCode: "3232321",
      itemName: "Rice Bag",
      uom: "kg",
      quantity: 25,
      pricePerUnit: 100,
      stockAvailable: 120,
      total: 2500,
      totalWithTax: 2625,
    },
    {
      sno: 2,
      itemCode: "3232322",
      itemName: "Milk Pack",
      uom: "litre",
      quantity: 18,
      pricePerUnit: 100,
      stockAvailable: 250,
      total: 1800,
      totalWithTax: 1890,
    },
    {
      sno: 3,
      itemCode: "3232323",
      itemName: "LED Bulb",
      uom: "pcs",
      quantity: 40,
      pricePerUnit: 80,
      stockAvailable: 300,
      total: 3200,
      totalWithTax: 3360,
    },
    {
      sno: 4,
      itemCode: "3232324",
      itemName: "Notebook",
      uom: "pcs",
      quantity: 15,
      pricePerUnit: 65,
      stockAvailable: 180,
      total: 975,
      totalWithTax: 1024,
    },
    {
      sno: 5,
      itemCode: "3232325",
      itemName: "Office Chair",
      uom: "pcs",
      quantity: 5,
      pricePerUnit: 900,
      stockAvailable: 20,
      total: 4500,
      totalWithTax: 4725,
    },
    {
      sno: 6,
      itemCode: "3232326",
      itemName: "Course Book",
      uom: "pcs",
      quantity: 21,
      pricePerUnit: 100,
      stockAvailable: 60,
      total: 2100,
      totalWithTax: 2205,
    },
    {
      sno: 7,
      itemCode: "3232327",
      itemName: "Screen Protector",
      uom: "pcs",
      quantity: 32,
      pricePerUnit: 50,
      stockAvailable: 150,
      total: 1600,
      totalWithTax: 1680,
    },
    {
      sno: 8,
      itemCode: "3232328",
      itemName: "Dress Shirt",
      uom: "pcs",
      quantity: 25,
      pricePerUnit: 122,
      stockAvailable: 90,
      total: 3050,
      totalWithTax: 3202,
    },
    {
      sno: 9,
      itemCode: "3232329",
      itemName: "Story Book",
      uom: "pcs",
      quantity: 29,
      pricePerUnit: 30,
      stockAvailable: 100,
      total: 870,
      totalWithTax: 914,
    },
    {
      sno: 10,
      itemCode: "3232330",
      itemName: "Electric Kettle",
      uom: "pcs",
      quantity: 11,
      pricePerUnit: 200,
      stockAvailable: 50,
      total: 2200,
      totalWithTax: 2310,
    },
    {
      sno: 11,
      itemCode: "3232331",
      itemName: "Bluetooth Speaker",
      uom: "pcs",
      quantity: 14,
      pricePerUnit: 1500,
      stockAvailable: 40,
      total: 21000,
      totalWithTax: 22050,
    },
    {
      sno: 12,
      itemCode: "3232332",
      itemName: "Water Bottle",
      uom: "pcs",
      quantity: 50,
      pricePerUnit: 60,
      stockAvailable: 200,
      total: 3000,
      totalWithTax: 3150,
    },
    {
      sno: 13,
      itemCode: "3232333",
      itemName: "USB Cable",
      uom: "pcs",
      quantity: 100,
      pricePerUnit: 25,
      stockAvailable: 500,
      total: 2500,
      totalWithTax: 2625,
    },
    {
      sno: 14,
      itemCode: "3232334",
      itemName: "Desk Lamp",
      uom: "pcs",
      quantity: 20,
      pricePerUnit: 450,
      stockAvailable: 30,
      total: 9000,
      totalWithTax: 9450,
    },
    {
      sno: 15,
      itemCode: "3232335",
      itemName: "Notebook Set",
      uom: "set",
      quantity: 18,
      pricePerUnit: 300,
      stockAvailable: 90,
      total: 5400,
      totalWithTax: 5670,
    },
  ];
  public tabs: string[] = ["price", "item", "short"];
  public activeBtn: string = "price";
 public notified : boolean = false
  constructor(private router: Router) {}

  addProduct() {
    this.router.navigate(["/inventory/add-product"]);
  }

  showSidebar: boolean = false;
  toggleSidebar() {
    this.showSidebar = !this.showSidebar;
  }
  setActiveTab(tab: any) {
    this.activeBtn = tab;
  }
  notify(){
    this.notified = !this.notified
  }
}
