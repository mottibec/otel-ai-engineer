import { useEffect, useState } from "react";
import { apiClient } from "../services/api";
import type { ObservabilityPlan } from "../types/plan";

export function usePlan(planId: string | undefined) {
  const [plan, setPlan] = useState<ObservabilityPlan | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!planId) {
      setPlan(null);
      setLoading(false);
      return;
    }

    const fetchPlan = async () => {
      try {
        setLoading(true);
        const data = await apiClient.getPlan(planId);
        setPlan(data);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err : new Error(String(err)));
        setPlan(null);
      } finally {
        setLoading(false);
      }
    };

    fetchPlan();
  }, [planId]);

  const refreshPlan = async () => {
    if (!planId) return;

    try {
      setLoading(true);
      const data = await apiClient.getPlan(planId);
      setPlan(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err : new Error(String(err)));
    } finally {
      setLoading(false);
    }
  };

  return { plan, loading, error, refreshPlan };
}

