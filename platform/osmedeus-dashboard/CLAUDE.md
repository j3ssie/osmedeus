# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Osmedeus Dashboard - A UI dashboard for the Osmedeus Workflow Engine built with Next.js App Router, Tailwind CSS v4, and Shadcn UI (new-york style).

## Commands

```bash
bun install          # Install dependencies
bun dev              # Start dev server with Turbopack (http://localhost:3000)
bun run build        # Production build
bun run lint         # Run ESLint
bunx shadcn@latest add <component>  # Add Shadcn components
```

## Architecture

### Route Structure (App Router)
- `app/(auth)/` - Public routes (login page)
- `app/(dashboard)/` - Protected routes requiring authentication
  - `/` - Dashboard home with stats
  - `/scans`, `/scans/new` - Scan management
  - `/settings` - User settings
  - `/assets`, `/assets/workspaces/[id]` - Asset workspaces and HTTP assets
  - `/workflows/[id]` - Workflow editor

### Authentication
Mock auth via `providers/auth-provider.tsx` using localStorage (`osmedeus_session` key). Default credentials: any username with 4+ character password. The `AuthProvider` handles redirect logic between public/protected routes.

### API Layer
All API calls go through `lib/api/` which currently uses mock data from `lib/mock/data/`. To integrate real APIs:
1. Update `NEXT_PUBLIC_API_URL` env var
2. Replace mock implementations in `lib/api/*.ts`
3. The `PaginatedResponse<T>` type in `lib/types/api.ts` defines the pagination contract

### Workflow Editor
The workflow editor at `/workflows/[id]` visualizes YAML workflows using React Flow (@xyflow/react):
- `components/workflow-editor/utils/yaml-parser.ts` - Converts YAML to React Flow nodes/edges
- `components/workflow-editor/utils/layout-engine.ts` - Dagre-based auto-layout
- `components/workflow-editor/nodes/` - Custom node types: bash, parallel, function, foreach, start, end
- Workflow types defined in `lib/types/workflow.ts`

### Data Tables
Row-heavy tables run on AG Grid (`ag-grid-community` + `ag-grid-react`) through the single wrapper in `components/ui/data-grid.tsx`, which registers the community modules once, exports `DataGrid` / `GridPagination` / `GridRefreshOverlay` plus shared cell renderers, and defines `osmedeusGridTheme` — a `themeQuartz.withParams({...})` whose every value is a `var(--…)` from `app/globals.css`, so the grid follows light/dark without a remount.

On AG Grid: HTTP assets, workspaces, scans, vulnerabilities, events. Still on the plain `Table` primitive: registry, artifacts, schedules. Reach for `DataGrid` for anything that can grow past ~50 rows; the `Table` primitive is fine for short, bespoke layouts.

Notes when working on grids:
- AG Grid params must reference emitted variables (`var(--wash-primary)`), not Tailwind `@theme inline` aliases (`--color-primary-wash`), which are substituted into utilities rather than emitted.
- Pass `rowHeight` as a `DataGrid` prop, not through the spread, so the auto-height calculation matches the real row height.
- Assets and workspaces sort server-side (comparators return `0`, `onSortChanged` reports the clicked column); everything else sorts in the grid via per-column comparators.
- Truncation is a CSS problem, not a renderer one: AG Grid wraps every cell renderer in `.ag-cell-wrapper > .ag-cell-value`, both flex items at `min-width: auto`, so without the `min-width: 0` rules in `globals.css` a long URL sizes them to itself and paints across the next column. A cell renderer's own `truncate` also needs `block` — `overflow` does nothing on an inline box.
- Columns fill the grid width through `flex`, which `defaultColDef` sets to `1`: a column opts out with `flex: 0` plus an explicit `width` (status chips, action buttons), and everything else splits the remainder by its `flex` with `minWidth` as the floor. AG Grid 35 silently ignored `flex` — 36 resolves it — so a table that ends short of the card edge means the version regressed, not that the column defs are wrong.

### Component Organization
- `components/ui/` - Shadcn primitives (do not edit directly, use `bunx shadcn@latest add`)
- `components/layout/` - App shell (sidebar, topbar, mobile nav)
- `components/shared/` - Reusable components (EmptyState, ErrorState, LoadingSkeleton)
- Feature components in `components/scans/`, `components/assets/`, `components/dashboard/`

### Theming
Tailwind CSS v4 with CSS variables defined in `app/globals.css` — the single source of truth for the color palette (light `:root` and `.dark`). Theme switching via `next-themes` in `providers/theme-provider.tsx`.

Two layers of tokens:
1. `--og-*` — the raw design tokens (surfaces, text ladder, tone marks and their soft fills, chart ramps, radii). Light is a warm paper palette (cream planes, warm greys, an electric-blue primary); dark is warm greys on a near-black page with a mid-blue primary. They are not inverses of each other.
2. The Shadcn semantic names (`--background`, `--card`, `--primary`, …) derived from layer 1, so every Shadcn component keeps working and `providers/color-vars-provider.tsx` can still override the palette at runtime from a stored preset.

Prefer the Shadcn spelling in components. Reach for the `--og`-only utilities only where Shadcn has no equivalent: `bg-page/surface/raised/sunken`, `text-ink/body/faint`, `border-border-subtle/strong`, `bg-<tone>-soft` + `text-<tone>` chips, `*-wash` row tints, `viz-*`/`seq-*`, `shadow-glow`, `rounded-{frame,card,control,pill}`, `px-gutter`, `tracking-label`, `text-2xs`/`text-md`. Do not introduce raw Tailwind palette colours (`text-green-500`, `bg-red-100`, …) — the codebase has none.

## Key Files

| File | Purpose |
|------|---------|
| `app/globals.css` | Tailwind v4 config + CSS variables |
| `components.json` | Shadcn configuration |
| `lib/types/` | TypeScript types for domain models |
| `lib/api/client.ts` | API base config (mock delay, headers) |
| `lib/api/active-org.ts` | Active org selection + which endpoints are org-scoped |
| `providers/org-provider.tsx` | Org list/selection context behind the topbar switcher |
| `example-workflow.yaml` | Reference YAML workflow structure |

## Org Scope

Orgs group multiple workspaces under one tenant. The topbar `OrgSwitcher` picks an
active org; `lib/api/http.ts` then injects `?org=<uuid>` into every request to an
org-scoped endpoint (`/assets`, `/vulnerabilities`, `/runs`, `/workspaces` — the
list lives in `lib/api/active-org.ts`).

That means **list pages need no org code of their own**. Two consequences worth
knowing:

- To deliberately bypass the filter, pass an explicit `org` — `fetchWorkspacesList({ org: "" })`
  does this so the org management screen can list workspaces that belong to other
  orgs. The interceptor only fills in `org` when it is `undefined`.
- Changing the active org triggers a full page reload, because each page fetches
  in its own mount effect and there is no shared query cache to invalidate.

No active org means no `org` param at all, so the dashboard spans every org — the
same default the CLI and API use.
