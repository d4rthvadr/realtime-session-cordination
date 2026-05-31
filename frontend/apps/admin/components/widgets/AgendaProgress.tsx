"use client";

import { useMemo, useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  ClipboardList,
  ArrowUp,
  ArrowDown,
  XCircle,
  Plus,
  Play,
  Pause,
  RotateCcw,
  Square,
  Minus,
} from "lucide-react";
import type {
  ProgramItemCreateInput,
  ProgramItemSnapshot,
} from "@/lib/actions";

const PROGRAM_ITEM_TYPES = [
  "announcement",
  "break",
  "keynote",
  "lecture",
  "panel",
  "q&a",
] as const;

interface AgendaProgressProps {
  items: ProgramItemSnapshot[];
  isPending: boolean;
  error: string | null;
  onCreateAction: (input: ProgramItemCreateInput) => void;
  onCancelAction: (itemId: string) => void;
  onStartAction: (itemId: string) => void;
  onPauseAction: (itemId: string) => void;
  onResumeAction: (itemId: string) => void;
  onEndAction: (itemId: string) => void;
  onAdjustTimeAction: (itemId: string, deltaSeconds: number) => void;
  onReorderAction: (items: Array<{ id: string; position: number }>) => void;
  runtimeEnabled: boolean;
}

export default function AgendaProgress({
  items,
  isPending,
  error,
  onCreateAction,
  onCancelAction,
  onStartAction,
  onPauseAction,
  onResumeAction,
  onEndAction,
  onAdjustTimeAction,
  onReorderAction,
  runtimeEnabled,
}: AgendaProgressProps) {
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [type, setType] = useState("lecture");
  const [hostName, setHostName] = useState("");
  const [location, setLocation] = useState("");
  const [startAt, setStartAt] = useState("");
  const [endAt, setEndAt] = useState("");
  const [validationError, setValidationError] = useState<string | null>(null);

  const sortedItems = useMemo(
    () => [...items].sort((a, b) => a.position - b.position),
    [items],
  );

  const resetForm = () => {
    setIsCreateModalOpen(false);
    setTitle("");
    setType("lecture");
    setHostName("");
    setLocation("");
    setStartAt("");
    setEndAt("");
    setValidationError(null);
  };

  const handleCreate = () => {
    setValidationError(null);
    if (!title || !startAt || !endAt) {
      setValidationError("Please complete the title and schedule fields.");
      return;
    }

    if (
      !PROGRAM_ITEM_TYPES.includes(type as (typeof PROGRAM_ITEM_TYPES)[number])
    ) {
      setValidationError("Please choose a valid program item type.");
      return;
    }

    if (new Date(startAt).getTime() >= new Date(endAt).getTime()) {
      setValidationError("End time must be after the start time.");
      return;
    }

    const expectedDurationMinutes = Math.max(
      1,
      Math.round(
        (new Date(endAt).getTime() - new Date(startAt).getTime()) / 60000,
      ),
    );

    onCreateAction({
      title,
      type,
      hostName: hostName || undefined,
      location: location || undefined,
      scheduledStart: new Date(startAt).toISOString(),
      scheduledEnd: new Date(endAt).toISOString(),
      expectedDurationMinutes,
      position: sortedItems.length + 1,
    });
    setIsCreateModalOpen(false);
    resetForm();
  };

  const moveItem = (itemId: string, direction: -1 | 1) => {
    const index = sortedItems.findIndex((item) => item.id === itemId);
    const target = index + direction;
    if (index < 0 || target < 0 || target >= sortedItems.length) {
      return;
    }

    const next = [...sortedItems];
    const tmp = next[index];
    next[index] = next[target];
    next[target] = tmp;

    onReorderAction(
      next.map((item, idx) => ({ id: item.id, position: idx + 1 })),
    );
  };

  return (
    <Card className="col-span-12 md:col-span-6 bg-purple-50">
      <CardContent className="p-4 sm:p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2 text-purple-700">
            <ClipboardList className="w-5 h-5" />
            <span className="text-xs font-semibold uppercase tracking-wider">
              Program Timeline
            </span>
          </div>
          <div className="flex items-center gap-2">
            <Badge
              variant="outline"
              className="text-purple-700 text-xs tracking-wider"
            >
              {sortedItems.length} ITEM{sortedItems.length === 1 ? "" : "S"}
            </Badge>
            <Dialog
              open={isCreateModalOpen}
              onOpenChange={setIsCreateModalOpen}
            >
              <DialogTrigger asChild>
                <Button
                  type="button"
                  size="sm"
                  className="h-8 rounded-full bg-purple-700 text-xs hover:bg-purple-800"
                  disabled={isPending}
                >
                  <Plus className="mr-1 h-3 w-3" />
                  Add Item
                </Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-[560px]">
                <DialogHeader>
                  <DialogTitle>Add Program Item</DialogTitle>
                  <DialogDescription>
                    Create a scheduled timeline item for this session.
                  </DialogDescription>
                </DialogHeader>

                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <div className="space-y-2 sm:col-span-2">
                    <Label htmlFor="program-item-title">
                      Program item title
                    </Label>
                    <Input
                      id="program-item-title"
                      value={title}
                      onChange={(e) => setTitle(e.target.value)}
                      placeholder="Program item title"
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="program-item-type">Type</Label>
                    <select
                      id="program-item-type"
                      value={type}
                      onChange={(e) => setType(e.target.value)}
                      className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                    >
                      {PROGRAM_ITEM_TYPES.map((itemType) => (
                        <option key={itemType} value={itemType}>
                          {itemType.toUpperCase()}
                        </option>
                      ))}
                    </select>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="program-item-host">Host name</Label>
                    <Input
                      id="program-item-host"
                      value={hostName}
                      onChange={(e) => setHostName(e.target.value)}
                      placeholder="Host name (optional)"
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="program-item-location">Location</Label>
                    <Input
                      id="program-item-location"
                      value={location}
                      onChange={(e) => setLocation(e.target.value)}
                      placeholder="Location (optional)"
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="program-item-start">Start time</Label>
                    <Input
                      id="program-item-start"
                      type="datetime-local"
                      value={startAt}
                      onChange={(e) => setStartAt(e.target.value)}
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="program-item-end">End time</Label>
                    <Input
                      id="program-item-end"
                      type="datetime-local"
                      value={endAt}
                      onChange={(e) => setEndAt(e.target.value)}
                    />
                  </div>
                </div>

                {validationError ? (
                  <div className="rounded border border-red-300 bg-red-50 p-2 text-sm text-red-700">
                    {validationError}
                  </div>
                ) : null}

                <DialogFooter>
                  <Button
                    type="button"
                    variant="outline"
                    className="rounded-full"
                    disabled={isPending}
                    onClick={resetForm}
                  >
                    Cancel
                  </Button>
                  <Button
                    type="button"
                    className="rounded-full bg-purple-700 hover:bg-purple-800"
                    disabled={isPending || !title || !startAt || !endAt}
                    onClick={handleCreate}
                  >
                    {isPending ? "Adding..." : "Add Program Item"}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>
        </div>

        {error ? (
          <div className="mb-4 rounded border border-red-300 bg-red-50 p-2 text-xs text-red-700">
            {error}
          </div>
        ) : null}

        <div className="space-y-2 mb-4 max-h-56 overflow-y-auto pr-1">
          {sortedItems.length === 0 ? (
            <div className="text-sm text-purple-700">No program items yet.</div>
          ) : (
            sortedItems.map((item) => (
              <div
                key={item.id}
                className="rounded border border-purple-200 bg-white p-2 text-xs"
              >
                <div className="flex items-center justify-between gap-2">
                  <div className="font-semibold text-purple-900">
                    {item.position}. {item.title}
                  </div>
                  <Badge
                    variant="outline"
                    className={
                      item.status === "canceled"
                        ? "border-amber-300 text-amber-700"
                        : item.status === "in_progress"
                          ? "border-sky-300 text-sky-700"
                          : item.status === "paused"
                            ? "border-indigo-300 text-indigo-700"
                            : item.status === "ended"
                              ? "border-slate-300 text-slate-700"
                              : "border-emerald-300 text-emerald-700"
                    }
                  >
                    {item.status.toUpperCase()}
                  </Badge>
                </div>
                <div className="mt-1 text-purple-700">
                  {new Date(item.scheduledStart).toLocaleTimeString([], {
                    hour: "2-digit",
                    minute: "2-digit",
                  })}
                  {" - "}
                  {new Date(item.scheduledEnd).toLocaleTimeString([], {
                    hour: "2-digit",
                    minute: "2-digit",
                  })}
                  {item.hostName ? ` • ${item.hostName}` : ""}
                </div>
                <div className="mt-2 flex gap-1">
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    className="h-7 rounded-full"
                    disabled={isPending}
                    onClick={() => moveItem(item.id, -1)}
                  >
                    <ArrowUp className="h-3 w-3" />
                  </Button>
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    className="h-7 rounded-full"
                    disabled={isPending}
                    onClick={() => moveItem(item.id, 1)}
                  >
                    <ArrowDown className="h-3 w-3" />
                  </Button>
                  {item.status === "scheduled" ? (
                    <Button
                      type="button"
                      size="sm"
                      variant="outline"
                      className="h-7 rounded-full text-emerald-700 border-emerald-300"
                      disabled={isPending || !runtimeEnabled}
                      onClick={() => onStartAction(item.id)}
                    >
                      <Play className="mr-1 h-3 w-3" />
                      Start
                    </Button>
                  ) : null}
                  {item.status === "in_progress" || item.status === "paused" ? (
                    <Button
                      type="button"
                      size="sm"
                      variant="outline"
                      className="h-7 rounded-full text-amber-700 border-amber-300"
                      disabled={isPending || !runtimeEnabled}
                      onClick={() =>
                        item.status === "in_progress"
                          ? onPauseAction(item.id)
                          : onResumeAction(item.id)
                      }
                    >
                      {item.status === "in_progress" ? (
                        <>
                          <Pause className="mr-1 h-3 w-3" />
                          Pause
                        </>
                      ) : (
                        <>
                          <RotateCcw className="mr-1 h-3 w-3" />
                          Resume
                        </>
                      )}
                    </Button>
                  ) : null}
                  {item.status === "in_progress" || item.status === "paused" ? (
                    <Button
                      type="button"
                      size="sm"
                      variant="outline"
                      className="h-7 rounded-full text-sky-700 border-sky-300"
                      disabled={isPending || !runtimeEnabled}
                      onClick={() => onEndAction(item.id)}
                    >
                      <Square className="mr-1 h-3 w-3" />
                      End
                    </Button>
                  ) : null}
                  {item.status === "in_progress" || item.status === "paused" ? (
                    <>
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        className="h-7 rounded-full"
                        disabled={isPending || !runtimeEnabled}
                        onClick={() => onAdjustTimeAction(item.id, 60)}
                      >
                        <Plus className="mr-1 h-3 w-3" />
                        60s
                      </Button>
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        className="h-7 rounded-full"
                        disabled={isPending || !runtimeEnabled}
                        onClick={() => onAdjustTimeAction(item.id, -60)}
                      >
                        <Minus className="mr-1 h-3 w-3" />
                        60s
                      </Button>
                    </>
                  ) : null}
                  {item.status === "scheduled" ? (
                    <Button
                      type="button"
                      size="sm"
                      variant="outline"
                      className="h-7 rounded-full text-amber-700 border-amber-300"
                      disabled={isPending}
                      onClick={() => onCancelAction(item.id)}
                    >
                      <XCircle className="mr-1 h-3 w-3" />
                      Cancel
                    </Button>
                  ) : null}
                </div>
              </div>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  );
}
