"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { Bell, User, BarChart3, Menu } from "lucide-react";
import { cn } from "@/lib/utils";

export default function DashboardNav() {
  const pathname = usePathname();
  const [isOpen, setIsOpen] = useState(false);

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
            <Link href="/dashboard" className="flex items-center gap-2">
              <div className="w-10 h-10 rounded-full bg-blue-600 flex items-center justify-center">
                <BarChart3 className="w-6 h-6 text-white" />
              </div>
              <span className="text-xl font-bold text-slate-900">SyncTime</span>
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

          {/* Right Side: Mobile Menu, Notifications, Profile */}
          <div className="flex items-center gap-3">
            {/* Mobile Menu Button */}
            <Sheet open={isOpen} onOpenChange={setIsOpen}>
              <SheetTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="rounded-full hover:bg-slate-100 md:hidden"
                >
                  <Menu className="w-5 h-5 text-slate-600" />
                </Button>
              </SheetTrigger>
              <SheetContent side="left" className="w-[280px] sm:w-[320px]">
                <SheetHeader className="mb-6">
                  <SheetTitle className="flex items-center gap-2">
                    <div className="w-10 h-10 rounded-full bg-blue-600 flex items-center justify-center">
                      <BarChart3 className="w-6 h-6 text-white" />
                    </div>
                    <span className="text-xl font-bold">SyncTime</span>
                  </SheetTitle>
                </SheetHeader>
                <nav className="flex flex-col gap-4">
                  {navLinks.map((link) => {
                    const isActive =
                      pathname === link.href ||
                      (link.href !== "/dashboard" &&
                        pathname?.startsWith(link.href));
                    return (
                      <Link
                        key={link.href}
                        href={link.href}
                        onClick={() => setIsOpen(false)}
                        className={cn(
                          "text-base font-semibold py-3 px-4 rounded-lg transition-colors",
                          isActive
                            ? "bg-blue-50 text-blue-600"
                            : "text-slate-700 hover:bg-slate-50",
                        )}
                      >
                        {link.label}
                      </Link>
                    );
                  })}
                </nav>
              </SheetContent>
            </Sheet>

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
