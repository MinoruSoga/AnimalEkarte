import { useCallback, useEffect, useRef } from "react";

import { handleApiError } from "@/lib/handle-api-error";
import type { ActionState } from "@/types/form";

interface UseMedicalRecordPostSaveArgs {
  activeTab: string;
  formState: ActionState;
  markClean: () => void;
}

export function useMedicalRecordPostSave({
  activeTab,
  formState,
  markClean,
}: UseMedicalRecordPostSaveArgs) {
  const activeTabRef = useRef(activeTab);
  const clinicalPlanSaveRef = useRef<(() => Promise<void>) | null>(null);
  const estimateSaveRef = useRef<(() => Promise<void>) | null>(null);

  useEffect(() => {
    activeTabRef.current = activeTab;
  }, [activeTab]);

  const handleRegisterClinicalPlanSave = useCallback((fn: () => Promise<void>) => {
    clinicalPlanSaveRef.current = fn;
  }, []);

  const handleRegisterEstimateSave = useCallback((fn: () => Promise<void>) => {
    estimateSaveRef.current = fn;
  }, []);

  useEffect(() => {
    if (!formState.success) return;

    const currentTab = activeTabRef.current;

    const doPostSave = async () => {
      try {
        if (currentTab === "診察/治療プラン") {
          await (clinicalPlanSaveRef.current?.() ?? Promise.resolve());
        } else if (currentTab === "見積書") {
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
    handleRegisterClinicalPlanSave,
    handleRegisterEstimateSave,
  };
}
