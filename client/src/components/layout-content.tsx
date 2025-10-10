"use client";

import { usePathname } from "next/navigation";
import { Sidebar } from "@/components/navigation/sidebar";
import { Topbar } from "@/components/navigation/topbar";
import { useAuth } from "@/lib/auth-context";

export function LayoutContent({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const { isAuthenticated, isLoading } = useAuth();
  
  // Check if we're on an auth page
  const isAuthPage = pathname?.startsWith('/auth');

  // Show content directly for auth pages
  if (isAuthPage) {
    return <>{children}</>;
  }

  // Show loading state
  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
      </div>
    );
  }

  // If not authenticated and not on auth page, show children (will redirect via protected route)
  if (!isAuthenticated) {
    return <>{children}</>;
  }

  // Show full layout for authenticated users
  return (
    <div className="grid min-h-screen md:grid-cols-[240px_1fr]">
      <div className="hidden md:flex">
        <Sidebar />
      </div>
      <div className="flex flex-col">
        <Topbar />
        <main className="flex-1 overflow-y-auto bg-background p-6">
          <div className="mx-auto w-full max-w-7xl space-y-6">{children}</div>
        </main>
      </div>
    </div>
  );
}