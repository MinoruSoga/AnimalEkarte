import { useState, useEffect } from "react";
import { ClinicInfo, DEFAULT_CLINIC_INFO } from "../types";
import { toast } from "sonner";

const STORAGE_KEY = "figma_make_clinic_info";

export const useClinicInfo = () => {
  const [clinicInfo, setClinicInfo] = useState<ClinicInfo>(DEFAULT_CLINIC_INFO);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        setClinicInfo({ ...DEFAULT_CLINIC_INFO, ...JSON.parse(stored) });
      }
    } catch {
      // Use default clinic info if loading fails
    } finally {
      setLoading(false);
    }
  }, []);

  const updateClinicInfo = (newInfo: ClinicInfo) => {
    try {
      setClinicInfo(newInfo);
      localStorage.setItem(STORAGE_KEY, JSON.stringify(newInfo));
      toast.success("病院情報を更新しました");
    } catch {
      toast.error("病院情報の保存に失敗しました");
    }
  };

  return {
    clinicInfo,
    updateClinicInfo,
    loading
  };
};
