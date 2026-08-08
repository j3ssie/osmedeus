## Goal
Build a polished, production-grade UI Dashboard for the **Osmedeus Workflow Engine** that helps users **run, manage, and monitor security scans**. Prioritize a **clean, modern, sleek aesthetic** with excellent UX states (loading/empty/error), strong visual hierarchy, and consistent spacing/typography.

## Hard Requirements (must follow)
- Runtime: **Bun 1.3.5**
- Framework: **Next.js 16.1.1 (App Router)** + **TypeScript 5.9.3**
- Styling: **Tailwind CSS 4.1.18** + **Shadcn 3.6.2**
  - Follow local Shadcn rules: prefer Shadcn components/classes, minimal custom CSS, no deprecated Tailwind v3 config patterns.
- Theme: Use this Shadcn preset  
  `bunx shadcn@latest add https://tweakcn.com/r/themes/cmjy4to6r000104jv1vg6feir`  
  or match the palette in `color-scheme.css`.
- Icons: **lucide-react 0.562.0**
- Workflow editor: **reactflow 12.10.0**

## UI / Design Direction (modern & sleek)
- Use a **calm, dense-but-readable dashboard feel**: generous padding, crisp borders, soft shadows, subtle separators.
- Prefer **Card-based composition**, muted backgrounds, and consistent radii.
- Strong typography hierarchy: clear page titles, section headers, quiet helper text.
- Subtle interactions: hover/focus states, dropdown/menu polish, inline validation messages.
- Accessibility: keyboard navigation, focus rings, aria labels where appropriate, readable contrast in both themes.

## UX Requirements
- Every screen must include:
  - Loading state (skeletons where appropriate)
  - Empty state (helpful copy + primary action)
  - Error state (human-readable message + retry where relevant)
- Forms:
  - Basic validation (client-side), disabled submit while loading
  - Clear error messages per field
  - Toast/alert feedback for actions
- Navigation:
  - Responsive layout: desktop sidebar + topbar; mobile uses drawer
  - Active nav state, breadcrumbs where helpful
- Theme:
  - Implement **light/dark toggle** using Shadcn theme-controller
  - Persist theme preference

## Architecture Requirements
- Keep a clean separation between:
  - **UI components** (pages, layout, reusable components)
  - **Mock data + placeholder API functions** (so real APIs can replace later)
- Reuse composable primitives:
  - Stat cards, table wrapper, form field components, dialogs/modals, empty states

## Routes / Pages to Build

### 1) Home (`/`)
- Summary stat cards:
  - workflows, workspaces, subdomains, HTTP assets, vulnerabilities
- “Recent scans” preview list/table (compact, clickable rows)

### 2) Login (`/login`)
- Responsive login form (username + password)
- Sleek auth layout (centered card, subtle background)
- Mock auth: on success, navigate to Home
- Structure code so real auth can be plugged in later

### 3) Settings (`/settings`)
- Settings sections (carded):
  - Profile/Org (basic fields)
  - API endpoint
  - Theme toggle (if not global)
- Save to local state + localStorage

### 4) Scans
- Listing (`/scans`):
  - Table columns: id, workflow, target/workspace, status, startedAt, duration, actions
  - Status badge styles, row actions menu, empty/loading states
- New scan (`/scans/new`):
  - Select workflow
  - Select workspace/target
  - Optional scheduling: toggle + cron input (light validation + helper text)
  - On submit: toast/alert and add to mock list

### 5) Workflow Editor (`/workflows/[id]` OR `/workflow-editor`)
- Use React Flow to visualize/edit an Osmedeus YAML workflow (see `example-workflow.yaml`)
- Features:
  - Load YAML text (mock string/file), parse into nodes/edges
  - Drag/drop, basic node editing (name/type/params)
  - Read-only “YAML Preview” panel
  - “Save” action (mock)
- Layout:
  - Responsive split-pane: canvas + side panel, styled to match Shadcn

### 6) Assets (`/assets`)
- Workspaces list (cards or table)
- Workspace detail (`/assets/workspaces/[id]`):
  - Large HTTP assets table columns:
    - url, status_code, content_length, title, location
  - Filters:
    - text search (url/title)
    - status_code filter
    - content_length range (optional)
    - location filter
  - Pagination UI (mock), loading/empty/error states

## Deliverables
- Recommended folder/file structure for Next.js App Router
- Implement layout + all routes above
- Shadcn-first UI (minimal custom CSS)
- Mock data modules + placeholder API functions (clearly separated)
- Brief run instructions (bun) and clear “where to replace mocks with real APIs”

## Acceptance Criteria
- UI feels cohesive and “premium”: consistent spacing, typography, and component usage
- Fully responsive: works on mobile/tablet/desktop
- Theme toggle works + persists
- Every page has loading/empty/error states
- Code is clean, composable, and easy to extend with real APIs
