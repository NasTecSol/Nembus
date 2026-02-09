import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-bills',
  imports: [CommonModule,FormsModule,TranslateModule],
  templateUrl: './bills.component.html',
})
export class BillsComponent {
 public bills: any[] = [
    {
      billNo: "INV-1001",
      amount: 2500,
      user: "Ali Khan",
      paymentMethod: "Cash",
      dateTime: "2025-05-21 10:30 AM",
      description: "Payment for grocery ",
    },
    {
      billNo: "INV-1002",
      amount: 1800,
      user: "Sara Raza",
      paymentMethod: "Credit Card",
      dateTime: "2025-05-20 03:45 PM",
      description: "Monthly subscription",
    },
    {
      billNo: "INV-1003",
      amount: 3200,
      user: "Ahmed Malik",
      paymentMethod: "Bank Transfer",
      dateTime: "2025-05-19 11:00 AM",
      description: "Electronics purchase",
    },
    {
      billNo: "INV-1004",
      amount: 975,
      user: "Hina Iqbal",
      paymentMethod: "Cash",
      dateTime: "2025-05-18 09:15 AM",
      description: "Stationery order",
    },
    {
      billNo: "INV-1005",
      amount: 4500,
      user: "Usman Sheikh",
      paymentMethod: "Debit Card",
      dateTime: "2025-05-17 01:20 PM",
      description: "Furniture advance",
    },
    {
      billNo: "INV-1006",
      amount: 2100,
      user: "Nadia Farooq",
      paymentMethod: "UPI",
      dateTime: "2025-05-16 04:50 PM",
      description: "Online course",
    },
    {
      billNo: "INV-1007",
      amount: 1600,
      user: "Bilal Zahid",
      paymentMethod: "Credit Card",
      dateTime: "2025-05-15 10:10 AM",
      description: "Mobile repair",
    },
    {
      billNo: "INV-1008",
      amount: 3050,
      user: "Zara Ali",
      paymentMethod: "Cash",
      dateTime: "2025-05-14 02:00 PM",
      description: "Fashion items",
    },
    {
      billNo: "INV-1009",
      amount: 870,
      user: "Danish Hussain",
      paymentMethod: "Bank Transfer",
      dateTime: "2025-05-13 12:45 PM",
      description: "Book purchase",
    },
    {
      billNo: "INV-1010",
      amount: 2200,
      user: "Sana Rafiq",
      paymentMethod: "Debit Card",
      dateTime: "2025-05-12 05:30 PM",
      description: "Utility bill",
    },
  ];

  public items = [
    {
      name: "Dairy Milk",
      quantity: 3,
      price: 222,
      avatar: "https://i.pravatar.cc/300?img=1",
    },
    {
      name: "KitKat",
      quantity: 2,
      price: 180,
      avatar: "https://i.pravatar.cc/300?img=2",
    },
    {
      name: "Snickers",
      quantity: 1,
      price: 150,
      avatar: "https://i.pravatar.cc/300?img=3",
    },
  ];
   public selectedBill: any = this.bills[0];

    selectUser(bill: any) {
    this.selectedBill = bill;
  }
}
