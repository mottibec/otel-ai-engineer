import { useEffect, useState } from "react";
import { apiClient } from "../services/api";
import type { ObservabilityPlan } from "../types/plan";

export function usePlans() {
  const [plans, setPlans] = useState<ObservabilityPlan[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchPlans = async () => {
      try {
        setLoading(true);
        const data = await apiClient.listPlans();
        setPlans(data);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err : new Error(String(err)));
      } finally {
        setLoading(false);
      }
    };

    fetchPlans();
  }, []);

  const refreshPlans = async () => {
    try {
      setLoading(true);
      const data = await apiClient.listPlans();
      setPlans(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err : new Error(String(err)));
    } finally {
      setLoading(false);
    }
  };

  return { plans, loading, error, refreshPlans };
}

