import { NextResponse, type NextRequest } from "next/server";
import { jwtVerify } from "jose";

const AUTH_COOKIE_NAME = "admin_auth_token";
const JWT_SECRET = new TextEncoder().encode(
  process.env.JWT_SECRET || "dev-secret-key",
);

function isProtectedPath(pathname: string): boolean {
  return (
    pathname === "/" ||
    pathname.startsWith("/dashboard") ||
    pathname.startsWith("/sessions")
  );
}

function isAuthPath(pathname: string): boolean {
  return (
    pathname === "/signin" || pathname === "/signup" || pathname === "/verify"
  );
}

async function verifyAuth(token: string): Promise<{ role: string } | null> {
  try {
    const verified = await jwtVerify(token, JWT_SECRET);
    const payload = verified.payload as Record<string, unknown>;
    const role = payload.role as string | undefined;
    return role ? { role } : null;
  } catch (err) {
    console.error("Token verification failed:", err);
    return null;
  }
}

export async function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;
  const token = req.cookies.get(AUTH_COOKIE_NAME)?.value;

  // Check auth for protected paths
  if (isProtectedPath(pathname) && !token) {
    const signInUrl = new URL("/signin", req.url);
    signInUrl.searchParams.set("next", pathname);
    return NextResponse.redirect(signInUrl);
  }

  // Redirect authenticated users away from auth paths
  if (isAuthPath(pathname) && token) {
    return NextResponse.redirect(new URL("/dashboard", req.url));
  }

  // Gate /dashboard routes to admin role only
  if (pathname.startsWith("/dashboard") && token) {
    const auth = await verifyAuth(token);
    if (!auth || auth.role !== "admin") {
      // Redirect non-admin users to home or signin
      return NextResponse.redirect(new URL("/signin", req.url));
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/",
    "/dashboard/:path*",
    "/sessions/:path*",
    "/signin",
    "/signup",
    "/verify",
  ],
};
