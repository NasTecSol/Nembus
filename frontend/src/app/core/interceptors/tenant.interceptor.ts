import { Injectable } from '@angular/core';
import {
  HttpEvent,
  HttpHandler,
  HttpInterceptor,
  HttpRequest
} from '@angular/common/http';
import { Observable } from 'rxjs';
@Injectable()
export class TenantInterceptor implements HttpInterceptor {

  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    const tenantId = localStorage.getItem('x-tenant-id') || '';

    if (req.url.includes('/api/auth/login')) {
      const loginRequest = req.clone({
        setHeaders: {
          'x-tenant-id': tenantId,
          'Content-Type': 'application/json',
        }
      });
      return next.handle(loginRequest);
    }
    const token = localStorage.getItem('token');
    const authRequest = req.clone({
      setHeaders: {
        'x-tenant-id': tenantId,
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {})
      }
    });

    return next.handle(authRequest);
  }
}
