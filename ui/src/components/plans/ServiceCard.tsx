import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { StatusBadge } from "@/components/common/StatusBadge";
import { Button } from "@/components/ui/button";
import { DelegateToAgent } from "@/components/common/DelegateToAgent";
import { UserCog } from "lucide-react";
import type { InstrumentedService } from "@/types/plan";

interface ServiceCardProps {
  service: InstrumentedService;
  onAction?: () => void;
}

export function ServiceCard({ service, onAction }: ServiceCardProps) {
  const [showDelegate, setShowDelegate] = useState(false);

  return (
    <>
      <Card className="transition-colors hover:bg-accent">
        <CardHeader>
          <div className="flex items-start justify-between">
            <div className="flex-1" onClick={onAction}>
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
            {service.git_repo_url && (
              <div className="text-sm">
                <span className="font-medium">Git Repo:</span>{" "}
                <a
                  href={service.git_repo_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 hover:underline truncate block"
                  onClick={(e) => e.stopPropagation()}
                >
                  {service.git_repo_url}
                </a>
              </div>
            )}
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
            <div className="flex items-center justify-between mt-3">
              <div className="text-xs text-muted-foreground">
                Updated: {new Date(service.updated_at).toLocaleString()}
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
        resourceType="service"
        resourceId={service.id}
        resourceName={service.service_name}
      />
    </>
  );
}

