export default function Navigation() {
  return (
    <header className="fixed left-0 top-0 z-50 flex h-16 w-full items-center justify-between border-b border-outline-variant bg-surface/80 px-10 backdrop-blur-md">
      <div className="flex items-center gap-2">
        <span className="material-symbols-outlined text-headline-md text-primary">
          sync
        </span>
        <span className="font-headline font-bold text-headline-md text-primary">
          SyncTime
        </span>
      </div>
      <nav className="hidden items-center gap-8 md:flex">
        {/* <a
          className="font-label-md text-label-md text-primary transition-transform hover:opacity-80 active:scale-95"
          href="#"
        >
          Live
        </a>
        <a
          className="font-label-md text-label-md text-on-surface-variant transition-transform hover:opacity-80 active:scale-95"
          href="#"
        >
          Agenda
        </a>
        <a
          className="font-label-md text-label-md text-on-surface-variant transition-transform hover:opacity-80 active:scale-95"
          href="#"
        >
          People
        </a> */}
        <button className="rounded-full bg-primary px-6 py-2 font-label-md text-label-md text-on-primary transition-transform hover:opacity-80 active:scale-95">
          Create Free
        </button>
      </nav>
    </header>
  );
}
