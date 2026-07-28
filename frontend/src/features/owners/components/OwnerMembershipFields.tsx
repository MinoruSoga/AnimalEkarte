import { memo } from "react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { C, STYLE } from "@/lib/design-tokens";
import { MEMBERSHIP_TYPE_VALUES, type MembershipTypeLabel } from "../types";
import type { OwnerFieldSectionProps } from "./owner-info-field-shared";

interface MembershipTypeButtonsProps {
  value: MembershipTypeLabel;
  onChange: (type: MembershipTypeLabel) => void;
}

const MembershipTypeButtons = memo(function MembershipTypeButtons({
  value,
  onChange,
}: MembershipTypeButtonsProps) {
  return (
    <div className="flex gap-1.5 flex-wrap">
      {MEMBERSHIP_TYPE_VALUES.map((type) => (
        <Button
          key={type}
          type="button"
          variant={value === type ? "default" : "outline"}
          size="sm"
          // docs/spec/design-system.md: 選択状態は brand と同じ primary teal
          className={
            value === type
              ? `${C.bgActionPrimary} ${C.textOnActionPrimary} ${C.hoverBgActionPrimary} ${C.hoverTextOnActionPrimary} h-11 text-sm px-3 rounded-full transition-colors shadow-none border-transparent`
              : `h-11 text-sm ${C.text} ${C.hoverBgMedium} ${C.borderMedium} px-3`
          }
          onClick={() => onChange(type)}
        >
          {type}
        </Button>
      ))}
    </div>
  );
});

interface OwnerMembershipFieldsProps extends OwnerFieldSectionProps {
  canEditDiscount: boolean;
  onMembershipChange: (type: MembershipTypeLabel) => void;
}

export function OwnerMembershipFields({
  ownerData,
  fieldErrors,
  canEditDiscount,
  onChange,
  onClearError,
  onMembershipChange,
}: OwnerMembershipFieldsProps) {
  return (
    <>
      <div className="space-y-1.5">
        <Label className={`text-sm ${C.text60}`}>会員区分</Label>
        <MembershipTypeButtons value={ownerData.membershipType} onChange={onMembershipChange} />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="dmPreference" className={`text-sm ${C.text60}`}>DM</Label>
        <Select
          value={
            ownerData.dmPreference == null
              ? "unset"
              : ownerData.dmPreference
                ? "required"
                : "unneeded"
          }
          onValueChange={(value) => {
            onChange(
              "dmPreference",
              value === "unset" ? null : value === "required",
            );
          }}
        >
          <SelectTrigger id="dmPreference" className={STYLE.formInput}>
            <SelectValue placeholder="未設定" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="unset">未設定</SelectItem>
            <SelectItem value="required">必要</SelectItem>
            <SelectItem value="unneeded">不要</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="discountRate" className={`text-sm ${C.text60}`}>値引率 (%)</Label>
        <NumberInput
          id="discountRate"
          min={0}
          max={100}
          step={1}
          value={ownerData.discountRate || ""}
          disabled={!canEditDiscount}
          aria-invalid={!!fieldErrors.discountRate}
          aria-describedby={
            fieldErrors.discountRate
              ? "discountRate-error"
              : !canEditDiscount
                ? "discountRate-permission"
                : undefined
          }
          onChange={(value) => {
            onChange("discountRate", Number(value));
            onClearError("discountRate");
          }}
          suffix="%"
          className={`${STYLE.formInput} ${fieldErrors.discountRate ? STYLE.formInputError : ""}`}
        />
        <FormFieldError id="discountRate-error" message={fieldErrors.discountRate} />
        {!canEditDiscount ? (
          <p id="discountRate-permission" className={`text-xs ${C.text50}`}>
            値引率の変更には権限が必要です
          </p>
        ) : null}
      </div>
    </>
  );
}
