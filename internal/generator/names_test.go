package generator

import "testing"

func TestNewNames(t *testing.T) {
	tests := []struct {
		input      string
		wantName   string
		wantTable  string
		wantRoute  string
		wantPlural string
	}{
		{"product", "Product", "products", "products", "Products"},
		{"product_category", "ProductCategory", "product_categories", "product-categories", "ProductCategories"},
		{"user", "User", "users", "users", "Users"},
		{"order_item", "OrderItem", "order_items", "order-items", "OrderItems"},
		{"category", "Category", "categories", "categories", "Categories"},
		{"tax", "Tax", "taxes", "taxes", "Taxes"},
		{"address", "Address", "addresses", "addresses", "Addresses"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			n := NewNames(tt.input, "github.com/test/app")
			if n.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", n.Name, tt.wantName)
			}
			if n.TableName != tt.wantTable {
				t.Errorf("TableName = %q, want %q", n.TableName, tt.wantTable)
			}
			if n.RouteName != tt.wantRoute {
				t.Errorf("RouteName = %q, want %q", n.RouteName, tt.wantRoute)
			}
			if n.PluralName != tt.wantPlural {
				t.Errorf("PluralName = %q, want %q", n.PluralName, tt.wantPlural)
			}
		})
	}
}
