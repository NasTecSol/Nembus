import { Component, OnInit } from "@angular/core";
import { countries } from "../../../../utils/country-codes";
import { CommonModule, Location } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TranslateModule } from "@ngx-translate/core";
import { ToastyService } from "../../../../core/services/toasty.service";
import { Router } from "@angular/router";
import { HttpClient, HttpHeaders } from "@angular/common/http";
import { environment } from "../../../../../environments/environment";
@Component({
  selector: "app-add-user",
  imports: [CommonModule, FormsModule, TranslateModule],
  templateUrl: "./add-user.component.html",
})
export class AddUserComponent implements OnInit {
  public countries: any[] = countries;
  public selectedDialCode: string = "+92";
  public phoneNumber: string = "";
  baseUrl = environment.baseUrl;

  public firstName: string = "";
  public lastName: string = "";
  public email: string = "";
  public password: string = "";
  public isActive: boolean = false;
  roles: any[] = [];
selectedRole!: number;
selectedStore: any = null;
stores: any[] = [];




  constructor(
    private location: Location,
    private router: Router,
    private toasty: ToastyService,
    private http: HttpClient
  ) { }
  ngOnInit() {
    this.fetchRoles();
    this.fetchStores();
  }

  fetchRoles() {
    this.http.get<any>(`${this.baseUrl}/api/roles`).subscribe({
      next: (res) => {
        if (res.statusCode === 200) {
          this.roles = res.data.map((role: any) => ({
            id: role.id,
            name: role.name
          }));
        }
      },
      error: (err) => {
        console.error('Error fetching roles', err);
      }
    });
  }
  fetchStores() {
     this.http.get<any>(`${this.baseUrl}/api/stores`).subscribe({
      next: (res) => {
        if (res.statusCode === 200) {
          this.stores = res.data.map((store: any) => ({
            id: store.id,
            name: store.name
          }));
        }
      },
      error: (err) => {
        console.error('Error fetching roles', err);
      }
    });
  }
submit() {
  const payload = {
    email: this.email,
    employee_code: "",
    first_name: this.firstName,
    is_active: this.isActive,
    last_name: this.lastName,
    password: this.password,
    username: `${this.firstName}${this.lastName}`.toLowerCase()
  };
  this.http.post<any>(`${this.baseUrl}/api/users`, payload).subscribe({
    next: (res) => {
      const userId = res?.data?.id; 
      if (!userId) {
        this.toasty.error("User created but ID not returned");
        return;
      }
      const rolePayload = {
        role_id: this.selectedRole, 
        store_id: this.selectedStore,
        metadata: {}
      };
      this.http
        .post(`${this.baseUrl}/api/users/addUserRoles/${userId}`, rolePayload)
        .subscribe({
          next: () => {
            this.toasty.success("User and role assigned successfully");
          },
          error: (err) => {
            console.error("ROLE API ERROR:", err);
            this.toasty.error("User created but role assignment failed");
          }
        });
    },
    error: (err) => {
      console.error("USER API ERROR:", err);
      this.toasty.error("Failed to create user");
    }
  });
}



  goBack() {
    this.location.back();
  }
}
