import DashboardNav from "@/components/DashboardNav";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <>
      <DashboardNav />
      <main className="min-h-[calc(100vh-4rem)] bg-slate-50">{children}</main>
    </>
  );
}
