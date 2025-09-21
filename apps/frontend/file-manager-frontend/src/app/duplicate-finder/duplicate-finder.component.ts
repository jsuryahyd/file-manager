import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FileManagerApiService, FileEntry } from '../file-manager-api.service';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-duplicate-finder',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './duplicate-finder.component.html',
  styleUrls: ['./duplicate-finder.component.scss']
})
export class DuplicateFinderComponent {
  private readonly apiService = inject(FileManagerApiService);

  scanPath = signal('');
  duplicates = signal<FileEntry[][]>([]);
  isLoading = signal(false);
  selection = signal<{ [path: string]: boolean }>({});

  findDuplicates() {
    if (!this.scanPath()) {
      return;
    }
    this.isLoading.set(true);
    this.duplicates.set([]);
    this.selection.set({});
    this.apiService.findDuplicates(this.scanPath())
      .subscribe({
        next: (result) => {
          this.duplicates.set(result || []);
          this.isLoading.set(false);
        },
        error: (err) => {
          console.error('Failed to find duplicates', err);
          this.isLoading.set(false);
        }
      });
  }

  deleteSelected() {
    const toDelete = Object.keys(this.selection()).filter(path => this.selection()[path]);
    if (toDelete.length === 0) {
      return;
    }

    if (!confirm(`Are you sure you want to delete ${toDelete.length} files?`)) {
      return;
    }

    this.apiService.deleteFiles(toDelete).subscribe({
      next: () => {
        console.log('Files deleted successfully');
        // Refresh the list of duplicates
        this.findDuplicates();
      },
      error: (err) => {
        console.error('Failed to delete files', err);
      }
    });
  }

  hasSelection(): boolean {
    return Object.values(this.selection()).some(v => v);
  }

  onSelectionChange(path: string, isChecked: boolean) {
    this.selection.update(s => ({...s, [path]: isChecked}));
  }
}
