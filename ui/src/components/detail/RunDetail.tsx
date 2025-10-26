import { useRunStream } from "../../hooks/useRunStream";
import { useRun } from "../../hooks/useRun";
import { StatusBadge } from "../common/StatusBadge";
import { Timeline } from "./Timeline";
import { Conversation } from "./Conversation";
import { ToolExecutions } from "./ToolExecutions";
import { MetricsPanel } from "./MetricsPanel";
import { formatDate } from "../../utils/formatters";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { MessageSquare, Wrench, Clock, Square, Pause, Play } from "lucide-react";
import { apiClient } from "../../services/api";
import { useSWRConfig } from "swr";

interface RunDetailProps {
  runId: string;
}

export function RunDetail({ runId }: RunDetailProps) {
  const { run, loading } = useRun(runId);
  const { events, status, markMessageAsSent } = useRunStream(runId);
  const { mutate } = useSWRConfig();

  const handleStop = async () => {
    try {
      await apiClient.stopRun(runId);
      mutate(`/api/runs/${runId}`);
      mutate("/api/runs");
    } catch (error) {
      console.error("Failed to stop run:", error);
    }
  };

  const handlePause = async () => {
    try {
      await apiClient.pauseRun(runId);
      mutate(`/api/runs/${runId}`);
      mutate("/api/runs");
    } catch (error) {
      console.error("Failed to pause run:", error);
    }
  };

  const handleResume = async () => {
    try {
      await apiClient.resumeRun(runId);
      mutate(`/api/runs/${runId}`);
      mutate("/api/runs");
    } catch (error) {
      console.error("Failed to resume run:", error);
    }
  };


  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-muted-foreground">Loading run details...</div>
      </div>
    );
  }

  if (!run) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-destructive">Run not found</div>
      </div>
    );
  }

  const toolCallCount = events.filter((e) => e.type === "tool_call").length;

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="border-b border-border/50 bg-card px-4 py-3 space-y-2 flex-shrink-0">
        <div className="flex items-start justify-between gap-3">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <h1 className="text-base font-semibold truncate">{run.agent_name}</h1>
              <StatusBadge status={run.status} />
              <Badge variant={status === "connected" ? "default" : "outline"} className="text-[10px] h-4 px-1.5">
                {status === "connected" ? "● Live" : "○"}
              </Badge>
            </div>
            {/* Control buttons */}
            <div className="flex items-center gap-2 mt-2">
              {run.status === "running" && (
                <>
                  <Button size="sm" variant="outline" onClick={handlePause} className="h-7 text-xs">
                    <Pause className="h-3 w-3 mr-1" />
                    Pause
                  </Button>
                  <Button size="sm" variant="destructive" onClick={handleStop} className="h-7 text-xs">
                    <Square className="h-3 w-3 mr-1" />
                    Stop
                  </Button>
                </>
              )}
              {run.status === "paused" && (
                <>
                  <Button size="sm" variant="outline" onClick={handleResume} className="h-7 text-xs">
                    <Play className="h-3 w-3 mr-1" />
                    Resume
                  </Button>
                  <Button size="sm" variant="destructive" onClick={handleStop} className="h-7 text-xs">
                    <Square className="h-3 w-3 mr-1" />
                    Stop
                  </Button>
                </>
              )}
            </div>
            <p className="text-xs text-muted-foreground line-clamp-2 mb-1">{run.prompt}</p>
            <div className="flex items-center gap-3 text-[11px] text-muted-foreground">
              <span>{formatDate(run.start_time)}</span>
              <span>•</span>
              <span>{run.model}</span>
            </div>
          </div>
        </div>

        {/* Metrics */}
        <MetricsPanel run={run} />

        {/* Error Message */}
        {run.error && (
          <div className="bg-destructive/10 border border-destructive/20 rounded px-2 py-1.5">
            <div className="text-xs font-medium text-destructive mb-0.5">
              Error
            </div>
            <div className="text-xs text-destructive/90">{run.error}</div>
          </div>
        )}
      </div>

      {/* Tabs */}
      <Tabs defaultValue="conversation" className="flex-1 flex flex-col overflow-hidden">
        <TabsList className="w-full justify-start rounded-none border-b border-border/50 bg-background h-auto p-0 flex-shrink-0">
          <TabsTrigger
            value="conversation"
            className="rounded-none text-xs h-8 data-[state=active]:border-b-2 data-[state=active]:border-primary"
          >
            <MessageSquare className="h-3 w-3 mr-1.5" />
            Conversation
          </TabsTrigger>
          <TabsTrigger
            value="tools"
            className="rounded-none text-xs h-8 data-[state=active]:border-b-2 data-[state=active]:border-primary"
          >
            <Wrench className="h-3 w-3 mr-1.5" />
            Tools ({toolCallCount})
          </TabsTrigger>
          <TabsTrigger
            value="timeline"
            className="rounded-none text-xs h-8 data-[state=active]:border-b-2 data-[state=active]:border-primary"
          >
            <Clock className="h-3 w-3 mr-1.5" />
            Timeline ({events.length})
          </TabsTrigger>
        </TabsList>

        <TabsContent
          value="conversation"
          className="flex-1 m-0 p-4 flex flex-col min-h-0"
        >
          <Conversation events={events} runId={runId} runStatus={run.status} markMessageAsSent={markMessageAsSent} />
        </TabsContent>
        <TabsContent
          value="tools"
          className="flex-1 m-0 p-4 overflow-y-auto"
        >
          <ToolExecutions events={events} />
        </TabsContent>
        <TabsContent
          value="timeline"
          className="flex-1 m-0 p-4 overflow-y-auto"
        >
          <Timeline events={events} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
