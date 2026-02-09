import { CommonModule } from "@angular/common";
import { Component } from "@angular/core";
import { TranslateModule } from "@ngx-translate/core";
import { AssignShiftModalComponent } from "../assign-shift-modal/assign-shift-modal.component";
import { ChangeShiftModalComponent } from "../change-shift-modal/change-shift-modal.component";

@Component({
  selector: "shifts",
  imports: [
    CommonModule,
    TranslateModule,
    AssignShiftModalComponent,
    ChangeShiftModalComponent,
  ],
  templateUrl: "./shifts.component.html",
})
export class ShiftsComponent {
  public showAddShiftModal = false;
  public showChangeShiftModal = false;

  openAssignShiftModal() {
    this.showAddShiftModal = true;
  }
  openChangeShiftModal() {
    this.showChangeShiftModal = true;
  }

  confirmAssignShift(confirmed: boolean) {
    this.showAddShiftModal = false;
    if (confirmed) {
      console.log("Shift is assigned!");
    } else {
      console.log("Shift is cancel");
    }
  }
  confirmChangeShift(confirmed: boolean) {
    this.showChangeShiftModal = false;
    if (confirmed) {
      console.log("Shift is changed");
    } else {
      console.log("Cancelled");
    }
  }
}
