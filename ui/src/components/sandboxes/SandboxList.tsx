import { useState, useEffect } from "react";
import { apiClient } from "@/services/api";
import type { Sandbox } from "@/types/sandbox";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { FlaskConical, Plus, RefreshCw } from "lucide-react";

interface SandboxListProps {
  onSelectSandbox: (sandboxId: string) => void;
  onCreateSandbox: () => void;
}

export function SandboxList({ onSelectSandbox, onCreateSandbox }: SandboxListProps) {
  const [sandboxes, setSandboxes] = useState<Sandbox[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadSandboxes = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.listSandboxes();
      setSandboxes(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load sandboxes");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadSandboxes();
  }, []);

  const getStatusColor = (status: string) => {
    switch (status) {
      case "running":
        return "bg-green-500/10 text-green-500 border-green-500/20";
      case "creating":
        return "bg-blue-500/10 text-blue-500 border-blue-500/20";
      case "validating":
        return "bg-yellow-500/10 text-yellow-500 border-yellow-500/20";
      case "stopped":
        return "bg-gray-500/10 text-gray-500 border-gray-500/20";
      case "failed":
        return "bg-red-500/10 text-red-500 border-red-500/20";
      default:
        return "bg-gray-500/10 text-gray-500 border-gray-500/20";
    }
  };

  const getValidationStatusColor = (status?: string) => {
    if (!status) return "bg-gray-500/10 text-gray-500 border-gray-500/20";
    switch (status) {
      case "passed":
        return "bg-green-500/10 text-green-500 border-green-500/20";
      case "failed":
        return "bg-red-500/10 text-red-500 border-red-500/20";
      case "partial":
        return "bg-yellow-500/10 text-yellow-500 border-yellow-500/20";
      case "running":
        return "bg-blue-500/10 text-blue-500 border-blue-500/20";
      default:
        return "bg-gray-500/10 text-gray-500 border-gray-500/20";
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
            <Button onClick={loadSandboxes} className="mt-4" variant="outline">
              <RefreshCw className="mr-2 h-4 w-4" />
              Retry
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <FlaskConical className="h-8 w-8" />
            Collector Sandboxes
          </h1>
          <p className="text-muted-foreground mt-1">
            Test and validate OpenTelemetry collector configurations
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={loadSandboxes}>
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh
          </Button>
          <Button onClick={onCreateSandbox}>
            <Plus className="mr-2 h-4 w-4" />
            New Sandbox
          </Button>
        </div>
      </div>

      {sandboxes.length === 0 ? (
        <Card>
          <CardHeader className="text-center">
            <div className="text-6xl mb-4">🧪</div>
            <CardTitle>No Sandboxes Yet</CardTitle>
            <CardDescription>
              Create your first sandbox to test collector configurations in isolation
            </CardDescription>
          </CardHeader>
          <CardContent className="text-center">
            <Button onClick={onCreateSandbox}>
              <Plus className="mr-2 h-4 w-4" />
              Create Sandbox
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4">
          {sandboxes.map((sandbox) => (
            <Card
              key={sandbox.id}
              className="cursor-pointer hover:bg-accent/50 transition-colors"
              onClick={() => onSelectSandbox(sandbox.id)}
            >
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div className="space-y-1 flex-1">
                    <CardTitle className="text-xl">{sandbox.name}</CardTitle>
                    {sandbox.description && (
                      <CardDescription>{sandbox.description}</CardDescription>
                    )}
                  </div>
                  <Badge variant="outline" className={getStatusColor(sandbox.status)}>
                    {sandbox.status}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                  <div>
                    <div className="text-muted-foreground">Collector Version</div>
                    <div className="font-medium">{sandbox.collector_version}</div>
                  </div>
                  <div>
                    <div className="text-muted-foreground">Network</div>
                    <div className="font-medium font-mono text-xs">
                      {sandbox.network_name || "—"}
                    </div>
                  </div>
                  <div>
                    <div className="text-muted-foreground">Telemetry</div>
                    <div className="flex gap-1 mt-1">
                      {sandbox.telemetry_config.generate_traces && (
                        <Badge variant="secondary" className="text-xs">Traces</Badge>
                      )}
                      {sandbox.telemetry_config.generate_metrics && (
                        <Badge variant="secondary" className="text-xs">Metrics</Badge>
                      )}
                      {sandbox.telemetry_config.generate_logs && (
                        <Badge variant="secondary" className="text-xs">Logs</Badge>
                      )}
                      {!sandbox.telemetry_config.generate_traces &&
                        !sandbox.telemetry_config.generate_metrics &&
                        !sandbox.telemetry_config.generate_logs && (
                          <span className="text-muted-foreground">None</span>
                        )}
                    </div>
                  </div>
                  <div>
                    <div className="text-muted-foreground">Last Validation</div>
                    {sandbox.last_validation ? (
                      <Badge
                        variant="outline"
                        className={getValidationStatusColor(sandbox.last_validation.status)}
                      >
                        {sandbox.last_validation.status}
                      </Badge>
                    ) : (
                      <div className="text-muted-foreground">Not validated</div>
                    )}
                  </div>
                </div>

                {/* Show error logs if status is failed */}
                {sandbox.status === "failed" && sandbox.metadata?.error_logs && (
                  <div className="mt-4 p-3 bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800 rounded-md">
                    <div className="text-sm font-semibold text-red-700 dark:text-red-400 mb-2">
                      Container Failed - Error Logs:
                    </div>
                    <div className="space-y-1 max-h-32 overflow-y-auto">
                      {(sandbox.metadata.error_logs as string[]).map((log, idx) => (
                        <div key={idx} className="text-xs font-mono text-red-600 dark:text-red-300">
                          {log}
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                <div className="mt-4 text-xs text-muted-foreground">
                  Created {new Date(sandbox.created_at).toLocaleString()}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
