import { useState, useEffect } from "react";
import { ClinicInfo, DEFAULT_CLINIC_INFO } from "../types";
import { toast } from "sonner@2.0.3";

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
    } catch (error) {
      console.error("Failed to load clinic info", error);
    } finally {
      setLoading(false);
    }
  }, []);

  const updateClinicInfo = (newInfo: ClinicInfo) => {
    try {
      setClinicInfo(newInfo);
      localStorage.setItem(STORAGE_KEY, JSON.stringify(newInfo));
      toast.success("病院情報を更新しました");
    } catch (error) {
      console.error("Failed to save clinic info", error);
      toast.error("病院情報の保存に失敗しました");
    }
  };

  return {
    clinicInfo,
    updateClinicInfo,
    loading
  };
};
