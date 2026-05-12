"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { parseDurationToSeconds } from "@/lib/session";
import { useCreateSession } from "@/hooks/useCreateSession";

export default function SessionCreateForm() {
  const router = useRouter();

  const [title, setTitle] = useState("Kubernetes Workshop");
  const [speakerName, setSpeakerName] = useState("John Doe");
  const [durationMinutes, setDurationMinutes] = useState("30");
  const [validationError, setValidationError] = useState<string | null>(null);
  const { createSession, isSubmitting, error, clearError } = useCreateSession();

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setValidationError(null);
    clearError();

    const durationSeconds = parseDurationToSeconds(durationMinutes);
    if (!title.trim() || !speakerName.trim() || durationSeconds <= 0) {
      setValidationError(
        "Please provide title, speaker, and a valid positive duration.",
      );
      return;
    }

    const payload = await createSession({
      title: title.trim(),
      speakerName: speakerName.trim(),
      durationSeconds,
    });

    if (payload) {
      window.sessionStorage.setItem(
        `controlToken:${payload.session.id}`,
        payload.controlToken,
      );

      router.push(`/sessions/${payload.session.id}`);
    }
  };

  return (
    <form
      onSubmit={onSubmit}
      className="space-y-5 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm"
    >
      <div>
        <h2 className="text-xl font-semibold text-slate-900">Create Session</h2>
        <p className="text-sm text-slate-600">
          Define title, speaker, and session duration.
        </p>
      </div>

      <label className="block">
        <span className="mb-1 block text-sm font-medium text-slate-700">
          Title
        </span>
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          className="w-full rounded-md border border-slate-300 px-3 py-2 text-slate-900 outline-none ring-0 focus:border-slate-500"
          placeholder="Kubernetes Workshop"
        />
      </label>

      <label className="block">
        <span className="mb-1 block text-sm font-medium text-slate-700">
          Speaker
        </span>
        <input
          value={speakerName}
          onChange={(e) => setSpeakerName(e.target.value)}
          className="w-full rounded-md border border-slate-300 px-3 py-2 text-slate-900 outline-none ring-0 focus:border-slate-500"
          placeholder="John Doe"
        />
      </label>

      <label className="block">
        <span className="mb-1 block text-sm font-medium text-slate-700">
          Duration (minutes)
        </span>
        <input
          type="number"
          min={1}
          step={1}
          value={durationMinutes}
          onChange={(e) => setDurationMinutes(e.target.value)}
          className="w-full rounded-md border border-slate-300 px-3 py-2 text-slate-900 outline-none ring-0 focus:border-slate-500"
        />
      </label>

      {validationError ? (
        <p className="text-sm text-red-700">{validationError}</p>
      ) : null}
      {error ? <p className="text-sm text-red-700">{error}</p> : null}

      <button
        type="submit"
        disabled={isSubmitting}
        className="w-full rounded-md bg-slate-900 px-4 py-2 font-medium text-white transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:bg-slate-400"
      >
        {isSubmitting ? "Creating..." : "Create Session"}
      </button>
    </form>
  );
}
