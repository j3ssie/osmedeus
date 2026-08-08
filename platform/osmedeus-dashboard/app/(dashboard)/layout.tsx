"use client";

import * as React from "react";
import { usePathname } from "next/navigation";
import { useAuth } from "@/providers/auth-provider";
import { OrgProvider } from "@/providers/org-provider";
import { AppSidebar } from "@/components/layout/app-sidebar";
import { Topbar } from "@/components/layout/topbar";
import { Skeleton } from "@/components/ui/skeleton";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";
import { AnimatePresence, motion } from "motion/react";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isLoading, isAuthenticated } = useAuth();
  const [sidebarOpen, setSidebarOpen] = React.useState(true);
  const pathname = usePathname();
  const DISABLE_AUTH =
    typeof process !== "undefined" &&
    process.env.NEXT_PUBLIC_DISABLE_LOGIN === "true";

  // Persist the live open/collapsed state so it survives a refresh.
  const handleSidebarOpenChange = React.useCallback((open: boolean) => {
    setSidebarOpen(open);
    if (typeof window !== "undefined") {
      window.localStorage.setItem("osmedeus_sidebar_open", String(open));
    }
  }, []);

  // Restore the last state on mount: prefer the explicitly persisted state,
  // otherwise fall back to the "collapse by default" preference.
  React.useEffect(() => {
    if (typeof window === "undefined") return;
    const savedOpen = window.localStorage.getItem("osmedeus_sidebar_open");
    if (savedOpen !== null) {
      setSidebarOpen(savedOpen === "true");
      return;
    }
    const collapsedByDefault =
      window.localStorage.getItem("osmedeus_sidebar_collapsed_by_default") === "true";
    setSidebarOpen(!collapsedByDefault);
  }, []);

  React.useEffect(() => {
    if (typeof window === "undefined") return;
    const handler = (e: Event) => {
      const ev = e as CustomEvent<boolean>;
      handleSidebarOpenChange(ev.detail !== true);
    };
    window.addEventListener("osmedeus-sidebar-collapsed-by-default-changed", handler);
    return () => {
      window.removeEventListener("osmedeus-sidebar-collapsed-by-default-changed", handler);
    };
  }, [handleSidebarOpenChange]);

  // Show loading state while checking auth
  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="flex flex-col items-center gap-4">
          <Skeleton className="size-12 rounded-full" />
          <Skeleton className="h-4 w-32" />
        </div>
      </div>
    );
  }

  // Auth provider handles redirect, but we don't render if not authenticated
  if (!isAuthenticated && !DISABLE_AUTH) {
    return null;
  }

  return (
    <OrgProvider>
    <SidebarProvider open={sidebarOpen} onOpenChange={handleSidebarOpenChange}>
      <AppSidebar />
      <SidebarInset>
        <Topbar />
        <main className="flex-1 overflow-y-auto">
          <AnimatePresence initial={false} mode="wait">
            <motion.div
              key={pathname}
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -6 }}
              transition={{ duration: 0.2, ease: "easeOut" }}
              className="w-full max-w-none p-2.5 lg:px-gutter lg:py-2.5"
            >
              {children}
            </motion.div>
          </AnimatePresence>
        </main>
      </SidebarInset>
    </SidebarProvider>
    </OrgProvider>
  );
}
