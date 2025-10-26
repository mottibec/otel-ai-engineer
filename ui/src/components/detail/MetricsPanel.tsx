import type { Run } from "../../types/models";
import { formatNumber, formatDuration } from "../../utils/formatters";
import { Card, CardContent } from "@/components/ui/card";

interface MetricsPanelProps {
  run: Run;
}

export function MetricsPanel({ run }: MetricsPanelProps) {
  return (
    <div className="flex items-center gap-6 text-xs">
      <div>
        <span className="text-muted-foreground">Iterations:</span>{" "}
        <span className="font-semibold">{run.total_iterations}</span>
      </div>
      <div>
        <span className="text-muted-foreground">Tools:</span>{" "}
        <span className="font-semibold">{run.total_tool_calls}</span>
      </div>
      <div>
        <span className="text-muted-foreground">Tokens:</span>{" "}
        <span className="font-semibold">
          {formatNumber(run.total_tokens.total_tokens)}
        </span>
        <span className="text-muted-foreground/70 ml-1">
          (↓{formatNumber(run.total_tokens.input_tokens)} ↑
          {formatNumber(run.total_tokens.output_tokens)})
        </span>
      </div>
      <div>
        <span className="text-muted-foreground">Duration:</span>{" "}
        <span className="font-semibold">
          {run.duration ? formatDuration(run.duration) : "-"}
        </span>
      </div>
    </div>
  );
}
