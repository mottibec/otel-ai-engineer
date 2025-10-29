import { useParams, useNavigate } from "react-router-dom";
import { BackendDetail } from "@/components/backends/BackendDetail";
import { BackendList } from "@/components/backends/BackendList";
import { CreateBackendModal } from "@/components/backends/CreateBackendModal";
import { useState } from "react";

export function BackendsPage() {
  const { backendId } = useParams<{ backendId?: string }>();
  const navigate = useNavigate();
  const [showCreateBackend, setShowCreateBackend] = useState(false);

  const handleBackendSelected = (selectedBackendId: string) => {
    navigate(`/backends/${selectedBackendId}`);
  };

  const handleBackendCreated = (createdBackendId: string) => {
    navigate(`/backends/${createdBackendId}`);
    setShowCreateBackend(false);
  };

  if (backendId) {
    return (
      <>
        <BackendDetail key={backendId} backendId={backendId} />
        <CreateBackendModal
          isOpen={showCreateBackend}
          onClose={() => setShowCreateBackend(false)}
          onBackendCreated={handleBackendCreated}
        />
      </>
    );
  }

  return (
    <>
      <BackendList
        onSelectBackend={handleBackendSelected}
        onCreateBackend={() => setShowCreateBackend(true)}
      />
      <CreateBackendModal
        isOpen={showCreateBackend}
        onClose={() => setShowCreateBackend(false)}
        onBackendCreated={handleBackendCreated}
      />
    </>
  );
}

