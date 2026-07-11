import { memo } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { DatePicker } from "@/components/shared/DatePicker";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { C, STYLE } from "@/lib/design-tokens";
import { toJSTWallDate } from "@/lib/jst-date";
import type { ClinicMembership } from "@/types/auth";
import { MEMBERSHIP_TYPE_VALUES, type MembershipTypeLabel, type OwnerData } from "../types";

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
          // docs/DESIGN_SYSTEM.md: 構造色は brand teal #038B94 のみ（旧 accent ブルーから移行）
          className={
            value === type
              ? `${C.bgBrand} ${C.hoverBgBrand} ${C.textWhite} h-11 text-sm px-3 rounded-full transition-colors shadow-none border-transparent`
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

interface OwnerFieldSectionProps {
  ownerData: OwnerData;
  fieldErrors: Record<string, string>;
  onChange: (field: string, value: string | boolean | number | null | undefined) => void;
  onClearError: (field: string) => void;
}

interface OwnerBasicFieldsProps extends OwnerFieldSectionProps {
  isEdit: boolean;
  /** #84: 登録先医院の選択肢（ユーザー所属医院）。2件以上かつ新規登録時のみセレクトを表示 */
  clinicOptions?: ClinicMembership[];
  /** #84: 未選択時に表示する現在の医院ID */
  currentClinicId?: string | null;
}

export function OwnerBasicFields({
  ownerData,
  fieldErrors,
  isEdit,
  clinicOptions,
  currentClinicId,
  onChange,
  onClearError,
}: OwnerBasicFieldsProps) {
  // #84 Q12=A: 医院指定は登録フォームのみ。単一所属ユーザーには表示しない
  const showClinicSelect = !isEdit && (clinicOptions?.length ?? 0) >= 2;
  return (
    <>
      <div className="space-y-1.5">
        <Label htmlFor="ownerId" className={`text-sm ${C.text60}`}>飼主No</Label>
        {isEdit ? (
          <Input
            id="ownerId"
            type="text"
            value={ownerData.ownerId}
            disabled
            className={`${STYLE.formInput} disabled:opacity-50`}
          />
        ) : (
          <p className={`flex h-9 items-center px-3 text-sm ${C.text40} italic`}>
            登録時に自動採番されます
          </p>
        )}
      </div>

      {showClinicSelect ? (
        <div className="space-y-1.5">
          <Label htmlFor="clinicId" className={`text-sm ${C.text60}`}>登録先医院</Label>
          <Select
            value={ownerData.clinicId ?? currentClinicId ?? undefined}
            onValueChange={(value) => onChange("clinicId", value)}
          >
            <SelectTrigger id="clinicId" data-testid="owner-clinic-select" className={STYLE.formInput}>
              <SelectValue placeholder="医院を選択" />
            </SelectTrigger>
            <SelectContent>
              {clinicOptions?.map((membership) => (
                <SelectItem key={membership.clinicId} value={membership.clinicId}>
                  {membership.clinicName}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      ) : null}

      <div className="space-y-1.5">
        <Label htmlFor="company" className={`text-sm ${C.text60}`}>会社名</Label>
        <Input
          id="company"
          value={ownerData.company}
          onChange={(event) => onChange("company", event.target.value)}
          className={STYLE.formInput}
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="ownerName" className={`text-sm ${C.text60}`}>
          飼主名 <span className={C.textRequired}>*</span>
        </Label>
        <Input
          id="ownerName"
          value={ownerData.ownerName}
          maxLength={100}
          aria-invalid={!!fieldErrors.ownerName}
          aria-describedby={fieldErrors.ownerName ? "ownerName-error" : undefined}
          onChange={(event) => {
            onChange("ownerName", event.target.value);
            onClearError("ownerName");
          }}
          className={`${STYLE.formInput} ${fieldErrors.ownerName ? STYLE.formInputError : ""}`}
        />
        <FormFieldError id="ownerName-error" message={fieldErrors.ownerName} />
      </div>

      <div className="space-y-1.5 col-span-2 lg:col-span-1 lg:row-span-3">
        <Label className={`text-sm ${C.text60}`}>危険人物</Label>
        <div className="flex items-center space-x-2 mb-2 h-10">
          <Switch
            id="dangerous"
            checked={ownerData.isDangerous}
            onCheckedChange={(checked) => onChange("isDangerous", checked)}
            className="origin-left mr-2"
          />
          <label htmlFor="dangerous" className={`text-sm cursor-pointer ${C.text}`}>該当する</label>
        </div>
        <Label htmlFor="remarks" className={`text-sm ${C.text60}`}>備考・特記事項</Label>
        <Textarea
          id="remarks"
          rows={6}
          value={ownerData.remarks}
          maxLength={1000}
          onChange={(event) => onChange("remarks", event.target.value)}
          className={`text-sm ${C.text} min-h-[140px] resize-none ${C.borderMedium} p-3`}
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="ownerNameKana" className={`text-sm ${C.text60}`}>
          飼主名よみ <span className={C.textRequired}>*</span>
        </Label>
        <Input
          id="ownerNameKana"
          placeholder="はやし ふみあき"
          value={ownerData.ownerNameKana}
          aria-invalid={!!fieldErrors.ownerNameKana}
          aria-describedby={fieldErrors.ownerNameKana ? "ownerNameKana-error" : undefined}
          onChange={(event) => {
            onChange("ownerNameKana", event.target.value);
            onClearError("ownerNameKana");
          }}
          className={`${STYLE.formInput} ${fieldErrors.ownerNameKana ? STYLE.formInputError : ""}`}
        />
        <FormFieldError id="ownerNameKana-error" message={fieldErrors.ownerNameKana} />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="email" className={`text-sm ${C.text60}`}>メールアドレス</Label>
        <Input
          id="email"
          type="email"
          value={ownerData.email}
          aria-invalid={!!fieldErrors.email}
          aria-describedby={fieldErrors.email ? "email-error" : undefined}
          onChange={(event) => {
            onChange("email", event.target.value);
            onClearError("email");
          }}
          className={`${STYLE.formInput} ${fieldErrors.email ? STYLE.formInputError : ""}`}
        />
        <FormFieldError id="email-error" message={fieldErrors.email} />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="phone" className={`text-sm ${C.text60}`}>
          電話番号 <span className={C.textRequired}>*</span>
        </Label>
        <Input
          id="phone"
          placeholder="090-1234-5678"
          value={ownerData.phone}
          aria-invalid={!!fieldErrors.phone}
          aria-describedby={fieldErrors.phone ? "phone-error" : undefined}
          onChange={(event) => {
            onChange("phone", event.target.value);
            onClearError("phone");
          }}
          className={`${STYLE.formInput} ${fieldErrors.phone ? STYLE.formInputError : ""}`}
        />
        <FormFieldError id="phone-error" message={fieldErrors.phone} />
      </div>

      <div className="space-y-1.5 col-span-1 lg:col-span-2">
        <Label htmlFor="companyPhone" className={`text-sm ${C.text60}`}>会社 電話番号</Label>
        <Input
          id="companyPhone"
          value={ownerData.companyPhone}
          onChange={(event) => onChange("companyPhone", event.target.value)}
          className={STYLE.formInput}
        />
      </div>
    </>
  );
}

interface OwnerAddressFieldsProps extends OwnerFieldSectionProps {
  onPostalCodeLookup: (postalCodeField: string, addressField: string) => void;
}

export function OwnerAddressFields({
  ownerData,
  fieldErrors,
  onChange,
  onClearError,
  onPostalCodeLookup,
}: OwnerAddressFieldsProps) {
  return (
    <>
      <div className="space-y-1.5">
        <Label htmlFor="postalCode" className={`text-sm ${C.text60}`}>郵便番号</Label>
        <div className="flex gap-1.5">
          <Input
            id="postalCode"
            placeholder="123-4567"
            value={ownerData.postalCode}
            aria-invalid={!!fieldErrors.postalCode}
            aria-describedby={fieldErrors.postalCode ? "postalCode-error" : undefined}
            onChange={(event) => {
              onChange("postalCode", event.target.value);
              onClearError("postalCode");
            }}
            className={`${STYLE.formInput} ${fieldErrors.postalCode ? STYLE.formInputError : ""}`}
          />
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="shrink-0 h-9 text-xs px-2"
            onClick={() => onPostalCodeLookup("postalCode", "address1")}
          >
            検索
          </Button>
        </div>
        <FormFieldError id="postalCode-error" message={fieldErrors.postalCode} />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="address1" className={`text-sm ${C.text60}`}>住所1（会社）</Label>
        <Input
          id="address1"
          value={ownerData.address1}
          onChange={(event) => onChange("address1", event.target.value)}
          maxLength={200}
          className={STYLE.formInput}
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="homePostalCode" className={`text-sm ${C.text60}`}>郵便番号(自宅)</Label>
        <div className="flex gap-1.5">
          <Input
            id="homePostalCode"
            placeholder="123-4567"
            value={ownerData.homePostalCode || ""}
            aria-invalid={!!fieldErrors.homePostalCode}
            aria-describedby={fieldErrors.homePostalCode ? "homePostalCode-error" : undefined}
            onChange={(event) => {
              onChange("homePostalCode", event.target.value);
              onClearError("homePostalCode");
            }}
            className={`${STYLE.formInput} ${fieldErrors.homePostalCode ? STYLE.formInputError : ""}`}
          />
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="shrink-0 h-9 text-xs px-2"
            onClick={() => onPostalCodeLookup("homePostalCode", "homeAddress1")}
          >
            検索
          </Button>
        </div>
        <FormFieldError id="homePostalCode-error" message={fieldErrors.homePostalCode} />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="address2" className={`text-sm ${C.text60}`}>住所2（会社）</Label>
        <Input
          id="address2"
          value={ownerData.address2}
          onChange={(event) => onChange("address2", event.target.value)}
          maxLength={200}
          className={STYLE.formInput}
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="homeAddress1" className={`text-sm ${C.text60}`}>住所1(自宅)</Label>
        <Input
          id="homeAddress1"
          value={ownerData.homeAddress1}
          onChange={(event) => onChange("homeAddress1", event.target.value)}
          maxLength={200}
          className={STYLE.formInput}
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="birthDate" className={`text-sm ${C.text60}`}>飼主生年月日</Label>
        <DatePicker
          id="birthDate"
          value={ownerData.birthDate}
          onChange={(value) => onChange("birthDate", value)}
          placeholder="生年月日を選択…"
          disabledDays={{ after: toJSTWallDate(new Date()) }}
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="homeAddress2" className={`text-sm ${C.text60}`}>住所2(自宅)</Label>
        <Input
          id="homeAddress2"
          value={ownerData.homeAddress2}
          onChange={(event) => onChange("homeAddress2", event.target.value)}
          maxLength={200}
          className={STYLE.formInput}
        />
      </div>
    </>
  );
}

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
