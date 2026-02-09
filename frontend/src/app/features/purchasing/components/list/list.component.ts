import { CommonModule } from "@angular/common";
import { Component, OnInit } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { Router, RouterLink } from "@angular/router";
import { TranslateModule } from "@ngx-translate/core";

@Component({
  selector: "app-list",
  imports: [CommonModule, TranslateModule, FormsModule, RouterLink],
  templateUrl: "./list.component.html",
})
export class ListComponent implements OnInit {
  public tabs: string[] = ["Invoices", "List of Products"];
  public purchasingTabTranslationMap: any = {
    "Invoices": "PURCHASING.INVOICES",
    "List of Products": "PURCHASING.LIST_OF_PRODUCTS",
  };
  public activeTab: string = "Invoices";

  public purchasings: any[] = [
    { id: "12345678", supplierName: "Ali Traders", noOfProducts: 120 },
    { id: "23456789", supplierName: "Sara Supplies", noOfProducts: 85 },
    { id: "34567890", supplierName: "Ahmed Wholesalers", noOfProducts: 150 },
    { id: "45678901", supplierName: "Hina Imports", noOfProducts: 95 },
    { id: "56789012", supplierName: "Usman Goods", noOfProducts: 200 },
    { id: "67890123", supplierName: "Nadia Exports", noOfProducts: 70 },
    { id: "78901234", supplierName: "Bilal Distributors", noOfProducts: 180 },
    { id: "89012345", supplierName: "Zara Mart", noOfProducts: 140 },
    { id: "90123456", supplierName: "Danish Suppliers", noOfProducts: 90 },
    { id: "01234567", supplierName: "Sana Traders", noOfProducts: 110 },
  ];

  public selectedPurchasing: any = this.purchasings[0];
  public filteredInvoices: any = [];
  public filteredProducts: any = [];

  public invoices = [
    {
      billNo: "B001",
      amount: 2500.75,
      paymentMethod: "Credit Card",
      dateTime: "2025-05-29T14:30:00",
      noOfProducts: 5,
      discount: 150.5,
      purchasingId: "12345678",
    },
    {
      billNo: "B002",
      amount: 1200.0,
      paymentMethod: "Cash",
      dateTime: "2025-05-28T10:15:00",
      noOfProducts: 3,
      discount: 0,
      purchasingId: "23456789",
    },
    {
      billNo: "B003",
      amount: 5600.25,
      paymentMethod: "Bank Transfer",
      dateTime: "2025-05-27T18:45:00",
      noOfProducts: 8,
      discount: 300,
      purchasingId: "12345678",
    },
  ];
  public products: any[] = [
    {
      productId: "F001",
      productName: "Basmati Rice",
      priceBeforeDiscount: 1200,
      priceAfterDiscount: 1050,
      purchasingId: "12345678",
    },
    {
      productId: "F002",
      productName: "Sunflower Cooking",
      priceBeforeDiscount: 550,
      priceAfterDiscount: 500,
      purchasingId: "23456789",
    },
    {
      productId: "F003",
      productName: "Premium Tea",
      priceBeforeDiscount: 300,
      priceAfterDiscount: 270,
      purchasingId: "12345678",
    },
    {
      productId: "F004",
      productName: "Wheat Flour",
      priceBeforeDiscount: 950,
      priceAfterDiscount: 880,
      purchasingId: "34567890",
    },
    {
      productId: "F005",
      productName: "Fresh Eggs",
      priceBeforeDiscount: 300,
      priceAfterDiscount: 280,
      purchasingId: "12345678",
    },
  ];

  constructor(private router: Router) {}
  setActiveTab(tab: string) {
    this.activeTab = tab;
  }
  ngOnInit(): void {
    if (this.purchasings.length > 0) {
      this.selectPurchasing(this.purchasings[0]);
    }
  }

  selectPurchasing(purchasing: any) {
    this.selectedPurchasing = purchasing;
    this.filteredInvoices = this.invoices.filter(
      (invoice) => invoice.purchasingId === purchasing.id
    );
    this.filteredProducts = this.products.filter(
      (product) => product.purchasingId === purchasing.id
    );
  }

  createNewOrder() {
    this.router.navigate(["/purchasing/create"]);
  }
}
