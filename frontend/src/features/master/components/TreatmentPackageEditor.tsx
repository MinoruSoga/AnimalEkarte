/**
 * EMR-523: パッケージ編集パネル。未保存入力は競合・失敗時も保持する。
 * version は編集開始時捕捉版のみを CAS に使い、競合は再読込・比較へ誘導（無断上書きしない）。
 */
import { C, STYLE } from "@/lib/design-tokens";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Checkbox } from "@/components/ui/checkbox";
import { TreatmentPackageComponentList } from "./TreatmentPackageComponentList";
import type { TreatmentPackageDraft, ComponentDraft } from "../lib/treatment-package-form";

interface TreatmentPackageEditorProps {
  mode: "create" | "edit";
  draft: TreatmentPackageDraft;
  /** componentKey → 行別エラー */
  componentErrors?: ReadonlyMap<string, readonly string[]>;
  formErrors?: readonly string[];
  /** 409 検出時 true。再読込・比較を促し、入力は破棄しない。 */
  conflict: boolean;
  /** create 応答不明（timeout/5xx）。盲目再送しない。 */
  createUnknown: boolean;
  canSave: boolean;
  canDelete: boolean;
  saving: boolean;
  deleting: boolean;
  onChange: (patch: Partial<TreatmentPackageDraft>) => void;
  onComponentsChange: (components: ComponentDraft[]) => void;
  onSave: () => void;
  onDeleteRequest?: () => void;
  onClose: () => void;
}

export function TreatmentPackageEditor({
  mode,
  draft,
  componentErrors,
  formErrors = [],
  conflict,
  createUnknown,
  canSave,
  canDelete,
  saving,
  deleting,
  onChange,
  onComponentsChange,
  onSave,
  onDeleteRequest,
  onClose,
}: TreatmentPackageEditorProps) {
  const busy = saving || deleting;
  return (
    <div className="flex flex-col gap-4 p-4">
      <h2 className={`text-base font-semibold ${C.text}`}>
        {mode === "create" ? "パッケージ新規登録" : "パッケージ編集"}
      </h2>
      {conflict ? (
        <div role="alert" className={`rounded-lg border ${C.borderLight} p-3 text-sm ${C.danger}`}>
          他の変更と競合しました（版不一致）。最新を再読込して内容を比較してください。入力は保持されています。
        </div>
      ) : null}
      {createUnknown ? (
        <div role="alert" className={`rounded-lg border ${C.borderLight} p-3 text-sm ${C.danger}`}>
          登録結果を確認できませんでした。同じ内容を自動で再送しません。一覧を再読込して結果を確認してください。入力は保持されています。
        </div>
      ) : null}
      {formErrors.length > 0 ? (
        <ul role="alert" className={`text-sm ${C.danger}`}>
          {formErrors.map((message) => (
            <li key={message}>{message}</li>
          ))}
        </ul>
      ) : null}
      <div>
        <label className={`text-xs ${C.text60}`} htmlFor="tp-name">
          名称（必須）
        </label>
        <Input
          id="tp-name"
          aria-label="名称"
          disabled={busy}
          value={draft.name}
          onChange={(e) => onChange({ name: e.target.value })}
        />
      </div>
      <div>
        <label className={`text-xs ${C.text60}`} htmlFor="tp-description">
          説明
        </label>
        <Textarea
          id="tp-description"
          aria-label="説明"
          disabled={busy}
          value={draft.description}
          onChange={(e) => onChange({ description: e.target.value })}
        />
      </div>
      <label className="flex items-center gap-2 text-sm">
        <Checkbox
          checked={draft.isActive}
          disabled={busy}
          onCheckedChange={(checked) => onChange({ isActive: checked === true })}
        />
        有効
      </label>
      <fieldset className="min-w-0">
        <legend className={`text-sm font-medium ${C.text}`}>構成品（順序どおり登録）</legend>
        <TreatmentPackageComponentList
          components={draft.components}
          disabled={busy || !canSave}
          componentErrors={componentErrors}
          onChange={onComponentsChange}
        />
      </fieldset>
      <div className={`${STYLE.FLEX_BETWEEN} gap-2`}>
        <div>
          {mode === "edit" && canDelete && onDeleteRequest !== undefined ? (
            <button
              type="button"
              disabled={busy}
              onClick={onDeleteRequest}
              className="text-sm underline"
            >
              削除
            </button>
          ) : null}
        </div>
        <div className="flex gap-2">
          <button type="button" disabled={busy} onClick={onClose} className="text-sm underline">
            キャンセル
          </button>
          <button
            type="button"
            disabled={busy || !canSave}
            onClick={onSave}
            aria-busy={saving}
            className="text-sm font-medium underline"
          >
            {saving ? "保存中…" : "保存"}
          </button>
        </div>
      </div>
    </div>
  );
}
