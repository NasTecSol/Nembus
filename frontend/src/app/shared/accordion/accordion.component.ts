import { CommonModule } from "@angular/common";
import { Component, Input } from "@angular/core";
import { FormsModule } from "@angular/forms";

@Component({
  selector: "app-accordion",
  imports: [CommonModule, FormsModule],
  templateUrl: "./accordion.component.html",
})
export class AccordionComponent {
  @Input() items: any[] = [];
  @Input() showSwitch: boolean = false;
  @Input() itemListClass: string = "text-black font-semibold";
  toggle(index: number) {
    if (this.showSwitch) {
      this.items[index].open = !this.items[index].open;
    } else {
      this.items[index].open = !this.items[index].open;
      if (index !== 0) {
        this.items[0].open = false;
      }
    }
  }
}
