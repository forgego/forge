package catalog

import (
	"database/sql"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSearchProductsQueryFiltersWithBoundValues(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE products (id INTEGER, name TEXT, slug TEXT, sku TEXT, description TEXT, price REAL, stock_quantity INTEGER, category_id INTEGER, brand_id INTEGER, is_active BOOLEAN);
		INSERT INTO products VALUES (1, 'Red Widget', 'red', 'R1', 'first', 10, 2, 1, 4, 1);
		INSERT INTO products VALUES (2, 'Blue Widget', 'blue', 'B2', 'second', 20, 0, 2, 4, 1);
		INSERT INTO products VALUES (3, 'Green Gizmo', 'green', 'G3', 'third', 30, 3, 2, 5, 1);`)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		args []any
		want []int
	}{
		{"all", []any{"", 0, 0, 0, 0, false}, []int{3, 2, 1}},
		{"search stays data", []any{"Widget' OR 1=1 --", 0, 0, 0, 0, false}, nil},
		{"category and stock", []any{"", 2, 0, 0, 0, true}, []int{3}},
		{"brand and price", []any{"", 0, 4, 15, 25, false}, []int{2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := db.Query(searchProductsQuery, tc.args...)
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			var got []int
			for rows.Next() {
				var id int
				var name, slug, sku, description string
				var price float64
				var stock, category, brand int
				var active bool
				if err := rows.Scan(&id, &name, &slug, &sku, &description, &price, &stock, &category, &brand, &active); err != nil {
					t.Fatal(err)
				}
				got = append(got, id)
			}
			if err := rows.Err(); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
