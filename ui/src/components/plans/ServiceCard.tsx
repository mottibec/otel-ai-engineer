import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { StatusBadge } from "@/components/common/StatusBadge";
import type { InstrumentedService } from "@/types/plan";

interface ServiceCardProps {
  service: InstrumentedService;
  onAction?: () => void;
}

export function ServiceCard({ service, onAction }: ServiceCardProps) {
  return (
    <Card className="cursor-pointer transition-colors hover:bg-accent" onClick={onAction}>
      <CardHeader>
        <div className="flex items-start justify-between">
          <div>
            <CardTitle className="text-lg">{service.service_name}</CardTitle>
            <CardDescription className="mt-1">
              <div className="flex items-center gap-2">
                <span>{service.language}</span>
                {service.framework && <Badge variant="outline">{service.framework}</Badge>}
              </div>
            </CardDescription>
          </div>
          <StatusBadge status={service.status} />
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {service.sdk_version && (
            <div className="text-sm">
              <span className="font-medium">SDK:</span> {service.sdk_version}
            </div>
          )}
          {service.exporter_endpoint && (
            <div className="text-sm text-muted-foreground">
              <span className="font-medium">Endpoint:</span> {service.exporter_endpoint}
            </div>
          )}
          {service.code_changes_summary && (
            <div className="text-sm text-muted-foreground">
              <span className="font-medium">Changes:</span> {service.code_changes_summary}
            </div>
          )}
          <div className="text-xs text-muted-foreground">
            Updated: {new Date(service.updated_at).toLocaleString()}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

