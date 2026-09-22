# xk6-signalflow

An experimental k6 JavaScript extension backed by the Splunk Observability Cloud SignalFlow Go client.

## Development

Install xk6:

```bash
go install go.k6.io/xk6/cmd/xk6@latest
```

Build and run the example from this repository:

```bash
xk6 run examples/smoke.js
```

Or build a reusable binary:

```bash
xk6 build \
  --with github.com/your-org/xk6-signalflow=. \
  --output ./k6

./k6 run examples/smoke.js
```

Set credentials through the environment; do not put tokens in scripts or images:

```bash
SFX_REALM=us0 \
SFX_ACCESS_TOKEN=... \
xk6 run examples/smoke.js
```

## Current scope

The extension supports client creation, computation execution, controlled shutdown,
one-hour (or arbitrary) lookback windows, and typed access to the messages needed to
summarize a stream. `execute()` has a 30-second default timeout; pass `timeout` and
`lookback` in milliseconds to override the request window.

The computation handle exposes:

```js
const message = computation.next(2000); // null on timeout
const metadata = computation.metadata(message.datapoints[0].tsid, 1000);

const messages = [];
// collect messages with computation.next(...)
const resultMap = signalflow.group(messages);
const tsid = resultMap.listTimeSeriesIds()[0];
const series = resultMap.getTimeSeriesDataById(tsid);
series.metric;
series.dimensions;
series.getDatapoints();
series.getTimeSeriesDatapointsCount();
```

Messages are returned as plain JavaScript objects with `type`, channel/timestamp
fields, and type-specific data. Data messages include `datapointCount` and a
`datapoints` array; metadata, info, event, and expired-timeseries messages are
also exposed. `signalflow.group(messages)` organizes the returned message set by
TSID. Each `TimeSeriesSummary` exposes `metric`, `dimensions`, `getDatapoints()`,
and `getTimeSeriesDatapointsCount()`. Metadata dimensions are the non-`sf_`
properties from the SignalFlow metadata record.
