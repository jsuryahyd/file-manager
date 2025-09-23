import {
  Component,
  EventEmitter,
  inject,
  OnInit,
  Output,
  signal,
  ChangeDetectorRef,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormControl, ReactiveFormsModule } from '@angular/forms';
import { FileManagerApiService, FileEntry } from '../file-manager-api.service';
import { Observable, combineLatest, of, BehaviorSubject } from 'rxjs';
import { startWith, switchMap, map, debounceTime, take, catchError } from 'rxjs';

@Component({
  selector: 'app-file-explorer-modal',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './file-explorer-modal.component.html',
  styleUrls: ['./file-explorer-modal.component.scss'],
})
export class FileExplorerModalComponent implements OnInit {
  private readonly apiService = inject(FileManagerApiService);

  @Output() folderSelected = new EventEmitter<string>();
  @Output() closeModal = new EventEmitter<void>();

  pathControl = new FormControl('', { nonNullable: true });
  filterControl = new FormControl('', { nonNullable: true });

  columns = signal<FileEntry[][]>([]);
  selectedPath = signal('');

  private path$ = new BehaviorSubject<string>('');

  ngOnInit(): void {
    this.apiService
      .getHomeDir()
      .pipe(take(1))
      .subscribe((homeDir) => {
        this.path$.next(homeDir);
      });

    this.pathControl.valueChanges.subscribe((p) => this.path$.next(p));

    const filterChanges$ = this.filterControl.valueChanges.pipe(startWith(''), debounceTime(300));

    combineLatest([this.path$, filterChanges$])
      .pipe(switchMap(([path, filter]) => this.getColumnsObservable(path, filter)))
      .subscribe((cols) => {
        this.columns.set(cols);
      });

    this.path$.subscribe((p) => {
      this.selectedPath.set(p);
      this.pathControl.setValue(p, { emitEvent: false });
    });
  }

  private getColumnsObservable(path: string, filter: string): Observable<FileEntry[][]> {
    if (!path) {
      return of([]);
    }
    const normalizedPath = path.replace(/\//g, '/');

    const columnPaths: string[] = [];
    if (normalizedPath === '/') {
      columnPaths.push('/');
    } else {
      const segments = normalizedPath.split('/').filter((p) => p);
      if (/^[a-zA-Z]:$/.test(segments[0])) {
        // Windows
        let current = (segments.shift() as string) + '/';
        columnPaths.push(current);
        for (const segment of segments) {
          current += segment;
          columnPaths.push(current);
          current += '/';
        }
      } else {
        // Unix
        columnPaths.push('/');
        let current = '/';
        for (const segment of segments) {
          current += segment;
          columnPaths.push(current);
          current += '/';
        }
      }
    }

    const columnObservables = columnPaths.map((p) =>
      this.apiService.listFiles(p, 'dir').pipe(
        catchError((err) => {
          console.error('Error listing files for path:', p, err);
          return of([]); // Return an empty array on error
        })
      )
    );

    return combineLatest(columnObservables).pipe(
      map((columns) => {
        const lowerCaseFilter = filter.toLowerCase();
        if (!lowerCaseFilter) {
          return columns;
        }

        return columns.map((files, i) => {
          if (i === columns.length - 1) {
            // Only filter the last column
            return files.filter((file) => file.name.toLowerCase().includes(lowerCaseFilter));
          }
          return files;
        });
      })
    );
  }

  selectFolder(path: string) {
    this.folderSelected.emit(path);
  }

  navigateTo(path: string) {
    this.path$.next(path);
  }

  trackByPath(index: number, file: FileEntry): string {
    return file.path;
  }
}
