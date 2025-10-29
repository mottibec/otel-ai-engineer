import { 
  Plus, 
  PlayCircle, 
  FileText, 
  FlaskConical, 
  Server, 
  Database, 
  Briefcase,
  Bot
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
  SidebarTrigger,
  useSidebar,
} from "@/components/ui/sidebar";
import { RunList } from "../runs/RunList";
import { StartRunModal } from "../runs/StartRunModal";
import { useState, useEffect } from "react";
import { Link, useLocation } from "react-router-dom";

interface AgentSidebarProps {
  selectedRunId?: string;
  onRunCreated?: (runId: string) => void;
  isCreateRunModalOpen?: boolean;
  onModalClose?: () => void;
}

export function AgentSidebar({
  selectedRunId,
  onRunCreated,
  isCreateRunModalOpen,
  onModalClose
}: AgentSidebarProps) {
  const [isModalOpen, setIsModalOpen] = useState(isCreateRunModalOpen ?? false);
  const { state } = useSidebar();
  const location = useLocation();

  // Sync parent state to internal state
  useEffect(() => {
    if (isCreateRunModalOpen !== undefined) {
      setIsModalOpen(isCreateRunModalOpen);
    }
  }, [isCreateRunModalOpen]);

  const handleClose = () => {
    setIsModalOpen(false);
    onModalClose?.();
  };

  const navigationItems = [
    { key: "runs", label: "Runs", icon: PlayCircle, path: "/runs" },
    { key: "plans", label: "Plans", icon: FileText, path: "/plans" },
    { key: "agents", label: "Agents", icon: Bot, path: "/agents" },
    { key: "sandboxes", label: "Lab", icon: FlaskConical, path: "/sandboxes" },
    { key: "collectors", label: "Collectors", icon: Server, path: "/collectors" },
    { key: "backends", label: "Backends", icon: Database, path: "/backends" },
    { key: "agent-work", label: "Agent Work", icon: Briefcase, path: "/agent-work" },
  ];

  // Determine active route based on current location
  const isActiveRoute = (path: string) => {
    if (path === "/runs") {
      return location.pathname.startsWith("/runs");
    }
    return location.pathname.startsWith(path);
  };

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
            <div className="flex items-center justify-between">
              <div className="flex-1 min-w-0">
                <h1 className="text-sm font-semibold truncate">Agent Monitor</h1>
                <p className="text-[10px] text-muted-foreground truncate">
                  OpenTelemetry AI
                </p>
              </div>
              <SidebarTrigger className="h-6 w-6 flex-shrink-0" />
            </div>
          )}
        </SidebarHeader>

        <SidebarContent className="overflow-hidden">
          {state === "expanded" && (
            <>
              {/* Navigation Section */}
              <SidebarGroup>
                <SidebarGroupLabel>Navigation</SidebarGroupLabel>
                <SidebarMenu>
                  {navigationItems.map((item) => {
                    const Icon = item.icon;
                    const isActive = isActiveRoute(item.path);
                    return (
                      <SidebarMenuItem key={item.key}>
                        <SidebarMenuButton
                          asChild
                          isActive={isActive}
                          tooltip={item.label}
                        >
                          <Link to={item.path}>
                            <Icon />
                            <span>{item.label}</span>
                          </Link>
                        </SidebarMenuButton>
                      </SidebarMenuItem>
                    );
                  })}
                </SidebarMenu>
              </SidebarGroup>

              {/* Runs Section - always visible */}
              <SidebarGroup className="flex-1 overflow-hidden">
                <SidebarGroupLabel>Runs</SidebarGroupLabel>
                <div className="flex-1 overflow-hidden">
                  <RunList selectedRunId={selectedRunId} />
                </div>
              </SidebarGroup>
            </>
          )}
        </SidebarContent>
      </Sidebar>

      <StartRunModal
        isOpen={isModalOpen}
        onClose={handleClose}
        onRunCreated={onRunCreated}
      />
    </>
  );
}

