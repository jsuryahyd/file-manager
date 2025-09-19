import { Component, inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { FileManagerApiService } from '../file-manager-api.service';
import { FileExplorerModalComponent } from '../file-explorer-modal/file-explorer-modal.component';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-sync',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, FileExplorerModalComponent],
  templateUrl: './sync.component.html',
  styleUrls: ['./sync.component.scss']
})
export class SyncComponent {
  private readonly fb = inject(FormBuilder);
  private readonly apiService = inject(FileManagerApiService);

  isModalOpen = false;
  activeInput: 'source' | 'destination' | null = null;

  form = this.fb.group({
    source: ['', Validators.required],
    destination: ['', Validators.required]
  });

  openModal(inputType: 'source' | 'destination') {
    this.activeInput = inputType;
    this.isModalOpen = true;
  }

  closeModal() {
    this.isModalOpen = false;
    this.activeInput = null;
  }

  onFolderSelected(path: string) {
    if (this.activeInput) {
      this.form.get(this.activeInput)?.setValue(path);
    }
    this.closeModal();
  }

  sync() {
    if (this.form.valid) {
      const { source, destination } = this.form.value;
      this.apiService.syncFiles({ source: source!, destination: destination! })
        .subscribe({
          next: () => {
            console.log('Sync completed successfully');
            // Add any post-sync logic here, like a success message
          },
          error: (err) => {
            if (err.status === 409) {
              if (confirm('This is a new sync pair. Do you want to create it?')) {
                this.apiService.syncFiles({ source: source!, destination: destination! }, true)
                  .subscribe(() => {
                    console.log('Sync completed successfully');
                  });
              }
            } else {
              console.error('Sync failed', err);
            }
          }
        });
    }
  }
}
