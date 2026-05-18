export default function Footer() {
  return (
    <footer className="border-t border-outline-variant bg-surface-container-low py-8 md:py-12">
      <div className="mx-auto max-w-container-max px-4 md:px-10">
        <div className="mb-4 flex items-center justify-center gap-2 md:mb-8">
          <span className="material-symbols-outlined text-lg text-primary md:text-xl">
            sync
          </span>
          <span className="font-headline text-lg font-bold text-primary md:text-headline-md">
            SyncTime
          </span>
        </div>
        <p className="text-center text-xs text-on-surface-variant md:text-sm">
          © 2024 SyncTime. Precision coordination for high-stakes sessions.
        </p>
      </div>
    </footer>
  );
}
