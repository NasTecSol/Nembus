import { CommonModule } from "@angular/common";
import { Component, Output, EventEmitter, Input } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { countries } from "../../utils/country-codes";
import { TranslateModule } from "@ngx-translate/core";
@Component({
  selector: "app-create-employee-modal",
  imports: [CommonModule, FormsModule, TranslateModule],
  templateUrl: "./create-employee-modal.component.html",
})
export class CreateEmployeeModalComponent {
  @Input() buttonText: string = "Save";
  @Output() close = new EventEmitter<boolean>();
  public countries: any[] = countries;
  public selectedDialCode: string = "+92";
  public phoneNumber: string = "";
  public employee = {
    firstName: "",
    surname: "",
    role: "",
    dialCode: this.selectedDialCode,
    phoneNumber: "",
    loginId: "",
    password: "",
    address: "",
    email: "",
    authCard: "",
    pin: "",
    city: "",
    biometrics: "",
  };
  confirm() {
    this.close.emit(true);
  }

  cancel() {
    this.close.emit(false);
  }
}
