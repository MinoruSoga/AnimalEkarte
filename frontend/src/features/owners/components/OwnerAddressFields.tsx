import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { DatePicker } from "@/components/shared/DatePicker";
import { C, STYLE } from "@/lib/design-tokens";
import { toJSTWallDate } from "@/lib/jst-date";
import { calculateSaigram } from "@/lib/saigram";
import type { OwnerFieldSectionProps } from "./owner-info-field-shared";

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
  const saigram = calculateSaigram(ownerData.birthDate);

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
        <div className="flex flex-col gap-1.5">
          <DatePicker
            id="birthDate"
            value={ownerData.birthDate}
            onChange={(value) => onChange("birthDate", value)}
            placeholder="生年月日を選択…"
            disabledDays={{ after: toJSTWallDate(new Date()) }}
            className="w-full min-w-[12rem]"
          />
          <span
            role="status"
            aria-live="polite"
            aria-atomic="true"
            className={`text-sm ${C.text60}`}
          >
            {saigram ? (
              <>
                サイグラム: {saigram.typeNo}（{saigram.typeCode}）{saigram.ha}
              </>
            ) : (
              <span className="sr-only">サイグラム分類なし</span>
            )}
          </span>
        </div>
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
