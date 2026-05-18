export default function Navigation() {
  return (
    <header className="fixed left-0 top-0 z-50 flex h-16 w-full items-center justify-between border-b border-border bg-white/80 px-4 backdrop-blur-xl md:px-10">
      <div className="flex items-center gap-2">
        <span className="material-symbols-outlined text-xl text-primary md:text-headline-md">
          sync
        </span>
        <span className="font-headline text-lg font-bold text-primary-dark md:text-headline-md">
          SyncTime
        </span>
      </div>
      <nav className="flex items-center gap-4 md:gap-8">
        <button className="rounded-full bg-primary px-4 py-2 font-label-md text-xs text-white shadow-sm transition-all hover:shadow-md hover:bg-primary-light active:scale-95 md:px-6 md:text-label-md">
          Create Free
        </button>
      </nav>
    </header>
  );
}
