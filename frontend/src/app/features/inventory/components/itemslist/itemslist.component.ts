import { CommonModule } from "@angular/common";
import { HttpClient, HttpParams } from "@angular/common/http";
import { Component, OnInit } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { Router } from "@angular/router";
import { environment } from "../../../../../environments/environment";
import { TranslateModule } from "@ngx-translate/core";

type ProductStatus = "Active" | "Inactive";
type ProductType = "Physical" | "Digital";
type StockHealthStatus = "Optimal" | "Low Stock" | "Critical";
type AdjustmentAction = "+" | "-";
type StockUpdateAdjustmentType =
  | "Increment (+)"
  | "Decrement (-)"
  | "Set Exact Quantity";
type StockMovementType = "Stock In" | "Stock Out" | "Transfer" | "Adjustment";
type StockMovementStatus = "Completed" | "In Transit" | "Pending";

interface CatalogVariant {
  id: number;
  is_active: boolean;
  variant_sku: string;
  variant_name: string;
  variant_attributes?: Record<string, string>;
}

interface CatalogInventory {
  stock_id: number;
  store_id: number;
  reorder_level: number;
  max_stock_level: number;
  quantity_on_hand: number;
  reorder_quantity: number;
  quantity_on_order: number;
  product_variant_id: number;
  quantity_allocated: number;
  quantity_available: number;
  storage_location_id: number;
  storage_location_code: string;
  storage_location_name: string;
  quantity_in_transit?: number;
  metadata?: Record<string, any>;
  product_id?: number;
}

interface CatalogProductResponseItem {
  id: number;
  sku: string;
  name: string;
  description: string;
  is_active: boolean;
  category_id?: number;
  category_name: string;
  brand_name?: string;
  variants?: CatalogVariant[];
  inventory?: CatalogInventory[];
  [key: string]: any;
}

interface CatalogResponse {
  statusCode?: number;
  message?: string;
  data?: CatalogProductResponseItem[];
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

interface ProductRow {
  id: number;
  sku: string;
  name: string;
  category: string;
  categoryPath: string;
  categoryId: number | null;
  manufacturer: string;
  weight: string;
  type: ProductType;
  basePrice: string;
  basePriceValue: number | null;
  stock: number | null;
  status: ProductStatus;
  description: string;
  variants: CatalogVariant[];
  inventory: CatalogInventory[];
}

interface SummaryCard {
  label: string;
  value: string;
  iconClass: string;
  iconColorClass: string;
  iconBgClass: string;
}

interface EditKpiCard {
  title: string;
  value: string;
  subtitle: string;
  badgeText: string;
  badgeClass: string;
  iconClass: string;
  iconColorClass: string;
  iconBgClass: string;
}

interface LocationStockRow {
  stockId: number;
  store: string;
  exactLocation: string;
  onHand: number;
  allocated: number;
  available: number;
  reorderLevel: number;
  status: StockHealthStatus;
}

interface QuickAction {
  id: "transfer" | "barcode" | "supplier";
  title: string;
  subtitle: string;
  iconClass: string;
}

interface StockUpdateForm {
  locationStockId: string;
  adjustmentType: StockUpdateAdjustmentType;
  quantity: number;
  reasonCode: string;
  referenceNumber: string;
  notes: string;
}

interface StockMovementHistoryRow {
  id: string;
  dateTime: string;
  movementType: StockMovementType;
  reference: string;
  fromTo: string;
  location: string;
  quantity: number;
  performedBy: string;
  status: StockMovementStatus;
}

interface InventoryStockUpsertPayload {
  max_stock_level: string;
  metadata: Record<string, any>;
  product_id: number;
  product_variant_id: number | null;
  quantity_allocated: string;
  quantity_available: string;
  quantity_in_transit: string;
  quantity_on_hand: string;
  quantity_on_order: string;
  reorder_level: string;
  reorder_quantity: string;
  storage_location_id: number;
  store_id: number;
}

@Component({
  selector: "app-itemslist",
  standalone: true,
  imports: [CommonModule, FormsModule, TranslateModule],
  templateUrl: "./itemslist.component.html",
})
export class ItemslistComponent implements OnInit {
  private readonly apiUrl = environment.baseUrl;
  private readonly organizationId = 1;

  constructor(
    private readonly router: Router,
    private readonly http: HttpClient
  ) { }

  public readonly pageSizeOptions: number[] = [20, 50, 100];

  public readonly quickActions: QuickAction[] = [
    {
      id: "transfer",
      title: "ITEMS_LIST.QUICK_ACTIONS.TRANSFER",
      subtitle: "ITEMS_LIST.QUICK_ACTIONS.TRANSFER_SUB",
      iconClass: "fa fa-random",
    },
    {
      id: "barcode",
      title: "ITEMS_LIST.QUICK_ACTIONS.BARCODE",
      subtitle: "ITEMS_LIST.QUICK_ACTIONS.BARCODE_SUB",
      iconClass: "fa fa-barcode",
    },
    {
      id: "supplier",
      title: "ITEMS_LIST.QUICK_ACTIONS.SUPPLIER",
      subtitle: "ITEMS_LIST.QUICK_ACTIONS.SUPPLIER_SUB",
      iconClass: "fa fa-envelope",
    },
  ];

  public readonly movementTypes: string[] = [
    "Adjustment",
    "Transfer",
    "Cycle Count",
  ];
  public readonly stockUpdateAdjustmentTypes: StockUpdateAdjustmentType[] = [
    "Increment (+)",
    "Decrement (-)",
    "Set Exact Quantity",
  ];
  public readonly stockUpdateReasonCodes: string[] = [
    "Purchase Restock",
    "Sales Return",
    "Cycle Count Adjustment",
    "Manual Correction",
    "Damaged Goods",
  ];

  public readonly statusOptions: Array<ProductStatus | "All"> = [
    "All",
    "Active",
    "Inactive",
  ];

  public readonly typeOptions: Array<ProductType | "All"> = [
    "All",
    "Physical",
    "Digital",
  ];

  public searchTerm = "";
  public selectedCategory = "All";
  public selectedStatus: ProductStatus | "All" = "All";
  public selectedType: ProductType | "All" = "All";
  public currentPage = 1;
  public limit = 20;
  public isEditMode = false;
  public isLoading = false;
  public errorMessage = "";
  public isUpdateStockModalOpen = false;
  public isHistoryModalOpen = false;
  public isUpdatingStock = false;
  public stockUpdateErrorMessage = "";
  public selectedVariantId: number | null = null;
  public selectedStockId: number | null = null;

  public totalProducts = 0;
  public selectedProduct: ProductRow | null = null;
  public products: ProductRow[] = [];
  public categoryOptions: string[] = ["All"];

  private hasNextPage = false;
  private totalFromApi: number | null = null;
  private readonly categoryIdByName = new Map<string, number>();

  public adjustmentForm: {
    movementType: string;
    storageLocation: string;
    quantity: number;
    action: AdjustmentAction;
    notes: string;
  } = {
      movementType: "Adjustment",
      storageLocation: "",
      quantity: 0,
      action: "-",
      notes: "",
    };
  public stockUpdateForm: StockUpdateForm = this.createDefaultStockUpdateForm();
  public selectedHistoryMovementType: StockMovementType | "All Movements" =
    "All Movements";
  public selectedHistoryLocation = "All Locations";
  public historyDateRange = "Oct 01, 2023 - Oct 24, 2023";
  private historyRows: StockMovementHistoryRow[] = [];

  ngOnInit(): void {
    this.fetchProducts();
  }

  public get filteredProducts(): ProductRow[] {
    const query = this.searchTerm.trim().toLowerCase();

    return this.products.filter((product) => {
      const matchesSearch =
        query.length === 0 ||
        product.sku.toLowerCase().includes(query) ||
        product.name.toLowerCase().includes(query) ||
        product.category.toLowerCase().includes(query);

      const matchesCategory =
        this.selectedCategory === "All" ||
        product.category === this.selectedCategory;
      const matchesStatus =
        this.selectedStatus === "All" || product.status === this.selectedStatus;
      const matchesType =
        this.selectedType === "All" || product.type === this.selectedType;

      return matchesSearch && matchesCategory && matchesStatus && matchesType;
    });
  }

  public get summaryCards(): SummaryCard[] {
    const categoriesCount = new Set(this.products.map((p) => p.category)).size;
    const lowStockItems = this.products.filter((product) =>
      this.isLowStock(product)
    ).length;

    const inventoryValue = this.products.reduce((total, product) => {
      if (product.basePriceValue === null || product.stock === null) {
        return total;
      }
      return total + product.basePriceValue * product.stock;
    }, 0);

    return [
      {
        label: "ITEMS_LIST.SUMMARY.TOTAL_CATEGORIES",
        value: String(categoriesCount),
        iconClass: "fa fa-tags",
        iconColorClass: "text-[var(--color-selectedobj)]",
        iconBgClass: "bg-[var(--color-selectedobj)]/10",
      },
      {
        label: "ITEMS_LIST.SUMMARY.LOW_STOCK_ITEMS",
        value: String(lowStockItems),
        iconClass: "fa fa-exclamation-triangle",
        iconColorClass: "text-[#f59e0b]",
        iconBgClass: "bg-[#fef3c7]",
      },
      {
        label: "ITEMS_LIST.SUMMARY.INVENTORY_VALUE",
        value: inventoryValue > 0 ? this.formatCompactCurrency(inventoryValue) : "N/A",
        iconClass: "fa fa-money",
        iconColorClass: "text-[#f97316]",
        iconBgClass: "bg-[#ffedd5]",
      },
    ];
  }

  public get editKpis(): EditKpiCard[] {
    const rows = this.stockByLocationRows;
    const onHand = rows.reduce((sum, row) => sum + row.onHand, 0);
    const allocated = rows.reduce((sum, row) => sum + row.allocated, 0);
    const available = rows.reduce((sum, row) => sum + row.available, 0);
    const lowStockLocations = rows.filter((row) => row.status !== "Optimal").length;

    return [
      {
        title: "ITEMS_LIST.KPI.ON_HAND_TOTAL",
        value: onHand.toLocaleString(),
        subtitle: "ITEMS_LIST.KPI.ON_HAND_SUB",
        badgeText: "+0%",
        badgeClass: "text-[var(--color-selectedobj)] bg-[var(--color-selectedobj)]/10",
        iconClass: "fa fa-archive",
        iconColorClass: "text-[var(--color-selectedobj)]",
        iconBgClass: "bg-[var(--color-selectedobj)]/10",
      },
      {
        title: "ITEMS_LIST.KPI.ALLOCATED",
        value: allocated.toLocaleString(),
        subtitle: "ITEMS_LIST.KPI.ALLOCATED_SUB",
        badgeText: "ITEMS_LIST.KPI.BADGE_DEMAND",
        badgeClass: "text-[#f59e0b] bg-[#fffbeb]",
        iconClass: "fa fa-random",
        iconColorClass: "text-[#f59e0b]",
        iconBgClass: "bg-[#fef3c7]",
      },
      {
        title: "ITEMS_LIST.KPI.AVAILABLE",
        value: available.toLocaleString(),
        subtitle: "ITEMS_LIST.KPI.AVAILABLE_SUB",
        badgeText: "ITEMS_LIST.KPI.BADGE_READY",
        badgeClass: "text-[var(--color-selectedobj)] bg-[var(--color-selectedobj)]/10",
        iconClass: "fa fa-check-circle",
        iconColorClass: "text-[var(--color-selectedobj)]",
        iconBgClass: "bg-[var(--color-selectedobj)]/10",
      },
      {
        title: "ITEMS_LIST.KPI.LOW_STOCK_LOCATIONS",
        value: lowStockLocations.toString().padStart(2, "0"),
        subtitle: "ITEMS_LIST.KPI.LOW_STOCK_SUB",
        badgeText: lowStockLocations > 0 ? "ITEMS_LIST.KPI.BADGE_CRITICAL" : "ITEMS_LIST.KPI.BADGE_NORMAL",
        badgeClass:
          lowStockLocations > 0
            ? "text-[#ef4444] bg-[#fef2f2]"
            : "text-[var(--color-selectedobj)] bg-[var(--color-selectedobj)]/10",
        iconClass: "fa fa-exclamation-circle",
        iconColorClass: lowStockLocations > 0 ? "text-[#ef4444]" : "text-[var(--color-selectedobj)]",
        iconBgClass: lowStockLocations > 0 ? "bg-[#fef2f2]" : "bg-[var(--color-selectedobj)]/10",
      },
    ];
  }

  public get stockByLocationRows(): LocationStockRow[] {
    return this.getSelectedVariantInventoryRows().map((row) => {
      const onHand = this.toNumber(row.quantity_on_hand);
      const allocated = this.toNumber(row.quantity_allocated);
      const available = this.toNumber(row.quantity_available);
      const reorderLevel = this.toNumber(row.reorder_level);

      return {
        stockId: this.toNumber(row.stock_id),
        store: row.store_id ? `Store #${row.store_id}` : "Store",
        exactLocation:
          row.storage_location_name ||
          row.storage_location_code ||
          "N/A",
        onHand,
        allocated,
        available,
        reorderLevel,
        status: this.computeStockHealth(available, reorderLevel),
      };
    });
  }

  public get hasVariants(): boolean {
    return !!this.selectedProduct && this.selectedProduct.variants.length > 0;
  }

  public get selectedVariantName(): string {
    if (!this.selectedProduct || !this.hasVariants) {
      return "Default Variant";
    }

    return (
      this.selectedProduct.variants.find(
        (variant) => variant.id === this.selectedVariantId
      )?.variant_name ?? this.selectedProduct.variants[0]?.variant_name ?? "Variant"
    );
  }

  public get storageLocations(): string[] {
    const locations = this.stockByLocationRows.map((row) => row.store);
    const unique = Array.from(new Set(locations));
    return unique.length > 0 ? unique : ["Store #1"];
  }

  public get stockLocationOptions(): Array<{ value: string; label: string }> {
    const options = this.stockByLocationRows.map((row) => ({
      value: String(row.stockId),
      label: `${row.store} -> ${row.exactLocation}`,
    }));

    return options.length > 0
      ? options
      : [{ value: "default", label: "Default Location" }];
  }

  public get selectedStockLocationLabel(): string {
    if (!this.selectedStockId) {
      return "N/A";
    }

    return (
      this.stockLocationOptions.find(
        (option) => this.toNumber(option.value) === this.selectedStockId
      )?.label ?? "N/A"
    );
  }

  public get canSubmitStockUpdate(): boolean {
    return (
      !!this.getSelectedInventoryForUpdate() &&
      this.stockUpdateForm.quantity >= 0 &&
      !this.isUpdatingStock
    );
  }

  public get historyMovementTypeOptions(): Array<StockMovementType | "All Movements"> {
    const unique = new Set<StockMovementType>();
    this.historyRows.forEach((row) => unique.add(row.movementType));
    return ["All Movements", ...Array.from(unique)];
  }

  public get historyLocationOptions(): string[] {
    const unique = new Set<string>();
    this.historyRows.forEach((row) => unique.add(row.location));
    return ["All Locations", ...Array.from(unique)];
  }

  public get filteredHistoryRows(): StockMovementHistoryRow[] {
    return this.historyRows.filter((row) => {
      const matchesType =
        this.selectedHistoryMovementType === "All Movements" ||
        row.movementType === this.selectedHistoryMovementType;
      const matchesLocation =
        this.selectedHistoryLocation === "All Locations" ||
        row.location === this.selectedHistoryLocation;

      return matchesType && matchesLocation;
    });
  }

  public get offset(): number {
    return (this.currentPage - 1) * this.limit;
  }

  public get showingFrom(): number {
    return this.filteredProducts.length === 0 ? 0 : this.offset + 1;
  }

  public get showingTo(): number {
    return this.offset + this.filteredProducts.length;
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

  public clearFilters(): void {
    this.searchTerm = "";
    this.selectedCategory = "All";
    this.selectedStatus = "All";
    this.selectedType = "All";
    this.currentPage = 1;
    this.fetchProducts();
  }

  public onCategoryChange(): void {
    this.currentPage = 1;
    this.fetchProducts();
  }

  public onPageSizeChange(): void {
    this.currentPage = 1;
    this.fetchProducts();
  }

  public goToPreviousPage(): void {
    if (this.currentPage > 1) {
      this.currentPage -= 1;
      this.fetchProducts();
    }
  }

  public goToNextPage(): void {
    if (this.canGoNext) {
      this.currentPage += 1;
      this.fetchProducts();
    }
  }

  public setPage(page: number): void {
    if (page === this.currentPage || page < 1) {
      return;
    }

    this.currentPage = page;
    this.fetchProducts();
  }

  public goToAddProduct(): void {
    void this.router.navigate(["/inventory/add-product"]);
  }

  public openEditView(product: ProductRow): void {
    this.selectedProduct = product;
    this.selectedVariantId = product.variants[0]?.id ?? null;
    this.syncSelectedStockId();
    this.adjustmentForm.storageLocation = this.storageLocations[0] ?? "";
    this.stockUpdateForm = this.createDefaultStockUpdateForm();
    this.stockUpdateErrorMessage = "";
    this.isEditMode = true;
  }

  public backToList(): void {
    this.isUpdateStockModalOpen = false;
    this.isHistoryModalOpen = false;
    this.stockUpdateErrorMessage = "";
    this.selectedStockId = null;
    this.selectedVariantId = null;
    this.isEditMode = false;
  }

  public onVariantChange(): void {
    this.syncSelectedStockId();
    this.stockUpdateForm = this.createDefaultStockUpdateForm();
  }

  public selectStockRow(row: LocationStockRow): void {
    this.selectedStockId = row.stockId;
    this.stockUpdateForm.locationStockId = String(row.stockId);
  }

  public openUpdateStockModal(): void {
    this.syncSelectedStockId();
    this.stockUpdateForm = this.createDefaultStockUpdateForm();
    this.isHistoryModalOpen = false;
    this.stockUpdateErrorMessage = "";
    this.isUpdateStockModalOpen = true;
  }

  public closeUpdateStockModal(): void {
    this.stockUpdateErrorMessage = "";
    this.isUpdateStockModalOpen = false;
  }

  public openHistoryModal(): void {
    if (!this.selectedProduct) {
      return;
    }

    this.historyRows = this.buildInitialHistoryRows(this.selectedProduct);
    this.selectedHistoryMovementType = "All Movements";
    this.selectedHistoryLocation = "All Locations";
    this.isUpdateStockModalOpen = false;
    this.isHistoryModalOpen = true;
  }

  public closeHistoryModal(): void {
    this.isHistoryModalOpen = false;
  }

  public resetHistoryFilters(): void {
    this.selectedHistoryMovementType = "All Movements";
    this.selectedHistoryLocation = "All Locations";
  }

  public confirmStockUpdate(): void {
    if (!this.selectedProduct || this.isUpdatingStock) {
      return;
    }

    const targetInventory = this.getSelectedInventoryForUpdate();

    if (!targetInventory) {
      this.isUpdateStockModalOpen = false;
      return;
    }

    const quantityOnHand = Math.max(0, this.toNumber(this.stockUpdateForm.quantity));
    const previousOnHand = this.toNumber(targetInventory.quantity_on_hand);
    const payload = this.buildInventoryUpsertPayload(targetInventory, quantityOnHand);

    this.isUpdatingStock = true;
    this.stockUpdateErrorMessage = "";

    this.http
      .patch(`${this.apiUrl}/api/inventory-stock/upsert`, payload)
      .subscribe({
        next: () => {
          targetInventory.quantity_on_hand = quantityOnHand;
          this.selectedProduct!.stock = this.selectedProduct!.inventory.reduce(
            (sum, row) => sum + this.toNumber(row.quantity_on_hand),
            0
          );

          const quantityDelta = quantityOnHand - previousOnHand;
          this.historyRows = [
            {
              id: `manual-${Date.now()}`,
              dateTime: this.formatDateTime(new Date()),
              movementType: "Adjustment",
              reference:
                this.stockUpdateForm.referenceNumber.trim() || "MANUAL-ADJUSTMENT",
              fromTo: this.stockUpdateForm.reasonCode,
              location: this.selectedStockLocationLabel,
              quantity: quantityDelta,
              performedBy: "Current User",
              status: "Completed",
            },
            ...this.historyRows,
          ];

          this.isUpdatingStock = false;
          this.isUpdateStockModalOpen = false;
          this.stockUpdateForm = this.createDefaultStockUpdateForm();
          this.fetchProducts();
        },
        error: (error) => {
          console.error("Failed to upsert inventory stock:", error);
          this.stockUpdateErrorMessage = "Unable to update stock. Please try again.";
          this.isUpdatingStock = false;
        },
      });
  }

  public onQuickAction(actionId: QuickAction["id"]): void {
    if (actionId === "transfer") {
      void this.router.navigate(["/inventory/overview/levels"], {
        queryParams: { view: "transfer-step-1" },
      });
    }
  }

  public setAdjustmentAction(action: AdjustmentAction): void {
    this.adjustmentForm.action = action;
  }

  public submitAdjustment(): void {
    this.adjustmentForm.quantity = 0;
    this.adjustmentForm.notes = "";
  }

  public get selectedProductTitle(): string {
    if (!this.selectedProduct) {
      return "Product Details";
    }

    return `${this.selectedProduct.name} (${this.selectedProduct.sku})`;
  }

  public get selectedProductCategoryPath(): string {
    return this.selectedProduct?.categoryPath ?? "N/A";
  }

  public get selectedProductManufacturer(): string {
    return this.selectedProduct?.manufacturer ?? "N/A";
  }

  public get selectedProductWeight(): string {
    return this.selectedProduct?.weight ?? "N/A";
  }

  public get selectedProductTotalOnHand(): number {
    return this.stockByLocationRows.reduce((sum, row) => sum + row.onHand, 0);
  }

  public isLowStock(product: ProductRow): boolean {
    if (product.stock !== null && product.stock <= 20) {
      return true;
    }

    return product.inventory.some((row) => {
      const available = this.toNumber(row.quantity_available);
      const reorder = this.toNumber(row.reorder_level);
      return reorder > 0 && available <= reorder;
    });
  }

  public getLocationStatusClass(status: StockHealthStatus): string {
    if (status === "Critical") {
      return "bg-[#fef2f2] text-[#ef4444]";
    }
    if (status === "Low Stock") {
      return "bg-[#fffbeb] text-[#f59e0b]";
    }
    return "bg-[var(--color-selectedobj)]/10 text-[var(--color-selectedobj)]";
  }

  public getLocationStatusDotClass(status: StockHealthStatus): string {
    if (status === "Critical") {
      return "bg-[#ef4444]";
    }
    if (status === "Low Stock") {
      return "bg-[#f59e0b]";
    }
    return "bg-[var(--color-selectedobj)]";
  }

  public getHistoryTypeClass(type: StockMovementType): string {
    if (type === "Stock In") {
      return "bg-[#ecfdf3] text-[#027a48]";
    }
    if (type === "Stock Out") {
      return "bg-[#fef2f2] text-[#b42318]";
    }
    if (type === "Transfer") {
      return "bg-[#eff8ff] text-[#175cd3]";
    }
    return "bg-[#f5f8ff] text-[#444ce7]";
  }

  public getHistoryQuantityClass(quantity: number): string {
    return quantity >= 0 ? "text-[#039855]" : "text-[#d92d20]";
  }

  public getHistoryStatusDotClass(status: StockMovementStatus): string {
    if (status === "In Transit") {
      return "bg-[#f59e0b]";
    }
    if (status === "Pending") {
      return "bg-[#475467]";
    }
    return "bg-[#12b76a]";
  }

  public getHistoryStatusClass(status: StockMovementStatus): string {
    if (status === "In Transit") {
      return "text-[#b54708]";
    }
    if (status === "Pending") {
      return "text-[#475467]";
    }
    return "text-[#027a48]";
  }

  public formatSignedQuantity(quantity: number): string {
    return quantity >= 0 ? `+${quantity}` : String(quantity);
  }

  public trackBySku(_index: number, product: ProductRow): string {
    return product.sku;
  }

  public trackByLocation(_index: number, row: LocationStockRow): string {
    return String(row.stockId);
  }

  public trackByHistoryRow(_index: number, row: StockMovementHistoryRow): string {
    return row.id;
  }

  private fetchProducts(): void {
    this.isLoading = true;
    this.errorMessage = "";

    let params = new HttpParams()
      .set("organization_id", String(this.organizationId))
      .set("limit", String(this.limit))
      .set("offset", String(this.offset));

    const selectedCategoryId =
      this.selectedCategory === "All"
        ? undefined
        : this.categoryIdByName.get(this.selectedCategory);

    if (selectedCategoryId !== undefined) {
      params = params.set("category_id", String(selectedCategoryId));
    }

    this.http
      .get<CatalogResponse>(`${this.apiUrl}/api/products/catalog`, { params })
      .subscribe({
        next: (response) => {
          const items = Array.isArray(response?.data) ? response.data : [];
          this.products = items.map((item) => this.mapProduct(item));
          this.updateCategoryOptions(this.products);

          this.totalFromApi = this.extractTotal(response);
          this.hasNextPage = this.extractHasNext(response, this.products.length);

          this.totalProducts =
            this.totalFromApi ??
            (this.offset + this.products.length + (this.hasNextPage ? this.limit : 0));

          if (this.selectedProduct) {
            const updated = this.products.find(
              (product) => product.id === this.selectedProduct?.id
            );
            if (updated) {
              this.selectedProduct = updated;
            }
          }

          this.isLoading = false;
        },
        error: (error) => {
          console.error("❌ Failed to fetch catalog products:", error);
          this.products = [];
          this.totalProducts = 0;
          this.totalFromApi = null;
          this.hasNextPage = false;
          this.errorMessage = "Unable to load products. Please try again.";
          this.isLoading = false;
        },
      });
  }

  private mapProduct(item: CatalogProductResponseItem): ProductRow {
    const inventory = Array.isArray(item.inventory) ? item.inventory : [];
    const variants = Array.isArray(item.variants) ? item.variants : [];

    const category = (item.category_name || "Uncategorized").trim();
    const categoryId = this.getCategoryId(item);
    if (categoryId !== null) {
      this.categoryIdByName.set(category, categoryId);
    }

    const basePriceValue = this.getBasePrice(item);
    const stock = inventory.reduce(
      (sum, row) => sum + this.toNumber(row.quantity_on_hand),
      0
    );

    return {
      id: this.toNumber(item.id),
      sku: (item.sku || `SKU-${item.id}`).trim(),
      name: (item.name || "Unnamed Product").trim(),
      category,
      categoryPath: category,
      categoryId,
      manufacturer: (item.brand_name || "N/A").trim(),
      weight: this.extractWeight(variants),
      type: this.inferProductType(item),
      basePrice:
        basePriceValue !== null ? this.formatCurrency(basePriceValue) : "N/A",
      basePriceValue,
      stock: Number.isFinite(stock) ? stock : null,
      status: item.is_active ? "Active" : "Inactive",
      description: item.description || "",
      variants,
      inventory,
    };
  }

  private getCategoryId(item: CatalogProductResponseItem): number | null {
    const raw = item.category_id ?? item["categoryId"] ?? item["categoryID"];
    const parsed = Number(raw);
    return Number.isFinite(parsed) ? parsed : null;
  }

  private getBasePrice(item: CatalogProductResponseItem): number | null {
    const candidates = [
      item["base_price"],
      item["basePrice"],
      item["price"],
      item["selling_price"],
      item["retail_price"],
      item["effective_price"],
    ];

    for (const candidate of candidates) {
      const parsed = Number(candidate);
      if (Number.isFinite(parsed)) {
        return parsed;
      }
    }

    return null;
  }

  private inferProductType(item: CatalogProductResponseItem): ProductType {
    const rawType = String(item["product_type"] || "").toLowerCase();
    return rawType.includes("digital") ? "Digital" : "Physical";
  }

  private extractWeight(variants: CatalogVariant[]): string {
    for (const variant of variants) {
      const attrs = variant.variant_attributes ?? {};
      const weightCandidate =
        attrs["Weight"] ||
        attrs["weight"] ||
        attrs["Quantity"] ||
        attrs["quantity"];
      if (typeof weightCandidate === "string" && weightCandidate.trim().length > 0) {
        return weightCandidate.trim();
      }
    }

    return "N/A";
  }

  private updateCategoryOptions(rows: ProductRow[]): void {
    const existing = new Set(this.categoryOptions.filter((name) => name !== "All"));
    rows.forEach((row) => {
      if (row.category) {
        existing.add(row.category);
      }
    });

    this.categoryOptions = ["All", ...Array.from(existing).sort()];
    if (!this.categoryOptions.includes(this.selectedCategory)) {
      this.selectedCategory = "All";
    }
  }

  private extractTotal(response: CatalogResponse): number | null {
    const candidates = [
      response.total,
      response.total_count,
      response.count,
      response.meta?.total,
      response.pagination?.total,
    ];

    for (const candidate of candidates) {
      const parsed = Number(candidate);
      if (Number.isFinite(parsed)) {
        return parsed;
      }
    }

    return null;
  }

  private extractHasNext(response: CatalogResponse, currentLength: number): boolean {
    const explicitCandidates = [
      response.has_next,
      response.hasNext,
      response.meta?.has_next,
      response.meta?.hasNext,
      response.pagination?.has_next,
      response.pagination?.hasNext,
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

  private computeStockHealth(
    available: number,
    reorderLevel: number
  ): StockHealthStatus {
    if (reorderLevel <= 0) {
      return "Optimal";
    }

    if (available <= reorderLevel * 0.4) {
      return "Critical";
    }

    if (available <= reorderLevel) {
      return "Low Stock";
    }

    return "Optimal";
  }

  private toNumber(value: any): number {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : 0;
  }

  private formatCurrency(value: number): string {
    return value.toLocaleString(undefined, {
      style: "currency",
      currency: "USD",
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    });
  }

  private formatCompactCurrency(value: number): string {
    return value.toLocaleString(undefined, {
      style: "currency",
      currency: "USD",
      notation: "compact",
      maximumFractionDigits: 1,
    });
  }

  private createDefaultStockUpdateForm(): StockUpdateForm {
    const stockId = this.selectedStockId ?? this.stockByLocationRows[0]?.stockId ?? null;
    const selectedInventory = this.getSelectedVariantInventoryRows().find(
      (row) => this.toNumber(row.stock_id) === stockId
    );

    return {
      locationStockId: stockId ? String(stockId) : "default",
      adjustmentType: "Set Exact Quantity",
      quantity: this.toNumber(selectedInventory?.quantity_on_hand),
      reasonCode: this.stockUpdateReasonCodes[0],
      referenceNumber: "",
      notes: "",
    };
  }

  private getSelectedVariantInventoryRows(): CatalogInventory[] {
    if (!this.selectedProduct) {
      return [];
    }

    const allRows = this.selectedProduct.inventory ?? [];
    if (this.selectedVariantId === null) {
      return allRows;
    }

    const filtered = allRows.filter(
      (row) => this.toNumber(row.product_variant_id) === this.selectedVariantId
    );

    return filtered.length > 0 ? filtered : allRows;
  }

  private syncSelectedStockId(): void {
    const rows = this.getSelectedVariantInventoryRows();
    const firstStockId = this.toNumber(rows[0]?.stock_id);
    this.selectedStockId = firstStockId > 0 ? firstStockId : null;
  }

  private getSelectedInventoryForUpdate(): CatalogInventory | null {
    const rows = this.getSelectedVariantInventoryRows();
    if (rows.length === 0) {
      return null;
    }

    const selectedStockId =
      this.toNumber(this.stockUpdateForm.locationStockId) || this.selectedStockId;
    const selectedRow = rows.find(
      (row) => this.toNumber(row.stock_id) === selectedStockId
    );

    if (selectedRow) {
      return selectedRow;
    }

    return rows[0];
  }

  private buildInventoryUpsertPayload(
    inventory: CatalogInventory,
    quantityOnHand: number
  ): InventoryStockUpsertPayload {
    const variantId = this.toNumber(inventory.product_variant_id);

    return {
      max_stock_level: String(this.toNumber(inventory.max_stock_level)),
      metadata:
        inventory.metadata && typeof inventory.metadata === "object"
          ? inventory.metadata
          : {},
      product_id: this.toNumber(inventory.product_id) || this.selectedProduct?.id || 0,
      product_variant_id: variantId > 0 ? variantId : null,
      quantity_allocated: String(this.toNumber(inventory.quantity_allocated)),
      quantity_available: String(this.toNumber(inventory.quantity_available)),
      quantity_in_transit: String(this.toNumber(inventory.quantity_in_transit)),
      quantity_on_hand: String(quantityOnHand),
      quantity_on_order: String(this.toNumber(inventory.quantity_on_order)),
      reorder_level: String(this.toNumber(inventory.reorder_level)),
      reorder_quantity: String(this.toNumber(inventory.reorder_quantity)),
      storage_location_id: this.toNumber(inventory.storage_location_id),
      store_id: this.toNumber(inventory.store_id),
    };
  }

  private buildInitialHistoryRows(product: ProductRow): StockMovementHistoryRow[] {
    const primaryLocation =
      this.stockLocationOptions[0]?.label ?? "Main Warehouse -> Aisle A-12";
    const secondaryLocation =
      this.stockLocationOptions[1]?.label ?? "Retail Store -> Front Shelf";

    return [
      {
        id: `${product.id}-1`,
        dateTime: "2023-10-24 14:30",
        movementType: "Stock In",
        reference: "PO-44502",
        fromTo: "Global Vendor -> Main Whse",
        location: primaryLocation,
        quantity: 150,
        performedBy: "Sarah Jenkins",
        status: "Completed",
      },
      {
        id: `${product.id}-2`,
        dateTime: "2023-10-23 09:15",
        movementType: "Transfer",
        reference: "TR-982",
        fromTo: "Main Whse -> Retail Store",
        location: secondaryLocation,
        quantity: 25,
        performedBy: "Mark Thompson",
        status: "In Transit",
      },
      {
        id: `${product.id}-3`,
        dateTime: "2023-10-22 16:45",
        movementType: "Stock Out",
        reference: "SO-11209",
        fromTo: "Main Whse -> Customer",
        location: primaryLocation,
        quantity: -12,
        performedBy: "Auto-System",
        status: "Completed",
      },
      {
        id: `${product.id}-4`,
        dateTime: "2023-10-21 11:20",
        movementType: "Adjustment",
        reference: "ADJ-443",
        fromTo: "Main Whse",
        location: primaryLocation,
        quantity: -1,
        performedBy: "Elena Rodriguez",
        status: "Completed",
      },
      {
        id: `${product.id}-5`,
        dateTime: "2023-10-18 10:00",
        movementType: "Stock In",
        reference: "PO-44122",
        fromTo: "Returns -> Main Whse",
        location: primaryLocation,
        quantity: 2,
        performedBy: "Sarah Jenkins",
        status: "Completed",
      },
    ];
  }

  private getHistoryTypeFromAdjustment(
    adjustmentType: StockUpdateAdjustmentType
  ): StockMovementType {
    if (adjustmentType === "Decrement (-)") {
      return "Stock Out";
    }
    if (adjustmentType === "Set Exact Quantity") {
      return "Adjustment";
    }
    return "Stock In";
  }

  private formatDateTime(date: Date): string {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");
    const hour = String(date.getHours()).padStart(2, "0");
    const minute = String(date.getMinutes()).padStart(2, "0");
    return `${year}-${month}-${day} ${hour}:${minute}`;
  }
}