// React/Framework
import { memo, useMemo, useState } from "react";

// Internal
import { useGetAllVaccinesMaster } from "@/features/master";

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

  const { data: historyItems = [], isLoading } = useGetPetVaccinations(petId);
  const { data: vaccinesMaster = [] } = useGetAllVaccinesMaster();

  const vaccineOptions = useMemo(
    () => vaccinesMaster.filter((v) => v.isActive).map((v) => ({ value: v.id, label: v.name })),
    [vaccinesMaster]
  );

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
      />

      {/* Right Column: History */}
      <VaccinationHistory historyItems={historyItems} isLoading={isLoading} />
    </div>
  );
});
