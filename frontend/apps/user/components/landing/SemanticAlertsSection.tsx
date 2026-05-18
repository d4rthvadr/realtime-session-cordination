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
    <section className="border-y border-border-light bg-surface-secondary">
      <div className="mx-auto max-w-container-max px-4 py-12 md:px-10 md:py-16">
        <p className="font-semibold mb-8 text-center text-xs uppercase tracking-widest text-text-tertiary md:mb-10 md:text-sm">
          Universal Semantic Alerts
        </p>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 md:gap-6 lg:grid-cols-4">
          {alerts.map((alert, index) => (
            <div
              key={index}
              className="flex items-center gap-4 rounded-2xl border border-border bg-white p-5 shadow-sm transition-shadow hover:shadow-md"
            >
              <div
                className={`flex h-12 w-12 items-center justify-center rounded-full ${alert.bgColor} ${alert.textColor}`}
              >
                <span className="material-symbols-outlined">{alert.icon}</span>
              </div>
              <div>
                <p
                  className={`font-label-md text-label-md ${alert.labelColor === "text-slate-900" ? "text-slate-900" : alert.labelColor}`}
                >
                  {alert.label}
                </p>
                <p className="text-xs text-text-tertiary">
                  {alert.description}
                </p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
