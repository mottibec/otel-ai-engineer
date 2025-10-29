export type PlanStatus = "draft" | "pending" | "executing" | "partial" | "success" | "failed";

export interface ObservabilityPlan {
  id: string;
  name: string;
  description: string;
  environment: string;
  status: PlanStatus;
  created_at: string;
  updated_at: string;
  services?: InstrumentedService[];
  infrastructure?: InfrastructureComponent[];
  pipelines?: CollectorPipeline[];
  backends?: Backend[];
  dependencies?: PlanDependency[];
}

export interface InstrumentedService {
  id: string;
  plan_id: string;
  service_name: string;
  language: string;
  framework: string;
  sdk_version: string;
  config_file: string;
  status: string;
  code_changes_summary: string;
  target_path: string;
  exporter_endpoint: string;
  git_repo_url?: string;
  created_at: string;
  updated_at: string;
}

export interface InfrastructureComponent {
  id: string;
  plan_id: string;
  component_type: string;
  name: string;
  host: string;
  receiver_type: string;
  metrics_collected: string;
  status: string;
  config?: string;
  created_at: string;
  updated_at: string;
}

export interface CollectorPipeline {
  id: string;
  plan_id: string;
  collector_id: string;
  name: string;
  config_yaml: string;
  rules: string;
  status: string;
  target_type: string;
  created_at: string;
  updated_at: string;
}

export interface Backend {
  id: string;
  plan_id: string;
  backend_type: string;
  name: string;
  url: string;
  credentials: string;
  health_status: string;
  last_check?: string;
  datasource_uid?: string;
  config?: string;
  created_at: string;
  updated_at: string;
}

export interface PlanDependency {
  id: string;
  plan_id: string;
  source_id: string;
  source_type: string;
  target_id: string;
  target_type: string;
  dependency_type: string;
  created_at: string;
}

export interface TopologyNode {
  id: string;
  type: string;
  label: string;
  status: string;
  metadata?: Record<string, any>;
}

export interface TopologyEdge {
  source_id: string;
  target_id: string;
  type: string;
}

export interface TopologyGraph {
  nodes: TopologyNode[];
  edges: TopologyEdge[];
}

