import { useState, useEffect } from "react";
import { useSWRConfig } from "swr";
import { apiClient } from "@/services/api";
import type { Collector, CollectorConfig } from "@/types/collector";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Server,
  RefreshCw,
  Trash2,
  Loader2,
  Settings,
  Bot,
  Code,
  FileText,
} from "lucide-react";
import { AgentWorkBadges } from "@/components/common/AgentWorkBadge";
import { DelegateToAgent } from "@/components/common/DelegateToAgent";
import { CodeBlock } from "@/components/common/CodeBlock";

interface CollectorDetailProps {
  collectorId: string;
}

export function CollectorDetail({ collectorId }: CollectorDetailProps) {
  const { mutate } = useSWRConfig();
  const [collector, setCollector] = useState<Collector | null>(null);
  const [config, setConfig] = useState<CollectorConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [configLoading, setConfigLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [configError, setConfigError] = useState<string | null>(null);
  const [editingConfig, setEditingConfig] = useState(false);
  const [editedYaml, setEditedYaml] = useState("");
  const [updating, setUpdating] = useState(false);
  const [stopping, setStopping] = useState(false);
  const [showDelegateModal, setShowDelegateModal] = useState(false);
  const [logs, setLogs] = useState<string>("");
  const [logsLoading, setLogsLoading] = useState(false);
  const [logsError, setLogsError] = useState<string | null>(null);

  const loadCollector = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.getCollector(collectorId);
      setCollector(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load collector");
    } finally {
      setLoading(false);
    }
  };

  const loadConfig = async () => {
    try {
      setConfigLoading(true);
      setConfigError(null);
      const data = await apiClient.getCollectorConfig(collectorId);
      setConfig(data);
      setEditedYaml(data.yaml_content);
    } catch (err) {
      setConfigError(err instanceof Error ? err.message : "Failed to load config");
    } finally {
      setConfigLoading(false);
    }
  };

  useEffect(() => {
    loadCollector();
    loadConfig();
  }, [collectorId]);

  const handleUpdateConfig = async () => {
    if (!config || !editedYaml.trim()) return;

    setUpdating(true);
    try {
      await apiClient.updateCollectorConfig(collectorId, {
        yaml_config: editedYaml.trim(),
      });
      setEditingConfig(false);
      await loadConfig();
      await loadCollector();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to update config");
    } finally {
      setUpdating(false);
    }
  };

  const handleStopCollector = async () => {
    if (!window.confirm(`Are you sure you want to stop collector "${collector?.collector_name}"?`)) {
      return;
    }

    setStopping(true);
    try {
      await apiClient.stopCollector(collectorId);
      // Invalidate collectors list
      mutate("/api/collectors");
      // Navigate back to list (parent should handle this)
      window.history.back();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to stop collector");
    } finally {
      setStopping(false);
    }
  };

  const loadLogs = async () => {
    // Only load logs for Docker collectors
    if (collector?.target_type !== "docker") {
      return;
    }

    try {
      setLogsLoading(true);
      setLogsError(null);
      const data = await apiClient.getCollectorLogs(collectorId, 200);
      setLogs(data.logs);
    } catch (err) {
      setLogsError(err instanceof Error ? err.message : "Failed to load logs");
    } finally {
      setLogsLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="p-6 space-y-4">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-96 w-full" />
      </div>
    );
  }

  if (error || !collector) {
    return (
      <div className="p-6">
        <Card className="border-red-500/50 bg-red-500/5">
          <CardHeader>
            <CardTitle className="text-red-500">Error</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">{error || "Collector not found"}</p>
            <Button onClick={loadCollector} className="mt-4" variant="outline">
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
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <Server className="h-8 w-8" />
            {collector.collector_name}
          </h1>
          <div className="flex items-center gap-2 mt-2">
            <Badge variant={collector.status === "running" ? "default" : "secondary"}>
              {collector.status}
            </Badge>
            <Badge variant="outline">{collector.target_type}</Badge>
            <AgentWorkBadges works={collector.agent_work} />
          </div>
        </div>
        <div className="flex gap-2">
          <Button onClick={() => setShowDelegateModal(true)} variant="outline">
            <Bot className="mr-2 h-4 w-4" />
            Delegate to Agent
          </Button>
          <Button onClick={loadConfig} disabled={configLoading} variant="outline">
            {configLoading ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <RefreshCw className="mr-2 h-4 w-4" />
            )}
            Refresh Config
          </Button>
          <Button onClick={handleStopCollector} disabled={stopping} variant="destructive">
            {stopping ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Trash2 className="mr-2 h-4 w-4" />
            )}
            Stop Collector
          </Button>
        </div>
      </div>

      <Tabs defaultValue="overview" className="w-full" onValueChange={(value) => {
        if (value === "logs" && collector?.target_type === "docker" && !logs && !logsLoading) {
          loadLogs();
        }
      }}>
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="config">Configuration</TabsTrigger>
          <TabsTrigger value="logs" disabled={collector.target_type !== "docker"}>
            Logs
          </TabsTrigger>
        </TabsList>

        <TabsContent value="overview">
          <div className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Collector Information</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label>ID</Label>
                    <p className="text-sm text-muted-foreground">{collector.collector_id}</p>
                  </div>
                  <div>
                    <Label>Name</Label>
                    <p className="text-sm text-muted-foreground">{collector.collector_name}</p>
                  </div>
                  <div>
                    <Label>Target Type</Label>
                    <p className="text-sm text-muted-foreground">{collector.target_type}</p>
                  </div>
                  <div>
                    <Label>Status</Label>
                    <p className="text-sm text-muted-foreground">{collector.status}</p>
                  </div>
                  <div>
                    <Label>Deployed At</Label>
                    <p className="text-sm text-muted-foreground">
                      {new Date(collector.deployed_at).toLocaleString()}
                    </p>
                  </div>
                  {collector.config_path && (
                    <div>
                      <Label>Config Path</Label>
                      <p className="text-sm text-muted-foreground">{collector.config_path}</p>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="config">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Configuration</CardTitle>
                  <CardDescription>
                    {config
                      ? `Version ${config.config_version} - ${config.config_name}`
                      : "Loading configuration..."}
                  </CardDescription>
                </div>
                {config && !editingConfig && (
                  <Button onClick={() => setEditingConfig(true)} variant="outline" size="sm">
                    <Settings className="mr-2 h-4 w-4" />
                    Edit
                  </Button>
                )}
              </div>
            </CardHeader>
            <CardContent>
              {configLoading ? (
                <div className="space-y-2">
                  <Skeleton className="h-4 w-full" />
                  <Skeleton className="h-4 w-full" />
                  <Skeleton className="h-4 w-3/4" />
                </div>
              ) : configError ? (
                <div className="p-4 bg-destructive/10 border border-destructive rounded-md text-destructive text-sm">
                  {configError}
                </div>
              ) : config ? (
                <div className="space-y-4">
                  {editingConfig ? (
                    <>
                      <div className="space-y-2">
                        <Label>YAML Configuration</Label>
                        <Textarea
                          value={editedYaml}
                          onChange={(e) => setEditedYaml(e.target.value)}
                          disabled={updating}
                          className="font-mono text-sm"
                          rows={20}
                        />
                      </div>
                      <div className="flex gap-2">
                        <Button onClick={() => setEditingConfig(false)} variant="outline" disabled={updating}>
                          Cancel
                        </Button>
                        <Button onClick={handleUpdateConfig} disabled={updating}>
                          {updating ? (
                            <>
                              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                              Updating...
                            </>
                          ) : (
                            "Save Configuration"
                          )}
                        </Button>
                      </div>
                    </>
                  ) : (
                    <div className="space-y-2">
                      <div className="flex items-center gap-2 mb-2">
                        <Code className="h-4 w-4 text-muted-foreground" />
                        <Label>YAML Configuration</Label>
                      </div>
                      <CodeBlock language="yaml" code={config.yaml_content} />
                    </div>
                  )}
                </div>
              ) : (
                <div className="text-sm text-muted-foreground">No configuration available</div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="logs">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Container Logs</CardTitle>
                  <CardDescription>
                    Real-time logs from the collector container
                  </CardDescription>
                </div>
                <Button onClick={loadLogs} disabled={logsLoading} variant="outline" size="sm">
                  {logsLoading ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <RefreshCw className="mr-2 h-4 w-4" />
                  )}
                  Refresh
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {logsLoading ? (
                <div className="space-y-2">
                  <Skeleton className="h-4 w-full" />
                  <Skeleton className="h-4 w-full" />
                  <Skeleton className="h-4 w-3/4" />
                </div>
              ) : logsError ? (
                <div className="p-4 bg-destructive/10 border border-destructive rounded-md text-destructive text-sm">
                  {logsError}
                </div>
              ) : logs ? (
                <div className="space-y-2">
                  <div className="flex items-center gap-2 mb-2">
                    <FileText className="h-4 w-4 text-muted-foreground" />
                    <Label>Container Output</Label>
                  </div>
                  <div className="bg-muted rounded-md p-4 max-h-[600px] overflow-y-auto">
                    <pre className="text-xs font-mono whitespace-pre-wrap break-words">
                      {logs}
                    </pre>
                  </div>
                </div>
              ) : (
                <div className="text-sm text-muted-foreground">No logs available. Click Refresh to load logs.</div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <DelegateToAgent
        isOpen={showDelegateModal}
        onClose={() => setShowDelegateModal(false)}
        resourceType="collector"
        resourceId={collectorId}
        resourceName={collector.collector_name}
        onDelegated={async (runId, workId) => {
          // Reload collector to show new agent work
          await loadCollector();
        }}
      />
    </div>
  );
}

