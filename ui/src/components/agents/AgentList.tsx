import { useState, useEffect } from "react";
import { apiClient } from "@/services/api";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { CreateAgentModal } from "./CreateAgentModal";
import { CreateMetaAgentModal } from "./CreateMetaAgentModal";
import type { Agent } from "@/types/agent";

interface AgentListProps {
  onSelectAgent: (agentId: string) => void;
}

export function AgentList({ onSelectAgent }: AgentListProps) {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isMetaModalOpen, setIsMetaModalOpen] = useState(false);

  const loadAgents = async () => {
    try {
      setLoading(true);
      const data = await apiClient.listAgents();
      setAgents(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Failed to load agents"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadAgents();
  }, []);

  if (loading) {
    return (
      <div className="space-y-4 p-6">
        <div className="flex items-center justify-between">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-10 w-32" />
        </div>
        <div className="grid gap-4">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-32 w-full" />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <Card>
          <CardHeader>
            <CardTitle>Error</CardTitle>
            <CardDescription>{error.message}</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  const handleAgentCreated = () => {
    loadAgents();
    setIsCreateModalOpen(false);
  };

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Agents</h1>
        <div className="flex gap-2">
          <button
            onClick={loadAgents}
            className="px-4 py-2 text-sm border rounded-md hover:bg-accent"
          >
            Refresh
          </button>
          <button
            onClick={() => setIsMetaModalOpen(true)}
            className="px-4 py-2 text-sm border rounded-md hover:bg-accent"
          >
            Create Meta-Agent
          </button>
          <button
            onClick={() => setIsCreateModalOpen(true)}
            className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
          >
            Create Agent
          </button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {agents.map((agent) => (
          <Card
            key={agent.id}
            className="cursor-pointer hover:shadow-lg transition-shadow"
            onClick={() => onSelectAgent(agent.id)}
          >
            <CardHeader>
              <div className="flex items-start justify-between">
                <CardTitle className="text-lg">{agent.name}</CardTitle>
                <span
                  className={`text-xs px-2 py-1 rounded ${
                    agent.type === "built-in"
                      ? "bg-blue-100 text-blue-800"
                      : "bg-green-100 text-green-800"
                  }`}
                >
                  {agent.type}
                </span>
              </div>
              <CardDescription>{agent.description}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <div className="text-sm text-muted-foreground">
                  <span className="font-medium">Tools:</span>{" "}
                  {agent.tool_names?.length || agent.tools?.length || 0}
                </div>
                {agent.model && (
                  <div className="text-sm text-muted-foreground">
                    <span className="font-medium">Model:</span> {agent.model}
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <CreateAgentModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onAgentCreated={handleAgentCreated}
      />

      <CreateMetaAgentModal
        isOpen={isMetaModalOpen}
        onClose={() => setIsMetaModalOpen(false)}
        onAgentCreated={handleAgentCreated}
      />
    </div>
  );
}

