import type { AgentEvent } from "./events";

export type RunStatus = "running" | "paused" | "success" | "failed" | "cancelled";

export interface TokenUsage {
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
}

export interface Run {
  id: string;
  agent_id: string;
  agent_name: string;
  status: RunStatus;
  prompt: string;
  model: string;
  start_time: string;
  end_time?: string;
  duration?: string;
  total_iterations: number;
  total_tool_calls: number;
  total_tokens: TokenUsage;
  error?: string;
}

export interface RunWithEvents extends Run {
  events: AgentEvent[];
}

export interface Agent {
  id: string;
  name: string;
  description: string;
  model: string;
}

export interface StartRunRequest {
  agent_id: string;
  prompt: string;
  resume_from_run_id?: string;
}
