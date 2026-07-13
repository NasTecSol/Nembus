import { CommonModule } from "@angular/common";
import { Component, EventEmitter, OnInit, Output } from "@angular/core";
import { TranslateModule } from "@ngx-translate/core";
import { environment } from "../../../../../../../environments/environment";
import { HttpClient } from '@angular/common/http';
import { FormsModule } from "@angular/forms";
import { AddProductService } from "../../../../../../core/services/add-product.service";
import { ToastyService } from "../../../../../../core/services/toasty.service";

@Component({
  selector: 'add-tax-info',
  imports: [CommonModule, TranslateModule, FormsModule],
  templateUrl: './tax-info.component.html',

})
export class TaxInfoComponent implements OnInit {
  @Output() stepComplete = new EventEmitter<void>();
  apiUrl = environment.baseUrl;
  tax: any[] = [];
  selectedTaxId!: number;
  selectedTax: any = null;
  constructor(private http: HttpClient, public addProductService: AddProductService, private toasty: ToastyService,) { }

  ngOnInit(): void {
    this.fetchTax();

  }
  fetchTax() {
    this.http.get<any>(`${this.apiUrl}/api/tax-categories/active`).subscribe(res => {
      this.tax = res.data; // keep full objects
    });
  }
  onTaxChange(taxId: number) {
    this.selectedTaxId = Number(taxId);
    this.selectedTax = this.tax.find(t => t.id === this.selectedTaxId);
    this.addProductService.setField('tax_category_id', this.selectedTaxId);
  }

  onSubmit() {
    this.addProductService.submitProduct().subscribe({
      next: (res: any) => {
        console.log('Product created successfully:', res);
        this.addProductService.setCreatedProduct(res.data);
        this.addProductService.clearCreatedProductVariants();
        this.stepComplete.emit();

        const productId = Number(this.addProductService.getCreatedProduct()?.id || 0);
        const hasVariants = !!this.addProductService.getPayload()?.has_variants;

        if (!productId) {
          this.toasty.error('Invalid product id');
          return;
        }

        if (!hasVariants) {
          this.postBarcode(productId);
        }

        this.toasty.success('Product added successfully');
      },
      error: (err) => {
        console.error('Error while adding product:', err);
        this.toasty.error('Failed to add product');
      }
    });
  }

  postBarcode(productId: number) {
    const payload = {
      barcode: this.addProductService.getPayload().barcode || '',
      barcode_type: 'EAN13',
      is_primary: true,
      metadata: {},
      product_id: productId,
      product_variant_id: null
    };

    this.http.post(`${this.addProductService.apiUrl}/api/product-barcodes`, payload)
      .subscribe({
        next: (res) => {
          console.log('Barcode created successfully:', res);
        },
        error: (err) => {
          console.error('Failed to create barcode:', err);
        }
      });
  }



}