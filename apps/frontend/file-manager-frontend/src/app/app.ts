import { Component, signal, ChangeDetectionStrategy } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { provideHttpClient, withFetch } from '@angular/common/http';
import { FileManagerApiService } from './file-manager-api.service';
import { FileExplorerComponent } from './file-explorer/file-explorer.component';
import { bootstrapApplication } from '@angular/platform-browser';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, CommonModule, FormsModule, FileExplorerComponent],
  templateUrl: './app.html',
  styleUrl: './app.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class App {
  protected readonly title = signal('file-manager-frontend');
  files = signal<string[]>([]);
  srcDir = signal('');
  dstDir = signal('');
  syncResult = signal<string[]>([]);

  constructor(private api: FileManagerApiService) {}

  listFiles() {
    this.api.listFiles(this.srcDir()).subscribe((files) => this.files.set(files));
  }

  syncFiles() {
    this.api
      .syncFiles(this.srcDir(), this.dstDir())
      .subscribe((result) => this.syncResult.set(result));
  }
}

// No bootstrapping here, it's handled in main.ts
