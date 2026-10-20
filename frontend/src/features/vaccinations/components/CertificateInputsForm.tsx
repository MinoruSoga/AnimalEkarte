import { MoneyInput } from "@/components/shared/SidePeek";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { C } from "@/lib/design-tokens";
import {
  CERTIFICATE_FEE_KEYS,
  type CertificateFeeKey,
  type CertificateInputDraft,
  type DraftErrors,
} from "../lib/certificate-inputs";
import { BODY_SIZE_LABEL } from "../lib/certificate-messages";

interface CertificateInputsFormProps {
  draft: CertificateInputDraft;
  errors: DraftErrors;
  disabled: boolean;
  onChange: (patch: Partial<CertificateInputDraft>) => void;
}

/** Radix Select は空文字の value を持てないため、未選択の番兵値を使う。 */
const BODY_SIZE_UNSET = "unset";

const FEE_LABEL: Record<CertificateFeeKey, string> = {
  registration_fee: "登録手数料",
  injection_fee: "注射料",
  tag_fee: "注射済票費用",
  total: "合計",
};

/**
 * 狂犬病予防注射済証の職員入力 10 欄（すべて任意）。
 * - 料金は未入力（null）と 0 を区別する。合計は独立入力で、自動計算も補正もしない。
 * - 初期値は空欄。会計・マスタからの転記はしない。
 */
export function CertificateInputsForm({ draft, errors, disabled, onChange }: CertificateInputsFormProps) {
  const text = (id: string, label: string, key: keyof CertificateInputDraft) => (
    <div className="flex flex-col">
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        value={draft[key]}
        disabled={disabled}
        autoComplete="off"
        onChange={(e) => onChange({ [key]: e.target.value } as Partial<CertificateInputDraft>)}
      />
    </div>
  );
  return (
    <fieldset className="grid gap-4 sm:grid-cols-2" disabled={disabled}>
      <legend className={`mb-2 text-sm ${C.text60}`}>
        すべて任意です。未入力のまま発行できます。
      </legend>
      {text("cert-number", "証明番号", "certificate_number")}
      {text("cert-injection-tag", "注射済票番号", "injection_tag_number")}
      {text("cert-dog-tag", "犬鑑札番号", "dog_registration_tag_number")}
      <div className="flex flex-col">
        <Label htmlFor="cert-body-size">体格</Label>
        <Select
          value={draft.body_size === "" ? BODY_SIZE_UNSET : draft.body_size}
          disabled={disabled}
          onValueChange={(v) =>
            onChange({ body_size: v === BODY_SIZE_UNSET ? "" : (v as CertificateInputDraft["body_size"]) })
          }
        >
          <SelectTrigger id="cert-body-size">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={BODY_SIZE_UNSET}>未選択</SelectItem>
            {Object.entries(BODY_SIZE_LABEL).map(([value, label]) => (
              <SelectItem key={value} value={value}>
                {label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div className="flex flex-col sm:col-span-2">
        <Label htmlFor="cert-features">特徴</Label>
        <Textarea
          id="cert-features"
          value={draft.distinguishing_features}
          disabled={disabled}
          onChange={(e) => onChange({ distinguishing_features: e.target.value })}
        />
      </div>
      {text("cert-amount", "注射量", "injection_amount_text")}
      {CERTIFICATE_FEE_KEYS.map((key) => (
        <div key={key} className="flex flex-col">
          <Label htmlFor={`cert-${key}`}>{FEE_LABEL[key]}</Label>
          <MoneyInput
            variant="inline"
            id={`cert-${key}`}
            stringValue={draft[key]}
            onStringChange={(v) => onChange({ [key]: v } as Partial<CertificateInputDraft>)}
            suffix="円"
            align="right"
            placeholder=""
            disabled={disabled}
            error={errors[key]}
          />
        </div>
      ))}
    </fieldset>
  );
}
