import { CommonModule } from "@angular/common";
import { Component, EventEmitter, OnInit, Output } from "@angular/core";
import { TranslateModule } from "@ngx-translate/core";
import { environment } from "../../../../../../../environments/environment";
import { HttpClient } from '@angular/common/http';
import { FormsModule } from "@angular/forms";
import { AddProductService } from "../../../../../../core/services/add-product.service";

@Component({
  selector: 'add-wholesale',
  imports: [CommonModule, TranslateModule, FormsModule],
  templateUrl: './wholesale.component.html',
})
export class WholesaleComponent implements OnInit {
  @Output() stepComplete = new EventEmitter<void>();
  apiUrl = environment.baseUrl;
  selectedParentId: number | null = null;
  uom: { id: number; name: string; code: string; type: string }[] = [];
  selectedUomId: number | null = null;
  constructor(private http: HttpClient, private addProductService: AddProductService) { }
  ngOnInit(): void {


    const product = this.addProductService.getCreatedProduct();

    if (product?.base_uom_id) {
      this.fetchUomById(product.base_uom_id);
    }
  }

  uomName: string = '';

  fetchUomById(id: number) {
    this.http.get<any>(`${this.apiUrl}/api/uoms/${id}`).subscribe({
      next: (res) => {
        this.uomName = res.data.name;
        this.addProductService.setBaseUnit(res.data.name);
        this.stepComplete.emit();
        console.log('Fetched UOM:', res.data);
      },
      error: (err) => {
        console.error('Failed to fetch UOM by id', err);
      }
    });
  }




}