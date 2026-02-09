import { CommonModule } from "@angular/common";
import { Component } from "@angular/core";
import { TranslateModule } from "@ngx-translate/core";
import { DiscountOffersComponent } from "../discount-offers/discount-offers.component";
import { CouponsComponent } from "../coupons/coupons.component";
import { CouponTypesComponent } from "../coupon-types/coupon-types.component";
import { MembershipComponent } from "../membership/membership.component";
import { Router } from "@angular/router";

@Component({
  selector: "app-list",
  imports: [
    CommonModule,
    TranslateModule,
    DiscountOffersComponent,
    CouponsComponent,
    CouponTypesComponent,
    MembershipComponent,
    MembershipComponent,
  ],
  templateUrl: "./promotions-list.component.html",
})
export class PromotionsListComponent {
  public tabs: string[] = [
    "Promotions",
    "Discount Offers",
    "Coupons",
    "Membership Points",
  ];

  public promotionsTab: any = {
    "Promotions": "PROMOTIONS.PROMOTIONS",
    "Discount Offers": "PROMOTIONS.DISCOUNT_OFFERS",
    "Coupons": "PROMOTIONS.COUPONS",
    "Membership Points": "PROMOTIONS.MEMBERSHIP_POINTS"
};


  public activeTab: string = this.tabs[0];

  public couponTabs: string[] = ["Coupons", "Coupon Types"];

  public couponActiveTab: string = this.couponTabs[0];

  public promotions = [
    {
      no: 1,
      code: 1001,
      symbol: "1+1 offer",
      name: "Pepsi Carton",
      block: false,
      active: true,
      fromDate: "01-06-2025",
      toDate: "30-06-2025",
    },
    {
      no: 2,
      code: 1002,
      symbol: "1+1 offer",
      name: "Pepsi Carton",
      block: false,
      active: true,
      fromDate: "01-07-2025",
      toDate: "31-07-2025",
    },
    {
      no: 3,
      code: 1003,
      symbol: "1+1 offer",
      name: "Pepsi Carton",
      block: true,
      active: false,
      fromDate: "01-05-2025",
      toDate: "15-05-2025",
    },
  ];

  constructor(private router: Router) {}

  setActiveTab(tab: string) {
    this.activeTab = tab;
  }
  setCouponActiveTab(tab: string) {
    this.couponActiveTab = tab;
  }
  addPromotion() {
    this.router.navigate(["/promotions/add-promotion"]);
  }
  addDiscountOffer() {
    this.router.navigate(["/promotions/add-discount-offer"]);
  }
  addCoupon() {
    this.router.navigate(["/promotions/add-coupon"]);
  }
  addCouponType() {
    this.router.navigate(["/promotions/add-coupon-type"]);
  }
  addMembershipOffer() {
    this.router.navigate(["/promotions/add-membership-offer"]);
  }
}
