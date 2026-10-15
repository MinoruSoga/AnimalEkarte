import { memo, useCallback } from "react";
import { Link } from "react-router";
import { ChevronDown, FileText, PawPrint } from "lucide-react";
import { Tooltip } from "@/components/ui/tooltip";
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { DatePicker } from "@/components/shared/DatePicker";
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
import { isPetDeceasedForClinicalWrite } from "@/lib/transforms/pet";
import type { Pet } from "@/types";
import { NextVisitButton } from "./NextVisitButton";
import { VisitTypeSelect } from "./VisitTypeSelect";
import { calcAgePartsAt } from "@/lib/calc-age";
import { isMedicalRecordFinalizedStatus } from "../lib/medical-record-lock";
import { useGetVitals } from "../api/vitals";
import { latestVisitVitalChips } from "../lib/visit-vital-chips";

// EMR-412 R1/X13: カルテ用の属性表示。PatientContextHeader を復活させず、
// PatientInfoCard 内にカルテ専用の属性を描く（年齢・性別・避妊去勢・品種）。
// 欠損は推測しない（避妊去勢未記録は「—」）。
function formatNeuteredStatus(
  gender: string | undefined,
  neuteredDate: string | undefined,
): string {
  if (!neuteredDate) return "—";
  if (gender === "雄") return "去勢済";
  if (gender === "雌") return "避妊済";
  return "避妊・去勢済";
}

function buildPetInfoText(
  birthDate: string | undefined,
  species: string | undefined,
  now: Date,
): string {
  const parts: string[] = [];
  if (birthDate) {
    const { years, months } = calcAgePartsAt(birthDate, now);
    // PR679 F3: 1歳未満は「0歳Nヶ月」ではなく「Nヶ月」だけを表示する（旧仕様の復元）
    parts.push(
      years >= 1 ? `${birthDate}生（${years}歳${months}ヶ月）` : `${birthDate}生（${months}ヶ月）`,
    );
  }
  if (species) parts.push(species);
  return parts.join(" / ");
}

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
  onVisitTypeChange: (value: string) => void;
  onStaffClick: () => void;
  onOwnerClick: () => void;
  onDateChange?: (date: string) => void;
  hasLineIntegration?: boolean;
  medicalRecordId?: string;
  recordClinicId?: string;
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
                "inline-flex min-h-11 shrink-0 items-center rounded-md border bg-white px-2.5 text-sm whitespace-nowrap transition-colors",
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

/**
 * EMR-412 R1: カルテ用の患者属性行（生年月日＋年齢 / 種、性別、避妊去勢、品種、体重）。
 * PatientInfoCard の属性領域に描画する。
 * 欠損値は推測しない。
 */
const PetAttributesLine = memo(function PetAttributesLine({ pet }: { pet: Pet }) {
  const petInfoText = buildPetInfoText(pet.birthDate ?? undefined, undefined, new Date());
  // PR679 F4: !== undefined ではなく実値（非空・trim）で判定する。
  // transform が空文字を素通しするため breed:"" だけでも行が出てしまっていた。
  const hasPetAttributes = Boolean(
    pet.gender?.trim() || pet.neuteredDate?.trim() || pet.breed?.trim(),
  );
  const neuteredStatus = formatNeuteredStatus(pet.gender, pet.neuteredDate);
  const breedText = pet.breed || "—";
  return (
    <div className={`contents text-sm ${C.text60}`}>
      {petInfoText ? (
        <Tooltip content={petInfoText} className="min-w-0 max-w-full">
          <span className="flex items-center gap-1 min-w-0">
            <PawPrint className={`${ICON.xs} shrink-0`} aria-hidden="true" />
            <span className="min-w-0">{petInfoText}</span>
          </span>
        </Tooltip>
      ) : null}
      {hasPetAttributes ? (
        <>
          <span className="flex items-center gap-1 whitespace-nowrap">
            <span className={`text-2xs ${C.text65}`}>性別</span>
            <span className={`font-medium ${C.text}`}>{pet.gender || "—"}</span>
          </span>
          <span className="flex items-center gap-1 whitespace-nowrap">
            <span className={`text-2xs ${C.text65}`}>避妊去勢</span>
            <span className={`font-medium ${C.text}`}>{neuteredStatus}</span>
          </span>
          <Tooltip content={`品種: ${breedText}`} className="min-w-0 max-w-[180px]">
            <span className="flex items-center gap-1 min-w-0">
              <span className={`text-2xs ${C.text65} shrink-0`}>品種</span>
              <span className={`font-medium ${C.text} truncate`}>{breedText}</span>
            </span>
          </Tooltip>
        </>
      ) : null}
    </div>
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
  onVisitTypeChange,
  onStaffClick,
  onOwnerClick,
  onDateChange,
  hasLineIntegration,
  medicalRecordId,
  recordClinicId,
}: MedicalRecordStickyHeaderProps) {
  const { data: visitVitals } = useGetVitals(medicalRecordId ?? "", recordClinicId);
  const vitalsSummary = medicalRecordId ? latestVisitVitalChips(visitVitals) : null;
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

  const scrollingControls = (
    <>
      {/* 来院種別 */}
      <VisitTypeSelect
        value={visitType}
        onChange={onVisitTypeChange}
        disabled={!canEdit || isFinalized}
      />

      {/* 診療日 */}
      <div className="flex flex-col gap-0 shrink-0 min-w-[110px]">
        <span className={`text-xs ${C.text50}`}>診療日</span>
        {canEditDate ? (
          <DatePicker
            value={dateInputValue ?? ""}
            ariaLabel="診療日"
            showClearButton={false}
            onChange={(value) => {
              if (value && value !== dateInputValue) onDateChange!(value);
            }}
          />
        ) : (
          <span className={`h-8 flex items-center text-sm ${C.text}`}>
            {isNewRecord ? todayJSTISO() : (recordDate ?? "-")}
          </span>
        )}
      </div>

      {/* 次回予定（EMR-308）: 実予約 > tentative > 未設定を表示し、Popover は
          tentative appointment を upsert/除去する。カルテ確定状態・medical-records
          編集権限では無効化しない（tentative はペット属性）。reservations:create
          権限と死亡ペットガードはコントロール内部で判定する。 */}
      {!isNewRecord ? (
        <NextVisitButton
          petId={selectedPet.id}
          medicalRecordId={medicalRecordId}
          clinicId={recordClinicId}
          isPetDeceased={isPetDeceasedForClinicalWrite(selectedPet)}
          hasLineIntegration={hasLineIntegration}
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

  // EMR-412 K1: 確定バッジは pinned グループの先頭へ。操作欄は flex-wrap で折り返す。
  const pinnedControls = (
    <>
      {/* SPEC-GAP: 確定済みバッジ。臨床記録の真正性担保のため、確定状態を常時明示する */}
      {!isNewRecord && isFinalized ? (
        <StatusBadge colorClass={getMedicalRecordStatusColor("確定済")}>確定済</StatusBadge>
      ) : null}

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
    </>
  );

  return (
    <div className={`sticky top-0 z-10 ${C.bgPage}`}>
      <PatientInfoCard
        ownerName={selectedPet.ownerName}
        petName={selectedPet.name}
        petNumber={selectedPet.petNumber || selectedPet.id}
        weight={selectedPet.weight ?? ""}
        status={isPetDeceasedForClinicalWrite(selectedPet) ? "deceased" : "alive"}
        birthDate={selectedPet.birthDate ?? undefined}
        insuranceName={selectedPet.insuranceName ?? undefined}
        insuranceDetails={selectedPet.insuranceDetails ?? undefined}
        visitCount={visitCount}
        ownerIsDangerous={selectedPet.ownerIsDangerous}
        petDangerLevel={selectedPet.dangerLevel}
        petDangerReason={selectedPet.dangerReason}
        hideStaff
        hideReservationType
        sticky={false}
        onOwnerClick={!isNewRecord && canEdit && !isFinalized ? onOwnerClick : undefined}
        ownerDetailHref={
          selectedPet.ownerId ? paths.owners.detail.getHref(selectedPet.ownerId) : undefined
        }
        petDetailHref={
          selectedPet.ownerId && selectedPet.id
            ? paths.owners.detail.pet.getHref(selectedPet.ownerId, selectedPet.id)
            : undefined
        }
        petDetails={selectedPet.species || undefined}
        petAttributes={<PetAttributesLine pet={selectedPet} />}
        microchipNumber={selectedPet.microchipNumber?.trim() || undefined}
      />
      {/* EMR-412 K1/X13: 確定済・来院種別・診療日・次回予定・レポート・担当医の操作欄。flex-wrap で折返す。 */}
      <div className={`flex flex-wrap items-center gap-2 px-4 pb-1.5 ${C.bgPage}`}>
        <div className="flex flex-wrap items-center gap-2 min-w-0">{scrollingControls}</div>
        <div className="flex items-center gap-2 shrink-0 ml-auto">{pinnedControls}</div>
      </div>
      {vitalsSummary ? (
        <div className={`px-4 pb-1.5 -mt-1 flex items-center gap-2 text-sm ${C.text60}`}>
          <span
            aria-label="今回のバイタル"
            className="flex items-center flex-wrap gap-x-2 gap-y-0.5"
          >
            {vitalsSummary.temperature != null ? (
              <span className="whitespace-nowrap">T {vitalsSummary.temperature}</span>
            ) : null}
            {vitalsSummary.heartRate != null ? (
              <span className="whitespace-nowrap">HR {vitalsSummary.heartRate}</span>
            ) : null}
            {vitalsSummary.respirationRate != null ? (
              <span className="whitespace-nowrap">RR {vitalsSummary.respirationRate}</span>
            ) : null}
            {vitalsSummary.weight != null ? (
              <span className="whitespace-nowrap">
                測定体重 {vitalsSummary.weight}
                {vitalsSummary.weightUnit ?? ""}
              </span>
            ) : null}
          </span>
        </div>
      ) : null}
      {!isNewRecord && cohabitingPets.length > 0 ? (
        <CohabitingPetChips pets={cohabitingPets} />
      ) : null}
      <div className={`flex shrink-0 overflow-x-auto ${C.bgPage}`}>
        <UnifiedTabsList items={tabs} />
      </div>
    </div>
  );
}
