import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { StatusBadge } from "@/components/common/StatusBadge";
import type { Backend } from "@/types/plan";

interface BackendCardProps {
  backend: Backend;
  onAction?: () => void;
}

export function BackendCard({ backend, onAction }: BackendCardProps) {
  const getBackendIcon = (type: string) => {
    switch (type) {
      case "grafana":
        return "📊";
      case "prometheus":
        return "📈";
      case "jaeger":
      case "tempo":
        return "🔍";
      default:
        return "🔗";
    }
  };

  const getHealthColor = (healthStatus: string) => {
    switch (healthStatus) {
      case "healthy":
        return "green";
      case "unhealthy":
        return "red";
      default:
        return "yellow";
    }
  };

  return (
    <Card className="cursor-pointer transition-colors hover:bg-accent" onClick={onAction}>
      <CardHeader>
        <div className="flex items-start justify-between">
          <div>
            <CardTitle className="text-lg flex items-center gap-2">
              <span>{getBackendIcon(backend.backend_type)}</span>
              {backend.name}
            </CardTitle>
            <CardDescription className="mt-1">
              <div className="flex items-center gap-2">
                <span>{backend.backend_type}</span>
                <Badge variant="outline" style={{ 
                  color: `var(--${getHealthColor(backend.health_status)}-600)` 
                }}>
                  {backend.health_status}
                </Badge>
              </div>
            </CardDescription>
          </div>
          <StatusBadge status={backend.health_status} />
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          <div className="text-sm">
            <span className="font-medium">URL:</span> {backend.url}
          </div>
          {backend.last_check && (
            <div className="text-sm text-muted-foreground">
              <span className="font-medium">Last check:</span>{" "}
              {new Date(backend.last_check).toLocaleString()}
            </div>
          )}
          {backend.datasource_uid && (
            <div className="text-sm text-muted-foreground">
              <span className="font-medium">Datasource UID:</span> {backend.datasource_uid}
            </div>
          )}
          <div className="text-xs text-muted-foreground">
            Updated: {new Date(backend.updated_at).toLocaleString()}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

