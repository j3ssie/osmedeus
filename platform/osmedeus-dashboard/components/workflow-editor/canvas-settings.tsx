"use client";

import * as React from "react";

type CanvasSettings = {
  wrapLongText: boolean;
  showDetails: boolean;
};

const CanvasSettingsContext = React.createContext<CanvasSettings | null>(null);

export function CanvasSettingsProvider({
  wrapLongText,
  showDetails,
  children,
}: {
  wrapLongText: boolean;
  showDetails: boolean;
  children: React.ReactNode;
}) {
  return (
    <CanvasSettingsContext.Provider value={{ wrapLongText, showDetails }}>
      {children}
    </CanvasSettingsContext.Provider>
  );
}

export function useCanvasSettings(): CanvasSettings {
  return React.useContext(CanvasSettingsContext) ?? { wrapLongText: false, showDetails: true };
}
