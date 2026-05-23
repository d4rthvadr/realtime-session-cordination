import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";

export function SemanticAlertsSection() {
  const alerts = [
    {
      label: "SAFE",
      description: "Speaker has ample time.",
      icon: "check_circle",
      bgColor: "bg-green-100",
      textColor: "text-green-700",
      labelColor: "text-green-800",
    },
    {
      label: "WARNING",
      description: "2 minutes remaining.",
      icon: "warning",
      bgColor: "bg-amber-100",
      textColor: "text-amber-700",
      labelColor: "text-amber-800",
    },
    {
      label: "CRITICAL",
      description: "30 seconds to wrap.",
      icon: "error",
      bgColor: "bg-red-100",
      textColor: "text-red-700",
      labelColor: "text-red-800",
    },
    {
      label: "OVERTIME",
      description: "Elapsed tracking active.",
      icon: "timer_off",
      bgColor: "bg-slate-900",
      textColor: "text-white",
      labelColor: "text-slate-900",
    },
  ];

  return (
    <section className="border-y border-slate-200 bg-slate-50/70">
      <div className="mx-auto max-w-container-max px-4 py-12 md:px-10 md:py-16">
        <p className="mb-8 text-center text-xs font-semibold uppercase tracking-widest text-slate-500 md:mb-10 md:text-sm">
          Universal Semantic Alerts
        </p>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 md:gap-6 lg:grid-cols-4">
          {alerts.map((alert, index) => (
            <Card key={index} className="rounded-2xl border-slate-200">
              <CardContent className="flex items-center gap-4 p-5">
                <div
                  className={`flex h-12 w-12 items-center justify-center rounded-full ${alert.bgColor} ${alert.textColor}`}
                >
                  <span className="material-symbols-outlined">
                    {alert.icon}
                  </span>
                </div>
                <div>
                  <Badge
                    variant="outline"
                    className={`border-transparent px-0 py-0 text-xs ${alert.labelColor === "text-slate-900" ? "text-slate-900" : alert.labelColor}`}
                  >
                    {alert.label}
                  </Badge>
                  <p className="mt-1 text-xs text-slate-500">
                    {alert.description}
                  </p>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </section>
  );
}
