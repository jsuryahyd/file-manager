import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({ providedIn: 'root' })
export class FileManagerApiService {
  constructor(private http: HttpClient) {}

  listFiles(dir: string = '.') : Observable<string[]> {
    return this.http.get<string[]>(`http://localhost:8080/api/list?dir=${encodeURIComponent(dir)}`);
  }

  syncFiles(src: string, dst: string): Observable<string[]> {
    return this.http.get<string[]>(`http://localhost:8080/api/sync?src=${encodeURIComponent(src)}&dst=${encodeURIComponent(dst)}`);
  }
}
