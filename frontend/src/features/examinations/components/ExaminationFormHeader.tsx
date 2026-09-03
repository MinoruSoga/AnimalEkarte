import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { LabDeviceUnlinkedBanner } from "@/components/shared/LabDeviceUnlinkedBanner/LabDeviceUnlinkedBanner";
import { C } from "@/lib/design-tokens";
import type { Pet } from "@/types";
import type { ExaminationPrintModel } from "../lib/examination-print-model";
import { ExaminationPatientChangeDialog } from "./ExaminationPatientChangeDialog";
import { ExaminationUnconfirmDialog } from "./ExaminationUnconfirmDialog";
import { ExaminationPrintArea } from "./ExaminationPrintArea";

interface ExaminationFormHeaderProps {
  selectedPet: Pet | undefined;
  selectedDoctorName: string;
  isPetDeceased: boolean;
  isEdit: boolean;
  isPatientChangeLocked: boolean;
  onPatientSelect: (pet: Pet) => void;
  isPersistedConfirmed: boolean;
  canUnconfirm: boolean;
  examinationId: string | undefined;
  onUnconfirm: (reason: string) => Promise<boolean>;
  printModel: ExaminationPrintModel | null;
}

// FE-RC-045/046: ExaminationForm.tsx から患者カード・死亡バナー・患者変更/確定解除ダイアログ・
// 印刷ボタン領域を分離（本体は 2 カラムレイアウトの組み立てに専念する）。
export function ExaminationFormHeader({
  selectedPet,
  selectedDoctorName,
  isPetDeceased,
  isEdit,
  isPatientChangeLocked,
  onPatientSelect,
  isPersistedConfirmed,
  canUnconfirm,
  examinationId,
  onUnconfirm,
  printModel,
}: ExaminationFormHeaderProps) {
  return (
    <>
      {/* rerender-memo: PatientInfoCard — フォームフィールド変更では再レンダーしない */}
      {selectedPet ? (
        <PatientInfoCard
          ownerName={selectedPet.ownerName}
          petName={`${selectedPet.name}${selectedPet.species ? `(${selectedPet.species})` : ""}`}
          petNumber={selectedPet.petNumber || selectedPet.id}
          weight={selectedPet.weight || "-"}
          staffName={selectedDoctorName}
          reservationType="検査"
          petDetails={`${selectedPet.birthDate ? `${selectedPet.birthDate}生` : ""} / ${selectedPet.species}`}
          insuranceName={selectedPet.insuranceName}
          insuranceDetails={selectedPet.insuranceDetails}
          status={isPetDeceased ? "deceased" : "alive"}
        />
      ) : null}
      {selectedPet ? <LabDeviceUnlinkedBanner petId={selectedPet.id} /> : null}
      {isPetDeceased ? (
        <div
          role="status"
          aria-label="死亡ペットのため保存不可"
          className={`flex items-center gap-2 px-4 py-2.5 rounded-md border ${C.bgWarning50} ${C.borderWarning20} ${C.textWarning}`}
        >
          <span className="text-sm font-medium">死亡したペットの検査記録は保存できません</span>
        </div>
      ) : null}

      {isEdit && !isPatientChangeLocked ? (
        <div className="flex justify-end">
          <ExaminationPatientChangeDialog selectedPet={selectedPet} onSelect={onPatientSelect} />
        </div>
      ) : null}

      {isPersistedConfirmed && canUnconfirm && examinationId ? (
        <div className="flex justify-end">
          <ExaminationUnconfirmDialog onUnconfirm={onUnconfirm} />
        </div>
      ) : null}

      {isEdit && examinationId ? (
        <div className="flex justify-end print:hidden">
          <button
            type="button"
            data-testid="examination-print-button"
            className={`rounded-xs border px-3 py-1.5 text-sm ${C.borderLight} ${C.text60} ${C.hoverBgLight} disabled:opacity-50`}
            disabled={!printModel}
            onClick={() => window.print()}
          >
            印刷 / PDF出力
          </button>
        </div>
      ) : null}

      {printModel ? <ExaminationPrintArea model={printModel} /> : null}
    </>
  );
}
