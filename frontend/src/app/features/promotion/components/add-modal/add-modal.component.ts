import { CommonModule } from '@angular/common';
import { Component ,Input,Output,EventEmitter} from '@angular/core';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';
import { CheckIconComponent } from '../../../../shared/icons/check.component';

@Component({
  selector: 'app-add-modal',
  imports: [CommonModule,FormsModule,TranslateModule,CheckIconComponent],
  templateUrl: './add-modal.component.html',

})
export class AddModalComponent {
  @Input() isOpen = false;
  @Input() product: any = null;

  @Output() close = new EventEmitter<void>();
  @Output() submit = new EventEmitter<any>();

  formData = {
    discount: ''
  };

  handleSubmit() {
    this.submit.emit({ product: this.product, discount: this.formData.discount });
    this.handleClose();
  }

  handleClose() {
    this.close.emit();
  }
}
