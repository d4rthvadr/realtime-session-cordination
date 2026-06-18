"use server";

import { cookies } from "next/headers";
import { redirect } from "next/navigation";

const AUTH_BACKEND_URL =
  process.env.NEXT_PUBLIC_AUTH_BACKEND_URL || "http://localhost:8080";

export interface AuthResponse {
  success: boolean;
  error: string | null;
  message?: string;
}

export interface VerifyOTPResponse extends AuthResponse {
  token?: string;
  userId?: string;
}

const AUTH_COOKIE_NAME = "admin_auth_token";
const AUTH_COOKIE_MAX_AGE_SECONDS = 60 * 60 * 24 * 7;

interface GuestAuthPayload {
  token?: string;
  user?: {
    id?: string;
  };
  error?: string;
}

function setAdminAuthCookie(token: string) {
  const cookieStore = cookies();
  cookieStore.set({
    name: AUTH_COOKIE_NAME,
    value: token,
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/",
    maxAge: AUTH_COOKIE_MAX_AGE_SECONDS,
  });
}

function clearAdminAuthCookie() {
  const cookieStore = cookies();
  cookieStore.set({
    name: AUTH_COOKIE_NAME,
    value: "",
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/",
    maxAge: 0,
  });
}

async function requestGuestToken(): Promise<{
  token: string;
  userId?: string;
} | null> {
  const response = await fetch(`${AUTH_BACKEND_URL}/api/v1/auth/guest`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    cache: "no-store",
  });

  if (!response.ok) {
    return null;
  }

  const data = (await response.json()) as GuestAuthPayload;
  if (!data.token) {
    return null;
  }

  return {
    token: data.token,
    userId: data.user?.id,
  };
}

export async function signOutAdmin(): Promise<never> {
  clearAdminAuthCookie();
  redirect("/signin");
}

// Send OTP to email
export async function sendOTP(email: string): Promise<AuthResponse> {
  try {
    // TODO: Implement actual backend call
    // For now, simulate API call
    await new Promise((resolve) => setTimeout(resolve, 1000));

    // Mock validation
    if (!email || !email.includes("@")) {
      return {
        success: false,
        error: "Please provide a valid email address",
      };
    }

    // Simulate successful OTP send
    console.log(`Sending OTP to ${email}`);

    return {
      success: true,
      error: null,
      message: `OTP sent to ${email}`,
    };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to send OTP";
    return {
      success: false,
      error: message,
    };
  }
}

// Verify OTP code
export async function verifyOTP(
  email: string,
  code: string,
): Promise<VerifyOTPResponse> {
  try {
    // TODO: Implement actual backend call
    await new Promise((resolve) => setTimeout(resolve, 1000));

    // Mock validation
    if (!email || !code) {
      return {
        success: false,
        error: "Email and code are required",
      };
    }

    if (code.length !== 6) {
      return {
        success: false,
        error: "OTP code must be 6 digits",
      };
    }

    // Mock OTP verification, then mint a real backend JWT for current flows.
    console.log(`Verifying OTP for ${email}: ${code}`);

    const guestAuth = await requestGuestToken();
    if (!guestAuth) {
      return {
        success: false,
        error: "Failed to create authenticated session",
      };
    }

    setAdminAuthCookie(guestAuth.token);

    return {
      success: true,
      error: null,
      message: "OTP verified successfully",
      token: guestAuth.token,
      userId: guestAuth.userId,
    };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to verify OTP";
    return {
      success: false,
      error: message,
    };
  }
}

export async function continueAsGuest(): Promise<AuthResponse> {
  try {
    const guestAuth = await requestGuestToken();
    if (!guestAuth) {
      return {
        success: false,
        error: "Failed to create guest session",
      };
    }

    setAdminAuthCookie(guestAuth.token);

    return {
      success: true,
      error: null,
      message: "Signed in as guest",
    };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to continue as guest";
    return {
      success: false,
      error: message,
    };
  }
}

// Sign up with email
export async function signUp(email: string): Promise<AuthResponse> {
  try {
    // TODO: Implement actual backend call
    await new Promise((resolve) => setTimeout(resolve, 500));

    if (!email || !email.includes("@")) {
      return {
        success: false,
        error: "Please provide a valid email address",
      };
    }

    console.log(`Creating account for ${email}`);

    return {
      success: true,
      error: null,
      message: "Account created successfully",
    };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to create account";
    return {
      success: false,
      error: message,
    };
  }
}

// Sign in with email
export async function signIn(email: string): Promise<AuthResponse> {
  try {
    // TODO: Implement actual backend call
    await new Promise((resolve) => setTimeout(resolve, 500));

    if (!email || !email.includes("@")) {
      return {
        success: false,
        error: "Please provide a valid email address",
      };
    }

    console.log(`Signing in ${email}`);

    return {
      success: true,
      error: null,
      message: "Signed in successfully",
    };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to sign in";
    return {
      success: false,
      error: message,
    };
  }
}
