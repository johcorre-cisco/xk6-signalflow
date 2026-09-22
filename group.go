package signalflow

import "sort"

// TimeSeriesDataPoint is a datapoint separated from its original data message.
type TimeSeriesDataPoint struct {
	Timestamp uint64      `js:"timestamp"`
	Value     interface{} `js:"value"`
	Type      string      `js:"type"`
}

// ComputationResultMap organizes a returned message set by timeseries ID.
type ComputationResultMap struct {
	series              map[string]*TimeSeriesSummary
	TimeSeriesIds       []string            `js:"timeSeriesIds"`
	TimeSeriesCount     int                 `js:"timeSeriesCount"`
	Cardinality         int                 `js:"cardinality"`
	DimensionsAndValues map[string][]string `js:"dimensionsAndValues"`
}

// TimeSeriesSummary contains the datapoints and metadata for one timeseries.
type TimeSeriesSummary struct {
	TSID           string                `js:"tsid"`
	Metric         string                `js:"metric"`
	Dimensions     map[string]string     `js:"dimensions"`
	Datapoints     []TimeSeriesDataPoint `js:"datapoints"`
	Metadata       *Metadata             `js:"metadata"`
	DatapointCount int                   `js:"datapointCount"`
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
				summary.Datapoints = append(summary.Datapoints, TimeSeriesDataPoint{
					Timestamp: message.TimestampMs,
					Value:     datapoint.Value,
					Type:      datapoint.Type,
				})
			}
		case "metadata":
			if message.Metadata != nil {
				summary := result.summary(message.Metadata.TSID)
				summary.Metadata = message.Metadata
				summary.Metric = message.Metadata.Metric
				summary.Dimensions = message.Metadata.Custom
			}
		}
	}
	for _, summary := range result.series {
		sort.SliceStable(summary.Datapoints, func(i, j int) bool {
			return summary.Datapoints[i].Timestamp < summary.Datapoints[j].Timestamp
		})
		summary.DatapointCount = len(summary.Datapoints)
	}
	result.finalizeAggregateProperties()

	return result
}

func (m *ComputationResultMap) finalizeAggregateProperties() {
	m.TimeSeriesIds = make([]string, 0, len(m.series))
	m.DimensionsAndValues = make(map[string][]string)
	valuesByDimension := make(map[string]map[string]struct{})

	for tsid, summary := range m.series {
		m.TimeSeriesIds = append(m.TimeSeriesIds, tsid)
		for dimension, value := range summary.Dimensions {
			if valuesByDimension[dimension] == nil {
				valuesByDimension[dimension] = make(map[string]struct{})
			}
			valuesByDimension[dimension][value] = struct{}{}
		}
	}

	sort.Strings(m.TimeSeriesIds)
	m.TimeSeriesCount = len(m.TimeSeriesIds)
	m.Cardinality = m.TimeSeriesCount
	for dimension, values := range valuesByDimension {
		m.DimensionsAndValues[dimension] = make([]string, 0, len(values))
		for value := range values {
			m.DimensionsAndValues[dimension] = append(m.DimensionsAndValues[dimension], value)
		}
		sort.Strings(m.DimensionsAndValues[dimension])
	}
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

// GetTimeSeriesDataById returns the summary for a TSID.
func (m *ComputationResultMap) GetTimeSeriesDataById(tsid string) *TimeSeriesSummary {
	return m.series[tsid]
}
