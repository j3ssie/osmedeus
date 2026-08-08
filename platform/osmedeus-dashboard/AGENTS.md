# AGENTS.md

Agent-focused reference for the Osmedeus Dashboard codebase.

## Commands

```bash
bun install          # Install dependencies
bun dev              # Start dev server with Turbopack (http://localhost:3000)
bun run build        # Production build
bun run lint         # Run ESLint
```

## Architecture

**Tech Stack**: Next.js 16 (App Router), React 19, Tailwind CSS v4, TypeScript, Shadcn UI (new-york)

**Key Directories**:
- `app/(auth)/` & `app/(dashboard)/` - Next.js routes
- `components/` - React components (Shadcn in `ui/`, feature-specific subdirs, reusable in `shared/`)
- `lib/api/` - API calls (mock data from `lib/mock/`)
- `lib/types/` - TypeScript interfaces (api.ts, workflow.ts, scan.ts, asset.ts, event.ts, schedule.ts)
- `providers/` - Context providers (auth, theme)

**Core Features**:
- Auth: Mock via localStorage (`osmedeus_session`); any 4+ char password
- Workflow Editor: React Flow (@xyflow/react) + Dagre auto-layout at `components/workflow-editor/`
- API Layer: Abstracted in `lib/api/` with `PaginatedResponse<T>` contract

## Code Style

**Imports**: Use `@/*` path alias (e.g., `@/lib/utils`, `@/components/ui/button`). Order: React/Next → external libs → local imports.

**Components**: Functional with TypeScript, use `cn()` for conditional classes, `React.ComponentProps<"element">` for spread props pattern (see card.tsx).

**Types**: Strict mode enabled; centralize domain types in `lib/types/`; use Zod for validation (react-hook-form integration).

**Naming**: PascalCase for components/types, camelCase for functions/variables, SCREAMING_SNAKE_CASE for constants.

**Formatting**: ESLint (next/core-web-vitals + next/typescript), no Prettier—use ESLint auto-fix. Shadcn components use `data-slot` attributes.

**Error Handling**: Use `ErrorState` component in `components/shared/error-state.tsx`; leverage sonner for toast notifications.

**CSS**: Tailwind v4 with CSS variables in `app/globals.css`; theme switching via `next-themes`.

<!-- BEGIN:nextjs-agent-rules -->

# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->
