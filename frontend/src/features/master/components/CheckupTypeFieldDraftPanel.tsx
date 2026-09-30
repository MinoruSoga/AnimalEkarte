import { C } from "@/lib/design-tokens";
import type { CheckupTypeFieldRow } from "@/hooks/use-checkup-fields";

import {
  isCheckupSelectFieldType,
  type CheckupFieldDraft,
} from "./checkup-type-fields-editor-model";
import {
  CheckupFieldOptionsEditor,
  CheckupFieldTypeSelect,
  NumberBoundInput,
} from "./CheckupTypeFieldEditors";
import { FieldInput } from "./ExamTypeFieldEditors";
import { useCheckupTypeFieldSession } from "../hooks/use-checkup-type-field-session";

// EMR-225: 健診フィールド定義の作成/編集パネル（ExamTypeFieldEditorSession 同型）。
// 型別フォーム出し分け: number→単位+基準値、選択式→選択肢エディタ、boolean/text→項目名+種別のみ。

interface CheckupTypeFieldDraftFormProps {
  editingId: string | "new";
  draft: CheckupFieldDraft;
  error: string;
  onPatch: (patch: Partial<CheckupFieldDraft>) => void;
  onAddOption: () => void;
  onChangeOption: (index: number, patch: Partial<{ value: string; label: string }>) => void;
  onRemoveOption: (index: number) => void;
  onCancel: () => void;
  onSave: () => void;
}

function CheckupTypeFieldDraftForm({
  editingId,
  draft,
  error,
  onPatch,
  onAddOption,
  onChangeOption,
  onRemoveOption,
  onCancel,
  onSave,
}: CheckupTypeFieldDraftFormProps) {
  const isNumber = draft.fieldType === "number";
  const isSelect = isCheckupSelectFieldType(draft.fieldType);
  return (
    <>
      <h4 className={`text-sm font-medium ${C.text}`}>
        {editingId === "new" ? "健診項目を追加" : "健診項目を編集"}
      </h4>
      <FieldInput
        label="項目名"
        value={draft.name}
        onChange={(value) => onPatch({ name: value })}
        description="健診項目の名前です。健診結果の入力画面に表示されます。"
      />
      <CheckupFieldTypeSelect
        value={draft.fieldType}
        onChange={(value) => onPatch({ fieldType: value })}
      />
      {isNumber ? (
        <>
          <FieldInput
            label="単位"
            value={draft.unit}
            onChange={(value) => onPatch({ unit: value })}
            description="この項目の単位です（例: kg、cm）。"
          />
          <div className="grid grid-cols-2 gap-2">
            <NumberBoundInput
              label="基準値下限"
              value={draft.minValue}
              onChange={(value) => onPatch({ minValue: value })}
              description="正常とみなす範囲の下限値です。"
            />
            <NumberBoundInput
              label="基準値上限"
              value={draft.maxValue}
              onChange={(value) => onPatch({ maxValue: value })}
              description="正常とみなす範囲の上限値です。"
            />
          </div>
        </>
      ) : null}
      {isSelect ? (
        <CheckupFieldOptionsEditor
          options={draft.options}
          onAdd={onAddOption}
          onChange={onChangeOption}
          onRemove={onRemoveOption}
        />
      ) : null}
      {error ? (
        <p role="alert" className={`text-sm ${C.danger}`}>
          {error}
        </p>
      ) : null}
      <div className="flex justify-end gap-2">
        <button
          type="button"
          onClick={onCancel}
          className={`min-h-11 rounded-xxs px-3 text-sm ${C.text50} ${C.hoverBgLight}`}
        >
          キャンセル
        </button>
        <button
          type="button"
          onClick={onSave}
          className={`min-h-11 rounded-full px-4 text-sm ${C.bgActionPrimarySolid} ${C.textOnActionPrimary} ${C.hoverBgActionPrimarySolid} ${C.hoverTextOnActionPrimary} ${C.activeBgActionPrimarySolid} ${C.activeTextOnActionPrimary}`}
        >
          項目を保存
        </button>
      </div>
    </>
  );
}

interface CheckupTypeFieldEditorSessionProps {
  checkupTypeId: string;
  editingId: string | "new";
  editingField: CheckupTypeFieldRow | null;
  fieldCount: number;
  canCreate: boolean;
  canEdit: boolean;
  onDirtyChange?: (dirty: boolean) => void;
  onClose: () => void;
}

export function CheckupTypeFieldEditorSession({
  checkupTypeId,
  editingId,
  editingField,
  fieldCount,
  canCreate,
  canEdit,
  onDirtyChange,
  onClose,
}: CheckupTypeFieldEditorSessionProps) {
  const session = useCheckupTypeFieldSession({
    checkupTypeId,
    editingId,
    editingField,
    fieldCount,
    canCreate,
    canEdit,
    onDirtyChange,
    onClose,
  });

  return (
    <div className={`mt-4 space-y-3 rounded-xs border p-3 ${C.borderLight} ${C.bgPage}`}>
      <CheckupTypeFieldDraftForm
        editingId={editingId}
        draft={session.draft}
        error={session.error}
        onPatch={session.patchDraft}
        onAddOption={session.addOption}
        onChangeOption={session.updateOption}
        onRemoveOption={session.removeOption}
        onCancel={session.cancelEdit}
        onSave={() => {
          void session.saveField();
        }}
      />
    </div>
  );
}
