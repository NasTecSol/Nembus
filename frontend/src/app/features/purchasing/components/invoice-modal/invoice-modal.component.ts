import { Component, Input, Output, EventEmitter } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-invoice-modal',
  standalone: true,
  imports: [CommonModule, FormsModule,TranslateModule],
  templateUrl: './invoice-modal.component.html'
})
export class  InvoiceModalComponent {
  @Input() isOpen = false;
  @Output() close = new EventEmitter<void>();
  @Output() submit = new EventEmitter<any>();
 invoiceItems = [
    { productName: 'Knorr Noodles', quantity: 40, price: 400, discount: '5%', total: 380 },
    { productName: 'Maggi Noodles', quantity: 30, price: 300, discount: '10%', total: 270 },
    { productName: 'Pepsi Bottle', quantity: 10, price: 120, discount: '0%', total: 120 },
    { productName: 'Knorr Noodles', quantity: 40, price: 400, discount: '5%', total: 380 },
    { productName: 'Maggi Noodles', quantity: 30, price: 300, discount: '10%', total: 270 },
    { productName: 'Pepsi Bottle', quantity: 10, price: 120, discount: '0%', total: 120 },
  
    
  ];
  handleSubmit() {
    
  }
  handleClose() {
    this.close.emit();
  }
}


