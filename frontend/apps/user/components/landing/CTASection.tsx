import { Button } from "@/components/ui/button";

export function CTASection() {
  return (
    <section className="relative overflow-hidden border-t border-slate-200 bg-gradient-to-br from-slate-900 via-slate-900 to-emerald-900 py-20 md:py-28 lg:py-32">
      {/* Decorative Elements */}
      <div className="pointer-events-none absolute inset-0">
        <div className="absolute left-0 top-0 h-72 w-72 rounded-full bg-white/10 blur-3xl"></div>
        <div className="absolute right-0 bottom-0 h-72 w-72 rounded-full bg-emerald-300/20 blur-3xl"></div>
      </div>

      <div className="relative mx-auto max-w-4xl px-4 text-center md:px-10">
        <h2 className="font-headline text-3xl font-bold text-white md:text-5xl lg:text-6xl">
          Ready to Keep Your Presentations On Time?
        </h2>

        <div className="mt-8 md:mt-10">
          <Button className="h-12 bg-white px-8 text-base text-slate-900 shadow-2xl hover:bg-slate-100 md:h-14 md:px-10 md:text-lg">
            Create Your First Session — Free
          </Button>
        </div>

        <p className="mt-6 text-sm text-white/80 md:mt-8 md:text-base">
          No signup required • Works instantly • Share in seconds
        </p>
      </div>
    </section>
  );
}
