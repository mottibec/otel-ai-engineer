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
            No components in this plan yet. Add services, infrastructure, collectors, and backends to see the observability stack.
          </p>
        </CardContent>
      </Card>
    );
  }

  // Group nodes by type for better layout
  const services = nodes.filter((n) => n.type === "service");
  const infrastructure = nodes.filter((n) => n.type === "infrastructure");
  const pipelines = nodes.filter((n) => n.type === "pipeline");
  const backends = nodes.filter((n) => n.type === "backend");

  const getNodeColor = (status: string) => {
    switch (status) {
      case "success":
      case "healthy":
        return "#10b981"; // green
      case "pending":
      case "unknown":
        return "#f59e0b"; // yellow
      case "failed":
      case "unhealthy":
        return "#ef4444"; // red
      default:
        return "#6b7280"; // gray
    }
  };

  const getNodeTypeColor = (type: string) => {
    switch (type) {
      case "service":
        return "#3b82f6"; // blue
      case "infrastructure":
        return "#8b5cf6"; // purple
      case "pipeline":
        return "#10b981"; // green
      case "backend":
        return "#06b6d4"; // cyan
      default:
        return "#6b7280"; // gray
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

  // Calculate positions using a hierarchical layout
  // Services/Infrastructure -> Pipelines -> Backends
  const getNodePosition = (node: TopologyNode, index: number, total: number, layer: number) => {
    const layerY = [150, 300, 450]; // Top: services/infra, Middle: pipelines, Bottom: backends
    const spacing = 800 / (total + 1);
    const x = spacing * (index + 1);
    const y = layerY[layer] || 300;
    return { x, y };
  };

  // Build node positions map
  const nodePositions = new Map<string, { x: number; y: number }>();
  
  // Layer 0: Services and Infrastructure (top)
  [...services, ...infrastructure].forEach((node, idx) => {
    const pos = getNodePosition(node, idx, services.length + infrastructure.length, 0);
    nodePositions.set(node.id, pos);
  });
  
  // Layer 1: Pipelines (middle)
  pipelines.forEach((node, idx) => {
    const pos = getNodePosition(node, idx, pipelines.length, 1);
    nodePositions.set(node.id, pos);
  });
  
  // Layer 2: Backends (bottom)
  backends.forEach((node, idx) => {
    const pos = getNodePosition(node, idx, backends.length, 2);
    nodePositions.set(node.id, pos);
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Observability Stack Topology</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="relative overflow-auto" style={{ minHeight: "500px" }}>
          <svg
            width="100%"
            height="500"
            viewBox="0 0 800 500"
            className="border rounded-lg bg-muted/10"
          >
            <defs>
              <marker
                id="arrowhead"
                markerWidth="10"
                markerHeight="10"
                refX="9"
                refY="3"
                orient="auto"
              >
                <polygon points="0 0, 10 3, 0 6" fill="#6b7280" />
              </marker>
            </defs>

            {/* Render edges first (so they appear behind nodes) */}
            <g stroke="#6b7280" strokeWidth="2" fill="none" markerEnd="url(#arrowhead)">
              {edges.map((edge, idx) => {
                const sourcePos = nodePositions.get(edge.source_id);
                const targetPos = nodePositions.get(edge.target_id);
                
                if (!sourcePos || !targetPos) return null;
                
                return (
                  <line
                    key={idx}
                    x1={sourcePos.x}
                    y1={sourcePos.y}
                    x2={targetPos.x}
                    y2={targetPos.y}
                    strokeDasharray={edge.type === "data_flow" ? "5,5" : "none"}
                  />
                );
              })}
              
              {/* Auto-connect: Services/Infrastructure -> Pipelines -> Backends if no explicit edges */}
              {edges.length === 0 && (
                <>
                  {/* Connect services/infra to pipelines */}
                  {[...services, ...infrastructure].map((sourceNode) => {
                    const sourcePos = nodePositions.get(sourceNode.id);
                    if (!sourcePos) return null;
                    
                    return pipelines.map((targetNode) => {
                      const targetPos = nodePositions.get(targetNode.id);
                      if (!targetPos) return null;
                      return (
                        <line
                          key={`${sourceNode.id}-${targetNode.id}`}
                          x1={sourcePos.x}
                          y1={sourcePos.y}
                          x2={targetPos.x}
                          y2={targetPos.y}
                          strokeDasharray="5,5"
                          opacity="0.3"
                        />
                      );
                    });
                  })}
                  
                  {/* Connect pipelines to backends */}
                  {pipelines.map((sourceNode) => {
                    const sourcePos = nodePositions.get(sourceNode.id);
                    if (!sourcePos) return null;
                    
                    return backends.map((targetNode) => {
                      const targetPos = nodePositions.get(targetNode.id);
                      if (!targetPos) return null;
                      return (
                        <line
                          key={`${sourceNode.id}-${targetNode.id}`}
                          x1={sourcePos.x}
                          y1={sourcePos.y}
                          x2={targetPos.x}
                          y2={targetPos.y}
                          strokeDasharray="5,5"
                          opacity="0.3"
                        />
                      );
                    });
                  })}
                </>
              )}
            </g>

            {/* Render nodes */}
            {nodes.map((node) => {
              const pos = nodePositions.get(node.id);
              if (!pos) return null;
              const { x, y } = pos;
              
              return (
                <g key={node.id}>
                  {/* Node circle with type color */}
                  <circle
                    cx={x}
                    cy={y}
                    r="30"
                    fill={getNodeTypeColor(node.type)}
                    opacity="0.9"
                    stroke="white"
                    strokeWidth="2"
                  />
                  {/* Status indicator */}
                  <circle
                    cx={x + 22}
                    cy={y - 22}
                    r="6"
                    fill={getNodeColor(node.status)}
                    stroke="white"
                    strokeWidth="1"
                  />
                  {/* Node type icon */}
                  <text
                    x={x}
                    y={y + 8}
                    fontSize="24"
                    textAnchor="middle"
                    dominantBaseline="middle"
                  >
                    {getNodeIcon(node.type)}
                  </text>
                  {/* Node label background */}
                  <rect
                    x={x - 55}
                    y={y + 35}
                    width="110"
                    height="20"
                    rx="4"
                    fill="white"
                    stroke="#e5e7eb"
                    strokeWidth="1"
                    opacity="0.95"
                  />
                  {/* Node label */}
                  <text
                    x={x}
                    y={y + 48}
                    fontSize="11"
                    textAnchor="middle"
                    dominantBaseline="middle"
                    fill="#1f2937"
                    fontWeight="500"
                  >
                    {node.label.length > 15 ? node.label.substring(0, 15) + "..." : node.label}
                  </text>
                  {/* Node type badge */}
                  <text
                    x={x}
                    y={y - 35}
                    fontSize="9"
                    textAnchor="middle"
                    dominantBaseline="middle"
                    fill="#6b7280"
                    fontWeight="600"
                    textTransform="uppercase"
                  >
                    {node.type}
                  </text>
                </g>
              );
            })}
          </svg>

          {/* Legend */}
          <div className="mt-4 p-4 bg-muted/30 rounded-lg">
            <div className="text-sm font-semibold mb-2">Component Types</div>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="flex items-center gap-2 text-sm">
                <div className="w-4 h-4 rounded-full" style={{ backgroundColor: "#3b82f6" }} />
                <span>Service</span>
              </div>
              <div className="flex items-center gap-2 text-sm">
                <div className="w-4 h-4 rounded-full" style={{ backgroundColor: "#8b5cf6" }} />
                <span>Infrastructure</span>
              </div>
              <div className="flex items-center gap-2 text-sm">
                <div className="w-4 h-4 rounded-full" style={{ backgroundColor: "#10b981" }} />
                <span>Pipeline</span>
              </div>
              <div className="flex items-center gap-2 text-sm">
                <div className="w-4 h-4 rounded-full" style={{ backgroundColor: "#06b6d4" }} />
                <span>Backend</span>
              </div>
            </div>
            <div className="text-sm font-semibold mt-3 mb-2">Status</div>
            <div className="grid grid-cols-3 gap-4">
              <div className="flex items-center gap-2 text-sm">
                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: "#10b981" }} />
                <span>Healthy</span>
              </div>
              <div className="flex items-center gap-2 text-sm">
                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: "#f59e0b" }} />
                <span>Pending</span>
              </div>
              <div className="flex items-center gap-2 text-sm">
                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: "#ef4444" }} />
                <span>Failed</span>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

