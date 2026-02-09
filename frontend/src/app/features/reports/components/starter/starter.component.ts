import { CommonModule } from "@angular/common";
import { Component } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { TranslateModule } from "@ngx-translate/core";
import { CheckIconComponent } from "../../../../shared/icons/check.component";

@Component({
  selector: "app-customers",
  imports: [CommonModule, FormsModule, TranslateModule,CheckIconComponent],
  templateUrl: "./starter.component.html",
})
export class StarterComponent {
  public tabs: string[] = ["Customers", "Items", "Sales Employees"];
  public activeTab: string = this.tabs[0];
  public activeIndex: number = 0;

  setActiveTab(tab: any) {
    this.activeTab = tab;
    this.activeIndex = this.tabs.indexOf(tab);
  }
  next() {
    if (this.activeIndex < this.tabs.length - 1) {
      this.activeIndex++;
      this.activeTab = this.tabs[this.activeIndex];
    }
  }
  previous() {
    if (this.activeIndex > 0) {
      this.activeIndex--;
      this.activeTab = this.tabs[this.activeIndex];
    }
  }
  finish() {
    if (this.activeIndex === this.tabs.length - 1) {
      this.activeIndex = 0;
      this.activeTab = this.tabs[this.activeIndex];
    }
  }
}
