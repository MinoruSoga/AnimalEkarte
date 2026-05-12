import { C } from "@/lib/design-tokens";
import type { CheckupAlertCategory } from "@/features/checkups";

interface CheckupAlertCardProps {
  variant: "overdue" | "upcoming";
  data: CheckupAlertCategory | undefined;
  isLoading: boolean;
  isError: boolean;
  onClick: () => void;
}

export function CheckupAlertCard({ variant, data, isLoading, isError, onClick }: CheckupAlertCardProps) {
  const isOverdue = variant === "overdue";
  const label = isOverdue ? "定期検査 期限切れ" : "定期検査 期限間近 (30日以内)";
  const testId = isOverdue ? "checkup-alert-overdue" : "checkup-alert-upcoming";

  const cardClass = isOverdue
    ? `${C.bgDanger8} ${C.borderDanger}`
    : `${C.bgNotice40} ${C.borderNotice}`;
  const countClass = isOverdue ? C.danger : C.textNotice;

  return (
    <button
      type="button"
      data-testid={testId}
      onClick={onClick}
      className={`${cardClass} rounded-lg border p-4 flex flex-col gap-1 text-left w-full cursor-pointer hover:opacity-90 transition-opacity`}
    >
      <span className={`text-sm font-medium ${C.text60}`}>{label}</span>
      {isLoading ? (
        <span className={`text-xl font-semibold ${C.text40}`} data-testid={`${testId}-count`}>
          --
        </span>
      ) : isError ? (
        <span className={`text-base ${C.text40}`} data-testid={`${testId}-count`}>
          取得失敗
        </span>
      ) : (
        <span className={`text-xl font-semibold ${countClass}`} data-testid={`${testId}-count`}>
          {data?.count ?? 0}件
        </span>
      )}
    </button>
  );
}
