import { useState } from "react";
import { apiClient } from "@/services/api";
import type { InfrastructureComponent } from "@/types/plan";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Loader2 } from "lucide-react";

interface AddInfrastructureModalProps {
  isOpen: boolean;
  onClose: () => void;
  planId: string;
  onInfrastructureAdded: () => void;
}

export function AddInfrastructureModal({
  isOpen,
  onClose,
  planId,
  onInfrastructureAdded,
}: AddInfrastructureModalProps) {
  const [componentName, setComponentName] = useState("");
  const [componentType, setComponentType] = useState("");
  const [host, setHost] = useState("");
  const [receiverType, setReceiverType] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!componentName.trim()) {
      setError("Component name is required");
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const infrastructure: Partial<InfrastructureComponent> = {
        plan_id: planId,
        name: componentName.trim(),
        component_type: componentType || undefined,
        host: host || undefined,
        receiver_type: receiverType || undefined,
        metrics_collected: "",
        status: "pending",
      };

      await apiClient.createInfrastructure(planId, infrastructure as InfrastructureComponent);
      onInfrastructureAdded();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create infrastructure component");
    } finally {
      setSubmitting(false);
    }
  };

  const handleClose = () => {
    if (!submitting) {
      setComponentName("");
      setComponentType("");
      setHost("");
      setReceiverType("");
      setError(null);
      onClose();
    }
  };

  // Map component types to receiver types
  const getReceiverOptions = (type: string) => {
    switch (type) {
      case "database":
        return ["postgres", "mysql", "mongodb", "redis", "cassandra"];
      case "cache":
        return ["redis", "memcached"];
      case "queue":
        return ["kafka", "rabbitmq", "nats"];
      case "host":
        return ["hostmetrics"];
      case "kubernetes":
        return ["k8s_cluster", "k8s_events", "k8s_state"];
      default:
        return ["hostmetrics"];
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Add Infrastructure Component</DialogTitle>
          <DialogDescription>
            Add infrastructure components that need to be monitored via a dedicated collector.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Component Name */}
          <div className="space-y-2">
            <Label htmlFor="componentName">Component Name *</Label>
            <Input
              id="componentName"
              value={componentName}
              onChange={(e) => setComponentName(e.target.value)}
              placeholder="e.g., postgres-db, redis-cache, k8s-cluster"
              disabled={submitting}
              required
            />
          </div>

          {/* Component Type */}
          <div className="space-y-2">
            <Label htmlFor="componentType">Component Type</Label>
            <Select
              value={componentType}
              onValueChange={(value) => {
                setComponentType(value);
                setReceiverType(""); // Reset receiver when type changes
              }}
              disabled={submitting}
            >
              <SelectTrigger id="componentType">
                <SelectValue placeholder="Select component type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="database">Database</SelectItem>
                <SelectItem value="cache">Cache</SelectItem>
                <SelectItem value="queue">Message Queue</SelectItem>
                <SelectItem value="host">Host</SelectItem>
                <SelectItem value="kubernetes">Kubernetes</SelectItem>
                <SelectItem value="container">Container</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Host */}
          <div className="space-y-2">
            <Label htmlFor="host">Host</Label>
            <Input
              id="host"
              value={host}
              onChange={(e) => setHost(e.target.value)}
              placeholder="e.g., localhost:5432, redis.example.com:6379"
              disabled={submitting}
            />
          </div>

          {/* Receiver Type */}
          <div className="space-y-2">
            <Label htmlFor="receiverType">Receiver Type</Label>
            <Select
              value={receiverType}
              onValueChange={setReceiverType}
              disabled={submitting || !componentType}
            >
              <SelectTrigger id="receiverType">
                <SelectValue placeholder="Select receiver type" />
              </SelectTrigger>
              <SelectContent>
                {getReceiverOptions(componentType).map((receiver) => (
                  <SelectItem key={receiver} value={receiver}>
                    {receiver}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              The OpenTelemetry receiver type that will collect metrics from this infrastructure component.
            </p>
          </div>

          {/* Error Display */}
          {error && (
            <div className="p-3 text-sm text-red-500 bg-red-500/10 border border-red-500/20 rounded-md">
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
            <Button type="submit" disabled={submitting || !componentName.trim()}>
              {submitting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Adding...
                </>
              ) : (
                "Add Infrastructure"
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

