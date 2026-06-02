"use client";

"use client";

import { useEffect, useState, useTransition } from "react";
import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import SessionCreateModal from "@/components/SessionCreateModal";
import {
  getAnalyticsOverview,
  getSessionsList,
  SessionSnapshot,
  AnalyticsOverview,
} from "@/lib/actions";
import { formatClock } from "@/lib/session";
import {
  Clock,
  Users,
  Calendar,
  TrendingUp,
  ExternalLink,
  Plus,
} from "lucide-react";
import { cn } from "@/lib/utils";

export default function DashboardPage() {
  const [sessions, setSessions] = useState<SessionSnapshot[]>([]);
  const [overview, setOverview] = useState<AnalyticsOverview | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  const loadSessions = () => {
    startTransition(async () => {
      const [sessionsResult, overviewResult] = await Promise.all([
        getSessionsList(),
        getAnalyticsOverview(),
      ]);

      if (sessionsResult.error) {
        setError(sessionsResult.error);
      } else {
        setSessions(sessionsResult.sessions);
      }

      if (overviewResult.error) {
        setOverview(null);
      } else {
        setOverview(overviewResult.overview);
      }
    });
  };

  useEffect(() => {
    loadSessions();
  }, []);

  // Calculate stats from sessions
  const fallbackTotalSessions = sessions.length;
  const fallbackActiveSessions = sessions.filter(
    (s) => s.status === "LIVE" || s.status === "PAUSED",
  ).length;
  const fallbackTodaySessions = sessions.filter((s) => {
    const createdAt = new Date(s.createdAt);
    const today = new Date();
    return createdAt.toDateString() === today.toDateString();
  }).length;
  const fallbackAvgDuration =
    sessions.length > 0
      ? Math.round(
          sessions.reduce((acc, s) => acc + s.durationSeconds, 0) /
            sessions.length /
            60,
        )
      : 0;

  const totalSessions = overview?.totalSessions ?? fallbackTotalSessions;
  const activeSessions =
    overview?.liveSessions != null && overview?.pausedSessions != null
      ? overview.liveSessions + overview.pausedSessions
      : fallbackActiveSessions;
  const todaySessions = fallbackTodaySessions;
  const avgDuration =
    overview?.totalSessions && overview.totalSessions > 0
      ? Math.round(
          overview.totalSessionDurationSeconds / overview.totalSessions / 60,
        )
      : fallbackAvgDuration;

  const getStatusColor = (status: string) => {
    switch (status) {
      case "LIVE":
        return "bg-emerald-100 text-emerald-700 border-emerald-200";
      case "PAUSED":
        return "bg-amber-100 text-amber-700 border-amber-200";
      case "ENDED":
        return "bg-slate-100 text-slate-600 border-slate-200";
      default:
        return "bg-blue-100 text-blue-700 border-blue-200";
    }
  };

  return (
    <div className="max-w-[1600px] mx-auto p-4 sm:p-6 lg:p-8 space-y-6">
      {/* Header Section */}
      <section className="grid grid-cols-1 md:grid-cols-4 lg:grid-cols-5 gap-4">
        {/* Global Time Health Card */}
        <Card className="md:col-span-3 lg:col-span-4 border-slate-200">
          <CardHeader className="pb-3">
            <div className="flex justify-between items-start">
              <div>
                <CardTitle className="text-2xl font-bold text-slate-900">
                  Global Time Health
                </CardTitle>
                <p className="text-sm text-slate-600 mt-1">
                  Real-time enterprise synchronization status
                </p>
              </div>
              <Badge className="bg-emerald-50 text-emerald-700 border-emerald-200 hover:bg-emerald-50">
                <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse mr-2"></span>
                NETWORK OPTIMIZED
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
              <div className="space-y-1">
                <span className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
                  Total Sessions
                </span>
                <div className="flex items-baseline gap-1">
                  <span className="text-3xl md:text-4xl font-bold text-slate-900">
                    {totalSessions}
                  </span>
                  <span className="text-sm text-slate-500">total</span>
                </div>
              </div>
              <div className="space-y-1">
                <span className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
                  Active Sessions
                </span>
                <div className="flex items-baseline gap-1">
                  <span className="text-3xl md:text-4xl font-bold text-blue-600">
                    {activeSessions}
                  </span>
                  <span className="text-sm text-slate-500">LIVE</span>
                </div>
              </div>
              <div className="space-y-1">
                <span className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
                  Today&apos;s Sessions
                </span>
                <div className="flex items-baseline gap-1">
                  <span className="text-3xl md:text-4xl font-bold text-slate-900">
                    {todaySessions}
                  </span>
                  <span className="text-sm text-slate-500">today</span>
                </div>
              </div>
              <div className="space-y-1">
                <span className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
                  Avg Duration
                </span>
                <div className="flex items-baseline gap-1">
                  <span className="text-3xl md:text-4xl font-bold text-slate-900">
                    {avgDuration}
                  </span>
                  <span className="text-sm text-slate-500">min</span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Quick Action Card - Initialize Session */}
        <Card className="bg-slate-900 text-white border-slate-800 hover:scale-[1.02] transition-transform duration-300 cursor-pointer">
          <CardContent className="p-6 flex flex-col justify-between h-full">
            <div>
              <Plus className="w-8 h-8 mb-4 text-slate-400" />
              <h3 className="text-xl font-semibold mb-2">Initialize Session</h3>
              <p className="text-sm text-slate-400">
                Deploy a new atomic-locked session across the cluster.
              </p>
            </div>
            <div className="mt-6">
              <SessionCreateModal
                onSuccess={loadSessions}
                trigger={
                  <Button className="w-full bg-white text-slate-900 hover:bg-slate-100 rounded-full">
                    Rapid Launch
                    <ExternalLink className="w-4 h-4 ml-2" />
                  </Button>
                }
              />
            </div>
          </CardContent>
        </Card>
      </section>

      {/* Sessions List */}
      <section>
        <Card className="border-slate-200">
          <CardHeader className="border-b border-slate-200">
            <div className="flex justify-between items-center">
              <div>
                <CardTitle className="text-xl font-semibold">
                  Session Coordination Log
                </CardTitle>
                <p className="text-sm text-slate-600 mt-1">
                  Manage and monitor all synchronized sessions
                </p>
              </div>
              <SessionCreateModal onSuccess={loadSessions} />
            </div>
          </CardHeader>
          <CardContent className="p-0">
            {isPending && sessions.length === 0 ? (
              <div className="text-center py-12 text-slate-500">
                <Clock className="w-12 h-12 mx-auto mb-4 animate-spin text-slate-300" />
                <p>Loading sessions...</p>
              </div>
            ) : error ? (
              <div className="text-center py-12 text-red-600">
                <p>Error loading sessions: {error}</p>
              </div>
            ) : sessions.length === 0 ? (
              <div className="text-center py-12 text-slate-500">
                <Calendar className="w-12 h-12 mx-auto mb-4 text-slate-300" />
                <p className="mb-4">No sessions created yet</p>
                <SessionCreateModal
                  onSuccess={loadSessions}
                  trigger={
                    <Button variant="outline" className="rounded-full">
                      <Plus className="w-4 h-4 mr-2" />
                      Create Your First Session
                    </Button>
                  }
                />
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left border-collapse">
                  <thead className="bg-slate-50 border-b border-slate-200">
                    <tr>
                      <th className="p-4 text-xs font-semibold text-slate-600 uppercase tracking-wider">
                        Session
                      </th>
                      <th className="p-4 text-xs font-semibold text-slate-600 uppercase tracking-wider">
                        Speaker
                      </th>
                      <th className="p-4 text-xs font-semibold text-slate-600 uppercase tracking-wider">
                        Status
                      </th>
                      <th className="p-4 text-xs font-semibold text-slate-600 uppercase tracking-wider">
                        Duration
                      </th>
                      <th className="p-4 text-xs font-semibold text-slate-600 uppercase tracking-wider">
                        Remaining
                      </th>
                      <th className="p-4 text-xs font-semibold text-slate-600 uppercase tracking-wider">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100">
                    {sessions.map((session) => (
                      <tr
                        key={session.id}
                        className="hover:bg-slate-50 transition-colors"
                      >
                        <td className="p-4">
                          <div className="font-medium text-slate-900">
                            {session.title}
                          </div>
                          <div className="text-xs text-slate-500 font-mono mt-0.5">
                            #{session.id.slice(0, 8)}
                          </div>
                        </td>
                        <td className="p-4 text-sm text-slate-700">
                          {session.speakerName}
                        </td>
                        <td className="p-4">
                          <Badge
                            className={cn(
                              "text-xs font-semibold border",
                              getStatusColor(session.status),
                            )}
                          >
                            {session.status}
                          </Badge>
                        </td>
                        <td className="p-4 text-sm text-slate-700 font-mono">
                          {formatClock(session.durationSeconds, "00:00")}
                        </td>
                        <td className="p-4 text-sm text-slate-700 font-mono">
                          {formatClock(session.durationSeconds, "00:00")}
                        </td>
                        <td className="p-4">
                          <Link href={`/dashboard/sessions/${session.id}`}>
                            <Button
                              variant="outline"
                              size="sm"
                              className="rounded-full"
                            >
                              View Details
                              <ExternalLink className="w-3 h-3 ml-2" />
                            </Button>
                          </Link>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </CardContent>
        </Card>
      </section>
    </div>
  );
}
