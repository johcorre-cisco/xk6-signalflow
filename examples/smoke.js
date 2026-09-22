import signalflow from 'k6/x/signalflow';

export const options = {
  vus: 1,
  iterations: 1,
};

export default function () {
  const client = signalflow.client({
    realm: __ENV.SFX_REALM || 'us0',
    token: __ENV.SFX_ACCESS_TOKEN,
  });

  const computation = client.execute({
    program: "data('sf.org.num.orguser').publish()",
    resolution: 10000,
    immediate: true,
    timeout: 15000,
    lookback: 60 * 60 * 1000,
  });

  const summary = {
    messages: 0,
    dataMessages: 0,
    dataMessageSizes: [],
    datapoints: 0,
    infoMessages: 0,
    eventMessages: 0,
    expiredTimeseries: 0,
    metadata: 0,
    tsids: new Set(),
    sampleDatapoints: [],
    sampleMetadata: [],
  };
  const messages = [];

  const deadline = Date.now() + 10000;
  while (Date.now() < deadline) {
    const message = computation.next(Math.min(2000, deadline - Date.now()));
    if (message === null) {
      break;
    }

    summary.messages += 1;
    messages.push(message);
    if (message.type === 'data') {
      summary.dataMessages += 1;
      summary.dataMessageSizes.push(message.datapointCount || 0);
      summary.datapoints += message.datapointCount || 0;
      if (summary.sampleDatapoints.length < 5) {
        summary.sampleDatapoints.push(...(message.datapoints || []).slice(0, 5 - summary.sampleDatapoints.length));
      }
      for (const datapoint of message.datapoints || []) {
        summary.tsids.add(datapoint.tsid);
      }
    } else if (message.type === 'metadata') {
      summary.metadata += 1;
      if (summary.sampleMetadata.length < 3 && message.metadata !== null) {
        summary.sampleMetadata.push(message.metadata);
      }
    } else if (message.type === 'message') {
      summary.infoMessages += 1;
    } else if (message.type === 'event') {
      summary.eventMessages += 1;
    } else if (message.type === 'expired-tsid') {
      summary.expiredTimeseries += 1;
    }
  }

  const resultMap = signalflow.group(messages);
  const groupedIds = resultMap.listTimeSeriesIds();
  const firstSeries = groupedIds.length > 0 ? resultMap.getTimeSeriesDataById(groupedIds[0]) : null;

  const { tsids, ...printableSummary } = summary;
  console.log(JSON.stringify({
    ...printableSummary,
    uniqueTimeseries: summary.tsids.size,
    sampleTsids: [...summary.tsids].slice(0, 10),
    groupedTimeSeries: resultMap.timeSeriesCount(),
    firstSeries: firstSeries === null ? null : {
      tsid: groupedIds[0],
      metric: firstSeries.metric,
      dimensions: firstSeries.dimensions,
      datapoints: firstSeries.getTimeSeriesDatapointsCount(),
      sampleDatapoints: firstSeries.getDatapoints().slice(0, 3),
    },
  }));

  computation.stop();
  client.close();
}
