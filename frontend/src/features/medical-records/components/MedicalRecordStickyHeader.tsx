import { memo, useCallback } from "react";
import { Link } from "react-router";
import { ChevronDown, FileText } from "lucide-react";
import { PatientContextHeader } from "@/components/shared/PatientContextHeader";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { UnifiedTabsList } from "@/components/shared/UnifiedTabs";
import { Button } from "@/components/ui/button";
import { C, ICON } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { cn } from "@/lib/utils";
import { openOwnerReport } from "@/lib/owner-report-window";
import { usePermission } from "@/hooks/use-permission";
import { getMedicalRecordStatusColor } from "@/lib/status-helpers";
import { ResourceMedicalRecords } from "@/types/generated/models";
import { todayJSTISO } from "@/lib/jst-date";
import type { Pet } from "@/types";
import { NextVisitButton } from "./NextVisitButton";
import { VisitTypeSelect } from "./VisitTypeSelect";
import { isMedicalRecordFinalizedStatus } from "../lib/medical-record-lock";

interface MedicalRecordStickyHeaderProps {
  selectedPet: Pet;
  cohabitingPets: Pet[];
  staffName: string;
  visitType: string;
  visitCount: number;
  canEdit: boolean;
  isNewRecord: boolean;
  tabs: { value: string; label: string }[];
  recordDate?: string;
  recordStatus?: string;
  nextVisitDate: string;
  onVisitTypeChange: (value: string) => void;
  onStaffClick: () => void;
  onOwnerClick: () => void;
  onDateChange?: (date: string) => void;
  onNextVisitDatePatch: (date: string) => void;
  onNextVisitDateValidChange: (valid: boolean) => void;
  hasLineIntegration?: boolean;
}

const CohabitingPetChips = memo(function CohabitingPetChips({ pets }: { pets: Pet[] }) {
  return (
    <section
      aria-label="同居ペット"
      className={cn(
        "flex items-center gap-1.5 overflow-x-auto rounded-md p-2 [&::-webkit-scrollbar]:hidden",
        C.bgPage30,
      )}
      style={{ scrollbarWidth: "none" }}
    >
      <span className={`shrink-0 text-xs ${C.text50}`}>同居ペット</span>
      <div className="flex min-w-max gap-1.5">
        {pets.map((pet) => {
          const label = pet.species ? `${pet.name}（${pet.species}）` : pet.name;
          return (
            <Link
              key={pet.id}
              to={`${paths.medicalRecords.getHref()}?pet_id=${encodeURIComponent(pet.id)}`}
              className={cn(
                "h-8 shrink-0 rounded-md border bg-white px-2.5 text-sm leading-8 whitespace-nowrap transition-colors",
                C.text,
                C.borderMedium,
                C.hoverBgLight,
              )}
            >
              {label}
            </Link>
          );
        })}
      </div>
    </section>
  );
});

export function MedicalRecordStickyHeader({
  selectedPet,
  cohabitingPets,
  staffName,
  visitType,
  visitCount,
  canEdit,
  isNewRecord,
  tabs,
  recordDate,
  recordStatus,
  nextVisitDate,
  onVisitTypeChange,
  onStaffClick,
  onOwnerClick,
  onDateChange,
  onNextVisitDatePatch,
  onNextVisitDateValidChange,
  hasLineIntegration,
}: MedicalRecordStickyHeaderProps) {
  const isFinalized = isMedicalRecordFinalizedStatus(recordStatus);
  const canEditDate = canEdit && !isFinalized && !!onDateChange && !isNewRecord;
  const dateInputValue = recordDate ? recordDate.replace(/\//g, "-") : undefined;

  // #158: 飼主レポートを別ウィンドウで開く。view 権限がない場合はボタンを出さない。
  const { canView: canViewReport } = usePermission(ResourceMedicalRecords);
  const reportOwnerId = selectedPet.ownerId;
  const reportPetId = selectedPet.id;
  const handleOpenReport = useCallback(() => {
    openOwnerReport(reportOwnerId, reportPetId);
  }, [reportOwnerId, reportPetId]);

  const contextControls = (
    <>
      {/* SPEC-GAP: 確定済みバッジ。臨床記録の真正性担保のため、確定状態を常時明示する */}
      {!isNewRecord && isFinalized ? (
        <StatusBadge colorClass={getMedicalRecordStatusColor("確定済")}>確定済</StatusBadge>
      ) : null}

      {/* 来院種別 */}
      <VisitTypeSelect
        value={visitType}
        onChange={onVisitTypeChange}
        disabled={!canEdit || isFinalized}
      />

      {/* 診察日 */}
      <div className="flex flex-col gap-0 shrink-0 min-w-[110px]">
        <span className={`text-xs ${C.text50}`}>診察日</span>
        {canEditDate ? (
          <input
            key={dateInputValue}
            type="date"
            aria-label="診察日"
            defaultValue={dateInputValue}
            onChange={(e) => {
              if (e.target.value) onDateChange!(e.target.value);
            }}
            className={`h-11 text-sm ${C.text} bg-transparent rounded px-1 cursor-pointer outline-none focus-visible:ring-2 ${C.focusRingAccent40}`}
          />
        ) : (
          <span className={`h-8 flex items-center text-sm ${C.text}`}>
            {isNewRecord ? todayJSTISO() : (recordDate ?? "-")}
          </span>
        )}
      </div>

      {/* 担当医 */}
      {canEdit && !isFinalized ? (
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className={`h-8 shrink-0 text-sm gap-1 px-2 max-w-[160px] ${C.hoverBgPage} ${C.text} border-none`}
          onClick={onStaffClick}
          aria-label={`担当医: ${staffName}`}
        >
          <span className={`text-xs ${C.text50} mr-0.5 shrink-0`}>担当医</span>
          <span className="truncate">{staffName}</span>
          <ChevronDown className={`${ICON.sm} ${C.text40} shrink-0`} aria-hidden="true" />
        </Button>
      ) : (
        <div className="flex flex-col gap-0 shrink-0 max-w-[160px]">
          <span className={`text-xs ${C.text50}`}>担当医</span>
          <span className={`h-8 flex items-center text-sm ${C.text} truncate`}>{staffName}</span>
        </div>
      )}

      {/* 次回予定 */}
      {!isNewRecord ? (
        <NextVisitButton
          value={nextVisitDate}
          onChange={onNextVisitDatePatch}
          onValidationChange={onNextVisitDateValidChange}
          hasLineIntegration={hasLineIntegration}
          disabled={!canEdit || isFinalized}
        />
      ) : null}

      {/* #158: 飼主レポート（別ウィンドウ）。当該飼主・当該ペットを初期選択して開く。 */}
      {canViewReport ? (
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className={`h-8 shrink-0 text-sm gap-1 px-2 ${C.hoverBgPage} ${C.text} border-none`}
          onClick={handleOpenReport}
          aria-label="飼主レポートを開く"
        >
          <FileText className={`${ICON.sm} ${C.text40}`} aria-hidden="true" />
          レポート
        </Button>
      ) : null}
    </>
  );

  return (
    <div className={`sticky top-0 z-10 ${C.bgPage}`}>
      <PatientContextHeader
        ownerName={selectedPet.ownerName}
        petName={selectedPet.name}
        petNumber={selectedPet.petNumber || selectedPet.id}
        weight={selectedPet.weight ?? undefined}
        status={selectedPet.status === "死亡" ? "deceased" : "alive"}
        birthDate={selectedPet.birthDate ?? undefined}
        species={selectedPet.species}
        gender={selectedPet.gender}
        neuteredDate={selectedPet.neuteredDate}
        breed={selectedPet.breed}
        insuranceName={selectedPet.insuranceName ?? undefined}
        insuranceDetails={selectedPet.insuranceDetails ?? undefined}
        visitCount={visitCount}
        onOwnerClick={!isNewRecord && canEdit && !isFinalized ? onOwnerClick : undefined}
        contextControls={contextControls}
      />
      {!isNewRecord && cohabitingPets.length > 0 ? (
        <CohabitingPetChips pets={cohabitingPets} />
      ) : null}
      <div className={`flex shrink-0 overflow-x-auto ${C.bgPage}`}>
        <UnifiedTabsList items={tabs} />
      </div>
    </div>
  );
}
