"use client";

import * as React from "react";
import { cn } from "@/lib/utils";

function Breadcrumb({
  className,
  ...props
}: React.HTMLAttributes<HTMLElement>) {
  return (
    <nav
      aria-label="breadcrumb"
      data-slot="breadcrumb"
      className={cn("flex items-center", className)}
      {...props}
    />
  );
}

function BreadcrumbList({
  className,
  ...props
}: React.HTMLAttributes<HTMLOListElement>) {
  return (
    <ol
      data-slot="breadcrumb-list"
      className={cn("flex items-center gap-2", className)}
      {...props}
    />
  );
}

function BreadcrumbItem({
  className,
  ...props
}: React.HTMLAttributes<HTMLLIElement>) {
  return (
    <li
      data-slot="breadcrumb-item"
      className={cn("flex items-center", className)}
      {...props}
    />
  );
}

function BreadcrumbSeparator({
  className,
  children = ">",
  ...props
}: React.HTMLAttributes<HTMLSpanElement>) {
  return (
    <span
      role="presentation"
      aria-hidden="true"
      data-slot="breadcrumb-separator"
      className={cn("text-muted-foreground", className)}
      {...props}
    >
      {children}
    </span>
  );
}

function BreadcrumbPage({
  className,
  ...props
}: React.HTMLAttributes<HTMLSpanElement>) {
  return (
    <span
      aria-current="page"
      data-slot="breadcrumb-page"
      className={cn("text-lg font-semibold tracking-tight", className)}
      {...props}
    />
  );
}

function BreadcrumbText({
  className,
  ...props
}: React.HTMLAttributes<HTMLSpanElement>) {
  return (
    <span
      data-slot="breadcrumb-text"
      className={cn("text-sm text-muted-foreground font-normal", className)}
      {...props}
    />
  );
}

export {
  Breadcrumb,
  BreadcrumbList,
  BreadcrumbItem,
  BreadcrumbSeparator,
  BreadcrumbPage,
  BreadcrumbText,
};
