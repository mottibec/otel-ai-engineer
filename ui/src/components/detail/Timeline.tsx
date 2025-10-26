import type { AgentEvent } from "../../types/events";
import { formatTimestamp } from "../../utils/formatters";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

interface TimelineProps {
  events: AgentEvent[];
}

export function Timeline({ events }: TimelineProps) {
  const getEventIcon = (type: string) => {
    switch (type) {
      case "run_start":
        return "▶️";
      case "run_end":
        return "⏹️";
      case "iteration":
        return "🔄";
      case "message":
        return "💬";
      case "tool_call":
        return "🔧";
      case "tool_result":
        return "✅";
      case "api_request":
        return "📤";
      case "api_response":
        return "📥";
      case "error":
        return "❌";
      default:
        return "•";
    }
  };

  const getEventVariant = (
    type: string,
  ): "default" | "secondary" | "destructive" | "outline" => {
    switch (type) {
      case "run_start":
      case "tool_result":
        return "secondary";
      case "error":
        return "destructive";
      case "message":
      case "iteration":
        return "default";
      default:
        return "outline";
    }
  };

  const getEventColorClass = (type: string) => {
    switch (type) {
      case "run_start":
      case "tool_result":
        return "border-green-500 bg-green-50";
      case "run_end":
        return "border-border bg-muted";
      case "iteration":
        return "border-primary bg-primary/10";
      case "message":
        return "border-purple-500 bg-purple-50";
      case "tool_call":
        return "border-yellow-500 bg-yellow-50";
      case "error":
        return "border-destructive bg-destructive/10";
      default:
        return "border-border bg-muted";
    }
  };

  return (
    <div className="space-y-1">
      {events.map((event, index) => (
        <div key={event.id} className="flex items-start gap-2">
          <div className="flex flex-col items-center flex-shrink-0">
            <div
              className={cn(
                "w-5 h-5 rounded-full border flex items-center justify-center text-[10px]",
                getEventColorClass(event.type),
              )}
            >
              {getEventIcon(event.type)}
            </div>
            {index < events.length - 1 && (
              <div className="w-px h-4 bg-border/50 mt-0.5" />
            )}
          </div>
          <div className="flex-1 pb-2 min-w-0">
            <div className="flex items-center justify-between gap-2">
              <Badge
                variant={getEventVariant(event.type)}
                className="text-[10px] h-4 px-1.5 font-normal"
              >
                {event.type.replace("_", " ")}
              </Badge>
              <span className="text-[10px] text-muted-foreground whitespace-nowrap">
                {formatTimestamp(event.timestamp)}
              </span>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
