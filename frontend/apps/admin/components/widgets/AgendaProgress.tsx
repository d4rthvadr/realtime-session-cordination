"use client";

import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { ClipboardList } from "lucide-react";

interface AgendaProgressProps {
  currentItem: number;
  totalItems: number;
  currentTitle: string;
  timeRemaining: string;
  progress: number;
}

export default function AgendaProgress({
  currentItem,
  totalItems,
  currentTitle,
  timeRemaining,
  progress,
}: AgendaProgressProps) {
  return (
    <Card className="col-span-6 bg-purple-50">
      <CardContent className="p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2 text-purple-700">
            <ClipboardList className="w-5 h-5" />
            <span className="text-xs font-semibold uppercase tracking-wider">
              Agenda Progress
            </span>
          </div>
          <Badge
            variant="outline"
            className="text-purple-700 text-xs tracking-wider"
          >
            ITEM {currentItem} OF {totalItems}
          </Badge>
        </div>
        <div className="text-lg font-semibold text-purple-900 mb-2">
          {currentTitle}
        </div>
        <div className="flex items-center gap-2">
          <Progress
            value={Math.min(100, Math.max(0, progress))}
            className="flex-1 h-1.5 bg-purple-200"
          />
          <span className="text-sm text-purple-700">{timeRemaining} left</span>
        </div>
      </CardContent>
    </Card>
  );
}
