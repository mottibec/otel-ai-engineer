import { useParams, useNavigate } from "react-router-dom";
import { PlanDetail } from "@/components/plans/PlanDetail";
import { PlanList } from "@/components/plans/PlanList";

export function PlansPage() {
  const { planId } = useParams<{ planId?: string }>();
  const navigate = useNavigate();

  const handleSelectPlan = (selectedPlanId: string) => {
    navigate(`/plans/${selectedPlanId}`);
  };

  if (planId) {
    return <PlanDetail key={planId} planId={planId} />;
  }

  return <PlanList onSelectPlan={handleSelectPlan} />;
}

