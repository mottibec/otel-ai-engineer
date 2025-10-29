import { useState } from "react";
import { useSWRConfig } from "swr";
import { useAgents } from "../../hooks/useAgents";
import { apiClient } from "../../services/api";
import type { DelegateRequest, ResourceType } from "../../types/agent-work";
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
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { AlertTriangle, Loader2 } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";

interface DelegateToAgentProps {
  isOpen: boolean;
  onClose: () => void;
  resourceType: ResourceType;
  resourceId: string;
  resourceName?: string;
  onDelegated?: (runId: string, workId: string) => void;
}

export function DelegateToAgent({
  isOpen,
  onClose,
  resourceType,
  resourceId,
  resourceName,
  onDelegated,
}: DelegateToAgentProps) {
  const { mutate } = useSWRConfig();
  const { agents, loading: agentsLoading, error: agentsError } = useAgents();
  const [selectedAgentId, setSelectedAgentId] = useState<string>("");
  const [taskDescription, setTaskDescription] = useState<string>("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!selectedAgentId || !taskDescription.trim()) {
      setError("Please select an agent and enter a task description");
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const request: DelegateRequest = {
        resource_type: resourceType,
        resource_id: resourceId,
        agent_id: selectedAgentId,
        task_description: taskDescription.trim(),
      };

      const response = await apiClient.delegateToAgent(resourceType, resourceId, request);

      // Invalidate agent work cache for this resource
      mutate(`/api/agent-work/resource/${resourceType}/${resourceId}`);
      // Also invalidate the resource cache (e.g., backends, collectors)
      mutate(`/api/${resourceType}s/${resourceId}`);
      mutate(`/api/${resourceType}s`);

      // Reset form
      setSelectedAgentId("");
      setTaskDescription("");

      // Notify parent and close modal
      onDelegated?.(response.run_id, response.work_id);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delegate task to agent");
    } finally {
      setSubmitting(false);
    }
  };

  const handleClose = () => {
    if (!submitting) {
      setSelectedAgentId("");
      setTaskDescription("");
      setError(null);
      onClose();
    }
  };

  const selectedAgent = agents.find((a) => a.id === selectedAgentId);
  const resourceDisplayName = resourceName || resourceId;

  // Get default task description suggestions based on resource type
  const getDefaultTaskDescription = () => {
    switch (resourceType) {
      case "collector":
        return `Configure and optimize the OpenTelemetry collector "${resourceDisplayName}". Ensure proper data flow and pipeline configuration.`;
      case "backend":
        return `Configure the observability backend "${resourceDisplayName}". Set up datasources, dashboards, and alerts as needed.`;
      case "service":
        return `Instrument the service "${resourceDisplayName}" with OpenTelemetry. Add appropriate telemetry collection.`;
      case "infrastructure":
        return `Set up monitoring for the infrastructure component "${resourceDisplayName}". Configure metrics and alerts.`;
      case "pipeline":
        return `Configure the collector pipeline "${resourceDisplayName}". Set up processors and exporters.`;
      default:
        return `Complete the task for ${resourceType} "${resourceDisplayName}".`;
    }
  };

  const handleUseDefault = () => {
    setTaskDescription(getDefaultTaskDescription());
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-2xl max-w-[calc(100vw-2rem)]">
        <DialogHeader>
          <DialogTitle>Delegate to Agent</DialogTitle>
          <DialogDescription>
            Assign an agent to work on {resourceType} "{resourceDisplayName}"
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 overflow-hidden">
          {/* Resource Info */}
          <div className="p-3 bg-muted rounded-md text-sm">
            <div className="font-medium">Resource</div>
            <div className="text-muted-foreground">
              {resourceType} / {resourceId}
            </div>
          </div>

          {/* Agent Selection */}
          <div className="space-y-2 min-w-0">
            <Label htmlFor="agent">Select Agent</Label>
            {agentsLoading ? (
              <div className="flex items-center gap-2 text-muted-foreground text-sm">
                <Loader2 className="h-4 w-4 animate-spin" />
                Loading agents...
              </div>
            ) : agentsError ? (
              <Alert variant="destructive">
                <AlertTriangle className="h-4 w-4" />
                <AlertDescription>
                  Error loading agents: {agentsError.message}
                </AlertDescription>
              </Alert>
            ) : agents.length === 0 ? (
              <Alert>
                <AlertTriangle className="h-4 w-4" />
                <AlertDescription>No agents available</AlertDescription>
              </Alert>
            ) : (
              <Select
                value={selectedAgentId}
                onValueChange={setSelectedAgentId}
                disabled={submitting}
              >
                <SelectTrigger id="agent" className="w-full overflow-hidden">
                  <SelectValue placeholder="-- Select an agent --" />
                </SelectTrigger>
                <SelectContent className="w-[var(--radix-select-trigger-width)] max-w-[var(--radix-select-trigger-width)]">
                  {agents.map((agent) => (
                    <SelectItem key={agent.id} value={agent.id}>
                      <span
                        className="truncate block"
                        title={`${agent.name} - ${agent.description}`}
                      >
                        {agent.name}
                      </span>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
            {selectedAgent && (
              <div className="space-y-1">
                <div className="text-sm text-muted-foreground truncate">
                  Model: {selectedAgent.model}
                </div>
                <div
                  className="text-xs text-muted-foreground truncate"
                  title={selectedAgent.description}
                >
                  {selectedAgent.description}
                </div>
              </div>
            )}
          </div>

          {/* Task Description Input */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="task">Task Description</Label>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={handleUseDefault}
                disabled={submitting}
                className="h-7 text-xs"
              >
                Use Default
              </Button>
            </div>
            <Textarea
              id="task"
              value={taskDescription}
              onChange={(e) => setTaskDescription(e.target.value)}
              disabled={submitting}
              placeholder="Describe the task you want the agent to perform..."
              rows={6}
              className="resize-none"
            />
            <div className="text-xs text-muted-foreground">
              The agent will receive context about this resource and the task you specify.
            </div>
          </div>

          {/* Error Display */}
          {error && (
            <Alert variant="destructive">
              <AlertTriangle className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
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
            <Button
              type="submit"
              disabled={submitting || !selectedAgentId || !taskDescription.trim()}
            >
              {submitting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Delegating...
                </>
              ) : (
                "Delegate to Agent"
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

