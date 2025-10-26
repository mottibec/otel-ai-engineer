import useSWR from "swr";
import type { Run } from "../types/models";
import { apiClient } from "../services/api";

const fetcher = (runId: string) => apiClient.getRun(runId);

export function useRun(runId: string | undefined) {
  const {
    data: run,
    error,
    isLoading: loading,
  } = useSWR<Run>(runId ? `/api/runs/${runId}` : null, () =>
    runId ? fetcher(runId) : null
  );

  return {
    run,
    loading,
    error: error ? (error instanceof Error ? error.message : "Failed to fetch run") : null,
  };
}

