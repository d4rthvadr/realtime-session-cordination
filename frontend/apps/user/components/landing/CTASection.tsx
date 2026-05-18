export function CTASection() {
  return (
    <section className="border-t border-outline-variant bg-primary py-16 sm:py-24">
      <div className="mx-auto max-w-4xl px-10 text-center">
        <h2 className="font-headline text-headline-lg text-on-primary sm:text-display-lg">
          Ready to Keep Your Presentations On Time?
        </h2>

        <div className="mt-8">
          <button className="inline-flex items-center justify-center rounded-xl bg-secondary px-8 py-4 font-headline text-lg text-on-secondary transition-all hover:opacity-90 active:scale-95">
            Create Your First Session — Free
          </button>
        </div>

        <p className="mt-6 text-sm text-on-primary/70">
          No signup required • Works instantly • Share in seconds
        </p>
      </div>
    </section>
  );
}
