import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ADMIN_ROUTES } from "./admin.routes";
import { RouterModule } from '@angular/router';


@NgModule({
  declarations: [],
  imports: [CommonModule, RouterModule.forChild(ADMIN_ROUTES)],
})
export class AdminModule {}
