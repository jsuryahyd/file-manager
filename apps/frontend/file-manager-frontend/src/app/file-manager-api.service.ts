import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface FileEntry {
  name: string;
  isDir: boolean;
  size: number;
  modTime: string;
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
    return this.http.get<FileEntry[]>(`${this.apiUrl}/files?path=${encodeURIComponent(path)}`);
  }

  syncFiles(request: SyncRequest): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/sync`, request);
  }
}
