import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';

interface OptionItem {
  id: string;
  name: string;
}

interface PurchaseOrderLine {
  sku: string;
  productName: string;
  qty: number;
  uom: string;
  unitPrice: number;
  taxPercent: number;
}

@Component({
  selector: 'app-create-po',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslateModule],
  templateUrl: './create-po.component.html',
  styleUrl: './create-po.component.scss'
})
export class CreatePoComponent {
  submenuId = 0;
  submenuName = 'Create Purchase Order';
  submenuCode = 'create_po';

  supplierOptions: OptionItem[] = [
    { id: '1', name: 'Global Tech Solutions Inc.' },
    { id: '2', name: 'Prime Logistics Group' },
    { id: '3', name: 'Office Depot Wholesale' },
  ];

  warehouseOptions: OptionItem[] = [
    { id: '1', name: 'Main Distribution Center (East)' },
    { id: '2', name: 'West Coast Hub' },
    { id: '3', name: 'Regional Outlet 04' },
  ];

  uomOptions: string[] = ['PCS', 'BOX', 'KG'];

  po = {
    supplierId: '',
    warehouseId: '',
    expectedDate: '',
    referenceNumber: '',
    internalNotes: '',
    shippingCost: 45,
  };

  lines: PurchaseOrderLine[] = [
    {
      sku: 'PRD-0042',
      productName: 'High Performance Ethernet Cable 50ft',
      qty: 50,
      uom: 'PCS',
      unitPrice: 12.5,
      taxPercent: 8,
    },
    {
      sku: 'HW-9912',
      productName: 'Smart Hub Controller v3.0',
      qty: 10,
      uom: 'PCS',
      unitPrice: 145,
      taxPercent: 8,
    },
  ];

  addLine(): void {
    this.lines.push({
      sku: '',
      productName: '',
      qty: 1,
      uom: 'PCS',
      unitPrice: 0,
      taxPercent: 0,
    });
  }

  removeLine(index: number): void {
    if (this.lines.length === 1) {
      return;
    }
    this.lines.splice(index, 1);
  }

  lineSubtotal(line: PurchaseOrderLine): number {
    const qty = Number(line.qty) || 0;
    const unitPrice = Number(line.unitPrice) || 0;
    const taxPercent = Number(line.taxPercent) || 0;
    const base = qty * unitPrice;
    return base + base * (taxPercent / 100);
  }

  get subtotal(): number {
    return this.lines.reduce((sum, line) => {
      const qty = Number(line.qty) || 0;
      const unitPrice = Number(line.unitPrice) || 0;
      return sum + qty * unitPrice;
    }, 0);
  }

  get taxTotal(): number {
    return this.lines.reduce((sum, line) => {
      const qty = Number(line.qty) || 0;
      const unitPrice = Number(line.unitPrice) || 0;
      const taxPercent = Number(line.taxPercent) || 0;
      return sum + qty * unitPrice * (taxPercent / 100);
    }, 0);
  }

  get grandTotal(): number {
    const shipping = Number(this.po.shippingCost) || 0;
    return this.subtotal + this.taxTotal + shipping;
  }
}