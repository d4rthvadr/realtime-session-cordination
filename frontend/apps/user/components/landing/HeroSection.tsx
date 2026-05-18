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

      {/* Mockup - Desktop Version */}
      <div className="hidden w-full flex-1 flex-col overflow-hidden rounded-2xl border border-outline-variant bg-white shadow-lg lg:flex">
        {/* Window Header */}
        <div className="flex items-center justify-between border-b border-outline-variant bg-surface-container-high p-4">
          <div className="flex gap-2">
            <div className="h-3 w-3 rounded-full bg-red-500"></div>
            <div className="h-3 w-3 rounded-full bg-yellow-500"></div>
            <div className="h-3 w-3 rounded-full bg-green-500"></div>
          </div>
          <span className="font-label-md text-label-md text-on-surface-variant">
            Admin Control Panel
          </span>
        </div>

        {/* Split View */}
        <div className="flex h-[400px]">
          {/* Left: Controls */}
          <div className="flex w-1/3 flex-col gap-6 border-r border-outline-variant p-6">
            <div className="space-y-2">
              <label className="font-label-md text-label-md uppercase text-on-surface-variant">
                Session Status
              </label>
              <div className="flex w-fit items-center gap-2 rounded-full border border-green-600 bg-green-50 px-3 py-1">
                <span className="h-2 w-2 animate-pulse rounded-full bg-green-600"></span>
                <span className="text-xs font-bold text-green-700">LIVE</span>
              </div>
            </div>
            <div className="flex flex-col gap-3">
              <button className="flex w-full items-center justify-center gap-2 rounded-lg bg-secondary py-3 font-label-md text-white">
                <span className="material-symbols-outlined text-sm">pause</span>
                PAUSE
              </button>
              <button className="flex w-full items-center justify-center gap-2 rounded-lg border border-outline py-3 font-label-md">
                <span className="material-symbols-outlined text-sm">add</span>
                +1 MIN
              </button>
            </div>
          </div>

          {/* Right: Display */}
          <div className="relative flex w-2/3 flex-col items-center justify-center overflow-hidden bg-slate-900 text-white">
            <div className="pointer-events-none absolute inset-0 opacity-10">
              <div
                className="h-full w-full"
                style={{
                  backgroundImage:
                    "radial-gradient(circle at 2px 2px, white 1px, transparent 0)",
                  backgroundSize: "24px 24px",
                }}
              ></div>
            </div>
            <span className="font-label-md mb-4 text-label-md uppercase tracking-widest text-slate-400">
              Current Session: Opening Keynote
            </span>
            <div className="font-display text-timer-display">08:45</div>
            <div className="mt-8 flex gap-4">
              <span className="rounded bg-white/10 px-4 py-1 font-label-md text-sm">
                UP NEXT: CTO Remarks
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Mockup - Mobile Version */}
      <div className="w-full flex-col overflow-hidden rounded-2xl border border-outline-variant bg-white shadow-lg lg:hidden">
        {/* Admin Control Header */}
        <div className="flex items-center justify-between border-b border-outline-variant bg-surface-container-low px-4 py-3">
          <span className="font-label-md text-xs font-semibold uppercase tracking-wider text-on-surface-variant">
            Admin Control
          </span>
          <div className="flex items-center gap-1.5 rounded-full bg-red-50 px-2.5 py-1">
            <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-red-500"></span>
            <span className="text-[10px] font-bold text-red-600">LIVE</span>
          </div>
        </div>

        {/* Control Buttons */}
        <div className="grid grid-cols-3 gap-2 border-b border-outline-variant p-4">
          <button className="flex flex-col items-center gap-1.5 rounded-lg border border-outline-variant bg-surface py-3 transition-colors hover:bg-surface-container">
            <span className="material-symbols-outlined text-lg text-on-surface-variant">
              play_arrow
            </span>
            <span className="text-[10px] font-semibold uppercase tracking-wider text-on-surface-variant">
              Start
            </span>
          </button>
          <button className="flex flex-col items-center gap-1.5 rounded-lg border border-outline-variant bg-surface py-3 transition-colors hover:bg-surface-container">
            <span className="material-symbols-outlined text-lg text-on-surface-variant">
              pause
            </span>
            <span className="text-[10px] font-semibold uppercase tracking-wider text-on-surface-variant">
              Pause
            </span>
          </button>
          <button className="flex flex-col items-center gap-1.5 rounded-lg border border-outline-variant bg-surface py-3 transition-colors hover:bg-surface-container">
            <span className="material-symbols-outlined text-lg text-on-surface-variant">
              tune
            </span>
            <span className="text-[10px] font-semibold uppercase tracking-wider text-on-surface-variant">
              Adjust
            </span>
          </button>
        </div>

        {/* Timer Display */}
        <div className="flex flex-col items-center bg-white px-4 py-8">
          <span className="mb-2 text-xs text-on-surface-variant">
            Keynote Presentation
          </span>
          <div className="font-display text-6xl font-bold text-primary">
            08:45
          </div>
          {/* Progress Bar */}
          <div className="mt-6 w-full">
            <div className="h-1 w-full overflow-hidden rounded-full bg-surface-container">
              <div className="h-full w-2/3 bg-secondary"></div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
