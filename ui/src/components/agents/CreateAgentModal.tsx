import { useState, useEffect } from "react";
import { apiClient } from "@/services/api";
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
import type { ToolInfo } from "@/types/agent";

interface CreateAgentModalProps {
  isOpen: boolean;
  onClose: () => void;
  onAgentCreated: () => void;
}

export function CreateAgentModal({
  isOpen,
  onClose,
  onAgentCreated,
}: CreateAgentModalProps) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [systemPrompt, setSystemPrompt] = useState("");
  const [model, setModel] = useState("claude-sonnet-4-5-20250929");
  const [maxTokens, setMaxTokens] = useState(4096);
  const [tools, setTools] = useState<ToolInfo[]>([]);
  const [selectedTools, setSelectedTools] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      loadTools();
    }
  }, [isOpen]);

  const loadTools = async () => {
    try {
      setLoading(true);
      const response = await apiClient.listAllTools();
      setTools(response.tools);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load tools");
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim() || !description.trim()) {
      setError("Name and description are required");
      return;
    }

    if (selectedTools.length === 0) {
      setError("At least one tool must be selected");
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      await apiClient.createCustomAgent({
        name: name.trim(),
        description: description.trim(),
        system_prompt: systemPrompt || undefined,
        model: model || undefined,
        max_tokens: maxTokens || undefined,
        tool_names: selectedTools,
      });
      onAgentCreated();
      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create agent");
    } finally {
      setSubmitting(false);
    }
  };

  const handleClose = () => {
    setName("");
    setDescription("");
    setSystemPrompt("");
    setModel("claude-sonnet-4-5-20250929");
    setMaxTokens(4096);
    setSelectedTools([]);
    setError(null);
    onClose();
  };

  const toggleTool = (toolName: string) => {
    setSelectedTools((prev) =>
      prev.includes(toolName)
        ? prev.filter((t) => t !== toolName)
        : [...prev, toolName]
    );
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create Custom Agent</DialogTitle>
          <DialogDescription>
            Create a new agent by selecting tools and configuring its behavior
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="name">Name *</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div>
            <Label htmlFor="description">Description *</Label>
            <Textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              required
              rows={2}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="model">Model</Label>
              <Input
                id="model"
                value={model}
                onChange={(e) => setModel(e.target.value)}
                placeholder="claude-sonnet-4-5-20250929"
              />
            </div>
            <div>
              <Label htmlFor="maxTokens">Max Tokens</Label>
              <Input
                id="maxTokens"
                type="number"
                value={maxTokens}
                onChange={(e) => setMaxTokens(parseInt(e.target.value) || 4096)}
              />
            </div>
          </div>

          <div>
            <Label htmlFor="systemPrompt">System Prompt</Label>
            <Textarea
              id="systemPrompt"
              value={systemPrompt}
              onChange={(e) => setSystemPrompt(e.target.value)}
              rows={4}
              placeholder="Optional system prompt to guide agent behavior..."
            />
          </div>

          <div>
            <Label>Tools * (Select at least one)</Label>
            {loading ? (
              <div className="text-sm text-muted-foreground">Loading tools...</div>
            ) : (
              <div className="mt-2 border rounded p-4 max-h-64 overflow-y-auto space-y-2">
                {tools.map((tool) => (
                  <label
                    key={tool.name}
                    className="flex items-start gap-2 cursor-pointer hover:bg-accent p-2 rounded"
                  >
                    <input
                      type="checkbox"
                      checked={selectedTools.includes(tool.name)}
                      onChange={() => toggleTool(tool.name)}
                      className="mt-1"
                    />
                    <div className="flex-1">
                      <div className="font-medium">{tool.name}</div>
                      <div className="text-sm text-muted-foreground">
                        {tool.description}
                      </div>
                      <div className="text-xs text-muted-foreground mt-1">
                        Category: {tool.category}
                      </div>
                    </div>
                  </label>
                ))}
              </div>
            )}
          </div>

          {error && (
            <div className="text-sm text-destructive bg-destructive/10 p-2 rounded">
              {error}
            </div>
          )}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create Agent
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

