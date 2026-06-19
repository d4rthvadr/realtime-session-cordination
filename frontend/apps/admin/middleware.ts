import { NextResponse, type NextRequest } from "next/server";
import { jwtVerify } from "jose";

if (!process.env.JWT_SECRET) {
  throw new Error("JWT_SECRET environment variable is required");
}
const AUTH_COOKIE_NAME = "admin_auth_token";
const JWT_SECRET = new TextEncoder().encode(process.env.JWT_SECRET);

function isProtectedPath(pathname: string): boolean {
  return pathname === "/" || pathname.startsWith("/dashboard");
}

const authPaths = ["/signin", "/signup", "/verify"];

function isAuthPath(pathname: string): boolean {
  return authPaths.includes(pathname);
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

  const auth = token ? await verifyAuth(token) : null;

  // If token exists but is invalid, clear it to prevent redirect loops.
  if (token && !auth) {
    if (isAuthPath(pathname)) {
      const response = NextResponse.next();
      response.cookies.delete(AUTH_COOKIE_NAME);
      return response;
    }

    const signInUrl = new URL("/signin", req.url);
    signInUrl.searchParams.set("next", pathname);
    const response = NextResponse.redirect(signInUrl);
    response.cookies.delete(AUTH_COOKIE_NAME);
    return response;
  }

  // Redirect authenticated users away from auth paths.
  // Both user and admin roles can use dashboard routes.
  if (isAuthPath(pathname) && auth) {
    return NextResponse.redirect(new URL("/dashboard", req.url));
  }

  // Gate /dashboard/ops routes to admin role only.
  if (pathname.startsWith("/dashboard/ops") && auth) {
    if (auth.role !== "admin") {
      return NextResponse.redirect(new URL("/dashboard", req.url));
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/", "/dashboard/:path*", "/signin", "/signup", "/verify"],
};
