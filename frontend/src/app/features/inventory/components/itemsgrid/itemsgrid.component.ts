import { CommonModule } from "@angular/common";
import { Component } from "@angular/core";
import { Router } from "@angular/router";

@Component({
  selector: "app-itemsgrid",
  standalone: true,
  imports: [CommonModule],
  templateUrl: "./itemsgrid.component.html",
})
export class ItemsgridComponent {
  public products = [
    {
      id: 1,
      name: "Organic Apple",
      price: 1.99,
      image: "/images/products/image-1.png",
      categoryId: 2,
      category: "Fruits",
    },
    {
      id: 2,
      name: "Gourmet Cheese",
      price: 7.49,
      image: "/images/products/image-2.png",
      categoryId: 5,
      category: "Dairy & Eggs",
    },
    {
      id: 3,
      name: "Whole Wheat Bread",
      price: 2.49,
      image: "/images/products/image-3.png",
      categoryId: 4,
      category: "Bakery",
    },
    {
      id: 4,
      name: "Organic Honey",
      price: 5.99,
      image: "/images/products/image-4.png",
      categoryId: 8,
      category: "Honey",
    },
    {
      id: 5,
      name: "Fresh Strawberries",
      price: 3.99,
      image: "/images/products/image-5.png",
      categoryId: 2,
      category: "Fruits",
    },
    {
      id: 6,
      name: "Almond Butter",
      price: 8.49,
      image: "/images/products/image-6.png",
      categoryId: 12,
      category: "Butter",
    },
    {
      id: 7,
      name: "Herbal Tea Mix",
      price: 4.99,
      image: "/images/products/image-7.png",
      categoryId: 9,
      category: "Coffee",
    },
    {
      id: 8,
      name: "Dark Chocolate",
      price: 6.99,
      image: "/images/products/image-8.png",
      categoryId: 6,
      category: "Sweet",
    },
    {
      id: 9,
      name: "Extra Virgin Olive Oil",
      price: 10.99,
      image: "/images/products/image-9.png",
      categoryId: 10,
      category: "Oil",
    },
    {
      id: 10,
      name: "Fresh Blueberries",
      price: 4.49,
      image: "/images/products/image-10.png",
      categoryId: 2,
      category: "Fruits",
    },
    {
      id: 11,
      name: "Organic Spinach",
      price: 2.99,
      image: "/images/products/image-11.png",
      categoryId: 3,
      category: "Vegetables",
    },
    {
      id: 12,
      name: "Cinnamon Sticks",
      price: 3.49,
      image: "/images/products/image-12.png",
      categoryId: 11,
      category: "Spices",
    },
    {
      id: 13,
      name: "Quinoa Pack",
      price: 7.99,
      image: "/images/products/image-13.png",
      categoryId: 7,
      category: "Healthy",
    },
    {
      id: 14,
      name: "Avocado",
      price: 2.79,
      image: "/images/products/image-14.png",
      categoryId: 2,
      category: "Fruits",
    },
  ];

  public categories = [
    { id: 1, name: "All", image: "/images/products/image-1.png" },
    { id: 2, name: "Fruits", image: "/images/products/image-5.png" },
    { id: 3, name: "Vegetables", image: "/images/products/image-11.png" },
    { id: 4, name: "Bakery", image: "/images/products/image-3.png" },
    { id: 5, name: "Dairy & Eggs", image: "/images/products/image-2.png" },
    { id: 6, name: "Sweet", image: "/images/products/image-8.png" },
    { id: 7, name: "Healthy", image: "/images/products/image-13.png" },
    { id: 8, name: "Honey", image: "/images/products/image-4.png" },
    { id: 9, name: "Coffee", image: "/images/products/image-7.png" },
    { id: 10, name: "Oil", image: "/images/products/image-9.png" },
    { id: 11, name: "Spices", image: "/images/products/image-12.png" },
    { id: 12, name: "Butter", image: "/images/products/image-6.png" },
  ];

  public selectedCategory = this.categories[0];
  public filteredProducts = this.products;
  constructor(private router: Router) {}
  selectCategory(cat: any) {
    this.selectedCategory = cat;
    this.filteredProducts =
      cat.name === "All"
        ? this.products
        : this.products.filter((p) => p.categoryId === cat.id);
  }
  viewProductDetail(product: any) {
    this.router.navigate(["/inventory/product-detail", product.id], {
      state: {
        data: product,
      },
    });
  }
}
