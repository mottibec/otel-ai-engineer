import type { ReactNode } from "react";
import { getCookie } from "@/lib/utils";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";

interface LayoutProps {
  sidebar: ReactNode;
  children: ReactNode;
  onRunCreated?: (runId: string) => void;
}

export function Layout({ sidebar, children }: LayoutProps) {
  // Read the sidebar state from cookie, default to true (expanded) if not found
  const sidebarCookie = getCookie("sidebar_state");
  const defaultOpen = sidebarCookie === null ? true : sidebarCookie === "true";

  return (
    <SidebarProvider
      defaultOpen={defaultOpen}
      className="flex h-screen w-screen overflow-hidden"
    >
      {sidebar}
      <SidebarInset className="flex-1 overflow-hidden">
        <div className="flex-1 h-screen min-h-0 overflow-auto">
          {children}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
