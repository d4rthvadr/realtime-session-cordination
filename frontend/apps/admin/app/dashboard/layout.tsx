import { cookies } from "next/headers";
import { jwtVerify } from "jose";
import DashboardNav from "@/components/DashboardNav";

const AUTH_COOKIE_NAME = "admin_auth_token";
const JWT_SECRET = process.env.JWT_SECRET
  ? new TextEncoder().encode(process.env.JWT_SECRET)
  : null;

async function getDashboardRole(): Promise<string | null> {
  const token = cookies().get(AUTH_COOKIE_NAME)?.value;
  if (!token || !JWT_SECRET) {
    return null;
  }

  try {
    const verified = await jwtVerify(token, JWT_SECRET);
    const payload = verified.payload as Record<string, unknown>;
    const role = payload.role;
    return typeof role === "string" ? role : null;
  } catch {
    return null;
  }
}

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const role = await getDashboardRole();

  return (
    <>
      <DashboardNav isAdmin={role === "admin"} />
      <main className="min-h-[calc(100vh-4rem)] bg-slate-50">{children}</main>
    </>
  );
}
