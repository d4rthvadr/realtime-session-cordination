"use client";

import { useState, useTransition, useEffect, useRef } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { verifyOTP, sendOTP } from "@/lib/auth-actions";
import { ArrowLeft, Mail } from "lucide-react";

export default function VerifyOTPPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const email = searchParams.get("email") || "";
  const type = searchParams.get("type") || "signin";

  const [otp, setOtp] = useState(["", "", "", "", "", ""]);
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();
  const [isResending, setIsResending] = useState(false);
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  useEffect(() => {
    // Redirect if no email provided
    if (!email) {
      router.push(`/${type}`);
    }
  }, [email, type, router]);

  useEffect(() => {
    // Focus first input on mount
    inputRefs.current[0]?.focus();
  }, []);

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
      const result = await verifyOTP(email, code);

      if (result.error) {
        setError(result.error);
        // Clear OTP inputs on error
        setOtp(["", "", "", "", "", ""]);
        inputRefs.current[0]?.focus();
        return;
      }

      // Store auth token (in a real app, use httpOnly cookies or secure storage)
      if (result.token) {
        window.localStorage.setItem("authToken", result.token);
      }

      // Redirect to dashboard
      router.push("/dashboard");
    });
  };

  const handleResend = async () => {
    setIsResending(true);
    setError(null);

    const result = await sendOTP(email);

    if (result.error) {
      setError(result.error);
    }

    setTimeout(() => setIsResending(false), 1000);
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
          disabled={isResending}
          className="text-sm text-blue-600 hover:text-blue-700 font-medium transition-colors disabled:opacity-50"
        >
          {isResending ? "Sending..." : "Resend code"}
        </button>
      </div>

      {/* Info */}
      <div className="bg-blue-50 border border-blue-200 rounded-md p-4 flex gap-3">
        <Mail className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
        <div className="text-sm text-blue-900">
          <p className="font-medium mb-1">Didn&apos;t receive the code?</p>
          <p className="text-blue-700">
            Check your spam folder or try resending the code. The code expires
            in 10 minutes.
          </p>
        </div>
      </div>
    </div>
  );
}
