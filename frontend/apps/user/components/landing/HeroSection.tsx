export default function HeroSection() {
  return (
    <section className="relative mx-auto flex max-w-container-max flex-col items-center gap-16 overflow-hidden px-10 py-20 md:flex-row">
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

      {/* Left: Content */}
      <div className="flex-1 space-y-8">
        <h1 className="font-display max-w-xl text-display-lg text-primary">
          Keep Every Speaker On Time, Every Time.
        </h1>
        <p className="font-body-lg max-w-lg text-body-lg text-on-surface-variant">
          The real-time coordination platform for high-stakes sessions,
          keynotes, and broadcasts. Precision timing without the panic.
        </p>
        <div className="flex gap-4">
          <button className="rounded-xl bg-secondary px-8 py-4 font-headline text-label-md text-on-secondary transition-all hover:opacity-90 active:scale-95">
            Create Free Session
          </button>
          <button className="flex items-center gap-2 rounded-xl border border-outline px-8 py-4 font-headline text-label-md text-primary transition-all hover:bg-surface-container active:scale-95">
            <span className="material-symbols-outlined">play_circle</span>
            Watch Demo
          </button>
        </div>
      </div>

      {/* Right: Mockup */}
      <div className="flex w-full flex-1 flex-col overflow-hidden rounded-2xl border border-outline-variant bg-white shadow-lg">
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
    </section>
  );
}
