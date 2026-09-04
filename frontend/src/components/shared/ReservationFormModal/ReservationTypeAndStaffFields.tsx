import { C } from "@/lib/design-tokens";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { SearchableSelect, type SearchableSelectOption } from "@/components/ui/searchable-select";
import {
  ReservationTypePickerDialog,
  type ReservationTypePickerGroup,
} from "@/components/shared/ReservationFormModal/ReservationTypePickerDialog";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Label } from "@/components/ui/label";
import { ChevronDown } from "lucide-react";
import { cn } from "@/lib/utils";
import { MasterLink } from "@/components/shared/MasterLink";
import { isOneOf } from "@/lib/type-utils";
import type { Reservation } from "@/types";
import { FieldLabel } from "./ReservationDateTimeFields";

const VISIT_TYPE_VALUES = ["first", "revisit"] as const;

interface SelectedReservationType {
  color: string;
  name: string;
  isActive: boolean;
}

interface ReservationTypeAndStaffFieldsProps {
  formData: Partial<Reservation>;
  onChange: (data: Partial<Reservation>) => void;
  validationErrors?: Record<string, string>;
  typePickerOpen: boolean;
  setTypePickerOpen: (open: boolean) => void;
  reservationTypePickerGroups: ReservationTypePickerGroup[];
  selectedReservationType: SelectedReservationType | null;
  staffSelectOptions: SearchableSelectOption[];
  staffEmptyMessage: string;
}

export function ReservationTypeAndStaffFields({
  formData,
  onChange,
  validationErrors,
  typePickerOpen,
  setTypePickerOpen,
  reservationTypePickerGroups,
  selectedReservationType,
  staffSelectOptions,
  staffEmptyMessage,
}: ReservationTypeAndStaffFieldsProps) {
  return (
    <>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <FieldLabel
            required
            trailing={<MasterLink category="reservationType" label="編集" className="text-2xs" />}
          >
            予約区分
          </FieldLabel>
          {/* BUG-341/scroll: サブダイアログ選択(カテゴリチップ + カードリスト)。Dialog 自身の
              scroll-lock で wheel スクロールが確実に動作する(popover-in-dialog の wheel ブロック回避) */}
          <button
            type="button"
            data-testid="res-type-trigger"
            onClick={() => setTypePickerOpen(true)}
            aria-invalid={Boolean(validationErrors?.type)}
            className={cn(
              "flex h-9 w-full items-center justify-between gap-2 rounded-md border bg-white px-3 text-sm transition-colors",
              C.borderMediumLight,
              C.hoverBgSubtle,
              validationErrors?.type && C.borderDanger,
            )}
          >
            {selectedReservationType ? (
              <span className="flex min-w-0 items-center gap-2">
                <span
                  className="size-3 shrink-0 rounded-full"
                  style={{ backgroundColor: selectedReservationType.color }}
                />
                <span className={cn("line-clamp-1", C.text)}>
                  {selectedReservationType.name}
                  {!selectedReservationType.isActive ? (
                    <span className={cn("ml-1 shrink-0 text-2xs", C.text40)}>（無効）</span>
                  ) : null}
                </span>
              </span>
            ) : (
              <span className={C.text40}>選択してください</span>
            )}
            <ChevronDown className={cn("size-4 shrink-0 opacity-50", C.text)} />
          </button>
          <ReservationTypePickerDialog
            open={typePickerOpen}
            onOpenChange={setTypePickerOpen}
            groups={reservationTypePickerGroups}
            selectedId={formData.type || ""}
            onSelect={(id) => onChange({ ...formData, type: id })}
          />
          {validationErrors?.type ? (
            <FormFieldError id="res-type-error" message={validationErrors.type} />
          ) : null}
        </div>
        <div className="space-y-1.5">
          <FieldLabel>来院区分</FieldLabel>
          <RadioGroup
            value={formData.visitType || ""}
            onValueChange={(v: string) => {
              if (isOneOf(v, VISIT_TYPE_VALUES)) {
                onChange({ ...formData, visitType: v });
              }
            }}
            className="flex gap-2 pt-1"
          >
            <div className="flex-1">
              <RadioGroupItem value="first" id="first" className="sr-only" />
              <Label
                htmlFor="first"
                className={cn(
                  `block h-9 rounded-full border-2 px-3 py-1.5 text-center text-sm font-medium cursor-pointer transition-colors ${C.text}`,
                  formData.visitType === "first"
                    ? `${C.borderBrand} ${C.bgBrand8}`
                    : `${C.borderMediumLight} bg-white ${C.hoverBgSubtle}`,
                )}
              >
                初診
              </Label>
            </div>
            <div className="flex-1">
              <RadioGroupItem value="revisit" id="revisit" className="sr-only" />
              <Label
                htmlFor="revisit"
                className={cn(
                  `block h-9 rounded-full border-2 px-3 py-1.5 text-center text-sm font-medium cursor-pointer transition-colors ${C.text}`,
                  formData.visitType === "revisit"
                    ? `${C.borderBrand} ${C.bgBrand8}`
                    : `${C.borderMediumLight} bg-white ${C.hoverBgSubtle}`,
                )}
              >
                再診
              </Label>
            </div>
          </RadioGroup>
        </div>
      </div>

      <div className="space-y-1.5">
        <FieldLabel trailing={<MasterLink category="staff" label="編集" className="text-2xs" />}>
          担当者
        </FieldLabel>
        <SearchableSelect
          value={formData.doctor || ""}
          onValueChange={(v) => onChange({ ...formData, doctor: v })}
          options={staffSelectOptions}
          placeholder="選択してください"
          searchPlaceholder="スタッフ名で検索..."
          emptyMessage={staffEmptyMessage}
          triggerTestId="res-staff-trigger"
        />
      </div>
    </>
  );
}
