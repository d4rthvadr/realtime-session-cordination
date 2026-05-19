"use client";

import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

type StatusCardVariant = "success" | "warning" | "info" | "error";

interface StatusCardProps {
  icon: React.ReactNode;
  label: string;
  value: string;
  subtitle: string;
  variant?: StatusCardVariant;
}

export default function StatusCard({
  icon,
  label,
  value,
  subtitle,
  variant = "info",
}: StatusCardProps) {
  const variantStyles = {
    success: {
      bg: "bg-green-50",
      text: "text-green-700",
      valueText: "text-green-900",
    },
    warning: {
      bg: "bg-orange-50",
      text: "text-orange-700",
      valueText: "text-orange-900",
    },
    info: {
      bg: "bg-blue-50",
      text: "text-blue-700",
      valueText: "text-blue-900",
    },
    error: {
      bg: "bg-red-50",
      text: "text-red-700",
      valueText: "text-red-900",
    },
  };

  const styles = variantStyles[variant];

  return (
    <Card className={cn("col-span-3", styles.bg)}>
      <CardContent className="p-6">
        <div className={cn("flex items-center gap-2 mb-4", styles.text)}>
          {icon}
          <span className="text-xs font-semibold uppercase tracking-wider">
            {label}
          </span>
        </div>
        <div className={cn("text-2xl font-semibold mb-1", styles.valueText)}>
          {value}
        </div>
        <div className={cn("text-sm", styles.text)}>{subtitle}</div>
      </CardContent>
    </Card>
  );
}

// Icon components for convenience
export function SignalIcon({ className = "w-5 h-5" }: { className?: string }) {
  return (
    <svg
      className={className}
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0"
      />
    </svg>
  );
}

export function CPUIcon({ className = "w-5 h-5" }: { className?: string }) {
  return (
    <svg
      className={className}
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"
      />
    </svg>
  );
}
