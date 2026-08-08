"use client";

import * as React from "react";
import { presets } from "@/theme-presets";

function getStored(key: string) {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(key);
}

function buildCss() {
  const presetName = getStored("osmedeus_theme_preset");
  const preset = presetName ? presets[presetName] : undefined;

  const lightPrimary = getStored("osmedeus_theme_light_primary");
  const lightSecondary = getStored("osmedeus_theme_light_secondary");
  const darkPrimary = getStored("osmedeus_theme_dark_primary");
  const darkSecondary = getStored("osmedeus_theme_dark_secondary");

  const lightVars: string[] = [];
  const darkVars: string[] = [];

  if (preset) {
    const light = preset.light || {};
    const dark = preset.dark || {};
    for (const [k, v] of Object.entries(light)) {
      lightVars.push(`--${k}: ${v};`);
    }
    for (const [k, v] of Object.entries(dark)) {
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

function applyCss(css: string) {
  if (typeof document === "undefined") return;
  let styleEl = document.getElementById("user-theme-colors") as HTMLStyleElement | null;
  if (!styleEl) {
    styleEl = document.createElement("style");
    styleEl.id = "user-theme-colors";
    document.head.appendChild(styleEl);
  }
  styleEl.textContent = css;
}

export function ColorVarsProvider() {
  React.useEffect(() => {
    const css = buildCss();
    if (css) applyCss(css);

    const handler = () => {
      const updated = buildCss();
      applyCss(updated);
    };
    window.addEventListener("osmedeus-theme-colors-updated", handler as EventListener);
    return () => {
      window.removeEventListener("osmedeus-theme-colors-updated", handler as EventListener);
    };
  }, []);
  return null;
}
