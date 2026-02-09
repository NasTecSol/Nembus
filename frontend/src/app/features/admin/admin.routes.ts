import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { GeneralSettingsComponent } from './components/general-settings/general-settings.component';
import { TaxConfigComponent } from './components/tax-config/tax-config.component';
import { ModuleListComponent } from './components/module-list/module-list.component';
import { MenueManageComponent } from './components/menue-manage/menue-manage.component';


export const ADMIN_ROUTES: Routes = [
   {
         path: "",
         redirectTo: "general",
         pathMatch: "full",
       },
       {
         path: "general",
         component: GeneralSettingsComponent,
       },
       {
         path: "tax",
         component: TaxConfigComponent,
       },
        {
         path: "list",
         component: ModuleListComponent,
       },
        {
         path: "menus",
         component: MenueManageComponent,
       },
         {
         path: "permissions",
         component: MenueManageComponent,
       },

         {
    path: "settings",
    children: [
      {
        path: "",
        redirectTo: "general",
        pathMatch: "full",
      },
      {
        path: "general",
        loadComponent: () =>
          import("./components/general-settings/general-settings.component").then(
            (m) => m.GeneralSettingsComponent
          ),
      },
      {
        path: "tax",
        loadComponent: () =>
          import("./components/tax-config/tax-config.component").then(
            (m) => m.TaxConfigComponent
          ),
      },
   
    
    ],
  },
          {
    path: "ui-modules",
    children: [
      {
        path: "",
        redirectTo: "list",
        pathMatch: "full",
      },
      {
        path: "list",
        loadComponent: () =>
          import("./components/module-list/module-list.component").then(
            (m) => m.ModuleListComponent
          ),
      },  
      {
        path: "menus",
        loadComponent: () =>
          import("./components/menue-manage/menue-manage.component").then(
            (m) => m.MenueManageComponent
          ),
      }, 
      {
        path: "menus",
        loadComponent: () =>
          import("./components/menue-manage/menue-manage.component").then(
            (m) => m.MenueManageComponent
          ),
      },
       {
        path: "permissions",
        loadComponent: () =>
          import("./components/permissions-manage/permissions-manage.component").then(
            (m) => m.PermissionsManageComponent
          ),
      },
   
    
    ],
  },

];


