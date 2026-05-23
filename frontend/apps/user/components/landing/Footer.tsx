export default function Footer() {
  return (
    <footer className="border-t border-slate-800 bg-[#090b10] py-12 md:py-16">
      <div className="mx-auto max-w-container-max px-4 md:px-10">
        <div className="mb-6 flex items-center justify-center gap-2 md:mb-8">
          <span className="material-symbols-outlined text-xl text-slate-200 md:text-2xl">
            sync
          </span>
          <span className="font-headline text-xl font-bold text-slate-100 md:text-2xl">
            SyncTime
          </span>
        </div>
        <p className="text-center text-sm text-slate-500 md:text-base">
          © 2024 SyncTime. Precision coordination for high-stakes sessions.
        </p>
      </div>
    </footer>
  );
}
