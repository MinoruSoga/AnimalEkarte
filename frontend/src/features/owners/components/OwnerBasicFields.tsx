import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { C, STYLE } from "@/lib/design-tokens";
import type { ClinicMembership } from "@/types/auth";
import type { OwnerFieldSectionProps } from "./owner-info-field-shared";

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

      <div className="space-y-1.5 col-span-1 sm:col-span-2 lg:col-span-1 lg:row-span-3">
        <Label className={`text-sm ${C.text60}`}>危険人物</Label>
        <div className="flex items-center space-x-2 mb-2 h-10">
          <Switch
            id="dangerous"
            aria-label="危険人物に該当する"
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
