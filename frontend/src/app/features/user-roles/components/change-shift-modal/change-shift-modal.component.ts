import { Component, Output, EventEmitter } from "@angular/core";
import { TranslateModule } from "@ngx-translate/core";

@Component({
  selector: "change-shift-modal",
  imports: [TranslateModule],
  templateUrl: "./change-shift-modal.component.html",
})
export class ChangeShiftModalComponent {
  @Output() close = new EventEmitter<boolean>();

  confirm() {
    this.close.emit(true);
  }

  cancel() {
    this.close.emit(false);
  }
}
