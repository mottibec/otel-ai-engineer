// Event types matching the Go backend
export type EventType =
  | 'run_start'
  | 'run_end'
  | 'iteration'
  | 'message'
  | 'tool_call'
  | 'tool_result'
  | 'file_change'
  | 'error'
  | 'api_request'
  | 'api_response';

export interface AgentEvent {
  id: string;
  timestamp: string;
  agent_id: string;
  agent_name: string;
  run_id: string;
  type: EventType;
  data: unknown;
}

export interface RunStartData {
  prompt: string;
  model: string;
  max_tokens: number;
  system_prompt?: string;
}

export interface RunEndData {
  success: boolean;
  error?: string;
  total_tool_calls: number;
  total_iterations: number;
  duration: string;
}

export interface IterationData {
  iteration: number;
  total_messages: number;
}

export interface UsageInfo {
  input_tokens: number;
  output_tokens: number;
}

export interface ContentBlock {
  type: 'text' | 'tool_use' | 'tool_result';
  text?: string;
  tool_use?: {
    id: string;
    name: string;
    input: unknown;
  };
  tool_result?: {
    tool_use_id: string;
    content: string;
    is_error: boolean;
  };
}

export interface MessageData {
  role: string;
  content: ContentBlock[];
  stop_reason?: string;
  model?: string;
  usage?: UsageInfo;
}

export interface ToolCallData {
  tool_use_id: string;
  tool_name: string;
  input: unknown;
}

export interface ToolResultData {
  tool_use_id: string;
  tool_name: string;
  result?: unknown;
  error?: string;
  is_error: boolean;
  duration: string;
}

export interface APIRequestData {
  model: string;
  max_tokens: number;
  tool_count: number;
}

export interface APIResponseData {
  stop_reason: string;
  model: string;
  usage?: UsageInfo;
  content_count: number;
}

export interface ErrorData {
  message: string;
  stack_trace?: string;
}
