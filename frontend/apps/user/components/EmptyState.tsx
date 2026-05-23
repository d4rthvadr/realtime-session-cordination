import Image from "next/image";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

interface EmptyStateProps {
  title: string;
  description: string;
  actionText?: string;
  onAction?: () => void;
  isDark?: boolean;
}

export default function EmptyState({
  title,
  description,
  actionText,
  onAction,
  isDark = false,
}: EmptyStateProps) {
  return (
    <Card
      className={
        isDark
          ? "border-slate-800 bg-slate-900/70 text-slate-100"
          : "border-slate-200 bg-white"
      }
    >
      <CardContent className="flex flex-col items-center justify-center p-8 text-center sm:p-12">
        <div className="relative h-40 w-56 sm:h-48 sm:w-64">
          <Image
            src="/images/session-not-found-empty-state.svg"
            alt="Empty state"
            fill
            className="object-contain"
            priority
          />
        </div>

        <h3
          className={`mt-6 text-xl font-semibold ${
            isDark ? "text-slate-100" : "text-slate-900"
          }`}
        >
          {title}
        </h3>
        <p
          className={`mt-2 max-w-2xl ${
            isDark ? "text-slate-400" : "text-slate-600"
          }`}
        >
          {description}
        </p>

        {actionText && onAction && (
          <Button
            onClick={onAction}
            className={
              isDark
                ? "mt-6 bg-slate-100 text-slate-900 hover:bg-white"
                : "mt-6"
            }
          >
            {actionText}
          </Button>
        )}
      </CardContent>
    </Card>
  );
}
