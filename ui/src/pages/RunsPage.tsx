import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { RunDetail } from "@/components/detail/RunDetail";
import { RunList } from "@/components/runs/RunList";
import { StartRunModal } from "@/components/runs/StartRunModal";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";

export function RunsPage() {
  const { runId } = useParams<{ runId?: string }>();
  const navigate = useNavigate();
  const [isCreateRunModalOpen, setIsCreateRunModalOpen] = useState(false);

  // Handle navigation to run from agent work badges
  useEffect(() => {
    const handleNavigateToRun = (event: CustomEvent<{ runId: string }>) => {
      navigate(`/runs/${event.detail.runId}`);
    };

    window.addEventListener("navigateToRun", handleNavigateToRun as EventListener);
    return () => {
      window.removeEventListener("navigateToRun", handleNavigateToRun as EventListener);
    };
  }, [navigate]);

  const handleRunCreated = (newRunId: string) => {
    navigate(`/runs/${newRunId}`, { replace: true });
    setIsCreateRunModalOpen(false);
  };

  const handleCreateRun = () => {
    setIsCreateRunModalOpen(true);
  };

  const handleSelectRun = (selectedRunId: string) => {
    // Navigation handled by Link in RunCard
  };

  if (runId) {
    return (
      <>
        <RunDetail key={runId} runId={runId} />
        <StartRunModal
          isOpen={isCreateRunModalOpen}
          onClose={() => setIsCreateRunModalOpen(false)}
          onRunCreated={handleRunCreated}
        />
      </>
    );
  }

  return (
    <>
      <div className="flex items-center justify-center h-full p-8">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <CardTitle className="text-2xl">Agent Runtime Monitor</CardTitle>
            <CardDescription>
              Select a run from the sidebar to view details
            </CardDescription>
          </CardHeader>
          <CardContent className="text-center text-sm text-muted-foreground">
            <Button onClick={handleCreateRun}>
              <Plus className="mr-2 h-4 w-4" />
              New Run
            </Button>
          </CardContent>
        </Card>
      </div>
      <StartRunModal
        isOpen={isCreateRunModalOpen}
        onClose={() => setIsCreateRunModalOpen(false)}
        onRunCreated={handleRunCreated}
      />
    </>
  );
}

