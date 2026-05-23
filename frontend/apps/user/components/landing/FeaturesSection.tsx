import { Card, CardContent } from "@/components/ui/card";

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
    <section className="border-t border-slate-200 bg-gradient-to-b from-slate-50 to-white py-16 md:py-24 lg:py-32">
      <div className="mx-auto max-w-container-max px-4 md:px-10">
        <div className="text-center">
          <h2 className="font-headline text-3xl font-bold text-slate-900 md:text-5xl">
            Everything You Need to Stay On Schedule
          </h2>
        </div>

        <div className="mt-12 grid gap-6 md:mt-16 md:gap-8 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((feature, index) => (
            <Card
              key={index}
              className="group rounded-2xl border-slate-200 transition-all hover:border-slate-300 hover:shadow-xl"
            >
              <CardContent className="p-8">
                <div className="flex items-start gap-4">
                  <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-xl bg-slate-900 shadow-md transition-transform group-hover:scale-110">
                    <span className="material-symbols-outlined text-xl text-white">
                      {feature.icon}
                    </span>
                  </div>
                  <div>
                    <h3 className="text-base font-semibold text-slate-900">
                      {feature.title}
                    </h3>
                    <p className="mt-2 text-sm leading-relaxed text-slate-600">
                      {feature.description}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </section>
  );
}
