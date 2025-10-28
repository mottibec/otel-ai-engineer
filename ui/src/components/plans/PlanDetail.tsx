import { usePlan } from "../../hooks/usePlan";
import { usePlanTopology } from "../../hooks/usePlanTopology";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { StatusBadge } from "@/components/common/StatusBadge";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ServiceCard } from "./ServiceCard";
import { InfrastructureCard } from "./InfrastructureCard";
import { PipelineCard } from "./PipelineCard";
import { BackendCard } from "./BackendCard";
import { TopologyGraph } from "./TopologyGraph";

interface PlanDetailProps {
  planId: string;
}

export function PlanDetail({ planId }: PlanDetailProps) {
  const { plan, loading, error } = usePlan(planId);
  const { topology, loading: topologyLoading } = usePlanTopology(planId);

  if (loading) {
    return (
      <div className="p-6 space-y-4">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-96 w-full" />
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

  if (!plan) {
    return (
      <div className="p-6">
        <Card>
          <CardHeader>
            <CardTitle>Plan Not Found</CardTitle>
          </CardHeader>
        </Card>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">{plan.name}</h1>
          {plan.description && (
            <p className="text-muted-foreground mt-1">{plan.description}</p>
          )}
        </div>
        <div className="flex items-center gap-2">
          <StatusBadge status={plan.status} />
          {(plan.status === "draft" || plan.status === "pending" || plan.status === "partial") && (
            <button
              onClick={async () => {
                try {
                  const response = await fetch(`/api/plans/${plan.id}/execute`, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                  });
                  if (!response.ok) {
                    throw new Error("Failed to execute plan");
                  }
                  await response.json();
                  // Refresh the plan to see updated status
                  window.location.reload();
                } catch (err) {
                  console.error("Failed to execute plan:", err);
                  alert("Failed to execute plan. Please try again.");
                }
              }}
              className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
            >
              Execute Plan
            </button>
          )}
        </div>
      </div>

      <Tabs defaultValue="overview" className="w-full">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="services">Services ({plan.services?.length || 0})</TabsTrigger>
          <TabsTrigger value="infrastructure">
            Infrastructure ({plan.infrastructure?.length || 0})
          </TabsTrigger>
          <TabsTrigger value="pipelines">Pipelines ({plan.pipelines?.length || 0})</TabsTrigger>
          <TabsTrigger value="backends">Backends ({plan.backends?.length || 0})</TabsTrigger>
          <TabsTrigger value="topology">Topology</TabsTrigger>
        </TabsList>

        <TabsContent value="overview">
          <div className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Plan Information</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <span className="font-medium">Environment:</span> {plan.environment || "N/A"}
                  </div>
                  <div>
                    <span className="font-medium">Status:</span> {plan.status}
                  </div>
                  <div>
                    <span className="font-medium">Created:</span>{" "}
                    {new Date(plan.created_at).toLocaleString()}
                  </div>
                  <div>
                    <span className="font-medium">Updated:</span>{" "}
                    {new Date(plan.updated_at).toLocaleString()}
                  </div>
                </div>
              </CardContent>
            </Card>

            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg">Services</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-3xl font-bold">{plan.services?.length || 0}</div>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg">Infrastructure</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-3xl font-bold">{plan.infrastructure?.length || 0}</div>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg">Pipelines</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-3xl font-bold">{plan.pipelines?.length || 0}</div>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg">Backends</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-3xl font-bold">{plan.backends?.length || 0}</div>
                </CardContent>
              </Card>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="services">
          <div className="space-y-4">
            {plan.services && plan.services.length > 0 ? (
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {plan.services.map((service) => (
                  <ServiceCard key={service.id} service={service} />
                ))}
              </div>
            ) : (
              <Card>
                <CardContent className="py-8 text-center text-muted-foreground">
                  No services configured
                </CardContent>
              </Card>
            )}
          </div>
        </TabsContent>

        <TabsContent value="infrastructure">
          <div className="space-y-4">
            {plan.infrastructure && plan.infrastructure.length > 0 ? (
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {plan.infrastructure.map((infra) => (
                  <InfrastructureCard key={infra.id} component={infra} />
                ))}
              </div>
            ) : (
              <Card>
                <CardContent className="py-8 text-center text-muted-foreground">
                  No infrastructure components configured
                </CardContent>
              </Card>
            )}
          </div>
        </TabsContent>

        <TabsContent value="pipelines">
          <div className="space-y-4">
            {plan.pipelines && plan.pipelines.length > 0 ? (
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {plan.pipelines.map((pipeline) => (
                  <PipelineCard key={pipeline.id} pipeline={pipeline} />
                ))}
              </div>
            ) : (
              <Card>
                <CardContent className="py-8 text-center text-muted-foreground">
                  No pipelines configured
                </CardContent>
              </Card>
            )}
          </div>
        </TabsContent>

        <TabsContent value="backends">
          <div className="space-y-4">
            {plan.backends && plan.backends.length > 0 ? (
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {plan.backends.map((backend) => (
                  <BackendCard key={backend.id} backend={backend} />
                ))}
              </div>
            ) : (
              <Card>
                <CardContent className="py-8 text-center text-muted-foreground">
                  No backends configured
                </CardContent>
              </Card>
            )}
          </div>
        </TabsContent>

        <TabsContent value="topology">
          {topologyLoading ? (
            <Skeleton className="h-96 w-full" />
          ) : (
            topology && <TopologyGraph nodes={topology.nodes} edges={topology.edges} />
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}

