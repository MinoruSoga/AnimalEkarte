/**
 * EMR-532/S08: 409 後の「最新取得 → 差分比較 → 明示選択」UI。
 * draft（あなたの入力）と最新 wire を並べて表示し、差分項目を強調する。
 * 自動保存・自動再送は行わない。選択は利用者のボタン操作のみ。
 */
import { C } from "@/lib/design-tokens";
import type { TreatmentPackage } from "../api/treatment-packages";
import {
  draftComponentSummary,
  draftPackageDiff,
  wireComponentSummary,
  type TreatmentPackageDraft,
} from "../lib/treatment-package-form";

interface TreatmentPackageConflictCompareProps {
  draft: TreatmentPackageDraft;
  latest: TreatmentPackage;
  /** 「最新を採用（入力を破棄）」 */
  onAdoptLatest: () => void;
  /** 「入力を保持して最新版を基準に編集を続ける」 */
  onKeepEditing: () => void;
}

function DiffCell({ value, differs }: { value: string; differs: boolean }) {
  return (
    <td className={`px-2 py-1 align-top ${differs ? `font-semibold ${C.danger}` : ""}`}>{value}</td>
  );
}

export function TreatmentPackageConflictCompare({
  draft,
  latest,
  onAdoptLatest,
  onKeepEditing,
}: TreatmentPackageConflictCompareProps) {
  const diff = draftPackageDiff(draft, latest);
  const yoursComponents = draft.components.map((c) => draftComponentSummary(c));
  const latestComponents = latest.components.map((c) => wireComponentSummary(c));
  const componentsRows = Math.max(yoursComponents.length, latestComponents.length);
  return (
    <div role="alert" className={`rounded-lg border ${C.borderLight} p-3 text-sm`}>
      <p className={`font-medium ${C.danger}`}>
        他の変更と競合しました（版不一致）。最新と入力を比較し、どちらを使うか選んでください。
      </p>
      <table className="mt-2 w-full text-xs" aria-label="競合比較">
        <thead>
          <tr className={`text-left ${C.text60}`}>
            <th className="px-2 py-1 font-medium">項目</th>
            <th className="px-2 py-1 font-medium">あなたの入力</th>
            <th className="px-2 py-1 font-medium">最新</th>
          </tr>
        </thead>
        <tbody>
          <tr className={`border-t ${C.borderLight}`}>
            <td className="px-2 py-1">名称</td>
            <DiffCell value={draft.name} differs={diff.name} />
            <DiffCell value={latest.name} differs={diff.name} />
          </tr>
          <tr className={`border-t ${C.borderLight}`}>
            <td className="px-2 py-1">説明</td>
            <DiffCell value={draft.description} differs={diff.description} />
            <DiffCell value={latest.description} differs={diff.description} />
          </tr>
          <tr className={`border-t ${C.borderLight}`}>
            <td className="px-2 py-1">有効</td>
            <DiffCell value={draft.isActive ? "有効" : "無効"} differs={diff.isActive} />
            <DiffCell value={latest.is_active ? "有効" : "無効"} differs={diff.isActive} />
          </tr>
          <tr className={`border-t ${C.borderLight}`}>
            <td className="px-2 py-1 align-top">
              構成（
              {diff.components
                ? `${yoursComponents.length}件 / ${latestComponents.length}件`
                : `${yoursComponents.length}件`}
              ）
            </td>
            <td className={`px-2 py-1 ${diff.components ? `font-semibold ${C.danger}` : ""}`}>
              <ul>
                {Array.from({ length: componentsRows }, (_, i) => (
                  <li key={i}>{yoursComponents[i] ?? "—"}</li>
                ))}
              </ul>
            </td>
            <td className={`px-2 py-1 ${diff.components ? `font-semibold ${C.danger}` : ""}`}>
              <ul>
                {Array.from({ length: componentsRows }, (_, i) => (
                  <li key={i}>{latestComponents[i] ?? "—"}</li>
                ))}
              </ul>
            </td>
          </tr>
        </tbody>
      </table>
      <div className="mt-3 flex flex-col gap-2">
        <button
          type="button"
          className={`min-h-[44px] rounded border ${C.borderLight} px-3 text-sm font-medium`}
          onClick={onAdoptLatest}
        >
          最新を採用（入力を破棄）
        </button>
        <button
          type="button"
          className={`min-h-[44px] rounded border ${C.borderLight} px-3 text-sm font-medium`}
          onClick={onKeepEditing}
        >
          入力を保持して最新版を基準に編集を続ける
        </button>
      </div>
    </div>
  );
}
