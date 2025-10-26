import { useRuns } from "../../hooks/useRuns";
import { RunCard } from "./RunCard";
import { ScrollArea } from "@/components/ui/scroll-area";

interface RunListProps {
  selectedRunId?: string;
  onSelectRun: (runId: string) => void;
}

export function RunList({ selectedRunId, onSelectRun }: RunListProps) {
  const { runs, loading, error } = useRuns();

  if (loading && runs.length === 0) {
    return (
      <div className="p-4 text-center text-muted-foreground">
        Loading runs...
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 text-center text-destructive">Error: {error}</div>
    );
  }

  if (runs.length === 0) {
    return (
      <div className="p-4 text-center text-muted-foreground">
        No runs found. Start an agent to see runs here.
      </div>
    );
  }

  return (
    <ScrollArea className="h-full">
      {runs.map((run) => (
        <RunCard
          key={run.id}
          run={run}
          onClick={() => onSelectRun(run.id)}
          isActive={run.id === selectedRunId}
        />
      ))}
    </ScrollArea>
  );
}
