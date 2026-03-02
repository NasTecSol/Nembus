import { CommonModule, Location } from "@angular/common";
import { Component } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { TranslateModule } from "@ngx-translate/core";
import { AddProductModal } from "../add-product-modal/add-product-modal.component";
import { InvoiceModalComponent } from "../invoice-modal/invoice-modal.component";

@Component({
  selector: "app-create",
  imports: [
    CommonModule,
    FormsModule,
    TranslateModule,
    AddProductModal,
    InvoiceModalComponent,
  ],
  templateUrl: "./create.component.html",
})
export class CreateComponent {
  public products = [
    {
      id: "02325698",
      name: "Knorr Noodles",
      quantity: 50,
      uom: "Kg",
      price: 500,
      stockInHand: 5,
      discountPercent: "5%",
      discountAmount: 250.0,
      warehouse: "",
      vatCode: "2502",
      total: 250.23,
      totalWithTax: 250.23,
    },
    {
      id: "02325699",
      name: "Maggi Noodles",
      quantity: 30,
      uom: "Kg",
      price: 300,
      stockInHand: 3,
      discountPercent: "10%",
      discountAmount: 90.0,
      warehouse: "",
      vatCode: "2503",
      total: 270.0,
      totalWithTax: 297.0,
    },
  ];

  constructor(private location: Location) {}
  goBack() {
    this.location.back();
  }

  public isAddProductModalOpen = false;
  public isInvoiceModalOpen = false;

  openProductModal() {
    this.isAddProductModalOpen = true;
  }

  closeProductModal() {
    this.isAddProductModalOpen = false;
  }

  handleProductSubmit(product: any) {
    console.log("Selected product:", product);
  }
  openInvoiceModal() {
    this.isInvoiceModalOpen = true;
  }
  closeInvoiceModal() {
    this.isInvoiceModalOpen = false;
  }
}
