import { Component } from '@angular/core';
import { countries } from '../../../../utils/country-codes';
import { CommonModule, Location } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';
@Component({
  selector: 'app-add-supplier',
  imports: [CommonModule,FormsModule,TranslateModule],
  templateUrl: './add-supplier.component.html',
})
export class AddSupplierComponent {
public countries: any[] = countries;
  public selectedDialCode: string = '+92';
  public phoneNumber: string = '';
constructor(private location:Location){}
  ngOnInit() {
    //console.log(this.countries);
  }
goBack(){
  this.location.back()
}
}
