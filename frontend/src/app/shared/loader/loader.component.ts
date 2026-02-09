import { Component, OnInit } from "@angular/core";

import { Observable } from "rxjs";
import { LoaderService } from "../../core/services/loader.service";
import { CommonModule } from "@angular/common";

@Component({
  selector: "app-loader",
  imports:[CommonModule],
  templateUrl: "./loader.component.html",
  styleUrl: "./loader.component.css",
})
export class LoaderComponent implements OnInit {
  public loading!: Observable<boolean>;
  constructor(private loaderService: LoaderService) {}
  ngOnInit(): void {
    this.loading = this.loaderService.loading$; // Subscribe to loading state
  }
}
