import { C } from "@/lib/design-tokens";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import type { CheckupTypeFieldRow, CheckupFieldResultInput } from "@/hooks/use-checkup-fields";

/** 1 フィールド分の入力値（型別。number は文字列で保持し送信時にパースする）。 */
export interface CheckupFieldValue {
  number?: string;
  text?: string;
  bool?: boolean;
  list?: string[];
}

interface DynamicCheckupFieldsProps {
  fields: CheckupTypeFieldRow[];
  values: Record<number, CheckupFieldValue>;
  onChange: (fieldId: number, value: CheckupFieldValue) => void;
}

const inputClass = `w-full rounded border ${C.borderLight} px-3 py-2 text-sm ${C.text}`;

function toggleListValue(list: string[] | undefined, value: string, checked: boolean): string[] {
  const current = list ?? [];
  if (checked) {
    return current.includes(value) ? current : [...current, value];
  }
  return current.filter((v) => v !== value);
}

/**
 * #211 健診パッケージの field_type に応じた動的入力フォーム。
 * 完全に制御コンポーネント（値は親が保持）。レイアウト系プロパティのアニメは行わない。
 */
export function DynamicCheckupFields({ fields, values, onChange }: DynamicCheckupFieldsProps) {
  if (fields.length === 0) return null;

  return (
    <div className="space-y-6" data-testid="dynamic-checkup-fields">
      {fields.map((field) => {
        const value = values[field.id] ?? {};
        const labelId = `checkup-field-${field.id}`;
        // multi_select/checklist は単一 input を持たないため、グループの aria-labelledby が
        // 参照できるよう Label 要素自身に id を付与する（htmlFor は単一 input 用に残す）。
        const fieldLabelId = `${labelId}-label`;
        return (
          <div key={field.id} className="space-y-2">
            <Label htmlFor={labelId} id={fieldLabelId}>
              {field.name}
              {field.unit ? (
                <span className={`ml-1 text-xs ${C.text50}`}>（{field.unit}）</span>
              ) : null}
            </Label>

            {field.fieldType === "boolean" ? (
              <div className={`flex items-center gap-2 text-sm ${C.text}`}>
                <input
                  id={labelId}
                  type="checkbox"
                  checked={value.bool ?? false}
                  onChange={(e) => onChange(field.id, { ...value, bool: e.target.checked })}
                />
                <span className={C.text50}>はい</span>
              </div>
            ) : null}

            {field.fieldType === "number" ? (
              <input
                id={labelId}
                type="number"
                className={inputClass}
                value={value.number ?? ""}
                min={field.minValue}
                max={field.maxValue}
                onChange={(e) => onChange(field.id, { ...value, number: e.target.value })}
              />
            ) : null}

            {field.fieldType === "single_select" ? (
              <select
                id={labelId}
                className={inputClass}
                value={value.text ?? ""}
                onChange={(e) => onChange(field.id, { ...value, text: e.target.value })}
              >
                <option value="">選択してください</option>
                {field.options.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            ) : null}

            {field.fieldType === "multi_select" || field.fieldType === "checklist" ? (
              <div className="flex flex-col gap-1.5" role="group" aria-labelledby={fieldLabelId}>
                {field.options.map((opt) => (
                  <label key={opt.value} className={`flex items-center gap-2 text-sm ${C.text}`}>
                    <input
                      type="checkbox"
                      checked={(value.list ?? []).includes(opt.value)}
                      onChange={(e) =>
                        onChange(field.id, {
                          ...value,
                          list: toggleListValue(value.list, opt.value, e.target.checked),
                        })
                      }
                    />
                    {opt.label}
                  </label>
                ))}
              </div>
            ) : null}

            {field.fieldType === "text" ? (
              <Textarea
                id={labelId}
                value={value.text ?? ""}
                onChange={(e) => onChange(field.id, { ...value, text: e.target.value })}
                className="min-h-[80px]"
              />
            ) : null}
          </div>
        );
      })}
    </div>
  );
}

/**
 * 動的フォームの値を BE の一括置換ペイロードに変換する。
 * 未入力（空文字 / 未選択 / boolean 未操作）のフィールドは送信しない。
 */
export function buildCheckupResultsPayload(
  fields: CheckupTypeFieldRow[],
  values: Record<number, CheckupFieldValue>,
): CheckupFieldResultInput[] {
  const out: CheckupFieldResultInput[] = [];
  for (const field of fields) {
    const value = values[field.id];
    if (!value) continue;
    const base = { checkup_type_field_id: field.id };

    switch (field.fieldType) {
      case "number": {
        const trimmed = (value.number ?? "").trim();
        if (trimmed === "") break;
        const n = Number(trimmed);
        if (!Number.isFinite(n)) break;
        out.push({ ...base, value_number: n });
        break;
      }
      case "boolean": {
        if (value.bool === undefined) break;
        out.push({ ...base, value_bool: value.bool });
        break;
      }
      case "single_select":
      case "text": {
        const text = (value.text ?? "").trim();
        if (text === "") break;
        out.push({ ...base, value_text: text });
        break;
      }
      case "multi_select":
      case "checklist": {
        const list = value.list ?? [];
        if (list.length === 0) break;
        out.push({ ...base, value_list: list });
        break;
      }
    }
  }
  return out;
}
