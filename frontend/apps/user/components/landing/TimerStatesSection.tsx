export default function TimerStatesSection() {
  const states = [
    {
      label: "Safe",
      time: ">25% remaining",
      color: "text-safe",
      bgColor: "bg-safe/10",
      borderColor: "border-safe/30",
    },
    {
      label: "Warning",
      time: "10-25% remaining",
      color: "text-warning",
      bgColor: "bg-warning/10",
      borderColor: "border-warning/30",
    },
    {
      label: "Critical",
      time: "<10% remaining",
      color: "text-critical",
      bgColor: "bg-critical/10",
      borderColor: "border-critical/30",
    },
    {
      label: "Overtime",
      time: "Time exceeded",
      color: "text-critical",
      bgColor: "bg-critical/20",
      borderColor: "border-critical",
      blink: true,
    },
  ];

  return (
    <section className="py-16 sm:py-24">
      <div className="mx-auto max-w-7xl px-6">
        <div className="text-center">
          <h2 className="text-3xl font-bold tracking-tight text-slate-50 sm:text-4xl">
            Intuitive Visual Feedback
          </h2>
          <p className="mx-auto mt-4 max-w-2xl text-lg text-slate-300">
            The countdown timer automatically changes appearance to signal
            urgency
          </p>
        </div>

        <div className="mt-12 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
          {states.map((state, index) => (
            <div
              key={index}
              className={`rounded-lg border ${state.borderColor} ${state.bgColor} p-6 text-center transition ${
                state.blink ? "animate-pulse" : ""
              }`}
            >
              <div className={`text-5xl font-bold ${state.color}`}>
                {state.label === "Overtime" ? "-02:15" : "12:30"}
              </div>
              <div className="mt-4 space-y-1">
                <div className={`text-sm font-semibold ${state.color}`}>
                  {state.label}
                </div>
                <div className="text-xs text-slate-400">{state.time}</div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
