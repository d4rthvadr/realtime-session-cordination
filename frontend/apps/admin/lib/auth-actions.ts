"use server";

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

    // Mock verification (accept any 6-digit code for demo)
    console.log(`Verifying OTP for ${email}: ${code}`);

    return {
      success: true,
      error: null,
      message: "OTP verified successfully",
      token: `mock-token-${Date.now()}`,
      userId: `user-${Date.now()}`,
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
