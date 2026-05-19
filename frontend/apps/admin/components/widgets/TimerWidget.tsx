"use client";

import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { Pause, RefreshCw } from "lucide-react";

interface TimerWidgetProps {
  currentTime: string;
  totalTime: string;
  progress: number;
  status: "CREATED" | "LIVE" | "PAUSED" | "ENDED";
  onPause?: () => void;
  onRefresh?: () => void;
}

export default function TimerWidget({
  currentTime,
  totalTime,
  progress,
  status,
  onPause,
  onRefresh,
}: TimerWidgetProps) {
  const getStatusLabel = () => {
    switch (status) {
      case "LIVE":
        return "LIVE SESSION";
      case "PAUSED":
        return "PAUSED";
      case "ENDED":
        return "ENDED";
      default:
        return "READY";
    }
  };

  const getStatusVariant = () => {
    switch (status) {
      case "LIVE":
        return "default" as const;
      case "PAUSED":
        return "warning" as const;
      case "ENDED":
        return "secondary" as const;
      default:
        return "success" as const;
    }
  };

  return (
    <Card className="col-span-8 p-6 flex flex-col justify-between">
      <div className="flex justify-between items-start">
        <Badge variant={getStatusVariant()} className="tracking-wider">
          {getStatusLabel()}
        </Badge>
        <div className="flex gap-2">
          {onPause && status === "LIVE" && (
            <Button
              onClick={onPause}
              variant="ghost"
              size="icon"
              aria-label="Pause session"
              className="rounded-full"
            >
              <Pause className="w-5 h-5" />
            </Button>
          )}
          {onRefresh && (
            <Button
              onClick={onRefresh}
              variant="ghost"
              size="icon"
              aria-label="Refresh session"
              className="rounded-full"
            >
              <RefreshCw className="w-5 h-5" />
            </Button>
          )}
        </div>
      </div>

      <div className="flex items-baseline justify-center py-10">
        <span className="text-8xl md:text-9xl font-bold leading-none tracking-tighter">
          {currentTime}
        </span>
        <span className="text-2xl md:text-3xl font-semibold text-muted-foreground ml-4">
          / {totalTime}
        </span>
      </div>

      <Progress value={Math.min(100, Math.max(0, progress))} className="h-2" />
    </Card>
  );
}
