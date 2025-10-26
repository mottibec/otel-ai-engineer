import { useState } from "react";
import useSWR from "swr";
import { apiClient } from "../../services/api";
import { SpanRow } from "./SpanRow";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import type { Span } from "../../types/trace";

interface TraceViewProps {
  runId: string;
}

export function TraceView({ runId }: TraceViewProps) {
  const [selectedSpan, setSelectedSpan] = useState<Span | null>(null);
  
  const { data: trace, error, isLoading } = useSWR(
    `/api/runs/${runId}/trace`,
    () => apiClient.getRunTrace(runId),
    {
      refreshInterval: 2000, // Refresh every 2 seconds for live updates
    }
  );

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-full" />
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        Failed to load trace: {error instanceof Error ? error.message : "Unknown error"}
      </div>
    );
  }

  if (!trace) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        No trace data available
      </div>
    );
  }

  const rootSpan = trace.root_span;
  const traceStart = new Date(trace.start_time).getTime();
  const traceEnd = trace.end_time ? new Date(trace.end_time).getTime() : Date.now();
  const totalDuration = Math.max(traceEnd - traceStart, 1);

  return (
    <div className="space-y-4">
      {/* Trace Overview */}
      <Card className="p-4">
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <h3 className="font-semibold text-sm">Trace Overview</h3>
            <div className="text-xs text-muted-foreground">
              {trace.duration_ms}ms
            </div>
          </div>
          <div className="grid grid-cols-3 gap-4 text-xs">
            <div>
              <div className="text-muted-foreground">Trace ID</div>
              <div className="font-mono truncate">{trace.trace_id}</div>
            </div>
            <div>
              <div className="text-muted-foreground">Start Time</div>
              <div>{new Date(trace.start_time).toLocaleString()}</div>
            </div>
            <div>
              <div className="text-muted-foreground">Status</div>
              <div>{trace.end_time ? "Completed" : "Running"}</div>
            </div>
          </div>
        </div>
      </Card>

      {/* Waterfall Visualization */}
      <Card className="overflow-hidden">
        <div className="bg-muted/50 border-b px-4 py-2">
          <div className="flex items-center gap-2 text-xs font-medium">
            <div className="w-4 h-4" /> {/* Spacer for depth icons */}
            <div className="flex-1">Span</div>
            <div className="w-20">Duration</div>
            <div className="flex-1 mx-2">Timeline</div>
            <div className="w-32">Time</div>
          </div>
        </div>
        
        <div className="divide-y divide-border">
          <SpanRow
            span={rootSpan}
            depth={0}
            startTime={traceStart}
            totalDuration={totalDuration}
            onSpanClick={setSelectedSpan}
          />
        </div>
      </Card>

      {/* Selected Span Details */}
      {selectedSpan && (
        <Card className="p-4">
          <div className="space-y-3">
            <h3 className="font-semibold text-sm">Span Details</h3>
            
            <div className="grid grid-cols-2 gap-4 text-xs">
              <div>
                <div className="text-muted-foreground">Name</div>
                <div className="font-medium">{selectedSpan.name}</div>
              </div>
              <div>
                <div className="text-muted-foreground">Type</div>
                <div>{selectedSpan.type}</div>
              </div>
              <div>
                <div className="text-muted-foreground">Duration</div>
                <div>{selectedSpan.duration || "N/A"}</div>
              </div>
              <div>
                <div className="text-muted-foreground">Status</div>
                <div>{selectedSpan.error ? "Error" : "Success"}</div>
              </div>
            </div>

            {selectedSpan.error && selectedSpan.error_msg && (
              <div className="text-xs">
                <div className="text-muted-foreground mb-1">Error</div>
                <div className="text-destructive bg-destructive/10 p-2 rounded">
                  {selectedSpan.error_msg}
                </div>
              </div>
            )}

            {selectedSpan.tags && Object.keys(selectedSpan.tags).length > 0 && (
              <div className="text-xs">
                <div className="text-muted-foreground mb-1">Tags</div>
                <div className="space-y-1">
                  {Object.entries(selectedSpan.tags).map(([key, value]) => (
                    <div key={key} className="flex gap-2">
                      <span className="font-medium">{key}:</span>
                      <span className="text-muted-foreground">
                        {typeof value === "object" 
                          ? JSON.stringify(value) 
                          : String(value)}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </Card>
      )}
    </div>
  );
}
