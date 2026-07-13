import { ChangeDetectorRef, Component, OnDestroy, OnInit } from "@angular/core";
import { TranslateModule } from "@ngx-translate/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { HttpClient } from "@angular/common/http";
import { forkJoin, Subscription } from "rxjs";
import { environment } from "../../../../../../../environments/environment";
import { AddProductService } from "../../../../../../core/services/add-product.service";
import { ToastyService } from "../../../../../../core/services/toasty.service";

export interface PricingRow {
  priceListId: number | null;
  productVariantId: number | null;
  barcode: string;
  price: number | null;
  minQty: number | null;
  maxQty: number | null;
  duplicateError?: boolean;
}

export interface UomRow {
  uomId: string | null;
  uomName: string;
  convFactor: number | null;
  convUnit: string;
  priceListId: number | null;
  productVariantId: number | null;
  barcode: string;
  price: number | null;
  minQty: number | null;
  maxQty: number | null;
  duplicateError?: boolean;
}

@Component({
  selector: "add-price",
  standalone: true,
  imports: [TranslateModule, FormsModule, CommonModule],
  templateUrl: "./price.component.html",
})
export class PriceComponent implements OnInit, OnDestroy {
  apiUrl = environment.baseUrl;
  baseUnit: string = "";
  baseUomId: number | null = null;
  currencySymbol = "SAR ";
  hasVariants = false;
  productVariants: { id: number; name: string }[] = [];

  priceList: { id: number; name: string }[] = [];
  uomOptions: { value: string; label: string }[] = [];
  pricingRows: PricingRow[] = [this.emptyPricingRow()];
  uomRows: UomRow[] = [this.emptyUomRow()];

  private variantsSub?: Subscription;

  constructor(
    private http: HttpClient,
    private addProductService: AddProductService,
    private cdr: ChangeDetectorRef,
    private toasty: ToastyService
  ) { }

  ngOnInit(): void {
    this.fetchPriceLists();
    this.fetchUomOptions();

    this.baseUnit = this.addProductService.getBaseUnit() ?? "";
    this.hasVariants = !!this.addProductService.getPayload()?.has_variants;

    const product = this.addProductService.getCreatedProduct();
    this.baseUomId = product?.base_uom_id ?? null;

    this.loadCreatedVariants();
    this.variantsSub = this.addProductService
      .getCreatedProductVariants$()
      .subscribe(() => this.loadCreatedVariants());
  }

  ngOnDestroy(): void {
    this.variantsSub?.unsubscribe();
  }

  get pricingGridTemplate(): string {
    return this.hasVariants
      ? "100px 1fr 160px 100px 100px 180px 160px 52px"
      : "100px 1fr 160px 100px 100px 52px";
  }

  get uomGridTemplate(): string {
    return this.hasVariants
      ? "44px 150px 20px 110px 110px 1fr 85px 85px 120px 180px 160px 52px"
      : "44px 150px 20px 110px 110px 1fr 85px 85px 120px 52px";
  }

  fetchPriceLists(): void {
    this.http.get<any>(`${this.apiUrl}/api/price-lists/active`).subscribe((res) => {
      this.priceList = res.data.map((pl: any) => ({ id: pl.id, name: pl.name }));
    });
  }

  fetchUomOptions(): void {
    this.http.get<any>(`${this.apiUrl}/api/uoms/active`).subscribe((res) => {
      this.uomOptions = res.data.map((u: any) => ({
        value: String(u.id),
        label: u.name,
      }));
    });
  }

  availablePriceLists(currentIndex: number): { id: number; name: string }[] {
    if (this.hasVariants) {
      return this.priceList;
    }

    const usedIds = this.pricingRows
      .filter((_, i) => i !== currentIndex)
      .map((r) => r.priceListId)
      .filter((id): id is number => id !== null);

    return this.priceList.filter((pl) => !usedIds.includes(pl.id));
  }

  addPricingRow(): void {
    const last = this.pricingRows[this.pricingRows.length - 1];
    if (!last.priceListId || !last.price) return;
    if (this.checkPricingDuplicate(last, this.pricingRows.length - 1)) return;
    this.pricingRows.push(this.emptyPricingRow());
  }

  removePricingRow(index: number): void {
    if (this.pricingRows.length > 1) {
      this.pricingRows.splice(index, 1);
      this.clearPricingErrors();
    }
  }

  onPricingChange(index: number): void {
    this.checkPricingDuplicate(this.pricingRows[index], index);
  }

  private checkPricingDuplicate(row: PricingRow, index: number): boolean {
    const isDuplicate = this.pricingRows.some(
      (r, i) =>
        i !== index &&
        r.priceListId !== null &&
        Number(r.priceListId) === Number(row.priceListId) &&
        (!this.hasVariants ||
          (r.productVariantId !== null &&
            row.productVariantId !== null &&
            Number(r.productVariantId) === Number(row.productVariantId)))
    );
    row.duplicateError = isDuplicate;
    return isDuplicate;
  }

  private clearPricingErrors(): void {
    this.pricingRows.forEach((r) => (r.duplicateError = false));
    this.pricingRows.forEach((r, i) => this.checkPricingDuplicate(r, i));
  }

  private emptyPricingRow(): PricingRow {
    return {
      priceListId: null,
      productVariantId: null,
      barcode: "",
      price: null,
      minQty: null,
      maxQty: null,
      duplicateError: false,
    };
  }

  getConvUnitOptions(rowIndex: number): string[] {
    const options = new Set<string>([this.baseUnit || "-"]);
    for (let i = 0; i < rowIndex; i++) {
      const prev = this.uomRows[i];
      if (prev?.uomName) options.add(prev.uomName);
    }
    return Array.from(options);
  }

  onUomChange(index: number): void {
    const row = this.uomRows[index];
    if (row.uomId) {
      const match = this.uomOptions.find((opt) => opt.value === row.uomId);
      row.uomName = match?.label ?? "";
    } else {
      row.uomName = "";
    }

    row.convUnit = this.baseUnit || "-";
    row.duplicateError = false;
    this.checkUomDuplicate(row, index);
    this.cdr.detectChanges();
  }

  onUomPriceListChange(index: number): void {
    this.checkUomDuplicate(this.uomRows[index], index);
  }

  addUomRow(): void {
    const last = this.uomRows[this.uomRows.length - 1];
    if (!last.uomId || !last.priceListId || !last.price) return;
    if (this.checkUomDuplicate(last, this.uomRows.length - 1)) return;
    this.uomRows.push(this.emptyUomRow());
  }

  removeUomRow(index: number): void {
    if (this.uomRows.length > 1) {
      this.uomRows.splice(index, 1);
      this.clearUomErrors();
    }
  }

  private checkUomDuplicate(row: UomRow, index: number): boolean {
    const isDuplicate = this.uomRows.some(
      (r, i) =>
        i !== index &&
        r.uomId !== null &&
        r.uomId === row.uomId &&
        r.priceListId !== null &&
        Number(r.priceListId) === Number(row.priceListId) &&
        (!this.hasVariants ||
          (r.productVariantId !== null &&
            row.productVariantId !== null &&
            Number(r.productVariantId) === Number(row.productVariantId)))
    );
    row.duplicateError = isDuplicate;
    return isDuplicate;
  }

  private clearUomErrors(): void {
    this.uomRows.forEach((r) => (r.duplicateError = false));
    this.uomRows.forEach((r, i) => this.checkUomDuplicate(r, i));
  }

  private emptyUomRow(): UomRow {
    return {
      uomId: null,
      uomName: "",
      convFactor: null,
      convUnit: this.baseUnit || "-",
      priceListId: null,
      productVariantId: null,
      barcode: "",
      price: null,
      minQty: null,
      maxQty: null,
      duplicateError: false,
    };
  }

  getPriceListName(id: number | string | null): string | null {
    if (id === null || id === undefined || id === "") return null;
    return this.priceList.find((pl) => pl.id === Number(id))?.name ?? null;
  }

  getVariantName(id: number | string | null): string | null {
    if (id === null || id === undefined || id === "") return null;
    return this.productVariants.find((v) => v.id === Number(id))?.name ?? null;
  }

  hasAnySummary(): boolean {
    const hasPricing = this.pricingRows.some((r) => r.price);
    const hasUom = this.uomRows.some((r) => r.uomId && r.convFactor && r.price);
    return hasPricing || hasUom;
  }

  get hasAnyError(): boolean {
    return (
      this.pricingRows.some((r) => r.duplicateError) ||
      this.uomRows.some((r) => r.duplicateError)
    );
  }

  onSubmit(): void {
    if (this.hasAnyError) return;

    const product = this.addProductService.getCreatedProduct();
    if (!product?.id) {
      this.toasty.error("Product not found. Complete previous steps first.");
      return;
    }

    if (this.hasVariants && this.productVariants.length === 0) {
      this.toasty.error("Please save product variants first.");
      return;
    }

    const productId = Number(product.id);
    const baseRows = this.pricingRows.filter((r) => r.priceListId && r.price != null);
    const packRows = this.uomRows.filter((r) => r.uomId && r.priceListId && r.price != null);

    if (this.hasVariants) {
      const baseMissingVariant = baseRows.some((r) => !r.productVariantId);
      const packMissingVariant = packRows.some((r) => !r.productVariantId);
      if (baseMissingVariant || packMissingVariant) {
        this.toasty.error("Please select product variant for all filled price rows.");
        return;
      }
    }

    const priceEntries: any[] = [];

    baseRows.forEach((row) => {
      priceEntries.push({
        is_active: true,
        price_list_id: Number(row.priceListId),
        product_id: productId,
        product_variant_id: this.hasVariants ? Number(row.productVariantId) : null,
        uom_id: this.baseUomId,
        min_quantity: row.minQty ?? 0,
        max_quantity: row.maxQty ?? 0,
        price: row.price,
        valid_from: null,
        valid_to: null,
        metadata: { additionalProp1: {} },
      });
    });

    packRows.forEach((row) => {
      priceEntries.push({
        is_active: true,
        price_list_id: Number(row.priceListId),
        product_id: productId,
        product_variant_id: this.hasVariants ? Number(row.productVariantId) : null,
        uom_id: Number(row.uomId),
        min_quantity: row.minQty ?? 0,
        max_quantity: row.maxQty ?? 0,
        price: row.price,
        valid_from: null,
        valid_to: null,
        metadata: { additionalProp1: {} },
      });
    });

    const priceRequests = priceEntries.map((payload) =>
      this.addProductService.postProductPrice(payload)
    );

    const barcodeRequests = this.hasVariants
      ? this.buildVariantBarcodeRequests(productId)
      : [];

    const requests = [...priceRequests, ...barcodeRequests];

    if (requests.length === 0) {
      this.toasty.error("Please add at least one pricing row.");
      return;
    }

    forkJoin(requests).subscribe({
      next: () => {
        this.toasty.success("Prices saved successfully!");
      },
      error: (err) => {
        console.error("Price/barcode creation failed", err);
        this.toasty.error("Price creation failed");
      },
    });
  }

  private loadCreatedVariants(): void {
    const variants = this.addProductService.getCreatedProductVariants() || [];
    this.productVariants = variants
      .map((variant: any) => ({
        id: Number(variant?.id || 0),
        name: variant?.variant_name || "",
      }))
      .filter((variant) => variant.id > 0 && !!variant.name);
  }

  private buildVariantBarcodeRequests(productId: number) {
    const rows = [...this.pricingRows, ...this.uomRows] as Array<{
      productVariantId: number | null;
      barcode: string;
    }>;
    const seen = new Set<string>();

    return rows
      .map((row) => ({
        variantId: Number(row.productVariantId || 0),
        barcode: (row.barcode || "").trim(),
      }))
      .filter((row) => row.variantId > 0 && !!row.barcode)
      .filter((row) => {
        const key = `${row.variantId}|${row.barcode}`;
        if (seen.has(key)) return false;
        seen.add(key);
        return true;
      })
      .map((row) => {
        const payload = {
          barcode: row.barcode,
          barcode_type: "EAN13",
          is_primary: true,
          metadata: {},
          product_id: productId,
          product_variant_id: row.variantId,
        };
        return this.http.post(`${this.addProductService.apiUrl}/api/product-barcodes`, payload);
      });
  }
}