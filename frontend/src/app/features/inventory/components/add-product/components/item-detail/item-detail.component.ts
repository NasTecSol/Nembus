import { CommonModule } from "@angular/common";
import { Component } from "@angular/core";
import { TranslateModule } from "@ngx-translate/core";

@Component({
  selector: "add-item-detail",
  imports: [CommonModule,TranslateModule],
  templateUrl: "./item-detail.component.html",
})
export class ItemDetailComponent {
  public isTyping: boolean = false;

  onClickPlusIcon() {
    this.isTyping = true;
  }

  onConfirm() {
    // Logic to save the value
    this.isTyping = false;
  }

  onCancel() {
    // Logic to clear/reset the input if needed
    this.isTyping = false;
  }
}
