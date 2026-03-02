import { NgModule } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterModule } from "@angular/router";
import { SETTINGS_ROUTES } from "./setting.routes";


@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    RouterModule.forChild(SETTINGS_ROUTES),
  ],
})
export class SettingsModule {}
