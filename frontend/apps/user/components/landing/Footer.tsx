export default function Footer() {
  return (
    <footer className="border-t border-outline-variant bg-surface-container-low py-12">
      <div className="mx-auto max-w-container-max px-10">
        <div className="mb-8 flex items-center justify-center gap-2">
          <span className="material-symbols-outlined text-primary">sync</span>
          <span className="font-headline font-bold text-headline-md text-primary">
            SyncTime
          </span>
        </div>
        <p className="text-center text-sm text-on-surface-variant">
          © 2024 SyncTime. Precision coordination for high-stakes sessions.
        </p>
      </div>
    </footer>
  );
}
