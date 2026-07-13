import { CommonModule } from "@angular/common";
import { HttpClient } from "@angular/common/http";
import { Component, OnInit } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { ActivatedRoute } from "@angular/router";
import { catchError, forkJoin, map, of } from "rxjs";
import { environment } from "../../../../../environments/environment";
import { TranslateModule } from "@ngx-translate/core";
import { TranslateService } from "@ngx-translate/core";

type StockStatus = "In Stock" | "Low Stock" | "Out of Stock";
type TrendDirection = "up" | "down";
type MarkerTone = "teal" | "violet";
type ViewMode =
  | "overview"
  | "adjustment"
  | "transfer-step-1"
  | "transfer-step-2"
  | "transfer-step-3";

interface StockLevelRow {
  rowId: string;
  sku: string;
  itemName: string;
  category: string;
  location: string;
  inStock: number;
  unitValue: number;
  status: StockStatus;
  storeId: number;
  storeName: string;
}

interface StoreApiItem {
  id: number;
  name: string;
  code?: string;
  is_active?: boolean;
}

interface StoreApiResponse {
  data?: StoreApiItem[];
}

interface StoreTab {
  id: number | "all";
  name: string;
  fullName: string;
}

interface PosStoreProduct {
  sku?: string;
  barcode?: string;
  name?: string;
  category_name?: string;
  effective_price?: number;
  is_in_stock?: boolean;
  is_low_stock?: boolean;
  location?: string;
  storage_location_name?: string;
  quantity_on_hand?: number;
  quantity_available?: number;
  stock?: number;
  current_stock?: number;
  package_n_price?: Array<{
    price?: number;
    product_name?: string;
    barcodes?: string[];
  }>;
  [key: string]: any;
}

interface PosStoreProductsResponse {
  data?: PosStoreProduct[];
}

interface QuickAction {
  id: "adjust" | "transfer" | "count";
  title: string;
  iconClass: string;
}

interface KpiCard {
  label: string;
  value: string;
  change: string;
  trend: TrendDirection;
  subtitle: string;
  iconClass: string;
  iconColorClass: string;
  iconBgClass: string;
}

interface MapMarker {
  label: string;
  top: string;
  left: string;
  tone: MarkerTone;
}

interface TransferItem {
  id: string;
  name: string;
  categoryPath: string;
  sku: string;
  availableStock: number;
  transferQty: number;
  unit: string;
}

interface AdjustmentLine {
  id: string;
  productName: string;
  variant: string;
  sku: string;
  currentStock: number;
  newCount: number;
  reasonCode: string;
  unitCost: number;
}

@Component({
  selector: "app-stock-levels",
  imports: [CommonModule, FormsModule, TranslateModule],
  templateUrl: "./stock-levels.component.html",
  styleUrl: "./stock-levels.component.scss",
})
export class StockLevelsComponent implements OnInit {
  private readonly apiUrl = environment.baseUrl;
  public currentView: ViewMode = "overview";
  public isHeaderFilterOpen = false;
  public isLoadingStores = false;
  public isLoadingProducts = false;
  public overviewErrorMessage = "";

  public stores: StoreApiItem[] = [];
  public selectedStoreId: number | "all" = "all";
  public selectedCategory = "All Categories";
  public selectedLocation = "All Locations";
  public selectedStatus: StockStatus | "All Statuses" = "All Statuses";
  public stockRows: StockLevelRow[] = [];
  public currentPage = 1;
  public pageSize = 10;

  constructor(
    private readonly route: ActivatedRoute,
    private readonly http: HttpClient,
    private readonly translate: TranslateService
  ) { }

  ngOnInit(): void {
    this.route.queryParamMap.subscribe((params) => {
      const view = params.get("view");
      if (
        view === "adjustment" ||
        view === "transfer-step-1" ||
        view === "transfer-step-2" ||
        view === "transfer-step-3"
      ) {
        this.currentView = view;
      }
    });

    this.fetchStores();
  }

  public originStore = "Main Distribution Center - North";
  public originStorageLocation = "Warehouse A - Row 12 (Ambient)";
  public destinationStore = "";
  public destinationStorageLocation = "Main Floor - Section G";
  public transferInitiatedBy = "Alex Chen (Inventory Mgr)";
  public transferReferenceId = "TRF-2023-00942";
  public transferRemarks = "";
  public transferSearchTerm = "";
  public adjustmentTargetStore = "Central Distribution Center (CDC)";
  public adjustmentStorageLocation = "All Locations";
  public adjustmentReferenceId = "IA-2023-11-04-001";
  public adjustmentInitiatedBy = "John Doe (Lead Warehouse Manager)";
  public adjustmentCountType: "Cycle Count" | "Manual Adjustment" =
    "Manual Adjustment";
  public adjustmentRemarks = "";

  public readonly quickActions: QuickAction[] = [
    { id: "adjust", title: "STOCK_LEVELS.QUICK_ACTIONS.ADJUST", iconClass: "fa fa-pencil" },
    { id: "transfer", title: "STOCK_LEVELS.QUICK_ACTIONS.TRANSFER", iconClass: "fa fa-random" },
    { id: "count", title: "STOCK_LEVELS.QUICK_ACTIONS.COUNT", iconClass: "fa fa-check-square-o" },
  ];

  public readonly pageSizeOptions: number[] = [10, 20, 50];

  public readonly statusOptions: Array<StockStatus | "All Statuses"> = [
    "All Statuses",
    "In Stock",
    "Low Stock",
    "Out of Stock",
  ];

  public readonly originStoreOptions: string[] = [
    "Main Distribution Center - North",
    "Central Warehouse (WH-01)",
    "Main Warehouse - South",
  ];

  public readonly originLocationOptions: string[] = [
    "Warehouse A - Row 12 (Ambient)",
    "Rack B4 - Electronics",
    "Bulk Zone - Row 05",
  ];

  public readonly destinationStoreOptions: string[] = [
    "Choose destination...",
    "Downtown Retail Outlet (STR-05)",
    "Regional Hub - East",
    "Coastal Center",
  ];

  public readonly destinationLocationOptions: string[] = [
    "Main Floor - Section G",
    "Back Store - Shelf C",
    "Receiving Dock - Bay 2",
  ];
  public readonly adjustmentReasonOptions: string[] = [
    "Cycle Count",
    "Damaged Stock",
    "Receiving Variance",
    "Manual Correction",
  ];
  public readonly adjustmentStoreOptions: string[] = [
    "Central Distribution Center (CDC)",
    "Regional Hub East",
    "Downtown Retail Outlet",
  ];
  public readonly adjustmentStorageOptions: string[] = [
    "All Locations",
    "Aisle A / Bin 01",
    "Aisle B / Bin 14",
    "Bulk Receiving Zone",
  ];

  public readonly mapMarkers: MapMarker[] = [
    { label: "Chicago Hub", top: "62%", left: "38%", tone: "violet" },
    { label: "Dallas DC", top: "66%", left: "50%", tone: "teal" },
    { label: "Atlanta Store", top: "58%", left: "62%", tone: "violet" },
  ];

  public transferItems: TransferItem[] = [
    {
      id: "item-1",
      name: "Industrial Ceiling Fan (Silver)",
      categoryPath: "Electronics - Home Appliances",
      sku: "FAN-IND-22-SLV",
      availableStock: 142,
      transferQty: 12,
      unit: "Units",
    },
    {
      id: "item-2",
      name: "LED Workshop Light Bar",
      categoryPath: "Lighting - Industrial",
      sku: "LIT-WS-900X",
      availableStock: 24,
      transferQty: 5,
      unit: "Units",
    },
    {
      id: "item-3",
      name: "Heavy Duty Storage Rack",
      categoryPath: "Furniture - Storage",
      sku: "STR-RCK-HD-B",
      availableStock: 8,
      transferQty: 2,
      unit: "Units",
    },
  ];

  public adjustmentLines: AdjustmentLine[] = [
    {
      id: "adj-1",
      productName: "EliteBook 840 G8",
      variant: "Blue - 1TB - 8GB",
      sku: "840-G8",
      currentStock: 42,
      newCount: 40,
      reasonCode: "Cycle Count",
      unitCost: 218,
    },
    {
      id: "adj-2",
      productName: "Logitech MX Master 3S",
      variant: "Black",
      sku: "MXM3S-GR",
      currentStock: 124,
      newCount: 128,
      reasonCode: "Receiving Variance",
      unitCost: 109,
    },
    {
      id: "adj-3",
      productName: "Dell UltraSharp 27in 4K",
      variant: "Matte Panel",
      sku: "DELL-U27-4K",
      currentStock: 15,
      newCount: 14,
      reasonCode: "Manual Correction",
      unitCost: 249.5,
    },
  ];

  public get currentStep(): number {
    if (this.currentView === "transfer-step-1") {
      return 1;
    }
    if (this.currentView === "transfer-step-2") {
      return 2;
    }
    if (this.currentView === "transfer-step-3") {
      return 3;
    }
    return 0;
  }

  public get transferProgressPercent(): number {
    if (this.currentStep <= 1) {
      return 33;
    }
    if (this.currentStep === 2) {
      return 66;
    }
    return 100;
  }

  public get totalAdjustmentGains(): number {
    return this.adjustmentLines.reduce((sum, line) => {
      const delta = this.getAdjustmentDelta(line);
      if (delta <= 0) {
        return sum;
      }
      return sum + delta * line.unitCost;
    }, 0);
  }

  public get totalAdjustmentLosses(): number {
    return this.adjustmentLines.reduce((sum, line) => {
      const delta = this.getAdjustmentDelta(line);
      if (delta >= 0) {
        return sum;
      }
      return sum + Math.abs(delta * line.unitCost);
    }, 0);
  }

  public get netAdjustmentImpact(): number {
    return this.totalAdjustmentGains - this.totalAdjustmentLosses;
  }

  public get storeTabs(): StoreTab[] {
    return [
      { id: "all", name: "STOCK_LEVELS.TABS.ALL_STORES", fullName: "STOCK_LEVELS.TABS.ALL_STORES" },
      ...this.stores.map((store) => ({
        id: store.id,
        name: this.truncateText(store.name, 16),
        fullName: store.name,
      })),
    ];
  }

  public get categoryOptions(): string[] {
    const categories = Array.from(
      new Set(this.stockRows.map((row) => row.category))
    ).sort();
    return ["All Categories", ...categories];
  }

  public get locationOptions(): string[] {
    const locations = Array.from(
      new Set(this.stockRows.map((row) => row.location))
    ).sort();
    return ["All Locations", ...locations];
  }

  public get filteredRows(): StockLevelRow[] {
    return this.stockRows.filter((row) => {
      const matchesCategory =
        this.selectedCategory === "All Categories" ||
        row.category === this.selectedCategory;
      const matchesLocation =
        this.selectedLocation === "All Locations" ||
        row.location === this.selectedLocation;
      const matchesStatus =
        this.selectedStatus === "All Statuses" || row.status === this.selectedStatus;
      return matchesCategory && matchesLocation && matchesStatus;
    });
  }

  public get paginatedRows(): StockLevelRow[] {
    const start = (this.currentPage - 1) * this.pageSize;
    return this.filteredRows.slice(start, start + this.pageSize);
  }

  public get totalPages(): number {
    return Math.max(1, Math.ceil(this.filteredRows.length / this.pageSize));
  }

  public get showingFrom(): number {
    if (this.filteredRows.length === 0) {
      return 0;
    }
    return (this.currentPage - 1) * this.pageSize + 1;
  }

  public get showingTo(): number {
    return Math.min(this.currentPage * this.pageSize, this.filteredRows.length);
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

  public get filteredTransferItems(): TransferItem[] {
    const term = this.transferSearchTerm.trim().toLowerCase();
    if (!term) {
      return this.transferItems;
    }

    return this.transferItems.filter((item) => {
      return (
        item.name.toLowerCase().includes(term) ||
        item.sku.toLowerCase().includes(term) ||
        item.categoryPath.toLowerCase().includes(term)
      );
    });
  }

  public get totalTransferQty(): number {
    return this.transferItems.reduce(
      (sum, item) => sum + Math.max(0, Number(item.transferQty) || 0),
      0
    );
  }

  public get kpiCards(): KpiCard[] {
    const totalStockValue = this.stockRows.reduce(
      (sum, row) => sum + row.inStock * row.unitValue,
      0
    );
    const outOfStockItems = this.stockRows.filter(
      (row) => row.status === "Out of Stock"
    ).length;
    const lowStockItems = this.stockRows.filter(
      (row) => row.status === "Low Stock"
    ).length;

    return [
      {
        label: "STOCK_LEVELS.KPI.TOTAL_STOCK",
        value: this.formatCurrency(totalStockValue),
        change: "+2.5%",
        trend: "up",
        subtitle: "STOCK_LEVELS.KPI.TOTAL_STOCK_SUB",
        iconClass: "fa fa-money",
        iconColorClass: "text-[var(--color-selectedobj)]",
        iconBgClass: "bg-[var(--color-selectedobj)]/10",
      },
      {
        label: "STOCK_LEVELS.KPI.OUT_OF_STOCK",
        value: `${outOfStockItems}`,
        change: "-4%",
        trend: "down",
        subtitle: "STOCK_LEVELS.KPI.OUT_OF_STOCK_SUB",
        iconClass: "fa fa-exclamation-triangle",
        iconColorClass: "text-[#ef4444]",
        iconBgClass: "bg-[#fef2f2]",
      },
      {
        label: "STOCK_LEVELS.KPI.LOW_STOCK",
        value: `${lowStockItems}`,
        change: "-10%",
        trend: "down",
        subtitle: "STOCK_LEVELS.KPI.LOW_STOCK_SUB",
        iconClass: "fa fa-bell",
        iconColorClass: "text-[#f59e0b]",
        iconBgClass: "bg-[#fffbeb]",
      },
    ];
  }

  public setStoreTab(storeId: number | "all"): void {
    if (this.selectedStoreId === storeId) {
      return;
    }
    this.selectedStoreId = storeId;
    this.currentPage = 1;
    this.fetchProductsForSelectedStore();
  }

  public toggleHeaderFilterPanel(): void {
    this.isHeaderFilterOpen = !this.isHeaderFilterOpen;
  }

  public onFilterChange(): void {
    this.currentPage = 1;
  }

  public onPageSizeChange(): void {
    this.currentPage = 1;
  }

  public goToPreviousPage(): void {
    if (this.currentPage > 1) {
      this.currentPage -= 1;
    }
  }

  public goToNextPage(): void {
    if (this.currentPage < this.totalPages) {
      this.currentPage += 1;
    }
  }

  public setPage(page: number): void {
    if (page < 1 || page > this.totalPages || page === this.currentPage) {
      return;
    }
    this.currentPage = page;
  }

  public clearFilters(): void {
    this.selectedCategory = "All Categories";
    this.selectedLocation = "All Locations";
    this.selectedStatus = "All Statuses";
    this.currentPage = 1;
    this.isHeaderFilterOpen = false;
  }

  private fetchStores(): void {
    this.isLoadingStores = true;
    this.overviewErrorMessage = "";

    this.http.get<StoreApiResponse>(`${this.apiUrl}/api/stores`).subscribe({
      next: (response) => {
        const stores = Array.isArray(response?.data) ? response.data : [];
        this.stores = stores.filter(
          (store) => store?.id && store?.name && store.is_active !== false
        );
        this.isLoadingStores = false;
        if (this.stores.length === 0) {
          this.overviewErrorMessage = "No active stores found.";
          this.stockRows = [];
          return;
        }
        this.fetchProductsForSelectedStore();
      },
      error: (error) => {
        console.error("Failed to fetch stores:", error);
        this.stores = [];
        this.stockRows = [];
        this.isLoadingStores = false;
        this.overviewErrorMessage = "Unable to load stores right now.";
      },
    });
  }

  private fetchProductsForSelectedStore(): void {
    if (this.stores.length === 0) {
      this.stockRows = [];
      return;
    }

    this.isLoadingProducts = true;
    this.overviewErrorMessage = "";

    if (this.selectedStoreId === "all") {
      const requests = this.stores.map((store) =>
        this.http
          .get<PosStoreProductsResponse>(
            `${this.apiUrl}/api/pos/stores/${store.id}/products`
          )
          .pipe(
            map((response) => this.mapStoreProducts(response, store)),
            catchError((error) => {
              console.error(`Failed to fetch products for store ${store.id}:`, error);
              return of([] as StockLevelRow[]);
            })
          )
      );

      forkJoin(requests).subscribe({
        next: (rowsByStore) => {
          this.stockRows = rowsByStore.flat();
          this.currentPage = 1;
          this.isLoadingProducts = false;
        },
        error: () => {
          this.stockRows = [];
          this.isLoadingProducts = false;
          this.overviewErrorMessage = "Unable to load products for stores.";
        },
      });
      return;
    }

    const selectedStore = this.stores.find(
      (store) => store.id === this.selectedStoreId
    );
    if (!selectedStore) {
      this.stockRows = [];
      this.isLoadingProducts = false;
      return;
    }

    this.http
      .get<PosStoreProductsResponse>(
        `${this.apiUrl}/api/pos/stores/${selectedStore.id}/products`
      )
      .subscribe({
        next: (response) => {
          this.stockRows = this.mapStoreProducts(response, selectedStore);
          this.currentPage = 1;
          this.isLoadingProducts = false;
        },
        error: (error) => {
          console.error("Failed to fetch store products:", error);
          this.stockRows = [];
          this.isLoadingProducts = false;
          this.overviewErrorMessage = "Unable to load products for selected store.";
        },
      });
  }

  private mapStoreProducts(
    response: PosStoreProductsResponse,
    store: StoreApiItem
  ): StockLevelRow[] {
    const products = Array.isArray(response?.data) ? response.data : [];

    return products.map((product, index) => {
      const inStock = this.extractNumber([
        product.quantity_on_hand,
        product.quantity_available,
        product.stock,
        product.current_stock,
      ]);
      const unitValue = this.extractNumber([
        product.effective_price,
        product.package_n_price?.[0]?.price,
      ]);
      const nameCandidate =
        product.name ||
        product.package_n_price?.[0]?.product_name ||
        "Unnamed Product";
      const skuCandidate =
        product.sku ||
        product.barcode ||
        product.package_n_price?.[0]?.barcodes?.[0] ||
        `SKU-${store.id}-${index + 1}`;

      return {
        rowId: `${store.id}-${skuCandidate}-${index}`,
        sku: String(skuCandidate),
        itemName: String(nameCandidate),
        category: String(product.category_name || "Uncategorized"),
        location: String(
          product.storage_location_name || product.location || store.name
        ),
        inStock,
        unitValue,
        status: this.resolveStockStatus(product, inStock),
        storeId: store.id,
        storeName: store.name,
      };
    });
  }

  private resolveStockStatus(
    product: PosStoreProduct,
    inStock: number
  ): StockStatus {
    if (product.is_low_stock === true) {
      return "Low Stock";
    }

    if (product.is_in_stock === false || inStock <= 0) {
      return "Out of Stock";
    }

    if (inStock <= 10) {
      return "Low Stock";
    }

    return "In Stock";
  }

  public onQuickAction(actionId: QuickAction["id"]): void {
    if (actionId === "adjust") {
      this.startAdjustmentScreen();
      return;
    }

    if (actionId === "transfer") {
      this.startTransferWizard();
    }
  }

  public startAdjustmentScreen(): void {
    this.currentView = "adjustment";
  }

  public startTransferWizard(): void {
    this.currentView = "transfer-step-1";
  }

  public backToInventoryOverview(): void {
    this.currentView = "overview";
  }

  public goToTransferStep(step: 1 | 2 | 3): void {
    if (step === 1) {
      this.currentView = "transfer-step-1";
      return;
    }

    if (step === 2) {
      this.currentView = "transfer-step-2";
      return;
    }

    this.currentView = "transfer-step-3";
  }

  public addTransferItem(): void {
    this.transferItems.push({
      id: `item-${Date.now()}`,
      name: "New Transfer Item",
      categoryPath: "General - Misc",
      sku: `SKU-${this.transferItems.length + 100}`,
      availableStock: 16,
      transferQty: 1,
      unit: "Units",
    });
  }

  public removeTransferItem(id: string): void {
    const index = this.transferItems.findIndex((item) => item.id === id);
    if (index >= 0) {
      this.transferItems.splice(index, 1);
    }
  }

  public confirmTransfer(): void {
    this.backToInventoryOverview();
  }

  public applyReasonToAll(): void {
    const reason = this.adjustmentLines[0]?.reasonCode ?? this.adjustmentReasonOptions[0];
    this.adjustmentLines.forEach((line) => {
      line.reasonCode = reason;
    });
  }

  public zeroOutAllLines(): void {
    this.adjustmentLines.forEach((line) => {
      line.newCount = 0;
    });
  }

  public getAdjustmentDelta(line: AdjustmentLine): number {
    return (Number(line.newCount) || 0) - (Number(line.currentStock) || 0);
  }

  public getAdjustmentDeltaLabel(line: AdjustmentLine): string {
    const delta = this.getAdjustmentDelta(line);
    return delta >= 0 ? `+${delta}` : String(delta);
  }

  public getAdjustmentDeltaClass(line: AdjustmentLine): string {
    const delta = this.getAdjustmentDelta(line);
    if (delta > 0) {
      return "bg-[#ecfdf3] text-[#027a48]";
    }
    if (delta < 0) {
      return "bg-[#fef2f2] text-[#b42318]";
    }
    return "bg-[#f2f4f7] text-[#475467]";
  }

  public formatSignedCurrency(value: number): string {
    const abs = this.formatCurrency(Math.abs(value));
    return value >= 0 ? `+${abs}` : `-${abs}`;
  }

  public getCategoryClass(category: string): string {
    const normalized = category.toLowerCase();
    if (normalized === "electronics") {
      return "bg-[#eff8ff] text-[#175cd3]";
    }
    if (normalized === "machinery") {
      return "bg-[#f4f3ff] text-[#6938ef]";
    }
    if (normalized === "furniture") {
      return "bg-[#fff4ed] text-[#c4320a]";
    }
    return "bg-[#f2f4f7] text-[#475467]";
  }

  public getStatusClass(status: StockStatus): string {
    if (status === "Out of Stock") {
      return "text-[#d92d20]";
    }
    if (status === "Low Stock") {
      return "text-[#b54708]";
    }
    return "text-[#039855]";
  }

  public getStatusDotClass(status: StockStatus): string {
    if (status === "Out of Stock") {
      return "bg-[#f04438]";
    }
    if (status === "Low Stock") {
      return "bg-[#f59e0b]";
    }
    return "bg-[#12b76a]";
  }

  public getMarkerClass(tone: MarkerTone): string {
    return tone === "teal"
      ? "bg-[#14b8a6] shadow-[0_0_0_8px_rgba(20,184,166,0.18)]"
      : "bg-[#7f56d9] shadow-[0_0_0_8px_rgba(127,86,217,0.2)]";
  }

  public getTransferStockClass(value: number): string {
    if (value <= 25) {
      return "bg-[#fffaeb] text-[#b54708]";
    }
    return "bg-[#ecfdf3] text-[#027a48]";
  }

  public stepCircleClass(step: number): string {
    if (this.currentStep > step) {
      return "bg-[var(--color-selectedobj)] text-white";
    }
    if (this.currentStep === step) {
      return "bg-[var(--color-selectedobj)] text-white ring-4 ring-[var(--color-selectedobj)]/20";
    }
    return "bg-[#e4e7ec] text-[#667085]";
  }

  public stepLabelClass(step: number): string {
    return this.currentStep === step
      ? "text-[#101828]"
      : this.currentStep > step
        ? "text-[var(--color-selectedobj)]"
        : "text-[#667085]";
  }

  public trackBySku(_index: number, row: StockLevelRow): string {
    return row.rowId;
  }

  public trackByMarker(_index: number, marker: MapMarker): string {
    return marker.label;
  }

  public trackByTransferItem(_index: number, item: TransferItem): string {
    return item.id;
  }

  public trackByAdjustmentLine(_index: number, line: AdjustmentLine): string {
    return line.id;
  }

  private extractNumber(values: unknown[]): number {
    for (const value of values) {
      const parsed = Number(value);
      if (Number.isFinite(parsed)) {
        return parsed;
      }
    }
    return 0;
  }

  private truncateText(value: string, maxLength: number): string {
    const normalized = (value || "").trim();
    if (normalized.length <= maxLength) {
      return normalized;
    }
    return `${normalized.slice(0, maxLength - 1)}…`;
  }

  public formatCurrency(value: number): string {
    return value.toLocaleString(undefined, {
      style: "currency",
      currency: "USD",
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    });
  }
}