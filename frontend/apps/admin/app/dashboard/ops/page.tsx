import { getAnalyticsOpsStatus, getDLQList } from "@/lib/actions";
import OpsHealthCard from "@/components/widgets/OpsHealthCard";
import DLQSection from "@/components/widgets/DLQSection";

export default async function OpsPage() {
  const [opsResult, dlqResult] = await Promise.all([
    getAnalyticsOpsStatus(),
    getDLQList(50, 0),
  ]);

  return (
    <div className="min-h-screen bg-slate-50 px-4 py-8 sm:px-6">
      <div className="mx-auto max-w-5xl space-y-6">
        <div>
          <h1 className="text-xl font-semibold text-slate-900">Operations</h1>
          <p className="text-sm text-slate-500 mt-0.5">
            Analytics processor health and dead-letter queue management.
          </p>
        </div>

        <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
          <div className="lg:col-span-1">
            <OpsHealthCard
              freshness={opsResult.freshness ?? null}
              metrics={opsResult.metrics ?? null}
            />
          </div>

          <div className="lg:col-span-2">
            <DLQSection
              initialRows={dlqResult.rows}
              initialCount={dlqResult.count}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
