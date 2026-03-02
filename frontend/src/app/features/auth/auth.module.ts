import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { Auth_ROUTES } from './auth-routes';


@NgModule({
  declarations: [],
  imports: [
    CommonModule,
     RouterModule.forChild(Auth_ROUTES),
  ]
})
export class AuthModule { }



