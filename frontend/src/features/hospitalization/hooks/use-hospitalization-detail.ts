import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import type { CreateCarePlanDTO, UpdateCarePlanDTO, CreateVitalDTO, CreateCareLogDTO } from "../types";
import { useGetHospitalization } from "../api/get-hospitalization";
import { useUpdateHospitalization } from "../api/update-hospitalization";
import { dischargeWithBilling } from "../api/discharge-with-billing";
import {
  useGetCarePlanItems,
  useCreateCarePlanItem,
  useUpdateCarePlanItem,
  useDeleteCarePlanItem,
  type CarePlanItem as ApiCarePlanItem,
} from "../api/care-plan-items";
import {
  useGetDailyRecords,
  dailyRecordKeys,
  addVitalRecord,
  addCareLogRecord,
  type ApiDailyRecord,
} from "../api/daily-records";
import type { CarePlanItem, DailyRecord } from "@/types";

// ---- 型変換ヘルパー ----

/** ApiCarePlanItem (snake_case) → CarePlanItem (camelCase) */
function mapApiCarePlanItem(api: ApiCarePlanItem): CarePlanItem {
  return {
    id: api.id,
    hospitalizationId: api.hospitalization_id,
    type: api.type,
    name: api.name,
    description: api.description,
    timing: api.timing,
    status: api.status,
    notes: api.notes,
    medicineId: api.medicine_id,
    procedureId: api.procedure_id,
    hospitalizationPlanId: api.hospitalization_plan_id,
    unitPrice: api.unit_price,
    category: api.category,
    sortOrder: api.sort_order,
    createdAt: api.created_at,
    updatedAt: api.updated_at,
  };
}

/** ApiDailyRecord (snake_case) → DailyRecord (camelCase) */
function mapApiDailyRecord(api: ApiDailyRecord): DailyRecord {
  return {
    id: api.id,
    hospitalizationId: api.hospitalization_id,
    date: api.date,
    vitals: (api.vital_records ?? []).map((v) => ({
      id: v.id,
      time: v.time,
      temperature: v.temperature,
      heartRate: v.heart_rate,
      respirationRate: v.respiration_rate,
      weight: v.weight,
      notes: v.notes,
    })),
    careLogs: (api.care_log_records ?? []).map((c) => ({
      id: c.id,
      time: c.time,
      type: c.type as DailyRecord["careLogs"][0]["type"],
      status: c.status,
      value: c.value,
      notes: c.notes,
    })),
    staffNotes: (api.staff_note_records ?? []).map((s) => ({
      id: s.id,
      time: s.time,
      content: s.content,
    })),
  };
}

// ---- Hook ----

export const useHospitalizationDetail = (hospitalizationId?: string) => {
  const id = hospitalizationId ?? "";
  const queryClient = useQueryClient();

  const { data: hospitalization, isLoading } = useGetHospitalization(id);
  const { mutateAsync: updateHosp } = useUpdateHospitalization();

  // ---- CarePlan ----
  const { data: rawPlans } = useGetCarePlanItems(id);
  const { mutateAsync: createPlan } = useCreateCarePlanItem(id);
  const { mutateAsync: updatePlan } = useUpdateCarePlanItem(id);
  const { mutateAsync: deletePlan } = useDeleteCarePlanItem(id);

  const plans: CarePlanItem[] = (rawPlans ?? []).map(mapApiCarePlanItem);

  // ---- DailyRecords ----
  const { data: rawRecords } = useGetDailyRecords(id);
  const records: DailyRecord[] = (rawRecords ?? []).map(mapApiDailyRecord);

  // ---- Handlers ----

  const handleAddPlan = async (plan: CreateCarePlanDTO) => {
    try {
      await createPlan({
        type: plan.type,
        name: plan.name,
        description: plan.description,
        timing: plan.timing,
        status: plan.status,
        notes: plan.notes,
        medicine_id: plan.medicineId,
        procedure_id: plan.procedureId,
        hospitalization_plan_id: plan.hospitalizationPlanId,
        unit_price: plan.unitPrice,
        category: plan.category,
        sort_order: plan.sortOrder,
      });
      toast.success("ケアプランを追加しました");
    } catch (error) {
      handleApiError(error, "ケアプランの追加");
    }
  };

  const handleUpdatePlan = async (planId: string, updates: UpdateCarePlanDTO) => {
    try {
      await updatePlan({
        itemId: planId,
        input: {
          type: updates.type,
          name: updates.name,
          description: updates.description,
          timing: updates.timing,
          status: updates.status,
          notes: updates.notes,
          medicine_id: updates.medicineId,
          procedure_id: updates.procedureId,
          hospitalization_plan_id: updates.hospitalizationPlanId,
          unit_price: updates.unitPrice,
          category: updates.category,
          sort_order: updates.sortOrder,
        },
      });
      toast.success("ケアプランを更新しました");
    } catch (error) {
      handleApiError(error, "ケアプランの更新");
    }
  };

  const handleDeletePlan = async (planId: string) => {
    try {
      await deletePlan(planId);
      toast.success("ケアプランを削除しました");
    } catch (error) {
      handleApiError(error, "ケアプランの削除");
    }
  };

  const handleAddVital = async (date: string, data: CreateVitalDTO) => {
    try {
      const result = await addVitalRecord(id, date, {
        time: data.time,
        temperature: data.temperature,
        heart_rate: data.heartRate,
        respiration_rate: data.respirationRate,
        weight: data.weight,
        notes: data.notes,
      });
      queryClient.setQueryData(dailyRecordKeys.byDate(id, date), result);
      queryClient.invalidateQueries({ queryKey: dailyRecordKeys.all(id) });
      toast.success("バイタルを記録しました");
    } catch (error) {
      handleApiError(error, "バイタルの記録");
    }
  };

  const handleAddLog = async (date: string, data: CreateCareLogDTO) => {
    try {
      const result = await addCareLogRecord(id, date, {
        time: data.time,
        type: data.type,
        status: data.status,
        value: data.value,
        notes: data.notes,
      });
      queryClient.setQueryData(dailyRecordKeys.byDate(id, date), result);
      queryClient.invalidateQueries({ queryKey: dailyRecordKeys.all(id) });
      toast.success("記録を追加しました");
    } catch (error) {
      handleApiError(error, "記録の追加");
    }
  };

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
