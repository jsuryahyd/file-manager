import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface FileEntry {
  name: string;
  isDir: boolean;
  path: string;
}

export interface SyncRequest {
  source: string;
  destination: string;
}

@Injectable({ providedIn: 'root' })
export class FileManagerApiService {
  private readonly apiUrl = 'http://localhost:8080/api';

  constructor(private http: HttpClient) {}

  listFiles(path: string): Observable<FileEntry[]> {
    return this.http.get<FileEntry[]>(`${this.apiUrl}/files/list?path=${encodeURIComponent(path)}`);
  }

  syncFiles(request: SyncRequest, force = false): Observable<void> {
    let url = `${this.apiUrl}/sync`;
    if (force) {
      url += '?force=true';
    }
    return this.http.post<void>(url, request);
  }
}
