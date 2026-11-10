# Master product stock catalog

`GET /api/products/master-catalog/stock`

Uses the same authentication and `x-tenant-id` header as the master product catalog.
Returns full catalog details for inventory-tracked products, with stock filters and
the existing pagination format. `organization_id` is required; the stock filters
are optional and can be combined.

| Query parameter | Default | Meaning |
| --- | --- | --- |
| `organization_id` | Required | Positive organization ID |
| `low_stock_threshold` | `20` | Non-negative number, including decimals |
| `stock_status` | `low` | `low`, `non_low`, or `all` |
| `category_id` | `0` | Filter category; omitted or zero means all |
| `store_id` | `0` | Filter store; omitted or zero means all |
| `storage_location_id` | `0` | Filter location; omitted or zero means all |
| `limit` | `100` | Positive page size |
| `offset` | `0` | Non-negative number of products to skip |

Stock means the sum of `inventory_stock.quantity_available`, including all variants,
within the requested organization/store/location scope. A product with quantities
5 and 7 in two locations in a store has a store total of 12. Stock in another store
does not contribute when `store_id` is provided.

`low` selects totals strictly below the threshold; `non_low` selects totals at or
above it. `all` returns both groups. Each product includes `available_stock` and
`is_low_stock`, along with the original master catalog fields. The nested
`inventory` contains only records matching the requested store and location,
including all its locations when only a store is supplied.

Without store/location filters, inventory-tracked products without inventory rows
have zero stock and an empty inventory array. With a store/location filter, only
products with an inventory row in that scope are returned. Null quantities count
as zero. Products with `track_inventory` disabled are excluded.

Examples:

```text
# Low stock across the organization
/api/products/master-catalog/stock?organization_id=1&low_stock_threshold=20

# Low stock in a category and store, across all the store's locations
/api/products/master-catalog/stock?organization_id=1&category_id=2&store_id=3&low_stock_threshold=20

# Low stock in a specific storage location
/api/products/master-catalog/stock?organization_id=1&store_id=3&storage_location_id=4&low_stock_threshold=20

# Products at or above the threshold
/api/products/master-catalog/stock?organization_id=1&stock_status=non_low&low_stock_threshold=20

# Both classifications, second page
/api/products/master-catalog/stock?organization_id=1&stock_status=all&limit=20&offset=20
```

Response envelope:

```json
{
  "statusCode": 200,
  "message": "master product stock catalog fetched successfully",
  "data": {
    "total_count": 42,
    "total_pages": 3,
    "page": 2,
    "limit": 20,
    "data": []
  }
}
```

`data.data` contains the products for the requested page and is an empty array
when no products match or the offset exceeds the result count. Counts use the
same filters as the products. No database migration is required for this endpoint.

The PostgreSQL integration test uses `NEMBUS_STOCK_TEST_DATABASE_URL` pointing to
a dedicated database initialized with `db/schema` and `db/views_functions`.
Run it with `go test ./repository -run TestMasterProductStockCatalogPostgres -v`
from `packages/core`. Fixtures run in a transaction and are rolled back.
