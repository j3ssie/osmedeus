import type { Metadata } from "next";
import favicon from "@/favicon.ico";
import { ThemeProvider } from "@/providers/theme-provider";
import { ColorVarsProvider } from "@/providers/color-vars-provider";
import { AuthProvider } from "@/providers/auth-provider";
import { Toaster } from "sonner";
import "./globals.css";

export const metadata: Metadata = {
  title: "Osmedeus Dashboard",
  description: "Security scan management dashboard for Osmedeus Workflow Engine",
  icons: {
    icon: favicon.src,
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      {/* The page plane sits *under* the app canvas: chrome and auth screens
          show `--og-page-bg`, and `SidebarInset` paints `bg-background` over
          it for the dashboard column. */}
      <body className="min-h-screen bg-page antialiased">
        {/*
          No `disableTransitionOnChange`: the palette swap is cross-faded
          instead, by `components/layout/theme-toggle.tsx` setting
          `data-theme-transition` on <html> for the length of the swap.
        */}
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          <ColorVarsProvider />
          <AuthProvider>
            {children}
            <Toaster
              position="bottom-right"
              richColors
              toastOptions={{
                className: "border border-border",
              }}
            />
          </AuthProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
