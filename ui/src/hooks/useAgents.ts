import useSWR from "swr";
import { apiClient } from "../services/api";
import type { Agent } from "../types/models";

const fetcher = () => apiClient.listAgents();

export function useAgents() {
  const {
    data: agents = [],
    error,
    isLoading: loading,
  } = useSWR<Agent[]>("/api/agents", fetcher, {
    revalidateOnFocus: true,
    revalidateOnReconnect: true,
  });

  return {
    agents,
    loading,
    error: error ? (error instanceof Error ? error : new Error("Failed to fetch agents")) : null,
  };
}
