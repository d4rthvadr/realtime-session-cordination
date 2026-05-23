import { NextResponse, type NextRequest } from "next/server";

const AUTH_COOKIE_NAME = "admin_auth_token";

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

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;
  const token = req.cookies.get(AUTH_COOKIE_NAME)?.value;

  if (isProtectedPath(pathname) && !token) {
    const signInUrl = new URL("/signin", req.url);
    signInUrl.searchParams.set("next", pathname);
    return NextResponse.redirect(signInUrl);
  }

  if (isAuthPath(pathname) && token) {
    return NextResponse.redirect(new URL("/dashboard", req.url));
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
