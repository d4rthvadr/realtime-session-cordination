"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { parseDurationToSeconds } from "@/lib/session";
import { createSession } from "@/lib/actions";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";

interface SessionCreateModalProps {
  trigger?: React.ReactNode;
  onSuccess?: () => void;
}

export default function SessionCreateModal({
  trigger,
  onSuccess,
}: SessionCreateModalProps) {
  const router = useRouter();
  const [isPending, startTransition] = useTransition();
  const [open, setOpen] = useState(false);

  const [title, setTitle] = useState("Kubernetes Workshop");
  const [speakerName, setSpeakerName] = useState("John Doe");
  const [durationMinutes, setDurationMinutes] = useState("30");
  const [validationError, setValidationError] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const onSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setValidationError(null);
    setError(null);

    const durationSeconds = parseDurationToSeconds(durationMinutes);
    if (!title.trim() || !speakerName.trim() || durationSeconds <= 0) {
      setValidationError(
        "Please provide title, speaker, and a valid positive duration.",
      );
      return;
    }

    startTransition(async () => {
      const result = await createSession({
        name: title.trim(),
        duration: durationSeconds,
      });

      if (result.error) {
        setError(result.error);
        return;
      }

      if (result.session) {
        // Store control token if backend returns it
        if (
          "controlToken" in result.session &&
          typeof (result.session as any).controlToken === "string"
        ) {
          window.sessionStorage.setItem(
            `controlToken:${result.session.id}`,
            (result.session as any).controlToken,
          );
        }

        // Close modal and reset form
        setOpen(false);
        setTitle("Kubernetes Workshop");
        setSpeakerName("John Doe");
        setDurationMinutes("30");
        setValidationError(null);
        setError(null);

        // Call onSuccess callback if provided
        if (onSuccess) {
          onSuccess();
        }

        // Navigate to session detail page
        router.push(`/dashboard/sessions/${result.session.id}`);
      }
    });
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {trigger || (
          <Button className="rounded-full">
            <Plus className="w-4 h-4 mr-2" />
            Initialize Session
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Create Session</DialogTitle>
          <DialogDescription>
            Define title, speaker, and session duration.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="space-y-4 mt-4">
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium text-slate-700">
              Title
            </span>
            <input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full rounded-md border border-slate-300 px-3 py-2 text-slate-900 outline-none ring-0 focus:border-slate-500 focus:ring-2 focus:ring-slate-200"
              placeholder="Kubernetes Workshop"
            />
          </label>

          <label className="block">
            <span className="mb-1.5 block text-sm font-medium text-slate-700">
              Speaker
            </span>
            <input
              value={speakerName}
              onChange={(e) => setSpeakerName(e.target.value)}
              className="w-full rounded-md border border-slate-300 px-3 py-2 text-slate-900 outline-none ring-0 focus:border-slate-500 focus:ring-2 focus:ring-slate-200"
              placeholder="John Doe"
            />
          </label>

          <label className="block">
            <span className="mb-1.5 block text-sm font-medium text-slate-700">
              Duration (minutes)
            </span>
            <input
              type="number"
              min={1}
              step={1}
              value={durationMinutes}
              onChange={(e) => setDurationMinutes(e.target.value)}
              className="w-full rounded-md border border-slate-300 px-3 py-2 text-slate-900 outline-none ring-0 focus:border-slate-500 focus:ring-2 focus:ring-slate-200"
            />
          </label>

          {validationError ? (
            <p className="text-sm text-red-600">{validationError}</p>
          ) : null}

          {error ? <p className="text-sm text-red-600">{error}</p> : null}

          <div className="flex justify-end gap-3 pt-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => setOpen(false)}
              disabled={isPending}
              className="rounded-full"
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isPending} className="rounded-full">
              {isPending ? "Creating..." : "Create Session"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
