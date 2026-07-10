import { useCallback, useState } from "react";

/**
 * MedicalRecordForm のモーダル・印刷の開閉状態を集約するフック。
 * バイタル/スタッフ選択/オーナー検索/削除確認/印刷の各 UI 表示フラグを一箇所で管理する。
 */
export function useMedicalRecordFormModals() {
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  const [isVitalsOpen, setIsVitalsOpen] = useState(false);
  const [isPrinting, setIsPrinting] = useState(false);
  const [isStaffModalOpen, setIsStaffModalOpen] = useState(false);
  const [isOwnerSearchOpen, setIsOwnerSearchOpen] = useState(false);

  const handleStaffModalOpenChange = useCallback((open: boolean) => {
    setIsStaffModalOpen(open);
  }, []);

  const handleOpenStaffModal = useCallback(() => {
    setIsStaffModalOpen(true);
  }, []);

  const handleOpenOwnerSearch = useCallback(() => {
    setIsOwnerSearchOpen(true);
  }, []);

  const handlePrintClick = useCallback(() => {
    setIsPrinting(true);
    setTimeout(() => {
      window.print();
      setIsPrinting(false);
    }, 100);
  }, []);

  return {
    isDeleteConfirmOpen,
    setIsDeleteConfirmOpen,
    isVitalsOpen,
    setIsVitalsOpen,
    isPrinting,
    isStaffModalOpen,
    isOwnerSearchOpen,
    setIsOwnerSearchOpen,
    handleStaffModalOpenChange,
    handleOpenStaffModal,
    handleOpenOwnerSearch,
    handlePrintClick,
  };
}
