import { CommonModule } from "@angular/common";
import { Component } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { Router, RouterModule } from "@angular/router";
import { TranslateModule } from "@ngx-translate/core";

@Component({
  selector: "app-supplier-list",
  imports: [CommonModule, FormsModule, TranslateModule],
  templateUrl: "./supplier-list.component.html",
})
export class SupplierListComponent {
  public tabs: string[] = ["Information", "Contacts", "History"];
  public activeTab: string = "Information";

public suppliers: any[] = [
  { id: "12345678", name: "Ali", surName: "Traders" },
  { id: "23456789", name: "Sara", surName: "Supplies" },
  { id: "34567890", name: "Ahmed", surName: "Wholesalers" },
  { id: "45678901", name: "Hina", surName: "Imports" },
  { id: "56789012", name: "Usman", surName: "Goods" },
  { id: "67890123", name: "Nadia", surName: "Exports" },
  { id: "78901234", name: "Bilal", surName: "Distributors" },
  { id: "89012345", name: "Zara", surName: "Mart" },
  { id: "90123456", name: "Danish", surName: "Suppliers" },
  { id: "01234567", name: "Sana", surName: "Traders" },
];

  public selectedPurchasing: any = this.suppliers[0];
  public filteredInvoices: any = [];
  public filteredProducts: any = [];

public contacts = [
  {
    contractId: "C001",
    product: "Smartphone",
    onProductBought: 5,
    discounted: 150.5,
    expiry: "2025-06-30",
  },
  {
    contractId: "C002",
    product: "Laptop",
    onProductBought: 3,
    discounted: 0,
    expiry: "2025-07-10",
  },
  {
    contractId: "C003",
    product: "Tablet",
    onProductBought: 8,
    discounted: 300,
    expiry: "2025-06-25",
  },
];

 public history = [
  {
    contractId: "C001",
    dateOfContract: "2025-05-29",
    product: "Smartphone",
    salePrice: 2500.75,
    discounts: 150.5
  },
  {
    contractId: "C002",
    dateOfContract: "2025-05-28",
    product: "Headphones",
    salePrice: 1200.0,
    discounts: 0
  },
  {
    contractId: "C003",
    dateOfContract: "2025-05-27",
    product: "Laptop",
    salePrice: 5600.25,
    discounts: 300
  },
];


  constructor(private router: Router) {}
  setActiveTab(tab: string) {
    this.activeTab = tab;
  }
  ngOnInit(): void {
    if (this.suppliers.length > 0) {
      this.selectPurchasing(this.suppliers[0]);
    }
  }

  selectPurchasing(purchasing: any) {
    this.selectedPurchasing = purchasing;
    this.filteredInvoices = this.contacts.filter(
      (invoice) => invoice.contractId === purchasing.id
    );
    this.filteredProducts = this.history.filter(
      (product) => product.contractId === purchasing.id
    );
  }

  addNewSupplier() {
    this.router.navigate(["/suppliers/add-supplier"]);
  }
}
