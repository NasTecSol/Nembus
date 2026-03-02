import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, firstValueFrom } from 'rxjs';
import { map, catchError } from 'rxjs/operators';
import { environment } from '../../../environments/environment';

export interface Tenant {
  id: string;
  tenant_name: string;
  slug: string;
  db_conn_str: string;
  is_active: boolean;
  settings: any;
  created_at: Date;
  updated_at: Date;
}

export interface TenantApiResponse {
  statusCode: number;
  message: string;
  data: Tenant[];
}

@Injectable({
  providedIn: 'root'
})
export class TenantService {
  private baseUrl = environment.baseUrl;

  constructor(private http: HttpClient) {}

  // Get all active tenants
  async getActiveTenants(): Promise<Tenant[]> {
    try {
      const url = `${this.baseUrl}/api/tenants`;
      console.log('🔵 Fetching active tenants from:', url);

      const response = await firstValueFrom(
        this.http.get<any>(url).pipe(
          map((apiResponse) => {
            console.log('✅ Tenants API Response:', apiResponse);
            
            // Handle different response structures
            if (Array.isArray(apiResponse)) {
              return apiResponse;
            }
            if (apiResponse.data && Array.isArray(apiResponse.data)) {
              return apiResponse.data;
            }
            if (apiResponse.tenants && Array.isArray(apiResponse.tenants)) {
              return apiResponse.tenants;
            }
            
            console.error('❌ Unexpected response format:', apiResponse);
            return [];
          }),
          catchError((error) => {
            console.error('❌ Error fetching tenants:', error);
            throw error;
          })
        )
      );

      console.log(`✅ Fetched ${response.length} tenants`);
      return response;
    } catch (error) {
      console.error('❌ Error in getActiveTenants:', error);
      throw error;
    }
  }

  // Get all tenants (active and inactive)
  async getAllTenants(): Promise<Tenant[]> {
    try {
      const url = `${this.baseUrl}/api/tenants/all`;
      console.log('🔵 Fetching all tenants from:', url);

      const response = await firstValueFrom(
        this.http.get<any>(url).pipe(
          map((apiResponse) => {
            console.log('✅ All Tenants API Response:', apiResponse);
            
            // Handle different response structures
            if (Array.isArray(apiResponse)) {
              return apiResponse;
            }
            if (apiResponse.data && Array.isArray(apiResponse.data)) {
              return apiResponse.data;
            }
            if (apiResponse.tenants && Array.isArray(apiResponse.tenants)) {
              return apiResponse.tenants;
            }
            
            console.error('❌ Unexpected response format:', apiResponse);
            return [];
          }),
          catchError((error) => {
            console.error('❌ Error fetching all tenants:', error);
            throw error;
          })
        )
      );

      console.log(`✅ Fetched ${response.length} tenants`);
      return response;
    } catch (error) {
      console.error('❌ Error in getAllTenants:', error);
      throw error;
    }
  }

  // Update tenant (including activate/deactivate)
  async updateTenant(tenantId: string, fullTenantData: any): Promise<any> {
  try {
    const url = `${this.baseUrl}/api/tenants/${tenantId}`;
    console.log('🔵 Updating tenant:', url);
    console.log('🔵 Full tenant payload:', JSON.stringify(fullTenantData, null, 2));

    const response = await firstValueFrom(
      this.http.put<any>(url, fullTenantData).pipe(
        map((apiResponse) => {
          console.log('✅ Update Tenant Response:', apiResponse);
          return apiResponse;
        }),
        catchError((error) => {
          console.error('❌ Error updating tenant:', error);
          console.error('❌ Error response:', error.error);
          throw error;
        })
      )
    );

    return response;
  } catch (error) {
    console.error('❌ Error in updateTenant:', error);
    throw error;
  }
}
// Update the createTenant method in TenantService
async createTenant(tenantData: any): Promise<any> {
  try {
    const url = `${this.baseUrl}/api/tenants`;
    console.log('🔵 Creating tenant at:', url);
    console.log('🔵 Payload:', JSON.stringify(tenantData, null, 2));

    const response = await firstValueFrom(
      this.http.post<any>(url, tenantData).pipe(
        map((apiResponse) => {
          console.log('✅ Create Tenant Response:', apiResponse);
          return apiResponse;
        }),
        catchError((error) => {
          console.error('❌ Error creating tenant:', error);
          console.error('❌ Error details:', error.error);
          throw error;
        })
      )
    );

    return response;
  } catch (error) {
    console.error('❌ Error in createTenant:', error);
    throw error;
  }
}
  // Deactivate tenant (specific endpoint)
  async deactivateTenant(slug: string): Promise<any> {
    try {
      const url = `${this.baseUrl}/api/tenants/deactivate/${slug}`;
      console.log('🔵 Deactivating tenant:', url);

      const response = await firstValueFrom(
        this.http.put<any>(url, {}).pipe(
          map((apiResponse) => {
            console.log('✅ Deactivate Tenant Response:', apiResponse);
            return apiResponse;
          }),
          catchError((error) => {
            console.error('❌ Error deactivating tenant:', error);
            throw error;
          })
        )
      );

      return response;
    } catch (error) {
      console.error('❌ Error in deactivateTenant:', error);
      throw error;
    }
  }

  // Activate tenant (if there's a specific endpoint)
  async activateTenant(tenantId: string): Promise<any> {
    try {
      // If there's a specific activate endpoint, use it
      // Otherwise, use updateTenant with is_active: true
      return await this.updateTenant(tenantId, { is_active: true });
    } catch (error) {
      console.error('❌ Error in activateTenant:', error);
      throw error;
    }
  }
}