"use client";

import * as React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "@/providers/auth-provider";
import { ThemeToggle } from "./theme-toggle";
import { OrgSwitcher } from "./org-switcher";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";
import {
  Breadcrumb,
  BreadcrumbList,
  BreadcrumbItem,
  BreadcrumbSeparator,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { LogOutIcon, UserIcon, CircleUserRound, BookOpenIcon } from "lucide-react";
import { GithubIcon } from "@/components/icons/github-icon";
import { getBreadcrumbs } from "./breadcrumbs-util";
import { getRouteMeta } from "./route-meta";
import { getServerInfo } from "@/lib/api/system";
import { isDemoMode } from "@/lib/api/demo-mode";

export function Topbar() {
  const { user, logout } = useAuth();
  const pathname = usePathname();
  const [serverUp, setServerUp] = React.useState<boolean | null>(null);
  const [serverVersion, setServerVersion] = React.useState<string | null>(null);
  const [serverLicense, setServerLicense] = React.useState<string>("open-source");
  const [demoMode, setDemoMode] = React.useState<boolean>(false);

  const crumbs = React.useMemo(() => getBreadcrumbs(pathname || "/"), [pathname]);
  const meta = React.useMemo(() => getRouteMeta(pathname || "/"), [pathname]);

  React.useEffect(() => {
    let mounted = true;
    const demo = isDemoMode();
    setDemoMode(demo);
    if (demo) {
      setServerUp(false);
      setServerVersion("demo");
      setServerLicense("open-source");
      return () => {
        mounted = false;
      };
    }
    const check = async () => {
      try {
        const info = await getServerInfo();
        if (!mounted) return;
        setServerUp(true);
        setServerVersion(info.version);
        setServerLicense(info.license ?? "open-source");
      } catch {
        if (!mounted) return;
        setServerUp(false);
        setServerVersion(null);
        setServerLicense("open-source");
      }
    };
    check();
    const id = setInterval(check, 60000);
    return () => {
      mounted = false;
      clearInterval(id);
    };
  }, []);

  return (
    <header className="sticky top-0 z-40 flex h-12 items-center justify-between gap-4 border-b border-border bg-background px-gutter text-sm">
      <div className="flex items-center gap-2">
        <Breadcrumb>
          <BreadcrumbList>
            {crumbs.map((c, idx) => (
              <React.Fragment key={`${c.label}-${idx}`}>
                <BreadcrumbItem>
                  {c.isCurrent ? (
                    <BreadcrumbPage className="flex items-center gap-2">
                      {c.icon && (
                        <span className="inline-flex size-6 items-center justify-center rounded-control bg-primary-soft text-primary-fg">
                          <c.icon className="size-4" />
                        </span>
                      )}
                      {meta.title || c.label}
                      {meta.description && (
                        <span className="text-muted-foreground">
                          {" · "}{meta.description}
                        </span>
                      )}
                    </BreadcrumbPage>
                  ) : (
                    c.href ? (
                      <Link href={c.href} className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground">
                        {c.icon && (
                          <span className="inline-flex size-6 items-center justify-center rounded-md bg-primary-soft text-primary">
                            <c.icon className="size-4" />
                          </span>
                        )}
                        {c.label}
                      </Link>
                    ) : (
                      <span className="flex items-center gap-2 text-sm text-muted-foreground">
                        {c.icon && (
                          <span className="inline-flex size-6 items-center justify-center rounded-md bg-primary-soft text-primary">
                            <c.icon className="size-4" />
                          </span>
                        )}
                        {c.label}
                      </span>
                    )
                  )}
                </BreadcrumbItem>
                {idx < crumbs.length - 1 && <BreadcrumbSeparator />}
              </React.Fragment>
            ))}
          </BreadcrumbList>
        </Breadcrumb>
      </div>

      {/* Right side actions */}
      <div className="hidden md:flex items-center gap-2">
        <OrgSwitcher />
        <Tooltip>
          <TooltipTrigger asChild>
            <div
              aria-label="Server status"
              className="inline-flex items-center gap-2 rounded-control border border-border bg-muted px-2.5 py-1"
            >
              <span
                className={`inline-block size-2 rounded-full ${
                  serverUp === null
                    ? "bg-faint"
                    : serverUp
                      ? "bg-success"
                      : demoMode
                        ? "bg-warning"
                        : "bg-destructive"
                }`}
              />
              <span className="text-xs">
                {serverUp === null
                  ? "Checking..."
                  : serverUp
                    ? "Healthy"
                    : demoMode
                      ? "Demo Mode"
                      : "Offline"}
              </span>
            </div>
          </TooltipTrigger>
          <TooltipContent>
            {serverUp === null
              ? "Checking server..."
              : serverUp
                ? "Server is healthy"
                : demoMode
                  ? "Running in demo mode with mock data"
                  : "Server is offline"}
          </TooltipContent>
        </Tooltip>
        <Tooltip>
          <TooltipTrigger asChild>
            <div
              aria-label="Server license"
              className="inline-flex items-center gap-2 px-2 py-1 rounded-md border bg-muted"
            >
              <span className="inline-block size-2 rounded-full bg-purple" />
              <span className="text-xs">{serverLicense}</span>
            </div>
          </TooltipTrigger>
          <TooltipContent>License: {serverLicense}</TooltipContent>
        </Tooltip>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="outline" size="icon" className="size-8" aria-label="Open documentation" asChild>
              <a
                href="https://docs.osmedeus.org/"
                target="_blank"
                rel="noopener noreferrer"
              >
                <BookOpenIcon className="size-4" />
              </a>
            </Button>
          </TooltipTrigger>
          <TooltipContent>Documentation</TooltipContent>
        </Tooltip>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="outline" size="icon" className="size-8" aria-label="Open GitHub repository" asChild>
              <a
                href="https://github.com/j3ssie/osmedeus"
                target="_blank"
                rel="noopener noreferrer"
              >
                <GithubIcon className="size-4" />
              </a>
            </Button>
          </TooltipTrigger>
          <TooltipContent>GitHub</TooltipContent>
        </Tooltip>
        <Tooltip>
          <TooltipTrigger asChild>
            <ThemeToggle variant="outline" className="size-8" />
          </TooltipTrigger>
          <TooltipContent>Theme</TooltipContent>
        </Tooltip>

        {user && (
          <DropdownMenu>
            <Tooltip>
              <TooltipTrigger asChild>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    className="relative size-8 p-0"
                    aria-label="User menu"
                  >
                    <Avatar className="size-8 rounded-control">
                      <AvatarFallback className="bg-muted text-muted-foreground rounded-control">
                        <CircleUserRound className="size-5" />
                      </AvatarFallback>
                    </Avatar>
                  </Button>
                </DropdownMenuTrigger>
              </TooltipTrigger>
              <TooltipContent side="bottom">Profile</TooltipContent>
            </Tooltip>
            <DropdownMenuContent align="end" className="w-56">
              <DropdownMenuLabel>
                <div className="flex flex-col gap-1">
                  <p className="text-sm font-medium">{user.name ?? user.username}</p>
                  <p className="text-xs text-muted-foreground">{user.email}</p>
                </div>
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem asChild>
                <Link href="/settings" className="cursor-pointer">
                  <UserIcon className="mr-2 size-4" />
                  Profile Settings
                </Link>
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onClick={logout}
                className="text-destructive focus:text-destructive cursor-pointer"
              >
                <LogOutIcon className="mr-2 size-4" />
                Sign out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </div>
    </header>
  );
}
