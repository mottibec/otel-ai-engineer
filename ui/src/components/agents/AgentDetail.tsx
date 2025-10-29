import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { apiClient } from "@/services/api";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import type { Agent, ToolInfo } from "@/types/agent";

interface AgentDetailProps {
  agentId: string;
}

export function AgentDetail({ agentId }: AgentDetailProps) {
  const [agent, setAgent] = useState<Agent | null>(null);
  const [tools, setTools] = useState<ToolInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    const loadAgent = async () => {
      try {
        setLoading(true);
        const [agentData, toolsData] = await Promise.all([
          apiClient.getAgent(agentId),
          apiClient.getAgentTools(agentId).catch(() => []),
        ]);
        setAgent(agentData);
        setTools(toolsData);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err : new Error("Failed to load agent"));
      } finally {
        setLoading(false);
      }
    };

    loadAgent();
  }, [agentId]);

  const handleDelete = async () => {
    if (!agent || agent.type !== "custom") return;
    if (!confirm(`Are you sure you want to delete agent "${agent.name}"?`)) return;

    try {
      await apiClient.deleteCustomAgent(agentId);
      navigate("/agents");
    } catch (err) {
      alert("Failed to delete agent");
      console.error(err);
    }
  };

  if (loading) {
    return (
      <div className="p-6 space-y-4">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-96 w-full" />
      </div>
    );
  }

  if (error || !agent) {
    return (
      <div className="p-6">
        <Card>
          <CardHeader>
            <CardTitle>Error</CardTitle>
            <CardDescription>{error?.message || "Agent not found"}</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">{agent.name}</h1>
          {agent.description && (
            <p className="text-muted-foreground mt-1">{agent.description}</p>
          )}
        </div>
        <div className="flex items-center gap-2">
          <span
            className={`text-xs px-2 py-1 rounded ${
              agent.type === "built-in"
                ? "bg-blue-100 text-blue-800"
                : "bg-green-100 text-green-800"
            }`}
          >
            {agent.type}
          </span>
          {agent.type === "custom" && (
            <Button variant="destructive" onClick={handleDelete}>
              Delete
            </Button>
          )}
          <Button variant="outline" onClick={() => navigate("/agents")}>
            Back
          </Button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Configuration</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {agent.model && (
              <div>
                <span className="font-medium">Model:</span> {agent.model}
              </div>
            )}
            {agent.max_tokens && (
              <div>
                <span className="font-medium">Max Tokens:</span> {agent.max_tokens}
              </div>
            )}
            {agent.created_at && (
              <div>
                <span className="font-medium">Created:</span>{" "}
                {new Date(agent.created_at).toLocaleString()}
              </div>
            )}
            {agent.updated_at && (
              <div>
                <span className="font-medium">Updated:</span>{" "}
                {new Date(agent.updated_at).toLocaleString()}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>System Prompt</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground whitespace-pre-wrap">
              {agent.system_prompt || "No system prompt configured"}
            </p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Tools ({tools.length || agent.tool_names?.length || 0})</CardTitle>
          <CardDescription>
            Tools available to this agent
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {(tools.length > 0 ? tools : agent.tool_names || []).map((tool, idx) => (
              <div key={idx} className="border rounded p-3">
                <div className="font-medium">
                  {typeof tool === "string" ? tool : tool.name}
                </div>
                {typeof tool === "object" && tool.description && (
                  <div className="text-sm text-muted-foreground mt-1">
                    {tool.description}
                  </div>
                )}
              </div>
            ))}
            {tools.length === 0 && (!agent.tool_names || agent.tool_names.length === 0) && (
              <p className="text-muted-foreground">No tools configured</p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

