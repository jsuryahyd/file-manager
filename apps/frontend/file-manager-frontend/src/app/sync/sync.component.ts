import { Component, inject, OnDestroy, OnInit, signal } from '@angular/core';
import {
  FormBuilder,
  ReactiveFormsModule,
  Validators,
  AbstractControl,
  ValidationErrors,
} from '@angular/forms';
import { FileManagerApiService, SyncRequest, SyncJob } from '../file-manager-api.service';
import { FileExplorerModalComponent } from '../file-explorer-modal/file-explorer-modal.component';
import { CommonModule } from '@angular/common';
import { DiffViewComponent } from '../diff-view/diff-view.component';
import { SyncJobsTableComponent } from '../sync-jobs-table/sync-jobs-table.component';
import { Subject, timer } from 'rxjs';
import { takeUntil, switchMap } from 'rxjs/operators';
import { ToastService } from '../toast/toast.service';

// Custom validator for absolute paths
function absolutePathValidator(control: AbstractControl): ValidationErrors | null {
  const path = control.value.replace(/\\/, '\\');
  if (!path) {
    return null; // Let Validators.required handle empty values
  }

  // Simple check for common absolute path patterns (Windows and Unix-like)
  // const isAbsolutePath = /^([a-zA-Z]:\\|\/)/.test(path);
  const isAbsolutePath =
    /^(?:[a-zA-Z]:[\\/]|(?:\/|\/\/)[^/\\]+|(?:[a-zA-Z]:)?(?:[\\/][^/\\]+)*[\\/])(?:[^/\\]+[\\/])*(?:[^/\\]+\.[^/\\]+)?/.test(
      path
    );

  return isAbsolutePath ? null : { absolutePath: true };
}

@Component({
  selector: 'app-sync',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, FileExplorerModalComponent, DiffViewComponent, SyncJobsTableComponent],
  templateUrl: './sync.component.html',
  styleUrls: ['./sync.component.scss'],
})
export class SyncComponent implements OnInit, OnDestroy {
  private readonly fb = inject(FormBuilder);
  private readonly apiService = inject(FileManagerApiService);
  private readonly toastService = inject(ToastService);
  private destroy$ = new Subject<void>();

  isModalOpen = false;
  activeInput: 'source' | 'destination' | null = null;
  jobs = signal<SyncJob[]>([]);
  isPolling = false;

  form = this.fb.group({
    source: ['', [Validators.required, absolutePathValidator]],
    destination: ['', [Validators.required, absolutePathValidator]],
    checkDuplicates: [true],
    overwriteExisting: [false],
    recursive: [true],
    skipPatterns: [''],
  });

  peekResult: any | null = null;

  ngOnInit(): void {
    this.startPolling();
  }

  ngOnDestroy(): void {
    this.stopPolling();
  }

  startPolling() {
    if (this.isPolling) return;

    this.isPolling = true;
    this.destroy$ = new Subject<void>(); // Re-create the subject

    timer(0, 5000)
      .pipe(
        switchMap(() => this.apiService.getSyncJobs()),
        takeUntil(this.destroy$)
      )
      .subscribe(newJobs => {
        const prevJobs = this.jobs();
        this.jobs.set(newJobs);

        const justCompletedJob = newJobs.find(job => {
            const prevJob = prevJobs.find(pJob => pJob.id === job.id);
            return prevJob && prevJob.status === 'running' && job.status === 'completed';
        });

        if (justCompletedJob) {
            this.toastService.show('Sync completed successfully.');
        }

        if (!newJobs.some(job => job.status === 'running')) {
          this.stopPolling();
        }
      });
  }

  stopPolling() {
    if (!this.isPolling) return;
    this.isPolling = false;
    this.destroy$.next();
    this.destroy$.complete();
  }

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

  preview() {
    if (this.form.valid) {
      const {
        source,
        destination,
        checkDuplicates,
        overwriteExisting,
        recursive,
        skipPatterns,
      } = this.form.value;
      const request: SyncRequest = {
        source: source!,
        destination: destination!,
        checkDuplicates: checkDuplicates!,
        overwriteExisting: overwriteExisting!,
        recursive: recursive!,
        skipPatterns: skipPatterns ? skipPatterns.split('\n') : [],
      };

      this.peekResult = null;

      this.apiService.peekSync(request).subscribe({
        next: (result) => {
          this.peekResult = result;
          console.log('Peek result:', this.peekResult);
        },
        error: (err) => {
          console.error('Peek failed', err);
        },
      });
    }
  }

  sync() {
    if (this.form.valid) {
      const {
        source,
        destination,
        checkDuplicates,
        overwriteExisting,
        recursive,
        skipPatterns,
      } = this.form.value;
      const request: SyncRequest = {
        source: source!,
        destination: destination!,
        checkDuplicates: checkDuplicates!,
        overwriteExisting: overwriteExisting!,
        recursive: recursive!,
        skipPatterns: skipPatterns ? skipPatterns.split('\n') : [],
      };

      this.peekResult = null;

      this.apiService.syncFiles(request).subscribe({
        next: () => {
          this.toastService.show('Sync job started and is now in progress.');
          this.startPolling();
        },
        error: (err) => {
          if (err.status === 409) {
            if (confirm('This is a new sync pair. Do you want to create it?')) {
              this.apiService.syncFiles(request, true).subscribe(() => {
                this.toastService.show('Sync job started and is now in progress.');
                this.startPolling();
              });
            }
          } else {
            console.error('Sync failed', err);
          }
        },
      });
    }
  }
}
