import { useState } from "react";
import { useSWRConfig } from "swr";
import { useAgents } from "../../hooks/useAgents";
import { apiClient } from "../../services/api";
import type { StartRunRequest } from "../../types/models";
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

interface StartRunModalProps {
  isOpen: boolean;
  onClose: () => void;
  onRunCreated?: (runId: string) => void;
}

export function StartRunModal({
  isOpen,
  onClose,
  onRunCreated,
}: StartRunModalProps) {
  const { mutate } = useSWRConfig();
  const { agents, loading: agentsLoading, error: agentsError } = useAgents();
  const [selectedAgentId, setSelectedAgentId] = useState<string>("");
  const [prompt, setPrompt] = useState<string>("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!selectedAgentId || !prompt.trim()) {
      setError("Please select an agent and enter a prompt");
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const request: StartRunRequest = {
        agent_id: selectedAgentId,
        prompt: prompt.trim(),
      };

      const run = await apiClient.createRun(request);

      // Invalidate and revalidate the runs cache
      mutate("/api/runs");

      // Reset form
      setSelectedAgentId("");
      setPrompt("");

      // Notify parent and close modal
      onRunCreated?.(run.id);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to start run");
    } finally {
      setSubmitting(false);
    }
  };

  const handleClose = () => {
    if (!submitting) {
      setSelectedAgentId("");
      setPrompt("");
      setError(null);
      onClose();
    }
  };

  const selectedAgent = agents.find((a) => a.id === selectedAgentId);

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-2xl max-w-[calc(100vw-2rem)]">
        <DialogHeader>
          <DialogTitle>Start New Agent Run</DialogTitle>
          <DialogDescription>
            Select an agent and provide a prompt to start a new run
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 overflow-hidden">
          {/* Agent Selection */}
          <div className="space-y-2 min-w-0">
            <Label htmlFor="agent">Select Agent</Label>
            {agentsLoading ? (
              <div className="text-muted-foreground text-sm">
                Loading agents...
              </div>
            ) : agentsError ? (
              <div className="text-destructive text-sm">
                Error loading agents: {agentsError.message}
              </div>
            ) : agents.length === 0 ? (
              <div className="text-yellow-600 text-sm">No agents available</div>
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
                <div className="text-xs text-muted-foreground truncate" title={selectedAgent.description}>
                  {selectedAgent.description}
                </div>
              </div>
            )}
          </div>

          {/* Prompt Input */}
          <div className="space-y-2">
            <Label htmlFor="prompt">Prompt</Label>
            <Textarea
              id="prompt"
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              disabled={submitting}
              placeholder="Enter your task description..."
              rows={6}
              className="resize-none"
            />
          </div>

          {/* Error Display */}
          {error && (
            <div className="p-3 bg-destructive/10 border border-destructive rounded-md text-destructive text-sm">
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
            <Button
              type="submit"
              disabled={submitting || !selectedAgentId || !prompt.trim()}
            >
              {submitting ? "Starting..." : "Start Run"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
