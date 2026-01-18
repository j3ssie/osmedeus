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
"[externals]/fs [external] (fs, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("fs", () => require("fs"));

module.exports = mod;
}),
"[externals]/path [external] (path, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("path", () => require("path"));

module.exports = mod;
}),
"[project]/app/api/mock/api/workflows/route.ts [app-route] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "GET",
    ()=>GET,
    "dynamic",
    ()=>dynamic
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$server$2e$js__$5b$app$2d$route$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/server.js [app-route] (ecmascript)");
var __TURBOPACK__imported__module__$5b$externals$5d2f$fs__$5b$external$5d$__$28$fs$2c$__cjs$29$__ = __turbopack_context__.i("[externals]/fs [external] (fs, cjs)");
var __TURBOPACK__imported__module__$5b$externals$5d2f$path__$5b$external$5d$__$28$path$2c$__cjs$29$__ = __turbopack_context__.i("[externals]/path [external] (path, cjs)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$js$2d$yaml$2f$dist$2f$js$2d$yaml$2e$mjs__$5b$app$2d$route$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/js-yaml/dist/js-yaml.mjs [app-route] (ecmascript)");
;
;
;
;
const dynamic = "force-static";
function normalizeTags(raw) {
    if (Array.isArray(raw)) {
        return raw.filter((t)=>typeof t === "string").map((t)=>t.trim()).filter(Boolean);
    }
    if (typeof raw === "string") {
        return raw.split(",").map((t)=>t.trim()).filter(Boolean);
    }
    return [];
}
function listYamlFiles(rootDir, relDir = "") {
    const fullDir = __TURBOPACK__imported__module__$5b$externals$5d2f$path__$5b$external$5d$__$28$path$2c$__cjs$29$__["default"].join(rootDir, relDir);
    const entries = __TURBOPACK__imported__module__$5b$externals$5d2f$fs__$5b$external$5d$__$28$fs$2c$__cjs$29$__["default"].readdirSync(fullDir, {
        withFileTypes: true
    });
    const out = [];
    for (const e of entries){
        if (e.isDirectory()) {
            out.push(...listYamlFiles(rootDir, __TURBOPACK__imported__module__$5b$externals$5d2f$path__$5b$external$5d$__$28$path$2c$__cjs$29$__["default"].join(relDir, e.name)));
            continue;
        }
        if (e.isFile() && e.name.endsWith(".yaml")) {
            out.push(__TURBOPACK__imported__module__$5b$externals$5d2f$path__$5b$external$5d$__$28$path$2c$__cjs$29$__["default"].join(relDir, e.name));
        }
    }
    return out;
}
async function GET() {
    const offset = 0;
    const limit = 50;
    const dir = __TURBOPACK__imported__module__$5b$externals$5d2f$path__$5b$external$5d$__$28$path$2c$__cjs$29$__["default"].join(process.cwd(), "mock-workflows");
    let relFiles = [];
    try {
        relFiles = listYamlFiles(dir);
    } catch  {
        relFiles = [];
    }
    const items = relFiles.map((rel)=>{
        const full = __TURBOPACK__imported__module__$5b$externals$5d2f$path__$5b$external$5d$__$28$path$2c$__cjs$29$__["default"].join(dir, rel);
        let doc = {};
        try {
            const content = __TURBOPACK__imported__module__$5b$externals$5d2f$fs__$5b$external$5d$__$28$fs$2c$__cjs$29$__["default"].readFileSync(full, "utf-8");
            doc = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$js$2d$yaml$2f$dist$2f$js$2d$yaml$2e$mjs__$5b$app$2d$route$5d$__$28$ecmascript$29$__["default"].load(content) || {};
        } catch  {
            doc = {};
        }
        const fallbackId = __TURBOPACK__imported__module__$5b$externals$5d2f$path__$5b$external$5d$__$28$path$2c$__cjs$29$__["default"].basename(rel, ".yaml");
        const name = typeof doc?.name === "string" && doc.name.trim() ? doc.name.trim() : fallbackId;
        const wfKind = doc?.kind === "flow" ? "flow" : "module";
        const description = doc?.description || `Mock workflow from ${rel}`;
        const steps = Array.isArray(doc?.steps) ? doc.steps : [];
        const modules = Array.isArray(doc?.modules) ? doc.modules : [];
        const rawTags = normalizeTags(doc?.tags);
        const tagSet = new Set(rawTags);
        tagSet.add("mock-data");
        const tags = Array.from(tagSet);
        const params = Array.isArray(doc?.params) ? doc.params : [];
        const required_params = params.filter((p)=>p?.required).map((p)=>p?.name ?? "");
        return {
            name,
            kind: wfKind,
            description,
            tags,
            file_path: `/mock-workflows/${rel.replace(/\\/g, "/")}`,
            params,
            required_params,
            step_count: steps.length,
            module_count: modules.length,
            checksum: "",
            indexed_at: new Date().toISOString()
        };
    });
    const uniqueByName = new Map();
    items.forEach((wf)=>{
        const key = String(wf.name || "").trim();
        if (!key) return;
        if (!uniqueByName.has(key)) uniqueByName.set(key, wf);
    });
    const uniqueItems = Array.from(uniqueByName.values());
    const sliced = uniqueItems.slice(Math.max(0, offset), Math.max(0, offset) + Math.max(0, limit));
    return __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$server$2e$js__$5b$app$2d$route$5d$__$28$ecmascript$29$__["NextResponse"].json({
        data: sliced,
        pagination: {
            total: uniqueItems.length,
            offset,
            limit
        }
    });
}
}),
];

//# sourceMappingURL=%5Broot-of-the-server%5D__227952d5._.js.map