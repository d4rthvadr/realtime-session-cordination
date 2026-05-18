import Image from "next/image";

interface EmptyStateProps {
  title: string;
  description: string;
  actionText?: string;
  onAction?: () => void;
}

export default function EmptyState({
  title,
  description,
  actionText,
  onAction,
}: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center rounded-2xl border border-slate-200 bg-white p-12 shadow-sm">
      <div className="relative h-48 w-64">
        <Image
          src="/images/session-not-found-empty-state.svg"
          alt="Empty state"
          fill
          className="object-contain"
          priority
        />
      </div>

      <h3 className="mt-6 text-xl font-semibold text-slate-900">{title}</h3>
      <p className="mt-2 text-center text-slate-600">{description}</p>

      {actionText && onAction && (
        <button
          onClick={onAction}
          className="mt-6 rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-slate-700"
        >
          {actionText}
        </button>
      )}
    </div>
  );
}
