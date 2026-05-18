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
    <section className="border-y border-outline-variant bg-surface-container-low">
      <div className="mx-auto max-w-container-max px-10 py-12">
        <p className="font-label-md mb-8 text-center text-label-md uppercase tracking-widest text-on-surface-variant">
          Universal Semantic Alerts
        </p>
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
          {alerts.map((alert, index) => (
            <div
              key={index}
              className="flex items-center gap-4 rounded-xl border border-outline-variant bg-white p-4"
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
                <p className="text-xs text-on-surface-variant">
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
