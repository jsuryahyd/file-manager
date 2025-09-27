import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SyncJob } from '../file-manager-api.service';

@Component({
  selector: 'app-sync-jobs-table',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './sync-jobs-table.component.html',
  styleUrls: ['./sync-jobs-table.component.scss']
})
export class SyncJobsTableComponent {
  @Input() jobs: SyncJob[] = [];

  getErrorMessage(job: SyncJob): string | null {
    if (job.misc) {
      try {
        const miscData = JSON.parse(job.misc);
        return miscData.error || null;
      } catch (e) {
        return null;
      }
    }
    return null;
  }
}
