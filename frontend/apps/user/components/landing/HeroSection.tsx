export default function HeroSection() {
  return (
    <section className="relative mx-auto flex max-w-container-max flex-col items-center gap-8 overflow-hidden px-4 py-12 md:gap-16 md:px-10 md:py-20 lg:flex-row">
      {/* Dot Grid Background */}
      <div className="pointer-events-none absolute inset-0 -z-10 opacity-[0.03]">
        <div
          className="h-full w-full"
          style={{
            backgroundImage:
              "radial-gradient(circle at 1px 1px, black 1px, transparent 0)",
            backgroundSize: "32px 32px",
          }}
        ></div>
      </div>

      {/* Content */}
      <div className="flex-1 space-y-6 text-center md:space-y-8 lg:text-left">
        <h1 className="font-display text-3xl text-primary md:max-w-xl md:text-display-lg">
          Keep Every Speaker On Time, Every Time.
        </h1>
        <p className="font-body mx-auto text-base text-on-surface-variant md:max-w-lg md:text-body-lg lg:mx-0">
          A lightweight real-time platform that synchronizes speakers,
          moderators, and audiences around allocated presentation time through a
          shared public countdown.
        </p>
        <div className="flex flex-col gap-3 md:flex-row md:gap-4 lg:justify-start">
          <button className="w-full rounded-xl bg-secondary px-8 py-4 font-headline text-sm text-on-secondary transition-all hover:opacity-90 active:scale-95 md:w-auto md:text-label-md">
            Create Free Session
          </button>
          <button className="flex w-full items-center justify-center gap-2 rounded-xl border border-outline px-8 py-4 font-headline text-sm text-primary transition-all hover:bg-surface-container active:scale-95 md:w-auto md:text-label-md">
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
