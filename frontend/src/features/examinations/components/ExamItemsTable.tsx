// React/Framework
import { memo, useCallback } from "react";

// External
import { CheckCircle } from "lucide-react";

// Internal
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { C, ICON } from "@/lib/design-tokens";

// 検査項目テーブルの 1 行。
// status / isAbnormal は backend が ref_min/ref_max から導出した値を表示するだけ（FE で再計算しない）。
export interface ExamItemRow {
  /** クライアント側の React key 用ローカル ID（不変） */
  key: string;
  /** exam_type_fields.id（テンプレ由来。手動追加行では undefined） */
  examTypeFieldId?: number;
  name: string;
  inspectionValue: string;
  unit: string;
  normalValue: string;
  referenceValue: string;
  refMin?: number;
  refMax?: number;
  sortOrder: number;
  /** backend が導出した判定。新規行・未保存値では undefined */
  status?: "normal" | "high" | "low";
  isAbnormal?: boolean;
}

interface ExamItemsTableProps {
  items: ExamItemRow[];
  onChangeInspectionValue: (key: string, value: string) => void;
  disabled?: boolean;
}

function StatusBadge({ status }: { status?: "normal" | "high" | "low" }) {
  if (status === "high") {
    return (
      <Badge
        variant="destructive"
        className={`h-8 px-3 text-xs ${C.bgDanger} ${C.hoverBgDanger90}`}
      >
        HIGH
      </Badge>
    );
  }
  if (status === "low") {
    return (
      <Badge
        variant="outline"
        className={`h-8 px-3 text-xs ${C.textBrand} ${C.borderBrand} ${C.bgBrand5}`}
      >
        LOW
      </Badge>
    );
  }
  if (status === "normal") {
    return <CheckCircle className={`${ICON.action} ${C.textStatusGreen} opacity-50`} />;
  }
  // 未判定（保存前）
  return <span className={`text-xs ${C.text45}`}>-</span>;
}

/**
 * 検査項目テーブル — `ExaminationForm` に組み込む編集可能テーブル。
 *
 * 表示列: 項目名 / 結果値（編集可） / 単位 / 基準値 / 判定（backend 由来）
 * - 異常値判定は backend service 層で行われる。FE はサーバが返した status を表示するだけ。
 * - 行の追加/削除はテンプレベース運用とするため MVP では非対応。
 * - 確定済み (`disabled=true`) では input が disabled になる。
 */
export const ExamItemsTable = memo(function ExamItemsTable({
  items,
  onChangeInspectionValue,
  disabled = false,
}: ExamItemsTableProps) {
  const handleChange = useCallback(
    (key: string) => (e: React.ChangeEvent<HTMLInputElement>) => {
      onChangeInspectionValue(key, e.target.value);
    },
    [onChangeInspectionValue],
  );

  if (items.length === 0) {
    return (
      <div className={`p-4 rounded-lg border ${C.borderMedium} ${C.bgWhite}`}>
        <p className={`text-sm ${C.text45} text-center`}>
          検査種別を選択すると検査項目が表示されます
        </p>
      </div>
    );
  }

  return (
    <div className={`border ${C.borderMedium} rounded-lg ${C.bgWhite} overflow-hidden shadow-sm overflow-x-auto`}>
      {/* ヘッダー */}
      <div
        className={`min-w-[600px] grid grid-cols-[2fr_1.5fr_1fr_1.8fr_1.2fr] gap-0 border-b ${C.borderMedium} ${C.bgPage} text-sm font-bold ${C.text80} h-11 items-center`}
      >
        <div className={`p-2 border-r ${C.borderMedium} pl-3`}>項目名</div>
        <div className={`p-2 border-r ${C.borderMedium} text-right`}>結果値</div>
        <div className={`p-2 border-r ${C.borderMedium} text-center`}>単位</div>
        <div className={`p-2 border-r ${C.borderMedium} text-center`}>基準値</div>
        <div className="p-2 text-center">判定</div>
      </div>
      {items.map((item, idx) => (
        <div
          key={item.key}
          data-testid="exam-item-row"
          data-abnormal={String(!!item.isAbnormal)}
          className={`min-w-[600px] grid grid-cols-[2fr_1.5fr_1fr_1.8fr_1.2fr] gap-0 ${
            idx !== items.length - 1 ? `border-b ${C.borderMedium}` : ""
          } ${
            item.isAbnormal
              ? item.status === "high"
                ? C.bgDanger8
                : C.bgBrandLight8
              : C.bgWhite
          } text-sm ${C.text} items-center h-12`}
        >
          <div className={`p-2 border-r ${C.borderMedium} pl-3 font-medium`}>
            {item.name}
          </div>
          <div className={`p-1.5 border-r ${C.borderMedium}`}>
            <Input
              type="text"
              inputMode="decimal"
              value={item.inspectionValue}
              disabled={disabled}
              onChange={handleChange(item.key)}
              placeholder="-"
              className={`h-8 text-sm text-right font-mono ${C.bgWhite} ${C.borderMedium}`}
              aria-label={`${item.name}の結果値`}
            />
          </div>
          <div className={`p-2 border-r ${C.borderMedium} text-center ${C.text60} text-sm`}>
            {item.unit || "-"}
          </div>
          <div className={`p-2 border-r ${C.borderMedium} text-center ${C.text60} text-sm`}>
            {item.referenceValue || item.normalValue || "-"}
          </div>
          <div className="p-2 flex justify-center items-center">
            <StatusBadge status={item.status} />
          </div>
        </div>
      ))}
    </div>
  );
});
