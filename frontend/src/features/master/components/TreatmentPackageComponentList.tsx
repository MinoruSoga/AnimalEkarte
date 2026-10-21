/**
 * EMR-523: パッケージ構成品の編集リスト。
 * 7 master 種＋正規 manual_other の混在を許し、型別 ID・quantity を文字列のまま保持する。
 * 健診 set/検査 panel は owner 定義の1品目として並べ、nested package へは変換しない。
 */
import { ArrowDown, ArrowUp, Plus, Trash2 } from "lucide-react";
import { C, STYLE } from "@/lib/design-tokens";
import { Input } from "@/components/ui/input";
import { NativeSelect } from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import {
  ITEM_KIND_LABELS,
  isMasterItemKind,
  newComponentDraft,
  type ComponentDraft,
} from "../lib/treatment-package-form";
import { TREATMENT_PACKAGE_ITEM_KINDS, type TreatmentPackageItemKind } from "../api/treatment-packages";

interface TreatmentPackageComponentListProps {
  components: ComponentDraft[];
  onChange: (next: ComponentDraft[]) => void;
  disabled?: boolean;
  /** componentKey → 行内エラー（保存前検証・422 行別理由） */
  componentErrors?: ReadonlyMap<string, readonly string[]>;
}

export function TreatmentPackageComponentList({
  components,
  onChange,
  disabled = false,
  componentErrors,
}: TreatmentPackageComponentListProps) {
  const updateAt = (index: number, patch: Partial<ComponentDraft>) => {
    onChange(components.map((c, i) => (i === index ? { ...c, ...patch } : c)));
  };
  const removeAt = (index: number) => onChange(components.filter((_, i) => i !== index));
  const move = (index: number, direction: -1 | 1) => {
    const target = index + direction;
    if (target < 0 || target >= components.length) return;
    const next = [...components];
    const [moved] = next.splice(index, 1);
    next.splice(target, 0, moved!);
    onChange(next);
  };
  const addComponent = () => onChange([...components, newComponentDraft("consultation")]);

  return (
    <div className="flex flex-col gap-3">
      {components.map((c, index) => {
        const errors = componentErrors?.get(c.componentKey) ?? [];
        return (
          <div
            key={c.componentKey}
            data-testid={`package-component-${c.componentKey}`}
            className={`rounded-lg border ${C.borderLight} p-3`}
          >
            <div className="flex items-start gap-2">
              <div className="flex flex-col">
                <button
                  type="button"
                  aria-label="上へ移動"
                  disabled={disabled || index === 0}
                  onClick={() => move(index, -1)}
                  className="min-h-[22px] min-w-[22px] flex items-center justify-center"
                >
                  <ArrowUp className="size-4" aria-hidden="true" />
                </button>
                <button
                  type="button"
                  aria-label="下へ移動"
                  disabled={disabled || index === components.length - 1}
                  onClick={() => move(index, 1)}
                  className="min-h-[22px] min-w-[22px] flex items-center justify-center"
                >
                  <ArrowDown className="size-4" aria-hidden="true" />
                </button>
              </div>
              <div className="flex-1 grid grid-cols-2 gap-2">
                <div>
                  <label className={`text-xs ${C.text60}`}>種別</label>
                  <NativeSelect
                    aria-label="種別"
                    disabled={disabled}
                    value={c.kind}
                    onChange={(e) =>
                      updateAt(index, {
                        kind: e.target.value as TreatmentPackageItemKind,
                        masterId: "",
                      })
                    }
                  >
                    {TREATMENT_PACKAGE_ITEM_KINDS.map((kind) => (
                      <option key={kind} value={kind}>
                        {ITEM_KIND_LABELS[kind]}
                      </option>
                    ))}
                  </NativeSelect>
                </div>
                <div>
                  <label className={`text-xs ${C.text60}`}>数量</label>
                  <Input
                    aria-label="数量"
                    inputMode="decimal"
                    disabled={disabled}
                    value={c.quantity}
                    onChange={(e) => updateAt(index, { quantity: e.target.value })}
                  />
                </div>
                {isMasterItemKind(c.kind) ? (
                  <div className="col-span-2">
                    <label className={`text-xs ${C.text60}`}>
                      {ITEM_KIND_LABELS[c.kind]}マスタ ID（文字列のまま保持）
                    </label>
                    <Input
                      aria-label={`${ITEM_KIND_LABELS[c.kind]}マスタ ID`}
                      inputMode="numeric"
                      disabled={disabled}
                      value={c.masterId}
                      onChange={(e) => updateAt(index, { masterId: e.target.value.trim() })}
                    />
                  </div>
                ) : (
                  <ManualFields
                    draft={c}
                    disabled={disabled}
                    onChange={(manual) => updateAt(index, { manual })}
                  />
                )}
              </div>
              <button
                type="button"
                aria-label="構成品を削除"
                disabled={disabled}
                onClick={() => removeAt(index)}
                className="min-h-[44px] min-w-[44px] flex items-center justify-center"
              >
                <Trash2 className="size-4" aria-hidden="true" />
              </button>
            </div>
            {errors.length > 0 ? (
              <ul className={`mt-2 text-sm ${C.danger}`} role="alert">
                {errors.map((message) => (
                  <li key={message}>{message}</li>
                ))}
              </ul>
            ) : null}
          </div>
        );
      })}
      <button
        type="button"
        disabled={disabled}
        onClick={addComponent}
        className={STYLE.settingsRow}
      >
        <Plus className="size-4" aria-hidden="true" />
        構成品を追加
      </button>
    </div>
  );
}

function ManualFields({
  draft,
  disabled,
  onChange,
}: {
  draft: ComponentDraft;
  disabled: boolean;
  onChange: (manual: ComponentDraft["manual"]) => void;
}) {
  const m = draft.manual;
  return (
    <div className="col-span-2 grid grid-cols-2 gap-2">
      <div>
        <label className={`text-xs ${C.text60}`}>名称（必須）</label>
        <Input
          aria-label="手入力名称"
          disabled={disabled}
          value={m.name}
          onChange={(e) => onChange({ ...m, name: e.target.value })}
        />
      </div>
      <div>
        <label className={`text-xs ${C.text60}`}>理由（必須）</label>
        <Input
          aria-label="手入力理由"
          disabled={disabled}
          value={m.reason}
          onChange={(e) => onChange({ ...m, reason: e.target.value })}
        />
      </div>
      <div>
        <label className={`text-xs ${C.text60}`}>単位</label>
        <Input
          aria-label="単位"
          disabled={disabled}
          value={m.unitLabel}
          onChange={(e) => onChange({ ...m, unitLabel: e.target.value })}
        />
      </div>
      <div>
        <label className={`text-xs ${C.text60}`}>単価（円・整数）</label>
        <Input
          aria-label="単価"
          inputMode="numeric"
          disabled={disabled}
          value={m.unitPrice}
          onChange={(e) => onChange({ ...m, unitPrice: e.target.value.trim() })}
        />
      </div>
      <div>
        <label className={`text-xs ${C.text60}`}>税区分</label>
        <NativeSelect
          aria-label="税区分"
          disabled={disabled}
          value={m.taxType}
          onChange={(e) =>
            onChange({ ...m, taxType: e.target.value as ComponentDraft["manual"]["taxType"] })
          }
        >
          <option value="included">内税</option>
          <option value="excluded">外税</option>
          <option value="exempt">非課税</option>
        </NativeSelect>
      </div>
      <div>
        <label className={`text-xs ${C.text60}`}>税率（正確な十進）</label>
        <Input
          aria-label="税率"
          inputMode="decimal"
          disabled={disabled}
          value={m.taxRate}
          onChange={(e) => onChange({ ...m, taxRate: e.target.value.trim() })}
        />
      </div>
      <label className="col-span-2 flex items-center gap-2 text-sm">
        <Checkbox
          checked={m.isNonInsurance}
          disabled={disabled}
          onCheckedChange={(checked) => onChange({ ...m, isNonInsurance: checked === true })}
        />
        保険適用外
      </label>
    </div>
  );
}
