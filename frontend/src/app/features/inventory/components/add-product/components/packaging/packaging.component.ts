import { Component,OnInit } from '@angular/core';
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TranslateModule } from "@ngx-translate/core";

interface PackagingConversion {
  unitType: string;
  value: number | null;
}

@Component({
  selector: 'app-packaging',
  imports: [CommonModule, FormsModule, TranslateModule],
  templateUrl: './packaging.component.html',
  styleUrl: './packaging.component.scss'
})
export class PackagingComponent implements OnInit {
  baseUnit: string = 'Liter'; // This should come from the selected base unit in previous tab
  
  packagingConversions: PackagingConversion[] = [
    {
      unitType: '',
      value: null
    }
  ];

  ngOnInit(): void {
    // Initialize with base unit from service or parent component
    // this.baseUnit = this.productService.getBaseUnit();
  }

  addConversion(): void {
    this.packagingConversions.push({
      unitType: '',
      value: null
    });
  }

  removeConversion(index: number): void {
    if (this.packagingConversions.length > 1) {
      this.packagingConversions.splice(index, 1);
    }
  }

  onSubmit(): void {
    // Validate and save packaging conversions
    console.log('Packaging conversions:', this.packagingConversions);
  }
}