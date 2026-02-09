import { CommonModule, Location } from "@angular/common";
import { Component } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { TranslateModule } from "@ngx-translate/core";
import { CreateEmployeeModalComponent } from "../../../../shared/create-employee-modal/create-employee-modal.component";

@Component({
  selector: "app-add-store",
  imports: [
    CommonModule,
    FormsModule,
    TranslateModule,
    CreateEmployeeModalComponent,
  ],
  templateUrl: "./add-store.component.html",
})
export class AddStoreComponent {
  public storeTabs: string[] = [
    "General",
    "Accounting",
    "Shifts",
    "Employees",
    "Cash Counters",
    "Inventory",
  ];

  public storeTabTranslationMap: any = {
    General: "STORES_LOCATIONS.GENERAL",
    Accounting: "STORES_LOCATIONS.ACCOUNTING",
    Shifts: "STORES_LOCATIONS.SHIFTS_TAB",
    Employees: "STORES_LOCATIONS.EMPLOYEES",
    "Cash Counters": "STORES_LOCATIONS.CASH_COUNTERS",
    Inventory: "STORES_LOCATIONS.INVENTORY",
  };

  public employees = [
    { id: "E001", name: "Amir Syafiq", surname: "Syafiq", position: "Manager" },
    {
      id: "E002",
      name: "Fatima Noor",
      surname: "Noor",
      position: "HR Officer",
    },
    {
      id: "E003",
      name: "Ali Khan",
      surname: "Khan",
      position: "Software Engineer",
    },
    { id: "E004", name: "Zainab Ali", surname: "Ali", position: "UX Designer" },
    {
      id: "E005",
      name: "Bilal Ahmed",
      surname: "Ahmed",
      position: "DevOps Engineer",
    },
    { id: "E006", name: "Sara Iqbal", surname: "Iqbal", position: "Recruiter" },
    { id: "E007", name: "Hamza Tariq", surname: "Tariq", position: "Intern" },
  ];

  public storeActiveTab: string = this.storeTabs[0];
  public currentActiveTabIndex: number = 0;
  public isShiftAdded: boolean = false;
  public isCashCounterAdded: boolean = false;
  constructor(private location: Location) {}

  setActiveTab(tab: string) {
    this.storeActiveTab = tab;
  }
  nextTab() {
    const index = this.storeTabs.indexOf(this.storeActiveTab);
    if (index < this.storeTabs.length - 1) {
      this.storeActiveTab = this.storeTabs[index + 1];
    }
  }
  previousTab() {
    const index = this.storeTabs.indexOf(this.storeActiveTab);
    if (index > 0) {
      this.storeActiveTab = this.storeTabs[index - 1];
    }
  }
  cancel() {
    alert("Cancelled");
  }

  finish() {
    alert("Finished!");
  }
  goBack() {
    this.location.back();
  }
  addShift() {
    this.isShiftAdded = true;
  }
  deleteShift() {
    this.isShiftAdded = !this.isShiftAdded;
  }
  addCashCounter(){
    this.isCashCounterAdded = true
  }
  deleteCashCounter(){
    this.isCashCounterAdded = !this.isCashCounterAdded;
  }
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
