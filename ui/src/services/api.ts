import type { Run, Agent, StartRunRequest } from "../types/models";
import type { AgentEvent } from "../types/events";
import type { Trace } from "../types/trace";
import type {
  ObservabilityPlan,
  TopologyGraph,
  InstrumentedService,
  InfrastructureComponent,
} from "../types/plan";
import type {
  Sandbox,
  CreateSandboxRequest,
  StartTelemetryRequest,
  ValidateSandboxRequest,
  ValidationResult,
  LogEntry,
  CollectorMetrics,
} from "../types/sandbox";
import type {
  Collector,
  ConnectedAgent,
  CollectorConfig,
  DeployCollectorRequest,
  UpdateCollectorConfigRequest,
} from "../types/collector";
import type {
  Backend,
  CreateBackendRequest,
  UpdateBackendRequest,
  TestConnectionRequest,
  TestConnectionResponse,
  ConfigureGrafanaDatasourceRequest,
} from "../types/backend";
import type { AgentWork, DelegateRequest, DelegateResponse } from "../types/agent-work";
import type {
  Agent,
  ToolInfo,
  ListToolsResponse,
  CreateCustomAgentRequest,
  UpdateCustomAgentRequest,
  CreateMetaAgentRequest,
} from "../types/agent";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "/api";

export class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  async listRuns(options?: {
    limit?: number;
    offset?: number;
    status?: string;
  }): Promise<Run[]> {
    const params = new URLSearchParams();
    if (options?.limit) params.append("limit", options.limit.toString());
    if (options?.offset) params.append("offset", options.offset.toString());
    if (options?.status) params.append("status", options.status);

    const response = await fetch(`${this.baseUrl}/runs?${params}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch runs: ${response.statusText}`);
    }
    return response.json();
  }

  async getRun(runId: string): Promise<Run> {
    const response = await fetch(`${this.baseUrl}/runs/${runId}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch run: ${response.statusText}`);
    }
    return response.json();
  }

  async getEvents(runId: string, after?: string): Promise<AgentEvent[]> {
    const params = new URLSearchParams();
    if (after) params.append("after", after);

    const response = await fetch(
      `${this.baseUrl}/runs/${runId}/events?${params}`,
    );
    if (!response.ok) {
      throw new Error(`Failed to fetch events: ${response.statusText}`);
    }
    return response.json();
  }

  async getEventCount(runId: string): Promise<number> {
    const response = await fetch(`${this.baseUrl}/runs/${runId}/events/count`);
    if (!response.ok) {
      throw new Error(`Failed to fetch event count: ${response.statusText}`);
    }
    const data = await response.json();
    return data.count;
  }

  async getRunTrace(runId: string): Promise<Trace> {
    const response = await fetch(`${this.baseUrl}/runs/${runId}/trace`);
    if (!response.ok) {
      throw new Error(`Failed to fetch trace: ${response.statusText}`);
    }
    return response.json();
  }

  getWebSocketUrl(runId?: string): string {
    // Use current host for WebSocket connection
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host;
    const wsBaseUrl = `${protocol}//${host}/api`;
    return runId ? `${wsBaseUrl}/runs/${runId}/stream` : `${wsBaseUrl}/stream`;
  }

  async listAgents(): Promise<Agent[]> {
    const response = await fetch(`${this.baseUrl}/agents`);
    if (!response.ok) {
      throw new Error(`Failed to fetch agents: ${response.statusText}`);
    }
    return response.json();
  }

  async createRun(request: StartRunRequest): Promise<Run> {
    const response = await fetch(`${this.baseUrl}/runs`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      throw new Error(`Failed to create run: ${response.statusText}`);
    }
    return response.json();
  }

  async stopRun(runId: string): Promise<{ status: string; run_id: string }> {
    const response = await fetch(`${this.baseUrl}/runs/${runId}/stop`, {
      method: "POST",
    });
    if (!response.ok) {
      throw new Error(`Failed to stop run: ${response.statusText}`);
    }
    return response.json();
  }

  async pauseRun(runId: string): Promise<{ status: string; run_id: string }> {
    const response = await fetch(`${this.baseUrl}/runs/${runId}/pause`, {
      method: "POST",
    });
    if (!response.ok) {
      throw new Error(`Failed to pause run: ${response.statusText}`);
    }
    return response.json();
  }

  async addInstruction(runId: string, instruction: string): Promise<{ status: string; run_id: string; instruction: string }> {
    const response = await fetch(`${this.baseUrl}/runs/${runId}/instruction`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ instruction }),
    });
    if (!response.ok) {
      throw new Error(`Failed to add instruction: ${response.statusText}`);
    }
    return response.json();
  }

  async resumeRun(runId: string, message: string): Promise<Run> {
    const response = await fetch(`${this.baseUrl}/runs/${runId}/resume`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ message }),
    });
    if (!response.ok) {
      throw new Error(`Failed to resume run: ${response.statusText}`);
    }
    return response.json();
  }

  // Plan management methods
  async listPlans(): Promise<ObservabilityPlan[]> {
    const response = await fetch(`${this.baseUrl}/plans`);
    if (!response.ok) {
      throw new Error(`Failed to fetch plans: ${response.statusText}`);
    }
    return response.json();
  }

  async getPlan(planId: string): Promise<ObservabilityPlan> {
    const response = await fetch(`${this.baseUrl}/plans/${planId}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch plan: ${response.statusText}`);
    }
    return response.json();
  }

  async createPlan(plan: Partial<ObservabilityPlan>): Promise<ObservabilityPlan> {
    const response = await fetch(`${this.baseUrl}/plans`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(plan),
    });
    if (!response.ok) {
      throw new Error(`Failed to create plan: ${response.statusText}`);
    }
    return response.json();
  }

  async updatePlan(planId: string, updates: Partial<ObservabilityPlan>): Promise<{ status: string }> {
    const response = await fetch(`${this.baseUrl}/plans/${planId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(updates),
    });
    if (!response.ok) {
      throw new Error(`Failed to update plan: ${response.statusText}`);
    }
    return response.json();
  }

  async deletePlan(planId: string): Promise<{ status: string }> {
    const response = await fetch(`${this.baseUrl}/plans/${planId}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      throw new Error(`Failed to delete plan: ${response.statusText}`);
    }
    return response.json();
  }

  async getPlanTopology(planId: string): Promise<TopologyGraph> {
    const response = await fetch(`${this.baseUrl}/plans/${planId}/topology`);
    if (!response.ok) {
      throw new Error(`Failed to fetch topology: ${response.statusText}`);
    }
    return response.json();
  }

  async attachBackendToPlan(planId: string, backendId: string): Promise<Backend> {
    const response = await fetch(`${this.baseUrl}/plans/${planId}/backends/${backendId}/attach`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
    });
    if (!response.ok) {
      throw new Error(`Failed to attach backend to plan: ${response.statusText}`);
    }
    return response.json();
  }

  async createService(planId: string, service: InstrumentedService): Promise<InstrumentedService> {
    const response = await fetch(`${this.baseUrl}/plans/${planId}/services`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(service),
    });
    if (!response.ok) {
      throw new Error(`Failed to create service: ${response.statusText}`);
    }
    return response.json();
  }

  async createInfrastructure(
    planId: string,
    infrastructure: InfrastructureComponent,
  ): Promise<InfrastructureComponent> {
    const response = await fetch(`${this.baseUrl}/plans/${planId}/infrastructure`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(infrastructure),
    });
    if (!response.ok) {
      throw new Error(`Failed to create infrastructure: ${response.statusText}`);
    }
    return response.json();
  }

  async createPipelineFromCollector(
    planId: string,
    collectorId: string,
    options?: {
      name?: string;
      config_yaml?: string;
      rules?: string;
      target_type?: string;
    },
  ): Promise<unknown> {
    const response = await fetch(`${this.baseUrl}/plans/${planId}/pipelines/from-collector`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        collector_id: collectorId,
        ...options,
      }),
    });
    if (!response.ok) {
      throw new Error(`Failed to create pipeline from collector: ${response.statusText}`);
    }
    return response.json();
  }

  // Sandbox management methods
  async listSandboxes(): Promise<Sandbox[]> {
    const response = await fetch(`${this.baseUrl}/sandboxes`);
    if (!response.ok) {
      throw new Error(`Failed to fetch sandboxes: ${response.statusText}`);
    }
    const data = await response.json();
    return data.sandboxes || [];
  }

  async getSandbox(sandboxId: string): Promise<Sandbox> {
    const response = await fetch(`${this.baseUrl}/sandboxes/${sandboxId}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch sandbox: ${response.statusText}`);
    }
    return response.json();
  }

  async createSandbox(request: CreateSandboxRequest): Promise<Sandbox> {
    const response = await fetch(`${this.baseUrl}/sandboxes`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      throw new Error(`Failed to create sandbox: ${response.statusText}`);
    }
    const data = await response.json();
    return data.sandbox;
  }

  async startTelemetry(sandboxId: string, request: StartTelemetryRequest): Promise<{ success: boolean; message: string }> {
    const response = await fetch(`${this.baseUrl}/sandboxes/${sandboxId}/telemetry`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      throw new Error(`Failed to start telemetry: ${response.statusText}`);
    }
    return response.json();
  }

  async validateSandbox(sandboxId: string, request: ValidateSandboxRequest): Promise<ValidationResult> {
    const response = await fetch(`${this.baseUrl}/sandboxes/${sandboxId}/validate`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      throw new Error(`Failed to validate sandbox: ${response.statusText}`);
    }
    const data = await response.json();
    return data.validation;
  }

  async getSandboxLogs(sandboxId: string, tail: number = 100): Promise<LogEntry[]> {
    const response = await fetch(`${this.baseUrl}/sandboxes/${sandboxId}/logs?tail=${tail}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch sandbox logs: ${response.statusText}`);
    }
    const data = await response.json();
    return data.logs || [];
  }

  async getSandboxMetrics(sandboxId: string): Promise<CollectorMetrics> {
    const response = await fetch(`${this.baseUrl}/sandboxes/${sandboxId}/metrics`);
    if (!response.ok) {
      throw new Error(`Failed to fetch sandbox metrics: ${response.statusText}`);
    }
    const data = await response.json();
    return data.metrics;
  }

  async stopSandbox(sandboxId: string): Promise<{ success: boolean; message: string }> {
    const response = await fetch(`${this.baseUrl}/sandboxes/${sandboxId}/stop`, {
      method: "POST",
    });
    if (!response.ok) {
      throw new Error(`Failed to stop sandbox: ${response.statusText}`);
    }
    return response.json();
  }

  async deleteSandbox(sandboxId: string): Promise<{ success: boolean; message: string }> {
    const response = await fetch(`${this.baseUrl}/sandboxes/${sandboxId}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      throw new Error(`Failed to delete sandbox: ${response.statusText}`);
    }
    return response.json();
  }

  // Collector management methods
  async listCollectors(targetType?: string): Promise<{ total_count: number; collectors: Collector[] }> {
    const params = new URLSearchParams();
    if (targetType) params.append("target_type", targetType);

    const response = await fetch(`${this.baseUrl}/collectors?${params}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch collectors: ${response.statusText}`);
    }
    return response.json();
  }

  async getCollector(collectorId: string): Promise<Collector> {
    const response = await fetch(`${this.baseUrl}/collectors/${collectorId}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch collector: ${response.statusText}`);
    }
    return response.json();
  }

  async listConnectedAgents(): Promise<{ total_count: number; agents: ConnectedAgent[] }> {
    const response = await fetch(`${this.baseUrl}/collectors/connected`);
    if (!response.ok) {
      throw new Error(`Failed to fetch connected agents: ${response.statusText}`);
    }
    return response.json();
  }

  async getCollectorConfig(collectorId: string): Promise<CollectorConfig> {
    const response = await fetch(`${this.baseUrl}/collectors/${collectorId}/config`);
    if (!response.ok) {
      throw new Error(`Failed to fetch collector config: ${response.statusText}`);
    }
    return response.json();
  }

  async updateCollectorConfig(
    collectorId: string,
    request: UpdateCollectorConfigRequest,
  ): Promise<CollectorConfig> {
    const response = await fetch(`${this.baseUrl}/collectors/${collectorId}/config`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      throw new Error(`Failed to update collector config: ${response.statusText}`);
    }
    return response.json();
  }

  async deployCollector(request: DeployCollectorRequest): Promise<unknown> {
    const response = await fetch(`${this.baseUrl}/collectors`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      throw new Error(`Failed to deploy collector: ${response.statusText}`);
    }
    return response.json();
  }

  async stopCollector(collectorId: string, targetType?: string): Promise<unknown> {
    const params = new URLSearchParams();
    if (targetType) params.append("target_type", targetType);

    const response = await fetch(`${this.baseUrl}/collectors/${collectorId}?${params}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      throw new Error(`Failed to stop collector: ${response.statusText}`);
    }
    return response.json();
  }

  async getCollectorLogs(collectorId: string, tail: number = 100): Promise<{ logs: string; tail: number }> {
    const params = new URLSearchParams();
    if (tail) params.append("tail", tail.toString());

    const response = await fetch(`${this.baseUrl}/collectors/${collectorId}/logs?${params}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch collector logs: ${response.statusText}`);
    }
    return response.json();
  }

  // Backend management methods
  async listBackends(): Promise<Backend[]> {
    const response = await fetch(`${this.baseUrl}/backends`);
    if (!response.ok) {
      throw new Error(`Failed to fetch backends: ${response.statusText}`);
    }
    return response.json();
  }

  async getBackend(backendId: string): Promise<Backend> {
    const response = await fetch(`${this.baseUrl}/backends/${backendId}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch backend: ${response.statusText}`);
    }
    return response.json();
  }

  async createBackend(request: CreateBackendRequest): Promise<Backend> {
    const response = await fetch(`${this.baseUrl}/backends`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      throw new Error(`Failed to create backend: ${response.statusText}`);
    }
    return response.json();
  }

  async updateBackend(backendId: string, request: UpdateBackendRequest): Promise<Backend> {
    const response = await fetch(`${this.baseUrl}/backends/${backendId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      throw new Error(`Failed to update backend: ${response.statusText}`);
    }
    return response.json();
  }

  async deleteBackend(backendId: string): Promise<{ status: string; backend_id: string }> {
    const response = await fetch(`${this.baseUrl}/backends/${backendId}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      throw new Error(`Failed to delete backend: ${response.statusText}`);
    }
    return response.json();
  }

  async testConnection(backendId: string, request?: TestConnectionRequest): Promise<TestConnectionResponse> {
    const response = await fetch(`${this.baseUrl}/backends/${backendId}/test-connection`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request || {}),
    });
    if (!response.ok) {
      throw new Error(`Failed to test connection: ${response.statusText}`);
    }
    return response.json();
  }

  async configureGrafanaDatasource(
    backendId: string,
    request: ConfigureGrafanaDatasourceRequest,
  ): Promise<unknown> {
    const response = await fetch(`${this.baseUrl}/backends/${backendId}/configure-datasource`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      throw new Error(`Failed to configure datasource: ${response.statusText}`);
    }
    return response.json();
  }

  // Delegation methods
  async delegateToAgent(
    resourceType: string,
    resourceId: string,
    request: DelegateRequest,
  ): Promise<DelegateResponse> {
    const response = await fetch(`${this.baseUrl}/resources/${resourceType}/${resourceId}/delegate`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `Failed to delegate task: ${response.statusText}`);
    }
    return response.json();
  }

  // Agent work methods
  async listAgentWork(params?: {
    limit?: number;
    offset?: number;
    resourceType?: string;
    resourceId?: string;
    status?: string;
  }): Promise<AgentWork[]> {
    const queryParams = new URLSearchParams();
    if (params?.limit) queryParams.set("limit", params.limit.toString());
    if (params?.offset) queryParams.set("offset", params.offset.toString());
    if (params?.resourceType) queryParams.set("resource_type", params.resourceType);
    if (params?.resourceId) queryParams.set("resource_id", params.resourceId);
    if (params?.status) queryParams.set("status", params.status);

    const url = `${this.baseUrl}/agent-work${queryParams.toString() ? `?${queryParams.toString()}` : ""}`;
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`Failed to fetch agent work: ${response.statusText}`);
    }
    const data = await response.json();
    // Handle null response from backend - normalize to empty array
    return Array.isArray(data) ? data : [];
  }

  async getAgentWork(workId: string): Promise<AgentWork> {
    const response = await fetch(`${this.baseUrl}/agent-work/${workId}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch agent work: ${response.statusText}`);
    }
    return response.json();
  }

  async cancelAgentWork(workId: string): Promise<AgentWork> {
    const response = await fetch(`${this.baseUrl}/agent-work/${workId}/cancel`, {
      method: "POST",
    });
    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `Failed to cancel agent work: ${response.statusText}`);
    }
    return response.json();
  }

  async listAgents(): Promise<Agent[]> {
    const response = await fetch(`${this.baseUrl}/agents`);
    if (!response.ok) {
      throw new Error(`Failed to fetch agents: ${response.statusText}`);
    }
    return response.json();
  }

  async getAgent(agentId: string): Promise<Agent> {
    const response = await fetch(`${this.baseUrl}/agents/${agentId}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch agent: ${response.statusText}`);
    }
    return response.json();
  }

  async getAgentTools(agentId: string): Promise<ToolInfo[]> {
    const response = await fetch(`${this.baseUrl}/agents/${agentId}/tools`);
    if (!response.ok) {
      throw new Error(`Failed to fetch agent tools: ${response.statusText}`);
    }
    return response.json();
  }

  async listAllTools(): Promise<ListToolsResponse> {
    const response = await fetch(`${this.baseUrl}/tools`);
    if (!response.ok) {
      throw new Error(`Failed to fetch tools: ${response.statusText}`);
    }
    return response.json();
  }

  async createCustomAgent(request: CreateCustomAgentRequest): Promise<Agent> {
    const response = await fetch(`${this.baseUrl}/agents/custom`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `Failed to create custom agent: ${response.statusText}`);
    }
    return response.json();
  }

  async updateCustomAgent(
    agentId: string,
    request: UpdateCustomAgentRequest,
  ): Promise<Agent> {
    const response = await fetch(`${this.baseUrl}/agents/custom/${agentId}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `Failed to update custom agent: ${response.statusText}`);
    }
    return response.json();
  }

  async deleteCustomAgent(agentId: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/agents/custom/${agentId}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `Failed to delete custom agent: ${response.statusText}`);
    }
  }

  async createMetaAgent(request: CreateMetaAgentRequest): Promise<Agent> {
    const response = await fetch(`${this.baseUrl}/agents/meta`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `Failed to create meta-agent: ${response.statusText}`);
    }
    return response.json();
  }
}

export const apiClient = new ApiClient();
