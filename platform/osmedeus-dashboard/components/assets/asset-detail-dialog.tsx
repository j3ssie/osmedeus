"use client";

import * as React from "react";
import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
  DrawerDescription,
} from "@/components/ui/drawer";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { formatBytes, getStatusBadgeVariant } from "@/lib/utils";
import type { HttpAsset } from "@/lib/types/asset";
import {
  GlobeIcon,
  ServerIcon,
  ShieldIcon,
  CodeIcon,
  ClockIcon,
  ExternalLinkIcon,
} from "lucide-react";

interface AssetDetailDialogProps {
  asset: HttpAsset | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AssetDetailDialog({
  asset,
  open,
  onOpenChange,
}: AssetDetailDialogProps) {
  if (!asset) return null;
  const safeUrl = (asset.url ?? "").trim();

  return (
    <Drawer open={open} onOpenChange={onOpenChange} direction="right">
      <DrawerContent className="h-full w-full sm:max-w-2xl">
        <DrawerHeader>
          <DrawerTitle className="flex items-center gap-2">
            <GlobeIcon className="size-5" />
            Asset Details
          </DrawerTitle>
          <DrawerDescription className="font-mono text-xs break-all">
            {safeUrl || "-"}
          </DrawerDescription>
        </DrawerHeader>

        <ScrollArea className="flex-1 min-h-0">
          <div className="space-y-6 px-4 pb-6">
            <section className="space-y-3">
              <h3 className="text-sm font-semibold flex items-center gap-2">
                <GlobeIcon className="size-4" />
                Overview
              </h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 text-sm">
                <div>
                  <p className="text-muted-foreground">id</p>
                  <p className="font-mono text-xs">{asset.id}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">asset_value</p>
                  <p className="font-mono">{asset.assetValue || "-"}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">workspace</p>
                  <p>{asset.workspace || "-"}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">input</p>
                  <p className="font-mono text-xs">{asset.input || "-"}</p>
                </div>
                <div className="col-span-2">
                  <p className="text-muted-foreground">url</p>
                  <a
                    href={safeUrl || "#"}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="font-mono text-xs text-primary hover:underline flex items-center gap-1 break-all"
                  >
                    {safeUrl || "-"}
                    <ExternalLinkIcon className="size-3 shrink-0" />
                  </a>
                </div>
              </div>
            </section>

            <section className="space-y-3">
              <h3 className="text-sm font-semibold flex items-center gap-2">
                <CodeIcon className="size-4" />
                HTTP Request
              </h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 text-sm">
                <div>
                  <p className="text-muted-foreground">method</p>
                  <Badge variant="outline">{asset.method}</Badge>
                </div>
                <div>
                  <p className="text-muted-foreground">scheme</p>
                  <p>{asset.scheme || "-"}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">path</p>
                  <p className="font-mono text-xs">{asset.path || "/"}</p>
                </div>
              </div>
            </section>

            <section className="space-y-3">
              <h3 className="text-sm font-semibold">HTTP Response</h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 text-sm">
                <div>
                  <p className="text-muted-foreground">status_code</p>
                  <Badge variant={getStatusBadgeVariant(asset.statusCode)}>
                    {asset.statusCode}
                  </Badge>
                </div>
                <div>
                  <p className="text-muted-foreground">content_type</p>
                  <p className="text-xs">{asset.contentType || "-"}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">content_length</p>
                  <p>{formatBytes(asset.contentLength)}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">title</p>
                  <p className="truncate">{asset.title || "-"}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">words</p>
                  <p>{asset.words.toLocaleString()}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">lines</p>
                  <p>{asset.lines.toLocaleString()}</p>
                </div>
              </div>
            </section>

            <section className="space-y-3">
              <h3 className="text-sm font-semibold flex items-center gap-2">
                <ServerIcon className="size-4" />
                Network
              </h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 text-sm">
                <div>
                  <p className="text-muted-foreground">host_ip</p>
                  <p className="font-mono">{asset.hostIp || "-"}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">tls</p>
                  <p>{asset.tls || "-"}</p>
                </div>
                <div className="col-span-2">
                  <p className="text-muted-foreground">dns_records</p>
                  {asset.aRecords.length > 0 ? (
                    <div className="flex flex-wrap gap-1 mt-1">
                      {asset.aRecords.map((ip, i) => (
                        <Badge key={i} variant="outline" className="font-mono text-xs">
                          {ip}
                        </Badge>
                      ))}
                    </div>
                  ) : (
                    <p className="text-muted-foreground">-</p>
                  )}
                </div>
              </div>
            </section>

            <section className="space-y-3">
              <h3 className="text-sm font-semibold flex items-center gap-2">
                <ShieldIcon className="size-4" />
                Detection
              </h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 text-sm">
                <div>
                  <p className="text-muted-foreground">asset_type</p>
                  <Badge variant="secondary">{asset.assetType}</Badge>
                </div>
                <div>
                  <p className="text-muted-foreground">source</p>
                  <p>{asset.source || "-"}</p>
                </div>
                {asset.technologies.length > 0 && (
                  <div className="col-span-2">
                    <p className="text-muted-foreground">tech</p>
                    <div className="flex flex-wrap gap-1 mt-1">
                      {asset.technologies.map((tech, i) => (
                        <Badge key={i} variant="outline">
                          {tech}
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </section>

            <section className="space-y-3">
              <h3 className="text-sm font-semibold flex items-center gap-2">
                <ClockIcon className="size-4" />
                Meta
              </h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 text-sm">
                <div>
                  <p className="text-muted-foreground">time</p>
                  <p>{asset.responseTime || "-"}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">remarks</p>
                  <p>
                    {Array.isArray(asset.remarks)
                      ? asset.remarks.join(", ") || "-"
                      : asset.remarks || "-"}
                  </p>
                </div>
                <div>
                  <p className="text-muted-foreground">created_at</p>
                  <p className="text-xs">{asset.createdAt.toLocaleString()}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">updated_at</p>
                  <p className="text-xs">{asset.updatedAt.toLocaleString()}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">last_seen_at</p>
                  <p className="text-xs">
                    {asset.lastSeenAt?.toLocaleString?.() ?? "-"}
                  </p>
                </div>
              </div>
            </section>
          </div>
        </ScrollArea>
      </DrawerContent>
    </Drawer>
  );
}
