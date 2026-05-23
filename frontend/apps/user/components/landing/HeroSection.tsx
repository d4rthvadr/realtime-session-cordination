import { Button } from "@/components/ui/button";
import Image from "next/image";

export default function HeroSection() {
  return (
    <section className="relative mx-auto flex max-w-container-max flex-col items-center gap-8 overflow-hidden px-4 py-16 md:gap-16 md:px-10 md:py-28 lg:flex-row">
      {/* Gradient Background */}
      <div className="pointer-events-none absolute inset-0 -z-10">
        <div className="absolute inset-0 bg-gradient-to-br from-slate-50 via-white to-emerald-50/50"></div>
        <div className="absolute left-0 top-0 h-96 w-96 rounded-full bg-slate-200/60 blur-3xl"></div>
        <div className="absolute right-0 bottom-0 h-96 w-96 rounded-full bg-emerald-200/40 blur-3xl"></div>
      </div>

      {/* Content */}
      <div className="flex-1 space-y-6 text-center md:space-y-8 lg:text-left">
        <h1 className="font-display text-4xl font-bold text-slate-900 md:max-w-xl md:text-display-lg lg:text-6xl">
          Keep Every Speaker On Time, Every Time.
        </h1>
        <p className="font-body mx-auto text-base leading-relaxed text-slate-600 md:max-w-lg md:text-body-lg lg:mx-0">
          A lightweight real-time platform that synchronizes speakers,
          moderators, and audiences around allocated presentation time through a
          shared public countdown.
        </p>
        <div className="flex flex-col gap-3 md:flex-row md:gap-4 lg:justify-start">
          <Button className="h-11 w-full px-8 text-sm md:w-auto md:text-base">
            Create Free Session
          </Button>
          <Button
            variant="outline"
            className="flex h-11 w-full items-center justify-center gap-2 px-8 text-sm md:w-auto md:text-base"
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
          className="h-auto w-full rounded-3xl border border-slate-200 shadow-2xl"
        />
      </div>
    </section>
  );
}
