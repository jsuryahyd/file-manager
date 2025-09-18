import { Component, signal, ChangeDetectionStrategy } from '@angular/core';
import { RouterModule } from '@angular/router';
import { SyncComponent } from './sync/sync.component';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterModule, SyncComponent],
  templateUrl: './app.html',
  styleUrl: './app.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class App {
  protected readonly title = signal('File Manager');
}
