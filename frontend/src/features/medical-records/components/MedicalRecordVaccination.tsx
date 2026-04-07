// React/Framework
import { memo, useMemo, useState, useCallback } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

// Internal
import { useGetAllVaccinesMaster } from "@/features/master";
import { useCreateVaccination } from "@/features/vaccinations";
import { handleApiError } from "@/lib/handle-api-error";
import { usePermission } from "@/features/auth";

// Relative
import { useGetPetVaccinations } from "../api/get-pet-vaccinations";
import { VaccinationForm } from "./VaccinationForm";
import { VaccinationHistory } from "./VaccinationHistory";

interface MedicalRecordVaccinationProps {
  petId?: string;
}

export const MedicalRecordVaccination = memo(function MedicalRecordVaccination({
  petId,
}: MedicalRecordVaccinationProps) {
  const [vaccineName, setVaccineName] = useState("esophagitis");
  const [date, setDate] = useState("");
  const [supplemental, setSupplemental] = useState("");
  const [lot1, setLot1] = useState("");
  const [lot2, setLot2] = useState("");
  const [lot3, setLot3] = useState("");
  const [lot4, setLot4] = useState("");
  const [nextScheduleType, setNextScheduleType] = useState("4weeks");
  const [nextDate, setNextDate] = useState("");
  const [remarks, setRemarks] = useState("");

  const { canCreate } = usePermission("medical-records");
  const { data: historyItems = [], isLoading } = useGetPetVaccinations(petId);
  const { data: vaccinesMaster = [] } = useGetAllVaccinesMaster();
  const createVaccinationMutation = useCreateVaccination();
  const queryClient = useQueryClient();

  const vaccineOptions = useMemo(
    () => vaccinesMaster.filter((v) => v.isActive).map((v) => ({ value: v.id, label: v.name })),
    [vaccinesMaster]
  );

  const handleSave = useCallback(async () => {
    if (!canCreate) return;
    if (!petId || !vaccineName || !date) {
      toast.error("必須項目を入力してください");
      return;
    }

    try {
      await createVaccinationMutation.mutateAsync({
        pet_id: Number(petId),
        vaccine_id: Number(vaccineName),
        date,
        next_date: nextDate || null,
        lot1: lot1 || undefined,
        lot2: lot2 || undefined,
        lot3: lot3 || undefined,
        lot4: lot4 || undefined,
        remarks: remarks || undefined,
      });

      toast.success("ワクチン記録を保存しました");
      queryClient.invalidateQueries({ queryKey: ["vaccinations", "pet", petId] });

      // Clear form
      setVaccineName("esophagitis");
      setDate("");
      setSupplemental("");
      setLot1("");
      setLot2("");
      setLot3("");
      setLot4("");
      setNextScheduleType("4weeks");
      setNextDate("");
      setRemarks("");
    } catch (error) {
      handleApiError(error, "ワクチン記録の保存");
    }
  }, [
    canCreate,
    petId,
    vaccineName,
    date,
    nextDate,
    lot1,
    lot2,
    lot3,
    lot4,
    remarks,
    createVaccinationMutation,
    queryClient,
  ]);

  return (
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
        onSave={canCreate ? handleSave : undefined}
        isSaving={createVaccinationMutation.isPending}
      />

      {/* Right Column: History */}
      <VaccinationHistory historyItems={historyItems} isLoading={isLoading} />
    </div>
  );
});
