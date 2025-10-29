export interface Sandbox {
  id: string;
  name: string;
  description: string;
  collector_config: string;
  collector_version: string;
  status: SandboxStatus;
  created_at: string;
  updated_at: string;
  collector_container_id?: string;
  collector_container_name?: string;
  network_id?: string;
  network_name?: string;
  telemetry_config: TelemetryConfig;
  last_validation?: ValidationResult;
  tags?: Record<string, string>;
  metadata?: Record<string, any>;
}

export type SandboxStatus =
  | "creating"
  | "running"
  | "stopped"
  | "failed"
  | "validating";

export interface TelemetryConfig {
  generate_traces: boolean;
  generate_metrics: boolean;
  generate_logs: boolean;
  trace_rate: number;
  metric_rate: number;
  log_rate: number;
  trace_attributes?: Record<string, string>;
  metric_types?: string[];
  log_severity?: string[];
  otlp_endpoint: string;
  otlp_protocol: string;
  trace_duration?: number;
}

export interface ValidationResult {
  id: string;
  sandbox_id: string;
  status: ValidationStatus;
  started_at: string;
  completed_at: string;
  duration: number;
  checks: ValidationCheck[];
  summary: ValidationSummary;
  collector_logs?: LogEntry[];
  collector_metrics?: CollectorMetrics;
  issues?: ValidationIssue[];
  ai_analysis?: string;
  recommendations?: string[];
}

export type ValidationStatus =
  | "pending"
  | "running"
  | "passed"
  | "failed"
  | "partial";

export interface ValidationCheck {
  name: string;
  category: string;
  status: string;
  message: string;
  details?: string;
  severity: string;
  timestamp: string;
}

export interface ValidationSummary {
  total_checks: number;
  passed: number;
  failed: number;
  warnings: number;
  critical: number;
  traces_received: number;
  metrics_received: number;
  logs_received: number;
  traces_exported: number;
  metrics_exported: number;
  logs_exported: number;
  data_loss_percent: number;
}

export interface ValidationIssue {
  type: string;
  severity: string;
  component: string;
  message: string;
  description: string;
  suggestion?: string;
  timestamp: string;
}

export interface LogEntry {
  timestamp: string;
  level: string;
  message: string;
  fields?: Record<string, string>;
}

export interface CollectorMetrics {
  receiver_accepted_spans: number;
  receiver_refused_spans: number;
  receiver_accepted_metrics: number;
  receiver_refused_metrics: number;
  receiver_accepted_logs: number;
  receiver_refused_logs: number;
  processor_accepted_spans: number;
  processor_refused_spans: number;
  processor_dropped_spans: number;
  processor_accepted_metrics: number;
  processor_refused_metrics: number;
  processor_dropped_metrics: number;
  exporter_sent_spans: number;
  exporter_failed_spans: number;
  exporter_sent_metrics: number;
  exporter_failed_metrics: number;
  exporter_sent_logs: number;
  exporter_failed_logs: number;
  queue_size: number;
  queue_capacity: number;
  memory_usage_mb: number;
  cpu_usage_percent: number;
}

export interface CreateSandboxRequest {
  name: string;
  description?: string;
  collector_config: string;
  collector_version?: string;
  telemetry_config?: Partial<TelemetryConfig>;
  tags?: Record<string, string>;
}

export interface StartTelemetryRequest {
  duration?: number;
  telemetry_config?: Partial<TelemetryConfig>;
  auto_validate?: boolean;
}

export interface ValidateSandboxRequest {
  run_checks?: string[];
  collect_logs?: boolean;
  collect_metrics?: boolean;
  ai_analysis?: boolean;
}
