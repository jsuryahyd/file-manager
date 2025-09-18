import { Component, inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { FileManagerApiService } from '../file-manager-api.service';

@Component({
  selector: 'app-sync',
  standalone: true,
  imports: [ReactiveFormsModule],
  templateUrl: './sync.component.html',
  styleUrls: ['./sync.component.scss']
})
export class SyncComponent {
  private readonly fb = inject(FormBuilder);
  private readonly apiService = inject(FileManagerApiService);

  form = this.fb.group({
    source: ['', Validators.required],
    destination: ['', Validators.required]
  });

  sync() {
    if (this.form.valid) {
      const { source, destination } = this.form.value;
      this.apiService.syncFiles({ source: source!, destination: destination! })
        .subscribe(() => {
          console.log('Sync completed successfully');
          // Add any post-sync logic here, like a success message
        });
    }
  }
}
