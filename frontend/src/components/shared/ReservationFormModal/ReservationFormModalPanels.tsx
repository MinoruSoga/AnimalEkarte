import { memo, useCallback } from "react";
import { Calendar, CalendarCheck, PawPrint, Search, UserPlus, Users, X } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { ReservationRouteSelect } from "@/components/shared/ReservationRouteSelect";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import type { NewOwnerFormData } from "@/types/reservation-form";
import type { ReservationRoute } from "@/types/reservation-route";
import { C, ICON, PALETTE } from "@/lib/design-tokens";
import { cn } from "@/lib/utils";
import type { Pet, Reservation } from "@/types";
import { NewOwnerInlineForm } from "./NewOwnerInlineForm";
import { PatientSelectionTable } from "./PatientSelectionTable";
import { ReservationFormFields } from "./ReservationFormFields";

export type OwnerMode = "existing" | "new";
export type MobilePanel = "search" | "form";
type LstepStatus = "synced" | "not-linked" | "opt-out" | undefined;

interface ReservationModalHeaderProps {
  isEditMode: boolean;
  mobilePanel: MobilePanel;
  onMobilePanelChange: (panel: MobilePanel) => void;
  descriptionId: string;
}

export function ReservationModalHeader({
  isEditMode,
  mobilePanel,
  onMobilePanelChange,
  descriptionId,
}: ReservationModalHeaderProps) {
  return (
    <DialogHeader className="p-4 border-b shrink-0 h-auto flex flex-col gap-3 space-y-0">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {isEditMode ? (
            <CalendarCheck className={`${ICON.page} ${C.textNotice}`} />
          ) : (
            <Calendar className={`${ICON.page} ${C.textBrand}`} />
          )}
          <DialogTitle className={`text-sm font-bold ${C.text}`}>
            {isEditMode ? "予約編集" : "新規予約作成"}
          </DialogTitle>
        </div>
        <DialogDescription id={descriptionId} className="sr-only">
          左側のリストからペットを選択し、右側のフォームで予約情報を入力してください
        </DialogDescription>
      </div>
      <div className="flex items-center gap-6">
        <StepIndicator step={1} label="患者選択" active={mobilePanel === "search"} />
        <StepIndicator step={2} label="予約情報" active={mobilePanel === "form"} />
      </div>
      <div className="flex gap-2 lg:hidden border-t pt-3 -mx-4 px-4">
        <Button
          size="sm"
          variant={mobilePanel === "search" ? "default" : "outline"}
          onClick={() => onMobilePanelChange("search")}
          className="flex-1 h-9 text-sm"
        >
          患者選択
        </Button>
        <Button
          size="sm"
          variant={mobilePanel === "form" ? "default" : "outline"}
          onClick={() => onMobilePanelChange("form")}
          className="flex-1 h-9 text-sm"
        >
          予約情報
        </Button>
      </div>
    </DialogHeader>
  );
}

const StepIndicator = memo(function StepIndicator({
  step,
  label,
  active,
}: {
  step: number;
  label: string;
  active: boolean;
}) {
  return (
    <div className={`flex items-center gap-1.5 text-xs ${active ? C.textBrand : C.text30}`}>
      <span
        className={`w-5 h-5 rounded-full flex items-center justify-center text-2xs font-bold transition-colors ${
          active ? `${C.bgBrand} ${C.textOnBrand}` : `${C.bgPrimary10} ${C.text30}`
        }`}
      >
        {step}
      </span>
      {label}
    </div>
  );
});

interface ReservationPatientPanelProps {
  isEditMode: boolean;
  ownerMode: OwnerMode;
  mobilePanel: MobilePanel;
  selectedPets: Pet[];
  newOwnerData: NewOwnerFormData;
  newOwnerErrors: Record<string, string>;
  onOwnerModeChange: (mode: OwnerMode) => void;
  onPetSelect: (pet: Pet) => void;
  onNewOwnerChange: (data: NewOwnerFormData) => void;
}

export function ReservationPatientPanel({
  isEditMode,
  ownerMode,
  mobilePanel,
  selectedPets,
  newOwnerData,
  newOwnerErrors,
  onOwnerModeChange,
  onPetSelect,
  onNewOwnerChange,
}: ReservationPatientPanelProps) {
  return (
    <div
      className={cn(
        `w-full lg:w-7/12 border-b lg:border-b-0 lg:border-r ${C.bgSubtle} p-4 flex flex-col overflow-hidden min-h-[300px] lg:min-h-auto flex-1`,
        mobilePanel !== "search" && "hidden lg:flex",
      )}
    >
      {!isEditMode ? (
        <div className="flex rounded-lg border overflow-hidden mb-3 shrink-0">
          <button
            type="button"
            onClick={() => onOwnerModeChange("existing")}
            data-testid="mode-existing"
            className={`flex-1 px-3 py-1.5 text-xs font-medium transition-colors ${ownerMode === "existing" ? `${C.bgBrand} ${C.textOnBrand}` : `bg-white ${C.text60}`}`}
          >
            <Users size={12} className="inline mr-1" />既存飼主
          </button>
          <button
            type="button"
            onClick={() => onOwnerModeChange("new")}
            data-testid="mode-new"
            className={`flex-1 px-3 py-1.5 text-xs font-medium transition-colors ${ownerMode === "new" ? `${C.bgBrand} ${C.textOnBrand}` : `bg-white ${C.text60}`}`}
          >
            <UserPlus size={12} className="inline mr-1" />新規飼主
          </button>
        </div>
      ) : null}

      {ownerMode === "existing" ? (
        <>
          <div className="mb-3 flex items-center gap-2 shrink-0">
            <Search className={`${ICON.action} ${C.text60}`} />
            <Label className={`text-sm font-bold ${C.text}`}>患者検索</Label>
          </div>
          <div className="flex-1 overflow-hidden flex flex-col min-h-0">
            <PatientSelectionTable
              onSelect={onPetSelect}
              selectedPets={selectedPets}
            />
          </div>
        </>
      ) : (
        <NewOwnerInlineForm
          value={newOwnerData}
          onChange={onNewOwnerChange}
          errors={newOwnerErrors}
        />
      )}
    </div>
  );
}

interface ReservationDetailsPanelProps {
  mobilePanel: MobilePanel;
  selectedPets: Pet[];
  lstepStatus: LstepStatus;
  formData: Partial<Reservation>;
  validationErrors: Record<string, string>;
  holidayDates: Set<string>;
  isEditMode: boolean;
  reservationId?: string;
  canEdit: boolean;
  onSelectedPetsChange: (updater: (prev: Pet[]) => Pet[]) => void;
  onFormChange: (data: Partial<Reservation>) => void;
  onClearError: (field: string) => void;
  onMonthChange: (yearMonth: string) => void;
}

export function ReservationDetailsPanel({
  mobilePanel,
  selectedPets,
  lstepStatus,
  formData,
  validationErrors,
  holidayDates,
  isEditMode,
  reservationId,
  canEdit,
  onSelectedPetsChange,
  onFormChange,
  onClearError,
  onMonthChange,
}: ReservationDetailsPanelProps) {
  return (
    <div
      className={cn(
        "w-full lg:w-5/12 bg-white flex flex-col overflow-hidden min-h-[300px] lg:min-h-auto flex-1",
        mobilePanel !== "form" && "hidden lg:flex",
      )}
    >
      <div className="flex-1 overflow-y-auto p-4 flex flex-col gap-4">
        <SelectedPatientSummary
          selectedPets={selectedPets}
          validationError={validationErrors.patient}
          onSelectedPetsChange={onSelectedPetsChange}
        />

        <LineStatusNotice status={lstepStatus} />

        <div className="space-y-4">
          <Label className={`text-sm font-bold ${C.text}`}>予約詳細</Label>
          <ReservationFormFields
            formData={formData}
            onChange={onFormChange}
            validationErrors={validationErrors}
            onClearError={onClearError}
            holidayDates={holidayDates}
            onMonthChange={onMonthChange}
          />
          {isEditMode && reservationId ? (
            <ReservationRouteSelect
              reservationId={reservationId}
              value={(formData.reservationRoute as ReservationRoute | undefined) ?? null}
              disabled={!canEdit}
            />
          ) : null}
        </div>
      </div>
    </div>
  );
}

function SelectedPatientSummary({
  selectedPets,
  validationError,
  onSelectedPetsChange,
}: {
  selectedPets: Pet[];
  validationError?: string;
  onSelectedPetsChange: (updater: (prev: Pet[]) => Pet[]) => void;
}) {
  const handleRemovePet = useCallback(
    (petId: string) => {
      onSelectedPetsChange((prev) => prev.filter((item) => item.id !== petId));
    },
    [onSelectedPetsChange],
  );

  return (
    <div className={`rounded-lg border p-3 transition-colors ${selectedPets.length > 0 ? `${C.bgBrandLight50} ${C.borderBrandLight}` : `${C.bgPage} ${C.borderMediumLight}`}`}>
      <Label className={`text-2xs ${C.text40} font-semibold uppercase block mb-3`}>
        予約対象（選択中）
        <span style={{ color: C.danger }} className="ml-1 normal-case" aria-hidden="true">*</span>
      </Label>

      {selectedPets.length > 0 ? (
        <div className="flex flex-col gap-2">
          {selectedPets.map((pet) => (
            <SelectedPetChip
              key={pet.id}
              pet={pet}
              onRemove={handleRemovePet}
            />
          ))}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center h-20 text-center">
          <PawPrint className={`${ICON.lg} ${C.text15} mb-2`} />
          <div className={`text-xs ${C.text40}`}>左側から患者を選択してください</div>
        </div>
      )}
      {selectedPets.length === 0 ? <FormFieldError message={validationError} /> : null}
    </div>
  );
}

const SelectedPetChip = memo(function SelectedPetChip({
  pet,
  onRemove,
}: {
  pet: Pet;
  onRemove: (petId: string) => void;
}) {
  return (
    <div className={`flex items-center gap-2 bg-white p-2 rounded-lg border ${C.borderMediumLight}`}>
      <PawPrint className={`${ICON.action} ${C.text60} flex-shrink-0`} />
      <span className={`text-sm font-bold ${C.text}`}>{pet.name}</span>
      <Badge variant="outline" className={`text-2xs font-normal ${C.text60} ${C.bgPage} ${C.borderMediumLight} h-5`}>
        {pet.species}
      </Badge>
      <span className={`text-2xs ${C.text60} ml-auto`}>
        No. {pet.ownerId} {pet.ownerName}
      </span>
      <button
        type="button"
        onClick={() => onRemove(pet.id)}
        className={`ml-1 p-1 ${C.hoverBgDanger5} rounded transition-colors`}
      >
        <X className={`${ICON.action} ${C.danger} ${C.hoverTextDanger}`} />
      </button>
    </div>
  );
});

function LineStatusNotice({ status }: { status: LstepStatus }) {
  if (status === "not-linked") {
    return (
      <div className={`rounded-xs border ${C.borderNotice} ${C.bgNotice40} px-3 py-2 text-xs ${C.textNotice}`}>
        この飼い主はLINEアカウントが未連携のため、予約確定後のLINE自動通知は送信されません。
      </div>
    );
  }
  if (status === "opt-out") {
    return (
      <div className={`rounded-xs border ${C.borderMediumLight} ${C.bgPage30} px-3 py-2 text-xs ${C.text40}`}>
        この飼い主はLINEメッセージの受信を拒否しています。予約確定後のLINE自動通知は送信されません。
      </div>
    );
  }
  if (status === "synced") {
    return (
      <div
        className="rounded-xs border px-3 py-2 text-xs text-white flex items-center gap-1.5"
        style={{ backgroundColor: PALETTE.lineGreen, borderColor: PALETTE.lineGreen }}
      >
        LINE連携済み — 予約確定後に自動通知が送信されます。
      </div>
    );
  }
  return null;
}

interface ReservationModalFooterProps {
  ownerMode: OwnerMode;
  selectedPetsCount: number;
  isEditMode: boolean;
  canSave: boolean;
  isSubmitting?: boolean;
  onClose: () => void;
}

export function ReservationModalFooter({
  ownerMode,
  selectedPetsCount,
  isEditMode,
  canSave,
  isSubmitting = false,
  onClose,
}: ReservationModalFooterProps) {
  return (
    <DialogFooter className="p-4 border-t bg-white shrink-0 h-14 flex items-center justify-between gap-2">
      <div className="flex items-center gap-1.5">
        {ownerMode === "new" ? (
          <>
            <UserPlus className={`${ICON.action} ${C.text60}`} />
            <span className={`text-sm ${C.text60}`}>新規飼主モード</span>
          </>
        ) : (
          <>
            <PawPrint className={`${ICON.action} ${C.text60}`} />
            <span className={`text-sm ${C.text60}`}>{selectedPetsCount}頭 選択中</span>
          </>
        )}
      </div>
      <div className="flex gap-2">
        <Button variant="outline" onClick={onClose} className="h-10 text-sm" disabled={isSubmitting}>
          キャンセル
        </Button>
        {canSave ? (
          // FE-RC-023: 手組みの pending/disabled ロジックを共有 SubmitButton（useFormStatus 連動）へ統一
          <SubmitButton loadingText="送信中...">
            {isEditMode ? "更新する" : "予約を確定"}
          </SubmitButton>
        ) : null}
      </div>
    </DialogFooter>
  );
}
