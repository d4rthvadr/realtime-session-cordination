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
    <section className="border-t border-outline-variant bg-surface-container-low py-12 md:py-16 lg:py-24">
      <div className="mx-auto max-w-container-max px-4 md:px-10">
        <div className="text-center">
          <h2 className="font-headline text-2xl text-primary md:text-headline-lg">
            Everything You Need to Stay On Schedule
          </h2>
        </div>

        <div className="mt-8 grid gap-4 md:mt-12 md:gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((feature, index) => (
            <div
              key={index}
              className="rounded-xl border border-outline-variant bg-white p-6 transition hover:shadow-md"
            >
              <div className="flex items-start gap-4">
                <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg bg-secondary/10">
                  <span className="material-symbols-outlined text-secondary">
                    {feature.icon}
                  </span>
                </div>
                <div>
                  <h3 className="font-headline text-label-md text-primary">
                    {feature.title}
                  </h3>
                  <p className="mt-2 text-sm text-on-surface-variant">
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
