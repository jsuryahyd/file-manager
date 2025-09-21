import { Routes } from '@angular/router';
import { SyncComponent } from './sync/sync.component';
import { DuplicateFinderComponent } from './duplicate-finder/duplicate-finder.component';

export const routes: Routes = [
    { path: '', component: SyncComponent },
    { path: 'duplicates', component: DuplicateFinderComponent }
];