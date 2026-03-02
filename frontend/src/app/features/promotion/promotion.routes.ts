import { Routes } from "@angular/router";
import { PromotionsListComponent } from "./components/promotions-list/promotions-list.component";
import { AddPromotionComponent } from "./components/add-promotion/add-promotion.component";
import { AddDiscountOfferComponent } from "./components/add-discount-offer/add-discount-offer.component";
import { AddCouponComponent } from "./components/add-coupon/add-coupon.component";
import { AddCouponTypeComponent } from "./components/add-coupon-type/add-coupon-type.component";
import { AddMembershipOfferComponent } from "./components/add-membership-offer/add-membership-offer.component";

export const PROMOTION_ROUTES: Routes = [
  {
    path: "",
    component: PromotionsListComponent,
  },
  {
    path: "add-promotion",
    component: AddPromotionComponent,
  },
  {
    path: "add-discount-offer",
    component: AddDiscountOfferComponent,
  },
  {
    path: "add-coupon",
    component: AddCouponComponent,
  },
  {
    path: "add-coupon-type",
    component: AddCouponTypeComponent,
  },
  {
    path: "add-membership-offer",
    component: AddMembershipOfferComponent,
  },
];
