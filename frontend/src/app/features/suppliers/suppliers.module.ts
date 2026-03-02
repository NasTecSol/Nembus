import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SUPPLIERS_ROUTES } from './suppliers.routes';
import { RouterModule } from '@angular/router';

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    RouterModule.forChild(SUPPLIERS_ROUTES)
  ]
})
export class SuppliersModule { }
