package core

import (
	"encoding/json"
	"testing"

	"github.com/forgego/forge/admin/components"
)

func TestDashboardLayoutSerializesStats(t *testing.T) {
	component := components.Stats("Revenue", "$42", "+5%")
	data, err := json.Marshal(component)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Type  string            `json:"type"`
		Props map[string]string `json:"props"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "stats" || got.Props["label"] != "Revenue" || got.Props["value"] != "$42" || got.Props["change"] != "+5%" {
		t.Fatalf("stats component = %+v", got)
	}
}

func TestDashboardLayoutSerializesLineChart(t *testing.T) {
	component := components.Chart("Orders", "line", []map[string]any{{"name": "Mon", "orders": 4}})
	data, err := json.Marshal(component)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Type  string                     `json:"type"`
		Props map[string]json.RawMessage `json:"props"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "chart" || string(got.Props["chartType"]) != `"line"` || string(got.Props["title"]) != `"Orders"` {
		t.Fatalf("line chart component = %+v", got)
	}
	var points []map[string]any
	if err := json.Unmarshal(got.Props["data"], &points); err != nil || len(points) != 1 || points[0]["orders"] != float64(4) {
		t.Fatalf("line chart data = %v, error = %v", points, err)
	}
}
