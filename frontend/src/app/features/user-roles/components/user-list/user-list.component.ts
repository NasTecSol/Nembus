import { CommonModule } from "@angular/common";
import { Component } from "@angular/core";
import { Router } from "@angular/router";
import { TranslateModule } from "@ngx-translate/core";
import { UserDetailsComponent } from "../user-details/user-details.component";
import { PermissionsComponent } from "../permissions/permissions.component";
import { ShiftsComponent } from "../shifts/shifts.component";
import { CashCountersComponent } from "../cash-counters/cash-counters.component";
import { EmployeesComponent } from "../employees/employees.component";

@Component({
  selector: "app-user-list",
  standalone: true,
  imports: [
    CommonModule,
    TranslateModule,
    UserDetailsComponent,
    PermissionsComponent,
    ShiftsComponent,
    CashCountersComponent,
    EmployeesComponent,
  ],
  templateUrl: "./user-list.component.html",
})
export class UserListComponent {
  constructor(private router: Router) {
    this.setVisibleFormTabs(this.roleActiveTab);
  }

  public roleTabs: string[] = ["ALL", "Cashiers", "Admin", "Managers"];
  public roleActiveTab: string = "ALL";

  public allFormTabs: string[] = [
    "User Details",
    "Permissions",
    "Shifts",
    "Cash Counters",
    "Employees",
  ];

  public visibleFormTabs: string[] = [];
  public formActiveTab: string = "User Details";


  public users: any[] = [
    { id: "12345678", name: "Ali", sName: "Khan" },
    { id: "23456789", name: "Sara", sName: "Raza" },
    { id: "34567890", name: "Ahmed", sName: "Malik" },
    { id: "45678901", name: "Hina", sName: "Iqbal" },
    { id: "56789012", name: "Usman", sName: "Sheikh" },
    { id: "67890123", name: "Nadia", sName: "Farooq" },
    { id: "78901234", name: "Bilal", sName: "Zahid" },
    { id: "89012345", name: "Zara", sName: "Ali" },
    { id: "90123456", name: "Danish", sName: "Hussain" },
    { id: "01234567", name: "Sana", sName: "Rafiq" },
  ];
    public selectedUser: any = this.users[0];

  private roleFormTabsMap: { [key: string]: string[] } = {
    ALL: this.allFormTabs,
    Cashiers: ["User Details", "Shifts", "Cash Counters"],
    Admin: this.allFormTabs,
    Managers: ["User Details", "Permissions", "Shifts", "Employees"],
  };
 public translateKeyMap: { [key: string]: string } = {
    "User Details": "USER_ROLES.USER_DETAILS",
    "Permissions": "USER_ROLES.PERMISSIONS",
    "Shifts": "USER_ROLES.SHIFTS",
    "Cash Counters": "USER_ROLES.CASH_COUNTERS",
    "Employees": "USER_ROLES.EMPLOYEES"
  };
  public roleTabTranslateMap: { [key: string]: string } = {
    ALL: "USER_ROLES.ALL",
    Cashiers: "USER_ROLES.CASHIERS",
    Admin: "USER_ROLES.ADMIN",
    Managers: "USER_ROLES.MANAGERS",
  };

  setActiveTab(tab: string) {
    this.roleActiveTab = tab;
    this.setVisibleFormTabs(tab);
  }

  setVisibleFormTabs(role: string) {
    this.visibleFormTabs = this.roleFormTabsMap[role] || [];
    if (!this.visibleFormTabs.includes(this.formActiveTab)) {
      this.formActiveTab =
        this.visibleFormTabs.length > 0 ? this.visibleFormTabs[0] : "";
    }
  }

  setFormActiveTab(tab: string) {
    this.formActiveTab = tab;
  }

  selectUser(user: any) {
    this.selectedUser = user;
  }

  navigateToAddUser() {
    this.router.navigate(["/users/create"]);
  }
}
