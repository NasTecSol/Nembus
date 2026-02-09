import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SetupComponent } from './components/setup/setup.component';

export const Auth_ROUTES: Routes = [
    {
    path: "",
    pathMatch: "full",
    redirectTo: "setup",
  },
  {
    path: "setup",
    component: SetupComponent,

  },
];
