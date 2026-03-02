import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { UI_ROUTES } from './ui-modules.routes';
import { RouterModule } from '@angular/router';



@NgModule({
  declarations: [],
  imports: [
    CommonModule,
   RouterModule.forChild(UI_ROUTES)
  ]
})
export class UiModulesModule { }
