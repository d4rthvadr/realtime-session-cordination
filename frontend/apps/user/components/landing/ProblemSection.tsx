import { Card, CardContent } from "@/components/ui/card";

export function ProblemSection() {
  const problems = [
    {
      title: "Hidden Timers",
      description:
        "Speakers miss small displays or handheld cards in high-pressure environments.",
      icon: "visibility_off",
    },
    {
      title: "Desync Events",
      description:
        "When plans change mid-session, communicating the new schedule is impossible.",
      icon: "sync_problem",
    },
    {
      title: "Schedule Drift",
      description:
        "A 5-minute overrun early in the day ruins the entire conference timeline.",
      icon: "running_with_errors",
    },
    {
      title: "Anxious Speakers",
      description:
        "Performance suffers when speakers are unsure how much time they actually have left.",
      icon: "person_cancel",
    },
  ];

  return (
    <section className="mx-auto max-w-container-max px-4 py-16 md:px-10 md:py-28">
      <div className="mb-12 text-center md:mb-20">
        <h2 className="font-headline text-3xl font-bold text-slate-100 md:text-5xl">
          Coordination shouldn&apos;t be chaotic.
        </h2>
        <p className="mt-4 text-base text-slate-400 md:mt-6 md:text-lg">
          Stop relying on hand signals and frantic texts.
        </p>
      </div>
      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 md:gap-8 lg:grid-cols-4">
        {problems.map((problem) => (
          <Card
            key={problem.title}
            className="group rounded-2xl border-slate-800 bg-slate-900/70 transition-all hover:border-slate-700 hover:shadow-[0_16px_40px_rgba(0,0,0,0.35)]"
          >
            <CardContent className="space-y-4 p-8">
              <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-slate-800 transition-transform group-hover:scale-110">
                <span className="material-symbols-outlined text-slate-200">
                  {problem.icon}
                </span>
              </div>
              <h3 className="text-base font-semibold text-slate-100">
                {problem.title}
              </h3>
              <p className="text-sm leading-relaxed text-slate-400">
                {problem.description}
              </p>
            </CardContent>
          </Card>
        ))}
      </div>
    </section>
  );
}
