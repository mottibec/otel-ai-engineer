import { useState, useEffect } from "react";
import { apiClient } from "@/services/api";
import type { Backend } from "@/types/backend";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Database, Plus, RefreshCw, Trash2, Activity } from "lucide-react";
import { AgentWorkBadges } from "@/components/common/AgentWorkBadge";
import { BackendHealthIndicator } from "./BackendHealthIndicator";

interface BackendListProps {
  onSelectBackend: (backendId: string) => void;
  onCreateBackend: () => void;
}

export function BackendList({ onSelectBackend, onCreateBackend }: BackendListProps) {
  const [backends, setBackends] = useState<Backend[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadBackends = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.listBackends();
      setBackends(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load backends");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadBackends();
  }, []);

  const handleDeleteBackend = async (e: React.MouseEvent, backendId: string) => {
    e.stopPropagation();
    if (!confirm("Are you sure you want to delete this backend?")) {
      return;
    }

    try {
      await apiClient.deleteBackend(backendId);
      loadBackends();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to delete backend");
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
            <Button onClick={loadBackends} className="mt-4" variant="outline">
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
            <Database className="h-6 w-6" />
            Backends
          </h2>
          <p className="text-sm text-muted-foreground mt-1">
            Manage your observability backends (Grafana, Prometheus, etc.)
          </p>
        </div>
        <div className="flex gap-2">
          <Button onClick={loadBackends} variant="outline" size="sm">
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh
          </Button>
          <Button onClick={onCreateBackend}>
            <Plus className="mr-2 h-4 w-4" />
            Add Backend
          </Button>
        </div>
      </div>

      {backends.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <Database className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
            <CardTitle className="mb-2">No backends configured</CardTitle>
            <CardDescription className="mb-4">
              Add your first observability backend to start visualizing and querying telemetry data.
            </CardDescription>
            <Button onClick={onCreateBackend}>
              <Plus className="mr-2 h-4 w-4" />
              Add Backend
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {backends.map((backend) => (
            <Card
              key={backend.id}
              className="cursor-pointer hover:bg-accent transition-colors"
              onClick={() => onSelectBackend(backend.id)}
            >
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="flex-1 min-w-0">
                    <CardTitle className="text-lg truncate">{backend.name}</CardTitle>
                    <CardDescription className="mt-1 text-xs">
                      {backend.id}
                    </CardDescription>
                  </div>
                  <div className="flex items-center gap-2 ml-2">
                    <BackendHealthIndicator status={backend.health_status} />
                    <AgentWorkBadges
                      works={backend.agent_work}
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
                    <span className="text-muted-foreground">Type:</span>
                    <Badge variant="outline" className="text-xs">
                      {backend.backend_type}
                    </Badge>
                  </div>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">URL:</span>
                    <span className="text-xs truncate max-w-[200px]">{backend.url}</span>
                  </div>
                  {backend.last_check && (
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">Last checked:</span>
                      <span className="text-xs">
                        {new Date(backend.last_check).toLocaleString()}
                      </span>
                    </div>
                  )}
                  <div className="flex justify-end gap-2 pt-2">
                    {backend.backend_type === "grafana" && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={(e) => {
                          e.stopPropagation();
                          onSelectBackend(backend.id);
                        }}
                        className="text-sm"
                      >
                        <Activity className="h-4 w-4 mr-1" />
                        Configure
                      </Button>
                    )}
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={(e) => handleDeleteBackend(e, backend.id)}
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

