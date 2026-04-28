// React/Framework
import { memo, useMemo, useState, useCallback, useTransition } from "react";

// Internal
import { useGetAllVaccinesMaster } from "@/hooks/use-treatment-master";
import { useCreateVaccination } from "@/hooks/use-vaccinations";
import { handleApiError } from "@/lib/handle-api-error";
import { C, PALETTE } from "@/lib/design-tokens";

// Relative
import { useGetPetVaccinations } from "../api/get-pet-vaccinations";
import type { PetVaccinationHistoryItem } from "../api/get-pet-vaccinations";
import { VaccinationForm } from "./VaccinationForm";
import { VaccinationHistory } from "./VaccinationHistory";

type LstepStatus = "synced" | "not-linked" | "opt-out";

interface MedicalRecordVaccinationProps {
  petId?: string;
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
  lstepStatus,
}: MedicalRecordVaccinationProps) {
  const [vaccineName, setVaccineName] = useState("");
  const [date, setDate] = useState("");
  const [supplemental, setSupplemental] = useState("");
  const [lot1, setLot1] = useState("");
  const [lot2, setLot2] = useState("");
  const [lot3, setLot3] = useState("");
  const [lot4, setLot4] = useState("");
  const [nextScheduleType, setNextScheduleType] = useState("4weeks");
  const [nextDate, setNextDate] = useState("");
  const [remarks, setRemarks] = useState("");

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

  const handleDuplicate = useCallback((item: PetVaccinationHistoryItem) => {
    setVaccineName(String(item.vaccineId));
    setDate(""); // 実施日は新しく入力させる
    setLot1(item.lot1);
    setLot2(item.lot2);
    setLot3(item.lot3);
    setLot4(item.lot4);
    setNextDate(item.nextDate);
    setRemarks(item.remarks);
  }, []);

  const handleSave = useCallback(() => {
    if (!petId || !vaccineName || !date) return;
    startSaveTransition(async () => {
      try {
        await createVaccination({
          pet_id: Number(petId),
          vaccine_id: Number(vaccineName),
          date,
          lot1: lot1 || undefined,
          lot2: lot2 || undefined,
          lot3: lot3 || undefined,
          lot4: lot4 || undefined,
          next_date: nextDate || null,
          remarks: remarks || undefined,
        });
        // フォームをリセット
        setVaccineName("");
        setDate("");
        setSupplemental("");
        setLot1("");
        setLot2("");
        setLot3("");
        setLot4("");
        setNextScheduleType("4weeks");
        setNextDate("");
        setRemarks("");
      } catch (err) {
        handleApiError(err, "予防接種の登録");
      }
    });
  }, [petId, vaccineName, date, lot1, lot2, lot3, lot4, nextDate, remarks, createVaccination]);

  return (
    <>
      {lstepStatus !== undefined ? (
        <div className="flex items-center gap-2 px-1 py-1.5 shrink-0">
          <LstepStatusBadge status={lstepStatus} />
        </div>
      ) : null}
    <div className="grid grid-cols-12 gap-4 h-[calc(100vh-220px)] min-h-[500px] overflow-y-auto pb-20 pr-1">
      {/* Left Column: Form */}
      <VaccinationForm
        vaccineOptions={vaccineOptions}
        vaccineName={vaccineName}
        setVaccineName={setVaccineName}
        date={date}
        setDate={setDate}
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
        setNextScheduleType={setNextScheduleType}
        nextDate={nextDate}
        setNextDate={setNextDate}
        remarks={remarks}
        setRemarks={setRemarks}
        onSave={handleSave}
        isSaving={isSaving}
      />

      {/* Right Column: History */}
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
