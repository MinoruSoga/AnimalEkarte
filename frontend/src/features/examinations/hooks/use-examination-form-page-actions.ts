/**
 * FE-RC-045: ExaminationForm.tsx を 400 行以内に分割するため、
 * 「markDirty/markClean を伴う入力ハンドラ」と「削除確認ダイアログ」の
 * ページ側グルーコードをここへ抽出する。
 */
import { useCallback, useState } from "react";
import type { NavigateFunction } from "react-router";
import { paths } from "@/config/paths";
import type { Pet } from "@/types";

export function useExaminationFormPageActions<TFormDataUpdate>({
  navigate,
  fromPath,
  markDirty,
  markClean,
  setFormData,
  setInspectionValue,
  setItemName,
  addManualItem,
  removeItem,
  setSelectedPets,
  handleDelete,
}: {
  navigate: NavigateFunction;
  fromPath: string | undefined;
  markDirty: () => void;
  markClean: () => void;
  setFormData: (next: TFormDataUpdate) => void;
  setInspectionValue: (key: string, value: string) => void;
  setItemName: (key: string, value: string) => void;
  addManualItem: () => void;
  removeItem: (key: string) => void;
  setSelectedPets: (pets: Pet[]) => void;
  handleDelete: (onSuccess: () => void) => void;
}) {
  const handleBack = useCallback(() => {
    if (fromPath) {
      navigate(fromPath);
    } else {
      navigate(paths.examinations.getHref());
    }
  }, [fromPath, navigate]);

  // rerender-memo: memo'd セクションに渡すハンドラを useCallback で安定化
  const handleSetFormData = useCallback(
    (next: TFormDataUpdate) => {
      markDirty();
      setFormData(next);
    },
    [markDirty, setFormData],
  );

  const handleInspectionValueChange = useCallback(
    (key: string, value: string) => {
      markDirty();
      setInspectionValue(key, value);
    },
    [markDirty, setInspectionValue],
  );

  const handleItemNameChange = useCallback(
    (key: string, value: string) => {
      markDirty();
      setItemName(key, value);
    },
    [markDirty, setItemName],
  );

  const handleAddItem = useCallback(() => {
    markDirty();
    addManualItem();
  }, [addManualItem, markDirty]);

  const handleRemoveItem = useCallback(
    (key: string) => {
      markDirty();
      removeItem(key);
    },
    [markDirty, removeItem],
  );

  const handlePatientSelect = useCallback(
    (pet: Pet) => {
      markDirty();
      setSelectedPets([pet]);
    },
    [markDirty, setSelectedPets],
  );

  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);

  const handleDeleteClick = useCallback(() => {
    setIsDeleteConfirmOpen(true);
  }, []);

  const handleDeleteCancel = useCallback(() => {
    setIsDeleteConfirmOpen(false);
  }, []);

  const handleDeleteConfirm = useCallback(() => {
    markClean();
    handleDelete(() => navigate(paths.examinations.getHref()));
    setIsDeleteConfirmOpen(false);
  }, [markClean, handleDelete, navigate]);

  return {
    handleBack,
    handleSetFormData,
    handleInspectionValueChange,
    handleItemNameChange,
    handleAddItem,
    handleRemoveItem,
    handlePatientSelect,
    isDeleteConfirmOpen,
    handleDeleteClick,
    handleDeleteCancel,
    handleDeleteConfirm,
  };
}
