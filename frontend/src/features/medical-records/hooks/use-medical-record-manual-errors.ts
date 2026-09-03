import { useEffect, useState } from "react";

const MEDICAL_RECORD_PRIORITY_FIELDS = ["treatment_policy", "diagnosis1_category_id"] as const;

/** エラーフィールドへのフォーカス移動を tab 切替後まで遅延させる時間 (FE5-6) */
const FOCUS_DELAY_MS = 50;

interface UseMedicalRecordManualErrorsArgs {
  setActiveTab: (tab: string) => void;
}

export function useMedicalRecordManualErrors({ setActiveTab }: UseMedicalRecordManualErrorsArgs) {
  const [manualErrors, setManualErrors] = useState<Record<string, string>>({});
  const [prevManualErrors, setPrevManualErrors] = useState(manualErrors);

  if (prevManualErrors !== manualErrors) {
    setPrevManualErrors(manualErrors);
    const errorFields = Object.keys(manualErrors);
    if (errorFields.length > 0) {
      const firstError =
        MEDICAL_RECORD_PRIORITY_FIELDS.find((field) => errorFields.includes(field)) ??
        errorFields[0];
      if (firstError === "treatment_policy") {
        setActiveTab("問診");
      } else if (firstError === "diagnosis1_category_id" || firstError === "diagnosis1_name_id") {
        setActiveTab("診察/治療プラン");
      }
    }
  }

  useEffect(() => {
    const errorFields = Object.keys(manualErrors);
    if (errorFields.length === 0) return;

    const firstError =
      MEDICAL_RECORD_PRIORITY_FIELDS.find((field) => errorFields.includes(field)) ?? errorFields[0];
    setTimeout(() => {
      const element = document.getElementById(firstError);
      if (element) {
        element.focus();
        element.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }, FOCUS_DELAY_MS);
  }, [manualErrors]);

  return {
    manualErrors,
    setManualErrors,
  };
}
