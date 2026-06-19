"use client";

import { Suspense, useState, useTransition, useEffect, useRef } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { verifyOTP, sendOTP } from "@/lib/auth-actions";
import { ArrowLeft, Mail } from "lucide-react";

function VerifyOTPContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const email = searchParams.get("email") || "";
  const type = searchParams.get("type") || "signin";
  const initialChallengeId = searchParams.get("challengeId") || "";
  const initialExpiresIn = parseInt(searchParams.get("expiresIn") || "10", 10);

  const [otp, setOtp] = useState(["", "", "", "", "", ""]);
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();
  const [challengeId, setChallengeId] = useState(initialChallengeId);
  const [resendCooldown, setResendCooldown] = useState(0);
  const [isResending, setIsResending] = useState(false);
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  useEffect(() => {
    if (!email) {
      router.push(`/${type}`);
    }
  }, [email, type, router]);

  useEffect(() => {
    inputRefs.current[0]?.focus();
  }, []);

  // Resend cooldown countdown
  useEffect(() => {
    if (resendCooldown <= 0) return;
    const id = setTimeout(() => setResendCooldown((s) => s - 1), 1000);
    return () => clearTimeout(id);
  }, [resendCooldown]);

  const handleChange = (index: number, value: string) => {
    // Only allow numbers
    if (value && !/^\d$/.test(value)) {
      return;
    }

    const newOtp = [...otp];
    newOtp[index] = value;
    setOtp(newOtp);

    // Auto-focus next input
    if (value && index < 5) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handleKeyDown = (index: number, e: React.KeyboardEvent) => {
    // Handle backspace
    if (e.key === "Backspace" && !otp[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  };

  const handlePaste = (e: React.ClipboardEvent) => {
    e.preventDefault();
    const pastedData = e.clipboardData.getData("text").trim();

    if (!/^\d{6}$/.test(pastedData)) {
      return;
    }

    const newOtp = pastedData.split("");
    setOtp(newOtp);
    inputRefs.current[5]?.focus();
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const code = otp.join("");

    if (code.length !== 6) {
      setError("Please enter all 6 digits");
      return;
    }

    startTransition(async () => {
      const result = await verifyOTP(email, code, type, challengeId);

      if (result.error) {
        setError(result.error);
        // Clear OTP inputs on error
        setOtp(["", "", "", "", "", ""]);
        inputRefs.current[0]?.focus();
        return;
      }

      // Redirect by role: admins go to dashboard, everyone else to sessions
      if (result.role === "admin") {
        router.push("/dashboard");
      } else {
        router.push("/sessions");
      }
    });
  };

  const handleResend = async () => {
    if (resendCooldown > 0) return;
    setIsResending(true);
    setError(null);

    const result = await sendOTP(email, type);

    if (result.error) {
      setError(result.error);
    } else {
      if (result.challengeId) setChallengeId(result.challengeId);
      setResendCooldown(30);
    }

    setIsResending(false);
  };

  return (
    <div className="space-y-6">
      {/* Back Button */}
      <Link
        href={`/${type}`}
        className="inline-flex items-center text-sm text-slate-600 hover:text-slate-900 transition-colors"
      >
        <ArrowLeft className="w-4 h-4 mr-2" />
        Back
      </Link>

      {/* Header */}
      <div className="space-y-2">
        <h1 className="text-3xl font-bold text-slate-900">Check your email</h1>
        <p className="text-slate-600">
          We sent a verification code to{" "}
          <span className="font-medium text-slate-900">{email}</span>
        </p>
      </div>

      {/* Form */}
      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="space-y-2">
          <Label className="text-slate-700">Verification Code</Label>
          <div className="flex gap-2 justify-between">
            {otp.map((digit, index) => (
              <Input
                key={index}
                ref={(el) => {
                  inputRefs.current[index] = el;
                }}
                type="text"
                inputMode="numeric"
                maxLength={1}
                value={digit}
                onChange={(e) => handleChange(index, e.target.value)}
                onKeyDown={(e) => handleKeyDown(index, e)}
                onPaste={index === 0 ? handlePaste : undefined}
                className="w-full h-14 text-center text-2xl font-semibold"
                disabled={isPending}
              />
            ))}
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
          disabled={isPending || otp.join("").length !== 6}
        >
          {isPending ? "Verifying..." : "Verify and Continue"}
        </Button>
      </form>

      {/* Resend */}
      <div className="text-center">
        <button
          type="button"
          onClick={handleResend}
          disabled={isResending || resendCooldown > 0}
          className="text-sm text-blue-600 hover:text-blue-700 font-medium transition-colors disabled:opacity-50"
        >
          {isResending
            ? "Sending..."
            : resendCooldown > 0
              ? `Resend in ${resendCooldown}s`
              : "Resend code"}
        </button>
      </div>

      {/* Info */}
      <div className="bg-blue-50 border border-blue-200 rounded-md p-4 flex gap-3">
        <Mail className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
        <div className="text-sm text-blue-900">
          <p className="font-medium mb-1">Didn&apos;t receive the code?</p>
          <p className="text-blue-700">
            Check your spam folder or try resending the code. The code expires
            in {initialExpiresIn} minute{initialExpiresIn !== 1 ? "s" : ""}.
          </p>
        </div>
      </div>
    </div>
  );
}

export default function VerifyOTPPage() {
  return (
    <Suspense
      fallback={<div className="text-sm text-slate-500">Loading...</div>}
    >
      <VerifyOTPContent />
    </Suspense>
  );
}
