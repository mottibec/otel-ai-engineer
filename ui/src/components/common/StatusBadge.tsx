import type { RunStatus } from "../../types/models";
import { Badge } from "@/components/ui/badge";

interface StatusBadgeProps {
  status: RunStatus;
}

export function StatusBadge({ status }: StatusBadgeProps) {
  const getStatusVariant = () => {
    switch (status) {
      case "running":
        return "default";
      case "success":
        return "secondary";
      case "failed":
        return "destructive";
      case "cancelled":
        return "outline";
      default:
        return "outline";
    }
  };

  const getStatusIcon = () => {
    switch (status) {
      case "running":
        return "●";
      case "success":
        return "✓";
      case "failed":
        return "✗";
      case "cancelled":
        return "○";
      default:
        return "?";
    }
  };

  return (
    <Badge variant={getStatusVariant()} className="text-[10px] h-4 px-1.5">
      {getStatusIcon()}
    </Badge>
  );
}
