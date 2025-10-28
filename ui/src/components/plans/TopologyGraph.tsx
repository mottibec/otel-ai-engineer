import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { TopologyNode, TopologyEdge } from "@/types/plan";

interface TopologyGraphProps {
  nodes: TopologyNode[];
  edges: TopologyEdge[];
}

export function TopologyGraph({ nodes, edges }: TopologyGraphProps) {
  if (nodes.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Topology</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-center py-8">
            No components in this plan yet
          </p>
        </CardContent>
      </Card>
    );
  }

  const getNodeColor = (status: string) => {
    switch (status) {
      case "success":
      case "healthy":
        return "bg-green-500";
      case "pending":
      case "unknown":
        return "bg-yellow-500";
      case "failed":
      case "unhealthy":
        return "bg-red-500";
      default:
        return "bg-gray-500";
    }
  };

  const getNodeIcon = (type: string) => {
    switch (type) {
      case "service":
        return "🔵";
      case "infrastructure":
        return "⚙️";
      case "pipeline":
        return "📊";
      case "backend":
        return "🔗";
      default:
        return "🔷";
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Topology</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="relative overflow-auto" style={{ minHeight: "400px" }}>
          {/* Simplified SVG visualization */}
          <svg
            width="100%"
            height="400"
            viewBox={`0 0 800 600`}
            className="border rounded-lg bg-muted/20"
          >
            {/* Render edges first (so they appear behind nodes) */}
            <g stroke="#9ca3af" strokeWidth="2" strokeDasharray="5,5">
              {edges.map((edge, idx) => {
                // Simple edge positioning (this would be enhanced with a proper layout algorithm)
                const sourceIdx = nodes.findIndex((n) => n.id === edge.source_id);
                const targetIdx = nodes.findIndex((n) => n.id === edge.target_id);
                
                if (sourceIdx === -1 || targetIdx === -1) return null;
                
                const x1 = 150 + (sourceIdx % 3) * 250;
                const y1 = 100 + Math.floor(sourceIdx / 3) * 150;
                const x2 = 150 + (targetIdx % 3) * 250;
                const y2 = 100 + Math.floor(targetIdx / 3) * 150;
                
                return <line key={idx} x1={x1} y1={y1} x2={x2} y2={y2} />;
              })}
            </g>

            {/* Render nodes */}
            {nodes.map((node, idx) => {
              const x = 150 + (idx % 3) * 250;
              const y = 100 + Math.floor(idx / 3) * 150;
              
              return (
                <g key={node.id}>
                  {/* Node circle */}
                  <circle
                    cx={x}
                    cy={y}
                    r="35"
                    fill="currentColor"
                    className={getNodeColor(node.status)}
                  />
                  {/* Node label background */}
                  <rect
                    x={x - 50}
                    y={y + 40}
                    width="100"
                    height="30"
                    rx="4"
                    fill="white"
                    stroke="currentColor"
                    className="text-border"
                    strokeWidth="1"
                  />
                  {/* Node type icon */}
                  <text
                    x={x}
                    y={y + 5}
                    fontSize="20"
                    textAnchor="middle"
                    dominantBaseline="middle"
                  >
                    {getNodeIcon(node.type)}
                  </text>
                  {/* Node label */}
                  <text
                    x={x}
                    y={y + 55}
                    fontSize="10"
                    textAnchor="middle"
                    dominantBaseline="middle"
                    fill="currentColor"
                    className="font-medium"
                  >
                    {node.label.length > 12 ? node.label.substring(0, 12) + "..." : node.label}
                  </text>
                  {/* Status indicator */}
                  <circle
                    cx={x + 30}
                    cy={y - 30}
                    r="8"
                    className={getNodeColor(node.status)}
                  />
                </g>
              );
            })}
          </svg>

          {/* Legend */}
          <div className="mt-4 grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="flex items-center gap-2 text-sm">
              <div className="w-4 h-4 bg-blue-500 rounded-full" />
              <span>Service</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <div className="w-4 h-4 bg-gray-500 rounded-full" />
              <span>Infrastructure</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <div className="w-4 h-4 bg-purple-500 rounded-full" />
              <span>Pipeline</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <div className="w-4 h-4 bg-cyan-500 rounded-full" />
              <span>Backend</span>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

