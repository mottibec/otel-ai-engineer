import { Routes, Route, Navigate, useLocation } from "react-router-dom";
import { Layout } from "./components/layout/Layout";
import { AgentSidebar } from "./components/layout/AgentSidebar";
import { RunsPage } from "./pages/RunsPage";
import { PlansPage } from "./pages/PlansPage";
import { SandboxesPage } from "./pages/SandboxesPage";
import { CollectorsPage } from "./pages/CollectorsPage";
import { BackendsPage } from "./pages/BackendsPage";
import { AgentWorkPage } from "./pages/AgentWorkPage";
import { AgentsPage } from "./pages/AgentsPage";
import { useState } from "react";

function App() {
  const location = useLocation();
  const [isCreateRunModalOpen, setIsCreateRunModalOpen] = useState(false);

  // Extract runId from URL if in runs route
  const runsMatch = location.pathname.match(/^\/runs\/([^/]+)$/);
  const selectedRunId = runsMatch ? runsMatch[1] : undefined;

  return (
    <Layout
      sidebar={
        <AgentSidebar
          selectedRunId={selectedRunId}
          onRunCreated={() => setIsCreateRunModalOpen(false)}
          isCreateRunModalOpen={isCreateRunModalOpen}
          onModalClose={() => setIsCreateRunModalOpen(false)}
        />
      }
      onRunCreated={() => setIsCreateRunModalOpen(false)}
    >
      <Routes>
        <Route path="/" element={<Navigate to="/runs" replace />} />
        <Route path="/runs" element={<RunsPage />} />
        <Route path="/runs/:runId" element={<RunsPage />} />
        <Route path="/plans" element={<PlansPage />} />
        <Route path="/plans/:planId" element={<PlansPage />} />
        <Route path="/sandboxes" element={<SandboxesPage />} />
        <Route path="/sandboxes/:sandboxId" element={<SandboxesPage />} />
        <Route path="/collectors" element={<CollectorsPage />} />
        <Route path="/collectors/:collectorId" element={<CollectorsPage />} />
        <Route path="/backends" element={<BackendsPage />} />
        <Route path="/backends/:backendId" element={<BackendsPage />} />
        <Route path="/agent-work" element={<AgentWorkPage />} />
        <Route path="/agents" element={<AgentsPage />} />
        <Route path="/agents/:agentId" element={<AgentsPage />} />
      </Routes>
    </Layout>
  );
}

export default App;
