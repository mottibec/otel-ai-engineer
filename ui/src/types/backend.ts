import type { AgentWork } from "./agent-work";

export interface Backend {
  id: string;
  plan_id?: string; // Optional for standalone backends
  backend_type: string; // "grafana", "prometheus", "jaeger", "custom"
  name: string;
  url: string;
  credentials: string;
  health_status: string; // "healthy", "unhealthy", "unknown"
  last_check?: string;
  datasource_uid?: string;
  config?: string;
  created_at: string;
  updated_at: string;
  agent_work?: AgentWork[];
}

export interface CreateBackendRequest {
  backend_type: string;
  name: string;
  url: string;
  username?: string;
  password?: string;
  credentials?: string;
  config?: Record<string, unknown>;
  plan_id?: string;
}

export interface UpdateBackendRequest {
  name?: string;
  url?: string;
  username?: string;
  password?: string;
  credentials?: string;
  config?: Record<string, unknown>;
  health_status?: string;
}

export interface TestConnectionRequest {
  url?: string;
  username?: string;
  password?: string;
}

export interface TestConnectionResponse {
  healthy: boolean;
  status: string;
  error?: string;
  datasources?: Datasource[];
}

export interface Datasource {
  id: number;
  uid: string;
  name: string;
  type: string;
  url: string;
}

export interface ConfigureGrafanaDatasourceRequest {
  datasource_name: string;
  datasource_type: string; // "otlp", "prometheus", "loki", "tempo"
  url: string;
  json_data?: Record<string, unknown>;
}

