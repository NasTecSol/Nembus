import { CommonModule } from "@angular/common";
import { HttpClient, HttpParams } from "@angular/common/http";
import { Component, OnDestroy, OnInit } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { Router } from "@angular/router";
import { environment } from "../../../../../environments/environment";
import { TranslateModule } from "@ngx-translate/core";

type BrandStatusFilter = "All Statuses" | "Active" | "Inactive";

interface BrandMetadata {
  country?: string;
  category?: string;
  [key: string]: unknown;
}

interface BrandApiItem {
  id: number;
  name?: string;
  code?: string;
  is_active?: boolean;
  metadata?: BrandMetadata;
  created_at?: string;
  updated_at?: string;
  [key: string]: unknown;
}

interface BrandsApiResponse {
  statusCode?: number;
  message?: string;
  data?: unknown;
  total?: number;
  total_count?: number;
  count?: number;
  has_next?: boolean;
  hasNext?: boolean;
  meta?: {
    total?: number;
    has_next?: boolean;
    hasNext?: boolean;
  };
  pagination?: {
    total?: number;
    has_next?: boolean;
    hasNext?: boolean;
  };
}

interface BrandRow {
  id: number;
  name: string;
  code: string;
  country: string;
  category: string;
  isActive: boolean;
  createdAt: string | null;
  updatedAt: string | null;
  logoText: string;
}

@Component({
  selector: "app-brand-list",
  imports: [CommonModule, FormsModule, TranslateModule],
  templateUrl: "./brand-list.component.html",
  styleUrl: "./brand-list.component.scss",
})
export class BrandListComponent implements OnInit, OnDestroy {
  private readonly apiUrl = environment.baseUrl;
  private searchDebounceTimer: ReturnType<typeof setTimeout> | null = null;
  private totalFromApi: number | null = null;
  private hasNextPage = false;

  public readonly pageSizeOptions: number[] = [10, 20, 50, 100];
  public readonly statusOptions: BrandStatusFilter[] = [
    "All Statuses",
    "Active",
    "Inactive",
  ];

  public searchTerm = "";
  public selectedStatus: BrandStatusFilter = "All Statuses";
  public currentPage = 1;
  public limit = 10;
  public isLoading = false;
  public errorMessage = "";

  public brands: BrandRow[] = [];
  public totalBrands = 0;

  constructor(
    private readonly http: HttpClient,
    private readonly router: Router
  ) {}

  public ngOnInit(): void {
    this.fetchBrands();
  }

  public ngOnDestroy(): void {
    if (this.searchDebounceTimer) {
      clearTimeout(this.searchDebounceTimer);
    }
  }

  public get offset(): number {
    return (this.currentPage - 1) * this.limit;
  }

  public get showingFrom(): number {
    return this.brands.length === 0 ? 0 : this.offset + 1;
  }

  public get showingTo(): number {
    return this.offset + this.brands.length;
  }

  public get totalPages(): number {
    if (this.totalFromApi !== null) {
      return Math.max(1, Math.ceil(this.totalFromApi / this.limit));
    }

    return this.hasNextPage ? this.currentPage + 1 : this.currentPage;
  }

  public get canGoNext(): boolean {
    if (this.totalFromApi !== null) {
      return this.currentPage < this.totalPages;
    }

    return this.hasNextPage;
  }

  public get pageTokens(): Array<number | "..."> {
    const total = this.totalPages;
    if (total <= 7) {
      return Array.from({ length: total }, (_, index) => index + 1);
    }

    const tokens: Array<number | "..."> = [1];
    const start = Math.max(2, this.currentPage - 1);
    const end = Math.min(total - 1, this.currentPage + 1);

    if (start > 2) {
      tokens.push("...");
    }

    for (let page = start; page <= end; page += 1) {
      tokens.push(page);
    }

    if (end < total - 1) {
      tokens.push("...");
    }

    tokens.push(total);
    return tokens;
  }

  public get brandHealthPercent(): number {
    if (this.brands.length === 0) {
      return 0;
    }

    const completeProfiles = this.brands.filter(
      (brand) =>
        brand.code !== "N/A" &&
        brand.country !== "N/A" &&
        brand.category !== "Uncategorized"
    ).length;

    return Math.round((completeProfiles / this.brands.length) * 100);
  }

  public get activeBrandsPercent(): number {
    if (this.brands.length === 0) {
      return 0;
    }

    const activeCount = this.brands.filter((brand) => brand.isActive).length;
    return Math.round((activeCount / this.brands.length) * 100);
  }

  public get recentUpdateSummary(): string {
    const latest = this.getLatestUpdatedBrand();
    if (!latest) {
      return "No updates found for the selected page.";
    }

    const date = this.toDate(latest.updatedAt ?? latest.createdAt);
    if (!date) {
      return `${latest.name} was updated recently.`;
    }

    return `${latest.name} was updated ${this.formatRelativeTime(date)}.`;
  }

  public onSearchChange(): void {
    this.currentPage = 1;
    this.queueFetch();
  }

  public onStatusChange(): void {
    this.currentPage = 1;
    this.fetchBrands();
  }

  public onPageSizeChange(): void {
    this.currentPage = 1;
    this.fetchBrands();
  }

  public clearFilters(): void {
    this.searchTerm = "";
    this.selectedStatus = "All Statuses";
    this.currentPage = 1;
    this.fetchBrands();
  }

  public goToPreviousPage(): void {
    if (this.currentPage > 1) {
      this.currentPage -= 1;
      this.fetchBrands();
    }
  }

  public goToNextPage(): void {
    if (this.canGoNext) {
      this.currentPage += 1;
      this.fetchBrands();
    }
  }

  public setPage(page: number): void {
    if (page === this.currentPage || page < 1) {
      return;
    }

    this.currentPage = page;
    this.fetchBrands();
  }

  public goToAddBrand(): void {
    void this.router.navigate(["/products/brands/new"]);
  }

  public trackByBrandId(_index: number, brand: BrandRow): number {
    return brand.id;
  }

  private queueFetch(): void {
    if (this.searchDebounceTimer) {
      clearTimeout(this.searchDebounceTimer);
    }

    this.searchDebounceTimer = setTimeout(() => {
      this.fetchBrands();
    }, 300);
  }

  private fetchBrands(): void {
    this.isLoading = true;
    this.errorMessage = "";

    let params = new HttpParams()
      .set("limit", String(this.limit))
      .set("offset", String(this.offset))
      .set("page", String(this.currentPage));

    const trimmedSearch = this.searchTerm.trim();
    if (trimmedSearch.length > 0) {
      params = params.set("search", trimmedSearch).set("q", trimmedSearch);
    }

    if (this.selectedStatus !== "All Statuses") {
      const isActive = this.selectedStatus === "Active";
      params = params
        .set("is_active", String(isActive))
        .set("status", this.selectedStatus.toLowerCase());
    }

    this.http.get<BrandsApiResponse>(`${this.apiUrl}/api/brands`, { params }).subscribe({
      next: (response) => {
        const items = this.extractItems(response);
        this.brands = items.map((item) => this.mapBrand(item));

        this.totalFromApi = this.extractTotal(response);
        this.hasNextPage = this.extractHasNext(response, this.brands.length);

        this.totalBrands =
          this.totalFromApi ??
          (this.offset + this.brands.length + (this.hasNextPage ? this.limit : 0));

        if (this.brands.length === 0 && this.currentPage > 1 && !this.hasNextPage) {
          this.currentPage -= 1;
          this.fetchBrands();
          return;
        }

        this.isLoading = false;
      },
      error: (error: unknown) => {
        console.error("Failed to fetch brands:", error);
        this.brands = [];
        this.totalBrands = 0;
        this.totalFromApi = null;
        this.hasNextPage = false;
        this.errorMessage = "Unable to load brands. Please try again.";
        this.isLoading = false;
      },
    });
  }

  private extractItems(response: BrandsApiResponse): BrandApiItem[] {
    if (Array.isArray(response.data)) {
      return response.data as BrandApiItem[];
    }

    const wrapper = this.getDataWrapper(response);
    if (!wrapper) {
      return [];
    }

    const itemCandidates = [wrapper["items"], wrapper["rows"], wrapper["data"]];
    for (const candidate of itemCandidates) {
      if (Array.isArray(candidate)) {
        return candidate as BrandApiItem[];
      }
    }

    return [];
  }

  private mapBrand(item: BrandApiItem): BrandRow {
    const id = this.parseNumber(item.id) ?? 0;
    const name = this.getString(item.name, "Unnamed Brand");
    const code = this.getString(item.code, id > 0 ? `BR-${id}` : "N/A");
    const metadata = this.isObject(item.metadata) ? (item.metadata as BrandMetadata) : {};

    const country = this.getString(metadata.country, "N/A");
    const category = this.getString(metadata.category, "Uncategorized");
    const isActive = Boolean(item.is_active);

    return {
      id,
      name,
      code,
      country,
      category,
      isActive,
      createdAt: this.toValidDateString(item.created_at),
      updatedAt: this.toValidDateString(item.updated_at),
      logoText: this.getLogoText(name, code),
    };
  }

  private extractTotal(response: BrandsApiResponse): number | null {
    const wrapper = this.getDataWrapper(response);
    const wrapperMeta = this.getNestedObject(wrapper, "meta");
    const wrapperPagination = this.getNestedObject(wrapper, "pagination");

    const candidates: unknown[] = [
      response.total,
      response.total_count,
      response.count,
      response.meta?.total,
      response.pagination?.total,
      wrapper?.["total"],
      wrapper?.["total_count"],
      wrapper?.["count"],
      wrapperMeta?.["total"],
      wrapperPagination?.["total"],
    ];

    for (const candidate of candidates) {
      const parsed = this.parseNumber(candidate);
      if (parsed !== null) {
        return parsed;
      }
    }

    return null;
  }

  private extractHasNext(response: BrandsApiResponse, currentLength: number): boolean {
    const wrapper = this.getDataWrapper(response);
    const wrapperMeta = this.getNestedObject(wrapper, "meta");
    const wrapperPagination = this.getNestedObject(wrapper, "pagination");

    const explicitCandidates: unknown[] = [
      response.has_next,
      response.hasNext,
      response.meta?.has_next,
      response.meta?.hasNext,
      response.pagination?.has_next,
      response.pagination?.hasNext,
      wrapper?.["has_next"],
      wrapper?.["hasNext"],
      wrapperMeta?.["has_next"],
      wrapperMeta?.["hasNext"],
      wrapperPagination?.["has_next"],
      wrapperPagination?.["hasNext"],
    ];

    for (const candidate of explicitCandidates) {
      if (typeof candidate === "boolean") {
        return candidate;
      }
    }

    if (this.totalFromApi !== null) {
      return this.offset + currentLength < this.totalFromApi;
    }

    return currentLength >= this.limit;
  }

  private getLatestUpdatedBrand(): BrandRow | null {
    if (this.brands.length === 0) {
      return null;
    }

    let latestBrand: BrandRow | null = null;
    let latestTimestamp = 0;

    for (const brand of this.brands) {
      const timestamp = this.toDate(brand.updatedAt ?? brand.createdAt)?.getTime() ?? 0;
      if (timestamp > latestTimestamp) {
        latestTimestamp = timestamp;
        latestBrand = brand;
      }
    }

    return latestBrand ?? this.brands[0];
  }

  private formatRelativeTime(date: Date): string {
    const diffMs = Date.now() - date.getTime();
    if (!Number.isFinite(diffMs) || diffMs < 0) {
      return `on ${date.toLocaleDateString()}`;
    }

    const minute = 60_000;
    const hour = 60 * minute;
    const day = 24 * hour;

    if (diffMs < minute) {
      return "just now";
    }

    if (diffMs < hour) {
      const minutes = Math.floor(diffMs / minute);
      return `${minutes} minute${minutes === 1 ? "" : "s"} ago`;
    }

    if (diffMs < day) {
      const hours = Math.floor(diffMs / hour);
      return `${hours} hour${hours === 1 ? "" : "s"} ago`;
    }

    const days = Math.floor(diffMs / day);
    if (days < 30) {
      return `${days} day${days === 1 ? "" : "s"} ago`;
    }

    return `on ${date.toLocaleDateString()}`;
  }

  private toDate(value: string | null): Date | null {
    if (typeof value !== "string" || value.trim().length === 0) {
      return null;
    }

    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? null : date;
  }

  private toValidDateString(value: unknown): string | null {
    if (typeof value !== "string" || value.trim().length === 0) {
      return null;
    }

    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return null;
    }

    return value;
  }

  private getLogoText(name: string, code: string): string {
    const normalized = name.replace(/[^a-zA-Z0-9]/g, "").trim();
    if (normalized.length > 0) {
      return normalized.charAt(0).toUpperCase();
    }

    return code.charAt(0).toUpperCase() || "B";
  }

  private getString(value: unknown, fallback: string): string {
    if (typeof value !== "string") {
      return fallback;
    }

    const trimmed = value.trim();
    return trimmed.length > 0 ? trimmed : fallback;
  }

  private parseNumber(value: unknown): number | null {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : null;
  }

  private getDataWrapper(response: BrandsApiResponse): Record<string, unknown> | null {
    if (this.isObject(response.data) && !Array.isArray(response.data)) {
      return response.data as Record<string, unknown>;
    }

    return null;
  }

  private getNestedObject(
    source: Record<string, unknown> | null,
    key: string
  ): Record<string, unknown> | null {
    if (!source) {
      return null;
    }

    const value = source[key];
    return this.isObject(value) ? (value as Record<string, unknown>) : null;
  }

  private isObject(value: unknown): value is object {
    return typeof value === "object" && value !== null;
  }
}
