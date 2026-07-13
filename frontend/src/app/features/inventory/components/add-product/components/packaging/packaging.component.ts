import { Component, OnDestroy, OnInit } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { TranslateModule } from "@ngx-translate/core";
import { forkJoin, Subscription } from "rxjs";
import { AddProductService } from "../../../../../../core/services/add-product.service";
import { ToastyService } from "../../../../../../core/services/toasty.service";

interface VariantAttribute {
  name: string;
  inputValue: string;
  values: string[];
}

interface VariantRow {
  key: string;
  attributes: Record<string, string>;
  variantName: string;
  sku: string;
  active: boolean;
}

@Component({
  selector: "app-packaging",
  imports: [CommonModule, FormsModule, TranslateModule],
  templateUrl: "./packaging.component.html",
  styleUrl: "./packaging.component.scss",
})
export class PackagingComponent implements OnInit, OnDestroy {
  newAttributeName = "";
  attributes: VariantAttribute[] = [];
  variantRows: VariantRow[] = [];

  private attributeSub?: Subscription;

  constructor(
    public addProductService: AddProductService,
    private toasty: ToastyService
  ) { }

  ngOnInit(): void {
    this.attributeSub = this.addProductService
      .getVariantAttributes$()
      .subscribe(attributes => this.setupAttributes(attributes));
  }

  ngOnDestroy(): void {
    this.attributeSub?.unsubscribe();
  }

  get productName(): string {
    return (this.addProductService.getPayload()?.name || "Product").trim();
  }

  setupAttributes(attributeNames: string[]): void {
    const existingMap = new Map(
      this.attributes.map(attribute => [attribute.name, attribute])
    );

    this.attributes = (attributeNames || []).map(name => {
      const existing = existingMap.get(name);
      return {
        name,
        inputValue: "",
        values: existing?.values || [],
      };
    });

    this.rebuildVariantRows();
  }

  addAttribute(): void {
    const names = this.newAttributeName
      .split(",")
      .map(name => name.trim())
      .filter(Boolean);

    if (names.length === 0) {
      return;
    }

    const existingNames = new Set(
      this.attributes.map(attribute => attribute.name.toLowerCase())
    );

    const nextAttributes = [...this.attributes];
    names.forEach(name => {
      const lowerName = name.toLowerCase();
      if (existingNames.has(lowerName)) {
        return;
      }

      nextAttributes.push({
        name,
        inputValue: "",
        values: [],
      });
      existingNames.add(lowerName);
    });

    this.attributes = nextAttributes;
    this.newAttributeName = "";
    this.syncAttributeNamesToService();
    this.rebuildVariantRows();
  }

  addAttributeByEnter(event: KeyboardEvent): void {
    if (event.key !== "Enter") {
      return;
    }

    event.preventDefault();
    this.addAttribute();
  }

  removeAttribute(index: number): void {
    this.attributes.splice(index, 1);
    this.syncAttributeNamesToService();
    this.rebuildVariantRows();
  }

  addAllAttributeValues(): void {
    let hasChanges = false;

    this.attributes.forEach(attribute => {
      const enteredValues = attribute.inputValue
        .split(",")
        .map(value => value.trim())
        .filter(Boolean);

      enteredValues.forEach(value => {
        const duplicate = attribute.values.some(
          existing => existing.toLowerCase() === value.toLowerCase()
        );

        if (!duplicate) {
          attribute.values.push(value);
          hasChanges = true;
        }
      });

      attribute.inputValue = "";
    });

    if (hasChanges) {
      this.rebuildVariantRows();
    }
  }

  addValueByEnter(event: KeyboardEvent): void {
    if (event.key !== "Enter") {
      return;
    }

    event.preventDefault();
    this.addAllAttributeValues();
  }

  hasPendingValues(): boolean {
    return this.attributes.some(attribute => attribute.inputValue.trim().length > 0);
  }

  removeAttributeValue(attributeIndex: number, valueIndex: number): void {
    this.attributes[attributeIndex]?.values.splice(valueIndex, 1);
    this.rebuildVariantRows();
  }

  onVariantRowChange(): void {
    this.syncVariantsToService();
  }

  onSubmit(): void {
    this.syncVariantsToService();
    this.saveProductVariants();
  }

  trackByAttribute(_: number, attribute: VariantAttribute): string {
    return attribute.name;
  }

  trackByVariant(_: number, variant: VariantRow): string {
    return variant.key;
  }

  private rebuildVariantRows(): void {
    const activeAttributes = this.attributes.filter(
      attribute => attribute.values.length > 0
    );

    if (activeAttributes.length === 0) {
      this.variantRows = [];
      this.syncVariantsToService();
      return;
    }

    const nameOrder = activeAttributes.map(attribute => attribute.name);
    const existingRows = new Map(this.variantRows.map(row => [row.key, row]));
    const combinations = this.getCombinations(activeAttributes);
    const baseSku = (this.addProductService.getPayload()?.sku || "VAR")
      .toString()
      .trim();

    this.variantRows = combinations.map((combination, index) => {
      const key = this.getCombinationKey(combination, nameOrder);
      const existing = existingRows.get(key);
      const generatedSku = this.buildSku(baseSku, combination, index);

      return {
        key,
        attributes: combination,
        variantName: existing?.variantName || this.buildVariantName(combination),
        sku: existing?.sku || generatedSku,
        active: existing?.active ?? true,
      };
    });

    this.syncVariantsToService();
  }

  private getCombinations(attributes: VariantAttribute[]): Record<string, string>[] {
    let combinations: Record<string, string>[] = [{}];

    attributes.forEach(attribute => {
      const next: Record<string, string>[] = [];

      combinations.forEach(combination => {
        attribute.values.forEach(value => {
          next.push({
            ...combination,
            [attribute.name]: value,
          });
        });
      });

      combinations = next;
    });

    return combinations;
  }

  private getCombinationKey(
    combination: Record<string, string>,
    nameOrder: string[]
  ): string {
    return nameOrder.map(name => `${name}:${combination[name]}`).join("|");
  }

  private buildVariantName(combination: Record<string, string>): string {
    const parts = Object.values(combination);
    if (parts.length === 0) {
      return this.productName;
    }

    return `${this.productName} - ${parts.join(" / ")}`;
  }

  private buildSku(
    baseSku: string,
    combination: Record<string, string>,
    index: number
  ): string {
    const suffix = Object.values(combination)
      .map(value =>
        value
          .replace(/[^a-zA-Z0-9]/g, "")
          .toUpperCase()
          .slice(0, 4)
      )
      .filter(Boolean)
      .join("-");

    if (!suffix) {
      return `${baseSku}-${index + 1}`;
    }

    return `${baseSku}-${suffix}`;
  }

  private syncVariantsToService(): void {
    this.addProductService.setProductVariants(
      this.variantRows.map(variant => ({
        variant_name: variant.variantName,
        variant_sku: variant.sku,
        is_active: variant.active,
        variant_attributes: variant.attributes,
      }))
    );
  }

  private syncAttributeNamesToService(): void {
    this.addProductService.setVariantAttributes(
      this.attributes.map(attribute => attribute.name)
    );
  }

  private saveProductVariants(): void {
    const productId = Number(this.addProductService.getCreatedProduct()?.id || 0);
    const variants = this.addProductService.getProductVariants() || [];

    if (!productId) {
      this.toasty.error("Please create the product first from Tax Info tab");
      return;
    }

    if (variants.length === 0) {
      this.toasty.error("Please add at least one variant before saving");
      return;
    }

    const requests = variants.map((variant: any) => {
      const payload = {
        is_active: variant.is_active ?? true,
        metadata: variant.metadata || {},
        product_id: productId,
        variant_attributes: variant.variant_attributes || {},
        variant_name: variant.variant_name || "",
        variant_sku: variant.variant_sku || "",
      };

      return this.addProductService.postProductVariant(payload);
    });

    forkJoin(requests).subscribe({
      next: (res) => {
        console.log("Product variants created successfully:", res);
        const createdVariants = (res || []).map((item: any) => item?.data ?? item);
        this.addProductService.setCreatedProductVariants(createdVariants);
        this.toasty.success("Product variants saved successfully");
      },
      error: (err) => {
        console.error("Failed to create product variants:", err);
        this.toasty.error("Failed to save product variants");
      },
    });
  }
}