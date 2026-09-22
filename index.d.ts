declare module 'k6/x/signalflow' {
  export type SignalFlowValue = unknown;

  export interface ClientOptions {
    realm?: string;
    token: string;
  }

  export interface ExecuteOptions {
    program: string;
    resolution?: number;
    immediate?: boolean;
    timeout?: number;
    lookback?: number;
  }

  export interface DataPoint {
    tsid: string;
    value: SignalFlowValue;
    type: string;
  }

  export interface Metadata {
    tsid: string;
    metric: string;
    originatingMetric: string;
    resolutionMs: number;
    createdOnMs: number;
    internal: Record<string, SignalFlowValue>;
    custom: Record<string, string>;
  }

  export interface BaseMessage {
    type: string;
    channel: string;
    tsid: string;
    timestampMs: number;
    logicalTimestampMs: number;
    datapointCount: number;
    datapoints: DataPoint[];
    event: string;
    messageCode: string;
    messageLevel: string;
    contents: Record<string, SignalFlowValue>;
    raw: Record<string, SignalFlowValue>;
    metadata: Metadata | null;
  }

  export interface DataMessage extends BaseMessage {
    type: 'data';
    datapoints: DataPoint[];
  }

  export interface MetadataMessage extends BaseMessage {
    type: 'metadata';
    tsid: string;
    metadata: Metadata;
  }

  export interface InfoMessage extends BaseMessage {
    type: 'message';
  }

  export interface EventMessage extends BaseMessage {
    type: 'event';
  }

  export interface ExpiredTimeseriesMessage extends BaseMessage {
    type: 'expired-tsid';
    tsid: string;
  }

  export type StreamMessage =
    | DataMessage
    | MetadataMessage
    | InfoMessage
    | EventMessage
    | ExpiredTimeseriesMessage;

  export interface Computation {
    next(timeoutMs?: number): StreamMessage | null;
    metadata(tsid: string, timeoutMs?: number): Metadata;
    stop(): void;
  }

  export interface Client {
    execute(options: ExecuteOptions): Computation;
    close(): void;
  }

  export interface TimeSeriesDataPoint extends DataPoint {
    timestamp: number;
  }

  export interface TimeSeriesSummary {
    readonly tsid: string;
    readonly metric: string;
    readonly dimensions: Record<string, string>;
    readonly datapoints: readonly TimeSeriesDataPoint[];
    readonly metadata: Metadata | null;
    readonly datapointCount: number;
  }

  export interface ComputationResultMap {
    readonly timeSeriesIds: readonly string[];
    readonly timeSeriesCount: number;
    readonly cardinality: number;
    readonly dimensionsAndValues: Readonly<Record<string, readonly string[]>>;
    getTimeSeriesDataById(tsid: string): TimeSeriesSummary | null;
  }

  export interface SignalFlowModule {
    client(options: ClientOptions): Client;
    group(messages: StreamMessage[]): ComputationResultMap;
  }

  const signalflow: SignalFlowModule;
  export default signalflow;
}
