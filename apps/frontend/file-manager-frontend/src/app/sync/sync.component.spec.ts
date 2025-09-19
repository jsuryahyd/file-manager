import { ComponentFixture, TestBed } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { SyncComponent } from './sync.component';
import { FileManagerApiService } from '../file-manager-api.service';
import { of, throwError } from 'rxjs';

describe('SyncComponent', () => {
  let component: SyncComponent;
  let fixture: ComponentFixture<SyncComponent>;
  let apiService: FileManagerApiService;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SyncComponent, HttpClientTestingModule],
      providers: [FileManagerApiService]
    })
    .compileComponents();

    fixture = TestBed.createComponent(SyncComponent);
    component = fixture.componentInstance;
    apiService = TestBed.inject(FileManagerApiService);
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should have an invalid form when source and destination are empty', () => {
    expect(component.form.valid).toBeFalsy();
  });

  it('should open the modal', () => {
    component.openModal('source');
    expect(component.isModalOpen).toBeTrue();
    expect(component.activeInput).toBe('source');
  });

  it('should close the modal', () => {
    component.closeModal();
    expect(component.isModalOpen).toBeFalsy();
    expect(component.activeInput).toBeNull();
  });

  it('should update the form on folder selection', () => {
    component.openModal('source');
    component.onFolderSelected('/test/path');
    expect(component.form.get('source')?.value).toBe('/test/path');
    expect(component.isModalOpen).toBeFalsy();
  });

  it('should call syncFiles on sync', () => {
    spyOn(apiService, 'syncFiles').and.returnValue(of(undefined));
    component.form.setValue({ source: '/src', destination: '/dst' });
    component.sync();
    expect(apiService.syncFiles).toHaveBeenCalledWith({ source: '/src', destination: '/dst' });
  });

  it('should handle 409 conflict on sync', () => {
    spyOn(apiService, 'syncFiles').and.callFake((req, force) => {
      if (force) {
        return of(undefined);
      }
      return throwError(() => ({ status: 409 }));
    });
    spyOn(window, 'confirm').and.returnValue(true);

    component.form.setValue({ source: '/src', destination: '/dst' });
    component.sync();

    expect(apiService.syncFiles).toHaveBeenCalledTimes(2);
    expect(window.confirm).toHaveBeenCalled();
  });
});
