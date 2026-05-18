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
    <section className="mx-auto max-w-container-max px-4 py-12 md:px-10 md:py-24">
      <div className="mb-8 text-center md:mb-16">
        <h2 className="font-headline text-2xl text-primary md:text-headline-lg">
          Coordination shouldn't be chaotic.
        </h2>
        <p className="mt-3 text-sm text-on-surface-variant md:mt-4 md:text-base">
          Stop relying on hand signals and frantic texts.
        </p>
      </div>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 md:gap-8 lg:grid-cols-4">
        {problems.map((problem, index) => (
          <div
            key={index}
            className="space-y-4 rounded-2xl border border-outline-variant bg-white p-8"
          >
            <span className="material-symbols-outlined text-secondary">
              {problem.icon}
            </span>
            <h3 className="font-headline text-label-md">{problem.title}</h3>
            <p className="text-sm text-on-surface-variant">
              {problem.description}
            </p>
          </div>
        ))}
      </div>
    </section>
  );
}
