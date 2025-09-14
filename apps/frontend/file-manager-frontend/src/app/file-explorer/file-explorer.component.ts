import { Component } from '@angular/core';

@Component({
  selector: 'app-file-explorer',
  templateUrl: './file-explorer.component.html',
  styleUrls: ['./file-explorer.component.scss'],
  standalone: true,
})
export class FileExplorerComponent {
  files: Array<{ name: string; type: 'file' | 'folder' }> = [
    { name: 'Documents', type: 'folder' },
    { name: 'Photos', type: 'folder' },
    { name: 'notes.txt', type: 'file' },
    { name: 'todo.md', type: 'file' },
  ];
  selected: Set<string> = new Set();

  toggleSelect(name: string) {
    if (this.selected.has(name)) {
      this.selected.delete(name);
    } else {
      this.selected.add(name);
    }
  }
}
