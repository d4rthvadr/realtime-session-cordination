export function FeaturesSection() {
  const features = [
    {
      title: "Instant Session Creation",
      description:
        "Create sessions in seconds with title, speaker, and duration",
      icon: "bolt",
    },
    {
      title: "Public Share Links",
      description: "Generate unique URLs viewable by anyone",
      icon: "link",
    },
    {
      title: "Host Controls",
      description: "Start, pause, resume, end, and adjust time on the fly",
      icon: "tune",
    },
    {
      title: "Real-Time Synchronization",
      description: "WebSocket-powered updates across all viewers instantly",
      icon: "sync",
    },
    {
      title: "Visual Urgency States",
      description:
        "Timer changes appearance: Safe → Warning → Critical → Overtime",
      icon: "traffic",
    },
    {
      title: "Zero Setup",
      description: "No authentication, no downloads, just create and share",
      icon: "rocket_launch",
    },
  ];

  return (
    <section className="border-t border-border-light bg-gradient-to-b from-surface-secondary to-white py-16 md:py-24 lg:py-32">
      <div className="mx-auto max-w-container-max px-4 md:px-10">
        <div className="text-center">
          <h2 className="font-headline text-3xl font-bold text-primary-dark md:text-5xl">
            Everything You Need to Stay On Schedule
          </h2>
        </div>

        <div className="mt-12 grid gap-6 md:mt-16 md:gap-8 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((feature, index) => (
            <div
              key={index}
              className="group rounded-2xl border border-border bg-white p-8 shadow-sm transition-all hover:shadow-xl hover:border-primary/20"
            >
              <div className="flex items-start gap-4">
                <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-primary to-primary-light shadow-md transition-transform group-hover:scale-110">
                  <span className="material-symbols-outlined text-xl text-white">
                    {feature.icon}
                  </span>
                </div>
                <div>
                  <h3 className="font-semibold text-base text-primary-dark">
                    {feature.title}
                  </h3>
                  <p className="mt-2 text-sm leading-relaxed text-text-secondary">
                    {feature.description}
                  </p>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
