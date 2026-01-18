module.exports = [
"[project]/components/workflow-editor/canvas-settings.tsx [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "CanvasSettingsProvider",
    ()=>CanvasSettingsProvider,
    "useCanvasSettings",
    ()=>useCanvasSettings
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react-jsx-dev-runtime.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react.js [app-ssr] (ecmascript)");
"use client";
;
;
const CanvasSettingsContext = /*#__PURE__*/ __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["createContext"](null);
function CanvasSettingsProvider({ wrapLongText, showDetails, children }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(CanvasSettingsContext.Provider, {
        value: {
            wrapLongText,
            showDetails
        },
        children: children
    }, void 0, false, {
        fileName: "[project]/components/workflow-editor/canvas-settings.tsx",
        lineNumber: 22,
        columnNumber: 5
    }, this);
}
function useCanvasSettings() {
    return __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useContext"](CanvasSettingsContext) ?? {
        wrapLongText: false,
        showDetails: true
    };
}
}),
"[project]/components/workflow-editor/nodes/base-node.tsx [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "BaseNode",
    ()=>BaseNode
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react-jsx-dev-runtime.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$react$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$locals$3e$__ = __turbopack_context__.i("[project]/node_modules/@xyflow/react/dist/esm/index.js [app-ssr] (ecmascript) <locals>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$system$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/@xyflow/system/dist/esm/index.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/utils.ts [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$canvas$2d$settings$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/workflow-editor/canvas-settings.tsx [app-ssr] (ecmascript)");
"use client";
;
;
;
;
;
function truncateText(input, maxLen) {
    const text = input.trim();
    if (text.length <= maxLen) return text;
    return `${text.slice(0, Math.max(0, maxLen - 1)).trimEnd()}…`;
}
function normalizeInlineText(input) {
    return input.replace(/\s+/g, " ").trim();
}
function isSensitiveKey(key) {
    const k = key.toLowerCase();
    return k.includes("password") || k.includes("passwd") || k.includes("secret") || k.includes("token") || k.includes("apikey") || k.includes("api_key") || k.includes("authorization");
}
function stringifyInlineValue(value) {
    if (value == null) return "null";
    if (typeof value === "string") return normalizeInlineText(value);
    if (typeof value === "number" || typeof value === "boolean" || typeof value === "bigint") {
        return String(value);
    }
    try {
        return normalizeInlineText(JSON.stringify(value));
    } catch  {
        return "";
    }
}
function extractMessageContentText(content) {
    if (typeof content === "string") return content;
    if (Array.isArray(content)) {
        const parts = [];
        for (const item of content){
            if (typeof item === "string") {
                parts.push(item);
                continue;
            }
            if (!item || typeof item !== "object") continue;
            const anyItem = item;
            if (anyItem.type === "text" && typeof anyItem.text === "string") {
                parts.push(anyItem.text);
                continue;
            }
            if (typeof anyItem.content === "string") {
                parts.push(anyItem.content);
                continue;
            }
            if (typeof anyItem.input === "string") {
                parts.push(anyItem.input);
                continue;
            }
        }
        if (parts.length > 0) return parts.join(" ");
        try {
            return JSON.stringify(content);
        } catch  {
            return "";
        }
    }
    if (content && typeof content === "object") {
        const anyObj = content;
        if (typeof anyObj.text === "string") return anyObj.text;
        if (typeof anyObj.content === "string") return anyObj.content;
        try {
            return JSON.stringify(content);
        } catch  {
            return "";
        }
    }
    return content == null ? "" : String(content);
}
function buildLlmMessagesSummary(step) {
    if (!step) return [];
    if (step.is_embedding) {
        const input = Array.isArray(step.embedding_input) ? step.embedding_input : [];
        if (input.length === 0) return [];
        const first = normalizeInlineText(String(input[0] ?? ""));
        const suffix = input.length > 1 ? ` (+${input.length - 1})` : "";
        return [
            truncateText(`input: ${first}${suffix}`, 140)
        ];
    }
    const messages = Array.isArray(step.messages) ? step.messages : [];
    if (messages.length === 0) return [];
    const maxLines = 3;
    const lines = [];
    for (const msg of messages.slice(0, maxLines)){
        if (!msg || typeof msg !== "object") continue;
        const anyMsg = msg;
        const role = typeof anyMsg.role === "string" ? anyMsg.role : "message";
        const name = typeof anyMsg.name === "string" ? anyMsg.name : "";
        const content = normalizeInlineText(extractMessageContentText(anyMsg.content));
        const roleLabel = name ? `${role}[${name}]` : role;
        const value = content ? truncateText(content, 140) : "";
        lines.push(value ? `${roleLabel}: ${value}` : `${roleLabel}`);
    }
    const remaining = messages.length - maxLines;
    if (remaining > 0) lines.push(`+${remaining} more`);
    return lines;
}
function buildFunctionMessagesSummary(step) {
    if (!step) return [];
    if (step.type !== "function") return [];
    const items = [];
    if (typeof step.function === "string" && step.function.trim()) {
        items.push(`fn: ${normalizeInlineText(step.function)}`);
    }
    const functions = Array.isArray(step.functions) ? step.functions : [];
    for (const fn of functions){
        if (typeof fn !== "string") continue;
        const text = fn.trim();
        if (!text) continue;
        items.push(`fn: ${normalizeInlineText(text)}`);
    }
    const parallelFunctions = Array.isArray(step.parallel_functions) ? step.parallel_functions : [];
    for (const fn of parallelFunctions){
        if (typeof fn !== "string") continue;
        const text = fn.trim();
        if (!text) continue;
        items.push(`pfn: ${normalizeInlineText(text)}`);
    }
    if (items.length === 0) return [];
    const maxLines = 3;
    const lines = items.slice(0, maxLines).map((v)=>truncateText(v, 140));
    const remaining = items.length - maxLines;
    if (remaining > 0) lines.push(`+${remaining} more`);
    return lines;
}
function buildHttpSummary(step) {
    if (!step) return [];
    if (step.type !== "http") return [];
    const lines = [];
    const url = typeof step.url === "string" ? step.url.trim() : "";
    const method = typeof step.method === "string" ? step.method.trim() : "";
    if (url) {
        const methodLabel = method ? method.toUpperCase() : "HTTP";
        lines.push(truncateText(`${methodLabel} ${normalizeInlineText(url)}`, 140));
    }
    if (step.headers && typeof step.headers === "object" && !Array.isArray(step.headers)) {
        const keys = Object.keys(step.headers).filter(Boolean);
        if (keys.length > 0) {
            const shown = keys.slice(0, 3).join(", ");
            const suffix = keys.length > 3 ? ` (+${keys.length - 3})` : "";
            lines.push(truncateText(`headers: ${shown}${suffix}`, 140));
        }
    }
    const body = typeof step.request_body === "string" ? step.request_body.trim() : "";
    if (body) {
        lines.push(truncateText(`body: ${normalizeInlineText(body)}`, 140));
    }
    return lines.slice(0, 3);
}
function buildModuleSummary(module) {
    if (!module) return [];
    const lines = [];
    const path = typeof module.path === "string" ? module.path.trim() : "";
    if (path) lines.push(truncateText(`path: ${normalizeInlineText(path)}`, 140));
    const dependsOn = Array.isArray(module.depends_on) ? module.depends_on : [];
    if (dependsOn.length > 0) {
        const shown = dependsOn.slice(0, 3).join(", ");
        const suffix = dependsOn.length > 3 ? ` (+${dependsOn.length - 3})` : "";
        lines.push(truncateText(`depends: ${normalizeInlineText(shown)}${suffix}`, 140));
    }
    const params = module.params && typeof module.params === "object" && !Array.isArray(module.params) ? module.params : null;
    if (params) {
        const entries = Object.entries(params).filter(([k])=>Boolean(k));
        if (entries.length > 0) {
            const shown = entries.slice(0, 3).map(([k, v])=>{
                if (isSensitiveKey(k)) return `${k}=***`;
                const valueText = stringifyInlineValue(v);
                return valueText ? `${k}=${valueText}` : k;
            });
            const suffix = entries.length > 3 ? ` (+${entries.length - 3})` : "";
            lines.push(truncateText(`params: ${shown.join(", ")}${suffix}`, 140));
        }
    }
    return lines.slice(0, 3);
}
function buildCommandSummary(step) {
    if (!step) return "";
    const hasStructuredArgs = step.speed_args !== undefined || step.config_args !== undefined || step.input_args !== undefined || step.output_args !== undefined;
    if (typeof step.command === "string" && step.command.trim()) {
        if (hasStructuredArgs) {
            const parts = [
                step.command,
                step.speed_args,
                step.config_args,
                step.input_args,
                step.output_args
            ].map((p)=>typeof p === "string" ? p.trim() : "").filter(Boolean);
            return parts.join(" ");
        }
        return step.command.trim();
    }
    if (Array.isArray(step.commands) && step.commands.length > 0) {
        const first = String(step.commands[0] || "").trim();
        if (!first) return "";
        if (step.commands.length === 1) return first;
        return `${first} (+${step.commands.length - 1})`;
    }
    if (Array.isArray(step.parallel_commands) && step.parallel_commands.length > 0) {
        const first = String(step.parallel_commands[0] || "").trim();
        if (!first) return "";
        if (step.parallel_commands.length === 1) return first;
        return `${first} (+${step.parallel_commands.length - 1})`;
    }
    return "";
}
function hasStructuredArgs(step) {
    if (!step) return false;
    return step.speed_args !== undefined || step.config_args !== undefined || step.input_args !== undefined || step.output_args !== undefined;
}
function BaseNode({ data, selected, icon, color, showHandles = true }) {
    const { wrapLongText, showDetails } = (0, __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$canvas$2d$settings$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useCanvasSettings"])();
    const commandSummary = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useMemo"](()=>buildCommandSummary(data.step), [
        data.step
    ]);
    const structured = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useMemo"](()=>hasStructuredArgs(data.step), [
        data.step
    ]);
    const functionMessageSummary = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useMemo"](()=>buildFunctionMessagesSummary(data.step), [
        data.step
    ]);
    const httpSummary = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useMemo"](()=>buildHttpSummary(data.step), [
        data.step
    ]);
    const moduleSummary = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useMemo"](()=>buildModuleSummary(data.module), [
        data.module
    ]);
    const llmMessageSummary = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useMemo"](()=>{
        if (data.step?.type !== "llm") return [];
        return buildLlmMessagesSummary(data.step);
    }, [
        data.step
    ]);
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Fragment"], {
        children: [
            showHandles && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$react$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$locals$3e$__["Handle"], {
                type: "target",
                position: __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$system$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Position"].Top,
                className: "!bg-border !border-2 !border-background !size-3"
            }, void 0, false, {
                fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                lineNumber: 309,
                columnNumber: 9
            }, this),
            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                className: (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["cn"])("flex items-center gap-3 rounded-lg border bg-card px-4 py-3 shadow-sm transition-all min-w-[200px]", selected && "ring-2 ring-ring ring-offset-2 ring-offset-background"),
                children: [
                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                        className: (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["cn"])("flex size-8 items-center justify-center rounded-md", color),
                        children: icon
                    }, void 0, false, {
                        fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                        lineNumber: 322,
                        columnNumber: 9
                    }, this),
                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                        className: "flex-1 min-w-0",
                        children: [
                            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("p", {
                                className: "text-sm font-medium truncate",
                                children: data.label
                            }, void 0, false, {
                                fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                                lineNumber: 331,
                                columnNumber: 11
                            }, this),
                            data.step?.type ? /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                className: "flex items-center gap-2",
                                children: [
                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("p", {
                                        className: "text-xs text-muted-foreground capitalize",
                                        children: data.step.type
                                    }, void 0, false, {
                                        fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                                        lineNumber: 334,
                                        columnNumber: 15
                                    }, this),
                                    showDetails && structured && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("span", {
                                        className: "rounded border px-1.5 py-0.5 text-[10px] text-muted-foreground",
                                        children: "args"
                                    }, void 0, false, {
                                        fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                                        lineNumber: 338,
                                        columnNumber: 17
                                    }, this)
                                ]
                            }, void 0, true, {
                                fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                                lineNumber: 333,
                                columnNumber: 13
                            }, this) : data.module ? /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("p", {
                                className: "text-xs text-muted-foreground",
                                children: "module"
                            }, void 0, false, {
                                fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                                lineNumber: 344,
                                columnNumber: 13
                            }, this) : null,
                            showDetails && commandSummary && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("p", {
                                className: (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["cn"])("mt-1 text-[11px] text-muted-foreground font-mono", wrapLongText ? "whitespace-pre-wrap break-words" : "truncate"),
                                children: commandSummary
                            }, void 0, false, {
                                fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                                lineNumber: 347,
                                columnNumber: 13
                            }, this),
                            showDetails && data.step?.type === "function" && functionMessageSummary.length > 0 && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                className: (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["cn"])("mt-1 space-y-0.5 text-[11px] text-muted-foreground font-mono", wrapLongText ? "whitespace-pre-wrap break-words" : ""),
                                children: functionMessageSummary.map((line, idx)=>/*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                        className: wrapLongText ? "" : "truncate",
                                        children: line
                                    }, `${data.label}-fn-msg-${idx}`, false, {
                                        fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                                        lineNumber: 365,
                                        columnNumber: 17
                                    }, this))
                            }, void 0, false, {
                                fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                                lineNumber: 358,
                                columnNumber: 13
                            }, this),
                            showDetails && data.step?.type === "http" && httpSummary.length > 0 && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                className: (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["cn"])("mt-1 space-y-0.5 text-[11px] text-muted-foreground font-mono", wrapLongText ? "whitespace-pre-wrap break-words" : ""),
                                children: httpSummary.map((line, idx)=>/*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                        className: wrapLongText ? "" : "truncate",
                                        children: line
                                    }, `${data.label}-http-${idx}`, false, {
                                        fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                                        lineNumber: 380,
                                        columnNumber: 17
                                    }, this))
                            }, void 0, false, {
                                fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                                lineNumber: 373,
                                columnNumber: 13
                            }, this),
                            showDetails && data.module && moduleSummary.length > 0 && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                className: (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["cn"])("mt-1 space-y-0.5 text-[11px] text-muted-foreground font-mono", wrapLongText ? "whitespace-pre-wrap break-words" : ""),
                                children: moduleSummary.map((line, idx)=>/*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                        className: wrapLongText ? "" : "truncate",
                                        children: line
                                    }, `${data.label}-module-${idx}`, false, {
                                        fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                                        lineNumber: 395,
                                        columnNumber: 17
                                    }, this))
                            }, void 0, false, {
                                fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                                lineNumber: 388,
                                columnNumber: 13
                            }, this),
                            showDetails && data.step?.type === "llm" && llmMessageSummary.length > 0 && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                className: (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["cn"])("mt-1 space-y-0.5 text-[11px] text-muted-foreground font-mono", wrapLongText ? "whitespace-pre-wrap break-words" : ""),
                                children: llmMessageSummary.map((line, idx)=>/*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                        className: wrapLongText ? "" : "truncate",
                                        children: line
                                    }, `${data.label}-llm-msg-${idx}`, false, {
                                        fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                                        lineNumber: 410,
                                        columnNumber: 17
                                    }, this))
                            }, void 0, false, {
                                fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                                lineNumber: 403,
                                columnNumber: 13
                            }, this)
                        ]
                    }, void 0, true, {
                        fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                        lineNumber: 330,
                        columnNumber: 9
                    }, this)
                ]
            }, void 0, true, {
                fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                lineNumber: 316,
                columnNumber: 7
            }, this),
            showHandles && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$react$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$locals$3e$__["Handle"], {
                type: "source",
                position: __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$system$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Position"].Bottom,
                className: "!bg-border !border-2 !border-background !size-3"
            }, void 0, false, {
                fileName: "[project]/components/workflow-editor/nodes/base-node.tsx",
                lineNumber: 420,
                columnNumber: 9
            }, this)
        ]
    }, void 0, true);
}
}),
"[project]/components/workflow-editor/nodes/index.tsx [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "BashNode",
    ()=>BashNode,
    "ContainerNode",
    ()=>ContainerNode,
    "EndNode",
    ()=>EndNode,
    "ForeachNode",
    ()=>ForeachNode,
    "FunctionNode",
    ()=>FunctionNode,
    "HttpNode",
    ()=>HttpNode,
    "LlmNode",
    ()=>LlmNode,
    "ParallelNode",
    ()=>ParallelNode,
    "StartNode",
    ()=>StartNode,
    "nodeTypes",
    ()=>nodeTypes
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react-jsx-dev-runtime.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$react$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$locals$3e$__ = __turbopack_context__.i("[project]/node_modules/@xyflow/react/dist/esm/index.js [app-ssr] (ecmascript) <locals>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$system$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/@xyflow/system/dist/esm/index.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$nodes$2f$base$2d$node$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/workflow-editor/nodes/base-node.tsx [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$terminal$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__TerminalIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/terminal.js [app-ssr] (ecmascript) <export default as TerminalIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$layers$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__LayersIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/layers.js [app-ssr] (ecmascript) <export default as LayersIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$square$2d$function$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__FunctionSquareIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/square-function.js [app-ssr] (ecmascript) <export default as FunctionSquareIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$repeat$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__RepeatIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/repeat.js [app-ssr] (ecmascript) <export default as RepeatIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$play$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__PlayIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/play.js [app-ssr] (ecmascript) <export default as PlayIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$flag$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__FlagIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/flag.js [app-ssr] (ecmascript) <export default as FlagIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$globe$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__GlobeIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/globe.js [app-ssr] (ecmascript) <export default as GlobeIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$brain$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__BrainIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/brain.js [app-ssr] (ecmascript) <export default as BrainIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$box$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__BoxIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/box.js [app-ssr] (ecmascript) <export default as BoxIcon>");
"use client";
;
;
;
;
function StartNode({ selected }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
        className: "flex flex-col items-center",
        children: [
            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                className: `flex size-12 items-center justify-center rounded-full bg-green-500/20 border-2 border-green-500 ${selected ? "ring-2 ring-ring ring-offset-2" : ""}`,
                children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$play$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__PlayIcon$3e$__["PlayIcon"], {
                    className: "size-5 text-green-600 dark:text-green-400"
                }, void 0, false, {
                    fileName: "[project]/components/workflow-editor/nodes/index.tsx",
                    lineNumber: 33,
                    columnNumber: 9
                }, this)
            }, void 0, false, {
                fileName: "[project]/components/workflow-editor/nodes/index.tsx",
                lineNumber: 28,
                columnNumber: 7
            }, this),
            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("p", {
                className: "mt-2 text-xs font-medium text-muted-foreground",
                children: "Start"
            }, void 0, false, {
                fileName: "[project]/components/workflow-editor/nodes/index.tsx",
                lineNumber: 35,
                columnNumber: 7
            }, this),
            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$react$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$locals$3e$__["Handle"], {
                type: "source",
                position: __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$system$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Position"].Bottom,
                className: "!bg-green-500 !border-2 !border-background !size-3"
            }, void 0, false, {
                fileName: "[project]/components/workflow-editor/nodes/index.tsx",
                lineNumber: 36,
                columnNumber: 7
            }, this)
        ]
    }, void 0, true, {
        fileName: "[project]/components/workflow-editor/nodes/index.tsx",
        lineNumber: 27,
        columnNumber: 5
    }, this);
}
function EndNode({ selected }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
        className: "flex flex-col items-center",
        children: [
            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$react$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$locals$3e$__["Handle"], {
                type: "target",
                position: __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$system$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Position"].Top,
                className: "!bg-red-500 !border-2 !border-background !size-3"
            }, void 0, false, {
                fileName: "[project]/components/workflow-editor/nodes/index.tsx",
                lineNumber: 49,
                columnNumber: 7
            }, this),
            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                className: `flex size-12 items-center justify-center rounded-full bg-red-500/20 border-2 border-red-500 ${selected ? "ring-2 ring-ring ring-offset-2" : ""}`,
                children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$flag$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__FlagIcon$3e$__["FlagIcon"], {
                    className: "size-5 text-red-600 dark:text-red-400"
                }, void 0, false, {
                    fileName: "[project]/components/workflow-editor/nodes/index.tsx",
                    lineNumber: 59,
                    columnNumber: 9
                }, this)
            }, void 0, false, {
                fileName: "[project]/components/workflow-editor/nodes/index.tsx",
                lineNumber: 54,
                columnNumber: 7
            }, this),
            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("p", {
                className: "mt-2 text-xs font-medium text-muted-foreground",
                children: "End"
            }, void 0, false, {
                fileName: "[project]/components/workflow-editor/nodes/index.tsx",
                lineNumber: 61,
                columnNumber: 7
            }, this)
        ]
    }, void 0, true, {
        fileName: "[project]/components/workflow-editor/nodes/index.tsx",
        lineNumber: 48,
        columnNumber: 5
    }, this);
}
function BashNode({ data, selected }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$nodes$2f$base$2d$node$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["BaseNode"], {
        data: data,
        selected: selected,
        icon: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$terminal$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__TerminalIcon$3e$__["TerminalIcon"], {
            className: "size-4 text-blue-600 dark:text-blue-400"
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/nodes/index.tsx",
            lineNumber: 72,
            columnNumber: 13
        }, void 0),
        color: "bg-blue-500/20"
    }, void 0, false, {
        fileName: "[project]/components/workflow-editor/nodes/index.tsx",
        lineNumber: 69,
        columnNumber: 5
    }, this);
}
function ParallelNode({ data, selected }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$nodes$2f$base$2d$node$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["BaseNode"], {
        data: data,
        selected: selected,
        icon: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$layers$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__LayersIcon$3e$__["LayersIcon"], {
            className: "size-4 text-purple-600 dark:text-purple-400"
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/nodes/index.tsx",
            lineNumber: 84,
            columnNumber: 13
        }, void 0),
        color: "bg-purple-500/20"
    }, void 0, false, {
        fileName: "[project]/components/workflow-editor/nodes/index.tsx",
        lineNumber: 81,
        columnNumber: 5
    }, this);
}
function FunctionNode({ data, selected }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$nodes$2f$base$2d$node$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["BaseNode"], {
        data: data,
        selected: selected,
        icon: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$square$2d$function$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__FunctionSquareIcon$3e$__["FunctionSquareIcon"], {
            className: "size-4 text-green-600 dark:text-green-400"
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/nodes/index.tsx",
            lineNumber: 96,
            columnNumber: 13
        }, void 0),
        color: "bg-green-500/20"
    }, void 0, false, {
        fileName: "[project]/components/workflow-editor/nodes/index.tsx",
        lineNumber: 93,
        columnNumber: 5
    }, this);
}
function ForeachNode({ data, selected }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$nodes$2f$base$2d$node$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["BaseNode"], {
        data: data,
        selected: selected,
        icon: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$repeat$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__RepeatIcon$3e$__["RepeatIcon"], {
            className: "size-4 text-orange-600 dark:text-orange-400"
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/nodes/index.tsx",
            lineNumber: 108,
            columnNumber: 13
        }, void 0),
        color: "bg-orange-500/20"
    }, void 0, false, {
        fileName: "[project]/components/workflow-editor/nodes/index.tsx",
        lineNumber: 105,
        columnNumber: 5
    }, this);
}
function HttpNode({ data, selected }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$nodes$2f$base$2d$node$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["BaseNode"], {
        data: data,
        selected: selected,
        icon: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$globe$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__GlobeIcon$3e$__["GlobeIcon"], {
            className: "size-4 text-cyan-600 dark:text-cyan-400"
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/nodes/index.tsx",
            lineNumber: 119,
            columnNumber: 13
        }, void 0),
        color: "bg-cyan-500/20"
    }, void 0, false, {
        fileName: "[project]/components/workflow-editor/nodes/index.tsx",
        lineNumber: 116,
        columnNumber: 5
    }, this);
}
function LlmNode({ data, selected }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$nodes$2f$base$2d$node$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["BaseNode"], {
        data: data,
        selected: selected,
        icon: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$brain$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__BrainIcon$3e$__["BrainIcon"], {
            className: "size-4 text-pink-600 dark:text-pink-400"
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/nodes/index.tsx",
            lineNumber: 130,
            columnNumber: 13
        }, void 0),
        color: "bg-pink-500/20"
    }, void 0, false, {
        fileName: "[project]/components/workflow-editor/nodes/index.tsx",
        lineNumber: 127,
        columnNumber: 5
    }, this);
}
function ContainerNode({ data, selected }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$nodes$2f$base$2d$node$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["BaseNode"], {
        data: data,
        selected: selected,
        icon: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$box$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__BoxIcon$3e$__["BoxIcon"], {
            className: "size-4 text-slate-600 dark:text-slate-400"
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/nodes/index.tsx",
            lineNumber: 141,
            columnNumber: 13
        }, void 0),
        color: "bg-slate-500/20"
    }, void 0, false, {
        fileName: "[project]/components/workflow-editor/nodes/index.tsx",
        lineNumber: 138,
        columnNumber: 5
    }, this);
}
const nodeTypes = {
    start: StartNode,
    end: EndNode,
    bash: BashNode,
    parallel: ParallelNode,
    "parallel-steps": ParallelNode,
    function: FunctionNode,
    foreach: ForeachNode,
    http: HttpNode,
    llm: LlmNode,
    container: ContainerNode,
    "remote-bash": ContainerNode,
    module: ContainerNode
};
}),
"[project]/components/workflow-editor/utils/layout-engine.ts [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "layoutWorkflow",
    ()=>layoutWorkflow
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$dagre$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/dagre/index.js [app-ssr] (ecmascript)");
;
const NODE_WIDTH = 220;
const NODE_HEIGHT = 80;
function layoutWorkflow(nodes, edges, orientation = "TB") {
    const g = new __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$dagre$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"].graphlib.Graph();
    const dashedEdges = edges.filter((e)=>typeof e.style?.strokeDasharray === "string");
    const labeledEdges = edges.filter((e)=>typeof e.label === "string" && e.label.trim().length > 0);
    const branchingEdges = dashedEdges.length > 0 ? dashedEdges : labeledEdges;
    const decisionOutDegree = new Map();
    for (const e of branchingEdges){
        decisionOutDegree.set(e.source, (decisionOutDegree.get(e.source) ?? 0) + 1);
    }
    const maxDecisionOut = Math.max(0, ...Array.from(decisionOutDegree.values()));
    const hasDecisions = dashedEdges.length > 0 || labeledEdges.length > 0;
    const ranksep = hasDecisions ? 170 : 80;
    const nodesep = hasDecisions ? 120 + Math.max(0, maxDecisionOut - 1) * 40 : 50;
    const edgesep = hasDecisions ? 40 : 10;
    g.setGraph({
        rankdir: orientation,
        ranksep,
        nodesep,
        edgesep,
        marginx: hasDecisions ? 60 : 20,
        marginy: hasDecisions ? 60 : 20
    });
    g.setDefaultEdgeLabel(()=>({}));
    // Add nodes to dagre
    nodes.forEach((node)=>{
        g.setNode(node.id, {
            width: NODE_WIDTH,
            height: NODE_HEIGHT
        });
    });
    // Add edges to dagre
    edges.forEach((edge)=>{
        g.setEdge(edge.source, edge.target);
    });
    // Run layout
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$dagre$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"].layout(g);
    // Apply positions back to nodes
    return nodes.map((node)=>{
        const nodeWithPosition = g.node(node.id);
        if (nodeWithPosition) {
            return {
                ...node,
                position: {
                    x: nodeWithPosition.x - NODE_WIDTH / 2,
                    y: nodeWithPosition.y - NODE_HEIGHT / 2
                }
            };
        }
        return node;
    });
}
}),
"[project]/components/workflow-editor/workflow-canvas.tsx [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "WorkflowCanvas",
    ()=>WorkflowCanvas
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react-jsx-dev-runtime.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$react$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$locals$3e$__ = __turbopack_context__.i("[project]/node_modules/@xyflow/react/dist/esm/index.js [app-ssr] (ecmascript) <locals>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$maximize$2d$2$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__Maximize2Icon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/maximize-2.js [app-ssr] (ecmascript) <export default as Maximize2Icon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$minimize$2d$2$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__Minimize2Icon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/minimize-2.js [app-ssr] (ecmascript) <export default as Minimize2Icon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$nodes$2f$index$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/workflow-editor/nodes/index.tsx [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$utils$2f$layout$2d$engine$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/workflow-editor/utils/layout-engine.ts [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$canvas$2d$settings$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/workflow-editor/canvas-settings.tsx [app-ssr] (ecmascript)");
"use client";
;
;
;
;
;
;
;
;
function WorkflowCanvas({ initialNodes, initialEdges, onNodeSelect, orientation = "TB", wrapLongText = false, showDetails = true, hideMiniMap = false, selectedNodeId = null, onCanvasReady }) {
    const wrapperRef = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useRef"](null);
    const [isFullscreen, setIsFullscreen] = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useState"](false);
    const instanceRef = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useRef"](null);
    // Apply layout to initial nodes
    const layoutedNodes = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useMemo"](()=>(0, __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$utils$2f$layout$2d$engine$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["layoutWorkflow"])(initialNodes, initialEdges, orientation), [
        initialNodes,
        initialEdges,
        orientation
    ]);
    const [nodes, setNodes, onNodesChange] = (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$react$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$locals$3e$__["useNodesState"])(layoutedNodes);
    const [edges, setEdges, onEdgesChange] = (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$react$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$locals$3e$__["useEdgesState"])(initialEdges);
    // Update nodes when initial data changes
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useEffect"](()=>{
        const newLayoutedNodes = (0, __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$utils$2f$layout$2d$engine$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["layoutWorkflow"])(initialNodes, initialEdges, orientation);
        setNodes(newLayoutedNodes.map((n)=>({
                ...n,
                selected: selectedNodeId ? n.id === selectedNodeId : false
            })));
        setEdges(initialEdges);
    }, [
        initialNodes,
        initialEdges,
        orientation,
        setNodes,
        setEdges,
        selectedNodeId
    ]);
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useEffect"](()=>{
        if (selectedNodeId === null) {
            setNodes((prev)=>prev.map((n)=>n.selected ? {
                        ...n,
                        selected: false
                    } : n));
            return;
        }
        setNodes((prev)=>prev.map((n)=>({
                    ...n,
                    selected: n.id === selectedNodeId
                })));
    }, [
        selectedNodeId,
        setNodes
    ]);
    const focusNode = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useCallback"]((nodeId)=>{
        const instance = instanceRef.current;
        if (!instance) return;
        const node = typeof instance.getNode === "function" ? instance.getNode(nodeId) : null;
        if (!node) return;
        const width = node.measured?.width ?? node.width ?? 0;
        const height = node.measured?.height ?? node.height ?? 0;
        const x = (node.positionAbsolute?.x ?? node.position?.x ?? 0) + width / 2;
        const y = (node.positionAbsolute?.y ?? node.position?.y ?? 0) + height / 2;
        if (typeof instance.setCenter === "function") {
            const currentZoom = typeof instance.getZoom === "function" ? instance.getZoom() : 1;
            const targetZoom = Math.max(currentZoom ?? 1, 1.09);
            instance.setCenter(x, y, {
                zoom: targetZoom,
                duration: 500
            });
        }
    }, []);
    const handleNodesChange = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useCallback"]((changes)=>{
        onNodesChange(changes);
        // Check for selection changes
        const selectedChange = changes.find((change)=>change.type === "select" && change.selected);
        if (selectedChange && selectedChange.type === "select") {
            onNodeSelect?.(selectedChange.id);
            return;
        }
        // Also check if all nodes are deselected
        const hasSelection = changes.some((change)=>change.type === "select" && change.selected);
        if (!hasSelection && changes.some((c)=>c.type === "select")) {
            // Check if any node is still selected
            const stillSelected = nodes.some((n)=>n.selected && !changes.find((c)=>c.type === "select" && c.id === n.id && !c.selected));
            if (!stillSelected) {
                onNodeSelect?.(null);
            }
        }
    }, [
        onNodesChange,
        onNodeSelect,
        nodes
    ]);
    const handleEdgesChange = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useCallback"]((changes)=>{
        onEdgesChange(changes);
    }, [
        onEdgesChange
    ]);
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useEffect"](()=>{
        const handler = ()=>{
            setIsFullscreen(Boolean(document.fullscreenElement));
        };
        document.addEventListener("fullscreenchange", handler);
        document.addEventListener("webkitfullscreenchange", handler);
        handler();
        return ()=>{
            document.removeEventListener("fullscreenchange", handler);
            document.removeEventListener("webkitfullscreenchange", handler);
        };
    }, []);
    const toggleFullscreen = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useCallback"](async ()=>{
        try {
            if (document.fullscreenElement) {
                await document.exitFullscreen();
                return;
            }
            const el = wrapperRef.current;
            if (!el) return;
            if (typeof el.requestFullscreen === "function") {
                await el.requestFullscreen();
                return;
            }
            if (typeof el.webkitRequestFullscreen === "function") {
                el.webkitRequestFullscreen();
            }
        } catch  {}
    }, []);
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
        ref: wrapperRef,
        className: "h-full w-full",
        children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$canvas$2d$settings$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["CanvasSettingsProvider"], {
            wrapLongText: wrapLongText,
            showDetails: showDetails,
            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$react$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$locals$3e$__["ReactFlow"], {
                nodes: nodes,
                edges: edges,
                onNodesChange: handleNodesChange,
                onEdgesChange: handleEdgesChange,
                nodeTypes: __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$nodes$2f$index$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["nodeTypes"],
                onInit: (instance)=>{
                    instanceRef.current = instance;
                    onCanvasReady?.({
                        focusNode
                    });
                },
                fitView: true,
                fitViewOptions: {
                    padding: 0.2
                },
                minZoom: 0.1,
                maxZoom: 2,
                defaultEdgeOptions: {
                    type: "smoothstep"
                },
                proOptions: {
                    hideAttribution: true
                },
                children: [
                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$react$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$locals$3e$__["Background"], {
                        variant: __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$react$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$locals$3e$__["BackgroundVariant"].Dots,
                        gap: 16,
                        size: 1,
                        className: "bg-muted/30"
                    }, void 0, false, {
                        fileName: "[project]/components/workflow-editor/workflow-canvas.tsx",
                        lineNumber: 192,
                        columnNumber: 11
                    }, this),
                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$react$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$locals$3e$__["Controls"], {
                        className: "rounded-lg border bg-card shadow-sm [&>button]:border-border [&>button]:bg-card [&>button:hover]:bg-muted",
                        showInteractive: false,
                        children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$react$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$locals$3e$__["ControlButton"], {
                            onClick: toggleFullscreen,
                            title: isFullscreen ? "Exit fullscreen" : "Fullscreen",
                            children: isFullscreen ? /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$minimize$2d$2$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__Minimize2Icon$3e$__["Minimize2Icon"], {
                                className: "size-4"
                            }, void 0, false, {
                                fileName: "[project]/components/workflow-editor/workflow-canvas.tsx",
                                lineNumber: 203,
                                columnNumber: 31
                            }, this) : /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$maximize$2d$2$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__Maximize2Icon$3e$__["Maximize2Icon"], {
                                className: "size-4"
                            }, void 0, false, {
                                fileName: "[project]/components/workflow-editor/workflow-canvas.tsx",
                                lineNumber: 203,
                                columnNumber: 70
                            }, this)
                        }, void 0, false, {
                            fileName: "[project]/components/workflow-editor/workflow-canvas.tsx",
                            lineNumber: 202,
                            columnNumber: 13
                        }, this)
                    }, void 0, false, {
                        fileName: "[project]/components/workflow-editor/workflow-canvas.tsx",
                        lineNumber: 198,
                        columnNumber: 11
                    }, this),
                    !hideMiniMap && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$xyflow$2f$react$2f$dist$2f$esm$2f$index$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$locals$3e$__["MiniMap"], {
                        className: "rounded-lg border bg-card shadow-sm",
                        nodeStrokeWidth: 3,
                        pannable: true,
                        zoomable: true
                    }, void 0, false, {
                        fileName: "[project]/components/workflow-editor/workflow-canvas.tsx",
                        lineNumber: 207,
                        columnNumber: 13
                    }, this)
                ]
            }, void 0, true, {
                fileName: "[project]/components/workflow-editor/workflow-canvas.tsx",
                lineNumber: 171,
                columnNumber: 9
            }, this)
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/workflow-canvas.tsx",
            lineNumber: 170,
            columnNumber: 7
        }, this)
    }, void 0, false, {
        fileName: "[project]/components/workflow-editor/workflow-canvas.tsx",
        lineNumber: 169,
        columnNumber: 5
    }, this);
}
}),
"[project]/components/ui/scroll-area.tsx [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "ScrollArea",
    ()=>ScrollArea,
    "ScrollBar",
    ()=>ScrollBar
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react-jsx-dev-runtime.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$radix$2d$ui$2f$react$2d$scroll$2d$area$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/@radix-ui/react-scroll-area/dist/index.mjs [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/utils.ts [app-ssr] (ecmascript)");
"use client";
;
;
;
function ScrollArea({ className, children, ...props }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$radix$2d$ui$2f$react$2d$scroll$2d$area$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Root"], {
        "data-slot": "scroll-area",
        className: (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["cn"])("relative overflow-hidden", className),
        ...props,
        children: [
            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$radix$2d$ui$2f$react$2d$scroll$2d$area$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Viewport"], {
                className: "h-full w-full rounded-[inherit]",
                children: children
            }, void 0, false, {
                fileName: "[project]/components/ui/scroll-area.tsx",
                lineNumber: 18,
                columnNumber: 7
            }, this),
            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(ScrollBar, {}, void 0, false, {
                fileName: "[project]/components/ui/scroll-area.tsx",
                lineNumber: 21,
                columnNumber: 7
            }, this),
            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$radix$2d$ui$2f$react$2d$scroll$2d$area$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Corner"], {}, void 0, false, {
                fileName: "[project]/components/ui/scroll-area.tsx",
                lineNumber: 22,
                columnNumber: 7
            }, this)
        ]
    }, void 0, true, {
        fileName: "[project]/components/ui/scroll-area.tsx",
        lineNumber: 13,
        columnNumber: 5
    }, this);
}
function ScrollBar({ className, orientation = "vertical", ...props }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$radix$2d$ui$2f$react$2d$scroll$2d$area$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["ScrollAreaScrollbar"], {
        "data-slot": "scroll-bar",
        orientation: orientation,
        className: (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["cn"])("flex touch-none select-none transition-colors", orientation === "vertical" && "h-full w-2.5 border-l border-l-transparent p-[1px]", orientation === "horizontal" && "h-2.5 flex-col border-t border-t-transparent p-[1px]", className),
        ...props,
        children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$radix$2d$ui$2f$react$2d$scroll$2d$area$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["ScrollAreaThumb"], {
            className: "relative flex-1 rounded-full bg-border"
        }, void 0, false, {
            fileName: "[project]/components/ui/scroll-area.tsx",
            lineNumber: 46,
            columnNumber: 7
        }, this)
    }, void 0, false, {
        fileName: "[project]/components/ui/scroll-area.tsx",
        lineNumber: 33,
        columnNumber: 5
    }, this);
}
;
}),
"[project]/components/ui/label.tsx [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "Label",
    ()=>Label
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react-jsx-dev-runtime.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$radix$2d$ui$2f$react$2d$label$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/@radix-ui/react-label/dist/index.mjs [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/utils.ts [app-ssr] (ecmascript)");
"use client";
;
;
;
function Label({ className, ...props }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$radix$2d$ui$2f$react$2d$label$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Root"], {
        "data-slot": "label",
        className: (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["cn"])("flex items-center gap-2 text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70", className),
        ...props
    }, void 0, false, {
        fileName: "[project]/components/ui/label.tsx",
        lineNumber: 12,
        columnNumber: 5
    }, this);
}
;
}),
"[project]/components/ui/tabs.tsx [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "Tabs",
    ()=>Tabs,
    "TabsContent",
    ()=>TabsContent,
    "TabsList",
    ()=>TabsList,
    "TabsTrigger",
    ()=>TabsTrigger
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react-jsx-dev-runtime.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$radix$2d$ui$2f$react$2d$tabs$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/@radix-ui/react-tabs/dist/index.mjs [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/utils.ts [app-ssr] (ecmascript)");
"use client";
;
;
;
function Tabs({ className, ...props }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$radix$2d$ui$2f$react$2d$tabs$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Root"], {
        "data-slot": "tabs",
        className: (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["cn"])("flex flex-col gap-2", className),
        ...props
    }, void 0, false, {
        fileName: "[project]/components/ui/tabs.tsx",
        lineNumber: 9,
        columnNumber: 5
    }, this);
}
function TabsList({ className, ...props }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$radix$2d$ui$2f$react$2d$tabs$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["List"], {
        "data-slot": "tabs-list",
        className: (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["cn"])("inline-flex h-9 items-center justify-center rounded-lg bg-muted p-1 text-muted-foreground", className),
        ...props
    }, void 0, false, {
        fileName: "[project]/components/ui/tabs.tsx",
        lineNumber: 22,
        columnNumber: 5
    }, this);
}
function TabsTrigger({ className, ...props }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$radix$2d$ui$2f$react$2d$tabs$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Trigger"], {
        "data-slot": "tabs-trigger",
        className: (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["cn"])("inline-flex items-center justify-center whitespace-nowrap rounded-md px-3 py-1 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-xs", className),
        ...props
    }, void 0, false, {
        fileName: "[project]/components/ui/tabs.tsx",
        lineNumber: 38,
        columnNumber: 5
    }, this);
}
function TabsContent({ className, ...props }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$radix$2d$ui$2f$react$2d$tabs$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Content"], {
        "data-slot": "tabs-content",
        className: (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["cn"])("mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2", className),
        ...props
    }, void 0, false, {
        fileName: "[project]/components/ui/tabs.tsx",
        lineNumber: 54,
        columnNumber: 5
    }, this);
}
;
}),
"[project]/components/workflow-editor/workflow-sidebar.tsx [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "WorkflowSidebar",
    ()=>WorkflowSidebar
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react-jsx-dev-runtime.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$scroll$2d$area$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/ui/scroll-area.tsx [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$input$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/ui/input.tsx [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/ui/label.tsx [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$badge$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/ui/badge.tsx [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$separator$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/ui/separator.tsx [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$tabs$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/ui/tabs.tsx [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/ui/button.tsx [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$terminal$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__TerminalIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/terminal.js [app-ssr] (ecmascript) <export default as TerminalIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$layers$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__LayersIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/layers.js [app-ssr] (ecmascript) <export default as LayersIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$square$2d$function$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__FunctionSquareIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/square-function.js [app-ssr] (ecmascript) <export default as FunctionSquareIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$repeat$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__RepeatIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/repeat.js [app-ssr] (ecmascript) <export default as RepeatIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$clock$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__ClockIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/clock.js [app-ssr] (ecmascript) <export default as ClockIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$clipboard$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__ClipboardIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/clipboard.js [app-ssr] (ecmascript) <export default as ClipboardIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$globe$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__GlobeIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/globe.js [app-ssr] (ecmascript) <export default as GlobeIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$brain$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__BrainIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/brain.js [app-ssr] (ecmascript) <export default as BrainIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$box$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__BoxIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/box.js [app-ssr] (ecmascript) <export default as BoxIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$light$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__Light$3e$__ = __turbopack_context__.i("[project]/node_modules/react-syntax-highlighter/dist/esm/light.js [app-ssr] (ecmascript) <export default as Light>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$languages$2f$hljs$2f$yaml$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/react-syntax-highlighter/dist/esm/languages/hljs/yaml.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$languages$2f$hljs$2f$bash$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/react-syntax-highlighter/dist/esm/languages/hljs/bash.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$languages$2f$hljs$2f$javascript$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/react-syntax-highlighter/dist/esm/languages/hljs/javascript.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$github$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/react-syntax-highlighter/dist/esm/styles/hljs/github.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$atom$2d$one$2d$dark$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/react-syntax-highlighter/dist/esm/styles/hljs/atom-one-dark.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2d$themes$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next-themes/dist/index.mjs [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$sonner$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/sonner/dist/index.mjs [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$js$2d$yaml$2f$dist$2f$js$2d$yaml$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/js-yaml/dist/js-yaml.mjs [app-ssr] (ecmascript)");
"use client";
;
;
;
;
;
;
;
;
;
;
;
;
;
;
;
;
;
;
;
function WorkflowSidebar({ selectedStep, selectedModule = null, yamlPreview, wrapLongText = false, onStepUpdate, workflowKind = null, allSteps = [], allModules = [], onNavigateToNode }) {
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$light$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__Light$3e$__["Light"].registerLanguage("yaml", __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$languages$2f$hljs$2f$yaml$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"]);
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$light$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__Light$3e$__["Light"].registerLanguage("bash", __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$languages$2f$hljs$2f$bash$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"]);
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$light$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__Light$3e$__["Light"].registerLanguage("javascript", __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$languages$2f$hljs$2f$javascript$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"]);
    const { theme } = (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2d$themes$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useTheme"])();
    const stepTypeIcons = {
        bash: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$terminal$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__TerminalIcon$3e$__["TerminalIcon"], {
            className: "size-4"
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
            lineNumber: 61,
            columnNumber: 11
        }, this),
        parallel: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$layers$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__LayersIcon$3e$__["LayersIcon"], {
            className: "size-4"
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
            lineNumber: 62,
            columnNumber: 15
        }, this),
        "parallel-steps": /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$layers$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__LayersIcon$3e$__["LayersIcon"], {
            className: "size-4"
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
            lineNumber: 63,
            columnNumber: 23
        }, this),
        function: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$square$2d$function$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__FunctionSquareIcon$3e$__["FunctionSquareIcon"], {
            className: "size-4"
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
            lineNumber: 64,
            columnNumber: 15
        }, this),
        foreach: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$repeat$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__RepeatIcon$3e$__["RepeatIcon"], {
            className: "size-4"
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
            lineNumber: 65,
            columnNumber: 14
        }, this),
        http: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$globe$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__GlobeIcon$3e$__["GlobeIcon"], {
            className: "size-4"
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
            lineNumber: 66,
            columnNumber: 11
        }, this),
        llm: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$brain$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__BrainIcon$3e$__["BrainIcon"], {
            className: "size-4"
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
            lineNumber: 67,
            columnNumber: 10
        }, this),
        container: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$box$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__BoxIcon$3e$__["BoxIcon"], {
            className: "size-4"
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
            lineNumber: 68,
            columnNumber: 16
        }, this),
        "remote-bash": /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$box$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__BoxIcon$3e$__["BoxIcon"], {
            className: "size-4"
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
            lineNumber: 69,
            columnNumber: 20
        }, this),
        module: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$box$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__BoxIcon$3e$__["BoxIcon"], {
            className: "size-4"
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
            lineNumber: 70,
            columnNumber: 13
        }, this)
    };
    const CodeHighlighter = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$light$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__Light$3e$__["Light"];
    const selectionType = selectedStep?.type ?? (selectedModule ? "module" : "");
    const selectionName = selectedStep?.name ?? selectedModule?.name ?? "";
    const [activeTab, setActiveTab] = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useState"]("properties");
    const badgeVariantForStepType = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useCallback"]((type)=>{
        switch(type){
            case "bash":
                return "info";
            case "remote-bash":
                return "info";
            case "container":
                return "warning";
            case "parallel":
                return "cyan";
            case "parallel-steps":
                return "cyan";
            case "function":
                return "purple";
            case "foreach":
                return "orange";
            case "http":
                return "success";
            case "llm":
                return "pink";
            default:
                return "secondary";
        }
    }, []);
    const bashResolvedCommand = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useMemo"](()=>{
        if (!selectedStep) return "";
        const hasStructuredArgs = selectedStep.speed_args !== undefined || selectedStep.config_args !== undefined || selectedStep.input_args !== undefined || selectedStep.output_args !== undefined;
        if (!hasStructuredArgs) return "";
        const parts = [
            selectedStep.command,
            selectedStep.speed_args,
            selectedStep.config_args,
            selectedStep.input_args,
            selectedStep.output_args
        ].map((p)=>typeof p === "string" ? p.trim() : "").filter(Boolean);
        return parts.join(" ");
    }, [
        selectedStep
    ]);
    const foreachStepYaml = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useMemo"](()=>{
        if (!selectedStep || selectedStep.type !== "foreach") return "";
        if (!selectedStep.step) return "";
        try {
            return __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$js$2d$yaml$2f$dist$2f$js$2d$yaml$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"].dump(selectedStep.step, {
                indent: 2,
                lineWidth: -1,
                noRefs: true,
                quotingType: '"'
            });
        } catch  {
            return "";
        }
    }, [
        selectedStep
    ]);
    const renderStringList = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useCallback"]((label, items, opts)=>{
        const language = opts?.language;
        const copyAllText = opts?.copyAllText;
        return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
            className: "space-y-2",
            children: [
                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                    className: "flex items-center justify-between gap-2",
                    children: [
                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                            children: [
                                label,
                                " (",
                                items.length,
                                ")"
                            ]
                        }, void 0, true, {
                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                            lineNumber: 170,
                            columnNumber: 13
                        }, this),
                        copyAllText !== undefined && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Button"], {
                            className: "rounded-md",
                            variant: "outline",
                            size: "icon",
                            onClick: async ()=>{
                                try {
                                    await navigator.clipboard.writeText(copyAllText);
                                    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$sonner$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["toast"].success("Copied to clipboard");
                                } catch  {
                                    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$sonner$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["toast"].error("Failed to copy");
                                }
                            },
                            children: [
                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$clipboard$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__ClipboardIcon$3e$__["ClipboardIcon"], {
                                    className: "size-4"
                                }, void 0, false, {
                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                    lineNumber: 187,
                                    columnNumber: 17
                                }, this),
                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("span", {
                                    className: "sr-only",
                                    children: "Copy all"
                                }, void 0, false, {
                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                    lineNumber: 188,
                                    columnNumber: 17
                                }, this)
                            ]
                        }, void 0, true, {
                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                            lineNumber: 174,
                            columnNumber: 15
                        }, this)
                    ]
                }, void 0, true, {
                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                    lineNumber: 169,
                    columnNumber: 11
                }, this),
                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                    className: "space-y-2",
                    children: items.map((item, idx)=>/*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                            className: "rounded-md border bg-muted/30 p-2",
                            children: [
                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                    className: "mb-2 flex items-center justify-between gap-2",
                                    children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$badge$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Badge"], {
                                        variant: "secondary",
                                        className: "text-[10px]",
                                        children: idx + 1
                                    }, void 0, false, {
                                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                        lineNumber: 197,
                                        columnNumber: 19
                                    }, this)
                                }, void 0, false, {
                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                    lineNumber: 196,
                                    columnNumber: 17
                                }, this),
                                language ? /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(CodeHighlighter, {
                                    language: language,
                                    style: theme === "dark" ? __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$atom$2d$one$2d$dark$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"] : __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$github$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"],
                                    customStyle: {
                                        margin: 0,
                                        background: "transparent",
                                        fontSize: "0.75rem",
                                        whiteSpace: "pre-wrap",
                                        wordBreak: "break-word"
                                    },
                                    codeTagProps: {
                                        style: {
                                            whiteSpace: "pre-wrap",
                                            wordBreak: "break-word"
                                        }
                                    },
                                    children: item
                                }, void 0, false, {
                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                    lineNumber: 202,
                                    columnNumber: 19
                                }, this) : /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                    className: "font-mono text-xs whitespace-pre-wrap break-words",
                                    children: item
                                }, void 0, false, {
                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                    lineNumber: 222,
                                    columnNumber: 19
                                }, this)
                            ]
                        }, `${idx}-${item}`, true, {
                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                            lineNumber: 195,
                            columnNumber: 15
                        }, this))
                }, void 0, false, {
                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                    lineNumber: 193,
                    columnNumber: 11
                }, this)
            ]
        }, void 0, true, {
            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
            lineNumber: 168,
            columnNumber: 9
        }, this);
    }, [
        CodeHighlighter,
        theme
    ]);
    const renderDecisionBlock = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useCallback"]((decision)=>{
        if (!decision) return null;
        if (Array.isArray(decision) && decision.length > 0) {
            return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                className: "space-y-2",
                children: [
                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                        children: "Decision Rules"
                    }, void 0, false, {
                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                        lineNumber: 239,
                        columnNumber: 11
                    }, this),
                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                        className: "space-y-2",
                        children: decision.map((rule, i)=>/*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                className: "rounded-md border p-2 text-xs font-mono overflow-hidden",
                                children: [
                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                        className: "text-muted-foreground whitespace-pre-wrap break-words",
                                        children: [
                                            "if ",
                                            rule?.condition
                                        ]
                                    }, void 0, true, {
                                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                        lineNumber: 243,
                                        columnNumber: 17
                                    }, this),
                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                        className: "text-primary whitespace-pre-wrap break-words",
                                        children: [
                                            "→ ",
                                            rule?.next
                                        ]
                                    }, void 0, true, {
                                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                        lineNumber: 244,
                                        columnNumber: 17
                                    }, this)
                                ]
                            }, i, true, {
                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                lineNumber: 242,
                                columnNumber: 15
                            }, this))
                    }, void 0, false, {
                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                        lineNumber: 240,
                        columnNumber: 11
                    }, this)
                ]
            }, void 0, true, {
                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                lineNumber: 238,
                columnNumber: 9
            }, this);
        }
        if (typeof decision === "object" && decision && typeof decision.switch === "string" && decision.cases && typeof decision.cases === "object") {
            const cases = decision.cases;
            const entries = Object.entries(cases).filter(([k])=>typeof k === "string" && k.trim().length > 0);
            const hasDefault = decision.default && (typeof decision.default.goto === "string" || typeof decision.default.next === "string");
            const defaultGoto = typeof decision.default?.goto === "string" ? decision.default.goto : typeof decision.default?.next === "string" ? decision.default.next : "";
            return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                className: "space-y-2",
                children: [
                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                        children: "Decision (switch)"
                    }, void 0, false, {
                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                        lineNumber: 260,
                        columnNumber: 11
                    }, this),
                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                        className: "rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words",
                        children: [
                            "switch ",
                            decision.switch
                        ]
                    }, void 0, true, {
                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                        lineNumber: 261,
                        columnNumber: 11
                    }, this),
                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                        className: "space-y-2",
                        children: [
                            entries.map(([key, value])=>{
                                const goto = typeof value?.goto === "string" ? value.goto : typeof value?.next === "string" ? value.next : "";
                                return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                    className: "rounded-md border p-2 text-xs font-mono overflow-hidden",
                                    children: [
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                            className: "text-muted-foreground whitespace-pre-wrap break-words",
                                            children: [
                                                "case ",
                                                key
                                            ]
                                        }, void 0, true, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 269,
                                            columnNumber: 19
                                        }, this),
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                            className: "text-primary whitespace-pre-wrap break-words",
                                            children: [
                                                "→ ",
                                                goto
                                            ]
                                        }, void 0, true, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 270,
                                            columnNumber: 19
                                        }, this)
                                    ]
                                }, key, true, {
                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                    lineNumber: 268,
                                    columnNumber: 17
                                }, this);
                            }),
                            hasDefault && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                className: "rounded-md border p-2 text-xs font-mono overflow-hidden",
                                children: [
                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                        className: "text-muted-foreground whitespace-pre-wrap break-words",
                                        children: "default"
                                    }, void 0, false, {
                                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                        lineNumber: 276,
                                        columnNumber: 17
                                    }, this),
                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                        className: "text-primary whitespace-pre-wrap break-words",
                                        children: [
                                            "→ ",
                                            defaultGoto
                                        ]
                                    }, void 0, true, {
                                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                        lineNumber: 277,
                                        columnNumber: 17
                                    }, this)
                                ]
                            }, void 0, true, {
                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                lineNumber: 275,
                                columnNumber: 15
                            }, this)
                        ]
                    }, void 0, true, {
                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                        lineNumber: 264,
                        columnNumber: 11
                    }, this)
                ]
            }, void 0, true, {
                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                lineNumber: 259,
                columnNumber: 9
            }, this);
        }
        return null;
    }, []);
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
        className: "flex h-full flex-col border-l bg-card",
        children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$tabs$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Tabs"], {
            value: activeTab,
            onValueChange: setActiveTab,
            className: "flex h-full flex-col",
            children: [
                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                    className: "border-b px-4 py-2",
                    children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$tabs$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["TabsList"], {
                        className: "w-full",
                        children: [
                            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$tabs$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["TabsTrigger"], {
                                value: "properties",
                                className: "flex-1",
                                children: "Properties"
                            }, void 0, false, {
                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                lineNumber: 293,
                                columnNumber: 13
                            }, this),
                            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$tabs$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["TabsTrigger"], {
                                value: "items",
                                className: "flex-1",
                                children: workflowKind === "flow" ? "Modules" : "Steps"
                            }, void 0, false, {
                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                lineNumber: 296,
                                columnNumber: 13
                            }, this),
                            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$tabs$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["TabsTrigger"], {
                                value: "yaml",
                                className: "flex-1",
                                children: "YAML"
                            }, void 0, false, {
                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                lineNumber: 299,
                                columnNumber: 13
                            }, this)
                        ]
                    }, void 0, true, {
                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                        lineNumber: 292,
                        columnNumber: 11
                    }, this)
                }, void 0, false, {
                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                    lineNumber: 291,
                    columnNumber: 9
                }, this),
                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$tabs$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["TabsContent"], {
                    value: "properties",
                    className: "flex-1 m-0 min-h-0",
                    children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$scroll$2d$area$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["ScrollArea"], {
                        className: "h-full",
                        children: selectedStep || selectedModule ? /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                            className: "p-4 space-y-6",
                            children: [
                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                    className: "space-y-2",
                                    children: [
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                            className: "flex items-center gap-2",
                                            children: [
                                                stepTypeIcons[selectionType],
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$badge$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Badge"], {
                                                    variant: "secondary",
                                                    className: "capitalize",
                                                    children: selectionType
                                                }, void 0, false, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 313,
                                                    columnNumber: 21
                                                }, this)
                                            ]
                                        }, void 0, true, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 311,
                                            columnNumber: 19
                                        }, this),
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("h3", {
                                            className: "text-lg font-semibold",
                                            children: selectionName
                                        }, void 0, false, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 317,
                                            columnNumber: 19
                                        }, this)
                                    ]
                                }, void 0, true, {
                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                    lineNumber: 310,
                                    columnNumber: 17
                                }, this),
                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$separator$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Separator"], {}, void 0, false, {
                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                    lineNumber: 320,
                                    columnNumber: 17
                                }, this),
                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                    className: "space-y-4",
                                    children: [
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                            className: "space-y-2",
                                            children: [
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                    htmlFor: "stepName",
                                                    children: "Name"
                                                }, void 0, false, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 325,
                                                    columnNumber: 21
                                                }, this),
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$input$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Input"], {
                                                    id: "stepName",
                                                    value: selectionName,
                                                    onChange: (e)=>selectedStep ? onStepUpdate?.(selectedStep.name, {
                                                            name: e.target.value
                                                        }) : undefined,
                                                    disabled: !selectedStep
                                                }, void 0, false, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 326,
                                                    columnNumber: 21
                                                }, this)
                                            ]
                                        }, void 0, true, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 324,
                                            columnNumber: 19
                                        }, this),
                                        selectedModule && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                            className: "space-y-4",
                                            children: [
                                                selectedModule.path && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "Path"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 340,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words",
                                                            children: selectedModule.path
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 341,
                                                            columnNumber: 27
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 339,
                                                    columnNumber: 25
                                                }, this),
                                                Array.isArray(selectedModule.depends_on) && selectedModule.depends_on.length > 0 && renderStringList("Depends On", selectedModule.depends_on.map((d)=>String(d)), {
                                                    copyAllText: selectedModule.depends_on.join("\n")
                                                }),
                                                selectedModule.condition && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "Condition"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 356,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words",
                                                            children: selectedModule.condition
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 357,
                                                            columnNumber: 27
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 355,
                                                    columnNumber: 25
                                                }, this),
                                                selectedModule.params && Object.keys(selectedModule.params).length > 0 && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "Params"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 365,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border bg-muted/30 p-2",
                                                            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(CodeHighlighter, {
                                                                language: "json",
                                                                style: theme === "dark" ? __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$atom$2d$one$2d$dark$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"] : __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$github$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"],
                                                                customStyle: {
                                                                    margin: 0,
                                                                    background: "transparent",
                                                                    fontSize: "0.75rem",
                                                                    whiteSpace: "pre-wrap",
                                                                    wordBreak: "break-word"
                                                                },
                                                                codeTagProps: {
                                                                    style: {
                                                                        whiteSpace: "pre-wrap",
                                                                        wordBreak: "break-word"
                                                                    }
                                                                },
                                                                children: JSON.stringify(selectedModule.params, null, 2)
                                                            }, void 0, false, {
                                                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                lineNumber: 367,
                                                                columnNumber: 29
                                                            }, this)
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 366,
                                                            columnNumber: 27
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 364,
                                                    columnNumber: 25
                                                }, this),
                                                selectedModule?.decision && renderDecisionBlock(selectedModule.decision),
                                                Array.isArray(selectedModule.on_success) && selectedModule.on_success.length > 0 && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "On Success"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 394,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border bg-muted/30 p-2",
                                                            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(CodeHighlighter, {
                                                                language: "json",
                                                                style: theme === "dark" ? __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$atom$2d$one$2d$dark$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"] : __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$github$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"],
                                                                customStyle: {
                                                                    margin: 0,
                                                                    background: "transparent",
                                                                    fontSize: "0.75rem",
                                                                    whiteSpace: "pre-wrap",
                                                                    wordBreak: "break-word"
                                                                },
                                                                codeTagProps: {
                                                                    style: {
                                                                        whiteSpace: "pre-wrap",
                                                                        wordBreak: "break-word"
                                                                    }
                                                                },
                                                                children: JSON.stringify(selectedModule.on_success, null, 2)
                                                            }, void 0, false, {
                                                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                lineNumber: 396,
                                                                columnNumber: 29
                                                            }, this)
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 395,
                                                            columnNumber: 27
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 393,
                                                    columnNumber: 25
                                                }, this),
                                                Array.isArray(selectedModule.on_error) && selectedModule.on_error.length > 0 && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "On Error"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 421,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border bg-muted/30 p-2",
                                                            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(CodeHighlighter, {
                                                                language: "json",
                                                                style: theme === "dark" ? __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$atom$2d$one$2d$dark$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"] : __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$github$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"],
                                                                customStyle: {
                                                                    margin: 0,
                                                                    background: "transparent",
                                                                    fontSize: "0.75rem",
                                                                    whiteSpace: "pre-wrap",
                                                                    wordBreak: "break-word"
                                                                },
                                                                codeTagProps: {
                                                                    style: {
                                                                        whiteSpace: "pre-wrap",
                                                                        wordBreak: "break-word"
                                                                    }
                                                                },
                                                                children: JSON.stringify(selectedModule.on_error, null, 2)
                                                            }, void 0, false, {
                                                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                lineNumber: 423,
                                                                columnNumber: 29
                                                            }, this)
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 422,
                                                            columnNumber: 27
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 420,
                                                    columnNumber: 25
                                                }, this)
                                            ]
                                        }, void 0, true, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 337,
                                            columnNumber: 21
                                        }, this),
                                        selectedStep && (selectedStep.type === "bash" || selectedStep.type === "remote-bash" || selectedStep.type === "container") && selectedStep.command && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                            className: "space-y-2",
                                            children: [
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "flex items-center justify-between gap-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            htmlFor: "command",
                                                            children: "Command"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 451,
                                                            columnNumber: 25
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Button"], {
                                                            className: "rounded-md",
                                                            variant: "outline",
                                                            size: "icon",
                                                            onClick: async ()=>{
                                                                try {
                                                                    await navigator.clipboard.writeText(selectedStep.command || "");
                                                                    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$sonner$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["toast"].success("Copied to clipboard");
                                                                } catch  {
                                                                    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$sonner$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["toast"].error("Failed to copy");
                                                                }
                                                            },
                                                            children: [
                                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$clipboard$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__ClipboardIcon$3e$__["ClipboardIcon"], {
                                                                    className: "size-4"
                                                                }, void 0, false, {
                                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                    lineNumber: 465,
                                                                    columnNumber: 27
                                                                }, this),
                                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("span", {
                                                                    className: "sr-only",
                                                                    children: "Copy command"
                                                                }, void 0, false, {
                                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                    lineNumber: 466,
                                                                    columnNumber: 27
                                                                }, this)
                                                            ]
                                                        }, void 0, true, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 452,
                                                            columnNumber: 25
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 450,
                                                    columnNumber: 23
                                                }, this),
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "rounded-md border bg-muted/30 p-2",
                                                    children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(CodeHighlighter, {
                                                        language: "bash",
                                                        style: theme === "dark" ? __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$atom$2d$one$2d$dark$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"] : __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$github$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"],
                                                        customStyle: {
                                                            margin: 0,
                                                            background: "transparent",
                                                            fontSize: "0.75rem",
                                                            whiteSpace: "pre-wrap",
                                                            wordBreak: "break-word"
                                                        },
                                                        codeTagProps: {
                                                            style: {
                                                                whiteSpace: "pre-wrap",
                                                                wordBreak: "break-word"
                                                            }
                                                        },
                                                        children: selectedStep.command
                                                    }, void 0, false, {
                                                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                        lineNumber: 470,
                                                        columnNumber: 25
                                                    }, this)
                                                }, void 0, false, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 469,
                                                    columnNumber: 23
                                                }, this)
                                            ]
                                        }, void 0, true, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 449,
                                            columnNumber: 21
                                        }, this),
                                        selectedStep && (selectedStep.type === "bash" || selectedStep.type === "remote-bash" || selectedStep.type === "container") && (selectedStep.speed_args !== undefined || selectedStep.config_args !== undefined || selectedStep.input_args !== undefined || selectedStep.output_args !== undefined) && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                            className: "space-y-2",
                                            children: [
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                    children: "Structured Args"
                                                }, void 0, false, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 500,
                                                    columnNumber: 23
                                                }, this),
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border p-2 text-xs font-mono overflow-hidden",
                                                            children: [
                                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                                    className: "text-muted-foreground whitespace-pre-wrap break-words",
                                                                    children: "speed_args"
                                                                }, void 0, false, {
                                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                    lineNumber: 503,
                                                                    columnNumber: 27
                                                                }, this),
                                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                                    className: "whitespace-pre-wrap break-words",
                                                                    children: selectedStep.speed_args ?? ""
                                                                }, void 0, false, {
                                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                    lineNumber: 504,
                                                                    columnNumber: 27
                                                                }, this)
                                                            ]
                                                        }, void 0, true, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 502,
                                                            columnNumber: 25
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border p-2 text-xs font-mono overflow-hidden",
                                                            children: [
                                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                                    className: "text-muted-foreground whitespace-pre-wrap break-words",
                                                                    children: "config_args"
                                                                }, void 0, false, {
                                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                    lineNumber: 507,
                                                                    columnNumber: 27
                                                                }, this),
                                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                                    className: "whitespace-pre-wrap break-words",
                                                                    children: selectedStep.config_args ?? ""
                                                                }, void 0, false, {
                                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                    lineNumber: 508,
                                                                    columnNumber: 27
                                                                }, this)
                                                            ]
                                                        }, void 0, true, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 506,
                                                            columnNumber: 25
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border p-2 text-xs font-mono overflow-hidden",
                                                            children: [
                                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                                    className: "text-muted-foreground whitespace-pre-wrap break-words",
                                                                    children: "input_args"
                                                                }, void 0, false, {
                                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                    lineNumber: 511,
                                                                    columnNumber: 27
                                                                }, this),
                                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                                    className: "whitespace-pre-wrap break-words",
                                                                    children: selectedStep.input_args ?? ""
                                                                }, void 0, false, {
                                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                    lineNumber: 512,
                                                                    columnNumber: 27
                                                                }, this)
                                                            ]
                                                        }, void 0, true, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 510,
                                                            columnNumber: 25
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border p-2 text-xs font-mono overflow-hidden",
                                                            children: [
                                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                                    className: "text-muted-foreground whitespace-pre-wrap break-words",
                                                                    children: "output_args"
                                                                }, void 0, false, {
                                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                    lineNumber: 515,
                                                                    columnNumber: 27
                                                                }, this),
                                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                                    className: "whitespace-pre-wrap break-words",
                                                                    children: selectedStep.output_args ?? ""
                                                                }, void 0, false, {
                                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                    lineNumber: 516,
                                                                    columnNumber: 27
                                                                }, this)
                                                            ]
                                                        }, void 0, true, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 514,
                                                            columnNumber: 25
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 501,
                                                    columnNumber: 23
                                                }, this),
                                                bashResolvedCommand && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "Resolved Command"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 522,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border bg-muted/30 p-2",
                                                            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(CodeHighlighter, {
                                                                language: "bash",
                                                                style: theme === "dark" ? __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$atom$2d$one$2d$dark$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"] : __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$github$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"],
                                                                customStyle: {
                                                                    margin: 0,
                                                                    background: "transparent",
                                                                    fontSize: "0.75rem",
                                                                    whiteSpace: "pre-wrap",
                                                                    wordBreak: "break-word"
                                                                },
                                                                codeTagProps: {
                                                                    style: {
                                                                        whiteSpace: "pre-wrap",
                                                                        wordBreak: "break-word"
                                                                    }
                                                                },
                                                                children: bashResolvedCommand
                                                            }, void 0, false, {
                                                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                lineNumber: 524,
                                                                columnNumber: 29
                                                            }, this)
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 523,
                                                            columnNumber: 27
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 521,
                                                    columnNumber: 25
                                                }, this)
                                            ]
                                        }, void 0, true, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 499,
                                            columnNumber: 21
                                        }, this),
                                        selectedStep && (selectedStep.type === "bash" || selectedStep.type === "remote-bash" || selectedStep.type === "container") && Array.isArray(selectedStep.commands) && selectedStep.commands.length > 0 && renderStringList("Commands", selectedStep.commands, {
                                            language: "bash",
                                            copyAllText: selectedStep.commands.join("\n")
                                        }),
                                        selectedStep && (selectedStep.type === "bash" || selectedStep.type === "remote-bash" || selectedStep.type === "container") && Array.isArray(selectedStep.parallel_commands) && selectedStep.parallel_commands.length > 0 && renderStringList("Parallel Commands", selectedStep.parallel_commands, {
                                            language: "bash",
                                            copyAllText: selectedStep.parallel_commands.join("\n")
                                        }),
                                        selectedStep && selectedStep.type === "function" && selectedStep.function && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                            className: "space-y-2",
                                            children: [
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                    children: "Function"
                                                }, void 0, false, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 565,
                                                    columnNumber: 23
                                                }, this),
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "rounded-md border bg-muted/30 p-2",
                                                    children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(CodeHighlighter, {
                                                        language: "javascript",
                                                        style: theme === "dark" ? __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$atom$2d$one$2d$dark$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"] : __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$github$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"],
                                                        customStyle: {
                                                            margin: 0,
                                                            background: "transparent",
                                                            fontSize: "0.75rem",
                                                            whiteSpace: "pre-wrap",
                                                            wordBreak: "break-word"
                                                        },
                                                        wrapLongLines: true,
                                                        codeTagProps: {
                                                            style: {
                                                                whiteSpace: "pre-wrap",
                                                                wordBreak: "break-word"
                                                            }
                                                        },
                                                        children: selectedStep.function
                                                    }, void 0, false, {
                                                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                        lineNumber: 567,
                                                        columnNumber: 25
                                                    }, this)
                                                }, void 0, false, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 566,
                                                    columnNumber: 23
                                                }, this)
                                            ]
                                        }, void 0, true, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 564,
                                            columnNumber: 21
                                        }, this),
                                        selectedStep && selectedStep.type === "function" && Array.isArray(selectedStep.functions) && selectedStep.functions.length > 0 && renderStringList("Functions", selectedStep.functions, {
                                            language: "javascript",
                                            copyAllText: selectedStep.functions.join("\n")
                                        }),
                                        selectedStep && selectedStep.type === "function" && Array.isArray(selectedStep.parallel_functions) && selectedStep.parallel_functions.length > 0 && renderStringList("Parallel Functions", selectedStep.parallel_functions, {
                                            language: "javascript",
                                            copyAllText: selectedStep.parallel_functions.join("\n")
                                        }),
                                        selectedStep && selectedStep.type === "http" && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                            className: "space-y-4",
                                            children: [
                                                selectedStep.url && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "URL"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 609,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words",
                                                            children: selectedStep.url
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 610,
                                                            columnNumber: 27
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 608,
                                                    columnNumber: 25
                                                }, this),
                                                selectedStep.method && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "Method"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 618,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words",
                                                            children: selectedStep.method
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 619,
                                                            columnNumber: 27
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 617,
                                                    columnNumber: 25
                                                }, this),
                                                selectedStep.headers && Object.keys(selectedStep.headers).length > 0 && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "Headers"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 627,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border bg-muted/30 p-2",
                                                            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(CodeHighlighter, {
                                                                language: "json",
                                                                style: theme === "dark" ? __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$atom$2d$one$2d$dark$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"] : __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$github$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"],
                                                                customStyle: {
                                                                    margin: 0,
                                                                    background: "transparent",
                                                                    fontSize: "0.75rem",
                                                                    whiteSpace: "pre-wrap",
                                                                    wordBreak: "break-word"
                                                                },
                                                                codeTagProps: {
                                                                    style: {
                                                                        whiteSpace: "pre-wrap",
                                                                        wordBreak: "break-word"
                                                                    }
                                                                },
                                                                children: JSON.stringify(selectedStep.headers, null, 2)
                                                            }, void 0, false, {
                                                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                lineNumber: 629,
                                                                columnNumber: 29
                                                            }, this)
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 628,
                                                            columnNumber: 27
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 626,
                                                    columnNumber: 25
                                                }, this),
                                                selectedStep.request_body && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "Request Body"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 654,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border bg-muted/30 p-2",
                                                            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(CodeHighlighter, {
                                                                language: "json",
                                                                style: theme === "dark" ? __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$atom$2d$one$2d$dark$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"] : __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$github$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"],
                                                                customStyle: {
                                                                    margin: 0,
                                                                    background: "transparent",
                                                                    fontSize: "0.75rem",
                                                                    whiteSpace: "pre-wrap",
                                                                    wordBreak: "break-word"
                                                                },
                                                                codeTagProps: {
                                                                    style: {
                                                                        whiteSpace: "pre-wrap",
                                                                        wordBreak: "break-word"
                                                                    }
                                                                },
                                                                children: String(selectedStep.request_body)
                                                            }, void 0, false, {
                                                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                lineNumber: 656,
                                                                columnNumber: 29
                                                            }, this)
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 655,
                                                            columnNumber: 27
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 653,
                                                    columnNumber: 25
                                                }, this)
                                            ]
                                        }, void 0, true, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 606,
                                            columnNumber: 21
                                        }, this),
                                        selectedStep && selectedStep.type === "llm" && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                            className: "space-y-4",
                                            children: [
                                                typeof selectedStep.is_embedding === "boolean" && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "Embedding"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 685,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border bg-muted/30 p-2 font-mono text-xs",
                                                            children: String(selectedStep.is_embedding)
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 686,
                                                            columnNumber: 27
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 684,
                                                    columnNumber: 25
                                                }, this),
                                                Array.isArray(selectedStep.embedding_input) && selectedStep.embedding_input.length > 0 && renderStringList("Embedding Input", selectedStep.embedding_input.map((v)=>String(v)), {
                                                    copyAllText: selectedStep.embedding_input.join("\n")
                                                }),
                                                Array.isArray(selectedStep.messages) && selectedStep.messages.length > 0 && renderStringList("Messages", selectedStep.messages.map((m)=>JSON.stringify(m, null, 2)), {
                                                    language: "json",
                                                    copyAllText: JSON.stringify(selectedStep.messages, null, 2)
                                                }),
                                                Array.isArray(selectedStep.tools) && selectedStep.tools.length > 0 && renderStringList("Tools", selectedStep.tools.map((t)=>JSON.stringify(t, null, 2)), {
                                                    language: "json",
                                                    copyAllText: JSON.stringify(selectedStep.tools, null, 2)
                                                }),
                                                selectedStep.tool_choice !== undefined && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "Tool Choice"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 715,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border bg-muted/30 p-2",
                                                            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(CodeHighlighter, {
                                                                language: "json",
                                                                style: theme === "dark" ? __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$atom$2d$one$2d$dark$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"] : __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$github$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"],
                                                                customStyle: {
                                                                    margin: 0,
                                                                    background: "transparent",
                                                                    fontSize: "0.75rem",
                                                                    whiteSpace: "pre-wrap",
                                                                    wordBreak: "break-word"
                                                                },
                                                                codeTagProps: {
                                                                    style: {
                                                                        whiteSpace: "pre-wrap",
                                                                        wordBreak: "break-word"
                                                                    }
                                                                },
                                                                children: JSON.stringify(selectedStep.tool_choice, null, 2)
                                                            }, void 0, false, {
                                                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                lineNumber: 717,
                                                                columnNumber: 29
                                                            }, this)
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 716,
                                                            columnNumber: 27
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 714,
                                                    columnNumber: 25
                                                }, this),
                                                selectedStep.llm_config && Object.keys(selectedStep.llm_config).length > 0 && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "LLM Config"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 742,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border bg-muted/30 p-2",
                                                            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(CodeHighlighter, {
                                                                language: "json",
                                                                style: theme === "dark" ? __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$atom$2d$one$2d$dark$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"] : __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$github$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"],
                                                                customStyle: {
                                                                    margin: 0,
                                                                    background: "transparent",
                                                                    fontSize: "0.75rem",
                                                                    whiteSpace: "pre-wrap",
                                                                    wordBreak: "break-word"
                                                                },
                                                                codeTagProps: {
                                                                    style: {
                                                                        whiteSpace: "pre-wrap",
                                                                        wordBreak: "break-word"
                                                                    }
                                                                },
                                                                children: JSON.stringify(selectedStep.llm_config, null, 2)
                                                            }, void 0, false, {
                                                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                lineNumber: 744,
                                                                columnNumber: 29
                                                            }, this)
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 743,
                                                            columnNumber: 27
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 741,
                                                    columnNumber: 25
                                                }, this),
                                                selectedStep.extra_llm_parameters && Object.keys(selectedStep.extra_llm_parameters).length > 0 && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "Extra LLM Parameters"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 769,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border bg-muted/30 p-2",
                                                            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(CodeHighlighter, {
                                                                language: "json",
                                                                style: theme === "dark" ? __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$atom$2d$one$2d$dark$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"] : __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$github$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"],
                                                                customStyle: {
                                                                    margin: 0,
                                                                    background: "transparent",
                                                                    fontSize: "0.75rem",
                                                                    whiteSpace: "pre-wrap",
                                                                    wordBreak: "break-word"
                                                                },
                                                                codeTagProps: {
                                                                    style: {
                                                                        whiteSpace: "pre-wrap",
                                                                        wordBreak: "break-word"
                                                                    }
                                                                },
                                                                children: JSON.stringify(selectedStep.extra_llm_parameters, null, 2)
                                                            }, void 0, false, {
                                                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                lineNumber: 771,
                                                                columnNumber: 29
                                                            }, this)
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 770,
                                                            columnNumber: 27
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 768,
                                                    columnNumber: 25
                                                }, this)
                                            ]
                                        }, void 0, true, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 682,
                                            columnNumber: 21
                                        }, this),
                                        selectedStep && (selectedStep.step_runner || selectedStep.step_runner_config) && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                            className: "space-y-4",
                                            children: [
                                                selectedStep.step_runner && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "Runner"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 800,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words",
                                                            children: selectedStep.step_runner
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 801,
                                                            columnNumber: 27
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 799,
                                                    columnNumber: 25
                                                }, this),
                                                selectedStep.step_runner_config && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "Runner Config"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 808,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border bg-muted/30 p-2",
                                                            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(CodeHighlighter, {
                                                                language: "json",
                                                                style: theme === "dark" ? __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$atom$2d$one$2d$dark$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"] : __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$github$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"],
                                                                customStyle: {
                                                                    margin: 0,
                                                                    background: "transparent",
                                                                    fontSize: "0.75rem",
                                                                    whiteSpace: "pre-wrap",
                                                                    wordBreak: "break-word"
                                                                },
                                                                codeTagProps: {
                                                                    style: {
                                                                        whiteSpace: "pre-wrap",
                                                                        wordBreak: "break-word"
                                                                    }
                                                                },
                                                                children: JSON.stringify(selectedStep.step_runner_config, null, 2)
                                                            }, void 0, false, {
                                                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                lineNumber: 810,
                                                                columnNumber: 29
                                                            }, this)
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 809,
                                                            columnNumber: 27
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 807,
                                                    columnNumber: 25
                                                }, this)
                                            ]
                                        }, void 0, true, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 797,
                                            columnNumber: 21
                                        }, this),
                                        selectedStep && (selectedStep.type === "parallel" || selectedStep.type === "parallel-steps") && Array.isArray(selectedStep.parallel_steps) && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                            className: "space-y-2",
                                            children: [
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                    children: [
                                                        "Parallel Steps (",
                                                        selectedStep.parallel_steps.length,
                                                        ")"
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 837,
                                                    columnNumber: 23
                                                }, this),
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: selectedStep.parallel_steps.map((ps)=>{
                                                        let psYaml = "";
                                                        try {
                                                            psYaml = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$js$2d$yaml$2f$dist$2f$js$2d$yaml$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"].dump(ps, {
                                                                indent: 2,
                                                                lineWidth: -1,
                                                                noRefs: true,
                                                                quotingType: '"'
                                                            });
                                                        } catch  {
                                                            psYaml = "";
                                                        }
                                                        return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border p-3 space-y-2",
                                                            children: [
                                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                                    className: "flex items-center gap-2",
                                                                    children: [
                                                                        stepTypeIcons[ps.type],
                                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("span", {
                                                                            className: "font-medium text-sm",
                                                                            children: ps.name
                                                                        }, void 0, false, {
                                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                            lineNumber: 856,
                                                                            columnNumber: 33
                                                                        }, this),
                                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$badge$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Badge"], {
                                                                            variant: "secondary",
                                                                            className: "capitalize",
                                                                            children: ps.type
                                                                        }, void 0, false, {
                                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                            lineNumber: 857,
                                                                            columnNumber: 33
                                                                        }, this)
                                                                    ]
                                                                }, void 0, true, {
                                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                    lineNumber: 854,
                                                                    columnNumber: 31
                                                                }, this),
                                                                psYaml && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                                    className: "rounded-md border bg-muted/30 p-2",
                                                                    children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(CodeHighlighter, {
                                                                        language: "yaml",
                                                                        style: theme === "dark" ? __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$atom$2d$one$2d$dark$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"] : __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$github$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"],
                                                                        customStyle: {
                                                                            margin: 0,
                                                                            background: "transparent",
                                                                            fontSize: "0.75rem",
                                                                            whiteSpace: "pre-wrap",
                                                                            wordBreak: "break-word"
                                                                        },
                                                                        codeTagProps: {
                                                                            style: {
                                                                                whiteSpace: "pre-wrap",
                                                                                wordBreak: "break-word"
                                                                            }
                                                                        },
                                                                        showLineNumbers: true,
                                                                        children: psYaml.trim()
                                                                    }, void 0, false, {
                                                                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                        lineNumber: 864,
                                                                        columnNumber: 35
                                                                    }, this)
                                                                }, void 0, false, {
                                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                    lineNumber: 863,
                                                                    columnNumber: 33
                                                                }, this)
                                                            ]
                                                        }, ps.name, true, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 853,
                                                            columnNumber: 29
                                                        }, this);
                                                    })
                                                }, void 0, false, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 838,
                                                    columnNumber: 23
                                                }, this)
                                            ]
                                        }, void 0, true, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 836,
                                            columnNumber: 21
                                        }, this),
                                        selectedStep && selectedStep.type === "foreach" && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Fragment"], {
                                            children: [
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "Input File"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 896,
                                                            columnNumber: 25
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md bg-muted p-3 font-mono text-xs",
                                                            children: selectedStep.input
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 897,
                                                            columnNumber: 25
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 895,
                                                    columnNumber: 23
                                                }, this),
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "Variable Name"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 902,
                                                            columnNumber: 25
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md bg-muted p-3 font-mono text-xs",
                                                            children: selectedStep.variable
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 903,
                                                            columnNumber: 25
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 901,
                                                    columnNumber: 23
                                                }, this),
                                                typeof selectedStep.threads === "number" && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                            children: "Threads"
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 909,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md bg-muted p-3 font-mono text-xs",
                                                            children: selectedStep.threads
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 910,
                                                            columnNumber: 27
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 908,
                                                    columnNumber: 25
                                                }, this),
                                                selectedStep.step && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-2",
                                                    children: [
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "flex items-center justify-between gap-2",
                                                            children: [
                                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                                    children: "Foreach Step"
                                                                }, void 0, false, {
                                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                    lineNumber: 919,
                                                                    columnNumber: 29
                                                                }, this),
                                                                foreachStepYaml && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Button"], {
                                                                    className: "rounded-md",
                                                                    variant: "outline",
                                                                    size: "icon",
                                                                    onClick: async ()=>{
                                                                        try {
                                                                            await navigator.clipboard.writeText(foreachStepYaml);
                                                                            __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$sonner$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["toast"].success("Copied to clipboard");
                                                                        } catch  {
                                                                            __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$sonner$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["toast"].error("Failed to copy");
                                                                        }
                                                                    },
                                                                    children: [
                                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$clipboard$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__ClipboardIcon$3e$__["ClipboardIcon"], {
                                                                            className: "size-4"
                                                                        }, void 0, false, {
                                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                            lineNumber: 934,
                                                                            columnNumber: 33
                                                                        }, this),
                                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("span", {
                                                                            className: "sr-only",
                                                                            children: "Copy foreach step YAML"
                                                                        }, void 0, false, {
                                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                            lineNumber: 935,
                                                                            columnNumber: 33
                                                                        }, this)
                                                                    ]
                                                                }, void 0, true, {
                                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                    lineNumber: 921,
                                                                    columnNumber: 31
                                                                }, this)
                                                            ]
                                                        }, void 0, true, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 918,
                                                            columnNumber: 27
                                                        }, this),
                                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border p-2 text-sm",
                                                            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                                className: "flex items-center gap-2",
                                                                children: [
                                                                    stepTypeIcons[selectedStep.step.type],
                                                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("span", {
                                                                        className: "font-medium",
                                                                        children: selectedStep.step.name
                                                                    }, void 0, false, {
                                                                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                        lineNumber: 943,
                                                                        columnNumber: 31
                                                                    }, this),
                                                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$badge$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Badge"], {
                                                                        variant: "secondary",
                                                                        className: "capitalize",
                                                                        children: selectedStep.step.type
                                                                    }, void 0, false, {
                                                                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                        lineNumber: 944,
                                                                        columnNumber: 31
                                                                    }, this)
                                                                ]
                                                            }, void 0, true, {
                                                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                lineNumber: 941,
                                                                columnNumber: 29
                                                            }, this)
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 940,
                                                            columnNumber: 27
                                                        }, this),
                                                        foreachStepYaml && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "rounded-md border bg-muted/30 p-2",
                                                            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(CodeHighlighter, {
                                                                language: "yaml",
                                                                style: theme === "dark" ? __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$atom$2d$one$2d$dark$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"] : __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$github$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"],
                                                                customStyle: {
                                                                    margin: 0,
                                                                    background: "transparent",
                                                                    fontSize: "0.75rem",
                                                                    whiteSpace: "pre-wrap",
                                                                    wordBreak: "break-word"
                                                                },
                                                                codeTagProps: {
                                                                    style: {
                                                                        whiteSpace: "pre-wrap",
                                                                        wordBreak: "break-word"
                                                                    }
                                                                },
                                                                showLineNumbers: true,
                                                                children: foreachStepYaml.trim()
                                                            }, void 0, false, {
                                                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                lineNumber: 952,
                                                                columnNumber: 31
                                                            }, this)
                                                        }, void 0, false, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 951,
                                                            columnNumber: 29
                                                        }, this)
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 917,
                                                    columnNumber: 25
                                                }, this)
                                            ]
                                        }, void 0, true)
                                    ]
                                }, void 0, true, {
                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                    lineNumber: 323,
                                    columnNumber: 17
                                }, this),
                                selectedStep && selectedStep.timeout && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Fragment"], {
                                    children: [
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$separator$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Separator"], {}, void 0, false, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 983,
                                            columnNumber: 21
                                        }, this),
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                            className: "flex items-center gap-2 text-sm text-muted-foreground",
                                            children: [
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$clock$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__ClockIcon$3e$__["ClockIcon"], {
                                                    className: "size-4"
                                                }, void 0, false, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 985,
                                                    columnNumber: 23
                                                }, this),
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("span", {
                                                    children: [
                                                        "Timeout: ",
                                                        selectedStep.timeout,
                                                        "s"
                                                    ]
                                                }, void 0, true, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 986,
                                                    columnNumber: 23
                                                }, this)
                                            ]
                                        }, void 0, true, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 984,
                                            columnNumber: 21
                                        }, this)
                                    ]
                                }, void 0, true),
                                selectedStep && selectedStep.pre_condition && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Fragment"], {
                                    children: [
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$separator$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Separator"], {}, void 0, false, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 994,
                                            columnNumber: 21
                                        }, this),
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                            className: "space-y-2",
                                            children: [
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                    className: "text-muted-foreground",
                                                    children: "Pre-condition"
                                                }, void 0, false, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 996,
                                                    columnNumber: 23
                                                }, this),
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "rounded-md bg-muted p-3 font-mono text-xs",
                                                    children: selectedStep.pre_condition
                                                }, void 0, false, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 997,
                                                    columnNumber: 23
                                                }, this)
                                            ]
                                        }, void 0, true, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 995,
                                            columnNumber: 21
                                        }, this)
                                    ]
                                }, void 0, true),
                                selectedStep && selectedStep.exports && Object.keys(selectedStep.exports).length > 0 && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Fragment"], {
                                    children: [
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$separator$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Separator"], {}, void 0, false, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 1007,
                                            columnNumber: 21
                                        }, this),
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                            className: "space-y-2",
                                            children: [
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                    children: "Exports"
                                                }, void 0, false, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 1009,
                                                    columnNumber: 23
                                                }, this),
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "space-y-1",
                                                    children: Object.entries(selectedStep.exports).map(([key, value])=>/*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                            className: "flex items-center gap-2 text-xs font-mono",
                                                            children: [
                                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$badge$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Badge"], {
                                                                    variant: "outline",
                                                                    className: "font-mono",
                                                                    children: key
                                                                }, void 0, false, {
                                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                    lineNumber: 1016,
                                                                    columnNumber: 29
                                                                }, this),
                                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("span", {
                                                                    className: "text-muted-foreground truncate",
                                                                    children: [
                                                                        "= ",
                                                                        value
                                                                    ]
                                                                }, void 0, true, {
                                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                                    lineNumber: 1019,
                                                                    columnNumber: 29
                                                                }, this)
                                                            ]
                                                        }, key, true, {
                                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                            lineNumber: 1012,
                                                            columnNumber: 27
                                                        }, this))
                                                }, void 0, false, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 1010,
                                                    columnNumber: 23
                                                }, this)
                                            ]
                                        }, void 0, true, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 1008,
                                            columnNumber: 21
                                        }, this)
                                    ]
                                }, void 0, true),
                                selectedStep?.decision && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Fragment"], {
                                    children: [
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$separator$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Separator"], {}, void 0, false, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 1032,
                                            columnNumber: 21
                                        }, this),
                                        renderDecisionBlock(selectedStep.decision)
                                    ]
                                }, void 0, true),
                                selectedStep && selectedStep.log && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Fragment"], {
                                    children: [
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$separator$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Separator"], {}, void 0, false, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 1039,
                                            columnNumber: 21
                                        }, this),
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                            className: "space-y-2",
                                            children: [
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$label$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Label"], {
                                                    children: "Log"
                                                }, void 0, false, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 1041,
                                                    columnNumber: 23
                                                }, this),
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                    className: "rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words",
                                                    children: selectedStep.log
                                                }, void 0, false, {
                                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                    lineNumber: 1042,
                                                    columnNumber: 23
                                                }, this)
                                            ]
                                        }, void 0, true, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 1040,
                                            columnNumber: 21
                                        }, this)
                                    ]
                                }, void 0, true)
                            ]
                        }, void 0, true, {
                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                            lineNumber: 308,
                            columnNumber: 15
                        }, this) : /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                            className: "flex h-full items-center justify-center p-4",
                            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("p", {
                                className: "text-sm text-muted-foreground text-center",
                                children: "Select a node to view its properties"
                            }, void 0, false, {
                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                lineNumber: 1051,
                                columnNumber: 17
                            }, this)
                        }, void 0, false, {
                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                            lineNumber: 1050,
                            columnNumber: 15
                        }, this)
                    }, void 0, false, {
                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                        lineNumber: 306,
                        columnNumber: 11
                    }, this)
                }, void 0, false, {
                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                    lineNumber: 305,
                    columnNumber: 9
                }, this),
                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$tabs$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["TabsContent"], {
                    value: "items",
                    className: "flex-1 m-0 min-h-0",
                    children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$scroll$2d$area$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["ScrollArea"], {
                        className: "h-full",
                        children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                            className: "p-3 space-y-2",
                            children: workflowKind === "flow" ? /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Fragment"], {
                                children: allModules.length === 0 ? /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                    className: "p-4 text-sm text-muted-foreground text-center",
                                    children: "No modules"
                                }, void 0, false, {
                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                    lineNumber: 1065,
                                    columnNumber: 21
                                }, this) : /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                    className: "space-y-1",
                                    children: allModules.map((m)=>{
                                        const isSelected = selectionName === m.name;
                                        return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Button"], {
                                            type: "button",
                                            variant: isSelected ? "secondary" : "ghost",
                                            className: "w-full justify-start",
                                            onClick: ()=>{
                                                onNavigateToNode?.(m.name);
                                            },
                                            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                className: "flex w-full items-center justify-between gap-3",
                                                children: [
                                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("span", {
                                                        className: "truncate font-mono text-xs",
                                                        children: m.name
                                                    }, void 0, false, {
                                                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                        lineNumber: 1083,
                                                        columnNumber: 31
                                                    }, this),
                                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$badge$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Badge"], {
                                                        variant: "secondary",
                                                        className: "capitalize",
                                                        children: "module"
                                                    }, void 0, false, {
                                                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                        lineNumber: 1084,
                                                        columnNumber: 31
                                                    }, this)
                                                ]
                                            }, void 0, true, {
                                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                lineNumber: 1082,
                                                columnNumber: 29
                                            }, this)
                                        }, m.name, false, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 1073,
                                            columnNumber: 27
                                        }, this);
                                    })
                                }, void 0, false, {
                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                    lineNumber: 1069,
                                    columnNumber: 21
                                }, this)
                            }, void 0, false) : /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Fragment"], {
                                children: allSteps.length === 0 ? /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                    className: "p-4 text-sm text-muted-foreground text-center",
                                    children: "No steps"
                                }, void 0, false, {
                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                    lineNumber: 1097,
                                    columnNumber: 21
                                }, this) : /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                    className: "space-y-1",
                                    children: allSteps.map((s)=>{
                                        const isSelected = selectionName === s.name;
                                        return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Button"], {
                                            type: "button",
                                            variant: isSelected ? "secondary" : "ghost",
                                            className: "w-full justify-start",
                                            onClick: ()=>{
                                                onNavigateToNode?.(s.name);
                                            },
                                            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                                className: "flex w-full items-center justify-between gap-3",
                                                children: [
                                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("span", {
                                                        className: "truncate font-mono text-xs",
                                                        children: s.name
                                                    }, void 0, false, {
                                                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                        lineNumber: 1115,
                                                        columnNumber: 31
                                                    }, this),
                                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$badge$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Badge"], {
                                                        variant: badgeVariantForStepType(s.type),
                                                        className: "capitalize",
                                                        children: s.type
                                                    }, void 0, false, {
                                                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                        lineNumber: 1116,
                                                        columnNumber: 31
                                                    }, this)
                                                ]
                                            }, void 0, true, {
                                                fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                                lineNumber: 1114,
                                                columnNumber: 29
                                            }, this)
                                        }, s.name, false, {
                                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                            lineNumber: 1105,
                                            columnNumber: 27
                                        }, this);
                                    })
                                }, void 0, false, {
                                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                                    lineNumber: 1101,
                                    columnNumber: 21
                                }, this)
                            }, void 0, false)
                        }, void 0, false, {
                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                            lineNumber: 1061,
                            columnNumber: 13
                        }, this)
                    }, void 0, false, {
                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                        lineNumber: 1060,
                        columnNumber: 11
                    }, this)
                }, void 0, false, {
                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                    lineNumber: 1059,
                    columnNumber: 9
                }, this),
                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$tabs$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["TabsContent"], {
                    value: "yaml",
                    className: "flex-1 m-0 min-h-0",
                    children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                        className: "h-full min-h-0 overflow-y-auto p-3",
                        children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(CodeHighlighter, {
                            language: "yaml",
                            style: theme === "dark" ? __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$atom$2d$one$2d$dark$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"] : __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$react$2d$syntax$2d$highlighter$2f$dist$2f$esm$2f$styles$2f$hljs$2f$github$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"],
                            customStyle: {
                                margin: 0,
                                background: "transparent",
                                fontSize: "0.75rem",
                                maxHeight: "100%",
                                overflowY: "auto",
                                whiteSpace: wrapLongText ? "pre-wrap" : "pre",
                                wordBreak: wrapLongText ? "break-word" : "normal"
                            },
                            codeTagProps: {
                                style: {
                                    whiteSpace: wrapLongText ? "pre-wrap" : "pre",
                                    wordBreak: wrapLongText ? "break-word" : "normal"
                                }
                            },
                            showLineNumbers: true,
                            children: yamlPreview
                        }, void 0, false, {
                            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                            lineNumber: 1133,
                            columnNumber: 13
                        }, this)
                    }, void 0, false, {
                        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                        lineNumber: 1132,
                        columnNumber: 11
                    }, this)
                }, void 0, false, {
                    fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
                    lineNumber: 1131,
                    columnNumber: 9
                }, this)
            ]
        }, void 0, true, {
            fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
            lineNumber: 290,
            columnNumber: 7
        }, this)
    }, void 0, false, {
        fileName: "[project]/components/workflow-editor/workflow-sidebar.tsx",
        lineNumber: 289,
        columnNumber: 5
    }, this);
}
}),
"[project]/components/workflow-editor/utils/yaml-parser.ts [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "parseWorkflowYaml",
    ()=>parseWorkflowYaml,
    "serializeWorkflowToYaml",
    ()=>serializeWorkflowToYaml,
    "updateStepInWorkflow",
    ()=>updateStepInWorkflow
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$js$2d$yaml$2f$dist$2f$js$2d$yaml$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/js-yaml/dist/js-yaml.mjs [app-ssr] (ecmascript)");
;
function isSwitchDecision(decision) {
    if (!decision || typeof decision !== "object") return false;
    const d = decision;
    return typeof d.switch === "string" && d.cases && typeof d.cases === "object" && !Array.isArray(d.cases);
}
function normalizeDecisionEdges(decision) {
    if (Array.isArray(decision)) {
        return decision.map((r)=>({
                condition: typeof r?.condition === "string" ? r.condition : "",
                next: typeof r?.next === "string" ? r.next : ""
            })).filter((r)=>r.next.trim().length > 0).map((r)=>({
                label: r.condition,
                next: r.next,
                kind: "rule"
            }));
    }
    if (isSwitchDecision(decision)) {
        const d = decision;
        const cases = d.cases ?? {};
        const edges = Object.entries(cases).filter(([key])=>typeof key === "string" && key.trim().length > 0).map(([key, value])=>{
            const goto = typeof value?.goto === "string" ? value.goto : typeof value?.next === "string" ? value.next : "";
            return {
                label: key,
                next: goto,
                kind: "case"
            };
        }).filter((e)=>e.next.trim().length > 0);
        const defGoto = typeof d.default?.goto === "string" ? d.default.goto : typeof d.default?.next === "string" ? d.default.next : "";
        if (typeof defGoto === "string" && defGoto.trim().length > 0) {
            edges.push({
                label: "default",
                next: defGoto,
                kind: "default"
            });
        }
        return edges;
    }
    return [];
}
function truncateLabel(label, maxLen = 30) {
    const trimmed = label.trim();
    if (trimmed.length <= maxLen) return trimmed;
    return trimmed.substring(0, maxLen) + "...";
}
function parseWorkflowYaml(yamlText) {
    const workflowAny = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$js$2d$yaml$2f$dist$2f$js$2d$yaml$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"].load(yamlText) ?? {};
    const kind = workflowAny?.kind === "flow" ? "flow" : "module";
    const workflow = workflowAny;
    const nodes = [];
    const edges = [];
    // Add start node
    nodes.push({
        id: "_start",
        type: "start",
        position: {
            x: 0,
            y: 0
        },
        data: {
            label: "Start",
            step: null,
            module: null
        }
    });
    if (kind === "module") {
        const wf = workflow;
        const steps = Array.isArray(wf?.steps) ? wf.steps : [];
        const stepIds = new Set(steps.map((s)=>typeof s?.name === "string" ? s.name : "").filter(Boolean));
        const missingNodeIds = new Set();
        steps.forEach((step, index)=>{
            const nodeId = step.name;
            nodes.push({
                id: nodeId,
                type: step.type,
                position: {
                    x: 0,
                    y: 0
                },
                data: {
                    label: step.name,
                    step,
                    module: null
                }
            });
            const prevNodeId = index === 0 ? "_start" : steps[index - 1].name;
            const prevStep = index > 0 ? steps[index - 1] : null;
            const prevHasSwitch = isSwitchDecision(prevStep?.decision);
            const prevDecisionEdges = prevStep ? normalizeDecisionEdges(prevStep.decision) : [];
            const hasDecisionToThis = prevDecisionEdges.some((d)=>d.next === nodeId);
            if (!prevHasSwitch && !hasDecisionToThis) {
                edges.push({
                    id: `${prevNodeId}->${nodeId}`,
                    source: prevNodeId,
                    target: nodeId,
                    type: "smoothstep",
                    animated: step.type === "parallel" || step.type === "parallel-steps"
                });
            }
            const decisionEdges = normalizeDecisionEdges(step.decision);
            if (decisionEdges.length > 0) {
                decisionEdges.forEach((rule)=>{
                    const next = rule.next;
                    if (next !== "_end" && !stepIds.has(next)) missingNodeIds.add(next);
                    edges.push({
                        id: `${nodeId}->${next}:${rule.kind}:${rule.label}`,
                        source: nodeId,
                        target: next,
                        type: "smoothstep",
                        label: truncateLabel(rule.label),
                        style: {
                            strokeDasharray: "5 5"
                        }
                    });
                });
            }
        });
        for (const missingId of missingNodeIds){
            if (stepIds.has(missingId)) continue;
            nodes.push({
                id: missingId,
                position: {
                    x: 0,
                    y: 0
                },
                data: {
                    label: `Missing: ${missingId}`,
                    step: null,
                    module: null
                }
            });
        }
        if (steps.length > 0) {
            const lastStep = steps[steps.length - 1];
            const lastHasSwitch = isSwitchDecision(lastStep?.decision);
            const hasAnyEndEdge = edges.some((e)=>e.source === lastStep.name && e.target === "_end");
            if (!lastHasSwitch && !hasAnyEndEdge) {
                edges.push({
                    id: `${lastStep.name}->_end`,
                    source: lastStep.name,
                    target: "_end",
                    type: "smoothstep"
                });
            }
        }
    } else {
        const wf = workflow;
        const modules = Array.isArray(wf?.modules) ? wf.modules : [];
        modules.forEach((m)=>{
            nodes.push({
                id: m.name,
                type: "module",
                position: {
                    x: 0,
                    y: 0
                },
                data: {
                    label: m.name,
                    step: null,
                    module: m
                }
            });
        });
        const nodeIds = new Set(modules.map((m)=>m.name));
        const missingNodeIds = new Set();
        const outDegree = new Map();
        modules.forEach((m)=>{
            const deps = Array.isArray(m.depends_on) ? m.depends_on.filter((d)=>typeof d === "string") : [];
            if (deps.length > 0) {
                deps.forEach((d)=>{
                    if (!nodeIds.has(d)) return;
                    edges.push({
                        id: `${d}->${m.name}`,
                        source: d,
                        target: m.name,
                        type: "smoothstep"
                    });
                    outDegree.set(d, (outDegree.get(d) || 0) + 1);
                });
            } else {
                edges.push({
                    id: `_start->${m.name}`,
                    source: "_start",
                    target: m.name,
                    type: "smoothstep"
                });
            }
            const decisionEdges = normalizeDecisionEdges(m.decision);
            if (decisionEdges.length > 0) {
                decisionEdges.forEach((rule)=>{
                    const next = rule.next;
                    if (next !== "_end" && !nodeIds.has(next)) missingNodeIds.add(next);
                    edges.push({
                        id: `${m.name}->${next}:${rule.kind}:${rule.label}`,
                        source: m.name,
                        target: next,
                        type: "smoothstep",
                        label: truncateLabel(rule.label),
                        style: {
                            strokeDasharray: "5 5"
                        }
                    });
                    outDegree.set(m.name, (outDegree.get(m.name) || 0) + 1);
                });
            }
        });
        for (const missingId of missingNodeIds){
            if (nodeIds.has(missingId)) continue;
            nodes.push({
                id: missingId,
                position: {
                    x: 0,
                    y: 0
                },
                data: {
                    label: `Missing: ${missingId}`,
                    step: null,
                    module: null
                }
            });
        }
        if (modules.length === 0) {
            edges.push({
                id: `_start->_end`,
                source: "_start",
                target: "_end",
                type: "smoothstep"
            });
        } else {
            modules.forEach((m)=>{
                const od = outDegree.get(m.name) || 0;
                if (od === 0) {
                    edges.push({
                        id: `${m.name}->_end`,
                        source: m.name,
                        target: "_end",
                        type: "smoothstep"
                    });
                }
            });
        }
    }
    // Add end node
    nodes.push({
        id: "_end",
        type: "end",
        position: {
            x: 0,
            y: 0
        },
        data: {
            label: "End",
            step: null,
            module: null
        }
    });
    return {
        nodes,
        edges,
        metadata: {
            name: workflowAny?.name || "",
            description: workflowAny?.description || "",
            kind
        },
        raw: workflow
    };
}
function serializeWorkflowToYaml(workflow) {
    return __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$js$2d$yaml$2f$dist$2f$js$2d$yaml$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"].dump(workflow, {
        indent: 2,
        lineWidth: -1,
        noRefs: true,
        quotingType: '"'
    });
}
function updateStepInWorkflow(workflow, stepName, updates) {
    return {
        ...workflow,
        steps: workflow.steps.map((step)=>step.name === stepName ? {
                ...step,
                ...updates
            } : step)
    };
}
}),
"[project]/components/shared/error-state.tsx [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "ErrorState",
    ()=>ErrorState
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react-jsx-dev-runtime.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/utils.ts [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/ui/button.tsx [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$circle$2d$alert$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__AlertCircleIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/circle-alert.js [app-ssr] (ecmascript) <export default as AlertCircleIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$refresh$2d$cw$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__RefreshCwIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/refresh-cw.js [app-ssr] (ecmascript) <export default as RefreshCwIcon>");
;
;
;
;
function ErrorState({ title = "Something went wrong", message = "We couldn't load the data. Please try again.", onRetry, className }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
        className: (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$utils$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["cn"])("flex flex-col items-center justify-center py-12 text-center", className),
        children: [
            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                className: "mb-4 rounded-full bg-destructive/10 p-4",
                children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$circle$2d$alert$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__AlertCircleIcon$3e$__["AlertCircleIcon"], {
                    className: "size-8 text-destructive"
                }, void 0, false, {
                    fileName: "[project]/components/shared/error-state.tsx",
                    lineNumber: 27,
                    columnNumber: 9
                }, this)
            }, void 0, false, {
                fileName: "[project]/components/shared/error-state.tsx",
                lineNumber: 26,
                columnNumber: 7
            }, this),
            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("h3", {
                className: "mb-1 text-lg font-semibold",
                children: title
            }, void 0, false, {
                fileName: "[project]/components/shared/error-state.tsx",
                lineNumber: 29,
                columnNumber: 7
            }, this),
            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("p", {
                className: "mb-4 max-w-sm text-sm text-muted-foreground",
                children: message
            }, void 0, false, {
                fileName: "[project]/components/shared/error-state.tsx",
                lineNumber: 30,
                columnNumber: 7
            }, this),
            onRetry && /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Button"], {
                onClick: onRetry,
                variant: "outline",
                children: [
                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$refresh$2d$cw$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__RefreshCwIcon$3e$__["RefreshCwIcon"], {
                        className: "mr-2 size-4"
                    }, void 0, false, {
                        fileName: "[project]/components/shared/error-state.tsx",
                        lineNumber: 33,
                        columnNumber: 11
                    }, this),
                    "Try again"
                ]
            }, void 0, true, {
                fileName: "[project]/components/shared/error-state.tsx",
                lineNumber: 32,
                columnNumber: 9
            }, this)
        ]
    }, void 0, true, {
        fileName: "[project]/components/shared/error-state.tsx",
        lineNumber: 20,
        columnNumber: 5
    }, this);
}
}),
"[project]/lib/mock/workflow-yamls.ts [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "MOCK_WORKFLOW_YAMLS",
    ()=>MOCK_WORKFLOW_YAMLS
]);
const MOCK_WORKFLOW_YAMLS = {
    "test-complex-docker-workflow": `name: test-complex-docker-workflow
kind: module
description: Complex workflow demonstrating bash, function steps with docker step_runner

params:
  - name: target
    required: true
  - name: output_dir
    default: /tmp/osm-complex-test
  - name: threads
    default: "5"

steps:
  # Step 1: Setup - Create directories using function
  - name: setup-workspace
    type: function
    log: "Setting up workspace for {{target}}"
    function: createDir("{{output_dir}}")
    exports:
      workspace_created: "output"

  # Step 2: Create input file with bash
  - name: create-target-list
    type: bash
    log: "Creating target list for {{target}}"
    commands:
      - mkdir -p {{output_dir}}/targets
      - |
        cat > {{output_dir}}/targets/hosts.txt << 'EOF'
        sub1.{{target}}
        sub2.{{target}}
        api.{{target}}
        www.{{target}}
        admin.{{target}}
        EOF
    exports:
      target_file: "{{output_dir}}/targets/hosts.txt"

  # Step 3: Docker-based DNS resolution simulation
  - name: dns-resolve
    type: remote-bash
    log: "Resolving DNS for targets in Docker"
    timeout: 60
    step_runner: docker
    step_runner_config:
      image: alpine:latest
      env:
        TARGET_DOMAIN: "{{target}}"
      volumes:
        - "{{output_dir}}:/workspace"
      workdir: /workspace
    command: |
      echo "Resolving DNS for $TARGET_DOMAIN"
      cat /workspace/targets/hosts.txt | while read host; do
        echo "$host -> 127.0.0.1" >> /workspace/dns-resolved.txt
      done
      echo "DNS resolution complete"
    exports:
      dns_output: "{{output_dir}}/dns-resolved.txt"

  # Step 4: Parallel docker commands - simulating port scanning
  - name: parallel-port-scan
    type: remote-bash
    log: "Running parallel port scans in Docker"
    timeout: 120
    step_runner: docker
    step_runner_config:
      image: alpine:latest
      volumes:
        - "{{output_dir}}:/workspace"
    parallel_commands:
      - 'echo "Scanning ports 1-1000 on {{target}}" && sleep 1 && echo "Port 80 open" > /workspace/ports-1.txt'
      - 'echo "Scanning ports 1001-2000 on {{target}}" && sleep 1 && echo "Port 443 open" > /workspace/ports-2.txt'
      - 'echo "Scanning ports 2001-3000 on {{target}}" && sleep 1 && echo "Port 8080 open" > /workspace/ports-3.txt'
      - 'echo "Scanning ports 3001-4000 on {{target}}" && sleep 1 && echo "Port 3306 open" > /workspace/ports-4.txt'

  # Step 5: Merge port scan results
  - name: merge-port-results
    type: bash
    log: "Merging port scan results"
    command: cat {{output_dir}}/ports-*.txt > {{output_dir}}/all-ports.txt
    exports:
      ports_file: "{{output_dir}}/all-ports.txt"

  # Step 6: Function to check file existence
  - name: verify-ports-file
    type: function
    log: "Verifying ports file exists"
    function: fileExists("{{ports_file}}")
    exports:
      ports_verified: "output"

  # Step 7: Docker-based HTTP probing with parallel steps
  - name: http-probe-parallel
    type: parallel-steps
    log: "Running parallel HTTP probes"
    parallel_steps:
      - name: probe-http
        type: remote-bash
        step_runner: docker
        step_runner_config:
          image: alpine:latest
          volumes:
            - "{{output_dir}}:/workspace"
        command: |
          echo "Probing HTTP on port 80"
          echo "http://{{target}}:80 [200]" > /workspace/http-80.txt
      - name: probe-https
        type: remote-bash
        step_runner: docker
        step_runner_config:
          image: alpine:latest
          volumes:
            - "{{output_dir}}:/workspace"
        command: |
          echo "Probing HTTPS on port 443"
          echo "https://{{target}}:443 [200]" > /workspace/https-443.txt
      - name: probe-alt
        type: remote-bash
        step_runner: docker
        step_runner_config:
          image: alpine:latest
          volumes:
            - "{{output_dir}}:/workspace"
        command: |
          echo "Probing alternate port 8080"
          echo "http://{{target}}:8080 [404]" > /workspace/http-8080.txt

  # Step 8: Foreach loop with docker - process each subdomain
  - name: process-subdomains
    type: foreach
    log: "Processing each subdomain"
    input: "{{output_dir}}/targets/hosts.txt"
    variable: subdomain
    threads: 3
    step:
      name: scan-subdomain
      type: remote-bash
      step_runner: docker
      step_runner_config:
        image: alpine:latest
        volumes:
          - "{{output_dir}}:/workspace"
      command: |
        echo "Scanning [[subdomain]]..."
        echo "[[subdomain]]: status=200, title=Example" >> /workspace/subdomain-results.txt

  # Step 9: Read results with function
  - name: read-subdomain-results
    type: function
    log: "Reading subdomain scan results"
    function: readFile("{{output_dir}}/subdomain-results.txt")
    exports:
      scan_results: "output"

  # Step 10: Decision based routing
  - name: check-results
    type: bash
    log: "Checking scan results"
    command: wc -l < {{output_dir}}/subdomain-results.txt
    exports:
      result_count: "output"
    decision:
      - condition: result_count == "0"
        next: "_end"
      - condition: result_count != "0"
        next: "generate-report"

  # Step 11: Generate final report in docker
  - name: generate-report
    type: remote-bash
    log: "Generating final report"
    timeout: 30
    step_runner: docker
    step_runner_config:
      image: alpine:latest
      volumes:
        - "{{output_dir}}:/workspace"
    commands:
      - echo "=== Scan Report for {{target}} ===" > /workspace/report.txt
      - echo "" >> /workspace/report.txt
      - echo "--- DNS Results ---" >> /workspace/report.txt
      - cat /workspace/dns-resolved.txt >> /workspace/report.txt 2>/dev/null || echo "No DNS results" >> /workspace/report.txt
      - echo "" >> /workspace/report.txt
      - echo "--- Open Ports ---" >> /workspace/report.txt
      - cat /workspace/all-ports.txt >> /workspace/report.txt 2>/dev/null || echo "No ports found" >> /workspace/report.txt
      - echo "" >> /workspace/report.txt
      - echo "--- Subdomain Results ---" >> /workspace/report.txt
      - cat /workspace/subdomain-results.txt >> /workspace/report.txt 2>/dev/null || echo "No subdomain results" >> /workspace/report.txt
      - echo "" >> /workspace/report.txt
      - echo "Report generated at $(date)" >> /workspace/report.txt
    exports:
      report_file: "{{output_dir}}/report.txt"

  # Step 12: Parallel functions to get file stats
  - name: get-file-stats
    type: function
    log: "Getting file statistics"
    parallel_functions:
      - fileLength("{{output_dir}}/report.txt")
      - fileExists("{{output_dir}}/all-ports.txt")
      - trim("  {{target}}  ")
    exports:
      file_stats: "output"

  # Step 13: Cleanup (optional - controlled by pre_condition)
  - name: cleanup-temp-files
    type: bash
    log: "Cleaning up temporary files"
    pre_condition: "false"
    command: rm -rf {{output_dir}}/ports-*.txt
    on_error:
      - action: log
        message: "Cleanup failed but continuing"
      - action: continue
`,
    "test-decision": `name: test-decision
kind: module
description: Test conditional step routing with decision

params:
  - name: target
    required: true

steps:
  - name: check-condition
    type: bash
    command: echo "{{target}}"
    exports:
      target_value: "output"
    decision:
      - condition: target_value == "skip"
        next: "_end"
      - condition: target_value == "jump"
        next: "final-step"

  - name: middle-step
    type: bash
    command: echo "middle executed"
    exports:
      middle_output: "output"

  - name: final-step
    type: bash
    command: echo "final executed"
    exports:
      final_output: "output"
`,
    "test-docker-flow": `name: test-docker-flow
kind: flow
description: Flow orchestrating multiple Docker-based security scanning modules

params:
  - name: target
    required: true
  - name: Output
    default: /tmp/osm-docker-flow
  - name: mode
    default: "full"
  - name: threads
    default: "10"
  - name: skip_vuln_scan
    default: "false"

modules:
  # Module 1: Initial reconnaissance
  - name: recon-module
    path: modules/test-docker-recon
    params:
      target: "{{target}}"
      output_dir: "{{Output}}/recon"
      threads: "{{threads}}"
    on_success:
      - action: log
        message: "Reconnaissance completed for {{target}}"
      - action: export
        key: recon_complete
        value: "true"
    on_error:
      - action: log
        message: "Reconnaissance failed for {{target}}"
      - action: abort

  # Module 2: Subdomain enumeration (depends on recon)
  - name: subdomain-module
    path: modules/test-docker-subdomain
    depends_on:
      - recon-module
    params:
      target: "{{target}}"
      output_dir: "{{Output}}/subdomains"
      wordlist: "/usr/share/wordlists/subdomains.txt"
    condition: "mode == 'full' || mode == 'subdomain'"
    on_success:
      - action: export
        key: subdomains_file
        value: "{{Output}}/subdomains/all.txt"

  # Module 3: Port scanning (parallel with subdomain)
  - name: portscan-module
    path: modules/test-docker-portscan
    depends_on:
      - recon-module
    params:
      target: "{{target}}"
      output_dir: "{{Output}}/ports"
      port_range: "1-10000"
      rate: "1000"
    condition: "mode == 'full' || mode == 'portscan'"

  # Module 4: HTTP probing (depends on subdomain results)
  - name: httpx-module
    path: modules/test-docker-httpx
    depends_on:
      - subdomain-module
    params:
      input: "{{subdomains_file}}"
      output_dir: "{{Output}}/http"
      threads: "{{threads}}"
    on_success:
      - action: export
        key: alive_hosts
        value: "{{Output}}/http/alive.txt"
      - action: export
        key: httpx_json
        value: "{{Output}}/http/httpx.json"
    decision:
      - condition: "fileLength('{{Output}}/http/alive.txt') == 0"
        next: "report-module"

  # Module 5: Technology detection (depends on HTTP probe)
  - name: tech-detect-module
    path: modules/test-docker-techdetect
    depends_on:
      - httpx-module
    params:
      input: "{{alive_hosts}}"
      output_dir: "{{Output}}/tech"

  # Module 6: Screenshot capture (parallel with tech detection)
  - name: screenshot-module
    path: modules/test-docker-screenshot
    depends_on:
      - httpx-module
    params:
      input: "{{alive_hosts}}"
      output_dir: "{{Output}}/screenshots"
      threads: "5"

  # Module 7: Vulnerability scanning (conditional)
  - name: vulnscan-module
    path: modules/test-docker-scanning
    depends_on:
      - httpx-module
      - tech-detect-module
    params:
      target: "{{target}}"
      Output: "{{Output}}/vulns"
      severity: "critical,high,medium"
      threads: "{{threads}}"
    condition: "skip_vuln_scan != 'true'"
    on_error:
      - action: log
        message: "Vulnerability scan encountered errors but continuing"
      - action: continue

  # Module 8: Directory bruteforcing (optional - depends on mode)
  - name: dirbrute-module
    path: modules/test-docker-dirbrute
    depends_on:
      - httpx-module
    params:
      input: "{{alive_hosts}}"
      output_dir: "{{Output}}/dirs"
      wordlist: "/usr/share/wordlists/common.txt"
      threads: "20"
    condition: "mode == 'full'"

  # Module 9: JavaScript analysis (depends on dir results)
  - name: js-analysis-module
    path: modules/test-docker-jsanalysis
    depends_on:
      - dirbrute-module
    params:
      input: "{{Output}}/dirs/js-files.txt"
      output_dir: "{{Output}}/js"
    condition: "mode == 'full'"

  # Module 10: Final report generation
  - name: report-module
    path: modules/test-docker-report
    depends_on:
      - screenshot-module
      - vulnscan-module
      - tech-detect-module
    params:
      target: "{{target}}"
      input_dir: "{{Output}}"
      output_dir: "{{Output}}/reports"
      format: "html,json,markdown"
    on_success:
      - action: log
        message: "Flow completed successfully for {{target}}"
      - action: notify
        message: "Security assessment complete: {{target}}"
`,
    "test-loop": `name: test-loop
kind: module
description: Test foreach loop with threading

params:
  - name: target
    required: true

steps:
  - name: create-input
    type: bash
    commands:
      - mkdir -p {{Output}}
      - printf 'one\\ntwo\\nthree\\nfour\\nfive\\n' > {{Output}}/items.txt

  - name: process-items
    type: foreach
    input: "{{Output}}/items.txt"
    variable: item
    threads: 2
    step:
      name: process-item
      type: bash
      command: echo "Processing [[item]] for {{target}}"
`,
    "comprehensive-flow-example": `# =============================================================================
# Flow Workflow: Comprehensive Example
# =============================================================================
# This file demonstrates ALL fields available in a flow-kind workflow.
# Flows orchestrate multiple modules with dependencies, conditions, and routing.
# =============================================================================

# -----------------------------------------------------------------------------
# WORKFLOW-LEVEL FIELDS
# Same as module workflows (kind, name, description, tags, params, etc.)
# -----------------------------------------------------------------------------

# kind: Workflow type - "flow" orchestrates multiple modules
kind: flow

# name: Unique identifier for this workflow (required)
name: comprehensive-flow-example

# description: Human-readable description
description: Demonstrates all flow-specific fields including modules, dependencies, conditions, and decisions

# tags: Comma-separated tags for filtering
tags: flow, comprehensive, example

# -----------------------------------------------------------------------------
# PARAMS SECTION
# Parameters available to all modules in this flow
# -----------------------------------------------------------------------------
params:
  - name: threads
    default: "10"

  - name: timeout
    default: "3600"

  - name: scan_depth
    default: "normal"

  - name: output_format
    default: "json"

# -----------------------------------------------------------------------------
# DEPENDENCIES SECTION
# Flow-level dependencies checked before any module executes
# -----------------------------------------------------------------------------
dependencies:
  commands:
    - nmap
    - nuclei
    - httpx

  files:
    - /tmp

  variables:
    - name: Target
      type: domain
      required: true

# -----------------------------------------------------------------------------
# REPORTS SECTION
# Reports aggregated from all modules in this flow
# -----------------------------------------------------------------------------
reports:
  - name: flow-summary
    path: "{{Output}}/flow-summary.json"
    type: json
    description: Aggregated results from all modules

  - name: vulnerabilities
    path: "{{Output}}/vulnerabilities.txt"
    type: text
    description: All discovered vulnerabilities

# -----------------------------------------------------------------------------
# PREFERENCES SECTION
# Flow-level preferences apply to all module executions
# -----------------------------------------------------------------------------
preferences:
  skip_workspace: false
  disable_notifications: false
  heuristics_check: 'basic'

# -----------------------------------------------------------------------------
# MODULES SECTION (Flow-specific)
# Ordered list of module references to execute
# =============================================================================
modules:
  # ===========================================================================
  # Module Reference: Basic Configuration
  # ===========================================================================
  - # name: Display name for this module execution (required)
    name: reconnaissance

    # path: Path to the module YAML file (required)
    # Can be relative to workflows directory or absolute
    path: modules/recon.yaml

    # params: Parameters to pass to this module
    # Overrides module defaults and flow-level params
    params:
      threads: "20"  # Override flow-level threads
      output_dir: "{{Output}}/recon"

  # ===========================================================================
  # Module Reference: With Dependencies (depends_on)
  # ===========================================================================
  - name: port-scanning
    path: modules/portscan.yaml

    # depends_on: List of module names that must complete before this module runs
    # Creates a DAG (Directed Acyclic Graph) for execution order
    depends_on:
      - reconnaissance

    params:
      target_list: "{{Output}}/recon/subdomains.txt"
      threads: "{{threads}}"

  # ===========================================================================
  # Module Reference: With Condition
  # ===========================================================================
  - name: web-scanning
    path: modules/webscan.yaml

    depends_on:
      - port-scanning

    # condition: JavaScript expression - module only runs if evaluates to true
    # Can reference exported variables from previous modules
    condition: 'fileLength("{{Output}}/portscan/http-services.txt") > 0'

    params:
      input: "{{Output}}/portscan/http-services.txt"

  # ===========================================================================
  # Module Reference: With on_success Handler
  # ===========================================================================
  - name: vulnerability-scanning
    path: modules/vuln-scan.yaml

    depends_on:
      - web-scanning

    condition: 'fileExists("{{Output}}/webscan/endpoints.txt")'

    params:
      endpoints: "{{Output}}/webscan/endpoints.txt"
      timeout: "{{timeout}}"

    # on_success: Actions to execute when this module completes successfully
    on_success:
      # action: log - Log a message
      - action: log
        message: "Vulnerability scanning completed for {{Target}}"

      # action: export - Export a variable for subsequent modules
      - action: export
        name: vuln_scan_complete
        value: "true"

      # action: notify - Send a notification
      - action: notify
        notify: "Vulnerability scan finished for {{Target}}"

      # action: run - Execute a follow-up step
      - action: run
        type: bash
        command: 'echo "Vuln scan done" >> {{Output}}/flow-log.txt'

      # action: run with functions
      - action: run
        type: function
        functions:
          - 'log_info("Module completed successfully")'

  # ===========================================================================
  # Module Reference: With on_error Handler
  # ===========================================================================
  - name: exploit-verification
    path: modules/exploit-verify.yaml

    depends_on:
      - vulnerability-scanning

    condition: '{{vuln_scan_complete}} == "true"'

    params:
      vulns_file: "{{Output}}/vuln-scan/vulnerabilities.json"

    # on_error: Actions to execute when this module fails
    on_error:
      # action: log - Log error message
      - action: log
        message: "Exploit verification failed for {{Target}}"
        # condition: Only execute if this condition is true
        condition: 'true'

      # action: continue - Allow flow to continue despite error
      - action: continue
        message: "Continuing flow despite exploit verification failure"

      # action: abort - Stop the entire flow
      # (Usually with a condition so it doesn't always abort)
      - action: abort
        message: "Critical failure - aborting flow"
        condition: 'false'  # Only abort under specific conditions

      # action: notify - Alert on failure
      - action: notify
        notify: "Module failed: exploit-verification for {{Target}}"

      # action: export - Export error state
      - action: export
        name: exploit_verify_failed
        value: "true"

  # ===========================================================================
  # Module Reference: With Decision Routing
  # ===========================================================================
  - name: deep-scan
    path: modules/deep-scan.yaml

    depends_on:
      - vulnerability-scanning

    # decision: Conditional routing based on results
    # Determines which module to execute next based on conditions
    decision:
      # condition: JavaScript expression to evaluate
      # next: Module name to jump to, or "_end" to finish flow
      - condition: 'fileLength("{{Output}}/vuln-scan/critical.txt") > 0'
        next: notification-critical

      - condition: 'fileLength("{{Output}}/vuln-scan/high.txt") > 0'
        next: notification-high

      # Default case - continue to next module in list
      - condition: 'true'
        next: cleanup

    params:
      scan_depth: "{{scan_depth}}"

  # ===========================================================================
  # Module Reference: Notification branches (targets of decision routing)
  # ===========================================================================
  - name: notification-critical
    path: modules/notify.yaml

    # Note: This module can be jumped to via decision routing
    # It won't run in normal sequential flow unless explicitly in depends_on

    params:
      severity: critical
      message: "Critical vulnerabilities found for {{Target}}"
      channel: security-alerts

    on_success:
      - action: export
        name: notification_sent
        value: "critical"

  - name: notification-high
    path: modules/notify.yaml

    params:
      severity: high
      message: "High severity vulnerabilities found for {{Target}}"
      channel: security-team

    on_success:
      - action: export
        name: notification_sent
        value: "high"

  # ===========================================================================
  # Module Reference: Parallel Module Execution
  # Modules with same depends_on and no inter-dependencies run in parallel
  # ===========================================================================
  - name: ssl-analysis
    path: modules/ssl-check.yaml

    depends_on:
      - port-scanning  # Same dependency as web-scanning

    params:
      input: "{{Output}}/portscan/ssl-services.txt"

  - name: dns-analysis
    path: modules/dns-check.yaml

    depends_on:
      - reconnaissance  # Can run in parallel with port-scanning

    params:
      domains: "{{Output}}/recon/subdomains.txt"

  # ===========================================================================
  # Module Reference: Cleanup/Final Module
  # ===========================================================================
  - name: cleanup
    path: modules/cleanup.yaml

    # depends_on multiple modules - waits for all to complete
    depends_on:
      - vulnerability-scanning
      - exploit-verification
      - ssl-analysis
      - dns-analysis

    # condition with multiple checks
    condition: 'true'  # Always run cleanup

    params:
      output_dir: "{{Output}}"
      format: "{{output_format}}"

    on_success:
      - action: log
        message: "Flow completed successfully for {{Target}}"

      - action: notify
        notify: "Security scan flow completed for {{Target}}"

      - action: export
        name: flow_status
        value: "completed"

    on_error:
      - action: log
        message: "Cleanup failed but flow results are preserved"

      - action: continue
        message: "Flow complete despite cleanup issues"
`,
    "triggers-example": `# =============================================================================
# Flow Workflow: All Trigger Types Example
# =============================================================================
# This file demonstrates ALL trigger types available in osmedeus workflows.
# Triggers define when/how a workflow should automatically execute.
# Trigger types: cron, event, watch, manual
# =============================================================================

kind: flow
name: triggers-example
description: Demonstrates all trigger types with comprehensive field documentation
tags: triggers, automation, scheduled

# -----------------------------------------------------------------------------
# TRIGGERS SECTION
# Define automatic execution triggers for this workflow
# Multiple triggers can be defined; any triggered condition will start execution
# =============================================================================
trigger:
  # ===========================================================================
  # TRIGGER TYPE: cron
  # Schedule-based execution using cron expressions
  # ===========================================================================
  - # name: Identifier for this trigger (for logging and management)
    name: daily-scan

    # on: Trigger type - cron, event, watch, or manual
    on: cron

    # schedule: Cron expression defining when to run
    # Format: minute hour day-of-month month day-of-week
    # Examples:
    #   "0 0 * * *"     - Every day at midnight
    #   "0 */6 * * *"   - Every 6 hours
    #   "0 9 * * 1-5"   - 9 AM on weekdays
    #   "0 0 1 * *"     - First day of every month at midnight
    schedule: "0 2 * * *"  # Every day at 2 AM

    # input: Defines where the target input comes from for scheduled runs
    input:
      # type: Input source type - file, event_data, function, or param
      type: file

      # path: For "file" type - path to file containing targets (one per line)
      path: "/data/targets/active-targets.txt"

    # enabled: Whether this trigger is active
    # true = trigger is active and will fire
    # false = trigger is defined but disabled
    enabled: true

  # ---------------------------------------------------------------------------
  # Cron trigger with function-based input
  # ---------------------------------------------------------------------------
  - name: weekly-full-scan
    on: cron
    schedule: "0 0 * * 0"  # Every Sunday at midnight

    input:
      # type: function - Generate input dynamically using a function
      type: function

      # function: JavaScript function to generate/retrieve targets
      # Can use built-in functions like db queries, API calls, etc.
      function: 'get_targets_from_db("scope:production")'

    enabled: true

  # ===========================================================================
  # TRIGGER TYPE: event
  # Event-driven execution based on system events
  # Events follow topic format: <component>.<event_type>
  # ===========================================================================
  - name: webhook-trigger
    on: event

    # event: Event configuration for event triggers
    event:
      # topic: Event topic to subscribe to
      # Common topics:
      #   webhook.received    - External webhook received
      #   assets.new          - New asset discovered
      #   assets.changed      - Asset data changed
      #   db.change           - Database record changed
      #   watch.files         - File system change detected
      topic: webhook.received

      # filters: JavaScript expressions to filter events
      # Event data available as 'event' object with fields:
      #   event.name      - Event name
      #   event.source    - Event source
      #   event.data      - JSON payload (string)
      #   event.data_type - Type of data
      # All filters must evaluate to true for trigger to fire
      filters:
        - 'event.source == "github"'
        - 'event.name == "push"'

    # input: How to extract target from event data
    input:
      # type: event_data - Extract from event payload
      type: event_data

      # field: JSON path to extract from event.data
      # Uses dot notation for nested fields
      field: "repository.html_url"

    enabled: true

  # ---------------------------------------------------------------------------
  # Event trigger for new asset discovery
  # ---------------------------------------------------------------------------
  - name: new-asset-scan
    on: event

    event:
      topic: assets.new

      filters:
        # Filter for specific asset types
        - 'event.data_type == "subdomain"'
        # Filter by source tool
        - 'event.source == "subfinder" || event.source == "amass"'

    input:
      type: event_data
      field: "hostname"

    enabled: true

  # ---------------------------------------------------------------------------
  # Event trigger with function-based input extraction
  # ---------------------------------------------------------------------------
  - name: vuln-alert-trigger
    on: event

    event:
      topic: webhook.received

      filters:
        - 'event.name == "vulnerability_alert"'
        - 'JSON.parse(event.data).severity == "critical"'

    input:
      # type: function - Use function to parse/transform event data
      type: function

      # function: Transform event data to target format
      function: 'jq("{{event.data}}", ".affected_host")'

    enabled: true

  # ===========================================================================
  # TRIGGER TYPE: watch
  # File system watch - triggers when files change
  # ===========================================================================
  - name: targets-file-watch
    on: watch

    # path: File or directory path to watch for changes
    # Supports glob patterns in some implementations
    path: "/data/targets/new-targets.txt"

    # input: How to get targets when file changes
    input:
      type: file
      path: "/data/targets/new-targets.txt"

    enabled: true

  # ---------------------------------------------------------------------------
  # Watch trigger on directory
  # ---------------------------------------------------------------------------
  - name: input-directory-watch
    on: watch

    path: "/data/incoming/"

    input:
      # type: function - Process newly added files
      type: function
      function: 'get_new_files("/data/incoming/", "*.txt")'

    enabled: true

  # ===========================================================================
  # TRIGGER TYPE: manual
  # Explicit manual trigger control
  # Used to enable/disable CLI execution for this workflow
  # ===========================================================================
  - name: manual-execution
    on: manual

    # For manual triggers, enabled controls whether CLI can run this workflow
    # enabled: true  - Allow: osmedeus run -f triggers-example -t target
    # enabled: false - Block CLI execution (only scheduled/event triggers work)
    enabled: true

    # input: Default input for manual execution
    # This is optional; CLI -t flag overrides this
    input:
      # type: param - Use a parameter as input
      type: param

      # name: Parameter name to use as target
      name: Target

  # ---------------------------------------------------------------------------
  # Disabled manual trigger example
  # This workflow can ONLY be triggered via cron/events, not CLI
  # ---------------------------------------------------------------------------
  # Uncomment to see the effect:
  # - name: block-manual
  #   on: manual
  #   enabled: false

# -----------------------------------------------------------------------------
# PARAMS SECTION
# -----------------------------------------------------------------------------
params:
  - name: scan_type
    default: "standard"

  - name: threads
    default: "10"

# -----------------------------------------------------------------------------
# MODULES SECTION
# The actual workflow steps to execute when any trigger fires
# -----------------------------------------------------------------------------
modules:
  - name: initial-recon
    path: modules/recon.yaml
    params:
      threads: "{{threads}}"

  - name: scanning
    path: modules/scan.yaml
    depends_on:
      - initial-recon
    params:
      scan_type: "{{scan_type}}"

  - name: reporting
    path: modules/report.yaml
    depends_on:
      - scanning

    on_success:
      - action: notify
        notify: "Triggered scan completed for {{Target}}"
        # condition: Only notify for certain triggers
        condition: 'true'

      - action: export
        name: completed_at
        value: "{{currentDate()}}"
`,
    "docker-runner-example": `# =============================================================================
# Module Workflow: Docker Runner Configuration Example
# =============================================================================
# This file demonstrates all Docker runner configuration fields at both
# the workflow level (for all steps) and step level (per-step override).
# =============================================================================

kind: module
name: docker-runner-example
description: Demonstrates Docker runner configuration with all available fields
tags: docker, runner, container

# -----------------------------------------------------------------------------
# RUNNER CONFIGURATION (Workflow-Level)
# Applies to all steps unless overridden at step level
# -----------------------------------------------------------------------------

# runner: Execution environment for this workflow
# Options: host (default - local machine), docker, ssh
runner: docker

# runner_config: Configuration for the selected runner type
runner_config:
  # -------------------------------------------------------------------------
  # DOCKER-SPECIFIC CONFIGURATION
  # -------------------------------------------------------------------------

  # image: Docker image to use (required for docker runner)
  # Format: registry/image:tag or just image:tag
  image: ubuntu:22.04

  # env: Environment variables to set inside the container
  # Map of VAR_NAME: value
  env:
    MY_VAR: my-value
    API_KEY: "{{api_key}}"  # Can use template variables
    THREADS: "{{threads}}"

  # volumes: Volume mounts in docker format
  # Format: host_path:container_path[:options]
  # Options: ro (read-only), rw (read-write)
  volumes:
    - "/tmp/osmedeus:/data"
    - "{{Output}}:/output"
    - "/etc/hosts:/etc/hosts:ro"

  # network: Docker network mode
  # Options: bridge (default), host, none, container:<name>, or network name
  network: host

  # persistent: Container lifecycle mode
  # true = reuse the same container across steps (faster, state preserved)
  # false = ephemeral, create new container per step (isolated, clean state)
  persistent: true

  # -------------------------------------------------------------------------
  # COMMON CONFIGURATION (applies to docker and ssh)
  # -------------------------------------------------------------------------

  # workdir: Working directory inside the container/remote
  # Commands will execute in this directory
  workdir: /app

params:
  - name: api_key
    default: "demo-key"

  - name: threads
    default: "5"

steps:
  # ===========================================================================
  # Step using workflow-level runner (docker with ubuntu:22.04)
  # ===========================================================================
  - name: use-workflow-runner
    type: bash
    log: "Running in workflow-level Docker container"
    command: 'echo "Running inside ubuntu:22.04 container"'

  # ===========================================================================
  # Step with per-step Docker runner override
  # Uses different image than workflow-level config
  # ===========================================================================
  - name: step-with-runner-override
    type: bash
    log: "Running in step-specific Docker container"

    # step_runner: Override runner type for this step only
    # Options: host, docker, ssh
    step_runner: docker

    # step_runner_config: Override runner configuration for this step
    # Same structure as runner_config but applies only to this step
    step_runner_config:
      # Use a different image for this specific step
      image: python:3.11-slim

      env:
        PYTHONPATH: /app

      volumes:
        - "{{Output}}:/output:rw"

      network: bridge

      persistent: false

      workdir: /app

    command: 'python3 -c "print(\\"Running in Python container\\")"'

  # ===========================================================================
  # Remote-bash step type with Docker (explicit remote-bash type)
  # remote-bash is specifically for executing commands in remote environments
  # ===========================================================================
  - name: remote-bash-docker
    # type: remote-bash is specifically for remote execution (docker/ssh)
    type: remote-bash
    log: "Remote bash execution in Docker"

    # step_runner: Required for remote-bash type - specifies execution environment
    # Must be "docker" or "ssh"
    step_runner: docker

    step_runner_config:
      image: alpine:latest
      workdir: /tmp

    # command/commands/parallel_commands: Same as bash step
    command: 'echo "Hello from Alpine container" > /tmp/output.txt'

    # step_remote_file: File path on remote (inside container) to copy after execution
    # This file will be copied from the container to the host
    step_remote_file: /tmp/output.txt

    # host_output_file: Local path where the remote file will be copied
    # Template variables are supported
    host_output_file: "{{Output}}/docker-output.txt"

  # ===========================================================================
  # Parallel commands in Docker container
  # ===========================================================================
  - name: docker-parallel-commands
    type: bash
    log: "Running parallel commands in Docker"
    step_runner: docker
    step_runner_config:
      image: ubuntu:22.04
      persistent: true

    parallel_commands:
      - 'sleep 2 && echo "Parallel job A completed"'
      - 'sleep 1 && echo "Parallel job B completed"'
      - 'sleep 3 && echo "Parallel job C completed"'

  # ===========================================================================
  # Foreach loop executing in Docker
  # ===========================================================================
  - name: docker-foreach
    type: foreach
    log: "Processing items in Docker containers"
    input: "{{Output}}/targets.txt"
    variable: target
    threads: 3

    step:
      name: process-in-docker
      type: bash
      step_runner: docker
      step_runner_config:
        image: curlimages/curl:latest
        network: host
      command: 'curl -s -o /dev/null -w "%{http_code}" "[[target]]"'
      exports:
        http_status: "{{stdout}}"

  # ===========================================================================
  # Step running on host (override workflow's docker runner)
  # ===========================================================================
  - name: run-on-host
    type: bash
    log: "Running on host machine (overriding workflow runner)"

    # Override to run locally instead of in container
    step_runner: host

    command: 'echo "This runs directly on the host machine"'

  # ===========================================================================
  # Docker step with all structured arguments
  # ===========================================================================
  - name: docker-with-args
    type: bash
    log: "Docker step with structured arguments"
    step_runner: docker
    step_runner_config:
      image: nuclei:latest
      volumes:
        - "{{Output}}:/output"
        - "/root/nuclei-templates:/templates:ro"
      workdir: /output

    command: nuclei
    speed_args: '-rate-limit 100 -c {{threads}}'
    config_args: '-t /templates/cves/'
    input_args: '-u {{Target}}'
    output_args: '-o /output/nuclei-results.txt'

    step_remote_file: /output/nuclei-results.txt
    host_output_file: "{{Output}}/nuclei-results.txt"

    exports:
      nuclei_output: "{{Output}}/nuclei-results.txt"
`,
    "ssh-runner-example": `# =============================================================================
# Module Workflow: SSH Runner Configuration Example
# =============================================================================
# This file demonstrates all SSH runner configuration fields at both
# the workflow level (for all steps) and step level (per-step override).
# =============================================================================

kind: module
name: ssh-runner-example
description: Demonstrates SSH runner configuration with all available fields
tags: ssh, runner, remote

# -----------------------------------------------------------------------------
# RUNNER CONFIGURATION (Workflow-Level)
# Applies to all steps unless overridden at step level
# -----------------------------------------------------------------------------

# runner: Execution environment for this workflow
# Options: host (default - local machine), docker, ssh
runner: ssh

# runner_config: Configuration for the selected runner type
runner_config:
  # -------------------------------------------------------------------------
  # SSH-SPECIFIC CONFIGURATION
  # -------------------------------------------------------------------------

  # host: SSH hostname or IP address (required for ssh runner)
  # Can use template variables for dynamic targeting
  host: "{{ssh_host}}"

  # port: SSH port number
  # Default: 22
  port: 22

  # user: SSH username for authentication
  user: "{{ssh_user}}"

  # key_file: Path to SSH private key file for key-based authentication
  # Preferred over password authentication for security
  key_file: "{{ssh_key_path}}"

  # password: SSH password for password-based authentication
  # WARNING: Not recommended - use key_file instead when possible
  # Can use template variables or environment references
  # password: "{{ssh_password}}"

  # -------------------------------------------------------------------------
  # COMMON CONFIGURATION (applies to docker and ssh)
  # -------------------------------------------------------------------------

  # workdir: Working directory on the remote machine
  # Commands will execute in this directory
  workdir: /home/scanner/workspace

params:
  - name: ssh_host
    default: "192.168.1.100"
    required: true

  - name: ssh_user
    default: "scanner"
    required: true

  - name: ssh_key_path
    default: "~/.ssh/id_rsa"

  - name: threads
    default: "10"

steps:
  # ===========================================================================
  # Step using workflow-level SSH runner
  # ===========================================================================
  - name: setup-remote-workspace
    type: bash
    log: "Setting up workspace on remote SSH server"
    command: 'mkdir -p /home/scanner/workspace/results && echo "Workspace ready"'

  # ===========================================================================
  # Remote-bash step type with SSH (explicit remote-bash type)
  # remote-bash is specifically designed for remote execution scenarios
  # ===========================================================================
  - name: remote-bash-ssh
    # type: remote-bash is explicitly for remote execution (docker/ssh)
    type: remote-bash
    log: "Remote bash execution via SSH"

    # step_runner: Required for remote-bash type - must be "docker" or "ssh"
    step_runner: ssh

    # step_runner_config: SSH configuration (inherits from workflow if not set)
    # Omitting this uses workflow-level runner_config
    step_runner_config:
      host: "{{ssh_host}}"
      port: 22
      user: "{{ssh_user}}"
      key_file: "{{ssh_key_path}}"
      workdir: /tmp

    # command: Command to execute on remote server
    command: 'hostname && whoami && pwd > /tmp/remote-info.txt'

    # step_remote_file: File on remote server to copy back to local host
    # This is useful for retrieving results from remote execution
    step_remote_file: /tmp/remote-info.txt

    # host_output_file: Local path where remote file will be copied
    host_output_file: "{{Output}}/remote-info.txt"

    exports:
      remote_file: "{{Output}}/remote-info.txt"

  # ===========================================================================
  # Step overriding SSH connection to different server
  # ===========================================================================
  - name: connect-to-secondary-server
    type: bash
    log: "Connecting to secondary server"

    # Override workflow runner with different SSH target
    step_runner: ssh

    step_runner_config:
      host: "192.168.1.101"  # Different server
      port: 2222             # Non-standard port
      user: admin
      key_file: "~/.ssh/secondary_key"
      workdir: /opt/scanner

    command: 'echo "Connected to secondary server" && uptime'

  # ===========================================================================
  # Multiple sequential commands via SSH
  # ===========================================================================
  - name: ssh-multiple-commands
    type: bash
    log: "Running multiple commands on remote"

    # commands: List of commands executed sequentially on remote
    commands:
      - 'echo "Step 1: Checking system"'
      - 'df -h'
      - 'echo "Step 2: Checking memory"'
      - 'free -m'
      - 'echo "Step 3: Checking processes"'
      - 'ps aux | head -10'

    std_file: "{{Output}}/system-check.txt"

  # ===========================================================================
  # Parallel commands on SSH (run concurrently on remote)
  # ===========================================================================
  - name: ssh-parallel-commands
    type: bash
    log: "Running parallel commands on remote SSH server"

    parallel_commands:
      - 'nmap -sS -p 80 {{Target}} > /tmp/port80.txt'
      - 'nmap -sS -p 443 {{Target}} > /tmp/port443.txt'
      - 'nmap -sS -p 22 {{Target}} > /tmp/port22.txt'

  # ===========================================================================
  # Run tool with structured arguments via SSH
  # ===========================================================================
  - name: ssh-nuclei-scan
    type: bash
    log: "Running nuclei scan via SSH"
    timeout: 3600

    command: nuclei
    speed_args: '-rate-limit 50 -c {{threads}}'
    config_args: '-t ~/nuclei-templates/cves/'
    input_args: '-u {{Target}}'
    output_args: '-o /home/scanner/workspace/nuclei-results.json -json'

    step_remote_file: /home/scanner/workspace/nuclei-results.json
    host_output_file: "{{Output}}/nuclei-results.json"

    exports:
      scan_results: "{{Output}}/nuclei-results.json"

  # ===========================================================================
  # Foreach loop with SSH execution
  # Processes multiple targets on remote server
  # ===========================================================================
  - name: ssh-foreach-targets
    type: foreach
    log: "Processing targets via SSH"

    # input: File containing targets (one per line)
    input: "{{Output}}/targets.txt"

    # variable: Loop variable accessed as [[variable]] in inner step
    variable: current_target

    # threads: Number of concurrent SSH executions
    threads: 5

    step:
      name: probe-target
      type: bash
      # Inner step inherits workflow-level SSH runner
      command: 'curl -s -o /dev/null -w "%{http_code}" "[[current_target]]" 2>/dev/null || echo "failed"'
      exports:
        probe_result: "{{stdout}}"

  # ===========================================================================
  # Step running on local host (override workflow's SSH runner)
  # Useful for local processing of results retrieved from remote
  # ===========================================================================
  - name: process-results-locally
    type: bash
    log: "Processing results on local host"

    # Override to run locally instead of via SSH
    step_runner: host

    command: 'cat "{{Output}}/nuclei-results.json" | jq -r ".info.severity" | sort | uniq -c'

    exports:
      severity_summary: "{{stdout}}"

  # ===========================================================================
  # Function step (always runs locally, regardless of workflow runner)
  # Note: Function steps execute on the host running osmedeus, not remote
  # ===========================================================================
  - name: log-completion
    type: function
    log: "Logging scan completion"
    function: 'log_info("SSH scan completed for {{Target}}")'

  # ===========================================================================
  # Cleanup step on remote server
  # ===========================================================================
  - name: cleanup-remote
    type: bash
    log: "Cleaning up remote workspace"
    command: 'rm -rf /home/scanner/workspace/temp/* 2>/dev/null; echo "Cleanup complete"'

    on_success:
      - action: log
        message: "Remote cleanup completed successfully"

    on_error:
      - action: continue
        message: "Cleanup failed but continuing workflow"
`,
    "all-step-types-example": `# =============================================================================
# Module Workflow: All Step Types Example
# =============================================================================
# This file demonstrates ALL fields available in a module-kind workflow,
# showcasing every step type with comprehensive comments.
# =============================================================================

# -----------------------------------------------------------------------------
# WORKFLOW-LEVEL FIELDS
# -----------------------------------------------------------------------------

# kind: Workflow type - either "module" (single unit with steps) or "flow" (orchestrates modules)
kind: module

# name: Unique identifier for this workflow (required)
name: all-step-types-example

# description: Human-readable description of what this workflow does
description: Demonstrates all step types and their fields with detailed comments

# tags: Comma-separated tags for filtering and categorization (parsed as []string)
tags: example, comprehensive, demo

# -----------------------------------------------------------------------------
# PARAMS SECTION
# Define workflow parameters that can be passed via CLI or referenced in templates
# -----------------------------------------------------------------------------
params:
  # name: Parameter identifier used in templates as {{param_name}}
  # default: Default value if not provided via CLI
  # required: If true, workflow fails without this value
  # generator: Function to generate value, e.g., uuid(), currentDate(), getEnvVar("KEY")
  - name: message
    default: "Hello World"
    required: false

  - name: output_dir
    default: "{{Output}}/results"  # Can reference built-in variables
    required: false

  - name: threads
    default: "10"
    required: false

  - name: run_id
    generator: uuid()  # Generates a unique ID automatically

# -----------------------------------------------------------------------------
# DEPENDENCIES SECTION
# Validate requirements before workflow execution
# -----------------------------------------------------------------------------
dependencies:
  # commands: List of binaries/commands that must exist in PATH
  commands:
    - echo
    - curl

  # files: List of files/directories that must exist
  files:
    - /tmp

  # variables: Define variable requirements with type validation
  # Types: domain, path, number, file, string
  variables:
    - name: Target
      type: string
      required: true

  # functions_conditions: JavaScript expressions that must evaluate to true
  functions_conditions:
    - '1 + 1 == 2'

# -----------------------------------------------------------------------------
# REPORTS SECTION
# Define output files produced by this workflow
# -----------------------------------------------------------------------------
reports:
  # name: Display name for the report
  # path: File path (can use templates like {{Output}})
  # type: Format type - text, csv, json, markdown, etc.
  # description: Human-readable description
  - name: main-output
    path: "{{Output}}/main-results.txt"
    type: text
    description: Main output file from the workflow

  - name: json-results
    path: "{{Output}}/results.json"
    type: json
    description: Structured JSON output

# -----------------------------------------------------------------------------
# PREFERENCES SECTION (Optional)
# Set CLI-like flags directly in the workflow. CLI flags always take precedence.
# -----------------------------------------------------------------------------
preferences:
  # skip_workspace: Equivalent to --disable-workspace-creation
  skip_workspace: false

  # disable_notifications: Equivalent to --disable-notification
  disable_notifications: true

  # disable_logging: Equivalent to --disable-logging
  disable_logging: false

  # heuristics_check: Equivalent to --heuristics-check (none, basic, advanced)
  heuristics_check: 'basic'

  # ci_output_format: Equivalent to --ci-output-format
  ci_output_format: false

  # silent: Equivalent to --silent
  silent: false

  # repeat: Equivalent to --repeat
  repeat: false

  # repeat_wait_time: Equivalent to --repeat-wait-time (e.g., 30s, 1h, 2h30m)
  repeat_wait_time: '60s'

  # clean_up_workspace: Equivalent to --clean-up-workspace
  clean_up_workspace: false

# -----------------------------------------------------------------------------
# STEPS SECTION
# The ordered list of execution steps for this module
# -----------------------------------------------------------------------------
steps:
  # ===========================================================================
  # STEP TYPE: bash
  # Execute shell commands on the host (or configured runner)
  # ===========================================================================
  - name: bash-single-command
    # type: Step type - bash, function, parallel-steps, foreach, remote-bash, http, llm
    type: bash

    # pre_condition: JavaScript expression - step only runs if this evaluates to true
    pre_condition: 'true'

    # log: Custom log message displayed when step starts (supports templates)
    log: "Executing single bash command for {{Target}}"

    # timeout: Maximum execution time in seconds (0 = no timeout)
    timeout: 60

    # command: Single command to execute
    command: 'echo "Processing target: {{Target}} with message: {{message}}"'

    # std_file: File path to save stdout/stderr output
    std_file: "{{Output}}/step1-output.txt"

    # exports: Variables to export for subsequent steps
    # Key = variable name, Value = extraction pattern or literal value
    exports:
      step1_result: "completed"

  # ---------------------------------------------------------------------------
  # Bash step with multiple sequential commands
  # ---------------------------------------------------------------------------
  - name: bash-multiple-commands
    type: bash
    log: "Running multiple sequential commands"

    # commands: List of commands executed sequentially
    commands:
      - 'echo "First command"'
      - 'echo "Second command"'
      - 'echo "Third command"'

  # ---------------------------------------------------------------------------
  # Bash step with parallel commands
  # ---------------------------------------------------------------------------
  - name: bash-parallel-commands
    type: bash
    log: "Running commands in parallel"

    # parallel_commands: List of commands executed concurrently
    parallel_commands:
      - 'echo "Parallel A" && sleep 1'
      - 'echo "Parallel B" && sleep 1'
      - 'echo "Parallel C" && sleep 1'

  # ---------------------------------------------------------------------------
  # Bash step with structured arguments
  # Arguments are joined in order: command + speed + config + input + output
  # ---------------------------------------------------------------------------
  - name: bash-structured-args
    type: bash
    log: "Using structured argument fields"

    command: 'echo'

    # speed_args: Performance-related arguments (e.g., thread count, rate limits)
    speed_args: '-n'

    # config_args: Configuration arguments (e.g., config file paths)
    config_args: ''

    # input_args: Input-related arguments (e.g., input file, target)
    input_args: '"Structured arguments test"'

    # output_args: Output-related arguments (e.g., output file, format)
    output_args: ''

  # ===========================================================================
  # STEP TYPE: function
  # Execute built-in utility functions via Otto JavaScript runtime
  # ===========================================================================
  - name: function-single
    type: function
    log: "Executing single function"

    # function: Single function call (JavaScript expression)
    function: 'log_info("Processing {{Target}} in function step")'

  # ---------------------------------------------------------------------------
  # Function step with multiple sequential functions
  # ---------------------------------------------------------------------------
  - name: function-multiple
    type: function
    log: "Executing multiple functions sequentially"

    # functions: List of functions executed sequentially
    functions:
      - 'log_info("Function 1")'
      - 'log_info("Function 2")'
      - 'log_info("Function 3")'

  # ---------------------------------------------------------------------------
  # Function step with parallel functions
  # ---------------------------------------------------------------------------
  - name: function-parallel
    type: function
    log: "Executing functions in parallel"

    # parallel_functions: List of functions executed concurrently
    parallel_functions:
      - 'log_info("Parallel Function A")'
      - 'log_info("Parallel Function B")'
      - 'log_info("Parallel Function C")'

  # ===========================================================================
  # STEP TYPE: parallel-steps
  # Execute multiple complete steps in parallel
  # ===========================================================================
  - name: parallel-step-container
    type: parallel-steps
    log: "Running multiple steps in parallel"

    # parallel_steps: List of Step objects executed concurrently
    parallel_steps:
      - name: parallel-inner-1
        type: bash
        command: 'echo "Inner parallel step 1"'

      - name: parallel-inner-2
        type: function
        function: 'log_info("Inner parallel step 2")'

      - name: parallel-inner-3
        type: bash
        command: 'echo "Inner parallel step 3"'

  # ===========================================================================
  # STEP TYPE: foreach
  # Iterate over input lines, executing inner step for each
  # ===========================================================================
  - name: foreach-example
    type: foreach
    log: "Iterating over items"

    # input: File path or direct content to iterate over (one item per line)
    input: "{{Output}}/items.txt"

    # variable: Name for the loop variable, accessed as [[variable]] in inner step
    variable: item

    # threads: Number of concurrent iterations (default: 1 = sequential)
    threads: 5

    # step: The inner step to execute for each item (single Step object)
    step:
      name: process-item
      type: bash
      command: 'echo "Processing [[item]]"'
      exports:
        processed_item: "[[item]]"

  # ===========================================================================
  # STEP TYPE: http
  # Make HTTP requests to external APIs
  # ===========================================================================
  - name: http-request
    type: http
    log: "Making HTTP request"
    timeout: 30

    # url: Target URL for the request (required for http type)
    url: "https://httpbin.org/post"

    # method: HTTP method - GET, POST, PUT, DELETE, PATCH, etc.
    method: POST

    # headers: Map of HTTP headers to send
    headers:
      Content-Type: application/json
      Authorization: "Bearer {{api_token}}"
      X-Custom-Header: custom-value

    # request_body: Request body content (typically JSON for POST/PUT)
    request_body: |
      {
        "target": "{{Target}}",
        "message": "{{message}}"
      }

    exports:
      http_response: "{{response.body}}"

  # ===========================================================================
  # STEP TYPE: llm
  # Make LLM API calls for AI-powered processing
  # ===========================================================================
  - name: llm-chat-completion
    type: llm
    log: "Calling LLM for analysis"
    timeout: 120

    # messages: Conversation messages for chat completion
    # role: system, user, assistant, or tool
    # content: Message text (can be string or multimodal array)
    messages:
      - role: system
        content: "You are a security analysis assistant."

      - role: user
        # content can be a simple string or complex multimodal content
        content: "Analyze this target: {{Target}}"

    # tools: Function tools available to the LLM
    tools:
      - type: function  # Currently only "function" type supported
        function:
          name: analyze_target
          description: Analyzes a target for security vulnerabilities
          # parameters: JSON Schema defining function parameters
          parameters:
            type: object
            properties:
              target:
                type: string
                description: The target to analyze
              depth:
                type: string
                enum: [shallow, deep]
            required:
              - target

    # tool_choice: How the model should choose tools
    # Can be: "auto", "none", "required", or {"type": "function", "function": {"name": "fn_name"}}
    tool_choice: auto

    # llm_config: Step-level LLM configuration overrides
    llm_config:
      # provider: Specific provider to use (overrides rotation)
      provider: openai

      # model: Model override for this step
      model: gpt-4

      # Generation parameters
      max_tokens: 1000
      temperature: 0.7
      top_p: 1.0

      # Request settings
      timeout: "60s"
      max_retries: 3
      stream: false

      # response_format: Control output format
      # type: "text", "json_object", or "json_schema"
      response_format:
        type: json_object

    # extra_llm_parameters: Additional provider-specific parameters
    extra_llm_parameters:
      seed: 42
      presence_penalty: 0.0

    exports:
      llm_analysis: "{{response.content}}"

  # ---------------------------------------------------------------------------
  # LLM step for embeddings
  # ---------------------------------------------------------------------------
  - name: llm-embedding
    type: llm
    log: "Generating text embeddings"

    # is_embedding: Flag to indicate this is an embedding request
    is_embedding: true

    # embedding_input: List of texts to generate embeddings for
    embedding_input:
      - "Security vulnerability in {{Target}}"
      - "Network reconnaissance results"
      - "Port scan findings"

    llm_config:
      model: text-embedding-3-small

    exports:
      embeddings: "{{response.embeddings}}"

  # ===========================================================================
  # COMMON STEP FIELDS: on_success, on_error, decision
  # These fields are available on ALL step types
  # ===========================================================================
  - name: step-with-handlers
    type: bash
    log: "Step demonstrating success/error handlers and decision routing"
    command: 'echo "Running step with all handler types"'

    # on_success: Actions to execute when step succeeds
    on_success:
      # action: Handler type - log, abort, continue, export, run, notify
      - action: log
        message: "Step completed successfully for {{Target}}"

      - action: export
        # name: Variable name to export
        name: success_flag
        # value: Value to export (can be string, number, or template)
        value: "true"

      - action: notify
        # notify: Notification message
        notify: "Step succeeded for {{Target}}"

      - action: run
        # type: Step type to run (bash or function)
        type: bash
        command: 'echo "Running follow-up command"'

      - action: run
        type: function
        functions:
          - 'log_info("Running follow-up function")'

    # on_error: Actions to execute when step fails
    on_error:
      - action: log
        message: "Step failed for {{Target}}"
        # condition: Only execute this action if condition evaluates to true
        condition: 'true'

      - action: notify
        notify: "Error in workflow for {{Target}}"

      # abort: Stops workflow execution immediately
      - action: abort
        message: "Aborting due to critical failure"
        condition: 'false'  # Only abort under specific conditions

      # continue: Allows workflow to continue despite error
      - action: continue
        message: "Continuing despite error"

    # decision: Conditional routing to other steps or workflow end
    decision:
      # condition: JavaScript expression to evaluate
      # next: Step name to jump to, or "_end" to finish workflow
      - condition: '{{success_flag}} == "true"'
        next: final-step

      - condition: '{{success_flag}} != "true"'
        next: _end  # Special value to end workflow

  # ---------------------------------------------------------------------------
  # Final step
  # ---------------------------------------------------------------------------
  - name: final-step
    type: function
    log: "Final step - workflow complete"
    function: 'log_info("All step types demonstrated for {{Target}}")'
`,
    "mock-all-step-types-example": `# =============================================================================
# Module Workflow: All Step Types Example
# =============================================================================
# This file demonstrates ALL fields available in a module-kind workflow,
# showcasing every step type with comprehensive comments.
# =============================================================================

# -----------------------------------------------------------------------------
# WORKFLOW-LEVEL FIELDS
# -----------------------------------------------------------------------------

# kind: Workflow type - either "module" (single unit with steps) or "flow" (orchestrates modules)
kind: module

# name: Unique identifier for this workflow (required)
name: mock-all-step-types-example

# description: Human-readable description of what this workflow does
description: Mock Demonstrates all step types and their fields with detailed comments

# tags: Comma-separated tags for filtering and categorization (parsed as []string)
tags: example, comprehensive, demo

# -----------------------------------------------------------------------------
# PARAMS SECTION
# Define workflow parameters that can be passed via CLI or referenced in templates
# -----------------------------------------------------------------------------
params:
  - name: message
    default: "Hello World"
    required: false

  - name: output_dir
    default: "{{Output}}/results"
    required: false

  - name: threads
    default: "10"
    required: false

  - name: run_id
    generator: uuid()

# -----------------------------------------------------------------------------
# DEPENDENCIES SECTION
# Validate requirements before workflow execution
# -----------------------------------------------------------------------------
dependencies:
  commands:
    - echo
    - curl

  files:
    - /tmp

  variables:
    - name: Target
      type: string
      required: true

  functions_conditions:
    - '1 + 1 == 2'

# -----------------------------------------------------------------------------
# REPORTS SECTION
# Define output files produced by this workflow
# -----------------------------------------------------------------------------
reports:
  - name: main-output
    path: "{{Output}}/main-results.txt"
    type: text
    description: Main output file from the workflow

  - name: json-results
    path: "{{Output}}/results.json"
    type: json
    description: Structured JSON output

# -----------------------------------------------------------------------------
# PREFERENCES SECTION (Optional)
# Set CLI-like flags directly in the workflow. CLI flags always take precedence.
# -----------------------------------------------------------------------------
preferences:
  skip_workspace: false
  disable_notifications: true
  disable_logging: false
  heuristics_check: 'basic'
  ci_output_format: false
  silent: false
  repeat: false
  repeat_wait_time: '60s'
  clean_up_workspace: false

# -----------------------------------------------------------------------------
# STEPS SECTION
# The ordered list of execution steps for this module
# -----------------------------------------------------------------------------
steps:
  - name: bash-single-command
    type: bash
    pre_condition: 'true'
    log: "Executing single bash command for {{Target}}"
    timeout: 60
    command: 'echo "Processing target: {{Target}} with message: {{message}}"'
    std_file: "{{Output}}/step1-output.txt"
    exports:
      step1_result: "completed"

  - name: bash-multiple-commands
    type: bash
    log: "Running multiple sequential commands"
    commands:
      - 'echo "First command"'
      - 'echo "Second command"'
      - 'echo "Third command"'

  - name: bash-parallel-commands
    type: bash
    log: "Running commands in parallel"
    parallel_commands:
      - 'echo "Parallel A" && sleep 1'
      - 'echo "Parallel B" && sleep 1'
      - 'echo "Parallel C" && sleep 1'

  - name: bash-structured-args
    type: bash
    log: "Using structured argument fields"
    command: 'echo'
    speed_args: '-n'
    config_args: ''
    input_args: '"Structured arguments test"'
    output_args: ''

  - name: function-single
    type: function
    log: "Executing single function"
    function: 'log_info("Processing {{Target}} in function step")'

  - name: function-multiple
    type: function
    log: "Executing multiple functions sequentially"
    functions:
      - 'log_info("Function 1")'
      - 'log_info("Function 2")'
      - 'log_info("Function 3")'

  - name: function-parallel
    type: function
    log: "Executing functions in parallel"
    parallel_functions:
      - 'log_info("Parallel Function A")'
      - 'log_info("Parallel Function B")'
      - 'log_info("Parallel Function C")'

  - name: parallel-step-container
    type: parallel-steps
    log: "Running multiple steps in parallel"
    parallel_steps:
      - name: parallel-inner-1
        type: bash
        command: 'echo "Inner parallel step 1"'
      - name: parallel-inner-2
        type: function
        function: 'log_info("Inner parallel step 2")'
      - name: parallel-inner-3
        type: bash
        command: 'echo "Inner parallel step 3"'

  - name: foreach-example
    type: foreach
    log: "Iterating over items"
    input: "{{Output}}/items.txt"
    variable: item
    threads: 5
    step:
      name: process-item
      type: bash
      command: 'echo "Processing [[item]]"'
      exports:
        processed_item: "[[item]]"

  - name: http-request
    type: http
    log: "Making HTTP request"
    timeout: 30
    url: "https://httpbin.org/post"
    method: POST
    headers:
      Content-Type: application/json
      Authorization: "Bearer {{api_token}}"
      X-Custom-Header: custom-value
    request_body: |
      {
        "target": "{{Target}}",
        "message": "{{message}}"
      }
    exports:
      http_response: "{{response.body}}"

  - name: llm-chat-completion
    type: llm
    log: "Calling LLM for analysis"
    timeout: 120
    messages:
      - role: system
        content: "You are a security analysis assistant."
      - role: user
        content: "Analyze this target: {{Target}}"
    tools:
      - type: function
        function:
          name: analyze_target
          description: Analyzes a target for security vulnerabilities
          parameters:
            type: object
            properties:
              target:
                type: string
                description: The target to analyze
              depth:
                type: string
                enum: [shallow, deep]
            required:
              - target
    tool_choice: auto
    llm_config:
      provider: openai
      model: gpt-4
      max_tokens: 1000
      temperature: 0.7
      top_p: 1.0
      timeout: "60s"
      max_retries: 3
      stream: false
      response_format:
        type: json_object
    extra_llm_parameters:
      seed: 42
      presence_penalty: 0.0
    exports:
      llm_analysis: "{{response.content}}"

  - name: llm-embedding
    type: llm
    log: "Generating text embeddings"
    is_embedding: true
    embedding_input:
      - "Security vulnerability in {{Target}}"
      - "Network reconnaissance results"
      - "Port scan findings"
    llm_config:
      model: text-embedding-3-small
    exports:
      embeddings: "{{response.embeddings}}"

  - name: step-with-handlers
    type: bash
    log: "Step demonstrating success/error handlers and decision routing"
    command: 'echo "Running step with all handler types"'
    on_success:
      - action: log
        message: "Step completed successfully for {{Target}}"
      - action: export
        name: success_flag
        value: "true"
      - action: notify
        notify: "Step succeeded for {{Target}}"
      - action: run
        type: bash
        command: 'echo "Running follow-up command"'
      - action: run
        type: function
        functions:
          - 'log_info("Running follow-up function")'
    on_error:
      - action: log
        message: "Step failed for {{Target}}"
        condition: 'true'
      - action: notify
        notify: "Error in workflow for {{Target}}"
      - action: abort
        message: "Aborting due to critical failure"
        condition: 'false'
      - action: continue
        message: "Continuing despite error"
    decision:
      - condition: '{{success_flag}} == "true"'
        next: final-step
      - condition: '{{success_flag}} != "true"'
        next: _end

  - name: final-step
    type: function
    log: "Final step - workflow complete"
    function: 'log_info("All step types demonstrated for {{Target}}")'
`
};
}),
"[project]/lib/api/workflows.ts [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "fetchMockWorkflowsList",
    ()=>fetchMockWorkflowsList,
    "fetchWorkflow",
    ()=>fetchWorkflow,
    "fetchWorkflowTags",
    ()=>fetchWorkflowTags,
    "fetchWorkflowYaml",
    ()=>fetchWorkflowYaml,
    "fetchWorkflows",
    ()=>fetchWorkflows,
    "fetchWorkflowsList",
    ()=>fetchWorkflowsList,
    "refreshWorkflowIndex",
    ()=>refreshWorkflowIndex,
    "saveWorkflowYaml",
    ()=>saveWorkflowYaml
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$http$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/api/http.ts [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$prefix$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/api/prefix.ts [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$demo$2d$mode$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/api/demo-mode.ts [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$mock$2f$workflow$2d$yamls$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/mock/workflow-yamls.ts [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$js$2d$yaml$2f$dist$2f$js$2d$yaml$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/js-yaml/dist/js-yaml.mjs [app-ssr] (ecmascript)");
;
;
;
;
;
function getCustomMockYamls() {
    if ("TURBOPACK compile-time truthy", 1) return {};
    //TURBOPACK unreachable
    ;
}
function getAllMockYamls() {
    const custom = getCustomMockYamls();
    return {
        ...__TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$mock$2f$workflow$2d$yamls$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["MOCK_WORKFLOW_YAMLS"],
        ...custom
    };
}
function getMockYamlEntries() {
    const out = [];
    Object.entries(__TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$mock$2f$workflow$2d$yamls$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["MOCK_WORKFLOW_YAMLS"]).forEach(([id, content])=>{
        if (typeof content !== "string" || !content.trim()) return;
        out.push({
            id,
            content,
            source: "builtin"
        });
    });
    const custom = getCustomMockYamls();
    Object.entries(custom).forEach(([id, content])=>{
        if (typeof content !== "string" || !content.trim()) return;
        out.push({
            id,
            content,
            source: "custom"
        });
    });
    return out;
}
function resolveMockYamlContent(idOrName) {
    const all = getAllMockYamls();
    const direct = all[idOrName];
    if (typeof direct === "string" && direct.trim()) return direct;
    const entries = getMockYamlEntries().slice().reverse();
    for (const { id: fallbackId, content } of entries){
        let doc = {};
        try {
            doc = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$js$2d$yaml$2f$dist$2f$js$2d$yaml$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"].load(content) || {};
        } catch  {
            doc = {};
        }
        const name = typeof doc?.name === "string" ? doc.name.trim() : "";
        if (name && name === idOrName) return content;
        if (fallbackId === idOrName) return content;
    }
    return null;
}
function getUniqueMockWorkflows() {
    const entries = getMockYamlEntries();
    const byName = new Map();
    const order = [];
    entries.forEach(({ id, content, source })=>{
        const wf = toWorkflowFromYaml(id, content);
        const key = (wf.name || "").trim() || id;
        const existing = byName.get(key);
        if (!existing) {
            byName.set(key, {
                wf,
                source
            });
            order.push(key);
            return;
        }
        if (existing.source === "builtin" && source === "custom") {
            byName.set(key, {
                wf,
                source
            });
        }
    });
    return order.map((k)=>byName.get(k).wf);
}
function normalizeTags(raw) {
    if (Array.isArray(raw)) {
        return raw.filter((t)=>typeof t === "string").map((t)=>t.trim()).filter(Boolean);
    }
    if (typeof raw === "string") {
        return raw.split(",").map((t)=>t.trim()).filter(Boolean);
    }
    return [];
}
function addMockDataTag(tags) {
    const set = new Set(tags);
    set.add("mock-data");
    return Array.from(set);
}
function getHttpErrorCode(e) {
    const msg = e instanceof Error ? e.message : "";
    const code = parseInt(msg.split(":")[0] || "0", 10);
    return Number.isFinite(code) ? code : 0;
}
function enableDemoMode() {
    if ("TURBOPACK compile-time falsy", 0) //TURBOPACK unreachable
    ;
}
function toWorkflowFromYaml(id, content) {
    let doc = {};
    try {
        doc = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$js$2d$yaml$2f$dist$2f$js$2d$yaml$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"].load(content) || {};
    } catch  {
        doc = {};
    }
    const steps = Array.isArray(doc?.steps) ? doc.steps : [];
    const modules = Array.isArray(doc?.modules) ? doc.modules : [];
    const kind = doc?.kind === "flow" ? "flow" : "module";
    const name = typeof doc?.name === "string" ? doc.name : id;
    const description = typeof doc?.description === "string" ? doc.description : "";
    const tags = addMockDataTag(normalizeTags(doc?.tags));
    const params = Array.isArray(doc?.params) ? doc.params : [];
    return {
        name,
        kind,
        description,
        tags,
        file_path: "",
        params,
        required_params: params.filter((p)=>p?.required).map((p)=>p?.name ?? ""),
        step_count: steps.length,
        module_count: modules.length,
        checksum: "",
        indexed_at: new Date().toISOString()
    };
}
function getMockWorkflowTags() {
    const tagSet = new Set();
    Object.values(getAllMockYamls()).forEach((content)=>{
        try {
            const doc = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$js$2d$yaml$2f$dist$2f$js$2d$yaml$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"].load(content) || {};
            normalizeTags(doc?.tags).forEach((t)=>tagSet.add(t));
        } catch  {}
    });
    tagSet.add("mock-data");
    return Array.from(tagSet.values()).sort();
}
async function fetchWorkflows() {
    if ((0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$demo$2d$mode$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["isDemoMode"])()) {
        return getUniqueMockWorkflows();
    }
    const res = await __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$http$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["http"].get(`${__TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$prefix$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["API_PREFIX"]}/workflows`);
    const data = res.data?.data || [];
    return data.map((w)=>({
            name: w.name ?? "",
            kind: w.kind === "flow" ? "flow" : "module",
            description: w.description ?? "",
            tags: Array.isArray(w.tags) ? w.tags : [],
            file_path: w.file_path ?? "",
            params: Array.isArray(w.params) ? w.params : [],
            required_params: Array.isArray(w.required_params) ? w.required_params : [],
            step_count: w.step_count ?? 0,
            module_count: w.module_count ?? 0,
            checksum: w.checksum ?? "",
            indexed_at: w.indexed_at ?? ""
        }));
}
async function fetchMockWorkflowsList(params = {}) {
    const all = getUniqueMockWorkflows();
    const filtered = all.filter((wf)=>{
        if (params.kind && wf.kind !== params.kind) return false;
        if (params.tags && params.tags.length > 0) {
            const tagSet = new Set((wf.tags || []).map((t)=>String(t)));
            if (!params.tags.some((t)=>tagSet.has(t))) return false;
        }
        if (params.search && params.search.trim()) {
            const q = params.search.trim().toLowerCase();
            const hay = `${wf.name ?? ""} ${wf.description ?? ""} ${(wf.tags || []).join(" ")}`.toLowerCase();
            if (!hay.includes(q)) return false;
        }
        return true;
    });
    const offset = typeof params.offset === "number" ? params.offset : 0;
    const limit = typeof params.limit === "number" ? params.limit : filtered.length;
    const paged = filtered.slice(Math.max(0, offset), Math.max(0, offset) + Math.max(0, limit));
    return {
        items: paged,
        pagination: {
            total: filtered.length,
            offset,
            limit
        }
    };
}
async function fetchWorkflowsList(params = {}) {
    if ((0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$demo$2d$mode$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["isDemoMode"])()) {
        const items = await fetchWorkflows();
        const filtered = items.filter((wf)=>{
            if (params.kind && wf.kind !== params.kind) return false;
            if (params.tags && params.tags.length > 0) {
                const tagSet = new Set((wf.tags || []).map((t)=>String(t)));
                if (!params.tags.some((t)=>tagSet.has(t))) return false;
            }
            if (params.search && params.search.trim()) {
                const q = params.search.trim().toLowerCase();
                const hay = `${wf.name ?? ""} ${wf.description ?? ""} ${(wf.tags || []).join(" ")}`.toLowerCase();
                if (!hay.includes(q)) return false;
            }
            return true;
        });
        const offset = typeof params.offset === "number" ? params.offset : 0;
        const limit = typeof params.limit === "number" ? params.limit : filtered.length;
        const paged = filtered.slice(Math.max(0, offset), Math.max(0, offset) + Math.max(0, limit));
        return {
            items: paged,
            pagination: {
                total: filtered.length,
                offset,
                limit
            }
        };
    }
    const query = {};
    if (params.source) query.source = params.source;
    if (params.tags && params.tags.length > 0) query.tags = params.tags.join(",");
    if (params.kind) query.kind = params.kind;
    if (params.search) query.search = params.search;
    if (typeof params.offset === "number") query.offset = params.offset;
    if (typeof params.limit === "number") query.limit = params.limit;
    try {
        const res = await __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$http$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["http"].get(`${__TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$prefix$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["API_PREFIX"]}/workflows`, {
            params: query
        });
        const data = res.data?.data || [];
        const pagination = res.data?.pagination || {
            total: data.length,
            offset: 0,
            limit: data.length
        };
        const items = data.map((w)=>({
                name: w.name ?? "",
                kind: w.kind === "flow" ? "flow" : "module",
                description: w.description ?? "",
                tags: Array.isArray(w.tags) ? w.tags.map((t)=>String(t)) : [],
                file_path: w.file_path ?? "",
                params: Array.isArray(w.params) ? w.params : [],
                required_params: Array.isArray(w.required_params) ? w.required_params : [],
                step_count: w.step_count ?? 0,
                module_count: w.module_count ?? 0,
                checksum: w.checksum ?? "",
                indexed_at: w.indexed_at ?? ""
            }));
        return {
            items,
            pagination: {
                total: Number(pagination.total) || items.length,
                offset: Number(pagination.offset) || 0,
                limit: Number(pagination.limit) || items.length
            }
        };
    } catch (e) {
        const code = getHttpErrorCode(e);
        if (code === 0) {
            enableDemoMode();
            return fetchMockWorkflowsList({
                kind: params.kind,
                tags: params.tags,
                search: params.search,
                offset: params.offset,
                limit: params.limit
            });
        }
        throw e;
    }
}
async function fetchWorkflow(id) {
    if ((0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$demo$2d$mode$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["isDemoMode"])()) {
        const content = resolveMockYamlContent(id);
        if (!content) return null;
        return toWorkflowFromYaml(id, content);
    }
    try {
        const res = await __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$http$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["http"].get(`${__TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$prefix$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["API_PREFIX"]}/workflows/${encodeURIComponent(id)}`, {
            params: {
                json: true
            }
        });
        const w = res.data;
        return {
            name: w.name ?? "",
            kind: w.kind === "flow" ? "flow" : "module",
            description: w.description ?? "",
            tags: Array.isArray(w.tags) ? w.tags : [],
            file_path: w.file_path ?? "",
            params: Array.isArray(w.params) ? w.params : [],
            required_params: Array.isArray(w.required_params) ? w.required_params : [],
            step_count: Array.isArray(w.steps) ? w.steps.length : w.step_count ?? 0,
            module_count: w.module_count ?? 0,
            checksum: w.checksum ?? "",
            indexed_at: w.indexed_at ?? ""
        };
    } catch (e) {
        const code = getHttpErrorCode(e);
        if (code === 404) throw new Error("WORKFLOW_NOT_FOUND");
        if (code === 401) throw new Error("UNAUTHORIZED");
        if (code === 0) {
            enableDemoMode();
            const content = resolveMockYamlContent(id);
            return content ? toWorkflowFromYaml(id, content) : null;
        }
        throw new Error("REQUEST_FAILED");
    }
}
async function fetchWorkflowYaml(id) {
    if ((0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$demo$2d$mode$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["isDemoMode"])()) {
        return resolveMockYamlContent(id);
    }
    try {
        const res = await __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$http$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["http"].get(`${__TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$prefix$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["API_PREFIX"]}/workflows/${encodeURIComponent(id)}`, {
            responseType: "text"
        });
        return typeof res.data === "string" ? res.data : res.data?.yaml ?? null;
    } catch (e) {
        const code = getHttpErrorCode(e);
        if (code === 404) throw new Error("WORKFLOW_NOT_FOUND");
        if (code === 401) throw new Error("UNAUTHORIZED");
        if (code === 0) {
            enableDemoMode();
            return resolveMockYamlContent(id);
        }
        throw new Error("REQUEST_FAILED");
    }
}
async function fetchWorkflowTags() {
    if ((0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$demo$2d$mode$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["isDemoMode"])()) {
        return getMockWorkflowTags();
    }
    try {
        const res = await __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$http$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["http"].get(`${__TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$prefix$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["API_PREFIX"]}/workflows/tags`);
        const tags = res.data?.tags || [];
        return Array.isArray(tags) ? tags.map((t)=>String(t)) : [];
    } catch (e) {
        const code = getHttpErrorCode(e);
        if (code === 0) {
            enableDemoMode();
            return getMockWorkflowTags();
        }
        throw e;
    }
}
async function refreshWorkflowIndex(force = false) {
    const res = await __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$http$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["http"].post(`${__TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$prefix$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["API_PREFIX"]}/workflows/refresh`, undefined, {
        params: force ? {
            force: true
        } : {}
    });
    return {
        message: res.data?.message || "",
        added: Number(res.data?.added || 0),
        updated: Number(res.data?.updated || 0),
        removed: Number(res.data?.removed || 0),
        errors: Array.isArray(res.data?.errors) ? res.data.errors : []
    };
}
async function saveWorkflowYaml(id, yamlText) {
    if (!id || !yamlText.trim()) return false;
    if ((0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$demo$2d$mode$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["isDemoMode"])()) {
        if ("TURBOPACK compile-time truthy", 1) return false;
        //TURBOPACK unreachable
        ;
    }
    try {
        if ("TURBOPACK compile-time truthy", 1) return false;
        //TURBOPACK unreachable
        ;
        let name;
        let kind;
        const form = undefined;
        const fileName = undefined;
        const blob = undefined;
    } catch (e) {
        const code = getHttpErrorCode(e);
        if (code === 0) {
            (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$demo$2d$mode$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["setDemoMode"])(true);
            return saveWorkflowYaml(id, yamlText);
        }
        return false;
    }
}
}),
"[project]/components/workflow-editor/workflow-editor-client.tsx [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "default",
    ()=>WorkflowEditorClient
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react-jsx-dev-runtime.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$client$2f$app$2d$dir$2f$link$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/client/app-dir/link.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/ui/button.tsx [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$badge$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/ui/badge.tsx [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$skeleton$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/ui/skeleton.tsx [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$workflow$2d$canvas$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/workflow-editor/workflow-canvas.tsx [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$workflow$2d$sidebar$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/workflow-editor/workflow-sidebar.tsx [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$utils$2f$yaml$2d$parser$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/workflow-editor/utils/yaml-parser.ts [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$shared$2f$error$2d$state$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/components/shared/error-state.tsx [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$workflows$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/api/workflows.ts [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$http$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/api/http.ts [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$navigation$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/navigation.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$sonner$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/sonner/dist/index.mjs [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$arrow$2d$left$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__ArrowLeftIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/arrow-left.js [app-ssr] (ecmascript) <export default as ArrowLeftIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$loader$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__LoaderIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/loader.js [app-ssr] (ecmascript) <export default as LoaderIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$arrow$2d$up$2d$down$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__ArrowUpDownIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/arrow-up-down.js [app-ssr] (ecmascript) <export default as ArrowUpDownIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$arrow$2d$left$2d$right$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__ArrowLeftRightIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/arrow-left-right.js [app-ssr] (ecmascript) <export default as ArrowLeftRightIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$clipboard$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__ClipboardIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/clipboard.js [app-ssr] (ecmascript) <export default as ClipboardIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$text$2d$align$2d$justify$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__AlignJustifyIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/text-align-justify.js [app-ssr] (ecmascript) <export default as AlignJustifyIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$map$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__MapIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/map.js [app-ssr] (ecmascript) <export default as MapIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$save$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__SaveIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/save.js [app-ssr] (ecmascript) <export default as SaveIcon>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$eye$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__EyeIcon$3e$__ = __turbopack_context__.i("[project]/node_modules/lucide-react/dist/esm/icons/eye.js [app-ssr] (ecmascript) <export default as EyeIcon>");
"use client";
;
;
;
;
;
;
;
;
;
;
;
;
;
;
;
function WorkflowEditorClient({ workflowId }) {
    const pathname = (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$navigation$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["usePathname"])();
    const splitRef = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useRef"](null);
    const canvasApiRef = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useRef"](null);
    const pendingFocusIdRef = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useRef"](null);
    const effectiveId = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useMemo"](()=>{
        if (workflowId && workflowId.length > 0) return workflowId;
        const seg = pathname?.split("/").filter(Boolean).pop() || "";
        return decodeURIComponent(seg);
    }, [
        workflowId,
        pathname
    ]);
    const [workflow, setWorkflow] = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useState"](null);
    const [parsedWorkflow, setParsedWorkflow] = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useState"](null);
    const [workflowData, setWorkflowData] = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useState"](null);
    const [yamlPreview, setYamlPreview] = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useState"]("");
    const [isLoading, setIsLoading] = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useState"](true);
    const [isSaving, setIsSaving] = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useState"](false);
    const [error, setError] = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useState"](null);
    const [selectedStepName, setSelectedStepName] = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useState"](null);
    const [orientation, setOrientation] = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useState"]("TB");
    const baseURL = (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$http$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["getHttpBaseURL"])();
    const [sidebarWidth, setSidebarWidth] = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useState"](320);
    const [wrapCanvasText, setWrapCanvasText] = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useState"](true);
    const [showCanvasDetails, setShowCanvasDetails] = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useState"](true);
    const [hideMiniMap, setHideMiniMap] = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useState"](true);
    const toolbarButtonClassName = "border-sky-400 text-sky-700 hover:bg-sky-500/10 dark:border-sky-300 dark:text-sky-200 dark:hover:bg-sky-300/15";
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useEffect"](()=>{
        if ("TURBOPACK compile-time truthy", 1) return;
        //TURBOPACK unreachable
        ;
        const raw = undefined;
        const n = undefined;
    }, []);
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useEffect"](()=>{
        if ("TURBOPACK compile-time truthy", 1) return;
        //TURBOPACK unreachable
        ;
        const raw = undefined;
    }, []);
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useEffect"](()=>{
        if ("TURBOPACK compile-time truthy", 1) return;
        //TURBOPACK unreachable
        ;
        const raw = undefined;
    }, []);
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useEffect"](()=>{
        if ("TURBOPACK compile-time truthy", 1) return;
        //TURBOPACK unreachable
        ;
        const raw = undefined;
    }, []);
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useEffect"](()=>{
        if ("TURBOPACK compile-time truthy", 1) return;
        //TURBOPACK unreachable
        ;
    }, [
        sidebarWidth
    ]);
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useEffect"](()=>{
        if ("TURBOPACK compile-time truthy", 1) return;
        //TURBOPACK unreachable
        ;
    }, [
        wrapCanvasText
    ]);
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useEffect"](()=>{
        if ("TURBOPACK compile-time truthy", 1) return;
        //TURBOPACK unreachable
        ;
    }, [
        showCanvasDetails
    ]);
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useEffect"](()=>{
        if ("TURBOPACK compile-time truthy", 1) return;
        //TURBOPACK unreachable
        ;
    }, [
        hideMiniMap
    ]);
    const loadWorkflow = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useCallback"](async ()=>{
        try {
            setIsLoading(true);
            setError(null);
            const [wf, yaml] = await Promise.all([
                (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$workflows$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["fetchWorkflow"])(effectiveId),
                (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$workflows$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["fetchWorkflowYaml"])(effectiveId)
            ]);
            if (!wf || !yaml) {
                setError(`Workflow not found: ${effectiveId}`);
                return;
            }
            setWorkflow(wf);
            const parsed = (0, __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$utils$2f$yaml$2d$parser$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["parseWorkflowYaml"])(yaml);
            setParsedWorkflow(parsed);
            setWorkflowData(parsed.raw);
            setYamlPreview(yaml);
        } catch (err) {
            const msg = err instanceof Error ? err.message : "";
            if (msg === "WORKFLOW_NOT_FOUND") {
                setError(`Workflow not found: ${effectiveId}`);
            } else if (msg === "NETWORK_ERROR") {
                setError(`Cannot reach API at ${baseURL}`);
            } else if (msg === "UNAUTHORIZED") {
                setError("Session expired. Please log in.");
            } else {
                setError("Failed to load workflow");
            }
        } finally{
            setIsLoading(false);
        }
    }, [
        effectiveId
    ]);
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useEffect"](()=>{
        loadWorkflow();
    }, [
        loadWorkflow
    ]);
    const selectedStep = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useMemo"](()=>{
        if (!selectedStepName || !workflowData) return null;
        if (workflowData.kind !== "module") return null;
        return workflowData.steps.find((s)=>s.name === selectedStepName) ?? null;
    }, [
        selectedStepName,
        workflowData
    ]);
    const selectedModule = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useMemo"](()=>{
        if (!selectedStepName || !workflowData) return null;
        if (workflowData.kind !== "flow") return null;
        const modules = workflowData.modules;
        if (!Array.isArray(modules)) return null;
        return modules.find((m)=>m.name === selectedStepName) ?? null;
    }, [
        selectedStepName,
        workflowData
    ]);
    const allSteps = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useMemo"](()=>{
        if (!workflowData || workflowData.kind !== "module") return [];
        return workflowData.steps ?? [];
    }, [
        workflowData
    ]);
    const allModules = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useMemo"](()=>{
        if (!workflowData || workflowData.kind !== "flow") return [];
        return workflowData.modules ?? [];
    }, [
        workflowData
    ]);
    const handleNavigateToNode = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useCallback"]((nodeId)=>{
        pendingFocusIdRef.current = nodeId;
        setSelectedStepName(nodeId);
        requestAnimationFrame(()=>{
            canvasApiRef.current?.focusNode(nodeId);
        });
    }, []);
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useEffect"](()=>{
        const id = pendingFocusIdRef.current;
        if (!id) return;
        if (id !== selectedStepName) return;
        pendingFocusIdRef.current = null;
        requestAnimationFrame(()=>{
            requestAnimationFrame(()=>{
                canvasApiRef.current?.focusNode(id);
            });
        });
    }, [
        selectedStepName
    ]);
    const handleNodeSelect = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useCallback"]((nodeId)=>{
        if (!nodeId || nodeId === "_start" || nodeId === "_end") {
            setSelectedStepName(null);
        } else {
            setSelectedStepName(nodeId);
        }
    }, []);
    const handleStepUpdate = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useCallback"]((stepName, updates)=>{
        if (!workflowData) return;
        if (workflowData.kind !== "module") return;
        const updatedWorkflow = (0, __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$utils$2f$yaml$2d$parser$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["updateStepInWorkflow"])(workflowData, stepName, updates);
        setWorkflowData(updatedWorkflow);
        const newYaml = (0, __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$utils$2f$yaml$2d$parser$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["serializeWorkflowToYaml"])(updatedWorkflow);
        setYamlPreview(newYaml);
        const parsed = (0, __TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$utils$2f$yaml$2d$parser$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["parseWorkflowYaml"])(newYaml);
        setParsedWorkflow(parsed);
    }, [
        workflowData
    ]);
    if (error) {
        return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
            className: "flex h-[calc(100vh-10rem)] items-center justify-center",
            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                className: "space-y-4 text-center",
                children: [
                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$shared$2f$error$2d$state$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["ErrorState"], {
                        title: "Workflow Error",
                        message: error,
                        onRetry: loadWorkflow
                    }, void 0, false, {
                        fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                        lineNumber: 217,
                        columnNumber: 11
                    }, this),
                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                        className: "flex items-center justify-center gap-2",
                        children: [
                            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Button"], {
                                variant: "outline",
                                asChild: true,
                                children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$client$2f$app$2d$dir$2f$link$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"], {
                                    href: "/workflows",
                                    children: "Back to Workflows"
                                }, void 0, false, {
                                    fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                    lineNumber: 220,
                                    columnNumber: 15
                                }, this)
                            }, void 0, false, {
                                fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                lineNumber: 219,
                                columnNumber: 13
                            }, this),
                            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Button"], {
                                variant: "outline",
                                asChild: true,
                                children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$client$2f$app$2d$dir$2f$link$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"], {
                                    href: "/settings",
                                    children: "Settings"
                                }, void 0, false, {
                                    fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                    lineNumber: 223,
                                    columnNumber: 15
                                }, this)
                            }, void 0, false, {
                                fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                lineNumber: 222,
                                columnNumber: 13
                            }, this)
                        ]
                    }, void 0, true, {
                        fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                        lineNumber: 218,
                        columnNumber: 11
                    }, this)
                ]
            }, void 0, true, {
                fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                lineNumber: 216,
                columnNumber: 9
            }, this)
        }, void 0, false, {
            fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
            lineNumber: 215,
            columnNumber: 7
        }, this);
    }
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
        className: "flex h-[calc(100vh-7rem)] flex-col",
        children: [
            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                className: "flex items-center justify-between border-b px-4 py-3",
                children: [
                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                        className: "flex items-center gap-4",
                        children: [
                            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Button"], {
                                variant: "outline",
                                size: "icon",
                                asChild: true,
                                children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$client$2f$app$2d$dir$2f$link$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"], {
                                    href: "/workflows",
                                    children: [
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$arrow$2d$left$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__ArrowLeftIcon$3e$__["ArrowLeftIcon"], {
                                            className: "size-4"
                                        }, void 0, false, {
                                            fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                            lineNumber: 237,
                                            columnNumber: 15
                                        }, this),
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("span", {
                                            className: "sr-only",
                                            children: "Back to workflows"
                                        }, void 0, false, {
                                            fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                            lineNumber: 238,
                                            columnNumber: 15
                                        }, this)
                                    ]
                                }, void 0, true, {
                                    fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                    lineNumber: 236,
                                    columnNumber: 13
                                }, this)
                            }, void 0, false, {
                                fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                lineNumber: 235,
                                columnNumber: 11
                            }, this),
                            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                children: isLoading ? /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                    className: "space-y-1",
                                    children: [
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$skeleton$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Skeleton"], {
                                            className: "h-6 w-40"
                                        }, void 0, false, {
                                            fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                            lineNumber: 244,
                                            columnNumber: 17
                                        }, this),
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$skeleton$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Skeleton"], {
                                            className: "h-4 w-24"
                                        }, void 0, false, {
                                            fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                            lineNumber: 245,
                                            columnNumber: 17
                                        }, this)
                                    ]
                                }, void 0, true, {
                                    fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                    lineNumber: 243,
                                    columnNumber: 15
                                }, this) : workflow ? /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Fragment"], {
                                    children: [
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                            className: "flex items-center gap-2",
                                            children: [
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("h1", {
                                                    className: "text-lg font-semibold",
                                                    children: workflow.name
                                                }, void 0, false, {
                                                    fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                                    lineNumber: 250,
                                                    columnNumber: 19
                                                }, this),
                                                /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$badge$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Badge"], {
                                                    variant: "secondary",
                                                    className: "capitalize",
                                                    children: workflow.kind
                                                }, void 0, false, {
                                                    fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                                    lineNumber: 251,
                                                    columnNumber: 19
                                                }, this)
                                            ]
                                        }, void 0, true, {
                                            fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                            lineNumber: 249,
                                            columnNumber: 17
                                        }, this),
                                        /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("p", {
                                            className: "text-sm text-muted-foreground",
                                            children: workflow.description
                                        }, void 0, false, {
                                            fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                            lineNumber: 255,
                                            columnNumber: 17
                                        }, this)
                                    ]
                                }, void 0, true) : null
                            }, void 0, false, {
                                fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                lineNumber: 241,
                                columnNumber: 11
                            }, this)
                        ]
                    }, void 0, true, {
                        fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                        lineNumber: 234,
                        columnNumber: 9
                    }, this),
                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                        className: "flex items-center gap-2",
                        children: [
                            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Button"], {
                                variant: "outline",
                                size: "sm",
                                onClick: ()=>setOrientation((prev)=>prev === "TB" ? "LR" : "TB"),
                                className: toolbarButtonClassName,
                                children: [
                                    orientation === "TB" ? /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$arrow$2d$up$2d$down$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__ArrowUpDownIcon$3e$__["ArrowUpDownIcon"], {
                                        className: "mr-2 size-4"
                                    }, void 0, false, {
                                        fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                        lineNumber: 271,
                                        columnNumber: 15
                                    }, this) : /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$arrow$2d$left$2d$right$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__ArrowLeftRightIcon$3e$__["ArrowLeftRightIcon"], {
                                        className: "mr-2 size-4"
                                    }, void 0, false, {
                                        fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                        lineNumber: 273,
                                        columnNumber: 15
                                    }, this),
                                    orientation === "TB" ? "Vertical" : "Horizontal"
                                ]
                            }, void 0, true, {
                                fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                lineNumber: 264,
                                columnNumber: 11
                            }, this),
                            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Button"], {
                                variant: "outline",
                                size: "sm",
                                onClick: ()=>setWrapCanvasText((v)=>!v),
                                className: toolbarButtonClassName,
                                children: [
                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$text$2d$align$2d$justify$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__AlignJustifyIcon$3e$__["AlignJustifyIcon"], {
                                        className: "mr-2 size-4"
                                    }, void 0, false, {
                                        fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                        lineNumber: 283,
                                        columnNumber: 13
                                    }, this),
                                    wrapCanvasText ? "Wrap lines on" : "Wrap lines off"
                                ]
                            }, void 0, true, {
                                fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                lineNumber: 277,
                                columnNumber: 11
                            }, this),
                            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Button"], {
                                variant: "outline",
                                size: "sm",
                                onClick: ()=>setShowCanvasDetails((v)=>!v),
                                className: toolbarButtonClassName,
                                children: [
                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$eye$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__EyeIcon$3e$__["EyeIcon"], {
                                        className: "mr-2 size-4"
                                    }, void 0, false, {
                                        fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                        lineNumber: 292,
                                        columnNumber: 13
                                    }, this),
                                    showCanvasDetails ? "Details on" : "Details off"
                                ]
                            }, void 0, true, {
                                fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                lineNumber: 286,
                                columnNumber: 11
                            }, this),
                            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Button"], {
                                variant: "outline",
                                size: "sm",
                                onClick: ()=>setHideMiniMap((v)=>!v),
                                className: toolbarButtonClassName,
                                children: [
                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$map$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__MapIcon$3e$__["MapIcon"], {
                                        className: "mr-2 size-4"
                                    }, void 0, false, {
                                        fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                        lineNumber: 301,
                                        columnNumber: 13
                                    }, this),
                                    hideMiniMap ? "Minimap off" : "Minimap on"
                                ]
                            }, void 0, true, {
                                fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                lineNumber: 295,
                                columnNumber: 11
                            }, this),
                            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Button"], {
                                variant: "outline",
                                size: "sm",
                                disabled: !yamlPreview,
                                onClick: async ()=>{
                                    try {
                                        await navigator.clipboard.writeText(yamlPreview);
                                        __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$sonner$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["toast"].success("Copied to clipboard");
                                    } catch  {
                                        __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$sonner$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["toast"].error("Failed to copy");
                                    }
                                },
                                children: [
                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$clipboard$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__ClipboardIcon$3e$__["ClipboardIcon"], {
                                        className: "mr-2 size-4"
                                    }, void 0, false, {
                                        fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                        lineNumber: 317,
                                        columnNumber: 13
                                    }, this),
                                    "Copy YAML"
                                ]
                            }, void 0, true, {
                                fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                lineNumber: 304,
                                columnNumber: 11
                            }, this),
                            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$ui$2f$button$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["Button"], {
                                variant: "outline",
                                size: "sm",
                                disabled: !yamlPreview || isSaving || isLoading,
                                onClick: async ()=>{
                                    if (!yamlPreview) return;
                                    setIsSaving(true);
                                    try {
                                        const ok = await (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$workflows$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["saveWorkflowYaml"])(effectiveId, yamlPreview);
                                        if (ok) {
                                            __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$sonner$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["toast"].success("Saved");
                                        } else {
                                            __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$sonner$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["toast"].error("Save failed");
                                        }
                                    } catch  {
                                        __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$sonner$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["toast"].error("Save failed");
                                    } finally{
                                        setIsSaving(false);
                                    }
                                },
                                children: [
                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$save$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__SaveIcon$3e$__["SaveIcon"], {
                                        className: "mr-2 size-4"
                                    }, void 0, false, {
                                        fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                        lineNumber: 341,
                                        columnNumber: 13
                                    }, this),
                                    isSaving ? "Saving..." : "Save"
                                ]
                            }, void 0, true, {
                                fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                lineNumber: 320,
                                columnNumber: 11
                            }, this)
                        ]
                    }, void 0, true, {
                        fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                        lineNumber: 263,
                        columnNumber: 9
                    }, this)
                ]
            }, void 0, true, {
                fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                lineNumber: 233,
                columnNumber: 7
            }, this),
            /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                ref: splitRef,
                className: "flex flex-1 overflow-hidden",
                children: [
                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                        className: "flex-1",
                        children: isLoading ? /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                            className: "flex h-full items-center justify-center",
                            children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                                className: "flex flex-col items-center gap-4",
                                children: [
                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$lucide$2d$react$2f$dist$2f$esm$2f$icons$2f$loader$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__$3c$export__default__as__LoaderIcon$3e$__["LoaderIcon"], {
                                        className: "size-8 animate-spin text-muted-foreground"
                                    }, void 0, false, {
                                        fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                        lineNumber: 352,
                                        columnNumber: 17
                                    }, this),
                                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("p", {
                                        className: "text-sm text-muted-foreground",
                                        children: "Loading workflow..."
                                    }, void 0, false, {
                                        fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                        lineNumber: 353,
                                        columnNumber: 17
                                    }, this)
                                ]
                            }, void 0, true, {
                                fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                                lineNumber: 351,
                                columnNumber: 15
                            }, this)
                        }, void 0, false, {
                            fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                            lineNumber: 350,
                            columnNumber: 13
                        }, this) : parsedWorkflow ? /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$workflow$2d$canvas$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["WorkflowCanvas"], {
                            initialNodes: parsedWorkflow.nodes,
                            initialEdges: parsedWorkflow.edges,
                            onNodeSelect: handleNodeSelect,
                            orientation: orientation,
                            wrapLongText: wrapCanvasText,
                            showDetails: showCanvasDetails,
                            hideMiniMap: hideMiniMap,
                            selectedNodeId: selectedStepName,
                            onCanvasReady: (api)=>{
                                canvasApiRef.current = api;
                                if (pendingFocusIdRef.current) {
                                    api.focusNode(pendingFocusIdRef.current);
                                }
                            }
                        }, void 0, false, {
                            fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                            lineNumber: 357,
                            columnNumber: 13
                        }, this) : null
                    }, void 0, false, {
                        fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                        lineNumber: 348,
                        columnNumber: 9
                    }, this),
                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                        role: "separator",
                        "aria-orientation": "vertical",
                        tabIndex: 0,
                        className: "w-1 cursor-col-resize bg-border/60 hover:bg-border focus:outline-none focus-visible:ring-2 focus-visible:ring-ring",
                        onPointerDown: (e)=>{
                            const startX = e.clientX;
                            const startWidth = sidebarWidth;
                            e.currentTarget.setPointerCapture(e.pointerId);
                            const handleMove = (ev)=>{
                                const rect = splitRef.current?.getBoundingClientRect();
                                const containerWidth = rect?.width ?? 0;
                                const deltaX = ev.clientX - startX;
                                const maxWidth = Math.max(260, containerWidth - 200);
                                const next = Math.min(maxWidth, Math.max(260, startWidth - deltaX));
                                setSidebarWidth(next);
                            };
                            const handleUp = (ev)=>{
                                window.removeEventListener("pointermove", handleMove);
                                window.removeEventListener("pointerup", handleUp);
                            };
                            window.addEventListener("pointermove", handleMove);
                            window.addEventListener("pointerup", handleUp);
                        },
                        onKeyDown: (e)=>{
                            const step = e.shiftKey ? 40 : 20;
                            if (e.key === "ArrowLeft") setSidebarWidth((w)=>Math.min(w + step, 900));
                            if (e.key === "ArrowRight") setSidebarWidth((w)=>Math.max(260, w - step));
                        }
                    }, void 0, false, {
                        fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                        lineNumber: 376,
                        columnNumber: 9
                    }, this),
                    /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])("div", {
                        className: "shrink-0",
                        style: {
                            width: sidebarWidth
                        },
                        children: /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$components$2f$workflow$2d$editor$2f$workflow$2d$sidebar$2e$tsx__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["WorkflowSidebar"], {
                            selectedStep: selectedStep,
                            selectedModule: selectedModule,
                            yamlPreview: yamlPreview,
                            wrapLongText: wrapCanvasText,
                            onStepUpdate: handleStepUpdate,
                            workflowKind: workflowData?.kind ?? null,
                            allSteps: allSteps,
                            allModules: allModules,
                            onNavigateToNode: handleNavigateToNode
                        }, void 0, false, {
                            fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                            lineNumber: 411,
                            columnNumber: 11
                        }, this)
                    }, void 0, false, {
                        fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                        lineNumber: 410,
                        columnNumber: 9
                    }, this)
                ]
            }, void 0, true, {
                fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
                lineNumber: 347,
                columnNumber: 7
            }, this)
        ]
    }, void 0, true, {
        fileName: "[project]/components/workflow-editor/workflow-editor-client.tsx",
        lineNumber: 232,
        columnNumber: 5
    }, this);
}
}),
];

//# sourceMappingURL=_3c5a6583._.js.map