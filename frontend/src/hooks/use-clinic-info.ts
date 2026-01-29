import { useState, useEffect } from "react";
import { toast } from "sonner";
import type { ClinicInfo } from "@/types";

const STORAGE_KEY = "figma_make_clinic_info";

const DEFAULT_CLINIC_INFO: ClinicInfo = {
  name: "ノア動物病院",
  branchName: "八王子院",
  postalCode: "192-0083",
  address: "東京都八王子市旭町1-1",
  phoneNumber: "042-123-4567",
  registrationNumber: "東京都獣医師会 第12345号",
};

export function useClinicInfo() {
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
    loading,
  };
}
