import type { Run, Agent, StartRunRequest } from "../types/models";
import type { AgentEvent } from "../types/events";

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
}

export const apiClient = new ApiClient();
