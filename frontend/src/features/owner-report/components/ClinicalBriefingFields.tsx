import type { ReactNode } from "react";

import { C } from "@/lib/design-tokens";

interface DataStatusProps {
  children?: ReactNode;
  isError?: boolean;
  isLoading?: boolean;
  noPermission?: boolean;
  emptyMessage?: string;
}

export function DataStatus({
  children,
  isError,
  isLoading,
  noPermission,
  emptyMessage = "記録なし",
}: DataStatusProps) {
  if (noPermission) return <p className={`text-xs ${C.text50}`}>閲覧権限なし</p>;
  if (isLoading)
    return (
      <p className={`text-xs ${C.text50}`} role="status">
        読み込み中...
      </p>
    );
  if (isError)
    return (
      <p className={`text-xs ${C.danger}`} role="alert">
        取得失敗
      </p>
    );
  return children ? <>{children}</> : <p className={`text-xs ${C.text50}`}>{emptyMessage}</p>;
}

interface BriefingFieldProps {
  label: string;
  value: ReactNode;
  alert?: boolean;
}

export function BriefingField({ label, value, alert = false }: BriefingFieldProps) {
  return (
    <div className={`min-w-0 border-l-2 pl-1.5 ${alert ? C.borderDanger : C.borderBrand}`}>
      <span className={`block text-2xs leading-snug font-semibold ${C.text50}`}>{label}</span>
      <strong
        className={`mt-0.5 block break-words text-sm leading-snug font-medium ${alert ? C.danger : C.text}`}
      >
        {value}
      </strong>
    </div>
  );
}

interface DetailFieldProps {
  label: string;
  value?: ReactNode;
}

export function DetailField({ label, value }: DetailFieldProps) {
  return (
    <div
      className={`grid min-w-0 grid-cols-[6rem_minmax(0,1fr)] gap-1 border-b py-1 ${C.borderDivider}`}
    >
      <dt className={`text-xs ${C.text50}`}>{label}</dt>
      <dd className={`min-w-0 break-words text-xs font-medium ${C.text}`}>{value || "-"}</dd>
    </div>
  );
}
