import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';

export interface FileEntry {
  name: string;
  isDir: boolean;
  path: string;
  isPotentialDuplicate?: boolean;
}

export interface SyncRequest {
  source: string;
  destination: string;
  checkDuplicates: boolean;
  overwriteExisting: boolean;
  recursive: boolean;
  peekMode: boolean;
  skipPatterns: string[];
}

export interface SyncResult {
  FilesCopied: string[];
  FilesSkipped: string[];
  Errors: any[];
}

@Injectable({ providedIn: 'root' })
export class FileManagerApiService {
  private readonly apiUrl = 'http://localhost:8080/api';

  constructor(private http: HttpClient) {}

  listFiles(path: string, fileType?: 'file' | 'dir'): Observable<FileEntry[]> {
    let url = `${this.apiUrl}/files/list?path=${encodeURIComponent(path)}`;
    if (fileType) {
      url += `&type=${fileType}`;
    }
    return this.http.get<FileEntry[]>(url);
  }

  syncFiles(request: SyncRequest, force = false): Observable<SyncResult | void> {
    let url = `${this.apiUrl}/sync`;
    if (force) {
      url += '?force=true';
    }
    return this.http.post<SyncResult | void>(url, request);
  }

  findDuplicates(path: string): Observable<FileEntry[][]> {
    return this.http.get<FileEntry[][]>(`${this.apiUrl}/duplicates/find?path=${encodeURIComponent(path)}`);
  }

  deleteFiles(paths: string[]): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/duplicates/delete`, { paths });
  }

  getHomeDir(): Observable<string> {
    return this.listFiles('', 'dir').pipe(
      map(entries => {
        if (entries.length > 0) {
          const firstPath = entries[0].path.replace(/\\/g, '/');
          const lastSlash = firstPath.lastIndexOf('/');
          if (lastSlash > 0) {
            const parent = firstPath.substring(0, lastSlash);
            if (/^[a-zA-Z]:$/.test(parent)) {
              return parent + '/';
            }
            return parent;
          } else if (lastSlash === 0) {
            return '/';
          }
          return firstPath;
        }
        return '/'; // fallback
      })
    );
  }
}
