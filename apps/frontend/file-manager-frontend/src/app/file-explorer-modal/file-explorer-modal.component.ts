import { Component, EventEmitter, inject, Output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FileManagerApiService, FileEntry } from '../file-manager-api.service';
import { Observable } from 'rxjs';

@Component({
  selector: 'app-file-explorer-modal',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './file-explorer-modal.component.html',
  styleUrls: ['./file-explorer-modal.component.scss']
})
export class FileExplorerModalComponent {
  private readonly apiService = inject(FileManagerApiService);

  @Output() folderSelected = new EventEmitter<string>();
  @Output() closeModal = new EventEmitter<void>();

  currentPath = '';
  files$: Observable<FileEntry[]>;

  constructor() {
    this.files$ = this.apiService.listFiles(this.currentPath);
  }

  selectFolder(path: string) {
    this.folderSelected.emit(path);
  }

  selectAndStopPropagation(event: MouseEvent, path: string) {
    event.stopPropagation();
    this.selectFolder(path);
  }

  navigateTo(path: string) {
    this.currentPath = path;
    this.files$ = this.apiService.listFiles(this.currentPath);
  }

  goUp() {
    const parentPath = this.currentPath.substring(0, this.currentPath.lastIndexOf('/'));
    this.navigateTo(parentPath);
  }
}
