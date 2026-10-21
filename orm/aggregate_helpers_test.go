package orm

import "testing"

func TestStdDevAggregateHelperBuildsStandardDeviation(t *testing.T) {
	agg := StdDev("price")
	if agg.Name != "stddev" || agg.Field != "price" || agg.Func != string(AggStdDev) {
		t.Fatalf("unexpected aggregate: %+v", agg)
	}
}

func TestVarianceAggregateHelperBuildsVariance(t *testing.T) {
	agg := Variance("price")
	if agg.Name != "variance" || agg.Field != "price" || agg.Func != string(AggVariance) {
		t.Fatalf("unexpected aggregate: %+v", agg)
	}
}
