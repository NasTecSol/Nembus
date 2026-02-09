import { CommonModule } from "@angular/common";
import { Component } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { TranslateModule } from "@ngx-translate/core";
import { CartSettingsComponent } from "../cart-settings/cart-settings.component";
import { StoreSettingsComponent } from "../store-settings/store-settings.component";
import { EmployeeSettingsComponent } from "../employee-settings/employee-settings.component";
import { PrintSettingsComponent } from "../print-settings/print-settings.component";

@Component({
  selector: "app-general-settings",
  imports: [
    CommonModule,
    FormsModule,
    TranslateModule,
    StoreSettingsComponent,
    EmployeeSettingsComponent,
    CartSettingsComponent,
    PrintSettingsComponent,
  ],
  templateUrl: "./general-settings.component.html",
})
export class GeneralSettingsComponent {

  public tabs:string[] = ["General","Store Settings","Employee Settings","Cart Settings","Print Settings"];
  public activeTab:string = "General";

  setActiveTab(tab:any){
    this.activeTab = tab;
  }
}
