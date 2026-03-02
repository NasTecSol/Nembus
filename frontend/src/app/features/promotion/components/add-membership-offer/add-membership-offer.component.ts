import { Location } from '@angular/common';
import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-add-membership-offer',
  imports: [TranslateModule],
  templateUrl: './add-membership-offer.component.html',

})
export class AddMembershipOfferComponent {


  constructor(private location:Location,private router:Router){
  }

  goBack(){
    this.location.back()
  }
  submit(){
    this.router.navigate(['/promotions'])
  }
}
