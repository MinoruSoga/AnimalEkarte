// React/Framework
import { useActionState, useState, useCallback } from "react";

// Internal
import { C } from "@/lib/design-tokens";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectTrigger, SelectValue } from "@/components/ui/select";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";

// Relative
import { TYPE_SELECT_ITEMS, TIMING_OPTIONS } from "./CarePlanBadges";
import { CarePlanRefSelect } from "./CarePlanRefSelect";
import { requiresRef, buildRefFields } from "../../lib/care-plan-item-model";

// Types
import type {
  CarePlanItem,
  CarePlanItemType,
  CarePlanTiming,
  UpdateCarePlanItemInput,
} from "../../api/care-plan-items";

/** item の既存 type に対応する FK 値を取り出す(type 未変更時の初期選択状態) */
function initialRefId(item: CarePlanItem): string | null {
  if (item.type === "medicine") return item.medicine_id ?? null;
  if (item.type === "treatment") return item.procedure_id ?? null;
  if (item.type === "item") return item.hospitalization_plan_id ?? null;
  return null;
}

/** 手入力「その他」理由の最大文字数（BE の 500 rune 契約と一致）。 */
const OTHER_REASON_MAX_LENGTH = 500;

interface EditRowState {
  error: string | null;
}

const INITIAL_STATE: EditRowState = { error: null };

interface EditRowProps {
  item: CarePlanItem;
  onSave: (input: UpdateCarePlanItemInput) => Promise<void>;
  onCancel: () => void;
}

export function EditRow({ item, onSave, onCancel }: EditRowProps) {
  const [name, setName] = useState(item.name);
  const [type, setType] = useState<CarePlanItemType>(item.type);
  const [timing, setTiming] = useState<CarePlanTiming[]>(item.timing);
  const [refId, setRefId] = useState<string | null>(initialRefId(item));
  /** 参照必須 type の保存単価。既存値で初期化し、マスタ再選択時はその price で上書き。0 は有限値として保持。 */
  const [refUnitPrice, setRefUnitPrice] = useState<number | null>(
    requiresRef(item.type, item.manual) ? item.unit_price : null,
  );
  /** type=item の手入力「その他」モード。既存手入力行は manual=true で初期化される。 */
  const [manual, setManual] = useState(item.manual);
  const [otherReason, setOtherReason] = useState(item.other_reason);
  const [manualUnitPrice, setManualUnitPrice] = useState(
    item.manual ? String(item.unit_price) : "",
  );

  const isManualItem = type === "item" && manual;
  const needsRef = requiresRef(type, manual);

  const handleTypeChange = useCallback((next: CarePlanItemType) => {
    setType(next);
    setRefId(null);
    setRefUnitPrice(null);
    // 手入力モードは type=item 専用。別 type へ切り替えたら理由・単価をクリアする。
    setManual(false);
    setOtherReason("");
    setManualUnitPrice("");
  }, []);

  /** 手入力⇔マスタ参照の切替。モードを抜ける側の入力値は捨てる（理由/参照の持ち越しを防ぐ）。 */
  const handleManualToggle = useCallback((checked: boolean) => {
    setManual(checked);
    if (checked) {
      setRefId(null);
      setRefUnitPrice(null);
    } else {
      setOtherReason("");
      setManualUnitPrice("");
    }
  }, []);

  const handleTimingToggle = useCallback((t: CarePlanTiming) => {
    setTiming((prev) => (prev.includes(t) ? prev.filter((x) => x !== t) : [...prev, t]));
  }, []);

  const [state, formAction, isSaving] = useActionState(
    async (_prevState: EditRowState, _formData: FormData): Promise<EditRowState> => {
      const trimmedName = name.trim();
      if (!trimmedName) return { error: "名称は必須です" };
      if (isManualItem) {
        const trimmedReason = otherReason.trim();
        if (!trimmedReason) return { error: "その他の理由を入力してください" };
        if ([...trimmedReason].length > OTHER_REASON_MAX_LENGTH) {
          return { error: "理由は500文字以内で入力してください" };
        }
        const price = manualUnitPrice.trim() === "" ? NaN : Number(manualUnitPrice);
        if (!Number.isInteger(price) || price < 0) {
          return { error: "単価は0以上の整数で入力してください" };
        }
        await onSave({
          name: trimmedName,
          type,
          timing,
          manual: true,
          other_reason: trimmedReason,
          unit_price: price,
          ...buildRefFields(type, null),
        });
      } else {
        if (needsRef && !refId) return { error: "参照するマスタを選択してください" };

        await onSave({
          name: trimmedName,
          type,
          timing,
          ...buildRefFields(type, refId),
          // 手入力からマスタ参照（または別 type）へ移る場合は manual=false を明示し、
          // BE が category=other / other_reason をクリアする。
          ...(item.manual ? { manual: false } : {}),
          // 参照必須 type は既存/再選択したマスタの単価を unit_price に維持・転記する。
          // 参照なし type への変更では残存価格を明示的に 0 へ戻す(退院会計は全項目の unit_price を写すため)。
          unit_price: needsRef ? (refUnitPrice ?? item.unit_price) : 0,
        });
      }
      return { error: null };
    },
    INITIAL_STATE,
  );

  const canSave =
    !!name.trim() &&
    (isManualItem ? !!otherReason.trim() && manualUnitPrice.trim() !== "" : !needsRef || !!refId);

  return (
    <form
      action={formAction}
      className={`flex flex-col gap-2 p-3 ${C.bgBrand5} rounded-lg border ${C.borderBrandLight}`}
    >
      {state.error ? (
        <p role="alert" className={`text-xs ${C.textNotionRed}`}>
          {state.error}
        </p>
      ) : null}
      <div className="flex gap-2 items-center">
        <Select value={type} onValueChange={(v) => handleTypeChange(v as CarePlanItemType)}>
          <SelectTrigger className="w-28 h-8 text-xs" aria-label="ケアプラン項目種別">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>{TYPE_SELECT_ITEMS}</SelectContent>
        </Select>
        <Input
          value={name}
          onChange={(e) => setName(e.target.value)}
          aria-label="ケアプラン項目名"
          className="h-8 text-sm flex-1"
          placeholder="名称"
        />
      </div>
      {type === "item" ? (
        <label className="flex items-center gap-1 cursor-pointer">
          <input
            type="checkbox"
            checked={manual}
            onChange={(e) => handleManualToggle(e.target.checked)}
            className="rounded"
          />
          <span className="text-xs">手入力（その他）</span>
        </label>
      ) : null}
      {isManualItem ? (
        <>
          <Input
            value={manualUnitPrice}
            onChange={(e) => setManualUnitPrice(e.target.value)}
            type="number"
            min={0}
            step={1}
            aria-label="手入力の単価"
            className="h-8 text-sm"
            placeholder="単価（円）を入力"
          />
          <Input
            value={otherReason}
            onChange={(e) => setOtherReason(e.target.value)}
            maxLength={OTHER_REASON_MAX_LENGTH}
            aria-label="その他理由"
            className="h-8 text-sm"
            placeholder="理由を入力（必須・500文字以内）"
          />
        </>
      ) : null}
      {needsRef ? (
        <CarePlanRefSelect
          type={type}
          value={refId}
          onChange={setRefId}
          onUnitPriceChange={setRefUnitPrice}
        />
      ) : null}
      <div className="flex items-center gap-3">
        <span className={`text-xs ${C.text50} shrink-0`}>タイミング:</span>
        <div className="flex gap-2">
          {TIMING_OPTIONS.map((opt) => (
            <label key={opt.value} className="flex items-center gap-1 cursor-pointer">
              <input
                type="checkbox"
                checked={timing.includes(opt.value)}
                onChange={() => handleTimingToggle(opt.value)}
                className="rounded"
              />
              <span className="text-xs">{opt.label}</span>
            </label>
          ))}
        </div>
      </div>
      <div className="flex gap-2 justify-end">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onCancel}
          disabled={isSaving}
          className="h-7 text-xs"
        >
          キャンセル
        </Button>
        <SubmitButton size="sm" disabled={!canSave} className="h-7 text-xs">
          保存
        </SubmitButton>
      </div>
    </form>
  );
}
