package signalflow

import (
	"testing"

	"github.com/signalfx/signalflow-client-go/v2/signalflow/messages"
	"github.com/signalfx/signalfx-go/idtool"
)

func TestWrapDataMessage(t *testing.T) {
	message := &messages.DataMessage{
		BaseChannelMessage: messages.BaseChannelMessage{Chan: "channel-1"},
		TimestampedMessage: messages.TimestampedMessage{TimestampMillis: 1234},
		Payloads: []messages.DataPayload{{
			Type: messages.ValTypeDouble,
			TSID: idtool.IDFromString("AAAAAAAAAAA"),
		}},
	}

	wrapped := wrapDataMessage(message)
	if wrapped.Type != messages.DataType {
		t.Fatalf("expected data type, got %q", wrapped.Type)
	}
	if wrapped.Channel != "channel-1" {
		t.Fatalf("expected channel-1, got %q", wrapped.Channel)
	}
	if wrapped.DatapointCount != 1 {
		t.Fatalf("expected one datapoint, got %d", wrapped.DatapointCount)
	}
}
