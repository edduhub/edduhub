"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { GraduationCap, CheckCircle, XCircle, Loader2 } from "lucide-react";
import { endpoints } from "@/lib/api-client";
import { logger } from "@/lib/logger";

export default function VerifyEmailPage() {
  const searchParams = useSearchParams();
  const [status, setStatus] = useState<"loading" | "success" | "error">("loading");
  const [message, setMessage] = useState<string>("");

  const flow = searchParams.get("flow");
  const token = searchParams.get("token");

  useEffect(() => {
    if (!flow || !token) {
      setStatus("error");
      setMessage("Invalid verification link. The link may be incomplete or expired.");
      return;
    }

    const verify = async () => {
      try {
        const apiBase = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
        const url = `${apiBase}${endpoints.auth.verifyEmail(flow, token)}`;
        const resp = await fetch(url, {
          method: "GET",
          credentials: "include",
          headers: { Accept: "application/json" },
        });
        const data = await resp.json();

        const msg = (data as { data?: { message?: string }; message?: string })?.data?.message ?? (data as { message?: string })?.message ?? "";
        if (resp.ok && (data as { success?: boolean })?.success !== false) {
          setStatus("success");
          setMessage(msg || "Your email has been verified successfully.");
        } else {
          setStatus("error");
          setMessage(msg || (data as { error?: string })?.error || "Email verification failed. The link may have expired.");
        }
      } catch (err) {
        logger.error("Email verification failed", err as Error);
        setStatus("error");
        setMessage("Verification failed. Please try again or request a new verification email.");
      }
    };

    verify();
  }, [flow, token]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-background via-background to-muted p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-3 text-center">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
            <GraduationCap className="h-6 w-6 text-primary" />
          </div>
          <CardTitle className="text-2xl">Email Verification</CardTitle>
          <CardDescription>
            {status === "loading"
              ? "Verifying your email address..."
              : status === "success"
                ? "Verification complete"
                : "Verification failed"}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {status === "loading" && (
            <div className="flex justify-center py-8">
              <Loader2 className="h-12 w-12 animate-spin text-primary" />
            </div>
          )}

          {status === "success" && (
            <>
              <div className="flex justify-center">
                <div className="rounded-full bg-green-100 p-4 dark:bg-green-900/30">
                  <CheckCircle className="h-12 w-12 text-green-600 dark:text-green-400" />
                </div>
              </div>
              <p className="text-center text-muted-foreground">{message}</p>
              <Button asChild className="w-full">
                <Link href="/auth/login">Sign in</Link>
              </Button>
            </>
          )}

          {status === "error" && (
            <>
              <div className="flex justify-center">
                <div className="rounded-full bg-destructive/10 p-4">
                  <XCircle className="h-12 w-12 text-destructive" />
                </div>
              </div>
              <p className="text-center text-muted-foreground">{message}</p>
              <div className="flex flex-col gap-2">
                <Button asChild className="w-full">
                  <Link href="/auth/login">Back to Sign in</Link>
                </Button>
                <Button asChild variant="outline" className="w-full">
                  <Link href="/auth/register">Create Account</Link>
                </Button>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
