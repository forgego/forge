package catalog

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers catalog API endpoints.
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	_ = ctx
	_ = database

	base := api.NewBaseSerializer(nil)

	router.Register("categories", &api.ViewSetConfig{
		Model:      &Category{},
		Queryset:   CategoryObjects,
		Serializer: base,
	})

	router.Register("brands", &api.ViewSetConfig{
		Model:      &Brand{},
		Queryset:   BrandObjects,
		Serializer: base,
	})

	router.Register("products", &api.ViewSetConfig{
		Model:      &Product{},
		Queryset:   ProductObjects,
		Serializer: base,
	})

	router.Register("product-variants", &api.ViewSetConfig{
		Model:      &ProductVariant{},
		Queryset:   ProductVariantObjects,
		Serializer: base,
	})

	router.Register("product-images", &api.ViewSetConfig{
		Model:      &ProductImage{},
		Queryset:   ProductImageObjects,
		Serializer: base,
	})

	router.Register("product-attributes", &api.ViewSetConfig{
		Model:      &ProductAttribute{},
		Queryset:   ProductAttributeObjects,
		Serializer: base,
	})

	router.Register("product-attribute-values", &api.ViewSetConfig{
		Model:      &ProductAttributeValue{},
		Queryset:   ProductAttributeValueObjects,
		Serializer: base,
	})

	// GET /catalog/stats - Aggregated catalog statistics
	router.Get("/catalog/stats", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var totalProducts, activeProducts, inStockProducts int64
		var avgPrice, minPrice, maxPrice float64
		var totalCategories, totalBrands int64

		row := database.QueryRowContext(ctx, `
			SELECT
				COUNT(*),
				COALESCE(SUM(CASE WHEN is_active THEN 1 ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN stock_quantity > 0 THEN 1 ELSE 0 END), 0),
				COALESCE(AVG(price), 0.0),
				COALESCE(MIN(price), 0.0),
				COALESCE(MAX(price), 0.0)
			FROM products
		`)
		if err := row.Scan(&totalProducts, &activeProducts, &inStockProducts, &avgPrice, &minPrice, &maxPrice); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_ = database.QueryRowContext(ctx, `SELECT COUNT(*) FROM categories`).Scan(&totalCategories)
		_ = database.QueryRowContext(ctx, `SELECT COUNT(*) FROM brands`).Scan(&totalBrands)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"total_products":    totalProducts,
			"active_products":   activeProducts,
			"in_stock_products": inStockProducts,
			"avg_price":         avgPrice,
			"min_price":         minPrice,
			"max_price":         maxPrice,
			"total_categories":  totalCategories,
			"total_brands":      totalBrands,
		})
	})

	// GET /catalog/search - Faceted search and filtering
	router.Get("/catalog/search", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		categoryID, _ := strconv.ParseInt(r.URL.Query().Get("category_id"), 10, 64)
		brandID, _ := strconv.ParseInt(r.URL.Query().Get("brand_id"), 10, 64)
		minPrice, _ := strconv.ParseFloat(r.URL.Query().Get("min_price"), 64)
		maxPrice, _ := strconv.ParseFloat(r.URL.Query().Get("max_price"), 64)
		inStock := r.URL.Query().Get("in_stock") == "true" || r.URL.Query().Get("in_stock") == "1"

		whereClauses := []string{"1=1"}
		var args []any

		if q != "" {
			whereClauses = append(whereClauses, "(name LIKE ? OR description LIKE ? OR sku LIKE ?)")
			likeTerm := "%" + q + "%"
			args = append(args, likeTerm, likeTerm, likeTerm)
		}
		if categoryID > 0 {
			whereClauses = append(whereClauses, "category_id = ?")
			args = append(args, categoryID)
		}
		if brandID > 0 {
			whereClauses = append(whereClauses, "brand_id = ?")
			args = append(args, brandID)
		}
		if minPrice > 0 {
			whereClauses = append(whereClauses, "price >= ?")
			args = append(args, minPrice)
		}
		if maxPrice > 0 {
			whereClauses = append(whereClauses, "price <= ?")
			args = append(args, maxPrice)
		}
		if inStock {
			whereClauses = append(whereClauses, "stock_quantity > 0")
		}

		query := "SELECT id, name, slug, sku, description, price, stock_quantity, category_id, brand_id, is_active FROM products WHERE " + strings.Join(whereClauses, " AND ") + " ORDER BY id DESC LIMIT 50"

		rows, err := database.QueryContext(ctx, query, args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type ProductResult struct {
			ID            int64   `json:"id"`
			Name          string  `json:"name"`
			Slug          string  `json:"slug"`
			SKU           string  `json:"sku"`
			Description   string  `json:"description"`
			Price         float64 `json:"price"`
			StockQuantity int     `json:"stock_quantity"`
			CategoryID    int64   `json:"category_id"`
			BrandID       int64   `json:"brand_id"`
			IsActive      bool    `json:"is_active"`
		}

		results := make([]ProductResult, 0)
		for rows.Next() {
			var p ProductResult
			if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.SKU, &p.Description, &p.Price, &p.StockQuantity, &p.CategoryID, &p.BrandID, &p.IsActive); err == nil {
				results = append(results, p)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"total":   len(results),
			"query":   q,
			"results": results,
		})
	})
}
