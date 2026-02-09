import { CommonModule } from "@angular/common";
import { Component } from "@angular/core";
import { TranslateModule } from "@ngx-translate/core";

@Component({
  selector: "permissions",
  imports: [CommonModule, TranslateModule],
  templateUrl: "./permissions.component.html",
})
export class PermissionsComponent {
  public dropdownOpen: boolean = false;
  public dropdownOpenIndex: number | null = null;

  public tableData = [
    {
      name: "Transmission",
      module: "Transmission of banks ECR",
      authorities: [
        "Can give authorities to others",
        "Edit Authority",
        "Remove Access",
      ],
      selectedAuthority: "Can give authorities to others", // default selected
    },
    {
      name: "Verification",
      module: "Verification of user data",
      authorities: ["Can verify records", "Edit Permissions", "Remove Access"],
      selectedAuthority: "Can verify records",
    },
  ];

  toggleDropdown(index: number) {
    this.dropdownOpenIndex = this.dropdownOpenIndex === index ? null : index;
  }

  selectAuthority(rowIndex: number, authority: string) {
    this.tableData[rowIndex].selectedAuthority = authority;
    this.dropdownOpenIndex = null; // close dropdown after selection
  }
}
