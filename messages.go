package signalflow

import (
	"context"
	"fmt"
	"time"

	"github.com/signalfx/signalflow-client-go/v2/signalflow/messages"
	"github.com/signalfx/signalfx-go/idtool"
)

// StreamMessage is the JavaScript-safe representation of a SignalFlow
// message. It intentionally contains plain Go values only; k6 can export
// these without exposing the underlying client library types or channels.
type StreamMessage struct {
	Type               string                 `js:"type"`
	Channel            string                 `js:"channel"`
	TimestampMs        uint64                 `js:"timestampMs"`
	LogicalTimestampMs uint64                 `js:"logicalTimestampMs"`
	TSID               string                 `js:"tsid"`
	DatapointCount     int                    `js:"datapointCount"`
	Datapoints         []DataPoint            `js:"datapoints"`
	Event              string                 `js:"event"`
	MessageCode        string                 `js:"messageCode"`
	MessageLevel       string                 `js:"messageLevel"`
	Contents           map[string]interface{} `js:"contents"`
	Raw                map[string]interface{} `js:"raw"`
	Metadata           *Metadata              `js:"metadata"`
}

type DataPoint struct {
	TSID  string      `js:"tsid"`
	Value interface{} `js:"value"`
	Type  string      `js:"type"`
}

type Metadata struct {
	TSID              string                 `js:"tsid"`
	Metric            string                 `js:"metric"`
	OriginatingMetric string                 `js:"originatingMetric"`
	ResolutionMs      int                    `js:"resolutionMs"`
	CreatedOnMs       int                    `js:"createdOnMs"`
	Internal          map[string]interface{} `js:"internal"`
	Custom            map[string]string      `js:"custom"`
}

// Next waits for the next normalized metadata, data, info, event, or
// expired-timeseries message. A timeout returns null to JavaScript without
// treating it as an error.
func (c *computationHandle) Next(timeoutMs int64) (*StreamMessage, error) {
	if timeoutMs <= 0 {
		timeoutMs = 1000
	}

	timer := time.NewTimer(time.Duration(timeoutMs) * time.Millisecond)
	defer timer.Stop()
	if len(c.pendingMessages) > 0 {
		message := c.pendingMessages[0]
		c.pendingMessages = c.pendingMessages[1:]
		return message, nil
	}

	dataCh := c.computation.Data()
	infoCh := c.computation.Info()
	eventCh := c.computation.Events()
	expirationCh := c.computation.Expirations()

	for dataCh != nil || infoCh != nil || eventCh != nil || expirationCh != nil {
		select {
		case message, ok := <-dataCh:
			if !ok {
				dataCh = nil
				continue
			}
			c.queueMetadataMessages(message)
			c.pendingMessages = append(c.pendingMessages, wrapDataMessage(message))
			queued := c.pendingMessages[0]
			c.pendingMessages = c.pendingMessages[1:]
			return queued, nil
		case message, ok := <-infoCh:
			if !ok {
				infoCh = nil
				continue
			}
			return wrapInfoMessage(message), nil
		case message, ok := <-eventCh:
			if !ok {
				eventCh = nil
				continue
			}
			return wrapEventMessage(message), nil
		case message, ok := <-expirationCh:
			if !ok {
				expirationCh = nil
				continue
			}
			return wrapExpirationMessage(message), nil
		case <-timer.C:
			return nil, nil
		}
	}

	if err := c.computation.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (c *computationHandle) queueMetadataMessages(message *messages.DataMessage) {
	for _, payload := range message.Payloads {
		tsid := payload.TSID.String()
		if tsid == "" || c.emittedMetadata[tsid] {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		metadata, err := c.metadataWithContext(tsid, ctx)
		cancel()
		if err != nil {
			continue
		}

		c.emittedMetadata[tsid] = true
		c.pendingMessages = append(c.pendingMessages, &StreamMessage{
			Type:     messages.MetadataType,
			TSID:     tsid,
			Metadata: metadata,
		})
	}
}

// Metadata retrieves the metadata associated with a timeseries ID observed in
// a data message. The underlying SignalFlow client caches metadata messages,
// so this also works when metadata arrived before the caller requested it.
func (c *computationHandle) Metadata(tsid string, timeoutMs int64) (*Metadata, error) {
	if tsid == "" {
		return nil, fmt.Errorf("timeseries ID is required")
	}
	if timeoutMs <= 0 {
		timeoutMs = 1000
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()
	return c.metadataWithContext(tsid, ctx)
}

func (c *computationHandle) metadataWithContext(tsid string, ctx context.Context) (*Metadata, error) {
	properties, err := c.computation.TSIDMetadata(ctx, idtool.IDFromString(tsid))
	if err != nil {
		return nil, err
	}

	return &Metadata{
		TSID:              tsid,
		Metric:            properties.Metric,
		OriginatingMetric: properties.OriginatingMetric,
		ResolutionMs:      properties.ResolutionMS,
		CreatedOnMs:       properties.CreatedOnMS,
		Internal:          properties.InternalProperties,
		Custom:            properties.CustomProperties,
	}, nil
}

func wrapDataMessage(message *messages.DataMessage) *StreamMessage {
	datapoints := make([]DataPoint, 0, len(message.Payloads))
	for _, payload := range message.Payloads {
		datapoints = append(datapoints, DataPoint{
			TSID:  payload.TSID.String(),
			Value: payload.Value(),
			Type:  payload.Type.String(),
		})
	}

	return &StreamMessage{
		Type:           messages.DataType,
		Channel:        message.Channel(),
		TimestampMs:    message.TimestampMillis,
		DatapointCount: len(datapoints),
		Datapoints:     datapoints,
	}
}

func wrapInfoMessage(message *messages.InfoMessage) *StreamMessage {
	return &StreamMessage{
		Type:               messages.MessageType,
		Channel:            message.Channel(),
		LogicalTimestampMs: message.LogicalTimestampMillis,
		MessageCode:        message.MessageBlock.Code,
		MessageLevel:       message.MessageBlock.Level,
		Contents:           message.MessageBlock.ContentsRaw,
	}
}

func wrapEventMessage(message *messages.EventMessage) *StreamMessage {
	return &StreamMessage{
		Type:    messages.EventType,
		Channel: message.Channel(),
		Raw:     message.JSONBase().RawData(),
	}
}

func wrapExpirationMessage(message *messages.ExpiredTSIDMessage) *StreamMessage {
	return &StreamMessage{
		Type:    messages.ExpiredTSIDType,
		Channel: message.Channel(),
		TSID:    message.TSID,
		Raw:     message.JSONBase().RawData(),
	}
}
