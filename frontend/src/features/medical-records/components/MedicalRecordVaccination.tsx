// React/Framework
import { memo, useMemo, useState, useCallback, useTransition } from "react";

// Internal
import { useGetAllVaccinesMaster } from "@/hooks/use-treatment-master";
import { useCreateVaccination } from "@/hooks/use-vaccinations";
import { C, PALETTE } from "@/lib/design-tokens";
import { todayJSTISO } from "@/lib/jst-date";
import { toast } from "sonner";

// Relative
import { useGetPetVaccinations } from "../api/get-pet-vaccinations";
import { EmptyState } from "@/components/shared/DataStates";
import { calculateNextDate, resolveScheduleTypeAfterManualDate } from "@/components/shared/NextScheduleField";
import { Button } from "@/components/ui/button";
import { VaccinationForm } from "./VaccinationForm";
import { VaccinationHistory } from "./VaccinationHistory";
import type { VaccinationHistoryItem } from "./VaccinationHistory";

type LstepStatus = "synced" | "not-linked" | "opt-out";

interface MedicalRecordVaccinationProps {
  petId?: string;
  medicalRecordId?: string;
  lstepStatus?: LstepStatus;
}

function LstepStatusBadge({ status }: { status: LstepStatus }) {
  if (status === "synced") {
    return (
      <span
        className="inline-flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full text-white"
        style={{ backgroundColor: PALETTE.lineGreen }}
      >
        LINE通知対象
      </span>
    );
  }
  if (status === "not-linked") {
    return (
      <span className={`inline-flex items-center text-xs font-medium px-2 py-0.5 rounded-full border ${C.textNotice} ${C.borderNotice} ${C.bgNotice40}`}>
        LINE未連携
      </span>
    );
  }
  return (
    <span className={`inline-flex items-center text-xs font-medium px-2 py-0.5 rounded-full border ${C.text40} ${C.borderMediumLight} ${C.bgPage30}`}>
      LINE受信拒否
    </span>
  );
}

export const MedicalRecordVaccination = memo(function MedicalRecordVaccination({
  petId,
  medicalRecordId,
  lstepStatus,
}: MedicalRecordVaccinationProps) {
  const [vaccineName, setVaccineName] = useState("");
  // BUG-501 / 仕様 15 §1.1: 実施日は JST 当日デフォルト（独立フォーム use-vaccination-form と揃える）
  const [date, setDate] = useState(() => todayJSTISO());
  const [supplemental, setSupplemental] = useState("");
  const [lot1, setLot1] = useState("");
  const [lot2, setLot2] = useState("");
  const [lot3, setLot3] = useState("");
  const [lot4, setLot4] = useState("");
  const [nextScheduleType, setNextScheduleType] = useState("4weeks");
  const [nextDate, setNextDate] = useState("");
  const [remarks, setRemarks] = useState("");
  const [isAdding, setIsAdding] = useState(false);
  // BUG-015: 未選択のまま追加すると early return で無音失敗していた → 明示 fieldErrors
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const [isSaving, startSaveTransition] = useTransition();

  const { data: historyItems = [], isLoading } = useGetPetVaccinations(petId);
  const { data: vaccinesMaster = [] } = useGetAllVaccinesMaster();
  const { mutateAsync: createVaccination } = useCreateVaccination();

  const vaccineOptions = useMemo(
    () => vaccinesMaster.flatMap((v) =>
      v.isActive ? [{ value: v.id, label: v.name }] : []
    ),
    [vaccinesMaster]
  );

  const handleDuplicate = useCallback((item: VaccinationHistoryItem) => {
    setIsAdding(true);
    setVaccineName(String(item.vaccineId));
    setDate(""); // 実施日は新しく入力させる
    setLot1(item.lot1);
    setLot2(item.lot2);
    setLot3(item.lot3);
    setLot4(item.lot4);
    setNextDate(item.nextDate);
    setRemarks(item.remarks);
  }, []);

  const handleVaccineNameChange = useCallback((value: string) => {
    setVaccineName(value);
    setFieldErrors((prev) => {
      if (!prev.vaccineId) return prev;
      const { vaccineId: _removed, ...rest } = prev;
      return rest;
    });
  }, []);

  const handleDateChange = useCallback((value: string) => {
    setDate(value);
    const calculated = calculateNextDate(value, nextScheduleType);
    if (calculated) setNextDate(calculated);
    setFieldErrors((prev) => {
      if (!prev.date) return prev;
      const { date: _removed, ...rest } = prev;
      return rest;
    });
  }, [nextScheduleType]);

  const handleNextScheduleTypeChange = useCallback((value: string) => {
    setNextScheduleType(value);
    const calculated = calculateNextDate(date, value);
    if (calculated) setNextDate(calculated);
  }, [date]);

  const handleNextDateChange = useCallback((value: string) => {
    setNextDate(value);
    setNextScheduleType(resolveScheduleTypeAfterManualDate(date, nextScheduleType, value));
  }, [date, nextScheduleType]);

  const handleSave = useCallback(() => {
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
      return;
    }
    if (!petId) return;
    if (!medicalRecordId) {
      setFieldErrors({ date: "カルテを保存してから接種を追加してください" });
      return;
    }

    setFieldErrors({});
    startSaveTransition(async () => {
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
        // フォームをリセット（実施日は次回追加でも当日デフォルト）
        setVaccineName("");
        setDate(todayJSTISO());
        setSupplemental("");
        setLot1("");
        setLot2("");
        setLot3("");
        setLot4("");
        setNextScheduleType("4weeks");
        setNextDate("");
        setRemarks("");
        setFieldErrors({});
        setIsAdding(false);
      } catch {
        // useCreateVaccination onError が handleApiError 済み
      }
    });
  }, [petId, medicalRecordId, vaccineName, date, supplemental, lot1, lot2, lot3, lot4, nextScheduleType, nextDate, remarks, createVaccination]);

  return (
    <>
      {lstepStatus !== undefined ? (
        <div className="flex items-center gap-2 px-1 py-1.5 shrink-0">
          <LstepStatusBadge status={lstepStatus} />
        </div>
      ) : null}
    <div className="grid grid-cols-1 gap-4 h-[calc(100vh-220px)] min-h-[500px] overflow-y-auto pb-20 pr-1 lg:grid-cols-5">
      {isAdding ? (
        <VaccinationForm
          vaccineOptions={vaccineOptions}
          vaccineName={vaccineName}
          setVaccineName={handleVaccineNameChange}
          date={date}
          setDate={handleDateChange}
          supplemental={supplemental}
          setSupplemental={setSupplemental}
          lot1={lot1}
          setLot1={setLot1}
          lot2={lot2}
          setLot2={setLot2}
          lot3={lot3}
          setLot3={setLot3}
          lot4={lot4}
          setLot4={setLot4}
          nextScheduleType={nextScheduleType}
          setNextScheduleType={handleNextScheduleTypeChange}
          nextDate={nextDate}
          setNextDate={handleNextDateChange}
          remarks={remarks}
          setRemarks={setRemarks}
          fieldErrors={fieldErrors}
          onSave={handleSave}
          isSaving={isSaving}
        />
      ) : historyItems.length > 0 ? (
        <div className="lg:col-span-3 flex flex-col gap-3">
          <ul className={`divide-y ${C.borderLight} ${C.bgWhite} rounded-lg border ${C.borderMedium}`}>
            {historyItems.map((item) => (
              <li key={item.id} className="px-3 py-2 text-sm">
                <div className={`font-medium ${C.text}`}>{item.name}</div>
                <div className={C.text60}>接種日 {item.date}</div>
              </li>
            ))}
          </ul>
          {petId ? (
            <Button type="button" variant="outline" size="sm" onClick={() => setIsAdding(true)}>
              記録を追加
            </Button>
          ) : null}
        </div>
      ) : (
        <EmptyState
          className="lg:col-span-3"
          message="接種記録がありません。下の「記録を追加」ボタンから追加してください。"
        >
          {petId ? (
            <Button type="button" variant="outline" size="sm" onClick={() => setIsAdding(true)}>
              記録を追加
            </Button>
          ) : null}
        </EmptyState>
      )}

      <VaccinationHistory
        historyItems={historyItems}
        isLoading={isLoading}
        onDuplicate={handleDuplicate}
        canCreate={!!petId}
      />
    </div>
    </>
  );
});
