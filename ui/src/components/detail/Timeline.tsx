import type { AgentEvent } from "../../types/events";
import { formatTimestamp } from "../../utils/formatters";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { 
  Play, 
  Square, 
  RotateCw, 
  MessageCircle, 
  Zap, 
  Check, 
  ArrowRight, 
  ArrowLeft, 
  AlertTriangle,
  GitMerge
} from "lucide-react";

interface TimelineProps {
  events: AgentEvent[];
}

const EVENT_ICONS: Record<string, React.ComponentType<any>> = {
  run_start: Play,
  run_end: Square,
  iteration: RotateCw,
  message: MessageCircle,
  tool_call: Zap,
  tool_result: Check,
  api_request: ArrowRight,
  api_response: ArrowLeft,
  error: AlertTriangle,
  agent_handoff: GitMerge,
  agent_handoff_complete: Check,
};

const EVENT_COLORS: Record<string, string> = {
  run_start: "text-blue-600 dark:text-blue-400",
  run_end: "text-gray-600 dark:text-gray-400",
  iteration: "text-purple-600 dark:text-purple-400",
  message: "text-slate-600 dark:text-slate-400",
  tool_call: "text-amber-600 dark:text-amber-400",
  tool_result: "text-green-600 dark:text-green-400",
  api_request: "text-cyan-600 dark:text-cyan-400",
  api_response: "text-cyan-600 dark:text-cyan-400",
  error: "text-red-600 dark:text-red-400",
  agent_handoff: "text-violet-600 dark:text-violet-400",
  agent_handoff_complete: "text-violet-600 dark:text-violet-400",
};

const EVENT_VARIANTS: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  run_start: "default",
  run_end: "secondary",
  iteration: "default",
  message: "outline",
  tool_call: "outline",
  tool_result: "secondary",
  api_request: "outline",
  api_response: "secondary",
  error: "destructive",
  agent_handoff: "default",
  agent_handoff_complete: "secondary",
};

export function Timeline({ events }: TimelineProps) {
  if (events.length === 0) {
    return (
      <div className="flex items-center justify-center h-32 text-sm text-muted-foreground">
        No events yet
      </div>
    );
  }

  return (
    <div className="relative">
      {/* Timeline line */}
      <div className="absolute left-4 top-0 bottom-0 w-px bg-border/50" />
      
      <div className="space-y-0">
        {events.map((event) => {
          const IconComponent = EVENT_ICONS[event.type];
          const color = EVENT_COLORS[event.type] || "text-muted-foreground";
          const variant = EVENT_VARIANTS[event.type] || "outline";
          
          return (
            <div 
              key={event.id} 
              className="relative flex items-start gap-3 py-3 group hover:bg-muted/30 transition-colors rounded-md -mx-2 px-2"
            >
              {/* Timeline dot */}
              <div className="relative z-10 flex-shrink-0 mt-0.5">
                <div className={cn(
                  "w-8 h-8 rounded-full flex items-center justify-center text-xs font-medium",
                  "bg-background border-2 shadow-sm",
                  color === "text-blue-600" || color === "text-blue-400" ? "border-blue-500 text-blue-600 dark:text-blue-400" :
                  color === "text-purple-600" || color === "text-purple-400" ? "border-purple-500 text-purple-600 dark:text-purple-400" :
                  color === "text-amber-600" || color === "text-amber-400" ? "border-amber-500 text-amber-600 dark:text-amber-400" :
                  color === "text-green-600" || color === "text-green-400" ? "border-green-500 text-green-600 dark:text-green-400" :
                  color === "text-red-600" || color === "text-red-400" ? "border-red-500 text-red-600 dark:text-red-400" :
                  "border-border text-muted-foreground"
                )}>
                  {IconComponent ? <IconComponent className="h-3.5 w-3.5" /> : "•"}
                </div>
              </div>

              {/* Event content */}
              <div className="flex-1 min-w-0 space-y-1.5">
                <div className="flex items-start justify-between gap-3">
                  <div className="flex items-center gap-2 flex-wrap">
                    <Badge
                      variant={variant}
                      className="text-[10px] h-5 px-2 font-normal rounded-md"
                    >
                      {event.type.replace(/_/g, " ")}
                    </Badge>
                  </div>
                  <span className="text-[10px] text-muted-foreground whitespace-nowrap font-mono tabular-nums">
                    {formatTimestamp(event.timestamp)}
                  </span>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
