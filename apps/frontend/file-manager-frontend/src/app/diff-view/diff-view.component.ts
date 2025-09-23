import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SyncResult } from '../file-manager-api.service';

@Component({
  selector: 'app-diff-view',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './diff-view.component.html',
  styleUrls: ['./diff-view.component.scss']
})
export class DiffViewComponent {
  @Input() result!: SyncResult;
}
