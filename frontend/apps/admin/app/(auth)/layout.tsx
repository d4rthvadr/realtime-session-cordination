import { BarChart3 } from "lucide-react";

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen flex">
      {/* Left Side - Auth Form */}
      <div className="flex-1 flex items-center justify-center p-8 bg-white">
        <div className="w-full max-w-md">{children}</div>
      </div>

      {/* Right Side - Marketing Content */}
      <div className="hidden lg:flex flex-1 bg-gradient-to-br from-blue-600 via-blue-500 to-indigo-600 p-12 items-center justify-center relative overflow-hidden">
        {/* Background Pattern */}
        <div className="absolute inset-0 bg-grid-white/[0.05] bg-[size:20px_20px]" />

        {/* Content */}
        <div className="relative z-10 text-white max-w-lg">
          <div className="flex items-center gap-2 mb-8">
            <div className="w-10 h-10 rounded-full bg-white/20 backdrop-blur-sm flex items-center justify-center">
              <BarChart3 className="w-6 h-6" />
            </div>
            <span className="text-xl font-bold">SyncTime</span>
          </div>

          <h2 className="text-4xl font-bold mb-4">
            Real-Time Session Coordination for Better Outcomes
          </h2>

          <p className="text-lg text-blue-100 mb-8">
            Stay synchronized with real-time insights, performance analytics,
            and data-driven session management. All in one powerful dashboard.
          </p>

          {/* Mock Dashboard Preview */}
          <div className="relative bg-white/10 backdrop-blur-md rounded-xl p-4 border border-white/20">
            <div className="space-y-3">
              {/* Mock Stats */}
              <div className="grid grid-cols-3 gap-3">
                {[
                  { label: "Sessions", value: "247" },
                  { label: "Active", value: "12" },
                  { label: "Uptime", value: "99.8%" },
                ].map((stat) => (
                  <div
                    key={stat.label}
                    className="bg-white/10 rounded-lg p-3 backdrop-blur-sm"
                  >
                    <div className="text-2xl font-bold">{stat.value}</div>
                    <div className="text-xs text-blue-200">{stat.label}</div>
                  </div>
                ))}
              </div>

              {/* Mock Chart */}
              <div className="bg-white/5 rounded-lg p-4 backdrop-blur-sm h-32 flex items-end gap-1">
                {[40, 65, 45, 80, 55, 70, 85, 60, 75, 90, 70, 85].map(
                  (height, i) => (
                    <div
                      key={i}
                      className="flex-1 bg-gradient-to-t from-white/40 to-white/60 rounded-sm"
                      style={{ height: `${height}%` }}
                    />
                  ),
                )}
              </div>

              {/* Mock Activity */}
              <div className="space-y-2">
                {[
                  "Session started",
                  "New attendee joined",
                  "Poll published",
                ].map((activity, i) => (
                  <div
                    key={i}
                    className="bg-white/10 rounded-lg px-3 py-2 backdrop-blur-sm flex items-center gap-2"
                  >
                    <div className="w-2 h-2 rounded-full bg-emerald-400" />
                    <span className="text-sm text-blue-100">{activity}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Footer */}
          <div className="mt-8 text-sm text-blue-200">
            © {new Date().getFullYear()} SyncTime. All rights reserved.
          </div>
        </div>
      </div>
    </div>
  );
}
