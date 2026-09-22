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

	if result.TimeSeriesCount() != 1 {
		t.Fatalf("expected one timeseries, got %d", result.TimeSeriesCount())
	}
	ids := result.ListTimeSeriesIds()
	if len(ids) != 1 || ids[0] != "tsid-1" {
		t.Fatalf("unexpected TSIDs: %#v", ids)
	}
	if result.GetTimeSeriesDatapointsCount("tsid-1") != 1 {
		t.Fatalf("expected one datapoint")
	}

	summary := result.GetTimeSeriesDataById("tsid-1")
	if summary.Metric != "cpu.utilization" {
		t.Fatalf("expected metric, got %q", summary.Metric)
	}
	if summary.Dimensions["host"] != "example" {
		t.Fatalf("expected host dimension")
	}
	if len(summary.GetDatapoints()) != 1 || summary.GetDatapoints()[0].TimestampMs != 1000 {
		t.Fatalf("unexpected grouped datapoints: %#v", summary.GetDatapoints())
	}
}
