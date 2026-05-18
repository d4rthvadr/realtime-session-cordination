export function CTASection() {
  return (
    <section className="border-t border-outline-variant bg-primary py-12 md:py-16 lg:py-24">
      <div className="mx-auto max-w-4xl px-4 text-center md:px-10">
        <h2 className="font-headline text-2xl text-on-primary md:text-headline-lg lg:text-display-lg">
          Ready to Keep Your Presentations On Time?
        </h2>

        <div className="mt-6 md:mt-8">
          <button className="inline-flex items-center justify-center rounded-xl bg-secondary px-6 py-3 font-headline text-base text-on-secondary transition-all hover:opacity-90 active:scale-95 md:px-8 md:py-4 md:text-lg">
            Create Your First Session — Free
          </button>
        </div>

        <p className="mt-4 text-xs text-on-primary/70 md:mt-6 md:text-sm">
          No signup required • Works instantly • Share in seconds
        </p>
      </div>
    </section>
  );
}
