import { Component, Input, Output, EventEmitter } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-add-product-modal',
  standalone: true,
  imports: [CommonModule, FormsModule,TranslateModule],
  templateUrl: './add-product-modal.component.html'
})
export class AddProductModal {
  @Input() isOpen = false;
  @Output() close = new EventEmitter<void>();
  @Output() submit = new EventEmitter<any>();

  products = [
    { id: '0234', name: 'Knoor Noodles', isAdded: false },
    { id: '0567', name: 'Lays Chips', isAdded: false },
    { id: '0789', name: 'Pepsi Can', isAdded: false }
  ];

  addProduct(product: any) {
    product.isAdded = !product.isAdded;
    //this.submit.emit(product);
  }

  handleClose() {
    this.close.emit();
  }
}
