package signalflow

import "sort"

// TimeSeriesDataPoint is a datapoint separated from its original data message.
type TimeSeriesDataPoint struct {
	TSID        string      `js:"tsid"`
	TimestampMs uint64      `js:"timestampMs"`
	Value       interface{} `js:"value"`
	Type        string      `js:"type"`
}

// ComputationResultMap organizes a returned message set by timeseries ID.
type ComputationResultMap struct {
	series map[string]*TimeSeriesSummary
}

// TimeSeriesSummary contains the datapoints and metadata for one timeseries.
type TimeSeriesSummary struct {
	TSID       string            `js:"tsid"`
	Metric     string            `js:"metric"`
	Dimensions map[string]string `js:"dimensions"`
	data       []TimeSeriesDataPoint
	metadata   *Metadata
}

// group is intentionally a module-level function. It does not depend on a
// computation handle and only organizes the messages supplied by the caller.
func (m *moduleInstance) group(messages []StreamMessage) *ComputationResultMap {
	result := &ComputationResultMap{series: make(map[string]*TimeSeriesSummary)}

	for _, message := range messages {
		switch message.Type {
		case "data":
			for _, datapoint := range message.Datapoints {
				summary := result.summary(datapoint.TSID)
				summary.data = append(summary.data, TimeSeriesDataPoint{
					TSID:        datapoint.TSID,
					TimestampMs: message.TimestampMs,
					Value:       datapoint.Value,
					Type:        datapoint.Type,
				})
			}
		case "metadata":
			if message.Metadata != nil {
				summary := result.summary(message.Metadata.TSID)
				summary.metadata = message.Metadata
				summary.Metric = message.Metadata.Metric
				summary.Dimensions = message.Metadata.Custom
			}
		}
	}

	return result
}

func (m *ComputationResultMap) summary(tsid string) *TimeSeriesSummary {
	if summary := m.series[tsid]; summary != nil {
		return summary
	}

	summary := &TimeSeriesSummary{
		TSID:       tsid,
		Dimensions: make(map[string]string),
	}
	m.series[tsid] = summary
	return summary
}

// ListTimeSeriesIds returns all TSIDs in deterministic order.
func (m *ComputationResultMap) ListTimeSeriesIds() []string {
	ids := make([]string, 0, len(m.series))
	for tsid := range m.series {
		ids = append(ids, tsid)
	}
	sort.Strings(ids)
	return ids
}

// TimeSeriesCount returns the number of unique TSIDs.
func (m *ComputationResultMap) TimeSeriesCount() int {
	return len(m.series)
}

// GetTimeSeriesDataById returns the summary for a TSID.
func (m *ComputationResultMap) GetTimeSeriesDataById(tsid string) *TimeSeriesSummary {
	return m.series[tsid]
}

// GetTimeSeriesMetadataById returns metadata for a TSID.
func (m *ComputationResultMap) GetTimeSeriesMetadataById(tsid string) *Metadata {
	if summary := m.series[tsid]; summary != nil {
		return summary.metadata
	}
	return nil
}

// GetTimeSeriesDatapointsCount returns the datapoint count for a TSID.
func (m *ComputationResultMap) GetTimeSeriesDatapointsCount(tsid string) int {
	if summary := m.series[tsid]; summary != nil {
		return len(summary.data)
	}
	return 0
}

// GetDatapoints returns this timeseries' datapoints in stream order.
func (s *TimeSeriesSummary) GetDatapoints() []TimeSeriesDataPoint {
	return s.data
}

// GetTimeSeriesMetadataById returns this summary's metadata. The summary is
// already bound to one TSID, so the optional argument is ignored.
func (s *TimeSeriesSummary) GetTimeSeriesMetadataById(_ ...string) *Metadata {
	return s.metadata
}

// GetTimeSeriesDatapointsCount returns this summary's datapoint count.
func (s *TimeSeriesSummary) GetTimeSeriesDatapointsCount(_ ...string) int {
	return len(s.data)
}
