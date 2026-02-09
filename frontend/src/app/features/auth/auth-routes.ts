import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SetupComponent } from './components/setup/setup.component';
import { LoginComponent } from './components/login/login.component';

export const Auth_ROUTES: Routes = [
  {
    path: "setup",
    loadComponent: () => import('./components/setup/setup.component').then(m => m.SetupComponent)
  },
  {
    path: "login",
    loadComponent: () => import('./components/login/login.component').then(m => m.LoginComponent)
  },
  {
    path: "",
    redirectTo: "setup",
    pathMatch: "full"
  }
];
