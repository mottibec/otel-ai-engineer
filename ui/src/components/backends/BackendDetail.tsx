import { useState, useEffect } from "react";
import { apiClient } from "@/services/api";
import type { Backend, ConfigureGrafanaDatasourceRequest } from "@/types/backend";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Database,
  RefreshCw,
  Activity,
  CheckCircle2,
  XCircle,
  Loader2,
  Settings,
  Bot,
} from "lucide-react";
import { AgentWorkBadges } from "@/components/common/AgentWorkBadge";
import { BackendHealthIndicator } from "./BackendHealthIndicator";
import { DelegateToAgent } from "@/components/common/DelegateToAgent";

interface BackendDetailProps {
  backendId: string;
}

export function BackendDetail({ backendId }: BackendDetailProps) {
  const [backend, setBackend] = useState<Backend | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<{ healthy: boolean; error?: string } | null>(null);
  const [editing, setEditing] = useState(false);
  const [editedUrl, setEditedUrl] = useState("");
  const [editedUsername, setEditedUsername] = useState("");
  const [editedPassword, setEditedPassword] = useState("");

  // Grafana datasource configuration
  const [configuringDatasource, setConfiguringDatasource] = useState(false);
  const [datasourceName, setDatasourceName] = useState("");
  const [datasourceType, setDatasourceType] = useState("otlp");
  const [datasourceUrl, setDatasourceUrl] = useState("http://lawrence:4318");

  // Delegation
  const [showDelegateModal, setShowDelegateModal] = useState(false);

  const loadBackend = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.getBackend(backendId);
      setBackend(data);
      setEditedUrl(data.url);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load backend");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadBackend();
  }, [backendId]);

  const handleTestConnection = async () => {
    if (!backend) return;

    setTesting(true);
    setTestResult(null);

    try {
      const result = await apiClient.testConnection(backendId);
      setTestResult({ healthy: result.healthy, error: result.error });
      // Reload backend to get updated health status
      await loadBackend();
    } catch (err) {
      setTestResult({
        healthy: false,
        error: err instanceof Error ? err.message : "Connection test failed",
      });
    } finally {
      setTesting(false);
    }
  };

  const handleSave = async () => {
    if (!backend) return;

    try {
      await apiClient.updateBackend(backendId, {
        url: editedUrl !== backend.url ? editedUrl : undefined,
        username: editedUsername || undefined,
        password: editedPassword || undefined,
      });
      setEditing(false);
      await loadBackend();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to update backend");
    }
  };

  const handleConfigureDatasource = async () => {
    if (!backend || backend.backend_type !== "grafana") return;

    setConfiguringDatasource(true);
    try {
      const request: ConfigureGrafanaDatasourceRequest = {
        datasource_name: datasourceName,
        datasource_type: datasourceType,
        url: datasourceUrl,
      };
      await apiClient.configureGrafanaDatasource(backendId, request);
      alert("Datasource configured successfully!");
      setDatasourceName("");
      setDatasourceUrl("http://lawrence:4318");
      await loadBackend();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to configure datasource");
    } finally {
      setConfiguringDatasource(false);
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

  if (error || !backend) {
    return (
      <div className="p-6">
        <Card className="border-red-500/50 bg-red-500/5">
          <CardHeader>
            <CardTitle className="text-red-500">Error</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">{error || "Backend not found"}</p>
            <Button onClick={loadBackend} className="mt-4" variant="outline">
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
            <Database className="h-8 w-8" />
            {backend.name}
          </h1>
          <div className="flex items-center gap-2 mt-2">
            <BackendHealthIndicator status={backend.health_status} />
            <Badge variant="outline">{backend.backend_type}</Badge>
            <AgentWorkBadges
              works={backend.agent_work}
              onClick={(runId) => {
                // Navigate to run view - handled by App via state update
                window.location.hash = `#run-${runId}`;
                // Trigger a custom event for App to handle
                window.dispatchEvent(new CustomEvent("navigateToRun", { detail: { runId } }));
              }}
            />
          </div>
        </div>
        <div className="flex gap-2">
          <Button onClick={() => setShowDelegateModal(true)} variant="outline">
            <Bot className="mr-2 h-4 w-4" />
            Delegate to Agent
          </Button>
          <Button onClick={handleTestConnection} disabled={testing} variant="outline">
            {testing ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Activity className="mr-2 h-4 w-4" />
            )}
            Test Connection
          </Button>
          {!editing ? (
            <Button onClick={() => setEditing(true)} variant="outline">
              <Settings className="mr-2 h-4 w-4" />
              Edit
            </Button>
          ) : (
            <>
              <Button onClick={() => setEditing(false)} variant="outline">
                Cancel
              </Button>
              <Button onClick={handleSave}>
                Save
              </Button>
            </>
          )}
        </div>
      </div>

      {testResult && (
        <Card className={testResult.healthy ? "border-green-500/50 bg-green-500/5" : "border-red-500/50 bg-red-500/5"}>
          <CardContent className="pt-6">
            <div className="flex items-center gap-2">
              {testResult.healthy ? (
                <CheckCircle2 className="h-5 w-5 text-green-500" />
              ) : (
                <XCircle className="h-5 w-5 text-red-500" />
              )}
              <span className={testResult.healthy ? "text-green-500" : "text-red-500"}>
                {testResult.healthy ? "Connection successful" : testResult.error || "Connection failed"}
              </span>
            </div>
          </CardContent>
        </Card>
      )}

      <Tabs defaultValue="overview" className="w-full">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          {backend.backend_type === "grafana" && (
            <TabsTrigger value="datasources">Data Sources</TabsTrigger>
          )}
        </TabsList>

        <TabsContent value="overview">
          <div className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Backend Information</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label>ID</Label>
                    <p className="text-sm text-muted-foreground">{backend.id}</p>
                  </div>
                  <div>
                    <Label>Type</Label>
                    <p className="text-sm text-muted-foreground">{backend.backend_type}</p>
                  </div>
                  <div>
                    <Label>URL</Label>
                    {editing ? (
                      <Input
                        value={editedUrl}
                        onChange={(e) => setEditedUrl(e.target.value)}
                        className="mt-1"
                      />
                    ) : (
                      <p className="text-sm text-muted-foreground">{backend.url}</p>
                    )}
                  </div>
                  <div>
                    <Label>Health Status</Label>
                    <p className="text-sm text-muted-foreground">{backend.health_status}</p>
                  </div>
                  {backend.last_check && (
                    <div>
                      <Label>Last Check</Label>
                      <p className="text-sm text-muted-foreground">
                        {new Date(backend.last_check).toLocaleString()}
                      </p>
                    </div>
                  )}
                  {backend.datasource_uid && (
                    <div>
                      <Label>Datasource UID</Label>
                      <p className="text-sm text-muted-foreground">{backend.datasource_uid}</p>
                    </div>
                  )}
                </div>

                {editing && (backend.backend_type === "grafana" || backend.backend_type === "custom") && (
                  <div className="space-y-2 pt-4 border-t">
                    <Label>Credentials (optional)</Label>
                    <Input
                      type="text"
                      placeholder="Username"
                      value={editedUsername}
                      onChange={(e) => setEditedUsername(e.target.value)}
                    />
                    <Input
                      type="password"
                      placeholder="Password"
                      value={editedPassword}
                      onChange={(e) => setEditedPassword(e.target.value)}
                    />
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {backend.backend_type === "grafana" && (
          <TabsContent value="datasources">
            <Card>
              <CardHeader>
                <CardTitle>Configure Data Source</CardTitle>
                <CardDescription>
                  Add a new data source to this Grafana instance
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="datasource-name">Data Source Name</Label>
                  <Input
                    id="datasource-name"
                    value={datasourceName}
                    onChange={(e) => setDatasourceName(e.target.value)}
                    placeholder="OTLP"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="datasource-type">Data Source Type</Label>
                  <Select value={datasourceType} onValueChange={setDatasourceType}>
                    <SelectTrigger id="datasource-type">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="otlp">OTLP</SelectItem>
                      <SelectItem value="prometheus">Prometheus</SelectItem>
                      <SelectItem value="loki">Loki</SelectItem>
                      <SelectItem value="tempo">Tempo</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="datasource-url">Data Source URL</Label>
                  <Input
                    id="datasource-url"
                    value={datasourceUrl}
                    onChange={(e) => setDatasourceUrl(e.target.value)}
                    placeholder="http://lawrence:4318"
                  />
                </div>

                <Button
                  onClick={handleConfigureDatasource}
                  disabled={!datasourceName || !datasourceUrl || configuringDatasource}
                >
                  {configuringDatasource && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                  Configure Data Source
                </Button>
              </CardContent>
            </Card>
          </TabsContent>
        )}
      </Tabs>

      <DelegateToAgent
        isOpen={showDelegateModal}
        onClose={() => setShowDelegateModal(false)}
        resourceType="backend"
        resourceId={backendId}
        resourceName={backend.name}
        onDelegated={async (runId, workId) => {
          // Reload backend to show new agent work
          await loadBackend();
        }}
      />
    </div>
  );
}

