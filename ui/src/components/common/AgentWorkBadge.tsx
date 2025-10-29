import type { AgentWork } from "../../types/agent-work";
import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Loader2 } from "lucide-react";

interface AgentWorkBadgeProps {
  work: AgentWork;
  onClick?: (runId: string) => void;
  showTooltip?: boolean;
}

export function AgentWorkBadge({ work, onClick, showTooltip = true }: AgentWorkBadgeProps) {
  const getStatusVariant = () => {
    switch (work.status) {
      case "running":
        return "default";
      case "completed":
        return "secondary";
      case "failed":
        return "destructive";
      case "cancelled":
        return "outline";
      default:
        return "outline";
    }
  };

  const getStatusIcon = () => {
    if (work.status === "running") {
      return <Loader2 className="h-3 w-3 animate-spin inline" />;
    }
    switch (work.status) {
      case "completed":
        return "✓";
      case "failed":
        return "✗";
      case "cancelled":
        return "○";
      default:
        return "?";
    }
  };

  const badge = (
    <Badge
      variant={getStatusVariant()}
      className={`text-xs h-5 px-2 ${onClick && work.run_id ? "cursor-pointer hover:opacity-80 transition-opacity" : "cursor-help"}`}
      onClick={onClick && work.run_id ? () => onClick(work.run_id) : undefined}
    >
      {getStatusIcon()} {work.agent_name}
    </Badge>
  );

  if (!showTooltip) {
    return badge;
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          {badge}
        </TooltipTrigger>
        <TooltipContent>
          <div className="space-y-1">
            <div className="font-semibold">{work.agent_name}</div>
            <div className="text-xs text-muted-foreground">{work.task_description}</div>
            <div className="text-xs text-muted-foreground">Status: {work.status}</div>
            {work.started_at && (
              <div className="text-xs text-muted-foreground">
                Started: {new Date(work.started_at).toLocaleString()}
              </div>
            )}
            {work.completed_at && (
              <div className="text-xs text-muted-foreground">
                Completed: {new Date(work.completed_at).toLocaleString()}
              </div>
            )}
            {work.error && (
              <div className="text-xs text-red-500">Error: {work.error}</div>
            )}
            {onClick && work.run_id && (
              <div className="text-xs text-blue-500 mt-1">Click to view run</div>
            )}
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

interface AgentWorkBadgesProps {
  works?: AgentWork[];
  onClick?: (runId: string) => void;
  maxVisible?: number;
}

export function AgentWorkBadges({ works, onClick, maxVisible = 3 }: AgentWorkBadgesProps) {
  if (!works || works.length === 0) {
    return null;
  }

  // Sort: running first, then by started_at descending
  const sortedWorks = [...works].sort((a, b) => {
    if (a.status === "running" && b.status !== "running") return -1;
    if (a.status !== "running" && b.status === "running") return 1;
    return new Date(b.started_at).getTime() - new Date(a.started_at).getTime();
  });

  const visibleWorks = sortedWorks.slice(0, maxVisible);
  const remainingCount = sortedWorks.length - visibleWorks.length;

  return (
    <div className="flex items-center gap-1 flex-wrap">
      {visibleWorks.map((work) => (
        <AgentWorkBadge key={work.id} work={work} onClick={onClick} />
      ))}
      {remainingCount > 0 && (
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <Badge variant="outline" className="text-xs h-5 px-2 cursor-help">
                +{remainingCount}
              </Badge>
            </TooltipTrigger>
            <TooltipContent>
              <div className="text-xs">
                {remainingCount} more agent {remainingCount === 1 ? "task" : "tasks"}
              </div>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      )}
    </div>
  );
}

