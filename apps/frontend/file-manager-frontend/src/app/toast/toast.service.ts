import { Injectable, signal } from '@angular/core';

export interface Toast {
  message: string;
  type: 'success' | 'error';
}

@Injectable({
  providedIn: 'root',
})
export class ToastService {
  toasts = signal<Toast[]>([]);

  show(message: string, type: 'success' | 'error' = 'success') {
    const newToast = { message, type };
    this.toasts.update(toasts => [...toasts, newToast]);
    setTimeout(() => this.remove(newToast), 5000);
  }

  remove(toast: Toast) {
    this.toasts.update(toasts => toasts.filter(t => t !== toast));
  }
}
