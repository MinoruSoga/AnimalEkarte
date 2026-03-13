// External
import Users from "lucide-react/dist/esm/icons/users";
import Clock from "lucide-react/dist/esm/icons/clock";
import CheckCircle from "lucide-react/dist/esm/icons/check-circle";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";

// Internal
import { C } from "@/lib/design-tokens";

// Types
import type { ColumnData } from "@/types";

interface SummaryCardProps {
  icon: React.ReactNode;
  label: string;
  count: number;
  colorClass: string;
  bgClass: string;
  borderClass: string;
  isLoading: boolean;
}

function SummaryCard({
  icon,
  label,
  count,
  colorClass,
  bgClass,
  borderClass,
  isLoading,
}: SummaryCardProps) {
  return (
    <div
      className={`flex items-center gap-3 px-4 py-3 rounded-lg border ${bgClass} ${borderClass} min-w-[140px] flex-1`}
    >
      <div className={`flex-shrink-0 ${colorClass}`}>{icon}</div>
      <div className="min-w-0">
        <p className={`text-xs font-medium ${C.text60} leading-tight`}>{label}</p>
        {isLoading ? (
          <div className={`mt-1 h-5 w-8 rounded ${C.bgSkeleton} animate-pulse`} />
        ) : (
          <p className={`text-xl font-bold ${colorClass} leading-tight tabular-nums`}>
            {count}
            <span className={`text-xs font-normal ${C.text60} ml-0.5`}>件</span>
          </p>
        )}
      </div>
    </div>
  );
}

interface DashboardSummaryProps {
  columns: ColumnData[];
  isLoading: boolean;
}

/** ダッシュボード上部の統計サマリーカード */
export function DashboardSummary({ columns, isLoading }: DashboardSummaryProps) {
  // 列名をキーにした件数計算
  const countByTitle = (title: string): number =>
    columns.find((c) => c.title === title)?.appointments.length ?? 0;

  const total = columns.reduce((sum, col) => sum + col.appointments.length, 0);
  const waiting = countByTitle("受付予約") + countByTitle("受付済");
  const inConsultation = countByTitle("診療中");
  const completed = countByTitle("会計済");

  return (
    <div className="flex flex-wrap gap-2 px-5 pt-4 pb-0">
      {/* 本日の予約総数 */}
      <SummaryCard
        icon={<Users className="size-4" />}
        label="本日の予約"
        count={total}
        colorClass={C.text}
        bgClass="bg-white"
        borderClass={`border-[rgba(55,53,47,0.12)]`}
        isLoading={isLoading}
      />

      {/* 待機中（受付予約 + 受付済） */}
      <SummaryCard
        icon={<Clock className="size-4" />}
        label="待機中"
        count={waiting}
        colorClass={C.textAccentDark}
        bgClass={C.bgAccentLight50}
        borderClass="border-[#B8D4E3]/60"
        isLoading={isLoading}
      />

      {/* 診察中 */}
      <SummaryCard
        icon={<Stethoscope className="size-4" />}
        label="診察中"
        count={inConsultation}
        colorClass={C.textStatusPurple}
        bgClass={C.bgStatusPurple60}
        borderClass="border-[#D6C6E1]/60"
        isLoading={isLoading}
      />

      {/* 完了（会計済） */}
      <SummaryCard
        icon={<CheckCircle className="size-4" />}
        label="完了"
        count={completed}
        colorClass={C.textStatusGreen}
        bgClass={C.bgStatusGreen60}
        borderClass="border-[#C3DFC3]/60"
        isLoading={isLoading}
      />
    </div>
  );
}
