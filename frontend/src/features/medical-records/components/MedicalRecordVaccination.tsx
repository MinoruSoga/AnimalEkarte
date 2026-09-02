// React/Framework
import { memo } from "react";

// Internal
import { C, PALETTE } from "@/lib/design-tokens";

// Relative
import { EmptyState } from "@/components/shared/DataStates";
import { Button } from "@/components/ui/button";
import { VaccinationForm } from "./VaccinationForm";
import { VaccinationHistory } from "./VaccinationHistory";
import { useMedicalRecordVaccinationForm } from "../hooks/use-medical-record-vaccination-form";

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
  // FE-RC-008: 埋め込みフォームの状態と保存アクションは extract hook に集約。
  // 本体は useActionState の formAction を <form action> に渡すだけの薄い belt。
  const {
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
    handleDuplicate,
    historyItems,
    isLoading,
  } = useMedicalRecordVaccinationForm(petId, medicalRecordId);

  return (
    <>
      {lstepStatus !== undefined ? (
        <div className="flex items-center gap-2 px-1 py-1.5 shrink-0">
          <LstepStatusBadge status={lstepStatus} />
        </div>
      ) : null}
      <form
        action={formAction}
        className="grid grid-cols-1 gap-4 h-[calc(100vh-220px)] min-h-[500px] overflow-y-auto pb-20 pr-1 lg:grid-cols-5"
      >
        {isAdding ? (
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
            fieldErrors={fieldErrors}
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
      </form>
    </>
  );
});
