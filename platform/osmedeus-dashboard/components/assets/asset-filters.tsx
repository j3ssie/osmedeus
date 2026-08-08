"use client";

import * as React from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Checkbox } from "@/components/ui/checkbox";
import { ScrollArea } from "@/components/ui/scroll-area";
import type { HttpAssetFilters } from "@/lib/types/asset";
import {
  BlocksIcon,
  SearchIcon,
  HashIcon,
  XIcon,
  ChevronsUpDownIcon,
  DatabaseIcon,
  LayersIcon,
  Columns3Icon,
  TagIcon,
  ListIcon,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface AssetFiltersProps {
  filters: HttpAssetFilters;
  onFiltersChange: (filters: HttpAssetFilters) => void;
  technologyOptions?: string[];
  assetTypeOptions?: string[];
  sourceOptions?: string[];
  remarkOptions?: string[];
  columnOptions?: Array<{ key: string; label: string }>;
  visibleColumns?: Record<string, boolean>;
  onVisibleColumnsChange?: (next: Record<string, boolean>) => void;
  pageSize?: number;
  pageSizeOptions?: number[];
  onPageSizeChange?: (pageSize: number) => void;
}

const statusCodeOptions = [
  { value: 200, label: "200 OK" },
  { value: 301, label: "301 Redirect" },
  { value: 302, label: "302 Found" },
  { value: 403, label: "403 Forbidden" },
  { value: 404, label: "404 Not Found" },
  { value: 500, label: "500 Server Error" },
];

export function AssetFilters({
  filters,
  onFiltersChange,
  technologyOptions = [],
  assetTypeOptions = [],
  sourceOptions = [],
  remarkOptions = [],
  columnOptions = [],
  visibleColumns,
  onVisibleColumnsChange,
  pageSize,
  pageSizeOptions = [],
  onPageSizeChange,
}: AssetFiltersProps) {
  const [searchValue, setSearchValue] = React.useState(filters.search ?? "");
  const [techSearch, setTechSearch] = React.useState("");
  const [assetTypeSearch, setAssetTypeSearch] = React.useState("");
  const [sourceSearch, setSourceSearch] = React.useState("");
  const [remarkSearch, setRemarkSearch] = React.useState("");

  const primaryTriggerClassName = "h-9 justify-between min-w-[140px]";

  // Count active filters
  const activeFilterCount = React.useMemo(() => {
    let count = 0;
    if (filters.search) count++;
    if (filters.statusCodes?.length) count++;
    if (filters.technologies?.length) count++;
    if (filters.assetTypes?.length) count++;
    if (filters.sources?.length) count++;
    if (filters.remarks?.length) count++;
    if (filters.contentTypes?.length) count++;
    if (filters.tlsVersion) count++;
    if (filters.location) count++;
    return count;
  }, [filters]);

  const visibleColumnCount = React.useMemo(() => {
    if (!visibleColumns) return 0;
    return Object.values(visibleColumns).filter(Boolean).length;
  }, [visibleColumns]);

  const applySearch = () => {
    if (searchValue !== filters.search) {
      onFiltersChange({ ...filters, search: searchValue || undefined });
    }
  };

  const handleStatusCodeToggle = (statusCode: number) => {
    const current = filters.statusCodes ?? [];
    const updated = current.includes(statusCode)
      ? current.filter((c) => c !== statusCode)
      : [...current, statusCode];
    const normalized = updated
      .filter((c, i) => updated.indexOf(c) === i)
      .sort((a, b) => a - b);
    onFiltersChange({
      ...filters,
      statusCodes: normalized.length > 0 ? normalized : undefined,
    });
  };

  const handleTechToggle = (tech: string) => {
    const current = filters.technologies ?? [];
    const updated = current.includes(tech)
      ? current.filter((t) => t !== tech)
      : [...current, tech];
    onFiltersChange({
      ...filters,
      technologies: updated.length > 0 ? updated : undefined,
    });
  };

  const handleAssetTypeToggle = (assetType: string) => {
    const current = filters.assetTypes ?? [];
    const updated = current.includes(assetType)
      ? current.filter((t) => t !== assetType)
      : [...current, assetType];
    onFiltersChange({
      ...filters,
      assetTypes: updated.length > 0 ? updated : undefined,
    });
  };

  const handleSourceToggle = (source: string) => {
    const current = filters.sources ?? [];
    const updated = current.includes(source)
      ? current.filter((t) => t !== source)
      : [...current, source];
    onFiltersChange({
      ...filters,
      sources: updated.length > 0 ? updated : undefined,
    });
  };

  const handleRemarkToggle = (remark: string) => {
    const current = filters.remarks ?? [];
    const updated = current.includes(remark)
      ? current.filter((t) => t !== remark)
      : [...current, remark];
    onFiltersChange({
      ...filters,
      remarks: updated.length > 0 ? updated : undefined,
    });
  };

  const handleClearFilters = () => {
    setSearchValue("");
    onFiltersChange({});
  };

  const filteredTechnologies = React.useMemo(() => {
    if (!techSearch) return technologyOptions;
    const search = techSearch.toLowerCase();
    return technologyOptions.filter((tech) =>
      tech.toLowerCase().includes(search)
    );
  }, [techSearch, technologyOptions]);

  const filteredAssetTypes = React.useMemo(() => {
    if (!assetTypeSearch) return assetTypeOptions;
    const search = assetTypeSearch.toLowerCase();
    return assetTypeOptions.filter((type) =>
      type.toLowerCase().includes(search)
    );
  }, [assetTypeOptions, assetTypeSearch]);

  const filteredSources = React.useMemo(() => {
    if (!sourceSearch) return sourceOptions;
    const search = sourceSearch.toLowerCase();
    return sourceOptions.filter((type) => type.toLowerCase().includes(search));
  }, [sourceOptions, sourceSearch]);

  const filteredRemarks = React.useMemo(() => {
    if (!remarkSearch) return remarkOptions;
    const search = remarkSearch.toLowerCase();
    return remarkOptions.filter((type) => type.toLowerCase().includes(search));
  }, [remarkOptions, remarkSearch]);

  const handleColumnToggle = (key: string) => {
    if (!visibleColumns || !onVisibleColumnsChange) return;
    const next = {
      ...visibleColumns,
      [key]: !visibleColumns[key],
    };
    if (Object.values(next).some(Boolean)) {
      onVisibleColumnsChange(next);
    }
  };

  return (
    <div className="space-y-3">
      {/* Primary Filters Row */}
      <div className="flex flex-wrap items-center gap-3">
        {/* Search */}
        <div className="relative w-full sm:w-[260px]">
          <SearchIcon className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search URL, title, or host..."
            value={searchValue}
            onChange={(e) => setSearchValue(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") applySearch();
            }}
            onBlur={applySearch}
            className="pl-9 h-9"
          />
        </div>

        {/* Status Code Multi-Select */}
        <Popover>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              className={cn(
                primaryTriggerClassName,
                (filters.statusCodes?.length ?? 0) > 0 && "border-primary"
              )}
            >
              <span className="flex items-center gap-2">
                <HashIcon className="size-4" />
                <span>Status Codes</span>
                {(filters.statusCodes?.length ?? 0) > 0 && (
                  <Badge variant="secondary" className="px-1.5 py-0 text-xs">
                    {filters.statusCodes?.length}
                  </Badge>
                )}
              </span>
              <ChevronsUpDownIcon className="size-4 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-[200px] p-2" align="start">
            <div className="space-y-1">
              {statusCodeOptions.map((option) => (
                <label
                  key={option.value}
                  className="flex items-center gap-2 px-2 py-1.5 rounded-sm hover:bg-muted cursor-pointer text-sm"
                >
                  <Checkbox
                    checked={filters.statusCodes?.includes(option.value) ?? false}
                    onCheckedChange={() => handleStatusCodeToggle(option.value)}
                  />
                  <span>{option.label}</span>
                </label>
              ))}
            </div>
            {(filters.statusCodes?.length ?? 0) > 0 && (
              <div className="pt-2 mt-2 border-t">
                <Button
                  variant="ghost"
                  size="sm"
                  className="w-full h-8"
                  onClick={() => onFiltersChange({ ...filters, statusCodes: undefined })}
                >
                  Clear selection
                </Button>
              </div>
            )}
          </PopoverContent>
        </Popover>

        {/* Technology Multi-Select */}
        <Popover>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              className={cn(
                primaryTriggerClassName,
                (filters.technologies?.length ?? 0) > 0 && "border-primary"
              )}
            >
              <span className="flex items-center gap-2">
                <BlocksIcon className="size-4" />
                <span>Technologies</span>
                {(filters.technologies?.length ?? 0) > 0 && (
                  <Badge variant="secondary" className="px-1.5 py-0 text-xs">
                    {filters.technologies?.length}
                  </Badge>
                )}
              </span>
              <ChevronsUpDownIcon className="size-4 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-[240px] p-0" align="start">
            <div className="p-2 border-b">
              <Input
                placeholder="Search technologies..."
                value={techSearch}
                onChange={(e) => setTechSearch(e.target.value)}
                className="h-8"
              />
            </div>
            <ScrollArea className="h-[240px]">
              <div className="p-2 space-y-1">
                {filteredTechnologies.map((tech) => (
                  <label
                    key={tech}
                    className="flex items-center gap-2 px-2 py-1.5 rounded-sm hover:bg-muted cursor-pointer text-sm"
                  >
                    <Checkbox
                      checked={filters.technologies?.includes(tech) ?? false}
                      onCheckedChange={() => handleTechToggle(tech)}
                    />
                    <span className="capitalize">{tech}</span>
                  </label>
                ))}
                {filteredTechnologies.length === 0 && (
                  <p className="text-sm text-muted-foreground text-center py-4">
                    No technologies found
                  </p>
                )}
              </div>
            </ScrollArea>
            {(filters.technologies?.length ?? 0) > 0 && (
              <div className="p-2 border-t">
                <Button
                  variant="ghost"
                  size="sm"
                  className="w-full h-8"
                  onClick={() =>
                    onFiltersChange({ ...filters, technologies: undefined })
                  }
                >
                  Clear selection
                </Button>
              </div>
            )}
          </PopoverContent>
        </Popover>

        {/* Asset Type Multi-Select */}
        <Popover>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              className={cn(
                primaryTriggerClassName,
                (filters.assetTypes?.length ?? 0) > 0 && "border-primary"
              )}
            >
              <span className="flex items-center gap-2">
                <LayersIcon className="size-4" />
                <span>Asset Type</span>
                {(filters.assetTypes?.length ?? 0) > 0 && (
                  <Badge variant="secondary" className="px-1.5 py-0 text-xs">
                    {filters.assetTypes?.length}
                  </Badge>
                )}
              </span>
              <ChevronsUpDownIcon className="size-4 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-[220px] p-0" align="start">
            <div className="p-2 border-b">
              <Input
                placeholder="Search asset types..."
                value={assetTypeSearch}
                onChange={(e) => setAssetTypeSearch(e.target.value)}
                className="h-8"
              />
            </div>
            <ScrollArea className="h-[220px]">
              <div className="p-2 space-y-1">
                {filteredAssetTypes.map((type) => (
                  <label
                    key={type}
                    className="flex items-center gap-2 px-2 py-1.5 rounded-sm hover:bg-muted cursor-pointer text-sm"
                  >
                    <Checkbox
                      checked={filters.assetTypes?.includes(type) ?? false}
                      onCheckedChange={() => handleAssetTypeToggle(type)}
                    />
                    <span>{type}</span>
                  </label>
                ))}
                {filteredAssetTypes.length === 0 && (
                  <p className="text-sm text-muted-foreground text-center py-4">
                    No asset types found
                  </p>
                )}
              </div>
            </ScrollArea>
            {(filters.assetTypes?.length ?? 0) > 0 && (
              <div className="p-2 border-t">
                <Button
                  variant="ghost"
                  size="sm"
                  className="w-full h-8"
                  onClick={() =>
                    onFiltersChange({ ...filters, assetTypes: undefined })
                  }
                >
                  Clear selection
                </Button>
              </div>
            )}
          </PopoverContent>
        </Popover>

        {/* Sources Multi-Select */}
        <Popover>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              className={cn(
                primaryTriggerClassName,
                (filters.sources?.length ?? 0) > 0 && "border-primary"
              )}
            >
              <span className="flex items-center gap-2">
                <DatabaseIcon className="size-4" />
                <span>Sources</span>
                {(filters.sources?.length ?? 0) > 0 && (
                  <Badge variant="secondary" className="px-1.5 py-0 text-xs">
                    {filters.sources?.length}
                  </Badge>
                )}
              </span>
              <ChevronsUpDownIcon className="size-4 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-[220px] p-0" align="start">
            <div className="p-2 border-b">
              <Input
                placeholder="Search sources..."
                value={sourceSearch}
                onChange={(e) => setSourceSearch(e.target.value)}
                className="h-8"
              />
            </div>
            <ScrollArea className="h-[220px]">
              <div className="p-2 space-y-1">
                {filteredSources.map((source) => (
                  <label
                    key={source}
                    className="flex items-center gap-2 px-2 py-1.5 rounded-sm hover:bg-muted cursor-pointer text-sm"
                  >
                    <Checkbox
                      checked={filters.sources?.includes(source) ?? false}
                      onCheckedChange={() => handleSourceToggle(source)}
                    />
                    <span>{source}</span>
                  </label>
                ))}
                {filteredSources.length === 0 && (
                  <p className="text-sm text-muted-foreground text-center py-4">
                    No sources found
                  </p>
                )}
              </div>
            </ScrollArea>
            {(filters.sources?.length ?? 0) > 0 && (
              <div className="p-2 border-t">
                <Button
                  variant="ghost"
                  size="sm"
                  className="w-full h-8"
                  onClick={() =>
                    onFiltersChange({ ...filters, sources: undefined })
                  }
                >
                  Clear selection
                </Button>
              </div>
            )}
          </PopoverContent>
        </Popover>

        {/* Remarks Multi-Select */}
        <Popover>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              className={cn(
                primaryTriggerClassName,
                (filters.remarks?.length ?? 0) > 0 && "border-primary"
              )}
            >
              <span className="flex items-center gap-2">
                <TagIcon className="size-4" />
                <span>Remarks</span>
                {(filters.remarks?.length ?? 0) > 0 && (
                  <Badge variant="secondary" className="px-1.5 py-0 text-xs">
                    {filters.remarks?.length}
                  </Badge>
                )}
              </span>
              <ChevronsUpDownIcon className="size-4 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-[220px] p-0" align="start">
            <div className="p-2 border-b">
              <Input
                placeholder="Search remarks..."
                value={remarkSearch}
                onChange={(e) => setRemarkSearch(e.target.value)}
                className="h-8"
              />
            </div>
            <ScrollArea className="h-[220px]">
              <div className="p-2 space-y-1">
                {filteredRemarks.map((remark) => (
                  <label
                    key={remark}
                    className="flex items-center gap-2 px-2 py-1.5 rounded-sm hover:bg-muted cursor-pointer text-sm"
                  >
                    <Checkbox
                      checked={filters.remarks?.includes(remark) ?? false}
                      onCheckedChange={() => handleRemarkToggle(remark)}
                    />
                    <span>{remark}</span>
                  </label>
                ))}
                {filteredRemarks.length === 0 && (
                  <p className="text-sm text-muted-foreground text-center py-4">
                    No remarks found
                  </p>
                )}
              </div>
            </ScrollArea>
            {(filters.remarks?.length ?? 0) > 0 && (
              <div className="p-2 border-t">
                <Button
                  variant="ghost"
                  size="sm"
                  className="w-full h-8"
                  onClick={() =>
                    onFiltersChange({ ...filters, remarks: undefined })
                  }
                >
                  Clear selection
                </Button>
              </div>
            )}
          </PopoverContent>
        </Popover>

        {visibleColumns && onVisibleColumnsChange && columnOptions.length > 0 && (
          <Popover>
            <PopoverTrigger asChild>
              <Button variant="outline" className={primaryTriggerClassName}>
                <span className="flex items-center gap-2">
                  <Columns3Icon className="size-4" />
                  <span>Columns</span>
                  {visibleColumnCount > 0 &&
                    visibleColumnCount !== columnOptions.length && (
                      <Badge variant="secondary" className="px-1.5 py-0 text-xs">
                        {visibleColumnCount}
                      </Badge>
                    )}
                </span>
                <ChevronsUpDownIcon className="size-4 opacity-50" />
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-[220px] p-2" align="start">
              <div className="space-y-1">
                {columnOptions.map((option) => {
                  const checked = visibleColumns[option.key] ?? false;
                  const disableToggle = checked && visibleColumnCount <= 1;
                  return (
                    <label
                      key={option.key}
                      className="flex items-center gap-2 px-2 py-1.5 rounded-sm hover:bg-muted cursor-pointer text-sm"
                    >
                      <Checkbox
                        checked={checked}
                        disabled={disableToggle}
                        onCheckedChange={() => handleColumnToggle(option.key)}
                      />
                      <span>{option.label}</span>
                    </label>
                  );
                })}
              </div>
            </PopoverContent>
          </Popover>
        )}

        {typeof pageSize === "number" &&
          pageSizeOptions.length > 0 &&
          onPageSizeChange && (
          <Select
            value={String(pageSize)}
            onValueChange={(value) => onPageSizeChange(Number(value))}
          >
            <SelectTrigger
              className={cn(
                primaryTriggerClassName,
                "w-[150px] border-primary text-primary"
              )}
            >
              <span className="flex items-center gap-2">
                <ListIcon className="size-4" />
                <SelectValue placeholder="Per Page" />
              </span>
            </SelectTrigger>
            <SelectContent>
              {pageSizeOptions.map((size) => (
                <SelectItem key={size} value={String(size)}>
                  {size} / page
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}

        {/* Filter Count & Clear */}
        {activeFilterCount > 0 && (
          <div className="flex items-center gap-2">
            <Badge variant="secondary" className="gap-1">
              {activeFilterCount} filter{activeFilterCount > 1 ? "s" : ""} active
            </Badge>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleClearFilters}
              className="h-9 gap-1 text-muted-foreground hover:text-foreground"
            >
              <XIcon className="size-4" />
              Clear all
            </Button>
          </div>
        )}
      </div>

      {/* Active Technology Badges */}
      {(filters.technologies?.length ?? 0) > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {filters.technologies?.map((tech) => (
            <Badge
              key={tech}
              variant="secondary"
              className="gap-1 pr-1 capitalize"
            >
              {tech}
              <button
                onClick={() => handleTechToggle(tech)}
                className="ml-1 rounded-full hover:bg-muted-foreground/20 p-0.5"
              >
                <XIcon className="size-3" />
              </button>
            </Badge>
          ))}
        </div>
      )}
    </div>
  );
}
