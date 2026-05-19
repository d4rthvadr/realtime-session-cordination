"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Bell, User } from "lucide-react";
import { cn } from "@/lib/utils";

export default function DashboardNav() {
  const pathname = usePathname();

  const navLinks = [
    { href: "/dashboard", label: "Dashboard" },
    { href: "/dashboard/sessions", label: "Sessions" },
  ];

  return (
    <header className="bg-white border-b border-slate-200 sticky top-0 z-50">
      <div className="max-w-[1600px] mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          {/* Logo and Nav Links */}
          <div className="flex items-center gap-8">
            <Link href="/dashboard" className="flex items-center">
              <span className="text-2xl font-bold text-slate-900">
                SyncTime
              </span>
            </Link>
            <nav className="hidden md:flex items-center gap-6">
              {navLinks.map((link) => {
                const isActive =
                  pathname === link.href ||
                  (link.href !== "/dashboard" &&
                    pathname?.startsWith(link.href));
                return (
                  <Link
                    key={link.href}
                    href={link.href}
                    className={cn(
                      "text-sm font-semibold transition-colors border-b-2 pb-1",
                      isActive
                        ? "text-blue-600 border-blue-600"
                        : "text-slate-600 border-transparent hover:text-slate-900",
                    )}
                  >
                    {link.label}
                  </Link>
                );
              })}
            </nav>
          </div>

          {/* Right Side: Notifications, Profile */}
          <div className="flex items-center gap-3">
            {/* Notifications */}
            <Button
              variant="ghost"
              size="icon"
              className="rounded-full hover:bg-slate-100"
            >
              <Bell className="w-5 h-5 text-slate-600" />
            </Button>

            {/* Profile */}
            <Button
              variant="ghost"
              size="icon"
              className="rounded-full hover:bg-slate-100"
            >
              <User className="w-5 h-5 text-slate-600" />
            </Button>
          </div>
        </div>
      </div>
    </header>
  );
}
