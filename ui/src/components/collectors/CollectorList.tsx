import { useState, useEffect } from "react";
import { apiClient } from "@/services/api";
import type { Collector } from "@/types/collector";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Server, Plus, RefreshCw, Trash2 } from "lucide-react";
import { AgentWorkBadges } from "@/components/common/AgentWorkBadge";
import { StatusBadge } from "@/components/common/StatusBadge";

interface CollectorListProps {
  onSelectCollector: (collectorId: string) => void;
  onCreateCollector: () => void;
}

export function CollectorList({ onSelectCollector, onCreateCollector }: CollectorListProps) {
  const [collectors, setCollectors] = useState<Collector[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadCollectors = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.listCollectors();
      setCollectors(data.collectors);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load collectors");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadCollectors();
  }, []);

  const handleStopCollector = async (e: React.MouseEvent, collectorId: string, targetType: string) => {
    e.stopPropagation();
    if (!confirm("Are you sure you want to stop this collector?")) {
      return;
    }

    try {
      await apiClient.stopCollector(collectorId, targetType);
      loadCollectors();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to stop collector");
    }
  };

  if (loading) {
    return (
      <div className="p-6 space-y-4">
        <div className="flex justify-between items-center">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-10 w-40" />
        </div>
        {[1, 2, 3].map((i) => (
          <Skeleton key={i} className="h-32 w-full" />
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <Card className="border-red-500/50 bg-red-500/5">
          <CardHeader>
            <CardTitle className="text-red-500">Error</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">{error}</p>
            <Button onClick={loadCollectors} className="mt-4" variant="outline">
              <RefreshCw className="mr-2 h-4 w-4" />
              Retry
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-4">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold flex items-center gap-2">
            <Server className="h-6 w-6" />
            Collectors
          </h2>
          <p className="text-sm text-muted-foreground mt-1">
            Manage your OpenTelemetry collectors
          </p>
        </div>
        <div className="flex gap-2">
          <Button onClick={loadCollectors} variant="outline" size="sm">
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh
          </Button>
          <Button onClick={onCreateCollector}>
            <Plus className="mr-2 h-4 w-4" />
            Deploy Collector
          </Button>
        </div>
      </div>

      {collectors.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <Server className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
            <CardTitle className="mb-2">No collectors deployed</CardTitle>
            <CardDescription className="mb-4">
              Deploy your first OpenTelemetry collector to start collecting telemetry data.
            </CardDescription>
            <Button onClick={onCreateCollector}>
              <Plus className="mr-2 h-4 w-4" />
              Deploy Collector
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {collectors.map((collector) => (
            <Card
              key={collector.collector_id}
              className="cursor-pointer hover:bg-accent transition-colors"
              onClick={() => onSelectCollector(collector.collector_id)}
            >
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="flex-1 min-w-0">
                    <CardTitle className="text-lg truncate">{collector.collector_name}</CardTitle>
                    <CardDescription className="mt-1 text-xs">
                      {collector.collector_id}
                    </CardDescription>
                  </div>
                  <div className="flex items-center gap-2 ml-2">
                    <StatusBadge status={collector.status} />
                    <AgentWorkBadges
                      works={collector.agent_work}
                      onClick={(runId) => {
                        window.dispatchEvent(new CustomEvent("navigateToRun", { detail: { runId } }));
                      }}
                    />
                  </div>
                </div>
              </CardHeader>
              <CardContent className="pt-0">
                <div className="space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">Target:</span>
                    <Badge variant="outline" className="text-xs">
                      {collector.target_type}
                    </Badge>
                  </div>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">Deployed:</span>
                    <span className="text-xs">
                      {new Date(collector.deployed_at).toLocaleString()}
                    </span>
                  </div>
                  <div className="flex justify-end gap-2 pt-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={(e) => handleStopCollector(e, collector.collector_id, collector.target_type)}
                      className="text-destructive hover:text-destructive"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}

