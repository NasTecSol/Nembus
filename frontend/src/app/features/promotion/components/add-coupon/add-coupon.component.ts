import { Location } from '@angular/common';
import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-add-coupon',
  imports: [TranslateModule],
  templateUrl: './add-coupon.component.html',
})
export class AddCouponComponent {

constructor(private location:Location,private router:Router){}
    goBack() {
    this.location.back();
  }
  submit(){
    this.router.navigate(['/promotions'])
  }
}
