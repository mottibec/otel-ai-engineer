import { Badge } from "@/components/ui/badge";

interface BackendHealthIndicatorProps {
  status: string;
}

export function BackendHealthIndicator({ status }: BackendHealthIndicatorProps) {
  const getStatusColor = () => {
    switch (status) {
      case "healthy":
        return "bg-green-500/10 text-green-500 border-green-500/20";
      case "unhealthy":
        return "bg-red-500/10 text-red-500 border-red-500/20";
      case "unknown":
      default:
        return "bg-gray-500/10 text-gray-500 border-gray-500/20";
    }
  };

  const getStatusIcon = () => {
    switch (status) {
      case "healthy":
        return "●";
      case "unhealthy":
        return "●";
      case "unknown":
      default:
        return "○";
    }
  };

  return (
    <Badge variant="outline" className={`text-xs h-5 px-2 ${getStatusColor()}`}>
      {getStatusIcon()} {status}
    </Badge>
  );
}

