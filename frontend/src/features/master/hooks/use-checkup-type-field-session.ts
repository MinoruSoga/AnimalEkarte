import { useCallback, useEffect, useState } from "react";

import type { CheckupTypeFieldRow } from "@/hooks/use-checkup-fields";

import { useCreateCheckupTypeField, useUpdateCheckupTypeField } from "../api/checkup-type-fields";
import {
  buildCheckupFieldCreateRequest,
  buildCheckupFieldUpdateRequest,
  checkupFieldToDraft,
  emptyCheckupFieldDraft,
  validateCheckupFieldDraft,
  type CheckupFieldDraft,
  type CheckupFieldOptionDraft,
} from "../components/checkup-type-fields-editor-model";

// EMR-225: 健診フィールド定義の編集セッション（exam-type の useExamTypeFieldSession 同型）。
// draft は name/field_type/unit/min/max/options の全体を保持し、保存時に
// 作成なら create 用・更新なら update 用（clear_* フラグ込み）の request を組み立てる。

interface UseCheckupTypeFieldSessionArgs {
  checkupTypeId: string;
  editingId: string | "new";
  editingField: CheckupTypeFieldRow | null;
  /** 新規作成時の sort_order 末尾採番に使う現在のフィールド件数 */
  fieldCount: number;
  canCreate: boolean;
  canEdit: boolean;
  onDirtyChange?: (dirty: boolean) => void;
  onClose: () => void;
}

export function useCheckupTypeFieldSession({
  checkupTypeId,
  editingId,
  editingField,
  fieldCount,
  canCreate,
  canEdit,
  onDirtyChange,
  onClose,
}: UseCheckupTypeFieldSessionArgs) {
  const createField = useCreateCheckupTypeField();
  const updateField = useUpdateCheckupTypeField();
  const [draft, setDraft] = useState<CheckupFieldDraft>(() =>
    editingField ? checkupFieldToDraft(editingField) : emptyCheckupFieldDraft(),
  );
  const [error, setError] = useState("");
  const [dirty, setDirty] = useState(false);

  useEffect(() => {
    onDirtyChange?.(dirty);
  }, [dirty, onDirtyChange]);

  useEffect(() => {
    return () => onDirtyChange?.(false);
  }, [onDirtyChange]);

  const patchDraft = useCallback((patch: Partial<CheckupFieldDraft>) => {
    setDraft((previous) => ({ ...previous, ...patch }));
    setError("");
    setDirty(true);
  }, []);

  const addOption = useCallback(() => {
    setDraft((previous) => ({
      ...previous,
      options: [...previous.options, { value: "", label: "" }],
    }));
    setDirty(true);
  }, []);

  const updateOption = useCallback((index: number, patch: Partial<CheckupFieldOptionDraft>) => {
    setDraft((previous) => ({
      ...previous,
      options: previous.options.map((option, i) =>
        i === index ? { ...option, ...patch } : option,
      ),
    }));
    setError("");
    setDirty(true);
  }, []);

  const removeOption = useCallback((index: number) => {
    setDraft((previous) => ({
      ...previous,
      options: previous.options.filter((_, i) => i !== index),
    }));
    setDirty(true);
  }, []);

  const saveField = useCallback(async () => {
    const validationError = validateCheckupFieldDraft(draft);
    if (validationError) {
      setError(validationError);
      return;
    }
    try {
      if (editingId === "new") {
        if (!canCreate) return;
        await createField.mutateAsync({
          checkupTypeId,
          req: buildCheckupFieldCreateRequest(draft, fieldCount + 1),
        });
      } else {
        if (!canEdit || !editingField) return;
        await updateField.mutateAsync({
          checkupTypeId,
          fieldId: editingId,
          req: buildCheckupFieldUpdateRequest(draft, editingField),
        });
      }
    } catch {
      return;
    }
    setError("");
    setDirty(false);
    if (editingId === "new") onClose();
  }, [
    canCreate,
    canEdit,
    checkupTypeId,
    createField,
    draft,
    editingField,
    editingId,
    fieldCount,
    onClose,
    updateField,
  ]);

  return {
    draft,
    error,
    saveField,
    patchDraft,
    addOption,
    updateOption,
    removeOption,
    cancelEdit: onClose,
  };
}
