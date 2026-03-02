import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { environment } from '../../../environments/environment';

export interface Organization {
  id?: number;
  name: string;
  code: string;
  legal_name: string | null;
  tax_id: string | null;
  currency_code: string;
  fiscal_year_variant: string | null;
  is_active: boolean;
  metadata?: any;
  created_at?: string;
  updated_at?: string;
}

export interface OrganizationListParams {
  limit?: number;
  offset?: number;
  is_active?: boolean;
}

// Matches the envelope your API returns
export interface ApiResponse<T> {
  statusCode: number;
  message: string;
  data: T;
}

@Injectable({
  providedIn: 'root'
})
export class OrganizationService {
  private apiUrl = `${environment.baseUrl}/api/organizations`;

  constructor(private http: HttpClient) {}

  private getHeaders(): HttpHeaders {
    const tenantId = localStorage.getItem('tenantId') || '';
    const token = localStorage.getItem('authToken') || '';

    return new HttpHeaders({
      'Content-Type': 'application/json',
      'x-tenant-id': tenantId,
      'Authorization': `Bearer ${token}`
    });
  }

  /**
   * Get all organizations
   * GET /api/organizations
   * FIX: unwrap the { statusCode, message, data } envelope via .pipe(map(...))
   */
  getOrganizations(params?: OrganizationListParams): Observable<Organization[]> {
    let httpParams = new HttpParams();

    if (params) {
      if (params.limit !== undefined) {
        httpParams = httpParams.set('limit', params.limit.toString());
      }
      if (params.offset !== undefined) {
        httpParams = httpParams.set('offset', params.offset.toString());
      }
      if (params.is_active !== undefined) {
        httpParams = httpParams.set('is_active', params.is_active.toString());
      }
    }

    return this.http.get<ApiResponse<Organization[]>>(this.apiUrl, {
      headers: this.getHeaders(),
      params: httpParams
    }).pipe(
      map(response => response.data)  // <-- extract the actual array
    );
  }

  /**
   * Get a single organization by ID
   * GET /api/organizations/:id
   */
  getOrganization(id: number): Observable<Organization> {
    return this.http.get<ApiResponse<Organization>>(`${this.apiUrl}/${id}`, {
      headers: this.getHeaders()
    }).pipe(
      map(response => response.data)
    );
  }

  /**
   * Create a new organization
   * POST /api/organizations
   */
  createOrganization(organization: Organization): Observable<Organization> {
    return this.http.post<ApiResponse<Organization>>(this.apiUrl, organization, {
      headers: this.getHeaders()
    }).pipe(
      map(response => response.data)
    );
  }

  /**
   * Update an existing organization
   * PATCH /api/organizations/:id
   */
updateOrganization(id: number, organization: Organization): Observable<Organization> {
  return this.http.put<ApiResponse<Organization>>(`${this.apiUrl}/${id}`, organization, {
    headers: this.getHeaders()
  }).pipe(
    map(response => response.data)
  );
}

  /**
   * Delete an organization
   * DELETE /api/organizations/:id
   */
  deleteOrganization(id: number): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`, {
      headers: this.getHeaders()
    });
  }

  /**
   * Bulk delete organizations
   * DELETE /api/organizations
   */
  deleteOrganizations(ids: number[]): Observable<void> {
    return this.http.request<void>('delete', this.apiUrl, {
      headers: this.getHeaders(),
      body: { ids }
    });
  }

  /**
   * Export organizations
   * GET /api/organizations/export
   */
  exportOrganizations(format: 'csv' | 'xlsx' = 'xlsx'): Observable<Blob> {
    return this.http.get(`${this.apiUrl}/export`, {
      headers: this.getHeaders(),
      params: { format },
      responseType: 'blob'
    });
  }
}