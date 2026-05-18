import Image from "next/image";

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
  isDark = true,
}: EmptyStateProps) {
  return (
    <div
      className={`flex flex-col items-center justify-center rounded-2xl border p-12 shadow-sm ${
        isDark
          ? "border-slate-800 bg-slate-900/70 backdrop-blur"
          : "border-slate-200 bg-white"
      }`}
    >
      <div className="relative h-48 w-64">
        <Image
          src="/images/session-not-found-empty-state.svg"
          alt="Empty state"
          fill
          className="object-contain"
          priority
        />
      </div>

      <h3
        className={`mt-6 text-xl font-semibold ${isDark ? "text-slate-100" : "text-slate-900"}`}
      >
        {title}
      </h3>
      <p
        className={`mt-2 text-center ${isDark ? "text-slate-300" : "text-slate-600"}`}
      >
        {description}
      </p>

      {actionText && onAction && (
        <button
          onClick={onAction}
          className={`mt-6 rounded-md px-4 py-2 text-sm font-medium transition ${
            isDark
              ? "border border-slate-700 bg-slate-800 text-slate-100 hover:bg-slate-700"
              : "bg-slate-900 text-white hover:bg-slate-700"
          }`}
        >
          {actionText}
        </button>
      )}
    </div>
  );
}
