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
        <h2 className="font-headline text-3xl font-bold text-primary-dark md:text-5xl">
          Coordination shouldn't be chaotic.
        </h2>
        <p className="mt-4 text-base text-text-secondary md:mt-6 md:text-lg">
          Stop relying on hand signals and frantic texts.
        </p>
      </div>
      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 md:gap-8 lg:grid-cols-4">
        {problems.map((problem, index) => (
          <div
            key={index}
            className="group space-y-4 rounded-2xl border border-border bg-white p-8 shadow-sm transition-all hover:shadow-lg hover:border-primary/20"
          >
            <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br from-primary/10 to-accent-purple/10 transition-transform group-hover:scale-110">
              <span className="material-symbols-outlined text-primary">
                {problem.icon}
              </span>
            </div>
            <h3 className="font-semibold text-base text-primary-dark">
              {problem.title}
            </h3>
            <p className="text-sm leading-relaxed text-text-secondary">
              {problem.description}
            </p>
          </div>
        ))}
      </div>
    </section>
  );
}
