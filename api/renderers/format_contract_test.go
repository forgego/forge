package renderers

import (
	"bytes"
	"encoding/csv"
	"reflect"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestCSVRendererProducesHeaderAndRows(t *testing.T) {
	r := NewCSVRenderer()
	if r.MediaType() != "text/csv" {
		t.Fatal(r.MediaType())
	}
	data := []struct {
		Name  string
		Count int
	}{{"apple", 2}}
	encoded, err := r.Render(data)
	if err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(bytes.NewReader(encoded)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 || len(records[0]) != 2 || records[0][0] != "Name" || records[1][0] != "apple" || records[1][1] != "2" {
		t.Fatalf("records=%q", records)
	}
	var buf bytes.Buffer
	if err := r.RenderToWriter(&buf, data); err != nil {
		t.Fatal(err)
	}
	if buf.String() != string(encoded) {
		t.Fatalf("writer=%q render=%q", buf.String(), encoded)
	}
}

func TestCSVRendererAcceptsStringKeyedMaps(t *testing.T) {
	type column string
	encoded, err := NewCSVRenderer().Render([]map[column]int{{"count": 2}})
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != "count\n2\n" {
		t.Fatalf("CSV=%q", encoded)
	}
}

func TestCSVRendererEscapesFormulaCellsButPreservesNumericValues(t *testing.T) {
	data := []map[string]any{{"value": "=1+1"}, {"value": -5}, {"value": "-5"}}
	encoded, err := NewCSVRenderer().Render(data)
	if err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(bytes.NewReader(encoded)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"value"}, {"'=1+1"}, {"-5"}, {"'-5"}}
	if !reflect.DeepEqual(records, want) {
		t.Fatalf("records=%q, want %q", records, want)
	}
}

func TestCSVRendererUsesJSONFieldNamesAndOmitsIgnoredFields(t *testing.T) {
	data := []struct {
		Name   string `json:"display_name,omitempty"`
		Secret string `json:"-"`
		Count  int
	}{{"apple", "hidden", 2}}
	encoded, err := NewCSVRenderer().Render(data)
	if err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(bytes.NewReader(encoded)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"display_name", "Count"}, {"apple", "2"}}
	if !reflect.DeepEqual(records, want) {
		t.Fatalf("records=%q, want %q", records, want)
	}
}

func TestYAMLRendererProducesYAMLDocument(t *testing.T) {
	r := NewYAMLRenderer()
	if r.MediaType() != "application/x-yaml" {
		t.Fatal(r.MediaType())
	}
	encoded, err := r.Render(map[string]interface{}{"name": "apple", "count": 2})
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]interface{}
	if err := yaml.Unmarshal(encoded, &value); err != nil {
		t.Fatal(err)
	}
	if value["name"] != "apple" || value["count"] != 2 {
		t.Fatalf("value=%v", value)
	}
	if bytes.HasPrefix(bytes.TrimSpace(encoded), []byte("{")) {
		t.Fatalf("JSON returned as YAML: %q", encoded)
	}
	var buf bytes.Buffer
	if err := r.RenderToWriter(&buf, map[string]interface{}{"name": "apple", "count": 2}); err != nil {
		t.Fatal(err)
	}
	if buf.String() != string(encoded) {
		t.Fatalf("writer=%q render=%q", buf.String(), encoded)
	}
}
