"use client";

import { useEffect, useState, useTransition } from "react";
import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import SessionCreateModal from "@/components/SessionCreateModal";
import { getSessionsList, SessionSnapshot } from "@/lib/actions";
import { formatLocalDate } from "@/lib/date-time";
import { formatClock } from "@/lib/session";
import { getSessionStatusBadgeClasses } from "@/lib/session-status";
import { Clock, Calendar, Plus, ExternalLink } from "lucide-react";
import { cn } from "@/lib/utils";

export default function SessionsListPage() {
  const [sessions, setSessions] = useState<SessionSnapshot[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  const loadSessions = () => {
    startTransition(async () => {
      const result = await getSessionsList();
      if (result.error) {
        setError(result.error);
      } else {
        setSessions(result.sessions);
      }
    });
  };

  useEffect(() => {
    loadSessions();
  }, []);

  return (
    <div className="max-w-[1600px] mx-auto p-4 sm:p-6 lg:p-8">
      <Card className="border-slate-200">
        <CardHeader className="border-b border-slate-200">
          <div className="flex justify-between items-center">
            <div>
              <CardTitle className="text-2xl font-bold">All Sessions</CardTitle>
              <p className="text-sm text-slate-600 mt-1">
                Browse and manage all synchronized sessions
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
                      Created
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
                            getSessionStatusBadgeClasses(session.status),
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
                      <td className="p-4 text-sm text-slate-500">
                        {formatLocalDate(session.createdAt)}
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
    </div>
  );
}
