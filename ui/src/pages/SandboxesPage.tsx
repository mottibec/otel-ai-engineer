import { useParams, useNavigate } from "react-router-dom";
import { SandboxList } from "@/components/sandboxes/SandboxList";
import { CreateSandboxModal } from "@/components/sandboxes/CreateSandboxModal";
import { useState } from "react";

export function SandboxesPage() {
  const { sandboxId } = useParams<{ sandboxId?: string }>();
  const navigate = useNavigate();
  const [showCreateSandbox, setShowCreateSandbox] = useState(false);

  const handleSandboxSelected = (selectedSandboxId: string) => {
    navigate(`/sandboxes/${selectedSandboxId}`);
  };

  const handleCreateSandbox = () => {
    setShowCreateSandbox(true);
  };

  const handleSandboxCreated = (createdSandboxId: string) => {
    navigate(`/sandboxes/${createdSandboxId}`);
    setShowCreateSandbox(false);
  };

  // TODO: Add SandboxDetail component if needed
  if (sandboxId) {
    // For now, just show the list
    return <SandboxList onSelectSandbox={handleSandboxSelected} onCreateSandbox={handleCreateSandbox} />;
  }

  return (
    <>
      <SandboxList onSelectSandbox={handleSandboxSelected} onCreateSandbox={handleCreateSandbox} />
      <CreateSandboxModal
        isOpen={showCreateSandbox}
        onClose={() => setShowCreateSandbox(false)}
        onSandboxCreated={handleSandboxCreated}
      />
    </>
  );
}

