import { useState } from "react";
import { useSWRConfig } from "swr";
import { apiClient } from "@/services/api";
import type { DeployCollectorRequest } from "@/types/collector";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Loader2 } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface DeployCollectorModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCollectorDeployed?: (collectorId: string) => void;
}

const DEFAULT_CONFIG = `receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024

exporters:
  otlp:
    endpoint: http://lawrence:4317

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]`;

export function DeployCollectorModal({
  isOpen,
  onClose,
  onCollectorDeployed,
}: DeployCollectorModalProps) {
  const { mutate } = useSWRConfig();
  const [name, setName] = useState("");
  const [targetType, setTargetType] = useState("docker");
  const [config, setConfig] = useState(DEFAULT_CONFIG);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim() || !config.trim()) {
      setError("Please provide a name and configuration");
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const request: DeployCollectorRequest = {
        collector_name: name.trim(),
        target_type: targetType,
        yaml_config: config.trim(),
      };

      const result = await apiClient.deployCollector(request);
      
      // Invalidate collectors cache
      mutate("/api/collectors");

      // Reset form
      setName("");
      setConfig(DEFAULT_CONFIG);

      // Get collector ID from result (structure may vary)
      const collectorId = (result as { collector_id?: string })?.collector_id || name;

      // Notify parent and close modal
      onCollectorDeployed?.(collectorId);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to deploy collector");
    } finally {
      setSubmitting(false);
    }
  };

  const handleClose = () => {
    if (!submitting) {
      setError(null);
      setName("");
      setConfig(DEFAULT_CONFIG);
      onClose();
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Deploy Collector</DialogTitle>
          <DialogDescription>
            Deploy a new OpenTelemetry collector instance with your configuration.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Collector Name</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="my-collector"
              disabled={submitting}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="target-type">Target Type</Label>
            <Select value={targetType} onValueChange={setTargetType} disabled={submitting}>
              <SelectTrigger id="target-type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="docker">Docker</SelectItem>
                <SelectItem value="kubernetes">Kubernetes</SelectItem>
                <SelectItem value="remote">Remote</SelectItem>
                <SelectItem value="local">Local</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="config">YAML Configuration</Label>
            <Textarea
              id="config"
              value={config}
              onChange={(e) => setConfig(e.target.value)}
              placeholder="Enter collector YAML configuration..."
              className="font-mono text-sm"
              rows={20}
              disabled={submitting}
              required
            />
          </div>

          {error && (
            <div className="p-3 text-sm text-red-500 bg-red-500/10 border border-red-500/20 rounded-md">
              {error}
            </div>
          )}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleClose} disabled={submitting}>
              Cancel
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Deploy Collector
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

