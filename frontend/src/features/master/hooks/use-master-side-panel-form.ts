import { useCallback, useEffect, useState } from "react";

interface UseMasterSidePanelFormOptions<TForm> {
  initialFormData: TForm;
  /**
   * FE-RC-029: onSave は必ず Promise<boolean>（または boolean）を返す契約にする。
   * false / reject 時は isDirty を落とさない（保存失敗を「未保存」のまま維持する）。
   */
  onSave: (data: TForm) => Promise<boolean> | boolean;
  onDirtyChange?: (dirty: boolean) => void;
  /**
   * フォームの妥当性を検証する。無効なら false を返す前提で、エラー表示は
   * validate 内で呼び出し側の state に副作用として設定する（フィールド数はケースにより異なるため）。
   */
  validate?: (data: TForm) => boolean;
}

export function useMasterSidePanelForm<TForm>({
  initialFormData,
  onSave,
  onDirtyChange,
  validate,
}: UseMasterSidePanelFormOptions<TForm>) {
  const [formData, setFormData] = useState<TForm>(initialFormData);
  const [isDirty, setIsDirty] = useState(false);

  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  const handleAction = useCallback(async () => {
    if (validate && !validate(formData)) return;
    const saved = await onSave(formData);
    if (saved) {
      setIsDirty(false);
      onDirtyChange?.(false);
    }
  }, [formData, onSave, onDirtyChange, validate]);

  return {
    formData,
    setFormData: setFormDataDirty,
    isDirty,
    setIsDirty,
    handleAction,
  };
}
