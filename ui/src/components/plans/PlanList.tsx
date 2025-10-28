import { useState } from "react";
import { usePlans } from "../../hooks/usePlans";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { StatusBadge } from "@/components/common/StatusBadge";
import { CreatePlanWizard } from "./CreatePlanWizard";
import type { ObservabilityPlan } from "../../types/plan";

interface PlanListProps {
  onSelectPlan: (planId: string) => void;
}

export function PlanList({ onSelectPlan }: PlanListProps) {
  const { plans, loading, error, refreshPlans } = usePlans();
  const [isWizardOpen, setIsWizardOpen] = useState(false);

  if (loading) {
    return (
      <div className="space-y-4 p-6">
        <div className="flex items-center justify-between">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-10 w-32" />
        </div>
        <div className="grid gap-4">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-32 w-full" />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <Card>
          <CardHeader>
            <CardTitle>Error</CardTitle>
            <CardDescription>{error.message}</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  const handlePlanCreated = (plan: ObservabilityPlan) => {
    refreshPlans();
    onSelectPlan(plan.id);
  };

  if (plans.length === 0) {
    return (
      <div className="p-6">
        <Card>
          <CardHeader>
            <CardTitle>No Plans</CardTitle>
            <CardDescription>
              Create your first observability plan to get started
            </CardDescription>
          </CardHeader>
          <CardContent>
            <button
              onClick={() => setIsWizardOpen(true)}
              className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
            >
              Create Your First Plan
            </button>
          </CardContent>
        </Card>

        <CreatePlanWizard
          isOpen={isWizardOpen}
          onClose={() => setIsWizardOpen(false)}
          onPlanCreated={handlePlanCreated}
        />
      </div>
    );
  }

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Observability Plans</h1>
        <div className="flex gap-2">
          <button
            onClick={refreshPlans}
            className="px-4 py-2 text-sm border rounded-md hover:bg-accent"
          >
            Refresh
          </button>
          <button
            onClick={() => setIsWizardOpen(true)}
            className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
          >
            New Plan
          </button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {plans.map((plan) => (
          <Card
            key={plan.id}
            className="cursor-pointer transition-colors hover:bg-accent"
            onClick={() => onSelectPlan(plan.id)}
          >
            <CardHeader>
              <div className="flex items-start justify-between">
                <div>
                  <CardTitle className="text-lg">{plan.name}</CardTitle>
                  {plan.description && (
                    <CardDescription className="mt-1 line-clamp-2">
                      {plan.description}
                    </CardDescription>
                  )}
                </div>
                <StatusBadge status={plan.status} />
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {plan.environment && (
                  <div className="text-sm">
                    <span className="font-medium">Environment:</span> {plan.environment}
                  </div>
                )}
                <div className="grid grid-cols-2 gap-2 text-sm text-muted-foreground">
                  <div>
                    <span className="font-medium">Services:</span> {plan.services?.length || 0}
                  </div>
                  <div>
                    <span className="font-medium">Infra:</span> {plan.infrastructure?.length || 0}
                  </div>
                  <div>
                    <span className="font-medium">Pipelines:</span> {plan.pipelines?.length || 0}
                  </div>
                  <div>
                    <span className="font-medium">Backends:</span> {plan.backends?.length || 0}
                  </div>
                </div>
                <div className="text-xs text-muted-foreground">
                  Updated: {new Date(plan.updated_at).toLocaleString()}
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <CreatePlanWizard
        isOpen={isWizardOpen}
        onClose={() => setIsWizardOpen(false)}
        onPlanCreated={handlePlanCreated}
      />
    </div>
  );
}

