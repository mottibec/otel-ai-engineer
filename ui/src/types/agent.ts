export interface ToolInfo {
  name: string;
  description: string;
  schema: any;
  category: string;
}

export interface Agent {
  id: string;
  name: string;
  description: string;
  model?: string;
  system_prompt?: string;
  max_tokens?: number;
  type: "built-in" | "custom";
  tools?: ToolInfo[];
  tool_names?: string[];
  created_at?: string;
  updated_at?: string;
}

export interface ListToolsResponse {
  tools: ToolInfo[];
  by_category: Record<string, ToolInfo[]>;
}

export interface CreateCustomAgentRequest {
  name: string;
  description: string;
  system_prompt?: string;
  model?: string;
  max_tokens?: number;
  tool_names: string[];
}

export interface UpdateCustomAgentRequest {
  name?: string;
  description?: string;
  system_prompt?: string;
  model?: string;
  max_tokens?: number;
  tool_names?: string[];
}

export interface CreateMetaAgentRequest {
  name: string;
  description: string;
  system_prompt?: string;
  model?: string;
  available_tool_names: string[];
}

