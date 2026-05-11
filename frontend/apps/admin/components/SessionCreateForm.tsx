"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { parseDurationToSeconds } from "@/lib/session";
import { useAdminSessionStore } from "@/store/adminSessionStore";

export default function SessionCreateForm() {
  const router = useRouter();
  const createSession = useAdminSessionStore((state) => state.createSession);

  const [title, setTitle] = useState("Kubernetes Workshop");
  const [speakerName, setSpeakerName] = useState("John Doe");
  const [durationMinutes, setDurationMinutes] = useState("30");
  const [error, setError] = useState<string | null>(null);

  const onSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const durationSeconds = parseDurationToSeconds(durationMinutes);
    if (!title.trim() || !speakerName.trim() || durationSeconds <= 0) {
      setError("Please provide title, speaker, and a valid positive duration.");
      return;
    }

    const session = createSession({
      title: title.trim(),
      speakerName: speakerName.trim(),
      durationSeconds,
    });

    setError(null);
    router.push(`/sessions/${session.id}`);
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

      {error ? <p className="text-sm text-red-700">{error}</p> : null}

      <button
        type="submit"
        className="w-full rounded-md bg-slate-900 px-4 py-2 font-medium text-white transition hover:bg-slate-700"
      >
        Create Session
      </button>
    </form>
  );
}
