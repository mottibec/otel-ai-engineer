import useSWR from "swr";
import type { Run } from "../types/models";
import { apiClient } from "../services/api";

const fetcher = () => apiClient.listRuns();

export function useRuns() {
  const {
    data: runs = [],
    error,
    isLoading: loading,
    mutate,
  } = useSWR<Run[]>("/api/runs", fetcher, {
    refreshInterval: 0, // Refresh every 2 seconds
    revalidateOnFocus: true,
    revalidateOnReconnect: true,
  });

  return {
    runs,
    loading,
    error: error
      ? error instanceof Error
        ? error.message
        : "Failed to fetch runs"
      : null,
    refetch: mutate,
  };
}
