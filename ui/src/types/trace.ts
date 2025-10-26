export type SpanType = 'tool' | 'api_call' | 'agent_handoff' | 'iteration' | 'trace';

export interface Span {
  id: string;
  type: SpanType;
  name: string;
  start_time: string;
  end_time?: string;
  duration?: string;
  duration_ms?: number;
  parent_span_id?: string;
  children?: Span[];
  tags?: Record<string, unknown>;
  error?: boolean;
  error_msg?: string;
}

export interface Trace {
  trace_id: string;
  root_span: Span;
  start_time: string;
  end_time?: string;
  duration?: string;
  duration_ms?: number;
}
