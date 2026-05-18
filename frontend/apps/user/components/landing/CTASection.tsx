export function CTASection() {
  return (
    <section className="relative overflow-hidden border-t border-border-light bg-gradient-to-br from-primary via-primary to-primary-light py-20 md:py-28 lg:py-32">
      {/* Decorative Elements */}
      <div className="pointer-events-none absolute inset-0">
        <div className="absolute left-0 top-0 h-72 w-72 rounded-full bg-white/10 blur-3xl"></div>
        <div className="absolute right-0 bottom-0 h-72 w-72 rounded-full bg-accent-cyan/20 blur-3xl"></div>
      </div>

      <div className="relative mx-auto max-w-4xl px-4 text-center md:px-10">
        <h2 className="font-headline text-3xl font-bold text-white md:text-5xl lg:text-6xl">
          Ready to Keep Your Presentations On Time?
        </h2>

        <div className="mt-8 md:mt-10">
          <button className="inline-flex items-center justify-center rounded-xl bg-white px-8 py-4 font-semibold text-base text-primary shadow-2xl transition-all hover:shadow-[0_20px_60px_rgba(255,255,255,0.3)] hover:scale-105 active:scale-95 md:px-10 md:py-5 md:text-lg">
            Create Your First Session — Free
          </button>
        </div>

        <p className="mt-6 text-sm text-white/80 md:mt-8 md:text-base">
          No signup required • Works instantly • Share in seconds
        </p>
      </div>
    </section>
  );
}
