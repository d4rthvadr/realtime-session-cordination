"use client";

import { useState, useTransition } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { DeadLetterRecord } from "@/lib/actions";
import { getDLQList, retryDLQItem } from "@/lib/actions";
import { cn } from "@/lib/utils";
import {
  AlertTriangle,
  ChevronLeft,
  ChevronRight,
  RefreshCw,
  X,
} from "lucide-react";

const PAGE_SIZE = 50;

interface DLQSectionProps {
  initialRows: DeadLetterRecord[];
  initialCount: number;
}

function truncate(str: string, max: number): string {
  if (!str) return "-";
  return str.length <= max ? str : str.slice(0, max) + "...";
}

function formatTimestamp(ts: string | null | undefined): string {
  if (!ts) return "-";
  const d = new Date(ts);
  if (!Number.isFinite(d.getTime())) return "-";
  return d.toLocaleString();
}

function DetailDialog({
  row,
  onClose,
  onRetry,
  retrying,
  retrySuccess,
}: {
  row: DeadLetterRecord;
  onClose: () => void;
  onRetry: (id: number) => void;
  retrying: boolean;
  retrySuccess: boolean;
}) {
  let parsedPayload: string | null = null;
  if (row.payloadJson) {
    try {
      parsedPayload = JSON.stringify(JSON.parse(row.payloadJson), null, 2);
    } catch {
      parsedPayload = row.payloadJson;
    }
  }

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-base font-semibold">
            Dead-letter row #{row.outboxId}
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-3 text-sm">
          <div className="grid grid-cols-2 gap-x-6 gap-y-2">
            <div>
              <p className="text-[10px] uppercase tracking-widest text-slate-400 mb-0.5">
                Event ID
              </p>
              <p className="font-mono text-xs break-all text-slate-700">
                {row.eventId}
              </p>
            </div>
            <div>
              <p className="text-[10px] uppercase tracking-widest text-slate-400 mb-0.5">
                Event key
              </p>
              <p className="font-mono text-xs text-slate-700">{row.eventKey}</p>
            </div>
            <div>
              <p className="text-[10px] uppercase tracking-widest text-slate-400 mb-0.5">
                Session ID
              </p>
              <p className="font-mono text-xs break-all text-slate-700">
                {row.sessionId}
              </p>
            </div>
            <div>
              <p className="text-[10px] uppercase tracking-widest text-slate-400 mb-0.5">
                Program Item ID
              </p>
              <p className="font-mono text-xs break-all text-slate-700">
                {row.programItemId ?? "-"}
              </p>
            </div>
            <div>
              <p className="text-[10px] uppercase tracking-widest text-slate-400 mb-0.5">
                Occurred at
              </p>
              <p className="text-xs text-slate-700">
                {formatTimestamp(row.occurredAt)}
              </p>
            </div>
            <div>
              <p className="text-[10px] uppercase tracking-widest text-slate-400 mb-0.5">
                Ingested at
              </p>
              <p className="text-xs text-slate-700">
                {formatTimestamp(row.ingestedAt)}
              </p>
            </div>
            <div>
              <p className="text-[10px] uppercase tracking-widest text-slate-400 mb-0.5">
                Failed at
              </p>
              <p className="text-xs text-slate-700">
                {formatTimestamp(row.failedAt)}
              </p>
            </div>
            <div>
              <p className="text-[10px] uppercase tracking-widest text-slate-400 mb-0.5">
                Attempt
              </p>
              <p className="font-mono text-xs text-slate-700">{row.attempt}</p>
            </div>
          </div>

          <div>
            <p className="text-[10px] uppercase tracking-widest text-slate-400 mb-0.5">
              Last error
            </p>
            <p className="text-xs text-red-700 bg-red-50 rounded p-2 font-mono break-all">
              {row.lastError || "-"}
            </p>
          </div>

          {parsedPayload && (
            <div>
              <p className="text-[10px] uppercase tracking-widest text-slate-400 mb-0.5">
                Payload
              </p>
              <pre className="text-xs bg-slate-50 rounded p-2 overflow-x-auto whitespace-pre-wrap break-all text-slate-700 border border-slate-200">
                {parsedPayload}
              </pre>
            </div>
          )}
        </div>

        <div className="flex items-center gap-3 pt-2">
          {retrySuccess ? (
            <Badge className="bg-emerald-50 text-emerald-700 border border-emerald-200">
              Queued for retry
            </Badge>
          ) : (
            <Button
              size="sm"
              variant="outline"
              disabled={retrying}
              onClick={() => onRetry(row.outboxId)}
              className="gap-1.5"
            >
              <RefreshCw
                className={cn("h-3.5 w-3.5", retrying && "animate-spin")}
              />
              {retrying ? "Queuing..." : "Retry"}
            </Button>
          )}
          <Button
            size="sm"
            variant="ghost"
            onClick={onClose}
            className="ml-auto gap-1.5"
          >
            <X className="h-3.5 w-3.5" />
            Close
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

export default function DLQSection({
  initialRows,
  initialCount,
}: DLQSectionProps) {
  const [rows, setRows] = useState<DeadLetterRecord[]>(initialRows);
  const [totalCount, setTotalCount] = useState(initialCount);
  const [page, setPage] = useState(0);
  const [selectedRow, setSelectedRow] = useState<DeadLetterRecord | null>(null);
  const [retriedIds, setRetriedIds] = useState<Set<number>>(new Set());
  const [retryingId, setRetryingId] = useState<number | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  const totalPages = Math.max(1, Math.ceil(totalCount / PAGE_SIZE));

  function loadPage(targetPage: number) {
    startTransition(async () => {
      setLoadError(null);
      const result = await getDLQList(PAGE_SIZE, targetPage * PAGE_SIZE);
      if (result.error) {
        setLoadError(result.error);
        return;
      }
      setRows(result.rows);
      setTotalCount(result.count);
      setPage(targetPage);
    });
  }

  async function handleRetry(outboxId: number) {
    setRetryingId(outboxId);
    try {
      const result = await retryDLQItem(outboxId);
      if (!result.error) {
        setRetriedIds((prev) => new Set(prev).add(outboxId));
        setTimeout(() => {
          setRows((prev) => prev.filter((r) => r.outboxId !== outboxId));
          setTotalCount((prev) => Math.max(0, prev - 1));
        }, 1200);
      }
    } finally {
      setRetryingId(null);
    }
  }

  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <AlertTriangle className="h-4 w-4 text-red-500" />
            <CardTitle className="text-base font-semibold text-slate-800">
              Dead-letter Queue
            </CardTitle>
            {totalCount > 0 && (
              <Badge className="bg-red-50 text-red-700 border border-red-200 text-[10px] font-bold">
                {totalCount}
              </Badge>
            )}
          </div>
          {totalPages > 1 && (
            <div className="flex items-center gap-1 text-xs text-slate-500">
              <Button
                size="icon"
                variant="ghost"
                className="h-7 w-7"
                disabled={page === 0 || isPending}
                onClick={() => loadPage(page - 1)}
              >
                <ChevronLeft className="h-3.5 w-3.5" />
              </Button>
              <span className="tabular-nums">
                {page + 1} / {totalPages}
              </span>
              <Button
                size="icon"
                variant="ghost"
                className="h-7 w-7"
                disabled={page >= totalPages - 1 || isPending}
                onClick={() => loadPage(page + 1)}
              >
                <ChevronRight className="h-3.5 w-3.5" />
              </Button>
            </div>
          )}
        </div>
      </CardHeader>

      <CardContent>
        {loadError && <p className="text-xs text-red-600 mb-3">{loadError}</p>}

        {rows.length === 0 ? (
          <p className="text-sm text-slate-400 text-center py-8">
            No dead-letter rows.
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead>
                <tr className="border-b border-slate-100">
                  <th className="text-left py-2 pr-3 text-[10px] uppercase tracking-widest text-slate-400 font-semibold w-14">
                    ID
                  </th>
                  <th className="text-left py-2 pr-3 text-[10px] uppercase tracking-widest text-slate-400 font-semibold">
                    Event key
                  </th>
                  <th className="text-left py-2 pr-3 text-[10px] uppercase tracking-widest text-slate-400 font-semibold hidden sm:table-cell">
                    Session
                  </th>
                  <th className="py-2 pr-3 text-[10px] uppercase tracking-widest text-slate-400 font-semibold w-14 text-right hidden md:table-cell">
                    Attempt
                  </th>
                  <th className="text-left py-2 pr-3 text-[10px] uppercase tracking-widest text-slate-400 font-semibold hidden lg:table-cell">
                    Failed at
                  </th>
                  <th className="text-left py-2 text-[10px] uppercase tracking-widest text-slate-400 font-semibold">
                    Error
                  </th>
                  <th className="w-20" />
                </tr>
              </thead>
              <tbody>
                {rows.map((row) => {
                  const isRetried = retriedIds.has(row.outboxId);
                  const isRetrying = retryingId === row.outboxId;
                  return (
                    <tr
                      key={row.outboxId}
                      className={cn(
                        "border-b border-slate-50 cursor-pointer hover:bg-slate-50 transition-colors",
                        isRetried && "opacity-50",
                      )}
                      onClick={() => setSelectedRow(row)}
                    >
                      <td className="py-2 pr-3 font-mono text-slate-500">
                        {row.outboxId}
                      </td>
                      <td className="py-2 pr-3 font-mono text-slate-700 whitespace-nowrap">
                        {row.eventKey}
                      </td>
                      <td className="py-2 pr-3 font-mono text-slate-500 hidden sm:table-cell">
                        {truncate(row.sessionId, 12)}
                      </td>
                      <td className="py-2 pr-3 text-right text-slate-600 hidden md:table-cell">
                        {row.attempt}
                      </td>
                      <td className="py-2 pr-3 text-slate-500 whitespace-nowrap hidden lg:table-cell">
                        {formatTimestamp(row.failedAt)}
                      </td>
                      <td className="py-2 pr-3 text-red-600">
                        {truncate(row.lastError, 48)}
                      </td>
                      <td
                        className="py-2 text-right"
                        onClick={(e) => e.stopPropagation()}
                      >
                        {isRetried ? (
                          <Badge className="bg-emerald-50 text-emerald-700 border border-emerald-200 text-[10px]">
                            queued
                          </Badge>
                        ) : (
                          <Button
                            size="sm"
                            variant="ghost"
                            className="h-6 px-2 text-[11px] gap-1"
                            disabled={isRetrying || isPending}
                            onClick={() => handleRetry(row.outboxId)}
                          >
                            <RefreshCw
                              className={cn(
                                "h-3 w-3",
                                isRetrying && "animate-spin",
                              )}
                            />
                            {isRetrying ? "..." : "Retry"}
                          </Button>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </CardContent>

      {selectedRow && (
        <DetailDialog
          row={selectedRow}
          onClose={() => setSelectedRow(null)}
          onRetry={handleRetry}
          retrying={retryingId === selectedRow.outboxId}
          retrySuccess={retriedIds.has(selectedRow.outboxId)}
        />
      )}
    </Card>
  );
}
