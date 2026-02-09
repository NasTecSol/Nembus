import { CommonModule } from "@angular/common";
import { Component } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { Router } from "@angular/router";
import { TranslateModule } from "@ngx-translate/core";

@Component({
  selector: "app-list",
  imports: [CommonModule, FormsModule, TranslateModule],
  templateUrl: "./list.component.html",
})
export class ListComponent {
  public dailyCash: any[] = [
    {
      empId: "EMP001",
      empName: "Ali Raza",
      dailyCollection: 12000,
      cash: 7000,
      card: 3000,
      coupens: 1000,
      receipt: 11000,
      leftoverAmount: 500,
      difference: 1000,
      netTotal: 12000,
      status: "Submitted",
    },
    {
      empId: "EMP002",
      empName: "Sana Tariq",
      dailyCollection: 15000,
      cash: 8000,
      card: 5000,
      coupens: 1500,
      receipt: 14800,
      leftoverAmount: 200,
      difference: 200,
      netTotal: 15000,
      status: "Pending",
    },
    {
      empId: "EMP003",
      empName: "Usman Ghani",
      dailyCollection: 9000,
      cash: 6000,
      card: 2000,
      coupens: 500,
      receipt: 8700,
      leftoverAmount: 300,
      difference: 300,
      netTotal: 9000,
      status: "Approval Pending",
    },
    {
      empId: "EMP004",
      empName: "Fatima Noor",
      dailyCollection: 11000,
      cash: 6000,
      card: 4000,
      coupens: 500,
      receipt: 10500,
      leftoverAmount: 500,
      difference: 500,
      netTotal: 11000,
      status: "Re-Count",
    },
    {
      empId: "EMP005",
      empName: "Hamza Ali",
      dailyCollection: 8000,
      cash: 5000,
      card: 2000,
      coupens: 500,
      receipt: 7700,
      leftoverAmount: 300,
      difference: 300,
      netTotal: 8000,
      status: "Submitted",
    },
    {
      empId: "EMP006",
      empName: "Ayesha Khan",
      dailyCollection: 10000,
      cash: 5000,
      card: 4000,
      coupens: 800,
      receipt: 9800,
      leftoverAmount: 200,
      difference: 200,
      netTotal: 10000,
      status: "Pending",
    },
    {
      empId: "EMP007",
      empName: "Zain Abbas",
      dailyCollection: 13000,
      cash: 8000,
      card: 4000,
      coupens: 500,
      receipt: 12500,
      leftoverAmount: 500,
      difference: 500,
      netTotal: 13000,
      status: "Approval Pending",
    },
    {
      empId: "EMP008",
      empName: "Rabia Shah",
      dailyCollection: 14000,
      cash: 9000,
      card: 4000,
      coupens: 700,
      receipt: 13500,
      leftoverAmount: 500,
      difference: 500,
      netTotal: 14000,
      status: "Re-Count",
    },
    {
      empId: "EMP009",
      empName: "Imran Nazir",
      dailyCollection: 9500,
      cash: 4000,
      card: 4000,
      coupens: 1200,
      receipt: 9200,
      leftoverAmount: 300,
      difference: 300,
      netTotal: 9500,
      status: "Submitted",
    },
    {
      empId: "EMP010",
      empName: "Sobia Javed",
      dailyCollection: 16000,
      cash: 10000,
      card: 5000,
      coupens: 800,
      receipt: 15800,
      leftoverAmount: 200,
      difference: 200,
      netTotal: 16000,
      status: "Pending",
    },
    {
      empId: "EMP011",
      empName: "Waseem Akhtar",
      dailyCollection: 10500,
      cash: 7000,
      card: 3000,
      coupens: 400,
      receipt: 10200,
      leftoverAmount: 300,
      difference: 300,
      netTotal: 10500,
      status: "Approval Pending",
    },
    {
      empId: "EMP012",
      empName: "Mariam Yousuf",
      dailyCollection: 8700,
      cash: 5000,
      card: 3000,
      coupens: 600,
      receipt: 8500,
      leftoverAmount: 200,
      difference: 200,
      netTotal: 8700,
      status: "Re-Count",
    },
    {
      empId: "EMP013",
      empName: "Bilal Ahmed",
      dailyCollection: 11500,
      cash: 6000,
      card: 4000,
      coupens: 1000,
      receipt: 11000,
      leftoverAmount: 500,
      difference: 500,
      netTotal: 11500,
      status: "Submitted",
    },
    {
      empId: "EMP014",
      empName: "Hina Farooq",
      dailyCollection: 9800,
      cash: 6000,
      card: 3000,
      coupens: 500,
      receipt: 9500,
      leftoverAmount: 300,
      difference: 300,
      netTotal: 9800,
      status: "Pending",
    },
    {
      empId: "EMP015",
      empName: "Tariq Mehmood",
      dailyCollection: 12500,
      cash: 7000,
      card: 5000,
      coupens: 300,
      receipt: 12000,
      leftoverAmount: 500,
      difference: 500,
      netTotal: 12500,
      status: "Approval Pending",
    },
  ];
  public empReports: any[] = [
    {
      empId: "EMP001",
      empName: "Ali Raza",
      balance: 12000,
      cash: 7000,
      card: 3000,
      coupens: 1000,
      receipt: 11000,
      leftoverAmount: 500,
      difference: 1000,
      netTotal: 12000,
      status: "Submitted",
    },
    {
      empId: "EMP002",
      empName: "Sana Tariq",
      balance: 15000,
      cash: 8000,
      card: 5000,
      coupens: 1500,
      receipt: 14800,
      leftoverAmount: 200,
      difference: 200,
      netTotal: 15000,
      status: "Pending",
    },
    {
      empId: "EMP003",
      empName: "Usman Ghani",
      balance: 9000,
      cash: 6000,
      card: 2000,
      coupens: 500,
      receipt: 8700,
      leftoverAmount: 300,
      difference: 300,
      netTotal: 9000,
      status: "Approval Pending",
    },
    {
      empId: "EMP004",
      empName: "Fatima Noor",
      balance: 11000,
      cash: 6000,
      card: 4000,
      coupens: 500,
      receipt: 10500,
      leftoverAmount: 500,
      difference: 500,
      netTotal: 11000,
      status: "Re-Count",
    },
  ];
  public bankDeposit: any[] = [
    {
      empId: "EMP001",
      empName: "Ali Raza",
      submitted: 24000,
      balance: 50000,
    },
    {
      empId: "EMP002",
      empName: "Sana Tariq",
      submitted: 24000,
      balance: 50000,
    },
    {
      empId: "EMP003",
      empName: "Usman Ghani",
      submitted: 24000,
      balance: 50000,
    },
  ];
  public cashCounters: any[] = [
    {
      counterId: "01",
      assignedCashiers: [
        {
          name: "Ali Raza",
          avatar: "https://ui-avatars.com/api/?name=Ali+Raza&background=random",
        },
        {
          name: "Ahmed Khan",
          avatar:
            "https://ui-avatars.com/api/?name=Ahmed+Khan&background=random",
        },
      ],
      balance: 12000,
      cash: 7000,
      card: 3000,
      coupens: 1000,
      receipt: 11000,
      leftoverAmount: 500,
      difference: 1000,
      netTotal: 12000,
      status: "Submitted",
    },
    {
      counterId: "02",
      assignedCashiers: [
        {
          name: "Sana Tariq",
          avatar:
            "https://ui-avatars.com/api/?name=Sana+Tariq&background=random",
        },
        {
          name: "Maria Shah",
          avatar:
            "https://ui-avatars.com/api/?name=Maria+Shah&background=random",
        },
      ],
      balance: 15000,
      cash: 8000,
      card: 5000,
      coupens: 1500,
      receipt: 14800,
      leftoverAmount: 200,
      difference: 200,
      netTotal: 15000,
      status: "Pending",
    },
    {
      counterId: "03",
      assignedCashiers: [
        {
          name: "Usman Ghani",
          avatar:
            "https://ui-avatars.com/api/?name=Usman+Ghani&background=random",
        },
        {
          name: "Bilal Aslam",
          avatar:
            "https://ui-avatars.com/api/?name=Bilal+Aslam&background=random",
        },
      ],
      balance: 9000,
      cash: 6000,
      card: 2000,
      coupens: 500,
      receipt: 8700,
      leftoverAmount: 300,
      difference: 300,
      netTotal: 9000,
      status: "Approval Pending",
    },
    {
      counterId: "04",
      assignedCashiers: [
        {
          name: "Fatima Noor",
          avatar:
            "https://ui-avatars.com/api/?name=Fatima+Noor&background=random",
        },
        {
          name: "Hina Malik",
          avatar:
            "https://ui-avatars.com/api/?name=Hina+Malik&background=random",
        },
      ],
      balance: 11000,
      cash: 6000,
      card: 4000,
      coupens: 500,
      receipt: 10500,
      leftoverAmount: 500,
      difference: 500,
      netTotal: 11000,
      status: "Re-Count",
    },
  ];

  public tabs: string[] = [
    "Daily Cash",
    "Bank Deposits",
    "Employee Reports",
    "Cash Counters",
  ];

  public cashCollectionTabTranslationMap: any = {
    "Daily Cash": "CASH_COLLECTION.DAILY_CASH",
    "Bank Deposits": "CASH_COLLECTION.BANK_DEPOSITS",
    "Employee Reports": "CASH_COLLECTION.EMPLOYEE_REPORTS",
    "Cash Counters": "CASH_COLLECTION.CASH_COUNTERS",
  };

  public activeTab = "Daily Cash";
  constructor(private router: Router) {}

  setActiveTab(tab: any) {
    this.activeTab = tab;
  }

  goToDetails(cash: any) {
    this.router.navigate(["cash-collection/cash-detail", cash.empId], {
      state: { data: cash },
    });
  }
}
