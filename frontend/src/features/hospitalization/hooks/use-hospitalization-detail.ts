import { isAxiosError } from "axios";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { useGetHospitalization } from "../api/get-hospitalization";
import { useUpdateHospitalization } from "../api/update-hospitalization";
import { dischargeWithBilling } from "../api/discharge-with-billing";

export const useHospitalizationDetail = (hospitalizationId?: string) => {
  const id = hospitalizationId ?? "";

  const { data: hospitalization, isLoading, isError, error } = useGetHospitalization(id);
  const isNotFound = isAxiosError(error) && error.response?.status === 404;
  const { mutateAsync: updateHosp } = useUpdateHospitalization();

  const dischargeHospitalization = async (createAccounting = false): Promise<{ success: boolean; accountingId?: number }> => {
    if (!hospitalizationId || !hospitalization) {
      return { success: false };
    }
    try {
      if (createAccounting) {
        const result = await dischargeWithBilling(hospitalizationId, {
          discharge_date: new Date().toISOString(),
          create_accounting: true,
        });
        toast.success("退院処理が完了しました");
        return { success: true, accountingId: result.accounting_id };
      }
      await updateHosp({
        id: hospitalizationId,
        req: {
          status: "discharged",
          end_date: new Date().toISOString(),
        },
      });
      toast.success("退院処理が完了しました");
      return { success: true };
    } catch (error) {
      handleApiError(error, "退院処理");
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
