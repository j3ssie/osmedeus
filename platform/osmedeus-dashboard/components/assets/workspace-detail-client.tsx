"use client";

import * as React from "react";
import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { AssetFilters } from "@/components/assets/asset-filters";
import { HttpAssetsTable } from "@/components/assets/http-assets-table";
import { ErrorState } from "@/components/shared/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { fetchWorkspace, fetchHttpAssets, fetchAssetStats } from "@/lib/api/assets";
import { formatNumber } from "@/lib/utils";
import type { Workspace, HttpAsset, HttpAssetFilters, AssetSortState, AssetSortField, AssetStats } from "@/lib/types/asset";
import { sortAssets } from "@/lib/utils";
import type { PaginatedResponse } from "@/lib/types/api";
import { ArrowLeftIcon, GlobeIcon, LinkIcon, AlertTriangleIcon } from "lucide-react";

interface WorkspaceDetailClientProps {
  workspaceId: string;
}

export default function WorkspaceDetailClient({ workspaceId }: WorkspaceDetailClientProps) {
  const [workspace, setWorkspace] = React.useState<Workspace | null>(null);
  const [assetsResponse, setAssetsResponse] = React.useState<PaginatedResponse<HttpAsset> | null>(null);
  const [isLoadingWorkspace, setIsLoadingWorkspace] = React.useState(true);
  const [isLoadingAssets, setIsLoadingAssets] = React.useState(true);
  const [assetStats, setAssetStats] = React.useState<AssetStats | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [filters, setFilters] = React.useState<HttpAssetFilters>({});
  const [pageSize, setPageSize] = React.useState(50);
  const [page, setPage] = React.useState(1);
  const [sortState, setSortState] = React.useState<AssetSortState>({
    field: null,
    direction: "asc",
  });

  React.useEffect(() => {
    const loadWorkspace = async () => {
      try {
        setIsLoadingWorkspace(true);
        const data = await fetchWorkspace(workspaceId);
        setWorkspace(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load workspace");
      } finally {
        setIsLoadingWorkspace(false);
      }
    };
    loadWorkspace();
  }, [workspaceId]);

  React.useEffect(() => {
    const loadStats = async () => {
      try {
        const workspaceName = workspace?.name ?? workspaceId;
        const stats = await fetchAssetStats(workspaceName);
        setAssetStats(stats);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load asset statistics");
      }
    };
    loadStats();
  }, [workspace?.name, workspaceId]);

  React.useEffect(() => {
    const loadAssets = async () => {
      try {
        setIsLoadingAssets(true);
        const data = await fetchHttpAssets(workspaceId, {
          page,
          pageSize,
          filters,
        });
        setAssetsResponse(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load assets");
      } finally {
        setIsLoadingAssets(false);
      }
    };
    loadAssets();
  }, [workspaceId, page, pageSize, filters]);

  const handleFiltersChange = React.useCallback((newFilters: HttpAssetFilters) => {
    setFilters(newFilters);
    setPage(1);
  }, []);

  const handleSort = React.useCallback((field: AssetSortField) => {
    setSortState((prev) => ({
      field,
      direction:
        prev.field === field && prev.direction === "asc" ? "desc" : "asc",
    }));
  }, []);

  const clientFilteredAssets = React.useMemo(() => {
    const assets = assetsResponse?.data ?? [];
    const q = (filters.search ?? "").trim().toLowerCase();

    return assets.filter((a) => {
      if (q) {
        const hay = [a.url, a.title ?? "", a.assetValue, a.hostIp ?? ""]
          .join(" ")
          .toLowerCase();
        if (!hay.includes(q)) return false;
      }
      if (filters.statusCodes?.length) {
        if (!filters.statusCodes.includes(a.statusCode)) return false;
      }
      if (filters.technologies?.length) {
        const techSet = new Set(
          a.technologies.map((t) => String(t).trim().toLowerCase())
        );
        const wanted = filters.technologies.map((t) => t.trim().toLowerCase());
        if (!wanted.some((t) => techSet.has(t))) return false;
      }
      if (filters.assetTypes?.length) {
        const assetType = String(a.assetType ?? "").trim().toLowerCase();
        const wanted = filters.assetTypes.map((t) => t.trim().toLowerCase());
        if (!wanted.includes(assetType)) return false;
      }
      if (filters.sources?.length) {
        const source = String(a.source ?? "").trim().toLowerCase();
        const wanted = filters.sources.map((t) => t.trim().toLowerCase());
        if (!wanted.includes(source)) return false;
      }
      if (filters.remarks?.length) {
        const remarks = Array.isArray(a.remarks)
          ? a.remarks
          : a.remarks
            ? [a.remarks]
            : [];
        const remarkSet = new Set(
          remarks.map((t) => String(t).trim().toLowerCase()).filter(Boolean)
        );
        const wanted = filters.remarks.map((t) => t.trim().toLowerCase());
        if (!wanted.some((t) => remarkSet.has(t))) return false;
      }
      return true;
    });
  }, [
    assetsResponse?.data,
    filters.search,
    filters.statusCodes,
    filters.technologies,
    filters.assetTypes,
    filters.sources,
    filters.remarks,
  ]);

  const sortedAssets = React.useMemo(() => {
    return sortAssets(clientFilteredAssets, sortState.field, sortState.direction);
  }, [clientFilteredAssets, sortState]);

  const fallbackAssetTypes = React.useMemo(() => {
    const values = (assetsResponse?.data ?? [])
      .map((asset) => asset.assetType)
      .filter((value) => value && value.trim().length > 0);
    return Array.from(new Set(values)).sort((a, b) => a.localeCompare(b));
  }, [assetsResponse?.data]);

  const fallbackSources = React.useMemo(() => {
    const values = (assetsResponse?.data ?? [])
      .map((asset) => asset.source)
      .filter((value) => value && value.trim().length > 0);
    return Array.from(new Set(values)).sort((a, b) => a.localeCompare(b));
  }, [assetsResponse?.data]);

  const fallbackRemarks = React.useMemo(() => {
    const values = (assetsResponse?.data ?? [])
      .flatMap((asset) =>
        Array.isArray(asset.remarks)
          ? asset.remarks
          : asset.remarks
            ? [asset.remarks]
            : []
      )
      .map((value) => String(value).trim())
      .filter((value) => value.length > 0);
    return Array.from(new Set(values)).sort((a, b) => a.localeCompare(b));
  }, [assetsResponse?.data]);

  const assetTypeOptions =
    assetStats?.assetTypes?.length ? assetStats.assetTypes : fallbackAssetTypes;
  const sourceOptions =
    assetStats?.sources?.length ? assetStats.sources : fallbackSources;
  const technologyOptions =
    assetStats?.technologies?.length ? assetStats.technologies : [];
  const remarkOptions =
    assetStats?.remarks?.length ? assetStats.remarks : fallbackRemarks;

  const hasActiveFilters = React.useMemo(() => {
    return !!(
      filters.search ||
      filters.statusCodes?.length ||
      filters.technologies?.length ||
      filters.assetTypes?.length ||
      filters.sources?.length ||
      filters.remarks?.length ||
      filters.contentTypes?.length ||
      filters.tlsVersion ||
      filters.location
    );
  }, [filters]);

  const pageSizeOptions = React.useMemo(() => [50, 100, 200, 500], []);

  if (error && !workspace) {
    return <ErrorState message={error} />;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-start gap-4">
        <Button variant="ghost" size="icon" asChild>
          <Link href="/inventory/assets">
            <ArrowLeftIcon className="size-4" />
            <span className="sr-only">Back to assets</span>
          </Link>
        </Button>
        <div className="flex-1">
          {isLoadingWorkspace ? (
            <div className="space-y-2">
              <Skeleton className="h-8 w-48" />
              <Skeleton className="h-4 w-32" />
            </div>
          ) : workspace ? (
            <>
              <h1 className="text-2xl font-bold tracking-tight">{workspace.name}</h1>
              <p className="font-mono text-muted-foreground">{workspace.local_path}</p>
            </>
          ) : null}
        </div>
      </div>

      {workspace && (
        <div className="grid gap-4 md:grid-cols-3">
          <Card>
            <CardContent className="flex items-center gap-4 pt-6">
              <div className="flex size-12 items-center justify-center rounded-lg bg-primary-soft">
                <GlobeIcon className="size-6 text-primary" />
              </div>
              <div>
                <p className="text-2xl font-bold">{formatNumber(workspace.total_subdomains)}</p>
                <p className="text-sm text-muted-foreground">Subdomains</p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="flex items-center gap-4 pt-6">
              <div className="flex size-12 items-center justify-center rounded-lg bg-primary-soft">
                <LinkIcon className="size-6 text-primary" />
              </div>
              <div>
                <p className="text-2xl font-bold">{formatNumber(workspace.total_urls)}</p>
                <p className="text-sm text-muted-foreground">HTTP Assets</p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="flex items-center gap-4 pt-6">
              <div className="flex size-12 items-center justify-center rounded-lg bg-destructive-soft">
                <AlertTriangleIcon className="size-6 text-destructive" />
              </div>
              <div className="flex items-center gap-2">
                <p className="text-2xl font-bold">{formatNumber(workspace.total_vulns)}</p>
                {workspace.total_vulns > 0 && (
                  <Badge variant="destructive" className="text-xs">
                    Action needed
                  </Badge>
                )}
              </div>
              <p className="text-sm text-muted-foreground sr-only">Vulnerabilities</p>
            </CardContent>
          </Card>
        </div>
      )}

      <AssetFilters
        filters={filters}
        onFiltersChange={handleFiltersChange}
        technologyOptions={technologyOptions}
        assetTypeOptions={assetTypeOptions}
        sourceOptions={sourceOptions}
        remarkOptions={remarkOptions}
        pageSize={pageSize}
        pageSizeOptions={pageSizeOptions}
        onPageSizeChange={(size) => {
          setPageSize(size);
          setPage(1);
        }}
      />

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">HTTP Assets</CardTitle>
        </CardHeader>
        <CardContent>
          <HttpAssetsTable
            assets={sortedAssets}
            isLoading={isLoadingAssets}
            pagination={assetsResponse?.pagination}
            sortState={sortState}
            onSort={handleSort}
            onPageChange={setPage}
            hasActiveFilters={hasActiveFilters}
          />
        </CardContent>
      </Card>
    </div>
  );
}
