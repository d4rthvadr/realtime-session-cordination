"use client";

import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

interface AttendeeStatsProps {
  totalOnline: number;
  participationRate: number;
  attentionLevel: "Low" | "Medium" | "High";
}

export default function AttendeeStats({
  totalOnline,
  participationRate,
  attentionLevel,
}: AttendeeStatsProps) {
  return (
    <Card className="col-span-12 md:col-span-4 bg-slate-900 border-slate-700 text-white">
      <CardHeader className="pb-3">
        <div className="flex justify-between items-center gap-2">
          <h3 className="text-base sm:text-lg font-semibold text-white">
            Attendees
          </h3>
          <Badge variant="secondary" className="text-xs tracking-wider">
            {totalOnline} ONLINE
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="pb-4 sm:pb-6">
        <div className="flex -space-x-3 mb-4 sm:mb-6">
          <div className="w-12 h-12 rounded-full border-2 border-slate-900 overflow-hidden bg-slate-400">
            <div className="w-full h-full bg-gradient-to-br from-blue-400 to-purple-500 flex items-center justify-center text-white font-semibold">
              A
            </div>
          </div>
          <div className="w-12 h-12 rounded-full border-2 border-slate-900 overflow-hidden bg-slate-400">
            <div className="w-full h-full bg-gradient-to-br from-emerald-400 to-teal-500 flex items-center justify-center text-white font-semibold">
              B
            </div>
          </div>
          <div className="w-12 h-12 rounded-full border-2 border-slate-900 overflow-hidden bg-slate-400">
            <div className="w-full h-full bg-gradient-to-br from-orange-400 to-red-500 flex items-center justify-center text-white font-semibold">
              C
            </div>
          </div>
          {totalOnline > 3 && (
            <div className="w-12 h-12 rounded-full border-2 border-slate-900 bg-slate-200 flex items-center justify-center text-slate-900 text-xs font-bold">
              +{totalOnline - 3}
            </div>
          )}
        </div>

        <div className="space-y-3">
          <div className="flex justify-between items-center text-slate-300">
            <span className="text-sm">Participation Rate</span>
            <span className="text-sm font-semibold">{participationRate}%</span>
          </div>
          <div className="flex justify-between items-center text-slate-300">
            <span className="text-sm">Avg. Attention</span>
            <span className="text-sm font-semibold">{attentionLevel}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
