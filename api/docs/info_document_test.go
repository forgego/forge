package docs

import (
	"encoding/json"
	"testing"
)

// Generated OpenAPI JSON carries the required info block.
func TestOpenAPIDocumentIncludesInfo(t *testing.T) {
	spec := NewOpenAPIGenerator("Catalog API", "v2").Generate()
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	info, ok := doc["info"].(map[string]any)
	if !ok {
		t.Fatalf("missing info: %s", data)
	}
	if info["title"] != "Catalog API" || info["version"] != "v2" {
		t.Fatalf("info = %v", info)
	}
}
