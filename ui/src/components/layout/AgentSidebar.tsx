import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarRail,
  SidebarTrigger,
  useSidebar,
} from "@/components/ui/sidebar";
import { RunList } from "../runs/RunList";
import { StartRunModal } from "../runs/StartRunModal";
import { useState } from "react";

interface AgentSidebarProps {
  selectedRunId?: string;
  onSelectRun: (runId: string) => void;
  onRunCreated?: (runId: string) => void;
}

export function AgentSidebar({
  selectedRunId,
  onSelectRun,
  onRunCreated,
}: AgentSidebarProps) {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const { state } = useSidebar();

  return (
    <>
      <Sidebar collapsible="icon" className="border-r border-border">
        <SidebarRail />
        <SidebarHeader className="border-b border-border px-2 py-2">
          {state === "collapsed" ? (
            <div className="flex flex-col items-center gap-2">
              <SidebarTrigger className="h-6 w-6" />
              <Button
                onClick={() => setIsModalOpen(true)}
                variant="outline"
                size="icon"
                className="h-8 w-8 flex-shrink-0"
                title="Start new agent run"
              >
                <Plus className="h-4 w-4" />
              </Button>
            </div>
          ) : (
            <div className="flex flex-col gap-2">
              {/* Logo + Collapse */}
              <div className="flex items-center justify-between">
                <div className="flex-1 min-w-0">
                  <h1 className="text-sm font-semibold truncate">Agent Monitor</h1>
                  <p className="text-[10px] text-muted-foreground truncate">
                    OpenTelemetry AI
                  </p>
                </div>
                <SidebarTrigger className="h-6 w-6 flex-shrink-0" />
              </div>
              {/* New Chat Button */}
              <Button
                onClick={() => setIsModalOpen(true)}
                variant="outline"
                size="sm"
                className="h-7 px-2 gap-1 w-full"
                title="Start new agent run"
              >
                <Plus className="h-3 w-3" />
                <span className="text-xs">New Chat</span>
              </Button>
            </div>
          )}
        </SidebarHeader>

        <SidebarContent className="overflow-hidden">
          {state === "expanded" && (
            <div className="flex-1 overflow-hidden">
              <RunList
                selectedRunId={selectedRunId}
                onSelectRun={onSelectRun}
              />
            </div>
          )}
        </SidebarContent>
      </Sidebar>

      <StartRunModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onRunCreated={onRunCreated}
      />
    </>
  );
}

