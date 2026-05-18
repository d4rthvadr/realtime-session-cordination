export default function HeroSection() {
  return (
    <section className="relative mx-auto flex max-w-container-max flex-col items-center gap-8 overflow-hidden px-4 py-16 md:gap-16 md:px-10 md:py-28 lg:flex-row">
      {/* Gradient Background */}
      <div className="pointer-events-none absolute inset-0 -z-10">
        <div className="absolute inset-0 bg-gradient-to-br from-primary/5 via-transparent to-accent-purple/5"></div>
        <div className="absolute left-0 top-0 h-96 w-96 rounded-full bg-primary/10 blur-3xl"></div>
        <div className="absolute right-0 bottom-0 h-96 w-96 rounded-full bg-accent-cyan/10 blur-3xl"></div>
      </div>

      {/* Content */}
      <div className="flex-1 space-y-6 text-center md:space-y-8 lg:text-left">
        <h1 className="font-display text-4xl font-bold text-primary-dark md:max-w-xl md:text-display-lg lg:text-6xl">
          Keep Every Speaker On Time, Every Time.
        </h1>
        <p className="font-body mx-auto text-base leading-relaxed text-text-secondary md:max-w-lg md:text-body-lg lg:mx-0">
          A lightweight real-time platform that synchronizes speakers,
          moderators, and audiences around allocated presentation time through a
          shared public countdown.
        </p>
        <div className="flex flex-col gap-3 md:flex-row md:gap-4 lg:justify-start">
          <button className="w-full rounded-xl bg-gradient-to-r from-primary to-primary-light px-8 py-4 font-semibold text-sm text-white shadow-lg transition-all hover:shadow-xl hover:scale-[1.02] active:scale-95 md:w-auto md:text-base">
            Create Free Session
          </button>
          <button className="flex w-full items-center justify-center gap-2 rounded-xl border border-border bg-white px-8 py-4 font-semibold text-sm text-primary-dark shadow-sm transition-all hover:shadow-md hover:bg-surface-secondary active:scale-95 md:w-auto md:text-base">
            <span className="material-symbols-outlined">play_circle</span>
            Watch Demo
          </button>
        </div>
      </div>

      {/* Mockup - Using screen.png asset */}
      <div className="w-full flex-1 lg:max-w-2xl">
        <img
          src="/images/screen.png"
          alt="SyncTime Admin Control Panel showing live timer at 08:45"
          className="h-auto w-full rounded-2xl shadow-2xl"
        />
      </div>
    </section>
  );
}
