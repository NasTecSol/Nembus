import { CommonModule } from "@angular/common";
import { Component } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { Router } from "@angular/router";
import { AccordionComponent } from "../../../../shared/accordion/accordion.component";
import { countries } from "../../../../utils/country-codes";
import { TranslateModule } from "@ngx-translate/core";
@Component({
  selector: "app-storeslist",
  imports: [CommonModule, FormsModule, TranslateModule, AccordionComponent],
  templateUrl: "./storeslist.component.html",
})
export class StoreslistComponent {
  constructor(private router: Router) {}

  public tabs: string[] = [
    "General",
    "Accounting",
    "Employees",
    "Cash Counters",
    "Shifts",
    "Settings",
  ];
  public tabTranslationMap: any = {
  "General": "GENERAL",
  "Accounting": "ACCOUNTING",
  "Employees": "EMPLOYEES",
  "Cash Counters": "CASH_COUNTERS",
  "Shifts": "SHIFTS_TAB",
  "Settings": "SETTINGS"
};
  public activeTab: string = "General";

  public showAddEmployeeForm: boolean = false;
  public showCashForm: boolean = false;
  public showShiftForm: boolean = false;

  public countries: any[] = countries;
  public selectedDialCode: string = "+92";
  public phoneNumber: string = "";

  public stores: any[] = [
    {
      storeNo: "0326",
      storeName: "Qadsiya",
      location: "Tabuk",
      inactive: "No",
      nettable: "Yes",
      binLocations: "Enabled",
    },
    {
      storeNo: "0327",
      storeName: "Al Noor",
      location: "Riyadh",
      inactive: "Yes",
      nettable: "No",
      binLocations: "Disabled",
    },
    {
      storeNo: "0328",
      storeName: "Al Salam",
      location: "Jeddah",
      inactive: "No",
      nettable: "Yes",
      binLocations: "Enabled",
    },
    {
      storeNo: "0329",
      storeName: "Haram",
      location: "Makkah",
      inactive: "No",
      nettable: "No",
      binLocations: "Disabled",
    },
    {
      storeNo: "0330",
      storeName: "Iman",
      location: "Medina",
      inactive: "Yes",
      nettable: "Yes",
      binLocations: "Enabled",
    },
    {
      storeNo: "0331",
      storeName: "Falah",
      location: "Dammam",
      inactive: "No",
      nettable: "Yes",
      binLocations: "Enabled",
    },
    {
      storeNo: "0332",
      storeName: "Safa",
      location: "Hail",
      inactive: "Yes",
      nettable: "No",
      binLocations: "Disabled",
    },
    {
      storeNo: "0333",
      storeName: "Ameen",
      location: "Abha",
      inactive: "No",
      nettable: "Yes",
      binLocations: "Enabled",
    },
    {
      storeNo: "0334",
      storeName: "Burhan",
      location: "Jazan",
      inactive: "No",
      nettable: "Yes",
      binLocations: "Enabled",
    },
    {
      storeNo: "0335",
      storeName: "Nour",
      location: "Khobar",
      inactive: "Yes",
      nettable: "No",
      binLocations: "Disabled",
    },
  ];

  // employees.component.ts
  public employees = [
    { id: "E001", name: "Amir Syafiq", position: "Manager" },
    { id: "E002", name: "Fatima Noor", position: "HR Officer" },
    { id: "E003", name: "Ali Khan", position: "Software Engineer" },
    { id: "E004", name: "Zainab Ali", position: "UX Designer" },
    { id: "E005", name: "Bilal Ahmed", position: "DevOps Engineer" },
    { id: "E006", name: "Sara Iqbal", position: "Recruiter" },
    { id: "E007", name: "Hamza Tariq", position: "Intern" },
    { id: "E008", name: "Nida Raza", position: "Project Lead" },
    { id: "E009", name: "Usman Javed", position: "Product Manager" },
    { id: "E010", name: "Ayesha Siddiq", position: "Scrum Master" },
  ];
  public cashCounters = [
    {
      counterId: "C001",
      name: "Qadsiya Drawer",
      assignedTo: "Aamir",
      shifts: "12am-8am",
    },
    {
      counterId: "C002",
      name: "Qadsiya Drawer",
      assignedTo: "Kashif",
      shifts: "8am-4pm",
    },
    {
      counterId: "C003",
      name: "Qadsiya Drawer",
      assignedTo: "Sara",
      shifts: "4pm-12am",
    },
    {
      counterId: "C004",
      name: "Qadsiya Drawer",
      assignedTo: "Ahmed",
      shifts: "12am-8am",
    },
    {
      counterId: "C005",
      name: "Qadsiya Drawer",
      assignedTo: "Fatima",
      shifts: "8am-4pm",
    },
    {
      counterId: "C006",
      name: "Qadsiya Drawer",
      assignedTo: "Bilal",
      shifts: "4pm-12am",
    },
    {
      counterId: "C007",
      name: "Qadsiya Drawer",
      assignedTo: "Nida",
      shifts: "12am-8am",
    },
    {
      counterId: "C008",
      name: "Qadsiya Drawer",
      assignedTo: "Hassan",
      shifts: "8am-4pm",
    },
    {
      counterId: "C009",
      name: "Qadsiya Drawer",
      assignedTo: "Zainab",
      shifts: "4pm-12am",
    },
    {
      counterId: "C010",
      name: "Qadsiya Drawer",
      assignedTo: "Usman",
      shifts: "12am-8am",
    },
  ];

  public shifts = [
    {
      shiftTime: "12am-8am",
      name: "Midnight Shift",
      assignedTo: "Aamir",
      setting: "Automatic",
    },
    {
      shiftTime: "8am-4pm",
      name: "Morning Shift",
      assignedTo: "Kashif",
      setting: "Manual",
    },
    {
      shiftTime: "4pm-12am",
      name: "Evening Shift",
      assignedTo: "Sara",
      setting: "Automatic",
    },
    {
      shiftTime: "10pm-6am",
      name: "Night Shift",
      assignedTo: "Ahmed",
      setting: "Manual",
    },
    {
      shiftTime: "9am-5pm",
      name: "Day Shift",
      assignedTo: "Fatima",
      setting: "Automatic",
    },
    {
      shiftTime: "6am-2pm",
      name: "Early Shift",
      assignedTo: "Bilal",
      setting: "Manual",
    },
    {
      shiftTime: "2pm-10pm",
      name: "Late Shift",
      assignedTo: "Nida",
      setting: "Automatic",
    },
    {
      shiftTime: "11pm-7am",
      name: "Graveyard Shift",
      assignedTo: "Hassan",
      setting: "Manual",
    },
    {
      shiftTime: "7am-3pm",
      name: "Sunrise Shift",
      assignedTo: "Zainab",
      setting: "Automatic",
    },
    {
      shiftTime: "3pm-11pm",
      name: "Twilight Shift",
      assignedTo: "Usman",
      setting: "Manual",
    },
  ];
  activeIndex: number | null = null;

  settings = [
    {
      name: "Store Cart Settings",
      itemList: ["On Product Removal, is biometric required?"],
      enabled: true,
      open: false,
    },
    {
      name: "Store Employee Settings",
      itemList: ["On Product Removal, is biometric required?"],
      enabled: false,
      open: false,
    },
  ];
  addStore() {
    this.router.navigate(["/stores/add-store"]);
  }
  setFormActiveTab(tab: string) {
    this.activeTab = tab;
  }
  onAddEmployeeClick() {
    this.showAddEmployeeForm = !this.showAddEmployeeForm;
  }
  onAddCashClick() {
    this.showCashForm = !this.showCashForm;
  }
  onAddShiftClick() {
    this.showShiftForm = !this.showShiftForm;
  }

  toggle(i: number) {
    this.settings[i].open = !this.settings[i].open;
  }

  public selectedStore: any = this.stores[0];

  selectStore(store: any) {
    this.selectedStore = store;
  }
}
