module.exports = [
"[externals]/next/dist/compiled/next-server/app-route-turbo.runtime.dev.js [external] (next/dist/compiled/next-server/app-route-turbo.runtime.dev.js, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("next/dist/compiled/next-server/app-route-turbo.runtime.dev.js", () => require("next/dist/compiled/next-server/app-route-turbo.runtime.dev.js"));

module.exports = mod;
}),
"[externals]/next/dist/compiled/@opentelemetry/api [external] (next/dist/compiled/@opentelemetry/api, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("next/dist/compiled/@opentelemetry/api", () => require("next/dist/compiled/@opentelemetry/api"));

module.exports = mod;
}),
"[externals]/next/dist/compiled/next-server/app-page-turbo.runtime.dev.js [external] (next/dist/compiled/next-server/app-page-turbo.runtime.dev.js, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("next/dist/compiled/next-server/app-page-turbo.runtime.dev.js", () => require("next/dist/compiled/next-server/app-page-turbo.runtime.dev.js"));

module.exports = mod;
}),
"[externals]/next/dist/server/app-render/work-unit-async-storage.external.js [external] (next/dist/server/app-render/work-unit-async-storage.external.js, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("next/dist/server/app-render/work-unit-async-storage.external.js", () => require("next/dist/server/app-render/work-unit-async-storage.external.js"));

module.exports = mod;
}),
"[externals]/next/dist/server/app-render/work-async-storage.external.js [external] (next/dist/server/app-render/work-async-storage.external.js, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("next/dist/server/app-render/work-async-storage.external.js", () => require("next/dist/server/app-render/work-async-storage.external.js"));

module.exports = mod;
}),
"[externals]/next/dist/shared/lib/no-fallback-error.external.js [external] (next/dist/shared/lib/no-fallback-error.external.js, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("next/dist/shared/lib/no-fallback-error.external.js", () => require("next/dist/shared/lib/no-fallback-error.external.js"));

module.exports = mod;
}),
"[externals]/next/dist/server/app-render/after-task-async-storage.external.js [external] (next/dist/server/app-render/after-task-async-storage.external.js, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("next/dist/server/app-render/after-task-async-storage.external.js", () => require("next/dist/server/app-render/after-task-async-storage.external.js"));

module.exports = mod;
}),
"[project]/lib/mock/data/workspaces.ts [app-route] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "mockWorkspaces",
    ()=>mockWorkspaces
]);
const mockWorkspaces = [
    {
        id: 1,
        name: "example.com",
        data_source: "local",
        local_path: "/home/user/osmedeus-base/workspaces/example.com",
        state_execution_log: "/home/user/osmedeus-base/workspaces/example.com/log/execution.log",
        state_completed_file: "/home/user/osmedeus-base/workspaces/example.com/state/completed",
        state_workflow_file: "/home/user/osmedeus-base/workspaces/example.com/state/workflow.yaml",
        state_workflow_folder: "/home/user/osmedeus-base/workspaces/example.com/state",
        total_assets: 150,
        total_subdomains: 1247,
        total_urls: 856,
        total_vulns: 23,
        vuln_critical: 2,
        vuln_high: 5,
        vuln_medium: 8,
        vuln_low: 8,
        vuln_potential: 3,
        risk_score: 7.5,
        tags: [
            "production",
            "priority"
        ],
        last_run: new Date(Date.now() - 3600000).toISOString(),
        run_workflow: "subdomain-enum",
        created_at: "2024-01-15T08:00:00Z",
        updated_at: new Date(Date.now() - 3600000).toISOString()
    },
    {
        id: 2,
        name: "testsite.org",
        data_source: "cloud",
        local_path: "/home/user/osmedeus-base/workspaces/testsite.org",
        state_execution_log: "/home/user/osmedeus-base/workspaces/testsite.org/log/execution.log",
        state_completed_file: "/home/user/osmedeus-base/workspaces/testsite.org/state/completed",
        state_workflow_file: "/home/user/osmedeus-base/workspaces/testsite.org/state/workflow.yaml",
        state_workflow_folder: "/home/user/osmedeus-base/workspaces/testsite.org/state",
        total_assets: 50,
        total_subdomains: 342,
        total_urls: 189,
        total_vulns: 7,
        vuln_critical: 0,
        vuln_high: 1,
        vuln_medium: 3,
        vuln_low: 3,
        vuln_potential: 2,
        risk_score: 4.2,
        tags: [
            "staging"
        ],
        last_run: new Date(Date.now() - 86400000).toISOString(),
        run_workflow: "port-scan",
        created_at: "2024-02-20T12:00:00Z",
        updated_at: new Date(Date.now() - 86400000).toISOString()
    },
    {
        id: 3,
        name: "acme.io",
        data_source: "imported",
        local_path: "/home/user/osmedeus-base/workspaces/acme.io",
        state_execution_log: "/home/user/osmedeus-base/workspaces/acme.io/log/execution.log",
        state_completed_file: "/home/user/osmedeus-base/workspaces/acme.io/state/completed",
        state_workflow_file: "/home/user/osmedeus-base/workspaces/acme.io/state/workflow.yaml",
        state_workflow_folder: "/home/user/osmedeus-base/workspaces/acme.io/state",
        total_assets: 320,
        total_subdomains: 2156,
        total_urls: 1432,
        total_vulns: 45,
        vuln_critical: 5,
        vuln_high: 12,
        vuln_medium: 15,
        vuln_low: 13,
        vuln_potential: 8,
        risk_score: 8.8,
        tags: [
            "production",
            "critical"
        ],
        last_run: new Date(Date.now() - 172800000).toISOString(),
        run_workflow: "full-scan",
        created_at: "2024-03-10T10:00:00Z",
        updated_at: new Date(Date.now() - 172800000).toISOString()
    },
    {
        id: 4,
        name: "secure.bank.com",
        data_source: "local",
        local_path: "/home/user/osmedeus-base/workspaces/secure.bank.com",
        state_execution_log: "/home/user/osmedeus-base/workspaces/secure.bank.com/log/execution.log",
        state_completed_file: "/home/user/osmedeus-base/workspaces/secure.bank.com/state/completed",
        state_workflow_file: "/home/user/osmedeus-base/workspaces/secure.bank.com/state/workflow.yaml",
        state_workflow_folder: "/home/user/osmedeus-base/workspaces/secure.bank.com/state",
        total_assets: 80,
        total_subdomains: 567,
        total_urls: 312,
        total_vulns: 12,
        vuln_critical: 1,
        vuln_high: 2,
        vuln_medium: 5,
        vuln_low: 4,
        vuln_potential: 0,
        risk_score: 5.5,
        tags: [
            "finance",
            "priority"
        ],
        last_run: new Date(Date.now() - 43200000).toISOString(),
        run_workflow: "vuln-scan",
        created_at: "2024-04-05T09:00:00Z",
        updated_at: new Date(Date.now() - 43200000).toISOString()
    },
    {
        id: 5,
        name: "startup.dev",
        data_source: "cloud",
        local_path: "/home/user/osmedeus-base/workspaces/startup.dev",
        state_execution_log: "/home/user/osmedeus-base/workspaces/startup.dev/log/execution.log",
        state_completed_file: "/home/user/osmedeus-base/workspaces/startup.dev/state/completed",
        state_workflow_file: "/home/user/osmedeus-base/workspaces/startup.dev/state/workflow.yaml",
        state_workflow_folder: "/home/user/osmedeus-base/workspaces/startup.dev/state",
        total_assets: 20,
        total_subdomains: 89,
        total_urls: 45,
        total_vulns: 3,
        vuln_critical: 0,
        vuln_high: 0,
        vuln_medium: 1,
        vuln_low: 2,
        vuln_potential: 5,
        risk_score: 2.1,
        tags: [
            "development"
        ],
        last_run: new Date(Date.now() - 604800000).toISOString(),
        run_workflow: "quick-scan",
        created_at: "2024-05-01T14:00:00Z",
        updated_at: new Date(Date.now() - 604800000).toISOString()
    },
    {
        id: 6,
        name: "megacorp.com",
        data_source: "local",
        local_path: "/home/user/osmedeus-base/workspaces/megacorp.com",
        state_execution_log: "/home/user/osmedeus-base/workspaces/megacorp.com/log/execution.log",
        state_completed_file: "/home/user/osmedeus-base/workspaces/megacorp.com/state/completed",
        state_workflow_file: "/home/user/osmedeus-base/workspaces/megacorp.com/state/workflow.yaml",
        state_workflow_folder: "/home/user/osmedeus-base/workspaces/megacorp.com/state",
        total_assets: 890,
        total_subdomains: 8934,
        total_urls: 4521,
        total_vulns: 89,
        vuln_critical: 8,
        vuln_high: 22,
        vuln_medium: 35,
        vuln_low: 24,
        vuln_potential: 15,
        risk_score: 9.2,
        tags: [
            "enterprise",
            "production",
            "critical"
        ],
        last_run: new Date(Date.now() - 259200000).toISOString(),
        run_workflow: "full-scan",
        created_at: "2024-01-01T08:00:00Z",
        updated_at: new Date(Date.now() - 259200000).toISOString()
    },
    {
        id: 7,
        name: "shop.retail.com",
        data_source: "imported",
        local_path: "/home/user/osmedeus-base/workspaces/shop.retail.com",
        state_execution_log: "/home/user/osmedeus-base/workspaces/shop.retail.com/log/execution.log",
        state_completed_file: "/home/user/osmedeus-base/workspaces/shop.retail.com/state/completed",
        state_workflow_file: "/home/user/osmedeus-base/workspaces/shop.retail.com/state/workflow.yaml",
        state_workflow_folder: "/home/user/osmedeus-base/workspaces/shop.retail.com/state",
        total_assets: 45,
        total_subdomains: 234,
        total_urls: 167,
        total_vulns: 5,
        vuln_critical: 0,
        vuln_high: 1,
        vuln_medium: 2,
        vuln_low: 2,
        vuln_potential: 3,
        risk_score: 3.5,
        tags: [
            "retail",
            "staging"
        ],
        last_run: new Date(Date.now() - 600000).toISOString(),
        run_workflow: "subdomain-enum",
        created_at: "2024-06-15T11:00:00Z",
        updated_at: new Date(Date.now() - 600000).toISOString()
    }
];
}),
"[project]/app/api/mock/api/workspaces/route.ts [app-route] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "GET",
    ()=>GET,
    "dynamic",
    ()=>dynamic
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$server$2e$js__$5b$app$2d$route$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/server.js [app-route] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$mock$2f$data$2f$workspaces$2e$ts__$5b$app$2d$route$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/mock/data/workspaces.ts [app-route] (ecmascript)");
;
;
const dynamic = "force-static";
async function GET() {
    const offset = 0;
    const limit = 20;
    // Mock data already has the correct structure
    const items = __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$mock$2f$data$2f$workspaces$2e$ts__$5b$app$2d$route$5d$__$28$ecmascript$29$__["mockWorkspaces"].map((w)=>({
            id: w.id,
            name: w.name,
            data_source: w.data_source,
            local_path: w.local_path,
            state_execution_log: w.state_execution_log,
            state_completed_file: w.state_completed_file,
            state_workflow_file: w.state_workflow_file,
            state_workflow_folder: w.state_workflow_folder,
            total_assets: w.total_assets,
            total_subdomains: w.total_subdomains,
            total_urls: w.total_urls,
            total_vulns: w.total_vulns,
            vuln_critical: w.vuln_critical,
            vuln_high: w.vuln_high,
            vuln_medium: w.vuln_medium,
            vuln_low: w.vuln_low,
            vuln_potential: w.vuln_potential,
            risk_score: w.risk_score,
            tags: w.tags,
            last_run: w.last_run,
            run_workflow: w.run_workflow,
            created_at: w.created_at,
            updated_at: w.updated_at
        }));
    const filtered = items;
    const sliced = filtered.slice(offset, offset + limit);
    const resp = {
        data: sliced,
        pagination: {
            total: filtered.length,
            offset,
            limit
        },
        mode: "database"
    };
    return __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$server$2e$js__$5b$app$2d$route$5d$__$28$ecmascript$29$__["NextResponse"].json(resp);
}
}),
];

//# sourceMappingURL=%5Broot-of-the-server%5D__8328cd25._.js.map