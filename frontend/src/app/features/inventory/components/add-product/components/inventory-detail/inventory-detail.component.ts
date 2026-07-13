import { CommonModule } from "@angular/common";
import { Component, EventEmitter, OnDestroy, OnInit, Output } from "@angular/core";
import { TranslateModule } from "@ngx-translate/core";
import { environment } from "../../../../../../../environments/environment";
import { HttpClient } from "@angular/common/http";
import { FormsModule } from "@angular/forms";
import { AddProductService } from "../../../../../../core/services/add-product.service";
import { ToastyService } from "../../../../../../core/services/toasty.service";
import { Subscription } from "rxjs";

interface ProductVariantOption {
  id: number;
  variant_name: string;
  variant_sku: string;
  is_active: boolean;
  variant_attributes: Record<string, string>;
  image_url: string;
}

@Component({
  selector: "add-inventory-detail",
  imports: [TranslateModule, CommonModule, FormsModule],
  templateUrl: "./inventory-detail.component.html",
})
export class InventoryDetailComponent implements OnInit, OnDestroy {
  @Output() stepComplete = new EventEmitter<void>();
  apiUrl = environment.baseUrl;
  stores: { id: number; name: string }[] = [];
  storageLocations: { id: number; name: string }[] = [];
  isSubmitting = false;
  errorMessage = "";
  successMessage = "";
  hasVariants = false;
  defaultVariantImage = "/images/products/image-1.png";
  productVariants: ProductVariantOption[] = [];
  selectedVariantCards: ProductVariantOption[] = [];
  private variantsSub?: Subscription;
  inventoryPayload: {
    max_stock_level: string;
    metadata: { [key: string]: any };
    product_id: number;
    product_variant_id: number | null;
    quantity_allocated: string;
    quantity_available: string;
    quantity_in_transit: string;
    quantity_on_hand: string;
    quantity_on_order: string;
    reorder_level: string;
    reorder_quantity: string;
    storage_location_id: number | null;
    store_id: number | null;
  } = {
      max_stock_level: "",
      metadata: {},
      product_id: 0,
      product_variant_id: null,
      quantity_allocated: "",
      quantity_available: "",
      quantity_in_transit: "",
      quantity_on_hand: "",
      quantity_on_order: "",
      reorder_level: "",
      reorder_quantity: "",
      storage_location_id: null,
      store_id: null,
    };

  constructor(
    private http: HttpClient,
    public addProductService: AddProductService,
    private toasty: ToastyService
  ) { }

  ngOnInit(): void {
    this.fetchStores();
    const product = this.addProductService.getCreatedProduct();
    this.hasVariants = !!this.addProductService.getPayload()?.has_variants;
    this.loadCreatedVariants();

    this.variantsSub = this.addProductService
      .getCreatedProductVariants$()
      .subscribe(() => this.loadCreatedVariants());

    if (product) {
      this.inventoryPayload.product_id = product.id ?? product.product_id ?? 0;
    }
  }

  ngOnDestroy(): void {
    this.variantsSub?.unsubscribe();
  }

  fetchStores(): void {
    this.http.get<any>(`${this.apiUrl}/api/stores`).subscribe({
      next: (res) => {
        this.stores = res.data.map((store: any) => ({
          id: store.id,
          name: store.name,
        }));
      },
      error: (err) => {
        console.error("Failed to fetch stores:", err);
      },
    });
  }

  onStoreChange(storeId: number | null): void {
    this.storageLocations = [];
    this.inventoryPayload.storage_location_id = null;

    if (!storeId) return;

    this.http
      .get<any>(`${this.apiUrl}/api/stores/${storeId}/storage-locations`)
      .subscribe({
        next: (res) => {
          this.storageLocations = res.data.map((loc: any) => ({
            id: loc.id,
            name: loc.name,
          }));
        },
        error: (err) => {
          console.error("Failed to fetch storage locations:", err);
        },
      });
  }

  onVariantSelectionChange(variantId: number | null): void {
    this.addVariantCardById(variantId);
  }

  onVariantImageError(event: Event): void {
    const element = event.target as HTMLImageElement;
    if (element.src.includes(this.defaultVariantImage)) {
      return;
    }

    element.src = this.defaultVariantImage;
  }

  trackByVariantCard(_: number, variant: ProductVariantOption): number {
    return variant.id;
  }

  onSubmit(): void {
    this.errorMessage = "";
    this.successMessage = "";

    if (!this.inventoryPayload.store_id) {
      this.errorMessage = "Please select a store.";
      return;
    }

    if (!this.inventoryPayload.storage_location_id) {
      this.errorMessage = "Please select a storage location.";
      return;
    }

    if (!this.inventoryPayload.product_id) {
      this.errorMessage = "Product not found. Please complete the previous step.";
      return;
    }

    if (this.hasVariants) {
      if (this.productVariants.length === 0) {
        this.errorMessage = "Please save product variants first.";
        return;
      }

      if (!this.inventoryPayload.product_variant_id) {
        this.errorMessage = "Please select a product variant.";
        return;
      }
    }

    this.addVariantCardById(this.inventoryPayload.product_variant_id);

    this.isSubmitting = true;
    const payload = {
      ...this.inventoryPayload,
      store_id: Number(this.inventoryPayload.store_id),
      storage_location_id: Number(this.inventoryPayload.storage_location_id),
      product_id: Number(this.inventoryPayload.product_id),
      product_variant_id: this.hasVariants
        ? Number(this.inventoryPayload.product_variant_id)
        : null,
      max_stock_level: String(this.inventoryPayload.max_stock_level),
      quantity_allocated: String(this.inventoryPayload.quantity_allocated),
      quantity_available: String(this.inventoryPayload.quantity_available),
      quantity_in_transit: String(this.inventoryPayload.quantity_in_transit),
      quantity_on_hand: String(this.inventoryPayload.quantity_on_hand),
      quantity_on_order: String(this.inventoryPayload.quantity_on_order),
      reorder_level: String(this.inventoryPayload.reorder_level),
      reorder_quantity: String(this.inventoryPayload.reorder_quantity),
    };

    this.addProductService.postInventory(payload).subscribe({
      next: (res) => {
        console.log("Inventory posted successfully:", res);
        this.successMessage = "Inventory details saved successfully.";
        this.stepComplete.emit();
        this.toasty.success('Inventory details saved successfully.');
        this.isSubmitting = false;
      },
      error: (err) => {
        console.error("Failed to post inventory:", err);
        this.errorMessage = "Failed to save inventory. Please try again.";
        this.toasty.error('Failed to save inventory. Please try again.');
        this.isSubmitting = false;
      },
    });
  }

  private loadCreatedVariants(): void {
    const variants = this.addProductService.getCreatedProductVariants() || [];
    this.productVariants = variants
      .map((variant: any) => ({
        id: Number(variant?.id || 0),
        variant_name: variant?.variant_name || "",
        variant_sku: variant?.variant_sku || "",
        is_active: variant?.is_active ?? true,
        variant_attributes: this.normalizeVariantAttributes(
          variant?.variant_attributes || variant?.metadata?.variant_attributes
        ),
        image_url: this.resolveVariantImage(variant),
      }))
      .filter(variant => variant.id > 0 && !!variant.variant_name);

    if (!this.hasVariants) {
      this.inventoryPayload.product_variant_id = null;
      return;
    }

    const selectedVariantId = Number(this.inventoryPayload.product_variant_id || 0);
    const stillExists = this.productVariants.some(v => v.id === selectedVariantId);
    if (!stillExists) {
      this.inventoryPayload.product_variant_id = null;
    }

    this.syncSelectedVariantCards();

    if (stillExists) {
      this.addVariantCardById(selectedVariantId);
    }
  }

  private addVariantCardById(variantId: number | null): void {
    const id = Number(variantId || 0);
    if (!id) {
      return;
    }

    const selected = this.productVariants.find(variant => variant.id === id);
    if (!selected) {
      return;
    }

    const exists = this.selectedVariantCards.some(variant => variant.id === id);
    if (exists) {
      return;
    }

    this.selectedVariantCards = [...this.selectedVariantCards, selected];
  }

  private syncSelectedVariantCards(): void {
    if (this.selectedVariantCards.length === 0) {
      return;
    }

    const variantMap = new Map(this.productVariants.map(variant => [variant.id, variant]));
    this.selectedVariantCards = this.selectedVariantCards.map(card => {
      return variantMap.get(card.id) || card;
    });
  }

  private resolveVariantImage(variant: any): string {
    const metadata = this.parseMetadata(variant?.metadata);
    const product = this.addProductService.getCreatedProduct() || {};
    const productMetadata = this.parseMetadata(product?.metadata);

    const candidates = [
      variant?.image_url,
      variant?.image,
      variant?.thumbnail_url,
      variant?.thumbnail,
      variant?.photo_url,
      metadata?.["image_url"],
      metadata?.["image"],
      metadata?.["thumbnail_url"],
      metadata?.["thumbnail"],
      product?.image_url,
      product?.image,
      productMetadata?.["image_url"],
      productMetadata?.["image"],
    ];

    const firstValid = candidates.find(
      candidate => typeof candidate === "string" && candidate.trim().length > 0
    );

    return firstValid || this.defaultVariantImage;
  }

  private parseMetadata(metadata: any): Record<string, any> {
    if (!metadata) {
      return {};
    }

    if (typeof metadata === "object") {
      return metadata;
    }

    if (typeof metadata === "string") {
      try {
        return JSON.parse(metadata);
      } catch {
        return {};
      }
    }

    return {};
  }

  private normalizeVariantAttributes(
    variantAttributes: any
  ): Record<string, string> {
    if (!variantAttributes) {
      return {};
    }

    if (typeof variantAttributes === "string") {
      try {
        const parsed = JSON.parse(variantAttributes);
        if (parsed && typeof parsed === "object" && !Array.isArray(parsed)) {
          return Object.entries(parsed).reduce((acc, [key, value]) => {
            acc[String(key)] = String(value ?? "");
            return acc;
          }, {} as Record<string, string>);
        }
      } catch {
        return {};
      }
      return {};
    }

    if (typeof variantAttributes === "object" && !Array.isArray(variantAttributes)) {
      return Object.entries(variantAttributes).reduce((acc, [key, value]) => {
        acc[String(key)] = String(value ?? "");
        return acc;
      }, {} as Record<string, string>);
    }

    return {};
  }
}