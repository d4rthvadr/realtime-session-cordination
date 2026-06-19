"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { continueAsGuest, sendOTP } from "@/lib/auth-actions";
import { Mail, ArrowRight } from "lucide-react";

export default function SignUpPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();
  const [isGuestPending, startGuestTransition] = useTransition();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!email.trim()) {
      setError("Please enter your email address");
      return;
    }

    if (!email.includes("@")) {
      setError("Please enter a valid email address");
      return;
    }

    startTransition(async () => {
      const result = await sendOTP(email, "signup");

      if (result.error) {
        setError(result.error);
        return;
      }

      const params = new URLSearchParams({
        email,
        type: "signup",
        ...(result.challengeId ? { challengeId: result.challengeId } : {}),
        ...(result.expiresInMinutes
          ? { expiresIn: String(result.expiresInMinutes) }
          : {}),
      });
      // Navigate to verification page with email in query
      router.push(`/verify?${params.toString()}`);
    });
  };

  const handleGuestMode = () => {
    setError(null);
    startGuestTransition(async () => {
      const result = await continueAsGuest();
      if (result.error) {
        setError(result.error);
        return;
      }

      router.push("/sessions");
    });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="space-y-2">
        <h1 className="text-3xl font-bold text-slate-900">Create an account</h1>
        <p className="text-slate-600">
          Enter your email to get started with SyncTime.
        </p>
      </div>

      {/* Form */}
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="email" className="text-slate-700">
            Email
          </Label>
          <div className="relative">
            <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
            <Input
              id="email"
              type="email"
              placeholder="Enter your email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="pl-10"
              disabled={isPending || isGuestPending}
            />
          </div>
        </div>

        {error && (
          <div className="text-sm text-red-600 bg-red-50 border border-red-200 rounded-md p-3">
            {error}
          </div>
        )}

        <Button
          type="submit"
          className="w-full rounded-full h-11"
          disabled={isPending || isGuestPending}
        >
          {isPending ? (
            "Sending OTP..."
          ) : (
            <>
              Continue with Email
              <ArrowRight className="w-4 h-4 ml-2" />
            </>
          )}
        </Button>

        <Button
          type="button"
          variant="outline"
          className="w-full rounded-full h-11"
          onClick={handleGuestMode}
          disabled={isPending || isGuestPending}
        >
          {isGuestPending ? "Entering guest mode..." : "Continue as Guest"}
        </Button>
      </form>

      {/* Divider */}
      <div className="relative">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-slate-200" />
        </div>
        <div className="relative flex justify-center text-xs uppercase">
          <span className="bg-white px-2 text-slate-500">or</span>
        </div>
      </div>

      {/* Sign In Link */}
      <div className="text-center text-sm text-slate-600">
        Already have an account?{" "}
        <Link
          href="/signin"
          className="font-semibold text-blue-600 hover:text-blue-700 transition-colors"
        >
          Sign in
        </Link>
      </div>

      {/* Footer */}
      <p className="text-xs text-slate-500 text-center mt-8">
        By continuing, you agree to our{" "}
        <Link href="/terms" className="underline hover:text-slate-700">
          Terms of Service
        </Link>{" "}
        and{" "}
        <Link href="/privacy" className="underline hover:text-slate-700">
          Privacy Policy
        </Link>
        .
      </p>
    </div>
  );
}
