import { Component,EventEmitter,Output } from '@angular/core';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'assign-shift-modal',
  imports: [TranslateModule],
  templateUrl: './assign-shift-modal.component.html',
})
export class AssignShiftModalComponent {
 @Output() close = new EventEmitter<boolean>();

  confirm() {
    this.close.emit(true); 
  }

  cancel() {
    this.close.emit(false); 
  }
}
