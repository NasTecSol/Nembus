import { CommonModule } from "@angular/common";
import { Component, OnInit } from "@angular/core";
import { TranslateModule } from "@ngx-translate/core";
import { environment } from "../../../../../../../environments/environment";
import { HttpClient } from '@angular/common/http';
import { FormsModule } from "@angular/forms";
import { AddProductService } from "../../../../../../core/services/add-product.service";
import { CheckIconComponent } from "../../../../../../shared/icons/check.component";

@Component({
  selector: "add-item-detail",
  imports: [CommonModule, TranslateModule, FormsModule, CheckIconComponent],
  templateUrl: "./item-detail.component.html",
})
export class ItemDetailComponent implements OnInit {
  public isTyping: boolean = false;
  public hasVariants: boolean = true;
  apiUrl = environment.baseUrl;
  productName: string = '';
  productDescription: string = '';
  productType: string = '';

  parentCategories: { id: number; name: string }[] = [];
  allCategories: any[] = [];
  childCategories: any[] = [];

  selectedParentId: number | null = null;
  stores: { id: number; name: string }[] = [];
  selectedStoreId: number | null = null;
  selectedChildCategoryId: number | null = null;

  uom: { id: number; name: string; code: string; type: string }[] = [];
  selectedUomId: number | null = null;
  barcodeValue: string = '';
  brands: { id: number; name: string }[] = [];
  selectedBrandId: number | null = null;


  constructor(private http: HttpClient, public addProductService: AddProductService) { }

  ngOnInit(): void {
    this.addProductService.setField("has_variants", this.hasVariants);
    this.fetchOragnization();
    this.fetchBrands();
    this.fetchParentCategories();
    this.fetchUOM();
  }
  fetchOragnization() {
    this.http.get<any>(`${this.apiUrl}/api/organizations`).subscribe(res => {
      this.stores = res.data.map((store: any) => ({
        id: store.id,
        name: store.name
      }));
    });
  }

  onStoreChange(value: string) {
    this.addProductService.setField('organization_id', Number(value));
  }

  fetchBrands() {
    this.http.get<any>(`${this.apiUrl}/api/brands/active`).subscribe(res => {
      this.brands = res.data.map((brand: any) => ({
        id: brand.id,
        name: brand.name
      }));
    });

  }
  onBrandChange(value: string) {
    this.addProductService.setField('brand_id', Number(value));
  }

  fetchParentCategories() {
    this.http.get<any>(`${this.apiUrl}/api/pos/categories`).subscribe(res => {
      this.allCategories = res.data;

      // unique parents
      const map = new Map<number, string>();

      this.allCategories.forEach(cat => {
        map.set(cat.parent_category_id, cat.parent_category_name);
      });

      this.parentCategories = Array.from(map, ([id, name]) => ({ id, name }));
    });
  }

  onParentChange(parentId: number) {
    const id = Number(parentId);

    this.childCategories = this.allCategories.filter(cat =>
      cat.parent_category_id === id
    );
    this.selectedChildCategoryId = null;
    this.addProductService.setVariantAttributes([]);

    console.log('Parent:', id);
    console.log('Children:', this.childCategories);
  }
  onChildCategoryChange(categoryId: number) {
    this.selectedChildCategoryId = Number(categoryId);
    this.addProductService.setField('category_id', this.selectedChildCategoryId);

    const selectedCategory = this.childCategories.find(
      cat => Number(cat.category_id) === this.selectedChildCategoryId
    );
    this.addProductService.setVariantAttributes(
      this.extractVariantAttributes(selectedCategory)
    );

    console.log('Saved category_id in service:', this.selectedChildCategoryId);
  }
  fetchUOM() {
    this.http.get<any>(`${this.apiUrl}/api/uoms/active`).subscribe(res => {
      this.uom = res.data.map((uom: any) => ({
        id: uom.id,
        name: uom.name,
        type: uom.uom_type,
        code: uom.code
      }));
    });
  }

  onDescriptionChange(value: string) {
    this.addProductService.setField('description', value);
  }

  onNameChange(value: string) {
    this.productName = value;
    this.addProductService.setField('name', value);
    const sku = value
      .trim()
      .toUpperCase()
      .replace(/\s+/g, '_');
    this.addProductService.setField('sku', sku);
    console.log('Generated SKU:', sku);
  }
  onItemTypeChange(value: string) {
    this.productType = value;
    this.addProductService.setField('product_type', value);
  }

  onSellableChange(event: Event) {
    const checked = (event.target as HTMLInputElement).checked;
    this.addProductService.setField('is_sellable', checked);
  }

  get isSellable(): boolean {
    return this.addProductService.getPayload().is_sellable;
  }
  onPurchasableChange(event: Event) {
    const checked = (event.target as HTMLInputElement).checked;
    this.addProductService.setField('is_purchasable', checked);
  }

  get isPurchasable(): boolean {
    return this.addProductService.getPayload().is_purchasable;
  }


  onUomChange(value: string) {
    this.addProductService.setField('base_uom_id', Number(value));
  }
  onBarcodeChange(value: string) {
    this.barcodeValue = value;
    // Save in main product payload
    this.addProductService.setField('barcode', value);
  }

  onHasVariantsChange(event: Event) {
    const checked = (event.target as HTMLInputElement).checked;
    this.hasVariants = checked;
    this.addProductService.setField("has_variants", checked);

    if (checked) {
      this.isTyping = false;
      this.barcodeValue = "";
      this.addProductService.setField("barcode", "");
    }
  }

  onClickPlusIcon() {
    this.isTyping = true;
  }
  onConfirm() {
    this.isTyping = false;
  }

  onCancel() {
    this.isTyping = false;
  }

  private extractVariantAttributes(category: any): string[] {
    if (!category) {
      return [];
    }

    const metadataAttributes = this.extractVariantAttributesFromMetadata(
      category.category_metadata ?? category.categoryMetadata
    );

    if (metadataAttributes.length > 0) {
      return metadataAttributes;
    }

    const candidates = [
      category.variant_attributes,
      category.variantAttributes,
      category.attributes,
      category.attribute_values,
      category.attributeValues,
      category.variant_options,
      category.variantOptions,
    ];

    const attributes: string[] = [];

    candidates.forEach(source => {
      if (!Array.isArray(source)) {
        return;
      }

      source.forEach((item: any) => {
        if (typeof item === "string") {
          attributes.push(item);
          return;
        }

        if (!item || typeof item !== "object") {
          return;
        }

        if (item.is_variant === false) {
          return;
        }

        const name =
          item.name ||
          item.attribute_name ||
          item.attributeName ||
          item.label ||
          item.title ||
          item.attribute?.name ||
          item.attribute?.attribute_name;

        if (name) {
          attributes.push(String(name));
        }
      });
    });

    return Array.from(
      new Set(attributes.map(attr => attr.trim()).filter(Boolean))
    );
  }

  private extractVariantAttributesFromMetadata(rawMetadata: unknown): string[] {
    if (!rawMetadata) {
      return [];
    }

    let metadataObj: any = rawMetadata;

    if (typeof rawMetadata === "string") {
      const trimmed = rawMetadata.trim();
      if (!trimmed) {
        return [];
      }

      metadataObj = this.tryParseJson(trimmed);

      if (!metadataObj) {
        const decoded = this.decodeBase64ToString(trimmed);
        metadataObj = decoded ? this.tryParseJson(decoded) : null;
      }
    }

    if (!metadataObj || typeof metadataObj !== "object") {
      return [];
    }

    const variantAttributes =
      metadataObj.variant_attributes ?? metadataObj.variantAttributes;

    if (!variantAttributes) {
      return [];
    }

    if (
      typeof variantAttributes === "object" &&
      !Array.isArray(variantAttributes)
    ) {
      return Object.keys(variantAttributes)
        .map(key => key.trim())
        .filter(Boolean);
    }

    if (Array.isArray(variantAttributes)) {
      return variantAttributes
        .map(item =>
          typeof item === "string"
            ? item
            : item?.name || item?.attribute_name || item?.attributeName
        )
        .map((name: any) => String(name || "").trim())
        .filter(Boolean);
    }

    return [];
  }

  private decodeBase64ToString(value: string): string | null {
    const normalized = value.replace(/-/g, "+").replace(/_/g, "/");
    const padding = (4 - (normalized.length % 4)) % 4;
    const padded = normalized + "=".repeat(padding);

    try {
      const binary = atob(padded);
      const bytes = Uint8Array.from(binary, char => char.charCodeAt(0));
      return new TextDecoder().decode(bytes);
    } catch {
      return null;
    }
  }

  private tryParseJson(value: string): any | null {
    try {
      return JSON.parse(value);
    } catch {
      return null;
    }
  }
}