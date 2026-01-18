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
"[project]/lib/mock/data/settings-yaml.ts [app-route] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "mockSettingsYaml",
    ()=>mockSettingsYaml
]);
const mockSettingsYaml = `# =============================================================================
# Osmedeus Configuration File
# =============================================================================

base_folder: ~/osmedeus-base

# -----------------------------------------------------------------------------
# Environment Configuration
# -----------------------------------------------------------------------------
environments:
  binaries_path: "{{base_folder}}/binaries"
  data_path: "{{base_folder}}/data"
  workflows_path: "{{base_folder}}/workflows"
  storages_path: "{{base_folder}}/storages"
  clouds_path: "{{base_folder}}/clouds"
  logs_path: "{{base_folder}}/logs"

# -----------------------------------------------------------------------------
# Server Configuration
# -----------------------------------------------------------------------------
server:
  host: "0.0.0.0"
  port: 8002
  enable_cors: true
  cors_origins:
    - "http://localhost:3000"
    - "http://127.0.0.1:3000"
  workspace_prefix_key: "[REDACTED]"
  simple_user_map_key: "[REDACTED]"
  jwt:
    secret_signing_key: "[REDACTED]"
    expiration_minutes: 180
    refresh_expiration_days: 7

# -----------------------------------------------------------------------------
# Database Configuration
# -----------------------------------------------------------------------------
database:
  host: "localhost"
  port: 5432
  name: "osmedeus"
  username: "[REDACTED]"
  password: "[REDACTED]"
  ssl_mode: "disable"
  max_idle_connections: 10
  max_open_connections: 100
  connection_max_lifetime: 3600

# -----------------------------------------------------------------------------
# Notification Configuration
# -----------------------------------------------------------------------------
notifications:
  slack:
    webhook_url: "[REDACTED]"
    channel: "#security-alerts"
    enabled: false
  discord:
    webhook_url: "[REDACTED]"
    enabled: false
  telegram:
    bot_token: "[REDACTED]"
    chat_id: "[REDACTED]"
    enabled: false

# -----------------------------------------------------------------------------
# Cloud Provider Configuration
# -----------------------------------------------------------------------------
cloud:
  provider: "digitalocean"
  digitalocean:
    api_token: "[REDACTED]"
    region: "nyc1"
    size: "s-2vcpu-4gb"
    image: "ubuntu-22-04-x64"
  aws:
    access_key_id: "[REDACTED]"
    secret_access_key: "[REDACTED]"
    region: "us-east-1"
  linode:
    api_token: "[REDACTED]"
    region: "us-east"

# -----------------------------------------------------------------------------
# Scan Configuration
# -----------------------------------------------------------------------------
scan:
  default_workflow: "general"
  max_concurrent_scans: 5
  timeout_minutes: 1440
  retry_failed_modules: true
  max_retries: 3

# -----------------------------------------------------------------------------
# LLM Configuration
# -----------------------------------------------------------------------------
llm:
  provider: "openai"
  openai:
    api_key: "[REDACTED]"
    model: "gpt-4"
    max_tokens: 4096
  anthropic:
    api_key: "[REDACTED]"
    model: "claude-3-sonnet-20240229"

# -----------------------------------------------------------------------------
# Logging Configuration
# -----------------------------------------------------------------------------
logging:
  level: "info"
  format: "json"
  output: "stdout"
  file_path: "{{base_folder}}/logs/osmedeus.log"
  max_size_mb: 100
  max_backups: 5
  max_age_days: 30
`;
}),
"[project]/app/api/mock/api/settings/yaml/route.ts [app-route] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "GET",
    ()=>GET,
    "PUT",
    ()=>PUT,
    "dynamic",
    ()=>dynamic
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$server$2e$js__$5b$app$2d$route$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/server.js [app-route] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$mock$2f$data$2f$settings$2d$yaml$2e$ts__$5b$app$2d$route$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/mock/data/settings-yaml.ts [app-route] (ecmascript)");
;
;
const dynamic = "force-static";
async function GET() {
    return new __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$server$2e$js__$5b$app$2d$route$5d$__$28$ecmascript$29$__["NextResponse"](__TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$mock$2f$data$2f$settings$2d$yaml$2e$ts__$5b$app$2d$route$5d$__$28$ecmascript$29$__["mockSettingsYaml"], {
        status: 200,
        headers: {
            "Content-Type": "text/yaml"
        }
    });
}
async function PUT(request) {
    try {
        const yaml = await request.text();
        if (!yaml || !yaml.trim()) {
            return __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$server$2e$js__$5b$app$2d$route$5d$__$28$ecmascript$29$__["NextResponse"].json({
                error: true,
                message: "YAML content cannot be empty"
            }, {
                status: 400
            });
        }
        // In a real implementation, this would validate and save the YAML
        // For mock purposes, we just return a success response
        return __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$server$2e$js__$5b$app$2d$route$5d$__$28$ecmascript$29$__["NextResponse"].json({
            message: "Configuration updated successfully",
            path: "/home/user/osmedeus-base/osm-settings.yaml",
            backup: "/home/user/osmedeus-base/osm-settings.yaml.backup"
        });
    } catch  {
        return __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$server$2e$js__$5b$app$2d$route$5d$__$28$ecmascript$29$__["NextResponse"].json({
            error: true,
            message: "Invalid YAML configuration"
        }, {
            status: 400
        });
    }
}
}),
];

//# sourceMappingURL=%5Broot-of-the-server%5D__cde62c81._.js.map