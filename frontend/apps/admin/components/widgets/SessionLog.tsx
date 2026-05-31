"use client";

import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { CheckCircle2, AlertTriangle, AlertCircle, Info } from "lucide-react";

export interface LogEntry {
  timestamp: string;
  message: string;
  type: "info" | "success" | "warning" | "error";
}

interface SessionLogProps {
  entries: LogEntry[];
  onExport?: () => void;
}

export default function SessionLog({ entries, onExport }: SessionLogProps) {
  const getIconForType = (type: LogEntry["type"]) => {
    switch (type) {
      case "success":
        return <CheckCircle2 className="w-4 h-4 text-green-500" />;
      case "warning":
        return <AlertTriangle className="w-4 h-4 text-orange-500" />;
      case "error":
        return <AlertCircle className="w-4 h-4 text-red-500" />;
      default:
        return <Info className="w-4 h-4 text-blue-500" />;
    }
  };

  return (
    <Card className="col-span-12">
      <CardHeader className="pb-3 sm:pb-6">
        <div className="flex justify-between items-center gap-2">
          <h3 className="text-base sm:text-lg font-semibold">Session Log</h3>
          {onExport && (
            <Button
              onClick={onExport}
              variant="ghost"
              size="sm"
              className="text-xs rounded-full"
            >
              Export CSV
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent className="px-4 sm:px-6">
        <div className="space-y-0 max-h-[200px] sm:max-h-[250px] overflow-y-auto pr-2 sm:pr-4 scrollbar-thin scrollbar-thumb-slate-300 scrollbar-track-transparent">
          {entries.length === 0 ? (
            <div className="text-center py-8 text-slate-400 text-sm">
              No log entries yet
            </div>
          ) : (
            entries.map((entry, index) => (
              <div
                key={index}
                className="flex items-center py-2 sm:py-3 border-b border-slate-100 last:border-0 gap-2 sm:gap-0"
              >
                <span className="w-16 sm:w-20 text-xs font-semibold text-slate-500 flex-shrink-0">
                  {entry.timestamp}
                </span>
                <span className="flex-1 text-xs sm:text-sm text-slate-900 px-2 sm:px-4">
                  {entry.message}
                </span>
                {getIconForType(entry.type)}
              </div>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  );
}
