import { useState } from "react";
import { Layout } from "./components/layout/Layout";
import { AgentSidebar } from "./components/layout/AgentSidebar";
import { RunDetail } from "./components/detail/RunDetail";
import { PlanList } from "./components/plans/PlanList";
import { PlanDetail } from "./components/plans/PlanDetail";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

type ViewMode = "runs" | "plans";

function App() {
  const [viewMode, setViewMode] = useState<ViewMode>("runs");
  const [selectedRunId, setSelectedRunId] = useState<string | undefined>();
  const [selectedPlanId, setSelectedPlanId] = useState<string | undefined>();

  const handleRunCreated = (runId: string) => {
    setSelectedRunId(runId);
    setViewMode("runs");
  };

  const handlePlanSelected = (planId: string) => {
    setSelectedPlanId(planId);
    setViewMode("plans");
  };

  const renderContent = () => {
    if (viewMode === "runs") {
      if (selectedRunId) {
        return <RunDetail key={selectedRunId} runId={selectedRunId} />;
      } else {
        return (
          <div className="flex items-center justify-center h-full p-8">
            <Card className="w-full max-w-md">
              <CardHeader className="text-center">
                <div className="text-6xl mb-4">🤖</div>
                <CardTitle className="text-2xl">Agent Runtime Monitor</CardTitle>
                <CardDescription>
                  Select a run from the sidebar to view details
                </CardDescription>
              </CardHeader>
              <CardContent className="text-center text-sm text-muted-foreground">
                Use the "New Run" button to start a new agent execution
              </CardContent>
            </Card>
          </div>
        );
      }
    } else {
      if (selectedPlanId) {
        return <PlanDetail key={selectedPlanId} planId={selectedPlanId} />;
      } else {
        return <PlanList onSelectPlan={handlePlanSelected} />;
      }
    }
  };

  return (
    <Layout
      sidebar={
        <AgentSidebar
          selectedRunId={selectedRunId}
          onSelectRun={setSelectedRunId}
          onRunCreated={handleRunCreated}
          viewMode={viewMode}
          onViewModeChange={setViewMode}
        />
      }
      onRunCreated={handleRunCreated}
    >
      {renderContent()}
    </Layout>
  );
}

export default App;
