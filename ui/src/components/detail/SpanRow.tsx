import { useState } from "react";
import { ChevronRight, ChevronDown, AlertCircle } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import type { Span } from "../../types/trace";

interface SpanRowProps {
  span: Span;
  depth: number;
  startTime: number;
  totalDuration: number;
  onSpanClick?: (span: Span) => void;
}

const SPAN_COLORS = {
  tool: "bg-yellow-500",
  api_call: "bg-blue-500",
  agent_handoff: "bg-purple-500",
  iteration: "bg-green-500",
  trace: "bg-gray-500",
};

const SPAN_TEXT_COLORS = {
  tool: "text-yellow-700",
  api_call: "text-blue-700",
  agent_handoff: "text-purple-700",
  iteration: "text-green-700",
  trace: "text-gray-700",
};

export function SpanRow({ span, depth, startTime, totalDuration, onSpanClick }: SpanRowProps) {
  const [isExpanded, setIsExpanded] = useState(true);

  const spanStart = new Date(span.start_time).getTime();
  const spanEnd = span.end_time ? new Date(span.end_time).getTime() : spanStart;
  const spanDuration = spanEnd - spanStart;

  const relativeStart = (spanStart - startTime) / totalDuration;
  const relativeWidth = spanDuration / totalDuration;

  const hasChildren = span.children && span.children.length > 0;
  const isClickable = !!onSpanClick;

  const handleClick = () => {
    if (hasChildren) {
      setIsExpanded(!isExpanded);
    }
    if (onSpanClick) {
      onSpanClick(span);
    }
  };

  return (
    <>
      <div
        className={cn(
          "group flex items-center gap-2 py-1 hover:bg-muted/50 transition-colors cursor-pointer text-xs",
          isClickable && "cursor-pointer"
        )}
        style={{ paddingLeft: `${depth * 1.5}rem` }}
        onClick={handleClick}
      >
        {/* Expand/Collapse Icon */}
        <div className="w-4 h-4 flex items-center justify-center flex-shrink-0">
          {hasChildren ? (
            isExpanded ? (
              <ChevronDown className="h-3 w-3 text-muted-foreground" />
            ) : (
              <ChevronRight className="h-3 w-3 text-muted-foreground" />
            )
          ) : (
            <div className="w-3 h-3" />
          )}
        </div>

        {/* Span Info */}
        <div className="flex-1 flex items-center gap-2 min-w-0">
          <Badge
            variant="outline"
            className={cn(
              "text-[10px] h-4 px-1.5 font-medium",
              SPAN_TEXT_COLORS[span.type]
            )}
          >
            {span.type}
          </Badge>
          <span className="truncate font-medium">{span.name}</span>
          {span.error && (
            <AlertCircle className="h-3 w-3 text-destructive flex-shrink-0" />
          )}
        </div>

        {/* Duration */}
        <div className="text-[10px] text-muted-foreground whitespace-nowrap">
          {span.duration_ms ? `${span.duration_ms}ms` : "..."}
        </div>

        {/* Visual Bar */}
        <div className="relative h-6 flex-1 mx-2 max-w-2xl bg-muted rounded overflow-hidden">
          {/* Position indicator */}
          <div
            className={cn(
              "absolute top-0 bottom-0 opacity-30 group-hover:opacity-50 transition-opacity",
              SPAN_COLORS[span.type]
            )}
            style={{
              left: `${relativeStart * 100}%`,
              width: `${Math.max(relativeWidth * 100, 0.5)}%`,
            }}
          />
          {/* Active bar */}
          <div
            className={cn(
              "absolute top-0 bottom-0 transition-all",
              SPAN_COLORS[span.type],
              span.error && "bg-red-500"
            )}
            style={{
              left: `${relativeStart * 100}%`,
              width: `${Math.max(relativeWidth * 100, 0.5)}%`,
            }}
          />
        </div>

        {/* Timestamp */}
        <div className="text-[10px] text-muted-foreground whitespace-nowrap">
          {new Date(span.start_time).toLocaleTimeString('en-US', {
            hour12: false,
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
            fractionalSecondDigits: 3,
          })}
        </div>
      </div>

      {/* Children */}
      {hasChildren && isExpanded && (
        <div>
          {span.children!.map((child) => (
            <SpanRow
              key={child.id}
              span={child}
              depth={depth + 1}
              startTime={startTime}
              totalDuration={totalDuration}
              onSpanClick={onSpanClick}
            />
          ))}
        </div>
      )}
    </>
  );
}
