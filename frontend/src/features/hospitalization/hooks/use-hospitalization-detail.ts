import { isAxiosError } from "axios";
import { toast } from "sonner";
import { useLayoutEffect, useRef } from "react";
import { handleApiError } from "@/lib/handle-api-error";
import { jstNowISOString } from "@/lib/jst-date";
import { usePermission } from "@/hooks/use-permission";
import { useGetHospitalization } from "../api/get-hospitalization";
import { useUpdateHospitalization } from "../api/update-hospitalization";
import { dischargeWithBilling } from "../api/discharge-with-billing";

export const useHospitalizationDetail = (hospitalizationId?: string) => {
  const id = hospitalizationId ?? "";

  const { data: hospitalization, isLoading, isError, error } = useGetHospitalization(id);
  const isNotFound = isAxiosError(error) && error.response?.status === 404;
  const { mutateAsync: updateHosp } = useUpdateHospitalization();
  const { canEdit } = usePermission("hospitalization");
  const canEditRef = useRef(canEdit);
  const petIsDeceasedRef = useRef(hospitalization?.petIsDeceased);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
    petIsDeceasedRef.current = hospitalization?.petIsDeceased;
  }, [canEdit, hospitalization?.petIsDeceased]);

  const dischargeHospitalization = async (
    createAccounting = false,
  ): Promise<{ success: boolean; accountingId?: number }> => {
    if (!hospitalizationId || !hospitalization) {
      return { success: false };
    }
    if (canEditRef.current !== true || petIsDeceasedRef.current === true) {
      return { success: false };
    }
    if (createAccounting) {
      // dischargeWithBilling は mutation ではない生の非同期呼び出しのため、
      // ここでの handleApiError が唯一の通知経路になる。
      try {
        const result = await dischargeWithBilling(hospitalizationId, {
          discharge_date: jstNowISOString(),
          create_accounting: true,
        });
        toast.success("退院処理が完了しました");
        return { success: true, accountingId: result.accounting_id };
      } catch (error) {
        handleApiError(error, "退院処理");
        return { success: false };
      }
    }
    try {
      // FE-RC-005: useUpdateHospitalization.onError が既に handleApiError でトースト表示済みのため、
      // ここでは再通知しない。
      await updateHosp({
        id: hospitalizationId,
        req: {
          status: "discharged",
          end_date: jstNowISOString(),
        },
      });
      toast.success("退院処理が完了しました");
      return { success: true };
    } catch {
      return { success: false };
    }
  };

  return {
    hospitalization: hospitalization ?? null,
    isLoading,
    isError,
    isNotFound,
    dischargeHospitalization,
  };
};
