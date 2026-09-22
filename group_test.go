package signalflow

import "testing"

func TestGroupBuildsTimeSeriesSummary(t *testing.T) {
	result := (&moduleInstance{}).group([]StreamMessage{
		{
			Type:        "data",
			TimestampMs: 1000,
			Datapoints: []DataPoint{{
				TSID:  "tsid-1",
				Value: 42,
				Type:  "double",
			}},
		},
		{
			Type: "metadata",
			Metadata: &Metadata{
				TSID:   "tsid-1",
				Metric: "cpu.utilization",
				Custom: map[string]string{"host": "example"},
			},
		},
	})

	if len(result.TimeSeriesIds) != 1 || result.TimeSeriesIds[0] != "tsid-1" {
		t.Fatalf("unexpected TSIDs: %#v", result.TimeSeriesIds)
	}
	if result.TimeSeriesCount != 1 || result.Cardinality != 1 {
		t.Fatalf("unexpected result counts: %d/%d", result.TimeSeriesCount, result.Cardinality)
	}
	if len(result.DimensionsAndValues["host"]) != 1 || result.DimensionsAndValues["host"][0] != "example" {
		t.Fatalf("unexpected dimension values: %#v", result.DimensionsAndValues)
	}
	summary := result.GetTimeSeriesDataById("tsid-1")
	if summary.Metric != "cpu.utilization" {
		t.Fatalf("expected metric, got %q", summary.Metric)
	}
	if summary.Dimensions["host"] != "example" {
		t.Fatalf("expected host dimension")
	}
	if summary.DatapointCount != 1 || len(summary.Datapoints) != 1 || summary.Datapoints[0].Timestamp != 1000 {
		t.Fatalf("unexpected grouped datapoints: %#v", summary.Datapoints)
	}
}
