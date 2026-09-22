package signalflow

import (
	"context"
	"fmt"
	"time"

	client "github.com/signalfx/signalflow-client-go/v2/signalflow"
	"go.k6.io/k6/v2/js/modules"
)

var (
	_ modules.Module   = (*rootModule)(nil)
	_ modules.Instance = (*moduleInstance)(nil)
)

type rootModule struct{}

func (*rootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &moduleInstance{vu: vu}
}

type moduleInstance struct {
	vu modules.VU
}

func (m *moduleInstance) Exports() modules.Exports {
	return modules.Exports{
		Default: map[string]any{
			"client": m.newClient,
			"group":  m.group,
		},
	}
}

type clientOptions struct {
	Realm string `js:"realm"`
	Token string `js:"token"`
}

func (m *moduleInstance) newClient(options clientOptions) (*clientHandle, error) {
	if options.Token == "" {
		return nil, fmt.Errorf("signalflow token is required")
	}

	flowClient, err := client.NewClient(
		client.StreamURLForRealm(options.Realm),
		client.AccessToken(options.Token),
		client.UserAgent("xk6-signalflow"),
	)
	if err != nil {
		return nil, err
	}

	return &clientHandle{client: flowClient}, nil
}

type executeOptions struct {
	Program    string `js:"program"`
	Resolution int64  `js:"resolution"`
	Immediate  bool   `js:"immediate"`
	Timeout    int64  `js:"timeout"`
	Lookback   int64  `js:"lookback"`
}

type clientHandle struct {
	client *client.Client
}

func (c *clientHandle) Execute(options executeOptions) (*computationHandle, error) {
	if options.Program == "" {
		return nil, fmt.Errorf("signalflow program is required")
	}

	request := &client.ExecuteRequest{
		Program:   options.Program,
		Immediate: options.Immediate,
	}
	if options.Resolution > 0 {
		request.Resolution = time.Duration(options.Resolution) * time.Millisecond
	}
	if options.Lookback > 0 {
		now := time.Now()
		request.Start = now.Add(-time.Duration(options.Lookback) * time.Millisecond)
		request.Stop = now
	}

	timeout := 30 * time.Second
	if options.Timeout > 0 {
		timeout = time.Duration(options.Timeout) * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	computation, err := c.client.Execute(ctx, request)
	if err != nil {
		return nil, err
	}

	return &computationHandle{
		computation:     computation,
		pendingMessages: make([]*StreamMessage, 0),
		emittedMetadata: make(map[string]bool),
	}, nil
}

func (c *clientHandle) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

type computationHandle struct {
	computation     *client.Computation
	pendingMessages []*StreamMessage
	emittedMetadata map[string]bool
}

func (c *computationHandle) Stop() error {
	return c.computation.Stop(context.Background())
}
