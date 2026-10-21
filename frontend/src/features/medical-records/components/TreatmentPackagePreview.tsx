/**
 * EMR-524: preview の行別表示（契約mock段階）。
 * blocking/warning を分離し、財務 unknown は警告・owner未解決は blocking。
 * total null を 0 にせず、税込額を再計算しない。
 */
import { C } from "@/lib/design-tokens";
import type { TreatmentPackagePreview as Preview } from "../api/treatment-package-applications";
import { classifyPreview } from "../lib/treatment-package-operation-state";

const KIND_LABELS: Record<string, string> = {
  consultation: "診察",
  procedure: "処置",
  medicine: "薬剤",
  examination_type: "検査",
  vaccine: "予防接種",
  checkup_type: "定期健診",
  merchandise_item: "物販",
  manual_other: "手入力",
};

export function TreatmentPackagePreview({
  preview,
  canApply,
  applyDisabledReason,
  applying,
  onApply,
  onClose,
}: {
  preview: Preview;
  canApply: boolean;
  applyDisabledReason?: string;
  applying: boolean;
  onApply: () => void;
  onClose: () => void;
}) {
  const { blocking, warnings } = classifyPreview(preview);
  return (
    <div className="flex flex-col gap-3" aria-label="パッケージプレビュー">
      <ul className="flex flex-col gap-2">
        {preview.items.map((item) => {
          const itemIssues = item.issues;
          return (
            <li key={item.component_key} className={`rounded-lg border ${C.borderLight} p-2`}>
              <div className="flex items-baseline justify-between gap-2">
                <span className="text-sm font-medium">{item.name}</span>
                <span className={`text-xs ${C.text60}`}>
                  {KIND_LABELS[item.item.kind] ?? item.item.kind} × {item.quantity}
                </span>
              </div>
              <div className={`text-xs ${C.text60}`}>
                {item.classification === "clinical"
                  ? "未実施として追加"
                  : item.classification === "commercial"
                    ? "選択明細として追加"
                    : "分類未解決"}
                {item.price.unit_price === null ? " / 価格未確定" : ` / ¥${item.price.unit_price}`}
              </div>
              {itemIssues.length > 0 ? (
                <ul className="mt-1 text-xs" role="list">
                  {itemIssues.map((issue) => (
                    <li
                      key={`${issue.code}-${issue.required_at}`}
                      className={issue.severity === "blocking" ? C.danger : C.text60}
                      data-severity={issue.severity}
                    >
                      {issue.severity === "blocking" ? "追加不可: " : "警告: "}
                      {issue.message}
                    </li>
                  ))}
                </ul>
              ) : null}
            </li>
          );
        })}
      </ul>
      <div className={`text-sm ${C.text}`} aria-label="合計">
        {preview.total.status === "available" && preview.total.amount !== null
          ? `合計 ¥${preview.total.amount}`
          : "合計は未確定です（各品目の価格を確認してください）"}
      </div>
      {blocking.length > 0 ? (
        <p role="alert" className={`text-sm ${C.danger}`}>
          {blocking.length} 件の構成品が追加できません。
        </p>
      ) : null}
      {warnings.length > 0 && blocking.length === 0 ? (
        <p className={`text-sm ${C.text60}`}>
          {warnings.length} 件の財務情報が未確定です。警告を確認のうえ適用できます。
        </p>
      ) : null}
      <div className="flex justify-end gap-3">
        <button type="button" className="text-sm underline" onClick={onClose}>
          戻る
        </button>
        <button
          type="button"
          className="text-sm font-medium underline"
          disabled={!canApply || applying}
          aria-busy={applying}
          title={applyDisabledReason}
          onClick={onApply}
        >
          {applying ? "適用中…" : "この内容で追加"}
        </button>
      </div>
      {applyDisabledReason !== undefined && !canApply ? (
        <p className={`text-xs ${C.text60}`}>{applyDisabledReason}</p>
      ) : null}
    </div>
  );
}
