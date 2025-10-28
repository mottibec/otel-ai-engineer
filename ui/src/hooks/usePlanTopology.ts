import { useEffect, useState } from "react";
import { apiClient } from "../services/api";
import type { TopologyGraph } from "../types/plan";

export function usePlanTopology(planId: string | undefined) {
  const [topology, setTopology] = useState<TopologyGraph | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!planId) {
      setTopology(null);
      setLoading(false);
      return;
    }

    const fetchTopology = async () => {
      try {
        setLoading(true);
        const data = await apiClient.getPlanTopology(planId);
        setTopology(data);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err : new Error(String(err)));
        setTopology(null);
      } finally {
        setLoading(false);
      }
    };

    fetchTopology();
  }, [planId]);

  const refreshTopology = async () => {
    if (!planId) return;

    try {
      setLoading(true);
      const data = await apiClient.getPlanTopology(planId);
      setTopology(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err : new Error(String(err)));
    } finally {
      setLoading(false);
    }
  };

  return { topology, loading, error, refreshTopology };
}

