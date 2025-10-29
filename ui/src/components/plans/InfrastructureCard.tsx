import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { StatusBadge } from "@/components/common/StatusBadge";
import { Button } from "@/components/ui/button";
import { DelegateToAgent } from "@/components/common/DelegateToAgent";
import { UserCog } from "lucide-react";
import type { InfrastructureComponent } from "@/types/plan";

interface InfrastructureCardProps {
  component: InfrastructureComponent;
  onAction?: () => void;
}

export function InfrastructureCard({ component, onAction }: InfrastructureCardProps) {
  const [showDelegate, setShowDelegate] = useState(false);

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
    <>
      <Card className="transition-colors hover:bg-accent">
        <CardHeader>
          <div className="flex items-start justify-between">
            <div className="flex-1" onClick={onAction}>
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
            <div className="flex items-center justify-between mt-3">
              <div className="text-xs text-muted-foreground">
                Updated: {new Date(component.updated_at).toLocaleString()}
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={(e) => {
                  e.stopPropagation();
                  setShowDelegate(true);
                }}
              >
                <UserCog className="h-3 w-3 mr-1" />
                Delegate
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      <DelegateToAgent
        isOpen={showDelegate}
        onClose={() => setShowDelegate(false)}
        resourceType="infrastructure"
        resourceId={component.id}
        resourceName={component.name}
      />
    </>
  );
}

