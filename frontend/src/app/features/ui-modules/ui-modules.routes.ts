import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { HomeComponent } from "./components/home/home.component";


export const UI_ROUTES: Routes = [
  {
         path: "",
         redirectTo: "list",
         pathMatch: "full",
       },

 
];

