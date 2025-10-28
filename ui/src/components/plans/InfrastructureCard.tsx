import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { StatusBadge } from "@/components/common/StatusBadge";
import type { InfrastructureComponent } from "@/types/plan";

interface InfrastructureCardProps {
  component: InfrastructureComponent;
  onAction?: () => void;
}

export function InfrastructureCard({ component, onAction }: InfrastructureCardProps) {
  const getComponentIcon = (type: string) => {
    switch (type) {
      case "database":
        return "🗄️";
      case "cache":
        return "💾";
      case "queue":
        return "📬";
      case "host":
        return "🖥️";
      default:
        return "⚙️";
    }
  };

  return (
    <Card className="cursor-pointer transition-colors hover:bg-accent" onClick={onAction}>
      <CardHeader>
        <div className="flex items-start justify-between">
          <div>
            <CardTitle className="text-lg flex items-center gap-2">
              <span>{getComponentIcon(component.component_type)}</span>
              {component.name}
            </CardTitle>
            <CardDescription className="mt-1">
              <div className="flex items-center gap-2">
                <span>{component.component_type}</span>
                {component.receiver_type && (
                  <Badge variant="outline">{component.receiver_type}</Badge>
                )}
              </div>
            </CardDescription>
          </div>
          <StatusBadge status={component.status} />
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {component.host && (
            <div className="text-sm">
              <span className="font-medium">Host:</span> {component.host}
            </div>
          )}
          {component.metrics_collected && (
            <div className="text-sm text-muted-foreground">
              <span className="font-medium">Metrics:</span> {component.metrics_collected}
            </div>
          )}
          <div className="text-xs text-muted-foreground">
            Updated: {new Date(component.updated_at).toLocaleString()}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

