export type AgentWorkStatus = "running" | "completed" | "failed" | "cancelled";

export type ResourceType = "collector" | "backend" | "service" | "infrastructure" | "pipeline" | "plan";

export interface AgentWork {
  id: string;
  resource_type: ResourceType;
  resource_id: string;
  run_id: string;
  agent_id: string;
  agent_name: string;
  task_description: string;
  status: AgentWorkStatus;
  started_at: string;
  completed_at?: string;
  error?: string;
  created_at: string;
  updated_at: string;
}

export interface DelegateRequest {
  resource_type: ResourceType;
  resource_id: string;
  agent_id: string;
  task_description: string;
  agent_params?: Record<string, unknown>;
}

export interface DelegateResponse {
  work_id: string;
  run_id: string;
  agent_id: string;
  agent_name: string;
  status: string;
  message: string;
}

