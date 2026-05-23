import { Button } from "@/components/ui/button";
import Image from "next/image";

export default function HeroSection() {
  return (
    <section className="relative mx-auto flex max-w-container-max flex-col items-center gap-8 overflow-hidden px-4 py-16 md:gap-16 md:px-10 md:py-28 lg:flex-row">
      {/* Gradient Background */}
      <div className="pointer-events-none absolute inset-0 -z-10">
        <div className="absolute inset-0 bg-gradient-to-br from-slate-950/70 via-[#0a0b0f] to-slate-900/40"></div>
        <div className="absolute left-0 top-0 h-96 w-96 rounded-full bg-indigo-500/10 blur-3xl"></div>
        <div className="absolute right-0 bottom-0 h-96 w-96 rounded-full bg-emerald-400/10 blur-3xl"></div>
      </div>

      {/* Content */}
      <div className="flex-1 space-y-6 text-center md:space-y-8 lg:text-left">
        <h1 className="font-display text-4xl font-bold text-slate-100 md:max-w-xl md:text-display-lg lg:text-6xl">
          Keep Every Speaker On Time, Every Time.
        </h1>
        <p className="font-body mx-auto text-base leading-relaxed text-slate-400 md:max-w-lg md:text-body-lg lg:mx-0">
          A lightweight real-time platform that synchronizes speakers,
          moderators, and audiences around allocated presentation time through a
          shared public countdown.
        </p>
        <div className="flex flex-col gap-3 md:flex-row md:gap-4 lg:justify-start">
          <Button className="h-11 w-full bg-slate-100 px-8 text-sm text-slate-900 hover:bg-white md:w-auto md:text-base">
            Create Free Session
          </Button>
          <Button
            variant="outline"
            className="flex h-11 w-full items-center justify-center gap-2 border-slate-700 bg-slate-900/40 px-8 text-sm text-slate-200 hover:bg-slate-800 md:w-auto md:text-base"
          >
            <span className="material-symbols-outlined">play_circle</span>
            Watch Demo
          </Button>
        </div>
      </div>

      {/* Mockup - Using screen.png asset */}
      <div className="w-full flex-1 lg:max-w-2xl">
        <Image
          src="/images/screen.png"
          alt="SyncTime Admin Control Panel showing live timer at 08:45"
          width={1280}
          height={720}
          className="h-auto w-full rounded-3xl border border-slate-800 bg-slate-950/50 shadow-[0_0_0_1px_rgba(255,255,255,0.03),0_24px_80px_rgba(0,0,0,0.45)]"
        />
      </div>
    </section>
  );
}
