import { useActionState, useCallback, useMemo, useState } from "react";
import { toast } from "sonner";

import { useGetAllVaccinesMaster } from "@/hooks/use-treatment-master";
import { useCreateVaccination } from "@/hooks/use-create-vaccination";
import { todayJSTISO } from "@/lib/jst-date";
import {
  calculateNextDate,
  resolveScheduleTypeAfterManualDate,
} from "@/components/shared/NextScheduleField";
import { useGetPetVaccinations } from "../api/get-pet-vaccinations";
import type { VaccinationHistoryItem } from "../components/VaccinationHistory";

interface VaccinationSaveState {
  success: boolean;
  timestamp: number;
}

const INITIAL_SAVE_STATE: VaccinationSaveState = { success: false, timestamp: 0 };

function makeDefaultNextScheduleType() {
  return "4weeks";
}

/**
 * FE-RC-008: MedicalRecordVaccination の埋め込み追加フォームの状態・保存アクションを抽出。
 * SearchableSelect / DatePicker / NextScheduleField は非ネイティブの制御コンポーネントのため、
 * 独立フォーム (features/vaccinations/hooks/use-vaccination-form.ts) と同じ規約で
 * ブラウザ FormData ではなく closure 化した React state を読む useActionState を使う。
 */
export function useMedicalRecordVaccinationForm(petId?: string, medicalRecordId?: string) {
  const [vaccineName, setVaccineNameRaw] = useState("");
  // BUG-501 / 仕様 15 §1.1: 実施日は JST 当日デフォルト（独立フォーム use-vaccination-form と揃える）
  const [date, setDateRaw] = useState(() => todayJSTISO());
  const [supplemental, setSupplemental] = useState("");
  const [lot1, setLot1] = useState("");
  const [lot2, setLot2] = useState("");
  const [lot3, setLot3] = useState("");
  const [lot4, setLot4] = useState("");
  const [nextScheduleType, setNextScheduleTypeRaw] = useState(makeDefaultNextScheduleType);
  const [nextDate, setNextDateRaw] = useState("");
  const [remarks, setRemarks] = useState("");
  const [isAdding, setIsAdding] = useState(false);
  // BUG-015: 未選択のまま追加すると early return で無音失敗していた → 明示 fieldErrors
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const { data: historyItems = [], isLoading } = useGetPetVaccinations(petId);
  const { data: vaccinesMaster = [] } = useGetAllVaccinesMaster();
  const { mutateAsync: createVaccination } = useCreateVaccination();

  const vaccineOptions = useMemo(
    () => vaccinesMaster.flatMap((v) => (v.isActive ? [{ value: v.id, label: v.name }] : [])),
    [vaccinesMaster],
  );

  const resetForm = useCallback(() => {
    setVaccineNameRaw("");
    // フォームをリセット（実施日は次回追加でも当日デフォルト）
    setDateRaw(todayJSTISO());
    setSupplemental("");
    setLot1("");
    setLot2("");
    setLot3("");
    setLot4("");
    setNextScheduleTypeRaw(makeDefaultNextScheduleType());
    setNextDateRaw("");
    setRemarks("");
    setFieldErrors({});
    setIsAdding(false);
  }, []);

  const handleDuplicate = useCallback((item: VaccinationHistoryItem) => {
    setIsAdding(true);
    setVaccineNameRaw(String(item.vaccineId));
    setDateRaw(""); // 実施日は新しく入力させる
    setLot1(item.lot1);
    setLot2(item.lot2);
    setLot3(item.lot3);
    setLot4(item.lot4);
    setNextDateRaw(item.nextDate);
    setRemarks(item.remarks);
  }, []);

  const setVaccineName = useCallback((value: string) => {
    setVaccineNameRaw(value);
    setFieldErrors((prev) => {
      if (!prev.vaccineId) return prev;
      const { vaccineId: _removed, ...rest } = prev;
      return rest;
    });
  }, []);

  const setDate = useCallback(
    (value: string) => {
      setDateRaw(value);
      const calculated = calculateNextDate(value, nextScheduleType);
      if (calculated) setNextDateRaw(calculated);
      setFieldErrors((prev) => {
        if (!prev.date) return prev;
        const { date: _removed, ...rest } = prev;
        return rest;
      });
    },
    [nextScheduleType],
  );

  const setNextScheduleType = useCallback(
    (value: string) => {
      setNextScheduleTypeRaw(value);
      const calculated = calculateNextDate(date, value);
      if (calculated) setNextDateRaw(calculated);
    },
    [date],
  );

  const setNextDate = useCallback(
    (value: string) => {
      setNextDateRaw(value);
      setNextScheduleTypeRaw(resolveScheduleTypeAfterManualDate(date, nextScheduleType, value));
    },
    [date, nextScheduleType],
  );

  const [, formAction, isSaving] = useActionState<VaccinationSaveState>(
    async (): Promise<VaccinationSaveState> => {
      // ヘッダー外の入力 (履歴の検索欄など) の Enter 送信では追加処理をしない
      if (!isAdding) return { success: false, timestamp: Date.now() };

      // 独立フォーム (use-vaccination-form) と同じ必須文言。未選択は API を叩かない。
      const errors: Record<string, string> = {};
      if (!vaccineName || vaccineName === "0") {
        errors.vaccineId = "ワクチン種別を選択してください";
      }
      if (!date) {
        errors.date = "接種日を入力してください";
      }
      if (Object.keys(errors).length > 0) {
        setFieldErrors(errors);
        return { success: false, timestamp: Date.now() };
      }
      if (!petId) return { success: false, timestamp: Date.now() };
      if (!medicalRecordId) {
        setFieldErrors({ date: "カルテを保存してから接種を追加してください" });
        return { success: false, timestamp: Date.now() };
      }

      setFieldErrors({});
      try {
        await createVaccination({
          pet_id: Number(petId),
          medical_record_id: Number(medicalRecordId),
          vaccine_id: Number(vaccineName),
          date,
          lot1: lot1 || undefined,
          lot2: lot2 || undefined,
          lot3: lot3 || undefined,
          lot4: lot4 || undefined,
          next_date: nextDate || null,
          supplemental: supplemental || undefined,
          next_schedule_type: nextScheduleType || undefined,
          remarks: remarks || undefined,
        });
        toast.success("接種記録を追加しました");
        resetForm();
        return { success: true, timestamp: Date.now() };
      } catch {
        // useCreateVaccination onError が handleApiError 済み
        return { success: false, timestamp: Date.now() };
      }
    },
    INITIAL_SAVE_STATE,
  );

  return {
    isAdding,
    setIsAdding,
    vaccineOptions,
    vaccineName,
    setVaccineName,
    date,
    setDate,
    supplemental,
    setSupplemental,
    lot1,
    setLot1,
    lot2,
    setLot2,
    lot3,
    setLot3,
    lot4,
    setLot4,
    nextScheduleType,
    setNextScheduleType,
    nextDate,
    setNextDate,
    remarks,
    setRemarks,
    fieldErrors,
    formAction,
    isSaving,
    handleDuplicate,
    historyItems,
    isLoading,
  };
}
