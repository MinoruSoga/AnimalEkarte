import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import type { CarePlanItem, DailyRecord } from "@/types";
import type {
  CreateCarePlanDTO,
  UpdateCarePlanDTO,
  CreateVitalDTO,
  CreateCareLogDTO,
} from "../types";
import { useGetHospitalization } from "../api/get-hospitalization";
import { useUpdateHospitalization } from "../api/update-hospitalization";

export const useHospitalizationDetail = (hospitalizationId?: string) => {
  const {
    data: hospitalization,
    isLoading,
  } = useGetHospitalization(hospitalizationId ?? "");

  const { mutateAsync: updateHosp } = useUpdateHospitalization();

  // CarePlan/DailyRecord はバックエンド側でサブエンティティAPIが未実装のため
  // 現在は「準備中」として通知し、Hospitalization に embed されたデータを使用する
  const plans: CarePlanItem[] = [];
  const records: DailyRecord[] = [];

  const handleAddPlan = async (_plan: CreateCarePlanDTO) => {
    toast.info("ケアプラン機能は準備中です");
  };

  const handleUpdatePlan = async (
    _planId: string,
    _updates: UpdateCarePlanDTO
  ) => {
    toast.info("ケアプラン機能は準備中です");
  };

  const handleDeletePlan = async (_planId: string) => {
    toast.info("ケアプラン機能は準備中です");
  };

  const handleAddVital = async (_date: string, _data: CreateVitalDTO) => {
    toast.info("バイタル記録機能は準備中です");
  };

  const handleAddLog = async (_date: string, _data: CreateCareLogDTO) => {
    toast.info("記録追加機能は準備中です");
  };

  const dischargeHospitalization = async (): Promise<boolean> => {
    if (!hospitalizationId || !hospitalization) {
      return false;
    }
    try {
      await updateHosp({
        id: hospitalizationId,
        req: {
          status: "discharged",
          end_date: new Date().toISOString(),
        },
      });
      toast.success("退院処理が完了しました");
      return true;
    } catch (error) {
      handleApiError(error, "退院処理");
      return false;
    }
  };

  return {
    hospitalization: hospitalization ?? null,
    plans,
    records,
    isLoading,
    handleAddPlan,
    handleUpdatePlan,
    handleDeletePlan,
    handleAddVital,
    handleAddLog,
    dischargeHospitalization,
  };
};
