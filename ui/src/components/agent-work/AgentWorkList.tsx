import { useState } from "react";
import { useSWRConfig } from "swr";
import { apiClient } from "@/services/api";
import type { AgentWork, ResourceType } from "@/types/agent-work";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Bot,
  Loader2,
  CheckCircle2,
  XCircle,
  Ban,
  ExternalLink,
  Clock,
} from "lucide-react";
import useSWR from "swr";
import { AgentWorkBadge } from "@/components/common/AgentWorkBadge";

interface AgentWorkListProps {
  onSelectRun?: (runId: string) => void;
}

const fetcher = () => apiClient.listAgentWork({ limit: 100 });

export function AgentWorkList({ onSelectRun }: AgentWorkListProps) {
  const { mutate } = useSWRConfig();
  const [statusFilter, setStatusFilter] = useState<string>("all");

  const { data: allWorks = [], error, isLoading } = useSWR<AgentWork[]>(
    "/api/agent-work",
    fetcher,
    {
      refreshInterval: 3000, // Poll every 3 seconds for real-time updates
      revalidateOnFocus: true,
    }
  );

  // Ensure allWorks is always an array (handle null/undefined responses)
  const safeWorks = Array.isArray(allWorks) ? allWorks : [];

  const handleCancelWork = async (workId: string) => {
    if (!window.confirm("Are you sure you want to cancel this agent work?")) {
      return;
    }

    try {
      await apiClient.cancelAgentWork(workId);
      // Invalidate and refetch
      mutate("/api/agent-work");
      // Also invalidate resource-specific caches
      const work = safeWorks.find((w) => w.id === workId);
      if (work) {
        mutate(`/api/agent-work/resource/${work.resource_type}/${work.resource_id}`);
        mutate(`/api/${work.resource_type}s/${work.resource_id}`);
        mutate(`/api/${work.resource_type}s`);
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to cancel agent work");
    }
  };

  // Filter works by status
  const filteredWorks = statusFilter === "all" 
    ? safeWorks 
    : safeWorks.filter((w) => w.status === statusFilter);

  // Group by status: running first, then others
  const sortedWorks = [...filteredWorks].sort((a, b) => {
    if (a.status === "running" && b.status !== "running") return -1;
    if (a.status !== "running" && b.status === "running") return 1;
    return new Date(b.started_at).getTime() - new Date(a.started_at).getTime();
  });

  const runningWorks = sortedWorks.filter((w) => w.status === "running");
  const completedWorks = sortedWorks.filter((w) => w.status === "completed");
  const failedWorks = sortedWorks.filter((w) => w.status === "failed");
  const cancelledWorks = sortedWorks.filter((w) => w.status === "cancelled");

  const getResourceTypeLabel = (type: ResourceType): string => {
    switch (type) {
      case "collector":
        return "Collector";
      case "backend":
        return "Backend";
      case "service":
        return "Service";
      case "infrastructure":
        return "Infrastructure";
      case "pipeline":
        return "Pipeline";
      case "plan":
        return "Plan";
      default:
        return type;
    }
  };

  const getResourceTypeIcon = (type: ResourceType) => {
    switch (type) {
      case "collector":
        return "📡";
      case "backend":
        return "🗄️";
      case "service":
        return "⚙️";
      case "infrastructure":
        return "🏗️";
      case "pipeline":
        return "🔀";
      case "plan":
        return "📋";
      default:
        return "📦";
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "running":
        return "text-blue-500";
      case "completed":
        return "text-green-500";
      case "failed":
        return "text-red-500";
      case "cancelled":
        return "text-gray-500";
      default:
        return "text-gray-500";
    }
  };

  if (isLoading) {
    return (
      <div className="p-6 space-y-4">
        <Skeleton className="h-8 w-64" />
        <div className="space-y-2">
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-24 w-full" />
        </div>
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
            <p className="text-sm text-muted-foreground">
              {error instanceof Error ? error.message : "Failed to load agent work"}
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-4 bg-background">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <Bot className="h-8 w-8" />
            Agent Work
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            Monitor and manage tasks assigned to agents
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Select value={statusFilter} onValueChange={setStatusFilter}>
            <SelectTrigger className="w-40">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Status</SelectItem>
              <SelectItem value="running">Running</SelectItem>
              <SelectItem value="completed">Completed</SelectItem>
              <SelectItem value="failed">Failed</SelectItem>
              <SelectItem value="cancelled">Cancelled</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Summary Stats */}
      <div className="grid grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Running</p>
                <p className="text-2xl font-bold text-blue-500">{runningWorks.length}</p>
              </div>
              <Loader2 className="h-8 w-8 text-blue-500 animate-spin" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Completed</p>
                <p className="text-2xl font-bold text-green-500">{completedWorks.length}</p>
              </div>
              <CheckCircle2 className="h-8 w-8 text-green-500" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Failed</p>
                <p className="text-2xl font-bold text-red-500">{failedWorks.length}</p>
              </div>
              <XCircle className="h-8 w-8 text-red-500" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Cancelled</p>
                <p className="text-2xl font-bold text-gray-500">{cancelledWorks.length}</p>
              </div>
              <Ban className="h-8 w-8 text-gray-500" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Work List */}
      {sortedWorks.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <Bot className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
            <CardTitle className="mb-2">No agent work found</CardTitle>
            <CardDescription>
              {statusFilter === "all"
                ? "No agent work has been created yet. Delegate tasks to agents to see them here."
                : `No agent work with status "${statusFilter}" found.`}
            </CardDescription>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-2">
          {sortedWorks.map((work) => (
            <Card key={work.id} className="hover:bg-accent transition-colors">
              <CardContent className="pt-6">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0 space-y-2">
                    <div className="flex items-center gap-2">
                      <span className="text-2xl">{getResourceTypeIcon(work.resource_type)}</span>
                      <Badge variant="outline" className="text-xs">
                        {getResourceTypeLabel(work.resource_type)}
                      </Badge>
                      <span className="text-sm text-muted-foreground font-mono">
                        {work.resource_id}
                      </span>
                    </div>
                    <div>
                      <div className="flex items-center gap-2 mb-1">
                        <AgentWorkBadge work={work} showTooltip={false} onClick={onSelectRun} />
                      </div>
                      <p className="text-sm text-muted-foreground">{work.task_description}</p>
                    </div>
                    <div className="flex items-center gap-4 text-xs text-muted-foreground">
                      <div className="flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        Started: {new Date(work.started_at).toLocaleString()}
                      </div>
                      {work.completed_at && (
                        <div className="flex items-center gap-1">
                          <CheckCircle2 className="h-3 w-3" />
                          Completed: {new Date(work.completed_at).toLocaleString()}
                        </div>
                      )}
                      {work.error && (
                        <div className={`flex items-center gap-1 ${getStatusColor("failed")}`}>
                          <XCircle className="h-3 w-3" />
                          Error: {work.error}
                        </div>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-2 flex-shrink-0">
                    {work.status === "running" && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleCancelWork(work.id)}
                      >
                        <Ban className="h-4 w-4 mr-1" />
                        Cancel
                      </Button>
                    )}
                    {work.run_id && onSelectRun && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => onSelectRun(work.run_id)}
                      >
                        <ExternalLink className="h-4 w-4 mr-1" />
                        View Run
                      </Button>
                    )}
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

