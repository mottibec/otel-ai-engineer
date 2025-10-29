import type { AgentWork } from "./agent-work";

export interface Collector {
  collector_id: string;
  collector_name: string;
  target_type: string;
  status: string;
  deployed_at: string;
  config_path?: string;
  agent_work?: AgentWork[];
}

export interface ConnectedAgent {
  id: string;
  name: string;
  status: string;
  version: string;
  last_seen?: string;
  group_id?: string;
  group_name?: string;
  description?: string;
  agent_work?: AgentWork[];
}

export interface CollectorConfig {
  config_id: string;
  config_name: string;
  config_version: number;
  yaml_content: string;
  agent_work?: AgentWork[];
}

export interface DeployCollectorRequest {
  collector_name: string;
  target_type?: string;
  yaml_config: string;
  parameters?: Record<string, unknown>;
}

export interface UpdateCollectorConfigRequest {
  yaml_config: string;
}

