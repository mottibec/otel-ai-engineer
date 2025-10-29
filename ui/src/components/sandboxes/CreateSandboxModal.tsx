import { useState } from "react";
import { useSWRConfig } from "swr";
import { apiClient } from "@/services/api";
import type { CreateSandboxRequest } from "@/types/sandbox";
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

interface CreateSandboxModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSandboxCreated?: (sandboxId: string) => void;
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
  debug:
    verbosity: detailed

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [debug]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [debug]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [debug]`;

export function CreateSandboxModal({
  isOpen,
  onClose,
  onSandboxCreated,
}: CreateSandboxModalProps) {
  const { mutate } = useSWRConfig();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
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
      const request: CreateSandboxRequest = {
        name: name.trim(),
        description: description.trim() || undefined,
        collector_config: config.trim(),
        collector_version: "latest",
      };

      const sandbox = await apiClient.createSandbox(request);

      // Invalidate and revalidate the sandboxes cache
      mutate("/api/sandboxes");

      // Reset form
      setName("");
      setDescription("");
      setConfig(DEFAULT_CONFIG);

      // Notify parent and close modal
      onSandboxCreated?.(sandbox.id);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create sandbox");
    } finally {
      setSubmitting(false);
    }
  };

  const handleClose = () => {
    if (!submitting) {
      setError(null);
      onClose();
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create Collector Sandbox</DialogTitle>
          <DialogDescription>
            Create an isolated testing environment for your OpenTelemetry collector
            configuration
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">
              Sandbox Name <span className="text-red-500">*</span>
            </Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g., test-traces-pipeline"
              disabled={submitting}
              autoFocus
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Input
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="What are you testing?"
              disabled={submitting}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="config">
              Collector Configuration (YAML) <span className="text-red-500">*</span>
            </Label>
            <Textarea
              id="config"
              value={config}
              onChange={(e) => setConfig(e.target.value)}
              placeholder="Paste your collector config here..."
              className="font-mono text-sm min-h-[400px]"
              disabled={submitting}
            />
            <p className="text-xs text-muted-foreground">
              A complete OpenTelemetry collector configuration in YAML format
            </p>
          </div>

          {error && (
            <div className="rounded-md bg-red-50 dark:bg-red-950 p-3 text-sm text-red-800 dark:text-red-200">
              {error}
            </div>
          )}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={handleClose}
              disabled={submitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create Sandbox
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
