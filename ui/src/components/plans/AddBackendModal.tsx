import { useState, useEffect } from "react";
import { apiClient } from "@/services/api";
import type { Backend } from "@/types/backend";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Loader2, Database, CheckCircle2 } from "lucide-react";
import { BackendHealthIndicator } from "@/components/backends/BackendHealthIndicator";

interface AddBackendModalProps {
  isOpen: boolean;
  onClose: () => void;
  planId: string;
  existingBackendIds: string[];
  onBackendAdded: () => void;
}

export function AddBackendModal({
  isOpen,
  onClose,
  planId,
  existingBackendIds,
  onBackendAdded,
}: AddBackendModalProps) {
  const [backends, setBackends] = useState<Backend[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [attachingId, setAttachingId] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      loadBackends();
    }
  }, [isOpen]);

  const loadBackends = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.listBackends();
      setBackends(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load backends");
    } finally {
      setLoading(false);
    }
  };

  const handleAttachBackend = async (backendId: string) => {
    try {
      setAttachingId(backendId);
      setError(null);
      await apiClient.attachBackendToPlan(planId, backendId);
      onBackendAdded();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to attach backend");
    } finally {
      setAttachingId(null);
    }
  };

  // Filter out backends already in the plan
  const availableBackends = backends.filter(
    (backend) => !existingBackendIds.includes(backend.id),
  );

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Add Backend to Plan</DialogTitle>
          <DialogDescription>
            Select an existing backend to associate with this plan
          </DialogDescription>
        </DialogHeader>

        {error && (
          <div className="p-3 text-sm text-red-500 bg-red-500/10 border border-red-500/20 rounded-md">
            {error}
          </div>
        )}

        {loading ? (
          <div className="space-y-4">
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-32 w-full" />
            ))}
          </div>
        ) : availableBackends.length === 0 ? (
          <Card>
            <CardContent className="py-12 text-center">
              <Database className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
              <CardTitle className="mb-2">No available backends</CardTitle>
              <CardDescription>
                All backends are already associated with this plan, or no backends exist yet.
                Create a new backend first.
              </CardDescription>
            </CardContent>
          </Card>
        ) : (
          <div className="grid gap-4 md:grid-cols-2">
            {availableBackends.map((backend) => (
              <Card
                key={backend.id}
                className={`cursor-pointer transition-colors ${
                  attachingId === backend.id
                    ? "bg-muted"
                    : "hover:bg-accent"
                }`}
              >
                <CardHeader className="pb-3">
                  <div className="flex items-start justify-between">
                    <div className="flex-1 min-w-0">
                      <CardTitle className="text-lg truncate">{backend.name}</CardTitle>
                      <CardDescription className="mt-1 text-xs">
                        {backend.id}
                      </CardDescription>
                    </div>
                    <div className="flex items-center gap-2 ml-2">
                      <BackendHealthIndicator status={backend.health_status} />
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="pt-0 space-y-3">
                  <div className="space-y-2">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">Type:</span>
                      <Badge variant="outline" className="text-xs">
                        {backend.backend_type}
                      </Badge>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">URL:</span>
                      <span className="text-xs truncate max-w-[200px]">{backend.url}</span>
                    </div>
                  </div>

                  <Button
                    onClick={() => handleAttachBackend(backend.id)}
                    disabled={attachingId !== null}
                    className="w-full"
                    size="sm"
                  >
                    {attachingId === backend.id ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Attaching...
                      </>
                    ) : (
                      <>
                        <CheckCircle2 className="mr-2 h-4 w-4" />
                        Add to Plan
                      </>
                    )}
                  </Button>
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        <div className="flex justify-end">
          <Button variant="outline" onClick={onClose}>
            Close
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

