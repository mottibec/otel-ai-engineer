import { useParams, useNavigate } from "react-router-dom";
import { AgentList } from "@/components/agents/AgentList";
import { AgentDetail } from "@/components/agents/AgentDetail";

export function AgentsPage() {
  const { agentId } = useParams<{ agentId?: string }>();
  const navigate = useNavigate();

  const handleSelectAgent = (selectedAgentId: string) => {
    navigate(`/agents/${selectedAgentId}`);
  };

  if (agentId) {
    return <AgentDetail key={agentId} agentId={agentId} />;
  }

  return <AgentList onSelectAgent={handleSelectAgent} />;
}

