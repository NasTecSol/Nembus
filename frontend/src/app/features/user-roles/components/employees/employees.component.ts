import { CommonModule } from "@angular/common";
import { Component } from "@angular/core";
import { CreateEmployeeModalComponent } from "../../../../shared/create-employee-modal/create-employee-modal.component";
import { TranslateModule } from "@ngx-translate/core";

@Component({
  selector: "employees",
  imports: [CommonModule, CreateEmployeeModalComponent, TranslateModule],
  templateUrl: "./employees.component.html",
})
export class EmployeesComponent {
  tableData = [
    { id: "001", name: "Amir Syafiq", role: "Cashier", counterId: "1001" },
    { id: "002", name: "Ayesha Khan", role: "Supervisor", counterId: "1002" },
    { id: "003", name: "Ali Raza", role: "Manager", counterId: "1003" },
    { id: "004", name: "Sara Malik", role: "Cashier", counterId: "1004" },
    { id: "005", name: "Zain Ahmed", role: "Assistant", counterId: "1005" },
    { id: "006", name: "Fatima Noor", role: "Supervisor", counterId: "1006" },
    { id: "007", name: "Usman Tariq", role: "Cashier", counterId: "1007" },
    { id: "008", name: "Hira Shah", role: "Manager", counterId: "1008" },
    { id: "009", name: "Bilal Khan", role: "Assistant", counterId: "1009" },
    { id: "010", name: "Nida Iqbal", role: "Cashier", counterId: "1010" },
  ];
  
  public showModal = false;

  openModal() {
    this.showModal = true;
  }

  confirmCreateEmployee(confirmed: boolean) {
    this.showModal = false;
    if (confirmed) {
      console.log("Employee is added");
    } else {
      console.log("Cancelled");
    }
  }
}
