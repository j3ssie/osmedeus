"use client";

import * as React from "react";
import Link from "next/link";
import Image from "next/image";
import { usePathname } from "next/navigation";
import {
  LayoutDashboardIcon,
  ScanSearchIcon,
  WorkflowIcon,
  FolderOpenIcon,
  SettingsIcon,
  ShieldIcon,
  ClipboardCheckIcon,
  ScrollText as ScrollTextIcon,
  SquareFunction as SquareFunctionIcon,
  ArchiveIcon,
  BrainIcon,
} from "lucide-react";
import { cn } from "@/lib/utils";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Button } from "@/components/ui/button";
import logo from "@/osmedeus-logo.png";

const navigation = [
  {
    name: "Dashboard",
    href: "/",
    icon: LayoutDashboardIcon,
  },
  {
    name: "Scans",
    href: "/scans",
    icon: ScanSearchIcon,
  },
  {
    name: "Workflows",
    href: "/workflows",
    icon: WorkflowIcon,
  },
  {
    name: "Inventory",
    href: "/inventory",
    icon: FolderOpenIcon,
  },
  {
    name: "Artifacts",
    href: "/inventory/artifacts",
    icon: ArchiveIcon,
  },
  {
    name: "Schedules",
    href: "/schedules",
    icon: ClipboardCheckIcon,
  },
  {
    name: "Events",
    href: "/events",
    icon: ScrollTextIcon,
  },
  {
    name: "Utilities Functions",
    href: "/utilities",
    icon: SquareFunctionIcon,
  },
  {
    name: "LLM Playground",
    href: "/llm",
    icon: BrainIcon,
  },
  {
    name: "Settings",
    href: "/settings",
    icon: SettingsIcon,
  },
];

interface MobileNavProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function MobileNav({ open, onOpenChange }: MobileNavProps) {
  const pathname = usePathname();

  const isActive = (href: string) => {
    if (href === "/") {
      return pathname === "/";
    }
    return pathname.startsWith(href);
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="left" className="w-72 p-0">
        <SheetHeader className="border-b border-border px-gutter py-4">
          <SheetTitle className="flex items-center gap-3">
            <div className="flex size-8 items-center justify-center rounded-control bg-primary">
              <ShieldIcon className="size-5 text-primary-foreground" />
            </div>
            <Link href="/" onClick={() => onOpenChange(false)} aria-label="Go home">
              <Button variant="ghost" size="sm" className="gap-2">
                <Image src={logo} alt="Osmedeus" className="hidden md:block h-8 w-auto rotate-0 scale-x-100" priority />
              </Button>
            </Link>
          </SheetTitle>
        </SheetHeader>

        <ScrollArea className="h-[calc(100vh-5rem)]">
          <nav className="flex flex-col gap-1 p-4">
            {navigation.map((item) => {
              const active = isActive(item.href);

              return (
                <Link
                  key={item.name}
                  href={item.href}
                  onClick={() => onOpenChange(false)}
                  className={cn(
                    "flex items-center gap-3 rounded-control px-3 py-2.5 text-sm transition-colors duration-150",
                    active
                      ? "bg-sidebar-accent font-medium text-sidebar-accent-foreground"
                      : "text-muted-foreground hover:text-foreground"
                  )}
                >
                  <item.icon
                    className={cn(
                      "size-5",
                      active ? "text-primary" : "text-muted-foreground"
                    )}
                  />
                  {item.name}
                </Link>
              );
            })}
          </nav>

          <Separator className="mx-4" />

          <div className="p-4">
            <p className="px-3 text-xs font-medium text-muted-foreground">
              Version 1.0.0
            </p>
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  );
}
