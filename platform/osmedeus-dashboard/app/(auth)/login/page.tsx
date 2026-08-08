"use client";

import * as React from "react";
import Image from "next/image";
import { useRouter } from "next/navigation";
import logo from "@/osmedeus-logo.png";
import { useAuth } from "@/providers/auth-provider";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ThemeToggle } from "@/components/layout/theme-toggle";
import { BookOpenIcon } from "lucide-react";
import { GithubIcon } from "@/components/icons/github-icon";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { LoaderIcon, AlertCircleIcon, HeartIcon } from "lucide-react";
import { toast } from "sonner";

export default function LoginPage() {
  const router = useRouter();
  const { login, isLoading: authLoading, isAuthenticated } = useAuth();
  const [username, setUsername] = React.useState("osmedeus");
  const [password, setPassword] = React.useState("");
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    if (!authLoading && isAuthenticated) {
      router.replace("/");
    }
  }, [authLoading, isAuthenticated, router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!username.trim()) {
      setError("Username is required");
      return;
    }

    if (!password.trim()) {
      setError("Password is required");
      return;
    }

    setIsSubmitting(true);

    try {
      await login(username, password);
      toast.success("Welcome back!", {
        description: "You have been signed in successfully.",
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setIsSubmitting(false);
    }
  };

  const isLoading = isSubmitting || authLoading;

  if (authLoading || isAuthenticated) {
    return (
      <div className="flex items-center justify-center py-12 text-muted-foreground">
        <LoaderIcon className="mr-2 size-4 animate-spin" />
        <span>Redirecting...</span>
      </div>
    );
  }

  return (
    <>
    <Card className="w-full shadow-lg">
      <CardHeader className="space-y-1 text-center">
        <div className="mx-auto mb-6 flex size-28 items-center justify-center rounded-2xl">
          <Image src={logo} alt="Osmedeus" priority className="h-28 w-auto logo-shadow" />
        </div>
        <CardTitle className="text-2xl">Welcome back</CardTitle>
        <CardDescription>
          Sign in to your Osmedeus Dashboard
        </CardDescription>
      </CardHeader>

      <form onSubmit={handleSubmit}>
        <CardContent className="space-y-4">
          {error && (
            <div className="flex items-center gap-2 rounded-control bg-destructive-soft p-3 text-sm text-destructive">
              <AlertCircleIcon className="size-4 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="username">Username</Label>
            <Input
              id="username"
              type="text"
              placeholder="osmedeus"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              disabled={isLoading}
              autoComplete="username"
              autoFocus
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="password">Password</Label>
            <Input
              id="password"
              type="password"
              placeholder="Enter your password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={isLoading}
              autoComplete="current-password"
            />
          </div>

        </CardContent>

        <CardFooter className="flex flex-col gap-4 mt-4">
          <p className="text-center text-xs text-muted-foreground">
            Your default password is in{" "}
            <strong className="font-semibold">$HOME/osmedeus-base/osm-settings.yaml</strong>.
          </p>
          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading ? (
              <>
                <LoaderIcon className="mr-2 size-4 animate-spin" />
                Signing in...
              </>
            ) : (
              "Sign in"
            )}
          </Button>
          <div className="flex items-center justify-center gap-3">
            <ThemeToggle variant="outline" size="sm" className="rounded-full px-3" ariaLabel="Toggle theme" label="Theme" />
            <Button variant="outline" size="sm" className="rounded-full px-3 gap-2" aria-label="Open documentation" asChild>
              <a href="https://docs.osmedeus.org/" target="_blank" rel="noopener noreferrer">
                <BookOpenIcon className="size-4" />
                <span className="text-xs">Docs</span>
              </a>
            </Button>
            <Button variant="outline" size="sm" className="rounded-full px-3 gap-2" aria-label="Open GitHub repository" asChild>
              <a href="https://github.com/j3ssie/osmedeus" target="_blank" rel="noopener noreferrer">
                <GithubIcon className="size-4" />
                <span className="text-xs">GitHub</span>
              </a>
            </Button>
          </div>
        </CardFooter>
      </form>
    </Card>
    <p className="mt-4 text-center text-xs text-muted-foreground">
      <code className="rounded-control bg-muted px-2 py-1 font-mono">
        Crafted with{" "}
        <HeartIcon className="size-3 text-destructive inline-block align-middle" aria-label="love" />{" "}
        by{" "}
        <a
          href="http://twitter.com/j3ssie"
          target="_blank"
          rel="noopener noreferrer"
          className="text-primary no-underline hover:underline"
        >
          @j3ssie
        </a>
      </code>
    </p>
    </>
  );
}
