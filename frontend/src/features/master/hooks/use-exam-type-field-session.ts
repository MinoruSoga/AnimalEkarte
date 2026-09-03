import { useCallback, useEffect, useState } from "react";

import {
  type ExaminationTypeField,
  useCreateExaminationTypeField,
  useReplaceExamTypeFieldReferenceRanges,
  useUpdateExaminationTypeField,
} from "../api/exam-types-master";
import {
  buildReferenceRangeRequest,
  toReferenceRangeDraft,
  validateReferenceRangeDrafts,
  type ReferenceRangeDraft,
} from "../components/exam-type-fields-editor-model";

export interface FieldDraft {
  name: string;
  inspectionValue: string;
  normalValue: string;
  unit: string;
}

const emptyFieldDraft = (): FieldDraft => ({
  name: "",
  inspectionValue: "",
  normalValue: "",
  unit: "",
});

const fieldToDraft = (field: ExaminationTypeField): FieldDraft => ({
  name: field.name,
  inspectionValue: field.inspectionValue,
  normalValue: field.normalValue,
  unit: field.unit,
});

interface UseExamTypeFieldSessionArgs {
  examTypeId: string;
  editingId: string | "new";
  editingField: ExaminationTypeField | null;
  canCreate: boolean;
  canEdit: boolean;
  onDirtyChange?: (dirty: boolean) => void;
  onClose: () => void;
}

export function useExamTypeFieldSession({
  examTypeId,
  editingId,
  editingField,
  canCreate,
  canEdit,
  onDirtyChange,
  onClose,
}: UseExamTypeFieldSessionArgs) {
  const createField = useCreateExaminationTypeField();
  const updateField = useUpdateExaminationTypeField();
  const replaceRanges = useReplaceExamTypeFieldReferenceRanges();
  const [fieldDraft, setFieldDraft] = useState<FieldDraft>(() =>
    editingField ? fieldToDraft(editingField) : emptyFieldDraft(),
  );
  const [rangeDrafts, setRangeDrafts] = useState<ReferenceRangeDraft[]>(() =>
    editingField ? editingField.referenceRanges.map(toReferenceRangeDraft) : [],
  );
  const [error, setError] = useState("");
  const [fieldDirty, setFieldDirty] = useState(false);
  const [rangeDirty, setRangeDirty] = useState(false);
  const hasDirtyDraft = fieldDirty || rangeDirty;

  useEffect(() => {
    onDirtyChange?.(hasDirtyDraft);
  }, [hasDirtyDraft, onDirtyChange]);

  useEffect(() => {
    return () => onDirtyChange?.(false);
  }, [onDirtyChange]);

  const saveField = useCallback(async () => {
    if (!fieldDraft.name.trim()) {
      setError("検査項目名を入力してください");
      return;
    }
    const req = {
      name: fieldDraft.name.trim(),
      inspection_value: fieldDraft.inspectionValue,
      normal_value: fieldDraft.normalValue,
      unit: fieldDraft.unit,
    };
    try {
      if (editingId === "new") {
        if (!canCreate) return;
        await createField.mutateAsync({ examTypeId, req });
      } else {
        if (!canEdit) return;
        await updateField.mutateAsync({ examTypeId, fieldId: editingId, req });
      }
    } catch {
      return;
    }
    setError("");
    setFieldDirty(false);
    if (editingId === "new") onClose();
  }, [canCreate, canEdit, createField, editingId, examTypeId, fieldDraft, onClose, updateField]);

  const saveRanges = useCallback(async () => {
    if (!canEdit || editingId === "new") return;
    const validationError = validateReferenceRangeDrafts(rangeDrafts);
    if (validationError) {
      setError(validationError);
      return;
    }
    try {
      await replaceRanges.mutateAsync({
        examTypeId,
        fieldId: editingId,
        ranges: buildReferenceRangeRequest(rangeDrafts),
      });
    } catch {
      return;
    }
    setError("");
    setRangeDirty(false);
  }, [canEdit, editingId, examTypeId, rangeDrafts, replaceRanges]);

  const toggleSpecies = useCallback((speciesId: string) => {
    setRangeDrafts((previous) => {
      if (previous.some((draft) => draft.animalSpeciesId === speciesId)) {
        return previous.filter((draft) => draft.animalSpeciesId !== speciesId);
      }
      return [
        ...previous,
        { animalSpeciesId: speciesId, mode: "numeric", min: "", max: "" },
      ];
    });
    setError("");
    setRangeDirty(true);
  }, []);

  const updateRange = useCallback((
    speciesId: string,
    update: (draft: ReferenceRangeDraft) => ReferenceRangeDraft,
  ) => {
    setRangeDrafts((previous) => previous.map((draft) =>
      draft.animalSpeciesId === speciesId ? update(draft) : draft
    ));
    setError("");
    setRangeDirty(true);
  }, []);

  const patchFieldDraft = useCallback((patch: Partial<FieldDraft>) => {
    setFieldDraft((previous) => ({ ...previous, ...patch }));
    setFieldDirty(true);
  }, []);

  return {
    fieldDraft,
    rangeDrafts,
    error,
    saveField,
    saveRanges,
    toggleSpecies,
    updateRange,
    patchFieldDraft,
    cancelEdit: onClose,
  };
}
