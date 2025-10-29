import { useNavigate } from "react-router-dom";
import { AgentWorkList } from "@/components/agent-work/AgentWorkList";

export function AgentWorkPage() {
  const navigate = useNavigate();

  const handleSelectRun = (runId: string) => {
    navigate(`/runs/${runId}`);
  };

  return (
    <div className="h-full w-full">
      <AgentWorkList onSelectRun={handleSelectRun} />
    </div>
  );
}

