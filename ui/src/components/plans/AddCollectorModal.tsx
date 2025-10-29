import { useState, useEffect } from "react";
import { apiClient } from "@/services/api";
import type { Collector } from "@/types/collector";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Loader2, Server, CheckCircle2, Settings } from "lucide-react";
import { StatusBadge } from "@/components/common/StatusBadge";

interface AddCollectorModalProps {
  isOpen: boolean;
  onClose: () => void;
  planId: string;
  onCollectorAdded: () => void;
}

export function AddCollectorModal({
  isOpen,
  onClose,
  planId,
  onCollectorAdded,
}: AddCollectorModalProps) {
  const [collectors, setCollectors] = useState<Collector[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedCollectorId, setSelectedCollectorId] = useState<string | null>(null);
  const [pipelineName, setPipelineName] = useState("");
  const [creating, setCreating] = useState(false);
  const [showConfig, setShowConfig] = useState(false);

  useEffect(() => {
    if (isOpen) {
      loadCollectors();
    } else {
      // Reset state when modal closes
      setSelectedCollectorId(null);
      setPipelineName("");
      setShowConfig(false);
      setError(null);
    }
  }, [isOpen]);

  const loadCollectors = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.listCollectors();
      setCollectors(data.collectors);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load collectors");
    } finally {
      setLoading(false);
    }
  };

  const handleSelectCollector = (collectorId: string, collectorName: string) => {
    setSelectedCollectorId(collectorId);
    // Auto-generate pipeline name
    if (!pipelineName) {
      setPipelineName(`${collectorName}-pipeline`);
    }
  };

  const handleCreatePipeline = async () => {
    if (!selectedCollectorId) {
      setError("Please select a collector");
      return;
    }

    try {
      setCreating(true);
      setError(null);
      await apiClient.createPipelineFromCollector(planId, selectedCollectorId, {
        name: pipelineName.trim() || undefined,
      });
      onCollectorAdded();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create pipeline");
    } finally {
      setCreating(false);
    }
  };

  const selectedCollector = collectors.find((c) => c.collector_id === selectedCollectorId);

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Add Collector to Plan</DialogTitle>
          <DialogDescription>
            Select a collector to create a pipeline entry for this plan
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
        ) : collectors.length === 0 ? (
          <Card>
            <CardContent className="py-12 text-center">
              <Server className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
              <CardTitle className="mb-2">No collectors available</CardTitle>
              <CardDescription>
                No collectors are deployed yet. Deploy a collector first to add it to this plan.
              </CardDescription>
            </CardContent>
          </Card>
        ) : (
          <>
            {!selectedCollectorId ? (
              <div className="grid gap-4 md:grid-cols-2">
                {collectors.map((collector) => (
                  <Card
                    key={collector.collector_id}
                    className={`cursor-pointer transition-colors hover:bg-accent ${
                      selectedCollectorId === collector.collector_id ? "bg-muted" : ""
                    }`}
                    onClick={() => handleSelectCollector(collector.collector_id, collector.collector_name)}
                  >
                    <CardHeader className="pb-3">
                      <div className="flex items-start justify-between">
                        <div className="flex-1 min-w-0">
                          <CardTitle className="text-lg truncate flex items-center gap-2">
                            <Server className="h-5 w-5" />
                            {collector.collector_name}
                          </CardTitle>
                          <CardDescription className="mt-1 text-xs">
                            {collector.collector_id}
                          </CardDescription>
                        </div>
                        <div className="flex items-center gap-2 ml-2">
                          <StatusBadge status={collector.status} />
                        </div>
                      </div>
                    </CardHeader>
                    <CardContent className="pt-0 space-y-2">
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-muted-foreground">Target:</span>
                        <Badge variant="outline" className="text-xs">
                          {collector.target_type}
                        </Badge>
                      </div>
                      {collector.deployed_at && (
                        <div className="flex items-center justify-between text-sm">
                          <span className="text-muted-foreground">Deployed:</span>
                          <span className="text-xs">
                            {new Date(collector.deployed_at).toLocaleString()}
                          </span>
                        </div>
                      )}
                      <div className="pt-2">
                        <Button
                          variant="outline"
                          size="sm"
                          className="w-full"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleSelectCollector(collector.collector_id, collector.collector_name);
                          }}
                        >
                          <CheckCircle2 className="mr-2 h-4 w-4" />
                          Select
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            ) : (
              <div className="space-y-4">
                <Card className="bg-muted">
                  <CardHeader>
                    <CardTitle className="text-lg">Selected Collector</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-muted-foreground">Name:</span>
                        <span className="text-sm font-medium">{selectedCollector?.collector_name}</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-muted-foreground">ID:</span>
                        <span className="text-sm font-mono text-xs">{selectedCollector?.collector_id}</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-muted-foreground">Target Type:</span>
                        <Badge variant="outline">{selectedCollector?.target_type}</Badge>
                      </div>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setSelectedCollectorId(null)}
                        className="mt-2"
                      >
                        Change Selection
                      </Button>
                    </div>
                  </CardContent>
                </Card>

                <div className="space-y-2">
                  <Label htmlFor="pipeline-name">Pipeline Name</Label>
                  <Input
                    id="pipeline-name"
                    value={pipelineName}
                    onChange={(e) => setPipelineName(e.target.value)}
                    placeholder="Enter pipeline name (optional)"
                    disabled={creating}
                  />
                  <p className="text-xs text-muted-foreground">
                    A default name will be generated if not provided
                  </p>
                </div>

                <div className="flex gap-2">
                  <Button
                    onClick={handleCreatePipeline}
                    disabled={creating}
                    className="flex-1"
                  >
                    {creating ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Creating...
                      </>
                    ) : (
                      <>
                        <CheckCircle2 className="mr-2 h-4 w-4" />
                        Add to Plan
                      </>
                    )}
                  </Button>
                  <Button variant="outline" onClick={onClose} disabled={creating}>
                    Cancel
                  </Button>
                </div>
              </div>
            )}
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}

