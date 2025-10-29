import { useParams, useNavigate } from "react-router-dom";
import { CollectorDetail } from "@/components/collectors/CollectorDetail";
import { CollectorList } from "@/components/collectors/CollectorList";
import { DeployCollectorModal } from "@/components/collectors/DeployCollectorModal";
import { useState } from "react";

export function CollectorsPage() {
  const { collectorId } = useParams<{ collectorId?: string }>();
  const navigate = useNavigate();
  const [showCreateCollector, setShowCreateCollector] = useState(false);

  const handleCollectorSelected = (selectedCollectorId: string) => {
    navigate(`/collectors/${selectedCollectorId}`);
  };

  const handleCollectorDeployed = (deployedCollectorId: string) => {
    navigate(`/collectors/${deployedCollectorId}`);
    setShowCreateCollector(false);
  };

  if (collectorId) {
    return (
      <>
        <CollectorDetail key={collectorId} collectorId={collectorId} />
        <DeployCollectorModal
          isOpen={showCreateCollector}
          onClose={() => setShowCreateCollector(false)}
          onCollectorDeployed={handleCollectorDeployed}
        />
      </>
    );
  }

  return (
    <>
      <CollectorList
        onSelectCollector={handleCollectorSelected}
        onCreateCollector={() => setShowCreateCollector(true)}
      />
      <DeployCollectorModal
        isOpen={showCreateCollector}
        onClose={() => setShowCreateCollector(false)}
        onCollectorDeployed={handleCollectorDeployed}
      />
    </>
  );
}

