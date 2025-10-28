import type { Run } from "../../types/models";
import { StatusBadge } from "../common/StatusBadge";
import {
  formatDate,
  formatDuration,
  formatNumber,
} from "../../utils/formatters";
import { cn } from "@/lib/utils";

interface RunCardProps {
  run: Run;
  onClick: () => void;
  isActive?: boolean;
}

export function RunCard({ run, onClick, isActive = false }: RunCardProps) {
  return (
    <div
      onClick={onClick}
      className={cn(
        "px-3 py-2 cursor-pointer transition-colors hover:bg-accent/50 border-b border-border/50",
        isActive && "bg-accent/70 border-l-2 border-l-primary",
      )}
    >
      <div className="flex items-start justify-between gap-2 mb-1">
        <h3 className="font-medium text-xs truncate flex-1">
          {run.agent_name}
        </h3>
        <StatusBadge status={run.status} />
      </div>

      <p className="text-[11px] text-muted-foreground line-clamp-1 mb-1.5">
        {run.prompt}
      </p>

      <div className="flex items-center justify-between text-[10px] text-muted-foreground mb-1">
        <span>{formatDate(run.start_time)}</span>
        {run.duration && <span>{formatDuration(run.duration)}</span>}
      </div>

      <div className="flex items-center gap-1.5">
        <span className="text-[10px] text-muted-foreground">
          {run.total_iterations} iter
        </span>
        <span className="text-[10px] text-muted-foreground/50">•</span>
        <span className="text-[10px] text-muted-foreground">
          {run.total_tool_calls} tools
        </span>
        <span className="text-[10px] text-muted-foreground/50">•</span>
        <span className="text-[10px] text-muted-foreground">
          {formatNumber(run.total_tokens.total_tokens)} tok
        </span>
      </div>
    </div>
  );
}
