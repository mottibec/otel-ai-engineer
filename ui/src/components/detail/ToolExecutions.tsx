import type {
  AgentEvent,
  ToolCallData,
  ToolResultData,
} from "../../types/events";
import { JsonViewer } from "../common/JsonViewer";
import { formatTimestamp } from "../../utils/formatters";
import { useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ChevronDown, ChevronRight } from "lucide-react";

interface ToolExecutionsProps {
  events: AgentEvent[];
}

interface ToolExecution {
  callEvent: AgentEvent;
  resultEvent?: AgentEvent;
}

export function ToolExecutions({ events }: ToolExecutionsProps) {
  const [expandedTools, setExpandedTools] = useState<Set<string>>(new Set());

  // Group tool calls with their results
  const toolExecutions: ToolExecution[] = [];
  const toolCalls = events.filter((e) => e.type === "tool_call");
  const toolResults = events.filter((e) => e.type === "tool_result");

  toolCalls.forEach((callEvent) => {
    const callData = callEvent.data as ToolCallData;
    const resultEvent = toolResults.find((r) => {
      const resultData = r.data as ToolResultData;
      return resultData.tool_use_id === callData.tool_use_id;
    });

    toolExecutions.push({ callEvent, resultEvent });
  });

  const toggleTool = (id: string) => {
    const newExpanded = new Set(expandedTools);
    if (newExpanded.has(id)) {
      newExpanded.delete(id);
    } else {
      newExpanded.add(id);
    }
    setExpandedTools(newExpanded);
  };

  if (toolExecutions.length === 0) {
    return (
      <div className="text-center text-muted-foreground py-8">
        No tool executions yet
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {toolExecutions.map((execution, index) => {
        const callData = execution.callEvent.data as ToolCallData;
        const resultData = execution.resultEvent?.data as
          | ToolResultData
          | undefined;
        const isExpanded = expandedTools.has(callData.tool_use_id);
        const isError = resultData?.is_error ?? false;

        return (
          <div
            key={callData.tool_use_id}
            className={`border rounded ${isError ? "border-destructive/20 bg-destructive/5" : "border-border/50"}`}
          >
            <div
              className="px-3 py-2 cursor-pointer hover:bg-accent/30 transition-colors"
              onClick={() => toggleTool(callData.tool_use_id)}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 flex-1 min-w-0">
                  {isExpanded ? (
                    <ChevronDown className="h-3 w-3 flex-shrink-0" />
                  ) : (
                    <ChevronRight className="h-3 w-3 flex-shrink-0" />
                  )}
                  <div className="flex-1 min-w-0">
                    <div className="text-xs font-medium truncate">
                      {index + 1}. {callData.tool_name}
                    </div>
                    <div className="text-[10px] text-muted-foreground">
                      {formatTimestamp(execution.callEvent.timestamp)}
                      {resultData && ` • ${resultData.duration}`}
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-1.5 flex-shrink-0">
                  {resultData && (
                    <Badge variant={isError ? "destructive" : "secondary"} className="text-[10px] h-4 px-1.5">
                      {isError ? "✗" : "✓"}
                    </Badge>
                  )}
                  {!resultData && (
                    <Badge variant="outline" className="text-[10px] h-4 px-1.5 bg-yellow-50 text-yellow-800 border-yellow-200">
                      ⏳
                    </Badge>
                  )}
                </div>
              </div>
            </div>

            {isExpanded && (
              <div className="border-t border-border/50 px-3 py-2 bg-muted/30 space-y-2.5">
                <div>
                  <h4 className="text-xs font-medium mb-1">Input</h4>
                  <JsonViewer data={callData.input} />
                </div>

                {resultData && (
                  <div>
                    <h4 className="text-xs font-medium mb-1">
                      {isError ? "Error" : "Result"}
                    </h4>
                    {isError ? (
                      <div className="border border-destructive/20 bg-destructive/10 rounded px-2 py-1.5 text-xs text-destructive">
                        {resultData.error}
                      </div>
                    ) : (
                      <JsonViewer data={resultData.result} />
                    )}
                  </div>
                )}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}
