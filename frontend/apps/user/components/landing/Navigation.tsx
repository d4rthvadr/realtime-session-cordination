import { Button } from "@/components/ui/button";

export default function Navigation() {
  return (
    <header className="fixed left-0 top-0 z-50 flex h-16 w-full items-center justify-between border-b border-slate-800 bg-[#0b0d12]/85 px-4 backdrop-blur-xl md:px-10">
      <div className="flex items-center gap-2">
        <span className="material-symbols-outlined text-xl text-slate-200 md:text-headline-md">
          sync
        </span>
        <span className="font-headline text-lg font-bold text-slate-100 md:text-headline-md">
          SyncTime
        </span>
      </div>
      <nav className="flex items-center gap-4">
        <Button className="h-9 bg-slate-100 px-5 text-xs text-slate-900 hover:bg-white md:h-10 md:px-6 md:text-sm">
          Create Free
        </Button>
      </nav>
    </header>
  );
}
