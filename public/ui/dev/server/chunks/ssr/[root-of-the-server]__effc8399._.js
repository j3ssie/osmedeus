module.exports = [
"[externals]/next/dist/compiled/next-server/app-page-turbo.runtime.dev.js [external] (next/dist/compiled/next-server/app-page-turbo.runtime.dev.js, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("next/dist/compiled/next-server/app-page-turbo.runtime.dev.js", () => require("next/dist/compiled/next-server/app-page-turbo.runtime.dev.js"));

module.exports = mod;
}),
"[project]/providers/theme-provider.tsx [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "ThemeProvider",
    ()=>ThemeProvider
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react-jsx-dev-runtime.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2d$themes$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next-themes/dist/index.mjs [app-ssr] (ecmascript)");
"use client";
;
;
function ThemeProvider({ children, ...props }) {
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(__TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2d$themes$2f$dist$2f$index$2e$mjs__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["ThemeProvider"], {
        ...props,
        children: children
    }, void 0, false, {
        fileName: "[project]/providers/theme-provider.tsx",
        lineNumber: 10,
        columnNumber: 10
    }, this);
}
}),
"[project]/config/theme.ts [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "defaultThemeState",
    ()=>defaultThemeState
]);
const defaultThemeState = {
    light: {
        background: "hsl(65, 35.6%, 88.4%)",
        foreground: "hsl(48 19.6078% 20%)",
        card: "hsl(48 33.3333% 97.0588%)",
        "card-foreground": "hsl(60 2.5641% 7.6471%)",
        popover: "hsl(0 0% 100%)",
        "popover-foreground": "hsl(50.7692 19.4030% 13.1373%)",
        primary: "hsl(247.8798 68.2873% 51.0713%)",
        "primary-foreground": "hsl(0 0% 100%)",
        secondary: "hsl(46.1538 22.8070% 88.8235%)",
        "secondary-foreground": "hsl(50.7692 8.4967% 30.0000%)",
        muted: "hsl(44.0000 29.4118% 90%)",
        "muted-foreground": "hsl(50.0000 2.3622% 50.1961%)",
        accent: "hsl(46.1538 22.8070% 88.8235%)",
        "accent-foreground": "hsl(50.7692 19.4030% 13.1373%)",
        destructive: "hsl(0 84.2365% 60.1961%)",
        "destructive-foreground": "hsl(0 0% 100%)",
        border: "hsl(50 7.5000% 84.3137%)",
        input: "hsl(50.7692 7.9755% 68.0392%)",
        ring: "hsl(247.8798 68.2873% 51.0713%)",
        "chart-1": "hsl(18.2813 57.1429% 43.9216%)",
        "chart-2": "hsl(251.4545 84.6154% 74.5098%)",
        "chart-3": "hsl(46.1538 28.2609% 81.9608%)",
        "chart-4": "hsl(256.5517 49.1525% 88.4314%)",
        "chart-5": "hsl(17.7778 60% 44.1176%)",
        sidebar: "hsl(51.4286 25.9259% 94.7059%)",
        "sidebar-foreground": "hsl(60 2.5210% 23.3333%)",
        "sidebar-primary": "hsl(247.8798 68.2873% 51.0713%)",
        "sidebar-primary-foreground": "hsl(0 0% 98.4314%)",
        "sidebar-accent": "hsl(46.1538 22.8070% 88.8235%)",
        "sidebar-accent-foreground": "hsl(0 0% 20.3922%)",
        "sidebar-border": "hsl(50 7.5000% 84.3137%)",
        "sidebar-ring": "hsl(247.8798 68.2873% 51.0713%)",
        "font-sans": "ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, 'Noto Sans', sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol', 'Noto Color Emoji'",
        "font-serif": 'ui-serif, Georgia, Cambria, "Times New Roman", Times, serif',
        "font-mono": 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
        radius: "0.5rem",
        "shadow-2xs": "0 1px 3px 0px hsl(0 0% 0% / 0.05)",
        "shadow-xs": "0 1px 3px 0px hsl(0 0% 0% / 0.05)",
        "shadow-sm": "0 1px 3px 0px hsl(0 0% 0% / 0.10), 0 1px 2px -1px hsl(0 0% 0% / 0.10)",
        shadow: "0 1px 3px 0px hsl(0 0% 0% / 0.10), 0 1px 2px -1px hsl(0 0% 0% / 0.10)",
        "shadow-md": "0 1px 3px 0px hsl(0 0% 0% / 0.10), 0 2px 4px -1px hsl(0 0% 0% / 0.10)",
        "shadow-lg": "0 1px 3px 0px hsl(0 0% 0% / 0.10), 0 4px 6px -1px hsl(0 0% 0% / 0.10)",
        "shadow-xl": "0 1px 3px 0px hsl(0 0% 0% / 0.10), 0 8px 10px -1px hsl(0 0% 0% / 0.10)",
        "shadow-2xl": "0 1px 3px 0px hsl(0 0% 0% / 0.25)"
    },
    dark: {
        background: "hsl(180,2%,10%)",
        foreground: "hsl(46.1538 9.7744% 73.9216%)",
        card: "hsl(60 2.7027% 14.5098%)",
        "card-foreground": "hsl(48 33.3333% 97.0588%)",
        popover: "hsl(60 2.1277% 18.4314%)",
        "popover-foreground": "hsl(60 5.4545% 89.2157%)",
        primary: "hsl(142.1569 71% 29%)",
        "primary-foreground": "hsl(0 0% 100%)",
        secondary: "hsl(48 33.3333% 97.0588%)",
        "secondary-foreground": "hsl(60 2.1277% 18.4314%)",
        muted: "hsl(60 3.8462% 10.1961%)",
        "muted-foreground": "hsl(51.4286 8.8608% 69.0196%)",
        accent: "hsl(48 10.6383% 9.2157%)",
        "accent-foreground": "hsl(51.4286 25.9259% 94.7059%)",
        destructive: "hsl(0 84.2365% 60.1961%)",
        "destructive-foreground": "hsl(0 0% 100%)",
        border: "hsl(60 5.0847% 23.1373%)",
        input: "hsl(52.5000 5.1282% 30.5882%)",
        ring: "hsl(116.2500 62.7451% 60%)",
        "chart-1": "hsl(18.2813 57.1429% 43.9216%)",
        "chart-2": "hsl(251.4545 84.6154% 74.5098%)",
        "chart-3": "hsl(48 10.6383% 9.2157%)",
        "chart-4": "hsl(248.2759 25.2174% 22.5490%)",
        "chart-5": "hsl(17.7778 60% 44.1176%)",
        sidebar: "hsl(30 3.3333% 11.7647%)",
        "sidebar-foreground": "hsl(46.1538 9.7744% 73.9216%)",
        "sidebar-primary": "hsl(116.2500 62.7451% 60%)",
        "sidebar-primary-foreground": "hsl(0 0% 98.4314%)",
        "sidebar-accent": "hsl(60 3.4483% 5.6863%)",
        "sidebar-accent-foreground": "hsl(46.1538 9.7744% 73.9216%)",
        "sidebar-border": "hsl(60 5.0847% 23.1373%)",
        "sidebar-ring": "hsl(116.2500 62.7451% 60%)",
        "shadow-2xs": "0 1px 3px 0px hsl(0 0% 0% / 0.05)",
        "shadow-xs": "0 1px 3px 0px hsl(0 0% 0% / 0.05)",
        "shadow-sm": "0 1px 3px 0px hsl(0 0% 0% / 0.10), 0 1px 2px -1px hsl(0 0% 0% / 0.10)",
        shadow: "0 1px 3px 0px hsl(0 0% 0% / 0.10), 0 1px 2px -1px hsl(0 0% 0% / 0.10)",
        "shadow-md": "0 1px 3px 0px hsl(0 0% 0% / 0.10), 0 2px 4px -1px hsl(0 0% 0% / 0.10)",
        "shadow-lg": "0 1px 3px 0px hsl(0 0% 0% / 0.10), 0 4px 6px -1px hsl(0 0% 0% / 0.10)",
        "shadow-xl": "0 1px 3px 0px hsl(0 0% 0% / 0.10), 0 8px 10px -1px hsl(0 0% 0% / 0.10)",
        "shadow-2xl": "0 1px 3px 0px hsl(0 0% 0% / 0.25)"
    },
    css: {}
};
}),
"[project]/theme-presets.ts [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

// Type Imports
__turbopack_context__.s([
    "getPresetThemeStyles",
    ()=>getPresetThemeStyles,
    "presets",
    ()=>presets
]);
// Config Imports
var __TURBOPACK__imported__module__$5b$project$5d2f$config$2f$theme$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/config/theme.ts [app-ssr] (ecmascript)");
;
function getPresetThemeStyles(name) {
    if (name === 'default') {
        return __TURBOPACK__imported__module__$5b$project$5d2f$config$2f$theme$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["defaultThemeState"];
    }
    const preset = presets[name];
    if (!preset) {
        return __TURBOPACK__imported__module__$5b$project$5d2f$config$2f$theme$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["defaultThemeState"];
    }
    return {
        light: {
            ...__TURBOPACK__imported__module__$5b$project$5d2f$config$2f$theme$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["defaultThemeState"].light,
            ...preset.light || {}
        },
        dark: {
            ...__TURBOPACK__imported__module__$5b$project$5d2f$config$2f$theme$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["defaultThemeState"].dark,
            ...preset.dark || {}
        },
        css: preset.css || {}
    };
}
const presets = {
    marshmallow: {
        light: {
            background: 'oklch(0.97 0.01 264.53)',
            foreground: 'oklch(0.22 0 0)',
            card: 'oklch(1.00 0 0)',
            'card-foreground': 'oklch(0.22 0 0)',
            popover: 'oklch(1.00 0 0)',
            'popover-foreground': 'oklch(0.22 0 0)',
            primary: 'oklch(0.80 0.14 349.25)',
            'primary-foreground': 'oklch(0 0 0)',
            secondary: 'oklch(0.94 0.07 98.08)',
            'secondary-foreground': 'oklch(0 0 0)',
            muted: 'oklch(0.92 0.01 268.52)',
            'muted-foreground': 'oklch(0.34 0 0)',
            accent: 'oklch(0.83 0.09 248.95)',
            'accent-foreground': 'oklch(0 0 0)',
            destructive: 'oklch(0.70 0.19 23.19)',
            border: 'oklch(0.85 0 0)',
            input: 'oklch(0.85 0 0)',
            ring: 'oklch(0.83 0.09 248.95)',
            'chart-1': 'oklch(0.80 0.14 349.25)',
            'chart-2': 'oklch(0.77 0.15 306.21)',
            'chart-3': 'oklch(0.83 0.09 248.95)',
            'chart-4': 'oklch(0.88 0.09 66.27)',
            'chart-5': 'oklch(0.94 0.14 130.35)',
            sidebar: 'oklch(1.00 0 0)',
            'sidebar-foreground': 'oklch(0.22 0 0)',
            'sidebar-primary': 'oklch(0.80 0.14 349.25)',
            'sidebar-primary-foreground': 'oklch(0 0 0)',
            'sidebar-accent': 'oklch(0.83 0.09 248.95)',
            'sidebar-accent-foreground': 'oklch(0 0 0)',
            'sidebar-border': 'oklch(0.85 0 0)',
            'sidebar-ring': 'oklch(0.83 0.09 248.95)',
            'font-sans': 'Gabriela, Geist Fallback, ui-sans-serif',
            'font-serif': 'Gabriela, Geist Fallback, ui-serif',
            'font-mono': 'Geist Mono, Geist Mono Fallback, ui-monospace',
            radius: '0rem',
            'shadow-color': 'oklch(0.83 0.09 248.95 )',
            'shadow-opacity': '0.10',
            'shadow-blur': '5px',
            'shadow-spread': '0px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '2px'
        },
        dark: {
            background: 'oklch(0.22 0 0)',
            foreground: 'oklch(0.97 0.01 264.53)',
            card: 'oklch(0.29 0 0)',
            'card-foreground': 'oklch(0.97 0.01 264.53)',
            popover: 'oklch(0.29 0 0)',
            'popover-foreground': 'oklch(0.97 0.01 264.53)',
            primary: 'oklch(0.80 0.14 349.25)',
            'primary-foreground': 'oklch(0.22 0 0)',
            secondary: 'oklch(0.77 0.15 306.21)',
            'secondary-foreground': 'oklch(0.22 0 0)',
            muted: 'oklch(0.32 0 0)',
            'muted-foreground': 'oklch(0.85 0 0)',
            accent: 'oklch(0.83 0.09 248.95)',
            'accent-foreground': 'oklch(0.22 0 0)',
            destructive: 'oklch(0.70 0.19 23.19)',
            border: 'oklch(0.39 0 0)',
            input: 'oklch(0.39 0 0)',
            ring: 'oklch(0.83 0.09 248.95)',
            'chart-1': 'oklch(0.80 0.14 349.25)',
            'chart-2': 'oklch(0.77 0.15 306.21)',
            'chart-3': 'oklch(0.83 0.09 248.95)',
            'chart-4': 'oklch(0.88 0.09 66.27)',
            'chart-5': 'oklch(0.94 0.14 130.35)',
            sidebar: 'oklch(0.29 0 0)',
            'sidebar-foreground': 'oklch(0.97 0.01 264.53)',
            'sidebar-primary': 'oklch(0.80 0.14 349.25)',
            'sidebar-primary-foreground': 'oklch(0.22 0 0)',
            'sidebar-accent': 'oklch(0.83 0.09 248.95)',
            'sidebar-accent-foreground': 'oklch(0.22 0 0)',
            'sidebar-border': 'oklch(0.39 0 0)',
            'sidebar-ring': 'oklch(0.83 0.09 248.95)',
            'shadow-color': 'oklch(0.83 0.09 248.95 / 0.10)',
            'shadow-opacity': '0.10',
            'shadow-blur': '2px',
            'shadow-spread': '0px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '1px'
        }
    },
    'art-deco': {
        light: {
            background: 'oklch(0.96 0.03 106.96)',
            foreground: 'oklch(0.40 0.07 91.45)',
            card: 'oklch(0.98 0.04 95.41)',
            'card-foreground': 'oklch(0.32 0 0)',
            popover: 'oklch(0.98 0.04 95.41)',
            'popover-foreground': 'oklch(0.32 0 0)',
            primary: 'oklch(0.77 0.14 91.05)',
            'primary-foreground': 'oklch(0 0 0)',
            secondary: 'oklch(0.67 0.13 61.29)',
            'secondary-foreground': 'oklch(0 0 0)',
            muted: 'oklch(0.93 0.03 106.91)',
            'muted-foreground': 'oklch(0.32 0 0)',
            accent: 'oklch(0.89 0.18 95.32)',
            'accent-foreground': 'oklch(0.32 0 0)',
            destructive: 'oklch(0.70 0.20 32.32)',
            border: 'oklch(0.83 0.11 92.68)',
            input: 'oklch(0.65 0.13 81.56)',
            ring: 'oklch(0.75 0.15 83.98)',
            'chart-1': 'oklch(0.89 0.18 95.32)',
            'chart-2': 'oklch(0.67 0.13 61.29)',
            'chart-3': 'oklch(0.65 0.13 81.56)',
            'chart-4': 'oklch(0.75 0.15 83.98)',
            'chart-5': 'oklch(0.77 0.14 91.05)',
            sidebar: 'oklch(0.96 0.03 106.96)',
            'sidebar-foreground': 'oklch(0.32 0 0)',
            'sidebar-primary': 'oklch(0.77 0.14 91.05)',
            'sidebar-primary-foreground': 'oklch(0.32 0 0)',
            'sidebar-accent': 'oklch(0.89 0.18 95.32)',
            'sidebar-accent-foreground': 'oklch(0.32 0 0)',
            'sidebar-border': 'oklch(0.65 0.13 81.56)',
            'sidebar-ring': 'oklch(0.75 0.15 83.98)',
            'font-sans': 'Delius Swash Caps',
            'font-serif': 'Delius Swash Caps',
            'font-mono': 'Delius Swash Caps',
            radius: '0.625rem',
            'shadow-color': 'oklch(0.70 0.17 28.12 / 30%)',
            'shadow-opacity': '0.05',
            'shadow-blur': '3px',
            'shadow-spread': '0px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '1px'
        },
        dark: {
            background: 'oklch(0.32 0 0)',
            foreground: 'oklch(0.96 0.03 106.96)',
            card: 'oklch(0.41 0 0)',
            'card-foreground': 'oklch(0.96 0.03 106.96)',
            popover: 'oklch(0.41 0 0)',
            'popover-foreground': 'oklch(0.96 0.03 106.96)',
            primary: 'oklch(0.84 0.17 82.56)',
            'primary-foreground': 'oklch(0 0 0)',
            secondary: 'oklch(0.47 0.11 50.84)',
            'secondary-foreground': 'oklch(0.96 0.03 106.96)',
            muted: 'oklch(0.44 0 0)',
            'muted-foreground': 'oklch(0.96 0.03 106.96)',
            accent: 'oklch(0.66 0.14 80.23)',
            'accent-foreground': 'oklch(0 0 0)',
            destructive: 'oklch(0.66 0.23 35.40)',
            border: 'oklch(0.47 0.11 50.84)',
            input: 'oklch(0.47 0.11 50.84)',
            ring: 'oklch(0.65 0.13 81.56)',
            'chart-1': 'oklch(0.75 0.15 83.98)',
            'chart-2': 'oklch(0.47 0.11 50.84)',
            'chart-3': 'oklch(0.65 0.13 81.56)',
            'chart-4': 'oklch(0.75 0.15 83.98)',
            'chart-5': 'oklch(0.65 0.13 81.56)',
            sidebar: 'oklch(0.32 0 0)',
            'sidebar-foreground': 'oklch(1.00 0 0)',
            'sidebar-primary': 'oklch(0.61 0.13 80.96)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.75 0.15 83.98)',
            'sidebar-accent-foreground': 'oklch(0.96 0.03 106.96)',
            'sidebar-border': 'oklch(0.47 0.11 50.84)',
            'sidebar-ring': 'oklch(0.65 0.13 81.56)',
            'shadow-color': 'oklch(0.00 0 0 / 0.05)',
            'shadow-opacity': '0.05',
            'shadow-blur': '3px',
            'shadow-spread': '0px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '1px'
        }
    },
    'vs-code': {
        light: {
            background: 'oklch(0.97 0.02 225.66)',
            foreground: 'oklch(0.15 0.02 269.18)',
            card: 'oklch(0.98 0.01 228.79)',
            'card-foreground': 'oklch(0.15 0.02 269.18)',
            popover: 'oklch(0.98 0.01 238.45)',
            'popover-foreground': 'oklch(0.15 0.02 269.18)',
            primary: 'oklch(0.71 0.15 239.07)',
            'primary-foreground': 'oklch(0.94 0.03 232.39)',
            secondary: 'oklch(0.91 0.03 229.20)',
            'secondary-foreground': 'oklch(0.15 0.02 269.18)',
            muted: 'oklch(0.89 0.02 225.69)',
            'muted-foreground': 'oklch(0.36 0.03 230.30)',
            accent: 'oklch(0.88 0.02 235.72)',
            'accent-foreground': 'oklch(0.34 0.05 229.72)',
            destructive: 'oklch(0.61 0.24 20.96)',
            border: 'oklch(0.82 0.02 240.77)',
            input: 'oklch(0.82 0.02 240.77)',
            ring: 'oklch(0.55 0.10 235.72)',
            'chart-1': 'oklch(0.57 0.11 228.97)',
            'chart-2': 'oklch(0.45 0.10 270.08)',
            'chart-3': 'oklch(0.65 0.15 159.03)',
            'chart-4': 'oklch(0.75 0.10 100.01)',
            'chart-5': 'oklch(0.55 0.15 299.88)',
            sidebar: 'oklch(0.93 0.01 238.46)',
            'sidebar-foreground': 'oklch(0.15 0.02 269.18)',
            'sidebar-primary': 'oklch(0.57 0.11 228.97)',
            'sidebar-primary-foreground': 'oklch(0.99 0.01 203.97)',
            'sidebar-accent': 'oklch(0.88 0.02 235.72)',
            'sidebar-accent-foreground': 'oklch(0.15 0.02 269.18)',
            'sidebar-border': 'oklch(0.82 0.02 240.77)',
            'sidebar-ring': 'oklch(0.57 0.11 228.97)',
            'font-sans': "'Source Code Pro', 'Geist', 'Geist Fallback', ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, 'Noto Sans', sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol', 'Noto Color Emoji'",
            'font-serif': "'Source Serif 4', 'Geist', 'Geist Fallback', ui-serif, Georgia, Cambria, 'Times New Roman', Times, serif",
            'font-mono': "'Source Code Pro', 'Geist Mono', 'Geist Mono Fallback', ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace",
            radius: '0rem',
            'shadow-color': 'oklch(0.49 0.09 235.45)',
            'shadow-opacity': '0.06',
            'shadow-blur': '2.5px',
            'shadow-spread': '0px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '1px'
        },
        dark: {
            background: 'oklch(0.18 0.02 271.27)',
            foreground: 'oklch(0.90 0.01 238.47)',
            card: 'oklch(0.22 0.02 271.67)',
            'card-foreground': 'oklch(0.90 0.01 238.47)',
            popover: 'oklch(0.22 0.02 271.67)',
            'popover-foreground': 'oklch(0.90 0.01 238.47)',
            primary: 'oklch(0.71 0.15 239.07)',
            'primary-foreground': 'oklch(0.94 0.03 232.39)',
            secondary: 'oklch(0.28 0.03 270.91)',
            'secondary-foreground': 'oklch(0.90 0.01 238.47)',
            muted: 'oklch(0.28 0.03 270.91)',
            'muted-foreground': 'oklch(0.60 0.03 269.46)',
            accent: 'oklch(0.28 0.03 270.91)',
            'accent-foreground': 'oklch(0.90 0.01 238.47)',
            destructive: 'oklch(0.64 0.25 19.69)',
            border: 'oklch(0.90 0.01 238.47 / 15%)',
            input: 'oklch(0.90 0.01 238.47 / 20%)',
            ring: 'oklch(0.66 0.13 227.15)',
            'chart-1': 'oklch(0.66 0.13 227.15)',
            'chart-2': 'oklch(0.60 0.10 269.83)',
            'chart-3': 'oklch(0.70 0.15 159.83)',
            'chart-4': 'oklch(0.80 0.10 100.65)',
            'chart-5': 'oklch(0.60 0.15 300.14)',
            sidebar: 'oklch(0.22 0.02 271.67)',
            'sidebar-foreground': 'oklch(0.90 0.01 238.47)',
            'sidebar-primary': 'oklch(0.66 0.13 227.15)',
            'sidebar-primary-foreground': 'oklch(0.18 0.02 271.27)',
            'sidebar-accent': 'oklch(0.28 0.03 270.91)',
            'sidebar-accent-foreground': 'oklch(0.90 0.01 238.47)',
            'sidebar-border': 'oklch(0.90 0.01 238.47 / 15%)',
            'sidebar-ring': 'oklch(0.66 0.13 227.15)',
            'shadow-color': 'oklch(0 0 0)',
            'shadow-opacity': '0.01',
            'shadow-blur': '2px',
            'shadow-spread': '0px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '1px'
        }
    },
    spotify: {
        light: {
            background: 'oklch(0.99 0 0)',
            foreground: 'oklch(0.35 0.02 165.48)',
            card: 'oklch(1.00 0 0)',
            'card-foreground': 'oklch(0.35 0.02 165.48)',
            popover: 'oklch(1.00 0 0)',
            'popover-foreground': 'oklch(0.35 0.02 165.48)',
            primary: 'oklch(0.67 0.17 153.85)',
            'primary-foreground': 'oklch(0.99 0.02 169.99)',
            secondary: 'oklch(0.90 0.02 238.66)',
            'secondary-foreground': 'oklch(0.20 0.02 266.02)',
            muted: 'oklch(0.90 0.02 240.73)',
            'muted-foreground': 'oklch(0.50 0.03 268.53)',
            accent: 'oklch(0.90 0.02 240.73)',
            'accent-foreground': 'oklch(0.35 0.02 165.48)',
            destructive: 'oklch(0.61 0.24 20.96)',
            border: 'oklch(0.94 0.01 238.46)',
            input: 'oklch(0.85 0.02 240.75)',
            ring: 'oklch(0.67 0.17 153.85)',
            'chart-1': 'oklch(0.67 0.17 153.85)',
            'chart-2': 'oklch(0.50 0.10 270.06)',
            'chart-3': 'oklch(0.72 0.12 201.79)',
            'chart-4': 'oklch(0.80 0.10 100.65)',
            'chart-5': 'oklch(0.60 0.15 300.14)',
            sidebar: 'oklch(0.98 0.01 238.45)',
            'sidebar-foreground': 'oklch(0.35 0.02 165.48)',
            'sidebar-primary': 'oklch(0.67 0.17 153.85)',
            'sidebar-primary-foreground': 'oklch(0.98 0.01 238.45)',
            'sidebar-accent': 'oklch(0.90 0.02 240.73)',
            'sidebar-accent-foreground': 'oklch(0.35 0.02 165.48)',
            'sidebar-border': 'oklch(0.85 0.02 240.75)',
            'sidebar-ring': 'oklch(0.67 0.17 153.85)',
            'font-sans': 'Lato, sans-serif',
            'font-serif': 'Merriweather, Geist, Geist Fallback, ui-serif, Georgia, Cambria, "Times New Roman", Times, serif',
            'font-mono': 'Roboto Mono, Geist Mono, Geist Mono Fallback, ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
            radius: '0.25rem',
            'shadow-color': 'oklch(0.35 0.05 163.50)',
            'shadow-opacity': '0.04',
            'shadow-blur': '3px',
            'shadow-spread': '0px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '1px'
        },
        dark: {
            background: 'oklch(0.15 0.02 269.18)',
            foreground: 'oklch(0.95 0.01 238.46)',
            card: 'oklch(0.20 0.02 266.02)',
            'card-foreground': 'oklch(0.95 0.01 238.46)',
            popover: 'oklch(0.20 0.02 266.02)',
            'popover-foreground': 'oklch(0.95 0.01 238.46)',
            primary: 'oklch(0.67 0.17 153.85)',
            'primary-foreground': 'oklch(0.15 0.02 269.18)',
            secondary: 'oklch(0.30 0.03 271.05)',
            'secondary-foreground': 'oklch(0.95 0.01 238.46)',
            muted: 'oklch(0.30 0.03 271.05)',
            'muted-foreground': 'oklch(0.60 0.03 269.46)',
            accent: 'oklch(0.30 0.03 271.05)',
            'accent-foreground': 'oklch(0.95 0.01 238.46)',
            destructive: 'oklch(0.64 0.25 19.69)',
            border: 'oklch(0.95 0.01 238.46 / 15%)',
            input: 'oklch(0.95 0.01 238.46 / 20%)',
            ring: 'oklch(0.67 0.17 153.85)',
            'chart-1': 'oklch(0.67 0.17 153.85)',
            'chart-2': 'oklch(0.60 0.10 269.83)',
            'chart-3': 'oklch(0.72 0.12 201.79)',
            'chart-4': 'oklch(0.80 0.10 100.65)',
            'chart-5': 'oklch(0.60 0.15 300.14)',
            sidebar: 'oklch(0.20 0.02 266.02)',
            'sidebar-foreground': 'oklch(0.95 0.01 238.46)',
            'sidebar-primary': 'oklch(0.67 0.17 153.85)',
            'sidebar-primary-foreground': 'oklch(0.15 0.02 269.18)',
            'sidebar-accent': 'oklch(0.30 0.03 271.05)',
            'sidebar-accent-foreground': 'oklch(0.95 0.01 238.46)',
            'sidebar-border': 'oklch(0.95 0.01 238.46 / 15%)',
            'sidebar-ring': 'oklch(0.67 0.17 153.85)',
            'shadow-color': 'oklch(0 0 0)',
            'shadow-opacity': '0.01',
            'shadow-blur': '2px',
            'shadow-spread': '0px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '1px'
        }
    },
    summer: {
        light: {
            background: 'oklch(0.98 0.01 78.24)',
            foreground: 'oklch(0.38 0.02 64.34)',
            card: 'oklch(0.97 0.02 74.09)',
            'card-foreground': 'oklch(0.38 0.02 64.34)',
            popover: 'oklch(0.96 0.04 81.50)',
            'popover-foreground': 'oklch(0.38 0.02 64.34)',
            primary: 'oklch(0.70 0.17 28.12)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.81 0.15 72.19)',
            'secondary-foreground': 'oklch(0.38 0.02 64.34)',
            muted: 'oklch(0.94 0.03 62.01)',
            'muted-foreground': 'oklch(0.62 0.06 59.53)',
            accent: 'oklch(0.64 0.22 28.81)',
            'accent-foreground': 'oklch(1.00 0 0)',
            destructive: 'oklch(0.57 0.20 26.41)',
            border: 'oklch(0.87 0.08 65.91)',
            input: 'oklch(0.96 0.03 79.26)',
            ring: 'oklch(0.70 0.17 28.12)',
            'chart-1': 'oklch(0.70 0.17 28.12)',
            'chart-2': 'oklch(0.81 0.15 72.19)',
            'chart-3': 'oklch(0.71 0.18 37.77)',
            'chart-4': 'oklch(0.89 0.15 91.22)',
            'chart-5': 'oklch(0.59 0.19 35.90)',
            sidebar: 'oklch(0.97 0.02 74.09)',
            'sidebar-foreground': 'oklch(0.38 0.02 64.34)',
            'sidebar-primary': 'oklch(0.70 0.17 28.12)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.81 0.15 72.19)',
            'sidebar-accent-foreground': 'oklch(0.38 0.02 64.34)',
            'sidebar-border': 'oklch(0.87 0.08 65.91)',
            'sidebar-ring': 'oklch(0.70 0.17 28.12)',
            'font-sans': 'Nunito, Segoe UI, Tahoma, Geneva, Verdana, sans-serif',
            'font-serif': 'Lora, ui-serif, Georgia, Cambria, Times New Roman, Times, serif',
            'font-mono': 'Fira Code, ui-monospace, SFMono-Regular',
            radius: '0.6rem',
            'shadow-color': 'oklch(0.70 0.17 28.12 / 30%)',
            'shadow-opacity': '0.05',
            'shadow-blur': '3px',
            'shadow-spread': '0px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '1px'
        },
        dark: {
            background: 'oklch(0.26 0.02 60.79)',
            foreground: 'oklch(0.87 0.08 65.91)',
            card: 'oklch(0.31 0.03 57.05)',
            'card-foreground': 'oklch(0.87 0.08 65.91)',
            popover: 'oklch(0.36 0.03 54.43)',
            'popover-foreground': 'oklch(0.87 0.08 65.91)',
            primary: 'oklch(0.70 0.17 28.12)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.81 0.15 72.19)',
            'secondary-foreground': 'oklch(0.26 0.02 60.79)',
            muted: 'oklch(0.56 0.05 58.96)',
            'muted-foreground': 'oklch(0.79 0.06 71.12)',
            accent: 'oklch(0.61 0.21 27.03)',
            'accent-foreground': 'oklch(1.00 0 0)',
            destructive: 'oklch(0.50 0.19 27.48)',
            border: 'oklch(0.45 0.05 59.00)',
            input: 'oklch(0.40 0.04 60.66)',
            ring: 'oklch(0.70 0.17 28.12)',
            'chart-1': 'oklch(0.70 0.17 28.12)',
            'chart-2': 'oklch(0.81 0.15 72.19)',
            'chart-3': 'oklch(0.71 0.18 37.77)',
            'chart-4': 'oklch(0.89 0.15 91.22)',
            'chart-5': 'oklch(0.59 0.19 35.90)',
            sidebar: 'oklch(0.31 0.03 57.05)',
            'sidebar-foreground': 'oklch(0.87 0.08 65.91)',
            'sidebar-primary': 'oklch(0.70 0.17 28.12)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.81 0.15 72.19)',
            'sidebar-accent-foreground': 'oklch(0.26 0.02 60.79)',
            'sidebar-border': 'oklch(0.45 0.05 59.00)',
            'sidebar-ring': 'oklch(0.70 0.17 28.12)',
            'shadow-color': 'oklch(0.70 0.17 28.12 / 70%)',
            'shadow-opacity': '0.05',
            'shadow-blur': '3px',
            'shadow-spread': '0px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '1px'
        }
    },
    'material-design': {
        light: {
            background: 'oklch(0.98 0.01 334.35)',
            foreground: 'oklch(0.22 0 0)',
            card: 'oklch(0.96 0.01 335.69)',
            'card-foreground': 'oklch(0.14 0 0)',
            popover: 'oklch(0.95 0.01 316.67)',
            'popover-foreground': 'oklch(0.40 0.04 309.35)',
            primary: 'oklch(0.51 0.21 286.50)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.49 0.04 300.23)',
            'secondary-foreground': 'oklch(1.00 0 0)',
            muted: 'oklch(0.96 0.01 335.69)',
            'muted-foreground': 'oklch(0.14 0 0)',
            accent: 'oklch(0.92 0.04 303.47)',
            'accent-foreground': 'oklch(0.14 0 0)',
            destructive: 'oklch(0.57 0.23 29.21)',
            border: 'oklch(0.83 0.02 308.26)',
            input: 'oklch(0.57 0.02 309.68)',
            ring: 'oklch(0.50 0.13 293.77)',
            'chart-1': 'oklch(0.61 0.21 279.42)',
            'chart-2': 'oklch(0.72 0.15 157.67)',
            'chart-3': 'oklch(0.66 0.17 324.24)',
            'chart-4': 'oklch(0.81 0.15 127.91)',
            'chart-5': 'oklch(0.68 0.17 258.25)',
            sidebar: 'oklch(0.99 0 0)',
            'sidebar-foreground': 'oklch(0.15 0 0)',
            'sidebar-primary': 'oklch(0.56 0.11 228.27)',
            'sidebar-primary-foreground': 'oklch(0.98 0 0)',
            'sidebar-accent': 'oklch(0.95 0 0)',
            'sidebar-accent-foreground': 'oklch(0.25 0 0)',
            'sidebar-border': 'oklch(0.90 0 0)',
            'sidebar-ring': 'oklch(0.56 0.11 228.27)',
            'font-sans': 'Roboto, sans-serif',
            'font-serif': 'Merriweather, serif',
            'font-mono': '"Geist Mono", "Geist Mono Fallback", ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
            radius: '1rem',
            'shadow-color': 'oklch(0 0 0 / 0.01)',
            'shadow-opacity': '0.01',
            'shadow-blur': '3px',
            'shadow-spread': '0px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '1px'
        },
        dark: {
            background: 'oklch(0.15 0.01 317.69)',
            foreground: 'oklch(0.95 0.01 321.50)',
            card: 'oklch(0.22 0.02 322.13)',
            'card-foreground': 'oklch(0.95 0.01 321.50)',
            popover: 'oklch(0.22 0.02 322.13)',
            'popover-foreground': 'oklch(0.95 0.01 321.50)',
            primary: 'oklch(0.60 0.22 279.81)',
            'primary-foreground': 'oklch(0.98 0.01 321.51)',
            secondary: 'oklch(0.45 0.03 294.79)',
            'secondary-foreground': 'oklch(0.95 0.01 321.50)',
            muted: 'oklch(0.22 0.01 319.50)',
            'muted-foreground': 'oklch(0.70 0.01 320.70)',
            accent: 'oklch(0.35 0.06 299.57)',
            'accent-foreground': 'oklch(0.95 0.01 321.50)',
            destructive: 'oklch(0.57 0.23 29.21)',
            border: 'oklch(0.40 0.04 309.35)',
            input: 'oklch(0.40 0.04 309.35)',
            ring: 'oklch(0.50 0.15 294.97)',
            'chart-1': 'oklch(0.50 0.25 274.99)',
            'chart-2': 'oklch(0.60 0.15 150.16)',
            'chart-3': 'oklch(0.65 0.20 309.96)',
            'chart-4': 'oklch(0.60 0.17 132.98)',
            'chart-5': 'oklch(0.60 0.20 255.25)',
            sidebar: 'oklch(0.20 0.01 317.74)',
            'sidebar-foreground': 'oklch(0.95 0.01 321.50)',
            'sidebar-primary': 'oklch(0.59 0.11 225.82)',
            'sidebar-primary-foreground': 'oklch(0.95 0.01 321.50)',
            'sidebar-accent': 'oklch(0.30 0.01 319.52)',
            'sidebar-accent-foreground': 'oklch(0.95 0.01 321.50)',
            'sidebar-border': 'oklch(0.35 0.01 319.53 / 30%)',
            'sidebar-ring': 'oklch(0.59 0.11 225.82)',
            'shadow-color': 'oklch(0 0 0 / 0.01)',
            'shadow-opacity': '0.01',
            'shadow-blur': '3px',
            'shadow-spread': '0px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '1px'
        }
    },
    marvel: {
        light: {
            background: 'oklch(0.98 0.01 25.23)',
            foreground: 'oklch(0.20 0.01 18.05)',
            card: 'oklch(0.95 0.01 25.23)',
            'card-foreground': 'oklch(0.18 0.01 29.18)',
            popover: 'oklch(0.94 0.01 25.23)',
            'popover-foreground': 'oklch(0.22 0.01 29.09)',
            primary: 'oklch(0.55 0.22 27.03)',
            'primary-foreground': 'oklch(0.98 0.01 100.72)',
            secondary: 'oklch(0.52 0.14 247.51)',
            'secondary-foreground': 'oklch(0.98 0.01 100.72)',
            muted: 'oklch(0.91 0.01 25.23)',
            'muted-foreground': 'oklch(0.38 0.01 17.71)',
            accent: 'oklch(0.86 0.04 33.03)',
            'accent-foreground': 'oklch(0.18 0.01 29.18)',
            destructive: 'oklch(0.56 0.23 29.23)',
            border: 'oklch(0.84 0.01 25.22)',
            input: 'oklch(0.80 0.01 25.22)',
            ring: 'oklch(0.50 0.12 244.86)',
            'chart-1': 'oklch(0.58 0.23 27.06)',
            'chart-2': 'oklch(0.61 0.18 251.95)',
            'chart-3': 'oklch(0.72 0.15 83.96)',
            'chart-4': 'oklch(0.67 0.15 144.89)',
            'chart-5': 'oklch(0.75 0.15 304.74)',
            sidebar: 'oklch(0.97 0 0)',
            'sidebar-foreground': 'oklch(0.20 0.01 18.05)',
            'sidebar-primary': 'oklch(0.52 0.14 247.51)',
            'sidebar-primary-foreground': 'oklch(0.98 0.01 100.72)',
            'sidebar-accent': 'oklch(0.69 0.14 79.64)',
            'sidebar-accent-foreground': 'oklch(0.20 0.01 18.05)',
            'sidebar-border': 'oklch(0.87 0.01 25.23)',
            'sidebar-ring': 'oklch(0.52 0.14 247.51)',
            'font-sans': 'Outfit, sans-serif',
            'font-serif': 'Merriweather, serif',
            'font-mono': 'Geist Mono, monospace',
            radius: '0rem',
            'shadow-color': 'oklch(0 0 0 / 0.01)',
            'shadow-opacity': '0.01',
            'shadow-blur': '3px',
            'shadow-spread': '0px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '1px'
        },
        dark: {
            background: 'oklch(0.12 0.01 38.49)',
            foreground: 'oklch(0.95 0.01 25.23)',
            card: 'oklch(0.18 0.01 29.18)',
            'card-foreground': 'oklch(0.95 0.01 25.23)',
            popover: 'oklch(0.18 0.01 29.18)',
            'popover-foreground': 'oklch(0.95 0.01 25.23)',
            primary: 'oklch(0.65 0.23 27.09)',
            'primary-foreground': 'oklch(0.98 0.01 100.72)',
            secondary: 'oklch(0.50 0.14 249.16)',
            'secondary-foreground': 'oklch(0.98 0.01 100.72)',
            muted: 'oklch(0.20 0.01 18.05)',
            'muted-foreground': 'oklch(0.70 0.01 25.22)',
            accent: 'oklch(0.59 0.12 78.11)',
            'accent-foreground': 'oklch(0.95 0.01 25.23)',
            destructive: 'oklch(0.56 0.23 29.23)',
            border: 'oklch(0.38 0.01 17.71)',
            input: 'oklch(0.38 0.01 17.71)',
            ring: 'oklch(0.49 0.14 250.75)',
            'chart-1': 'oklch(0.64 0.25 26.85)',
            'chart-2': 'oklch(0.66 0.19 250.17)',
            'chart-3': 'oklch(0.78 0.16 87.01)',
            'chart-4': 'oklch(0.68 0.15 144.94)',
            'chart-5': 'oklch(0.75 0.15 304.74)',
            sidebar: 'oklch(0.14 0.01 33.25)',
            'sidebar-foreground': 'oklch(0.95 0.01 25.23)',
            'sidebar-primary': 'oklch(0.50 0.14 249.16)',
            'sidebar-primary-foreground': 'oklch(0.98 0.01 100.72)',
            'sidebar-accent': 'oklch(0.59 0.12 78.11)',
            'sidebar-accent-foreground': 'oklch(0.95 0.01 25.23)',
            'sidebar-border': 'oklch(0.32 0.01 27.45 / 30%)',
            'sidebar-ring': 'oklch(0.50 0.14 249.16)',
            'shadow-color': 'oklch(0 0 0 / 0.01)',
            'shadow-opacity': '0.01',
            'shadow-blur': '3px',
            'shadow-spread': '0px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '1px'
        }
    },
    valorant: {
        light: {
            background: 'oklch(0.97 0.02 12.78)',
            foreground: 'oklch(0.24 0.07 17.81)',
            card: 'oklch(0.98 0.01 17.28)',
            'card-foreground': 'oklch(0.26 0.07 19)',
            popover: 'oklch(0.98 0.01 17.28)',
            'popover-foreground': 'oklch(0.26 0.07 19)',
            primary: 'oklch(0.67 0.22 21.34)',
            'primary-foreground': 'oklch(0.99 0.00 359.99)',
            secondary: 'oklch(0.95 0.02 11.28)',
            'secondary-foreground': 'oklch(0.24 0.07 17.81)',
            muted: 'oklch(0.98 0.01 17.28)',
            'muted-foreground': 'oklch(0.26 0.07 19)',
            accent: 'oklch(0.99 0.00 359.99)',
            'accent-foreground': 'oklch(0.43 0.13 20.62)',
            destructive: 'oklch(0.80 0.17 73.27)',
            border: 'oklch(0.91 0.05 11.40)',
            input: 'oklch(0.90 0.05 12.59)',
            ring: 'oklch(0.92 0.04 12.39)',
            'chart-1': 'oklch(0.86 0.18 88.49)',
            'chart-2': 'oklch(0.62 0.21 255.13)',
            'chart-3': 'oklch(0.54 0.29 297.82)',
            'chart-4': 'oklch(0.95 0.10 98.39)',
            'chart-5': 'oklch(0.87 0.12 100.28)',
            sidebar: 'oklch(0.97 0.02 12.78)',
            'sidebar-foreground': 'oklch(0.26 0.07 19)',
            'sidebar-primary': 'oklch(0.67 0.22 21.34)',
            'sidebar-primary-foreground': 'oklch(0.98 0.01 17.28)',
            'sidebar-accent': 'oklch(0.98 0.01 17.28)',
            'sidebar-accent-foreground': 'oklch(0.43 0.13 20.62)',
            'sidebar-border': 'oklch(0.91 0.05 11.40)',
            'sidebar-ring': 'oklch(0.92 0.04 12.39)',
            'font-sans': 'Barlow',
            'font-serif': 'Merriweather',
            'font-mono': 'JetBrains Mono',
            radius: '0rem',
            'shadow-color': 'oklch(0.3 0.0891 19.6)',
            'shadow-opacity': '0.08',
            'shadow-blur': '3px',
            'shadow-spread': '0px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '0px'
        },
        dark: {
            background: 'oklch(0.16 0.03 17.48)',
            foreground: 'oklch(0.99 0.00 359.99)',
            card: 'oklch(0.21 0.05 19.26)',
            'card-foreground': 'oklch(0.98 0 0)',
            popover: 'oklch(0.26 0.07 19)',
            'popover-foreground': 'oklch(0.99 0.00 359.99)',
            primary: 'oklch(0.67 0.22 21.34)',
            'primary-foreground': 'oklch(0.99 0.00 359.99)',
            secondary: 'oklch(0.3 0.0891 19.6)',
            'secondary-foreground': 'oklch(0.95 0.02 11.28)',
            muted: 'oklch(0.26 0.07 19)',
            'muted-foreground': 'oklch(0.99 0.00 359.99)',
            accent: 'oklch(0.43 0.13 20.62)',
            'accent-foreground': 'oklch(0.99 0.00 359.99)',
            destructive: 'oklch(0.80 0.17 73.27)',
            border: 'oklch(0.31 0.09 19.80)',
            input: 'oklch(0.39 0.12 20.37)',
            ring: 'oklch(0.50 0.16 20.89)',
            'chart-1': 'oklch(0.86 0.18 88.49)',
            'chart-2': 'oklch(0.62 0.21 255.13)',
            'chart-3': 'oklch(0.54 0.29 297.82)',
            'chart-4': 'oklch(0.95 0.10 98.39)',
            'chart-5': 'oklch(0.87 0.12 100.28)',
            sidebar: 'oklch(0.26 0.07 19)',
            'sidebar-foreground': 'oklch(0.99 0.00 359.99)',
            'sidebar-primary': 'oklch(0.67 0.22 21.34)',
            'sidebar-primary-foreground': 'oklch(0.99 0.00 359.99)',
            'sidebar-accent': 'oklch(0.43 0.13 20.62)',
            'sidebar-accent-foreground': 'oklch(0.99 0.00 359.99)',
            'sidebar-border': 'oklch(0.39 0.12 20.37)',
            'sidebar-ring': 'oklch(0.50 0.16 20.89)'
        }
    },
    'ghibli-studio': {
        light: {
            background: 'oklch(0.91 0.05 82.69)',
            foreground: 'oklch(0.41 0.08 79.04)',
            card: 'oklch(0.92 0.04 83.86)',
            'card-foreground': 'oklch(0.41 0.08 73.75)',
            popover: 'oklch(0.92 0.04 83.86)',
            'popover-foreground': 'oklch(0.41 0.08 73.75)',
            primary: 'oklch(0.71 0.10 111.99)',
            'primary-foreground': 'oklch(0.98 0.01 3.71)',
            secondary: 'oklch(0.88 0.05 83.41)',
            'secondary-foreground': 'oklch(0.51 0.08 79.21)',
            muted: 'oklch(0.86 0.06 83.48)',
            'muted-foreground': 'oklch(0.51 0.08 74.26)',
            accent: 'oklch(0.86 0.05 84.50)',
            'accent-foreground': 'oklch(0.26 0.02 358.42)',
            destructive: 'oklch(0.63 0.24 29.21)',
            border: 'oklch(0.74 0.06 79.81)',
            input: 'oklch(0.74 0.06 79.81)',
            ring: 'oklch(0.51 0.08 74.26)',
            'chart-1': 'oklch(0.66 0.19 41.68)',
            'chart-2': 'oklch(0.70 0.12 183.20)',
            'chart-3': 'oklch(0.48 0.08 211.46)',
            'chart-4': 'oklch(0.84 0.17 85.07)',
            'chart-5': 'oklch(0.74 0.17 60.21)',
            sidebar: 'oklch(0.87 0.06 83.96)',
            'sidebar-foreground': 'oklch(0.41 0.08 79.04)',
            'sidebar-primary': 'oklch(0.26 0.02 358.42)',
            'sidebar-primary-foreground': 'oklch(0.98 0.01 3.71)',
            'sidebar-accent': 'oklch(0.83 0.06 84.46)',
            'sidebar-accent-foreground': 'oklch(0.26 0.02 358.42)',
            'sidebar-border': 'oklch(0.91 0.00 0.43)',
            'sidebar-ring': 'oklch(0.71 0.00 0.37)',
            'font-sans': 'Nunito, sans-serif',
            'font-serif': 'PT Serif, serif',
            'font-mono': 'JetBrains Mono, monospace',
            radius: '0.625rem'
        },
        dark: {
            background: 'oklch(0.20 0.01 48.35)',
            foreground: 'oklch(0.88 0.05 79.26)',
            card: 'oklch(0.25 0.01 56.14)',
            'card-foreground': 'oklch(0.88 0.05 79.26)',
            popover: 'oklch(0.25 0.01 56.14)',
            'popover-foreground': 'oklch(0.88 0.05 79.26)',
            primary: 'oklch(0.64 0.05 115.39)',
            'primary-foreground': 'oklch(0.98 0.01 3.71)',
            secondary: 'oklch(0.33 0.02 60.70)',
            'secondary-foreground': 'oklch(0.88 0.05 83.41)',
            muted: 'oklch(0.27 0.01 39.35)',
            'muted-foreground': 'oklch(0.74 0.06 79.81)',
            accent: 'oklch(0.33 0.02 60.70)',
            'accent-foreground': 'oklch(0.86 0.05 84.50)',
            destructive: 'oklch(0.63 0.24 29.21)',
            border: 'oklch(0.33 0.02 60.70)',
            input: 'oklch(0.33 0.02 60.70)',
            ring: 'oklch(0.64 0.05 115.39)',
            'chart-1': 'oklch(0.66 0.19 41.68)',
            'chart-2': 'oklch(0.70 0.12 183.20)',
            'chart-3': 'oklch(0.48 0.08 211.46)',
            'chart-4': 'oklch(0.84 0.17 85.07)',
            'chart-5': 'oklch(0.74 0.17 60.21)',
            sidebar: 'oklch(0.23 0.01 56.09)',
            'sidebar-foreground': 'oklch(0.88 0.05 79.26)',
            'sidebar-primary': 'oklch(0.64 0.05 115.39)',
            'sidebar-primary-foreground': 'oklch(0.98 0.01 3.71)',
            'sidebar-accent': 'oklch(0.33 0.02 60.70)',
            'sidebar-accent-foreground': 'oklch(0.86 0.05 84.50)',
            'sidebar-border': 'oklch(0.33 0.02 60.70)',
            'sidebar-ring': 'oklch(0.64 0.05 115.39)'
        }
    },
    'modern-minimal': {
        light: {
            background: 'oklch(1.00 0 0)',
            foreground: 'oklch(0.32 0 0)',
            card: 'oklch(1.00 0 0)',
            'card-foreground': 'oklch(0.32 0 0)',
            popover: 'oklch(1.00 0 0)',
            'popover-foreground': 'oklch(0.32 0 0)',
            primary: 'oklch(0.62 0.19 259.76)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.97 0 0)',
            'secondary-foreground': 'oklch(0.45 0.03 257.68)',
            muted: 'oklch(0.98 0 0)',
            'muted-foreground': 'oklch(0.55 0.02 264.41)',
            accent: 'oklch(0.95 0.03 233.56)',
            'accent-foreground': 'oklch(0.38 0.14 265.59)',
            destructive: 'oklch(0.64 0.21 25.39)',
            border: 'oklch(0.93 0.01 261.82)',
            input: 'oklch(0.93 0.01 261.82)',
            ring: 'oklch(0.62 0.19 259.76)',
            'chart-1': 'oklch(0.62 0.19 259.76)',
            'chart-2': 'oklch(0.55 0.22 262.96)',
            'chart-3': 'oklch(0.49 0.22 264.43)',
            'chart-4': 'oklch(0.42 0.18 265.55)',
            'chart-5': 'oklch(0.38 0.14 265.59)',
            sidebar: 'oklch(0.98 0 0)',
            'sidebar-foreground': 'oklch(0.14 0 0)',
            'sidebar-primary': 'oklch(0.20 0 0)',
            'sidebar-primary-foreground': 'oklch(0.98 0 0)',
            'sidebar-accent': 'oklch(0.97 0 0)',
            'sidebar-accent-foreground': 'oklch(0.20 0 0)',
            'sidebar-border': 'oklch(0.92 0 0)',
            'sidebar-ring': 'oklch(0.71 0 0)',
            'font-serif': 'Source Serif 4, serif',
            'font-mono': 'JetBrains Mono, monospace',
            radius: '0.375rem'
        },
        dark: {
            background: 'oklch(0.20 0 0)',
            foreground: 'oklch(0.92 0 0)',
            card: 'oklch(0.27 0 0)',
            'card-foreground': 'oklch(0.92 0 0)',
            popover: 'oklch(0.27 0 0)',
            'popover-foreground': 'oklch(0.92 0 0)',
            primary: 'oklch(0.62 0.19 259.76)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.27 0 0)',
            'secondary-foreground': 'oklch(0.92 0 0)',
            muted: 'oklch(0.27 0 0)',
            'muted-foreground': 'oklch(0.72 0 0)',
            accent: 'oklch(0.38 0.14 265.59)',
            'accent-foreground': 'oklch(0.88 0.06 254.63)',
            destructive: 'oklch(0.64 0.21 25.39)',
            border: 'oklch(0.37 0 0)',
            input: 'oklch(0.37 0 0)',
            ring: 'oklch(0.62 0.19 259.76)',
            'chart-1': 'oklch(0.71 0.14 254.69)',
            'chart-2': 'oklch(0.62 0.19 259.76)',
            'chart-3': 'oklch(0.55 0.22 262.96)',
            'chart-4': 'oklch(0.49 0.22 264.43)',
            'chart-5': 'oklch(0.42 0.18 265.55)',
            sidebar: 'oklch(0.21 0.01 285.93)',
            'sidebar-foreground': 'oklch(0.99 0 0)',
            'sidebar-primary': 'oklch(0.49 0.24 264.40)',
            'sidebar-primary-foreground': 'oklch(0.99 0 0)',
            'sidebar-accent': 'oklch(0.27 0.01 286.10)',
            'sidebar-accent-foreground': 'oklch(0.99 0 0)',
            'sidebar-border': 'oklch(1.00 0 0 / 10%)',
            'sidebar-ring': 'oklch(0.55 0.02 285.93)'
        }
    },
    nature: {
        light: {
            background: 'oklch(0.97 0.01 80.72)',
            foreground: 'oklch(0.30 0.04 30.20)',
            card: 'oklch(0.97 0.01 80.72)',
            'card-foreground': 'oklch(0.30 0.04 30.20)',
            popover: 'oklch(0.97 0.01 80.72)',
            'popover-foreground': 'oklch(0.30 0.04 30.20)',
            primary: 'oklch(0.52 0.13 144.17)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.96 0.02 147.64)',
            'secondary-foreground': 'oklch(0.43 0.12 144.31)',
            muted: 'oklch(0.94 0.01 74.42)',
            'muted-foreground': 'oklch(0.45 0.05 39.21)',
            accent: 'oklch(0.90 0.05 146.04)',
            'accent-foreground': 'oklch(0.43 0.12 144.31)',
            destructive: 'oklch(0.54 0.19 26.72)',
            border: 'oklch(0.88 0.02 74.64)',
            input: 'oklch(0.88 0.02 74.64)',
            ring: 'oklch(0.52 0.13 144.17)',
            'chart-1': 'oklch(0.67 0.16 144.21)',
            'chart-2': 'oklch(0.58 0.14 144.18)',
            'chart-3': 'oklch(0.52 0.13 144.17)',
            'chart-4': 'oklch(0.43 0.12 144.31)',
            'chart-5': 'oklch(0.22 0.05 145.73)',
            sidebar: 'oklch(0.94 0.01 74.42)',
            'sidebar-foreground': 'oklch(0.30 0.04 30.20)',
            'sidebar-primary': 'oklch(0.52 0.13 144.17)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.90 0.05 146.04)',
            'sidebar-accent-foreground': 'oklch(0.43 0.12 144.31)',
            'sidebar-border': 'oklch(0.88 0.02 74.64)',
            'sidebar-ring': 'oklch(0.52 0.13 144.17)',
            'font-sans': 'Montserrat, sans-serif',
            'font-serif': 'Merriweather, serif',
            'font-mono': 'Source Code Pro, monospace',
            radius: '0.5rem'
        },
        dark: {
            background: 'oklch(0.27 0.03 150.77)',
            foreground: 'oklch(0.94 0.01 72.66)',
            card: 'oklch(0.33 0.03 146.99)',
            'card-foreground': 'oklch(0.94 0.01 72.66)',
            popover: 'oklch(0.33 0.03 146.99)',
            'popover-foreground': 'oklch(0.94 0.01 72.66)',
            primary: 'oklch(0.67 0.16 144.21)',
            'primary-foreground': 'oklch(0.22 0.05 145.73)',
            secondary: 'oklch(0.39 0.03 142.99)',
            'secondary-foreground': 'oklch(0.90 0.02 142.55)',
            muted: 'oklch(0.33 0.03 146.99)',
            'muted-foreground': 'oklch(0.86 0.02 76.10)',
            accent: 'oklch(0.58 0.14 144.18)',
            'accent-foreground': 'oklch(0.94 0.01 72.66)',
            destructive: 'oklch(0.54 0.19 26.72)',
            border: 'oklch(0.39 0.03 142.99)',
            input: 'oklch(0.39 0.03 142.99)',
            ring: 'oklch(0.67 0.16 144.21)',
            'chart-1': 'oklch(0.77 0.12 145.30)',
            'chart-2': 'oklch(0.72 0.14 144.89)',
            'chart-3': 'oklch(0.67 0.16 144.21)',
            'chart-4': 'oklch(0.63 0.15 144.20)',
            'chart-5': 'oklch(0.58 0.14 144.18)',
            sidebar: 'oklch(0.27 0.03 150.77)',
            'sidebar-foreground': 'oklch(0.94 0.01 72.66)',
            'sidebar-primary': 'oklch(0.67 0.16 144.21)',
            'sidebar-primary-foreground': 'oklch(0.22 0.05 145.73)',
            'sidebar-accent': 'oklch(0.58 0.14 144.18)',
            'sidebar-accent-foreground': 'oklch(0.94 0.01 72.66)',
            'sidebar-border': 'oklch(0.39 0.03 142.99)',
            'sidebar-ring': 'oklch(0.67 0.16 144.21)'
        }
    },
    'elegant-luxury': {
        light: {
            background: 'oklch(0.98 0.00 56.38)',
            foreground: 'oklch(0.22 0 0)',
            card: 'oklch(0.98 0.00 56.38)',
            'card-foreground': 'oklch(0.22 0 0)',
            popover: 'oklch(0.98 0.00 56.38)',
            'popover-foreground': 'oklch(0.22 0 0)',
            primary: 'oklch(0.47 0.15 24.94)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.96 0.04 89.09)',
            'secondary-foreground': 'oklch(0.48 0.10 75.12)',
            muted: 'oklch(0.94 0.01 53.44)',
            'muted-foreground': 'oklch(0.44 0.01 73.64)',
            accent: 'oklch(0.96 0.06 95.62)',
            'accent-foreground': 'oklch(0.40 0.13 25.72)',
            destructive: 'oklch(0.44 0.16 26.90)',
            border: 'oklch(0.94 0.03 80.99)',
            input: 'oklch(0.94 0.03 80.99)',
            ring: 'oklch(0.47 0.15 24.94)',
            'chart-1': 'oklch(0.51 0.19 27.52)',
            'chart-2': 'oklch(0.47 0.15 24.94)',
            'chart-3': 'oklch(0.40 0.13 25.72)',
            'chart-4': 'oklch(0.56 0.15 49.00)',
            'chart-5': 'oklch(0.47 0.12 46.20)',
            sidebar: 'oklch(0.94 0.01 53.44)',
            'sidebar-foreground': 'oklch(0.22 0 0)',
            'sidebar-primary': 'oklch(0.47 0.15 24.94)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.96 0.06 95.62)',
            'sidebar-accent-foreground': 'oklch(0.40 0.13 25.72)',
            'sidebar-border': 'oklch(0.94 0.03 80.99)',
            'sidebar-ring': 'oklch(0.47 0.15 24.94)',
            'font-sans': 'Poppins, sans-serif',
            'font-serif': 'Libre Baskerville, serif',
            'font-mono': 'IBM Plex Mono, monospace',
            radius: '0.375rem',
            'shadow-color': 'hsl(0 63% 18%)',
            'shadow-opacity': '0.12',
            'shadow-blur': '16px',
            'shadow-spread': '-2px',
            'shadow-offset-x': '1px',
            'shadow-offset-y': '1px'
        },
        dark: {
            background: 'oklch(0.22 0.01 56.04)',
            foreground: 'oklch(0.97 0.00 106.42)',
            card: 'oklch(0.27 0.01 34.30)',
            'card-foreground': 'oklch(0.97 0.00 106.42)',
            popover: 'oklch(0.27 0.01 34.30)',
            'popover-foreground': 'oklch(0.97 0.00 106.42)',
            primary: 'oklch(0.51 0.19 27.52)',
            'primary-foreground': 'oklch(0.98 0.00 56.38)',
            secondary: 'oklch(0.47 0.12 46.20)',
            'secondary-foreground': 'oklch(0.96 0.06 95.62)',
            muted: 'oklch(0.27 0.01 34.30)',
            'muted-foreground': 'oklch(0.87 0.00 56.37)',
            accent: 'oklch(0.56 0.15 49.00)',
            'accent-foreground': 'oklch(0.96 0.06 95.62)',
            destructive: 'oklch(0.64 0.21 25.33)',
            border: 'oklch(0.37 0.01 67.56)',
            input: 'oklch(0.37 0.01 67.56)',
            ring: 'oklch(0.51 0.19 27.52)',
            'chart-1': 'oklch(0.71 0.17 22.22)',
            'chart-2': 'oklch(0.64 0.21 25.33)',
            'chart-3': 'oklch(0.58 0.22 27.33)',
            'chart-4': 'oklch(0.84 0.16 84.43)',
            'chart-5': 'oklch(0.77 0.16 70.08)',
            sidebar: 'oklch(0.22 0.01 56.04)',
            'sidebar-foreground': 'oklch(0.97 0.00 106.42)',
            'sidebar-primary': 'oklch(0.51 0.19 27.52)',
            'sidebar-primary-foreground': 'oklch(0.98 0.00 56.38)',
            'sidebar-accent': 'oklch(0.56 0.15 49.00)',
            'sidebar-accent-foreground': 'oklch(0.96 0.06 95.62)',
            'sidebar-border': 'oklch(0.37 0.01 67.56)',
            'sidebar-ring': 'oklch(0.51 0.19 27.52)'
        }
    },
    'neo-brutalism': {
        light: {
            background: 'oklch(1.00 0 0)',
            foreground: 'oklch(0 0 0)',
            card: 'oklch(1.00 0 0)',
            'card-foreground': 'oklch(0 0 0)',
            popover: 'oklch(1.00 0 0)',
            'popover-foreground': 'oklch(0 0 0)',
            primary: 'oklch(0.65 0.24 26.97)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.97 0.21 109.77)',
            'secondary-foreground': 'oklch(0 0 0)',
            muted: 'oklch(0.96 0 0)',
            'muted-foreground': 'oklch(0.32 0 0)',
            accent: 'oklch(0.56 0.24 260.82)',
            'accent-foreground': 'oklch(1.00 0 0)',
            destructive: 'oklch(0 0 0)',
            border: 'oklch(0 0 0)',
            input: 'oklch(0 0 0)',
            ring: 'oklch(0.65 0.24 26.97)',
            'chart-1': 'oklch(0.65 0.24 26.97)',
            'chart-2': 'oklch(0.97 0.21 109.77)',
            'chart-3': 'oklch(0.56 0.24 260.82)',
            'chart-4': 'oklch(0.73 0.25 142.50)',
            'chart-5': 'oklch(0.59 0.27 328.36)',
            sidebar: 'oklch(0.96 0 0)',
            'sidebar-foreground': 'oklch(0 0 0)',
            'sidebar-primary': 'oklch(0.65 0.24 26.97)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.56 0.24 260.82)',
            'sidebar-accent-foreground': 'oklch(1.00 0 0)',
            'sidebar-border': 'oklch(0 0 0)',
            'sidebar-ring': 'oklch(0.65 0.24 26.97)',
            'font-sans': 'DM Sans, sans-serif',
            'font-serif': 'ui-serif, Georgia, Cambria, "Times New Roman", Times, serif',
            'font-mono': 'Space Mono, monospace',
            radius: '0px',
            'shadow-color': 'hsl(0 0% 0%)',
            'shadow-opacity': '1',
            'shadow-blur': '0px',
            'shadow-spread': '0px',
            'shadow-offset-x': '4px',
            'shadow-offset-y': '4px'
        },
        dark: {
            background: 'oklch(0 0 0)',
            foreground: 'oklch(1.00 0 0)',
            card: 'oklch(0.32 0 0)',
            'card-foreground': 'oklch(1.00 0 0)',
            popover: 'oklch(0.32 0 0)',
            'popover-foreground': 'oklch(1.00 0 0)',
            primary: 'oklch(0.70 0.19 23.19)',
            'primary-foreground': 'oklch(0 0 0)',
            secondary: 'oklch(0.97 0.20 109.62)',
            'secondary-foreground': 'oklch(0 0 0)',
            muted: 'oklch(0.32 0 0)',
            'muted-foreground': 'oklch(0.85 0 0)',
            accent: 'oklch(0.68 0.18 252.26)',
            'accent-foreground': 'oklch(0 0 0)',
            destructive: 'oklch(1.00 0 0)',
            border: 'oklch(1.00 0 0)',
            input: 'oklch(1.00 0 0)',
            ring: 'oklch(0.70 0.19 23.19)',
            'chart-1': 'oklch(0.70 0.19 23.19)',
            'chart-2': 'oklch(0.97 0.20 109.62)',
            'chart-3': 'oklch(0.68 0.18 252.26)',
            'chart-4': 'oklch(0.74 0.23 142.85)',
            'chart-5': 'oklch(0.61 0.25 328.07)',
            sidebar: 'oklch(0 0 0)',
            'sidebar-foreground': 'oklch(1.00 0 0)',
            'sidebar-primary': 'oklch(0.70 0.19 23.19)',
            'sidebar-primary-foreground': 'oklch(0 0 0)',
            'sidebar-accent': 'oklch(0.68 0.18 252.26)',
            'sidebar-accent-foreground': 'oklch(0 0 0)',
            'sidebar-border': 'oklch(1.00 0 0)',
            'sidebar-ring': 'oklch(0.70 0.19 23.19)'
        }
    },
    'pastel-dreams': {
        light: {
            background: 'oklch(0.97 0.01 314.78)',
            foreground: 'oklch(0.37 0.03 259.73)',
            card: 'oklch(1.00 0 0)',
            'card-foreground': 'oklch(0.37 0.03 259.73)',
            popover: 'oklch(1.00 0 0)',
            'popover-foreground': 'oklch(0.37 0.03 259.73)',
            primary: 'oklch(0.71 0.16 293.54)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.91 0.05 306.09)',
            'secondary-foreground': 'oklch(0.45 0.03 256.80)',
            muted: 'oklch(0.95 0.03 307.17)',
            'muted-foreground': 'oklch(0.55 0.02 264.36)',
            accent: 'oklch(0.94 0.03 321.94)',
            'accent-foreground': 'oklch(0.37 0.03 259.73)',
            destructive: 'oklch(0.81 0.10 19.57)',
            border: 'oklch(0.91 0.05 306.09)',
            input: 'oklch(0.91 0.05 306.09)',
            ring: 'oklch(0.71 0.16 293.54)',
            'chart-1': 'oklch(0.71 0.16 293.54)',
            'chart-2': 'oklch(0.61 0.22 292.72)',
            'chart-3': 'oklch(0.54 0.25 293.01)',
            'chart-4': 'oklch(0.49 0.24 292.58)',
            'chart-5': 'oklch(0.43 0.21 292.76)',
            sidebar: 'oklch(0.91 0.05 306.09)',
            'sidebar-foreground': 'oklch(0.37 0.03 259.73)',
            'sidebar-primary': 'oklch(0.71 0.16 293.54)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.94 0.03 321.94)',
            'sidebar-accent-foreground': 'oklch(0.37 0.03 259.73)',
            'sidebar-border': 'oklch(0.91 0.05 306.09)',
            'sidebar-ring': 'oklch(0.71 0.16 293.54)',
            'font-sans': 'Open Sans, sans-serif',
            'font-serif': 'Source Serif 4, serif',
            'font-mono': 'IBM Plex Mono, monospace',
            radius: '1.5rem',
            'shadow-color': 'hsl(0 0% 0%)',
            'shadow-opacity': '0.08',
            'shadow-blur': '16px',
            'shadow-spread': '-4px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '8px'
        },
        dark: {
            background: 'oklch(0.22 0.01 56.04)',
            foreground: 'oklch(0.93 0.03 272.79)',
            card: 'oklch(0.28 0.03 307.23)',
            'card-foreground': 'oklch(0.93 0.03 272.79)',
            popover: 'oklch(0.28 0.03 307.23)',
            'popover-foreground': 'oklch(0.93 0.03 272.79)',
            primary: 'oklch(0.79 0.12 295.75)',
            'primary-foreground': 'oklch(0.22 0.01 56.04)',
            secondary: 'oklch(0.34 0.04 308.85)',
            'secondary-foreground': 'oklch(0.87 0.01 258.34)',
            muted: 'oklch(0.28 0.03 307.23)',
            'muted-foreground': 'oklch(0.71 0.02 261.32)',
            accent: 'oklch(0.39 0.05 304.64)',
            'accent-foreground': 'oklch(0.87 0.01 258.34)',
            destructive: 'oklch(0.81 0.10 19.57)',
            border: 'oklch(0.34 0.04 308.85)',
            input: 'oklch(0.34 0.04 308.85)',
            ring: 'oklch(0.79 0.12 295.75)',
            'chart-1': 'oklch(0.79 0.12 295.75)',
            'chart-2': 'oklch(0.71 0.16 293.54)',
            'chart-3': 'oklch(0.61 0.22 292.72)',
            'chart-4': 'oklch(0.54 0.25 293.01)',
            'chart-5': 'oklch(0.49 0.24 292.58)',
            sidebar: 'oklch(0.34 0.04 308.85)',
            'sidebar-foreground': 'oklch(0.93 0.03 272.79)',
            'sidebar-primary': 'oklch(0.79 0.12 295.75)',
            'sidebar-primary-foreground': 'oklch(0.22 0.01 56.04)',
            'sidebar-accent': 'oklch(0.39 0.05 304.64)',
            'sidebar-accent-foreground': 'oklch(0.87 0.01 258.34)',
            'sidebar-border': 'oklch(0.34 0.04 308.85)',
            'sidebar-ring': 'oklch(0.79 0.12 295.75)'
        }
    },
    'clean-slate': {
        light: {
            background: 'oklch(0.98 0.00 247.86)',
            foreground: 'oklch(0.28 0.04 260.03)',
            card: 'oklch(1.00 0 0)',
            'card-foreground': 'oklch(0.28 0.04 260.03)',
            popover: 'oklch(1.00 0 0)',
            'popover-foreground': 'oklch(0.28 0.04 260.03)',
            primary: 'oklch(0.59 0.20 277.12)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.93 0.01 264.53)',
            'secondary-foreground': 'oklch(0.37 0.03 259.73)',
            muted: 'oklch(0.97 0.00 264.54)',
            'muted-foreground': 'oklch(0.55 0.02 264.36)',
            accent: 'oklch(0.93 0.03 272.79)',
            'accent-foreground': 'oklch(0.37 0.03 259.73)',
            destructive: 'oklch(0.64 0.21 25.33)',
            border: 'oklch(0.87 0.01 258.34)',
            input: 'oklch(0.87 0.01 258.34)',
            ring: 'oklch(0.59 0.20 277.12)',
            'chart-1': 'oklch(0.59 0.20 277.12)',
            'chart-2': 'oklch(0.51 0.23 276.97)',
            'chart-3': 'oklch(0.46 0.21 277.02)',
            'chart-4': 'oklch(0.40 0.18 277.37)',
            'chart-5': 'oklch(0.36 0.14 278.70)',
            sidebar: 'oklch(0.97 0.00 264.54)',
            'sidebar-foreground': 'oklch(0.28 0.04 260.03)',
            'sidebar-primary': 'oklch(0.59 0.20 277.12)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.93 0.03 272.79)',
            'sidebar-accent-foreground': 'oklch(0.37 0.03 259.73)',
            'sidebar-border': 'oklch(0.87 0.01 258.34)',
            'sidebar-ring': 'oklch(0.59 0.20 277.12)',
            'font-sans': 'Inter, sans-serif',
            'font-serif': 'Merriweather, serif',
            'font-mono': 'JetBrains Mono, monospace',
            radius: '0.5rem',
            'shadow-color': 'hsl(0 0% 0%)',
            'shadow-opacity': '0.1',
            'shadow-blur': '8px',
            'shadow-spread': '-1px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '4px'
        },
        dark: {
            background: 'oklch(0.21 0.04 265.75)',
            foreground: 'oklch(0.93 0.01 255.51)',
            card: 'oklch(0.28 0.04 260.03)',
            'card-foreground': 'oklch(0.93 0.01 255.51)',
            popover: 'oklch(0.28 0.04 260.03)',
            'popover-foreground': 'oklch(0.93 0.01 255.51)',
            primary: 'oklch(0.68 0.16 276.93)',
            'primary-foreground': 'oklch(0.21 0.04 265.75)',
            secondary: 'oklch(0.34 0.03 260.91)',
            'secondary-foreground': 'oklch(0.87 0.01 258.34)',
            muted: 'oklch(0.28 0.04 260.03)',
            'muted-foreground': 'oklch(0.71 0.02 261.32)',
            accent: 'oklch(0.37 0.03 259.73)',
            'accent-foreground': 'oklch(0.87 0.01 258.34)',
            destructive: 'oklch(0.64 0.21 25.33)',
            border: 'oklch(0.45 0.03 256.80)',
            input: 'oklch(0.45 0.03 256.80)',
            ring: 'oklch(0.68 0.16 276.93)',
            'chart-1': 'oklch(0.68 0.16 276.93)',
            'chart-2': 'oklch(0.59 0.20 277.12)',
            'chart-3': 'oklch(0.51 0.23 276.97)',
            'chart-4': 'oklch(0.46 0.21 277.02)',
            'chart-5': 'oklch(0.40 0.18 277.37)',
            sidebar: 'oklch(0.28 0.04 260.03)',
            'sidebar-foreground': 'oklch(0.93 0.01 255.51)',
            'sidebar-primary': 'oklch(0.68 0.16 276.93)',
            'sidebar-primary-foreground': 'oklch(0.21 0.04 265.75)',
            'sidebar-accent': 'oklch(0.37 0.03 259.73)',
            'sidebar-accent-foreground': 'oklch(0.87 0.01 258.34)',
            'sidebar-border': 'oklch(0.45 0.03 256.80)',
            'sidebar-ring': 'oklch(0.68 0.16 276.93)'
        }
    },
    'midnight-bloom': {
        light: {
            background: 'oklch(0.98 0 0)',
            foreground: 'oklch(0.32 0 0)',
            card: 'oklch(1.00 0 0)',
            'card-foreground': 'oklch(0.32 0 0)',
            popover: 'oklch(1.00 0 0)',
            'popover-foreground': 'oklch(0.32 0 0)',
            primary: 'oklch(0.57 0.20 283.08)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.82 0.07 249.35)',
            'secondary-foreground': 'oklch(0.32 0 0)',
            muted: 'oklch(0.82 0.02 91.62)',
            'muted-foreground': 'oklch(0.54 0 0)',
            accent: 'oklch(0.65 0.06 117.43)',
            'accent-foreground': 'oklch(1.00 0 0)',
            destructive: 'oklch(0.64 0.21 25.33)',
            border: 'oklch(0.87 0 0)',
            input: 'oklch(0.87 0 0)',
            ring: 'oklch(0.57 0.20 283.08)',
            'chart-1': 'oklch(0.57 0.20 283.08)',
            'chart-2': 'oklch(0.53 0.17 314.65)',
            'chart-3': 'oklch(0.34 0.18 301.68)',
            'chart-4': 'oklch(0.67 0.14 261.34)',
            'chart-5': 'oklch(0.59 0.10 245.74)',
            sidebar: 'oklch(0.98 0 0)',
            'sidebar-foreground': 'oklch(0.32 0 0)',
            'sidebar-primary': 'oklch(0.57 0.20 283.08)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.65 0.06 117.43)',
            'sidebar-accent-foreground': 'oklch(1.00 0 0)',
            'sidebar-border': 'oklch(0.87 0 0)',
            'sidebar-ring': 'oklch(0.57 0.20 283.08)',
            'font-sans': 'Montserrat, sans-serif',
            'font-serif': 'Playfair Display, serif',
            'font-mono': 'Source Code Pro, monospace',
            radius: '0.5rem',
            'shadow-color': 'hsl(0 0% 0%)',
            'shadow-opacity': '0.1',
            'shadow-blur': '10px',
            'shadow-spread': '-2px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '5px'
        },
        dark: {
            background: 'oklch(0.23 0.01 264.29)',
            foreground: 'oklch(0.92 0 0)',
            card: 'oklch(0.32 0.01 223.67)',
            'card-foreground': 'oklch(0.92 0 0)',
            popover: 'oklch(0.32 0.01 223.67)',
            'popover-foreground': 'oklch(0.92 0 0)',
            primary: 'oklch(0.57 0.20 283.08)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.34 0.18 301.68)',
            'secondary-foreground': 'oklch(0.92 0 0)',
            muted: 'oklch(0.39 0 0)',
            'muted-foreground': 'oklch(0.72 0 0)',
            accent: 'oklch(0.67 0.14 261.34)',
            'accent-foreground': 'oklch(0.92 0 0)',
            destructive: 'oklch(0.64 0.21 25.33)',
            border: 'oklch(0.39 0 0)',
            input: 'oklch(0.39 0 0)',
            ring: 'oklch(0.57 0.20 283.08)',
            'chart-1': 'oklch(0.57 0.20 283.08)',
            'chart-2': 'oklch(0.53 0.17 314.65)',
            'chart-3': 'oklch(0.34 0.18 301.68)',
            'chart-4': 'oklch(0.67 0.14 261.34)',
            'chart-5': 'oklch(0.59 0.10 245.74)',
            sidebar: 'oklch(0.23 0.01 264.29)',
            'sidebar-foreground': 'oklch(0.92 0 0)',
            'sidebar-primary': 'oklch(0.57 0.20 283.08)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.67 0.14 261.34)',
            'sidebar-accent-foreground': 'oklch(0.92 0 0)',
            'sidebar-border': 'oklch(0.39 0 0)',
            'sidebar-ring': 'oklch(0.57 0.20 283.08)'
        }
    },
    'sunset-horizon': {
        light: {
            background: 'oklch(0.99 0.01 56.32)',
            foreground: 'oklch(0.34 0.01 2.77)',
            card: 'oklch(1.00 0 0)',
            'card-foreground': 'oklch(0.34 0.01 2.77)',
            popover: 'oklch(1.00 0 0)',
            'popover-foreground': 'oklch(0.34 0.01 2.77)',
            primary: 'oklch(0.74 0.16 34.71)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.96 0.02 28.90)',
            'secondary-foreground': 'oklch(0.56 0.13 32.74)',
            muted: 'oklch(0.97 0.02 39.40)',
            'muted-foreground': 'oklch(0.49 0.05 26.45)',
            accent: 'oklch(0.83 0.11 58.00)',
            'accent-foreground': 'oklch(0.34 0.01 2.77)',
            destructive: 'oklch(0.61 0.21 22.24)',
            border: 'oklch(0.93 0.04 38.69)',
            input: 'oklch(0.93 0.04 38.69)',
            ring: 'oklch(0.74 0.16 34.71)',
            'chart-1': 'oklch(0.74 0.16 34.71)',
            'chart-2': 'oklch(0.83 0.11 58.00)',
            'chart-3': 'oklch(0.88 0.08 54.93)',
            'chart-4': 'oklch(0.82 0.11 40.89)',
            'chart-5': 'oklch(0.64 0.13 32.07)',
            sidebar: 'oklch(0.97 0.02 39.40)',
            'sidebar-foreground': 'oklch(0.34 0.01 2.77)',
            'sidebar-primary': 'oklch(0.74 0.16 34.71)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.83 0.11 58.00)',
            'sidebar-accent-foreground': 'oklch(0.34 0.01 2.77)',
            'sidebar-border': 'oklch(0.93 0.04 38.69)',
            'sidebar-ring': 'oklch(0.74 0.16 34.71)',
            'font-sans': 'Montserrat, sans-serif',
            'font-serif': 'Merriweather, serif',
            'font-mono': 'Ubuntu Mono, monospace',
            radius: '0.625rem',
            'shadow-color': 'hsl(0 0% 0%)',
            'shadow-opacity': '0.09',
            'shadow-blur': '12px',
            'shadow-spread': '-3px',
            'shadow-offset-x': '0px',
            'shadow-offset-y': '6px'
        },
        dark: {
            background: 'oklch(0.26 0.02 352.40)',
            foreground: 'oklch(0.94 0.01 51.32)',
            card: 'oklch(0.32 0.02 341.45)',
            'card-foreground': 'oklch(0.94 0.01 51.32)',
            popover: 'oklch(0.32 0.02 341.45)',
            'popover-foreground': 'oklch(0.94 0.01 51.32)',
            primary: 'oklch(0.74 0.16 34.71)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.36 0.02 342.27)',
            'secondary-foreground': 'oklch(0.94 0.01 51.32)',
            muted: 'oklch(0.32 0.02 341.45)',
            'muted-foreground': 'oklch(0.84 0.02 52.63)',
            accent: 'oklch(0.83 0.11 58.00)',
            'accent-foreground': 'oklch(0.26 0.02 352.40)',
            destructive: 'oklch(0.61 0.21 22.24)',
            border: 'oklch(0.36 0.02 342.27)',
            input: 'oklch(0.36 0.02 342.27)',
            ring: 'oklch(0.74 0.16 34.71)',
            'chart-1': 'oklch(0.74 0.16 34.71)',
            'chart-2': 'oklch(0.83 0.11 58.00)',
            'chart-3': 'oklch(0.88 0.08 54.93)',
            'chart-4': 'oklch(0.82 0.11 40.89)',
            'chart-5': 'oklch(0.64 0.13 32.07)',
            sidebar: 'oklch(0.26 0.02 352.40)',
            'sidebar-foreground': 'oklch(0.94 0.01 51.32)',
            'sidebar-primary': 'oklch(0.74 0.16 34.71)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.83 0.11 58.00)',
            'sidebar-accent-foreground': 'oklch(0.26 0.02 352.40)',
            'sidebar-border': 'oklch(0.36 0.02 342.27)',
            'sidebar-ring': 'oklch(0.74 0.16 34.71)'
        }
    },
    claude: {
        light: {
            background: 'oklch(0.98 0.01 95.10)',
            foreground: 'oklch(0.34 0.03 95.72)',
            card: 'oklch(0.98 0.01 95.10)',
            'card-foreground': 'oklch(0.19 0.00 106.59)',
            popover: 'oklch(1.00 0 0)',
            'popover-foreground': 'oklch(0.27 0.02 98.94)',
            primary: 'oklch(0.62 0.14 39.04)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.92 0.01 92.99)',
            'secondary-foreground': 'oklch(0.43 0.02 98.60)',
            muted: 'oklch(0.93 0.02 90.24)',
            'muted-foreground': 'oklch(0.61 0.01 97.42)',
            accent: 'oklch(0.92 0.01 92.99)',
            'accent-foreground': 'oklch(0.27 0.02 98.94)',
            destructive: 'oklch(0.19 0.00 106.59)',
            border: 'oklch(0.88 0.01 97.36)',
            input: 'oklch(0.76 0.02 98.35)',
            ring: 'oklch(0.59 0.17 253.06)',
            'chart-1': 'oklch(0.56 0.13 43.00)',
            'chart-2': 'oklch(0.69 0.16 290.41)',
            'chart-3': 'oklch(0.88 0.03 93.13)',
            'chart-4': 'oklch(0.88 0.04 298.18)',
            'chart-5': 'oklch(0.56 0.13 42.06)',
            sidebar: 'oklch(0.97 0.01 98.88)',
            'sidebar-foreground': 'oklch(0.36 0.01 106.65)',
            'sidebar-primary': 'oklch(0.62 0.14 39.04)',
            'sidebar-primary-foreground': 'oklch(0.99 0 0)',
            'sidebar-accent': 'oklch(0.92 0.01 92.99)',
            'sidebar-accent-foreground': 'oklch(0.33 0 0)',
            'sidebar-border': 'oklch(0.94 0 0)',
            'sidebar-ring': 'oklch(0.77 0 0)',
            radius: '0.5rem'
        },
        dark: {
            background: 'oklch(0.27 0.00 106.64)',
            foreground: 'oklch(0.81 0.01 93.01)',
            card: 'oklch(0.27 0.00 106.64)',
            'card-foreground': 'oklch(0.98 0.01 95.10)',
            popover: 'oklch(0.31 0.00 106.60)',
            'popover-foreground': 'oklch(0.92 0.00 106.48)',
            primary: 'oklch(0.67 0.13 38.76)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.98 0.01 95.10)',
            'secondary-foreground': 'oklch(0.31 0.00 106.60)',
            muted: 'oklch(0.22 0.00 106.71)',
            'muted-foreground': 'oklch(0.77 0.02 99.07)',
            accent: 'oklch(0.21 0.01 95.42)',
            'accent-foreground': 'oklch(0.97 0.01 98.88)',
            destructive: 'oklch(0.64 0.21 25.33)',
            border: 'oklch(0.36 0.01 106.89)',
            input: 'oklch(0.43 0.01 100.22)',
            ring: 'oklch(0.59 0.17 253.06)',
            'chart-1': 'oklch(0.56 0.13 43.00)',
            'chart-2': 'oklch(0.69 0.16 290.41)',
            'chart-3': 'oklch(0.21 0.01 95.42)',
            'chart-4': 'oklch(0.31 0.05 289.32)',
            'chart-5': 'oklch(0.56 0.13 42.06)',
            sidebar: 'oklch(0.24 0.00 67.71)',
            'sidebar-foreground': 'oklch(0.81 0.01 93.01)',
            'sidebar-primary': 'oklch(0.33 0 0)',
            'sidebar-primary-foreground': 'oklch(0.99 0 0)',
            'sidebar-accent': 'oklch(0.17 0.00 106.62)',
            'sidebar-accent-foreground': 'oklch(0.81 0.01 93.01)',
            'sidebar-border': 'oklch(0.94 0 0)',
            'sidebar-ring': 'oklch(0.77 0 0)'
        }
    },
    caffeine: {
        light: {
            background: 'oklch(0.98 0 0)',
            foreground: 'oklch(0.24 0 0)',
            card: 'oklch(0.99 0 0)',
            'card-foreground': 'oklch(0.24 0 0)',
            popover: 'oklch(0.99 0 0)',
            'popover-foreground': 'oklch(0.24 0 0)',
            primary: 'oklch(0.43 0.04 41.99)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.92 0.07 74.37)',
            'secondary-foreground': 'oklch(0.35 0.07 40.83)',
            muted: 'oklch(0.95 0 0)',
            'muted-foreground': 'oklch(0.50 0 0)',
            accent: 'oklch(0.93 0 0)',
            'accent-foreground': 'oklch(0.24 0 0)',
            destructive: 'oklch(0.63 0.19 33.34)',
            border: 'oklch(0.88 0 0)',
            input: 'oklch(0.88 0 0)',
            ring: 'oklch(0.43 0.04 41.99)',
            'chart-1': 'oklch(0.43 0.04 41.99)',
            'chart-2': 'oklch(0.92 0.07 74.37)',
            'chart-3': 'oklch(0.93 0 0)',
            'chart-4': 'oklch(0.94 0.05 75.50)',
            'chart-5': 'oklch(0.43 0.04 41.67)',
            sidebar: 'oklch(0.99 0 0)',
            'sidebar-foreground': 'oklch(0.26 0 0)',
            'sidebar-primary': 'oklch(0.33 0 0)',
            'sidebar-primary-foreground': 'oklch(0.99 0 0)',
            'sidebar-accent': 'oklch(0.98 0 0)',
            'sidebar-accent-foreground': 'oklch(0.33 0 0)',
            'sidebar-border': 'oklch(0.94 0 0)',
            'sidebar-ring': 'oklch(0.77 0 0)',
            radius: '0.5rem'
        },
        dark: {
            background: 'oklch(0.18 0 0)',
            foreground: 'oklch(0.95 0 0)',
            card: 'oklch(0.21 0 0)',
            'card-foreground': 'oklch(0.95 0 0)',
            popover: 'oklch(0.21 0 0)',
            'popover-foreground': 'oklch(0.95 0 0)',
            primary: 'oklch(0.92 0.05 66.17)',
            'primary-foreground': 'oklch(0.20 0.02 200.20)',
            secondary: 'oklch(0.32 0.02 63.70)',
            'secondary-foreground': 'oklch(0.92 0.05 66.17)',
            muted: 'oklch(0.25 0 0)',
            'muted-foreground': 'oklch(0.77 0 0)',
            accent: 'oklch(0.29 0 0)',
            'accent-foreground': 'oklch(0.95 0 0)',
            destructive: 'oklch(0.63 0.19 33.34)',
            border: 'oklch(0.24 0.01 91.75)',
            input: 'oklch(0.40 0 0)',
            ring: 'oklch(0.92 0.05 66.17)',
            'chart-1': 'oklch(0.92 0.05 66.17)',
            'chart-2': 'oklch(0.32 0.02 63.70)',
            'chart-3': 'oklch(0.29 0 0)',
            'chart-4': 'oklch(0.35 0.02 67.00)',
            'chart-5': 'oklch(0.92 0.05 67.09)',
            sidebar: 'oklch(0.21 0.01 285.89)',
            'sidebar-foreground': 'oklch(0.97 0.00 286.38)',
            'sidebar-primary': 'oklch(0.49 0.22 264.38)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.27 0.01 286.03)',
            'sidebar-accent-foreground': 'oklch(0.97 0.00 286.38)',
            'sidebar-border': 'oklch(0.27 0.01 286.03)',
            'sidebar-ring': 'oklch(0.87 0.01 286.29)'
        }
    },
    corporate: {
        light: {
            background: 'oklch(0.98 0 0)',
            foreground: 'oklch(0.21 0.03 264.67)',
            card: 'oklch(1.00 0 0)',
            'card-foreground': 'oklch(0.21 0.03 264.67)',
            popover: 'oklch(1.00 0 0)',
            'popover-foreground': 'oklch(0.21 0.03 264.67)',
            primary: 'oklch(0.48 0.20 260.48)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.97 0.00 264.70)',
            'secondary-foreground': 'oklch(0.37 0.03 259.73)',
            muted: 'oklch(0.97 0.00 264.70)',
            'muted-foreground': 'oklch(0.55 0.02 264.37)',
            accent: 'oklch(0.95 0.02 261.78)',
            'accent-foreground': 'oklch(0.48 0.20 260.48)',
            destructive: 'oklch(0.58 0.22 27.33)',
            border: 'oklch(0.93 0.01 264.60)',
            input: 'oklch(0.93 0.01 264.60)',
            ring: 'oklch(0.48 0.20 260.48)',
            'chart-1': 'oklch(0.48 0.20 260.48)',
            'chart-2': 'oklch(0.56 0.24 260.95)',
            'chart-3': 'oklch(0.40 0.16 259.09)',
            'chart-4': 'oklch(0.43 0.16 259.85)',
            'chart-5': 'oklch(0.29 0.07 260.37)',
            sidebar: 'oklch(0.97 0.00 264.70)',
            'sidebar-foreground': 'oklch(0.21 0.03 264.67)',
            'sidebar-primary': 'oklch(0.48 0.20 260.48)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.95 0.02 261.78)',
            'sidebar-accent-foreground': 'oklch(0.48 0.20 260.48)',
            'sidebar-border': 'oklch(0.93 0.01 264.60)',
            'sidebar-ring': 'oklch(0.48 0.20 260.48)',
            'font-sans': 'Inter, sans-serif',
            'font-serif': 'Source Serif 4, serif',
            'font-mono': 'IBM Plex Mono, monospace',
            radius: '0.375rem'
        },
        dark: {
            background: 'oklch(0.26 0.03 262.71)',
            foreground: 'oklch(0.93 0.01 264.60)',
            card: 'oklch(0.30 0.03 261.75)',
            'card-foreground': 'oklch(0.93 0.01 264.60)',
            popover: 'oklch(0.30 0.03 261.75)',
            'popover-foreground': 'oklch(0.93 0.01 264.60)',
            primary: 'oklch(0.56 0.24 260.95)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.35 0.04 262.16)',
            'secondary-foreground': 'oklch(0.93 0.01 264.60)',
            muted: 'oklch(0.30 0.03 261.75)',
            'muted-foreground': 'oklch(0.71 0.02 261.33)',
            accent: 'oklch(0.33 0.04 264.82)',
            'accent-foreground': 'oklch(0.93 0.01 264.60)',
            destructive: 'oklch(0.64 0.21 25.33)',
            border: 'oklch(0.35 0.04 262.16)',
            input: 'oklch(0.35 0.04 262.16)',
            ring: 'oklch(0.56 0.24 260.95)',
            'chart-1': 'oklch(0.56 0.24 260.95)',
            'chart-2': 'oklch(0.48 0.20 260.48)',
            'chart-3': 'oklch(0.69 0.17 256.00)',
            'chart-4': 'oklch(0.43 0.16 259.85)',
            'chart-5': 'oklch(0.29 0.07 260.37)',
            sidebar: 'oklch(0.26 0.03 262.71)',
            'sidebar-foreground': 'oklch(0.93 0.01 264.60)',
            'sidebar-primary': 'oklch(0.56 0.24 260.95)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.33 0.04 264.82)',
            'sidebar-accent-foreground': 'oklch(0.93 0.01 264.60)',
            'sidebar-border': 'oklch(0.35 0.04 262.16)',
            'sidebar-ring': 'oklch(0.56 0.24 260.95)'
        }
    },
    slack: {
        light: {
            background: 'oklch(1.00 0 0)',
            foreground: 'oklch(0.23 0.00 325.86)',
            card: 'oklch(0.98 0 0)',
            'card-foreground': 'oklch(0.23 0.00 325.86)',
            popover: 'oklch(0.98 0 0)',
            'popover-foreground': 'oklch(0.23 0.00 325.86)',
            primary: 'oklch(0.37 0.14 323.23)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.96 0.01 312.56)',
            'secondary-foreground': 'oklch(0.31 0.11 327.07)',
            muted: 'oklch(0.97 0.00 264.70)',
            'muted-foreground': 'oklch(0.49 0 0)',
            accent: 'oklch(0.88 0.02 323.34)',
            'accent-foreground': 'oklch(0.31 0.11 327.07)',
            destructive: 'oklch(0.59 0.22 11.50)',
            border: 'oklch(0.91 0 0)',
            input: 'oklch(0.91 0 0)',
            ring: 'oklch(0.37 0.14 323.23)',
            'chart-1': 'oklch(0.31 0.11 327.07)',
            'chart-2': 'oklch(0.37 0.14 323.23)',
            'chart-3': 'oklch(0.59 0.22 11.50)',
            'chart-4': 'oklch(0.77 0.13 223.19)',
            'chart-5': 'oklch(0.69 0.14 160.23)',
            sidebar: 'oklch(0.96 0.01 312.56)',
            'sidebar-foreground': 'oklch(0.23 0.00 325.86)',
            'sidebar-primary': 'oklch(0.37 0.14 323.23)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.88 0.02 323.34)',
            'sidebar-accent-foreground': 'oklch(0.31 0.11 327.07)',
            'sidebar-border': 'oklch(0.91 0 0)',
            'sidebar-ring': 'oklch(0.37 0.14 323.23)',
            'font-sans': 'Lato, sans-serif',
            'font-serif': 'Merriweather, serif',
            'font-mono': 'Roboto Mono, monospace',
            radius: '0.5rem'
        },
        dark: {
            background: 'oklch(0.23 0.01 255.60)',
            foreground: 'oklch(0.93 0 0)',
            card: 'oklch(0.26 0.01 255.58)',
            'card-foreground': 'oklch(0.93 0 0)',
            popover: 'oklch(0.26 0.01 255.58)',
            'popover-foreground': 'oklch(0.93 0 0)',
            primary: 'oklch(0.58 0.14 327.21)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.30 0.01 248.05)',
            'secondary-foreground': 'oklch(0.93 0 0)',
            muted: 'oklch(0.26 0.01 255.58)',
            'muted-foreground': 'oklch(0.68 0 0)',
            accent: 'oklch(0.33 0.03 326.23)',
            'accent-foreground': 'oklch(0.93 0 0)',
            destructive: 'oklch(0.59 0.22 11.50)',
            border: 'oklch(0.30 0.01 268.37)',
            input: 'oklch(0.30 0.01 268.37)',
            ring: 'oklch(0.58 0.14 327.21)',
            'chart-1': 'oklch(0.58 0.14 327.21)',
            'chart-2': 'oklch(0.77 0.13 223.19)',
            'chart-3': 'oklch(0.69 0.14 160.23)',
            'chart-4': 'oklch(0.59 0.22 11.50)',
            'chart-5': 'oklch(0.80 0.15 82.64)',
            sidebar: 'oklch(0.23 0.01 255.60)',
            'sidebar-foreground': 'oklch(0.93 0 0)',
            'sidebar-primary': 'oklch(0.58 0.14 327.21)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.33 0.03 326.23)',
            'sidebar-accent-foreground': 'oklch(0.93 0 0)',
            'sidebar-border': 'oklch(0.30 0.01 268.37)',
            'sidebar-ring': 'oklch(0.58 0.14 327.21)'
        }
    },
    gruvbox: {
        light: {
            background: 'hsl(52.94 73.91% 90.98%)',
            foreground: 'hsl(0.00 0.00% 15.69%)',
            card: 'hsl(45.56 67.50% 84.31%)',
            'card-foreground': 'hsl(0.00 0.00% 15.69%)',
            popover: 'hsl(45.56 67.50% 84.31%)',
            'popover-foreground': 'hsl(0.00 0.00% 15.69%)',
            primary: 'hsl(61.21, 66.22%, 44.12%)',
            'primary-foreground': 'hsl(52.94 73.91% 90.98%)',
            secondary: 'hsl(189.56 88.98% 24.90%)',
            'secondary-foreground': 'hsl(52.94 73.91% 90.98%)',
            muted: 'hsl(48.46 86.67% 88.24%)',
            'muted-foreground': 'hsl(27.50 10.71% 43.92%)',
            accent: 'hsl(48.46 86.67% 88.24%)',
            'accent-foreground': 'hsl(0.00 0.00% 15.69%)',
            destructive: 'hsl(357.71 100.00% 30.78%)',
            border: 'hsl(43.16 58.76% 80.98%)',
            input: 'hsl(43.16 58.76% 80.98%)',
            ring: 'hsl(19.19 96.63% 34.90%)',
            'chart-1': 'hsl(357.71 100.00% 30.78%)',
            'chart-2': 'hsl(57.20 79.26% 26.47%)',
            'chart-3': 'hsl(36.52 80.10% 39.41%)',
            'chart-4': 'hsl(189.56 88.98% 24.90%)',
            'chart-5': 'hsl(322.50 38.83% 40.39%)',
            sidebar: 'hsl(45.56 67.50% 84.31%)',
            'sidebar-foreground': 'hsl(0.00 0.00% 15.69%)',
            'sidebar-primary': 'hsl(19.19 96.63% 34.90%)',
            'sidebar-primary-foreground': 'hsl(52.94 73.91% 90.98%)',
            'sidebar-accent': 'hsl(48.46 86.67% 88.24%)',
            'sidebar-accent-foreground': 'hsl(0.00 0.00% 15.69%)',
            'sidebar-border': 'hsl(43.16 58.76% 80.98%)',
            'sidebar-ring': 'hsl(19.19 96.63% 34.90%)'
        },
        dark: {
            background: 'hsl(195.00 6.45% 12.16%)',
            foreground: 'hsl(48.46 86.67% 88.24%)',
            card: 'hsl(20.00 3.09% 19.02%)',
            'card-foreground': 'hsl(48.46 86.67% 88.24%)',
            popover: 'hsl(20.00 3.09% 19.02%)',
            'popover-foreground': 'hsl(48.46 86.67% 88.24%)',
            primary: 'hsl(61.21, 66.22%, 44.12%)',
            'primary-foreground': 'hsl(195.00 6.45% 12.16%)',
            secondary: 'hsl(39.56 73.39% 48.63%)',
            'secondary-foreground': 'hsl(195.00 6.45% 12.16%)',
            muted: 'hsl(20.00 5.26% 22.35%)',
            'muted-foreground': 'hsl(38.57 24.14% 65.88%)',
            accent: 'hsl(20.00 5.26% 22.35%)',
            'accent-foreground': 'hsl(48.46 86.67% 88.24%)',
            destructive: 'hsl(2.40 75.11% 45.69%)',
            border: 'hsl(21.82 7.38% 29.22%)',
            input: 'hsl(21.82 7.38% 29.22%)',
            ring: 'hsl(23.70 87.72% 44.71%)',
            'chart-1': 'hsl(2.40 75.11% 45.69%)',
            'chart-2': 'hsl(59.52 70.79% 34.90%)',
            'chart-3': 'hsl(39.56 73.39% 48.63%)',
            'chart-4': 'hsl(182.69 32.68% 40.20%)',
            'chart-5': 'hsl(332.66 33.62% 53.92%)',
            sidebar: 'hsl(20.00 3.09% 19.02%)',
            'sidebar-foreground': 'hsl(48.46 86.67% 88.24%)',
            'sidebar-primary': 'hsl(23.70 87.72% 44.71%)',
            'sidebar-primary-foreground': 'hsl(195.00 6.45% 12.16%)',
            'sidebar-accent': 'hsl(20.00 5.26% 22.35%)',
            'sidebar-accent-foreground': 'hsl(48.46 86.67% 88.24%)',
            'sidebar-border': 'hsl(21.82 7.38% 29.22%)',
            'sidebar-ring': 'hsl(23.70 87.72% 44.71%)'
        }
    },
    perplexity: {
        light: {
            background: 'oklch(0.95 0.01 196.81)',
            foreground: 'oklch(0.38 0.06 212.65)',
            card: 'oklch(0.97 0.01 196.73)',
            'card-foreground': 'oklch(0.38 0.06 212.65)',
            popover: 'oklch(0.97 0.01 196.73)',
            'popover-foreground': 'oklch(0.38 0.06 212.65)',
            primary: 'oklch(0.72 0.12 209.78)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.97 0.01 247.86)',
            'secondary-foreground': 'oklch(0.14 0.00 285.86)',
            muted: 'oklch(0.97 0.01 247.86)',
            'muted-foreground': 'oklch(0.55 0.04 257.42)',
            accent: 'oklch(0.96 0.02 205.23)',
            'accent-foreground': 'oklch(0.57 0.10 213.38)',
            destructive: 'oklch(0.64 0.21 25.33)',
            border: 'oklch(0.93 0.01 255.51)',
            input: 'oklch(0.93 0.01 255.51)',
            ring: 'oklch(0.72 0.12 209.78)',
            'chart-1': 'oklch(0.72 0.12 209.78)',
            'chart-2': 'oklch(0.57 0.10 213.38)',
            'chart-3': 'oklch(0.79 0.12 208.87)',
            'chart-4': 'oklch(0.76 0.11 208.84)',
            'chart-5': 'oklch(0.83 0.10 208.33)',
            sidebar: 'oklch(0.98 0.00 247.80)',
            'sidebar-foreground': 'oklch(0.14 0.00 285.86)',
            'sidebar-primary': 'oklch(0.72 0.12 209.78)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.96 0.02 205.23)',
            'sidebar-accent-foreground': 'oklch(0.57 0.10 213.38)',
            'sidebar-border': 'oklch(0.93 0.01 255.51)',
            'sidebar-ring': 'oklch(0.72 0.12 209.78)',
            'font-sans': 'Inter, sans-serif',
            'font-serif': 'Lora, serif',
            'font-mono': 'Roboto Mono, monospace',
            radius: '0.5rem'
        },
        dark: {
            background: 'oklch(0.21 0.02 224.44)',
            foreground: 'oklch(0.85 0.13 195.02)',
            card: 'oklch(0.23 0.03 216.05)',
            'card-foreground': 'oklch(0.85 0.13 195.02)',
            popover: 'oklch(0.23 0.03 216.05)',
            'popover-foreground': 'oklch(0.85 0.13 195.02)',
            primary: 'oklch(0.72 0.12 209.78)',
            'primary-foreground': 'oklch(1.00 0 0)',
            secondary: 'oklch(0.27 0.01 286.10)',
            'secondary-foreground': 'oklch(0.97 0.00 264.70)',
            muted: 'oklch(0.24 0 0)',
            'muted-foreground': 'oklch(0.71 0.01 286.14)',
            accent: 'oklch(0.24 0.00 286.20)',
            'accent-foreground': 'oklch(0.97 0.00 264.70)',
            destructive: 'oklch(0.64 0.21 25.33)',
            border: 'oklch(0.29 0.00 286.27)',
            input: 'oklch(0.29 0.00 286.27)',
            ring: 'oklch(0.72 0.12 209.78)',
            'chart-1': 'oklch(0.72 0.12 209.78)',
            'chart-2': 'oklch(0.79 0.12 208.87)',
            'chart-3': 'oklch(0.76 0.11 208.84)',
            'chart-4': 'oklch(0.83 0.10 208.33)',
            'chart-5': 'oklch(0.57 0.10 213.38)',
            sidebar: 'oklch(0.19 0 0)',
            'sidebar-foreground': 'oklch(0.97 0.00 264.70)',
            'sidebar-primary': 'oklch(0.72 0.12 209.78)',
            'sidebar-primary-foreground': 'oklch(1.00 0 0)',
            'sidebar-accent': 'oklch(0.24 0.00 286.20)',
            'sidebar-accent-foreground': 'oklch(0.97 0.00 264.70)',
            'sidebar-border': 'oklch(0.29 0.00 286.27)',
            'sidebar-ring': 'oklch(0.72 0.12 209.78)'
        }
    }
};
}),
"[project]/providers/color-vars-provider.tsx [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "ColorVarsProvider",
    ()=>ColorVarsProvider
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$theme$2d$presets$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/theme-presets.ts [app-ssr] (ecmascript)");
"use client";
;
;
function getStored(key) {
    if ("TURBOPACK compile-time truthy", 1) return null;
    //TURBOPACK unreachable
    ;
}
function buildCss() {
    const presetName = getStored("osmedeus_theme_preset");
    const preset = presetName ? __TURBOPACK__imported__module__$5b$project$5d2f$theme$2d$presets$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["presets"][presetName] : undefined;
    const lightPrimary = getStored("osmedeus_theme_light_primary");
    const lightSecondary = getStored("osmedeus_theme_light_secondary");
    const darkPrimary = getStored("osmedeus_theme_dark_primary");
    const darkSecondary = getStored("osmedeus_theme_dark_secondary");
    const lightVars = [];
    const darkVars = [];
    if (preset) {
        const light = preset.light || {};
        const dark = preset.dark || {};
        for (const [k, v] of Object.entries(light)){
            lightVars.push(`--${k}: ${v};`);
        }
        for (const [k, v] of Object.entries(dark)){
            darkVars.push(`--${k}: ${v};`);
        }
    }
    if (lightPrimary) {
        lightVars.push(`--primary: ${lightPrimary};`);
        lightVars.push(`--ring: ${lightPrimary};`);
        lightVars.push(`--sidebar-primary: ${lightPrimary};`);
        lightVars.push(`--sidebar-ring: ${lightPrimary};`);
    }
    if (lightSecondary) {
        lightVars.push(`--secondary: ${lightSecondary};`);
    }
    if (darkPrimary) {
        darkVars.push(`--primary: ${darkPrimary};`);
        darkVars.push(`--ring: ${darkPrimary};`);
        darkVars.push(`--sidebar-primary: ${darkPrimary};`);
        darkVars.push(`--sidebar-ring: ${darkPrimary};`);
    }
    if (darkSecondary) {
        darkVars.push(`--secondary: ${darkSecondary};`);
    }
    const lightBlock = lightVars.length ? `:root { ${lightVars.join(" ")} }` : "";
    const darkBlock = darkVars.length ? `.dark { ${darkVars.join(" ")} }` : "";
    const css = `${lightBlock}${darkBlock ? " " + darkBlock : ""}`;
    return css;
}
function applyCss(css) {
    if (typeof document === "undefined") return;
    let styleEl = document.getElementById("user-theme-colors");
    if (!styleEl) {
        styleEl = document.createElement("style");
        styleEl.id = "user-theme-colors";
        document.head.appendChild(styleEl);
    }
    styleEl.textContent = css;
}
function ColorVarsProvider() {
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useEffect"](()=>{
        const css = buildCss();
        if (css) applyCss(css);
        const handler = ()=>{
            const updated = buildCss();
            applyCss(updated);
        };
        window.addEventListener("osmedeus-theme-colors-updated", handler);
        return ()=>{
            window.removeEventListener("osmedeus-theme-colors-updated", handler);
        };
    }, []);
    return null;
}
}),
"[externals]/next/dist/server/app-render/action-async-storage.external.js [external] (next/dist/server/app-render/action-async-storage.external.js, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("next/dist/server/app-render/action-async-storage.external.js", () => require("next/dist/server/app-render/action-async-storage.external.js"));

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
"[externals]/util [external] (util, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("util", () => require("util"));

module.exports = mod;
}),
"[externals]/stream [external] (stream, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("stream", () => require("stream"));

module.exports = mod;
}),
"[externals]/path [external] (path, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("path", () => require("path"));

module.exports = mod;
}),
"[externals]/http [external] (http, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("http", () => require("http"));

module.exports = mod;
}),
"[externals]/https [external] (https, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("https", () => require("https"));

module.exports = mod;
}),
"[externals]/url [external] (url, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("url", () => require("url"));

module.exports = mod;
}),
"[externals]/fs [external] (fs, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("fs", () => require("fs"));

module.exports = mod;
}),
"[externals]/crypto [external] (crypto, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("crypto", () => require("crypto"));

module.exports = mod;
}),
"[externals]/http2 [external] (http2, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("http2", () => require("http2"));

module.exports = mod;
}),
"[externals]/assert [external] (assert, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("assert", () => require("assert"));

module.exports = mod;
}),
"[externals]/tty [external] (tty, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("tty", () => require("tty"));

module.exports = mod;
}),
"[externals]/os [external] (os, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("os", () => require("os"));

module.exports = mod;
}),
"[externals]/zlib [external] (zlib, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("zlib", () => require("zlib"));

module.exports = mod;
}),
"[externals]/events [external] (events, cjs)", ((__turbopack_context__, module, exports) => {

const mod = __turbopack_context__.x("events", () => require("events"));

module.exports = mod;
}),
"[project]/lib/api/demo-mode.ts [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

/**
 * Demo mode utility for runtime switching between mock and real API
 *
 * Priority:
 * 1. localStorage.osmedeus_demo_mode (runtime toggle)
 * 2. process.env.NEXT_PUBLIC_USE_MOCK (build-time fallback)
 */ /**
 * Check if demo mode is enabled (runtime check)
 * Call this in API functions instead of checking env var directly
 */ __turbopack_context__.s([
    "clearDemoModePreference",
    ()=>clearDemoModePreference,
    "getDemoModePreference",
    ()=>getDemoModePreference,
    "isDemoMode",
    ()=>isDemoMode,
    "setDemoMode",
    ()=>setDemoMode
]);
function isDemoMode() {
    // First check localStorage (runtime toggle)
    if ("TURBOPACK compile-time falsy", 0) //TURBOPACK unreachable
    ;
    // Fallback to env var (build-time setting)
    return process.env.NEXT_PUBLIC_USE_MOCK === "true";
}
function setDemoMode(enabled) {
    if ("TURBOPACK compile-time falsy", 0) //TURBOPACK unreachable
    ;
}
function getDemoModePreference() {
    if ("TURBOPACK compile-time falsy", 0) //TURBOPACK unreachable
    ;
    return null;
}
function clearDemoModePreference() {
    if ("TURBOPACK compile-time falsy", 0) //TURBOPACK unreachable
    ;
}
}),
"[project]/lib/api/prefix.ts [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "API_PREFIX",
    ()=>API_PREFIX
]);
const API_PREFIX = "/osm/api";
}),
"[project]/lib/api/http.ts [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "getHttpBaseURL",
    ()=>getHttpBaseURL,
    "http",
    ()=>http
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$axios$2f$lib$2f$axios$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/axios/lib/axios.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$demo$2d$mode$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/api/demo-mode.ts [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$prefix$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/api/prefix.ts [app-ssr] (ecmascript)");
;
;
;
const MOCK_API_PREFIX = "/api/mock/api";
function resolveBaseURL() {
    if ((0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$demo$2d$mode$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["isDemoMode"])()) return "";
    if ("TURBOPACK compile-time falsy", 0) //TURBOPACK unreachable
    ;
    const envUrl = process.env.BASE_API_URL || process.env.NEXT_PUBLIC_API_URL;
    if (envUrl) return envUrl.replace(/\/+$/, "");
    if ("TURBOPACK compile-time falsy", 0) //TURBOPACK unreachable
    ;
    return "";
}
const http = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$axios$2f$lib$2f$axios$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["default"].create({
    baseURL: resolveBaseURL(),
    headers: {
        "Content-Type": "application/json"
    }
});
http.interceptors.request.use((config)=>{
    if ((0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$demo$2d$mode$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["isDemoMode"])()) {
        config.baseURL = "";
        if (typeof config.url === "string") {
            if (config.url === __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$prefix$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["API_PREFIX"]) {
                config.url = MOCK_API_PREFIX;
            } else if (config.url.startsWith(`${__TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$prefix$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["API_PREFIX"]}/`)) {
                config.url = `${MOCK_API_PREFIX}${config.url.slice(__TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$prefix$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["API_PREFIX"].length)}`;
            }
        }
    }
    if ("TURBOPACK compile-time falsy", 0) //TURBOPACK unreachable
    ;
    return config;
});
http.interceptors.response.use((response)=>response, (error)=>{
    const status = error?.response?.status;
    const message = error?.response?.data?.message || error?.message || "Request failed";
    if ("TURBOPACK compile-time falsy", 0) //TURBOPACK unreachable
    ;
    return Promise.reject(new Error(`${status || 0}:${message}`));
});
function getHttpBaseURL() {
    return http.defaults.baseURL || "";
}
}),
"[project]/lib/api/auth.ts [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "login",
    ()=>login
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$http$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/api/http.ts [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$prefix$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/api/prefix.ts [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$demo$2d$mode$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/api/demo-mode.ts [app-ssr] (ecmascript)");
;
;
;
async function login(username, password) {
    if ((0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$demo$2d$mode$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["isDemoMode"])()) {
        const token = "mock-" + Buffer.from(username).toString("base64");
        return token;
    }
    const res = await __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$http$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["http"].post(`${__TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$prefix$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["API_PREFIX"]}/login`, {
        username,
        password
    });
    return res.data?.token;
}
}),
"[project]/providers/auth-provider.tsx [app-ssr] (ecmascript)", ((__turbopack_context__) => {
"use strict";

__turbopack_context__.s([
    "AuthProvider",
    ()=>AuthProvider,
    "useAuth",
    ()=>useAuth
]);
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react-jsx-dev-runtime.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/dist/server/route-modules/app-page/vendored/ssr/react.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$navigation$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/node_modules/next/navigation.js [app-ssr] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$auth$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__ = __turbopack_context__.i("[project]/lib/api/auth.ts [app-ssr] (ecmascript)");
"use client";
;
;
;
;
const AuthContext = /*#__PURE__*/ __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["createContext"](undefined);
const PUBLIC_PATHS = [
    "/login"
];
function AuthProvider({ children }) {
    const [user, setUser] = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useState"](null);
    const [isLoading, setIsLoading] = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useState"](true);
    const router = (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$navigation$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useRouter"])();
    const pathname = (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$navigation$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["usePathname"])();
    const DISABLE_AUTH = typeof process !== "undefined" && ("TURBOPACK compile-time value", "true") === "true";
    // Check for existing session on mount
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useEffect"](()=>{
        const checkSession = ()=>{
            try {
                const stored = localStorage.getItem("osmedeus_session");
                if (stored) {
                    const parsed = JSON.parse(stored);
                    setUser(parsed);
                }
            } catch  {
                localStorage.removeItem("osmedeus_session");
            } finally{
                setIsLoading(false);
            }
        };
        checkSession();
    }, []);
    // Redirect logic
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useEffect"](()=>{
        if (isLoading) return;
        if (DISABLE_AUTH) return;
        const isPublicPath = PUBLIC_PATHS.includes(pathname);
        if (!user && !isPublicPath) {
            router.push("/login");
        } else if (user && isPublicPath) {
            router.push("/");
        }
    }, [
        user,
        isLoading,
        pathname,
        router
    ]);
    const login = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useCallback"](async (username, password)=>{
        const token = await (0, __TURBOPACK__imported__module__$5b$project$5d2f$lib$2f$api$2f$auth$2e$ts__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["login"])(username, password);
        localStorage.setItem("osmedeus_token", token);
        const userData = {
            id: `user-${Date.now()}`,
            username,
            email: `${username}@osmedeus.io`,
            name: username.charAt(0).toUpperCase() + username.slice(1)
        };
        localStorage.setItem("osmedeus_session", JSON.stringify(userData));
        setUser(userData);
        router.push("/");
    }, [
        router
    ]);
    const logout = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useCallback"](()=>{
        localStorage.removeItem("osmedeus_token");
        localStorage.removeItem("osmedeus_session");
        setUser(null);
        router.push("/login");
    }, [
        router
    ]);
    __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useEffect"](()=>{
        if (!isLoading && DISABLE_AUTH && !user) {
            setUser({
                id: "guest",
                username: "guest",
                email: "guest@osmedeus.io",
                name: "Guest"
            });
        }
    }, [
        isLoading,
        DISABLE_AUTH,
        user
    ]);
    return /*#__PURE__*/ (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2d$jsx$2d$dev$2d$runtime$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["jsxDEV"])(AuthContext.Provider, {
        value: {
            user,
            isAuthenticated: DISABLE_AUTH || !!user,
            isLoading,
            login,
            logout
        },
        children: children
    }, void 0, false, {
        fileName: "[project]/providers/auth-provider.tsx",
        lineNumber: 104,
        columnNumber: 5
    }, this);
}
function useAuth() {
    const context = __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2f$dist$2f$server$2f$route$2d$modules$2f$app$2d$page$2f$vendored$2f$ssr$2f$react$2e$js__$5b$app$2d$ssr$5d$__$28$ecmascript$29$__["useContext"](AuthContext);
    if (context === undefined) {
        throw new Error("useAuth must be used within an AuthProvider");
    }
    return context;
}
}),
];

//# sourceMappingURL=%5Broot-of-the-server%5D__effc8399._.js.map