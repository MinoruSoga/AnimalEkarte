import { useCallback, useEffect, useRef } from "react";

import { handleApiError } from "@/lib/handle-api-error";
import type { ActionState } from "@/types/form";

interface UseMedicalRecordPostSaveArgs {
  activeTab: string;
  formState: ActionState;
  markClean: () => void;
}

/**
 * 保存成功後のタブ固有フォローアップ。
 * BUG-010: clinical-plan は親 save action の単一 versioned PATCH が正本のため、
 * post-save での再書き込み経路は持たない（見積書のみ登録 save を維持）。
 */
export function useMedicalRecordPostSave({
  activeTab,
  formState,
  markClean,
}: UseMedicalRecordPostSaveArgs) {
  const activeTabRef = useRef(activeTab);
  const estimateSaveRef = useRef<(() => Promise<void>) | null>(null);

  useEffect(() => {
    activeTabRef.current = activeTab;
  }, [activeTab]);

  const handleRegisterEstimateSave = useCallback((fn: () => Promise<void>) => {
    estimateSaveRef.current = fn;
  }, []);

  useEffect(() => {
    if (!formState.success) return;

    const currentTab = activeTabRef.current;

    const doPostSave = async () => {
      try {
        if (currentTab === "見積書") {
          await (estimateSaveRef.current?.() ?? Promise.resolve());
        }
      } catch (error) {
        handleApiError(error, "データの保存");
      }
      markClean();
    };

    doPostSave();
  }, [formState.success, formState.timestamp, markClean]);

  return {
    handleRegisterEstimateSave,
  };
}
