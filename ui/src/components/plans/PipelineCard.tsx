import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { StatusBadge } from "@/components/common/StatusBadge";
import type { CollectorPipeline } from "@/types/plan";

interface PipelineCardProps {
  pipeline: CollectorPipeline;
  onAction?: () => void;
}

export function PipelineCard({ pipeline, onAction }: PipelineCardProps) {
  return (
    <Card className="cursor-pointer transition-colors hover:bg-accent" onClick={onAction}>
      <CardHeader>
        <div className="flex items-start justify-between">
          <div>
            <CardTitle className="text-lg">{pipeline.name}</CardTitle>
            <CardDescription className="mt-1">
              <div className="flex items-center gap-2">
                <span>Collector: {pipeline.collector_id}</span>
                {pipeline.target_type && <Badge variant="outline">{pipeline.target_type}</Badge>}
              </div>
            </CardDescription>
          </div>
          <StatusBadge status={pipeline.status} />
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {pipeline.rules && (
            <div className="text-sm text-muted-foreground">
              <span className="font-medium">Rules:</span> Configured
            </div>
          )}
          <div className="text-sm text-muted-foreground">
            <span className="font-medium">Config:</span> YAML defined
          </div>
          <div className="text-xs text-muted-foreground">
            Updated: {new Date(pipeline.updated_at).toLocaleString()}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

