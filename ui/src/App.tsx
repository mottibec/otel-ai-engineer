import { useState } from "react";
import { Layout } from "./components/layout/Layout";
import { AgentSidebar } from "./components/layout/AgentSidebar";
import { RunDetail } from "./components/detail/RunDetail";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

function App() {
  const [selectedRunId, setSelectedRunId] = useState<string | undefined>();

  const handleRunCreated = (runId: string) => {
    // Automatically select the newly created run
    setSelectedRunId(runId);
  };

  return (
    <Layout
      sidebar={
        <AgentSidebar
          selectedRunId={selectedRunId}
          onSelectRun={setSelectedRunId}
          onRunCreated={handleRunCreated}
        />
      }
      onRunCreated={handleRunCreated}
    >
      {selectedRunId ? (
        <RunDetail key={selectedRunId} runId={selectedRunId} />
      ) : (
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
      )}
    </Layout>
  );
}

export default App;
