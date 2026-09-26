// React/Framework
import { useActionState, useState, useCallback } from "react";

// External
import { Plus } from "lucide-react";

// Internal
import { C, ICON } from "@/lib/design-tokens";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectTrigger, SelectValue } from "@/components/ui/select";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";

// Relative
import { TYPE_SELECT_ITEMS, TIMING_OPTIONS } from "./CarePlanBadges";
import { CarePlanRefSelect } from "./CarePlanRefSelect";
import { requiresRef, buildRefFields } from "../../lib/care-plan-item-model";

// Types
import type {
  CarePlanItemType,
  CarePlanTiming,
  CreateCarePlanItemInput,
} from "../../api/care-plan-items";

const INITIAL_TIMING: CarePlanTiming[] = ["morning"];

/** 手入力「その他」理由の最大文字数（BE の 500 rune 契約と一致）。 */
const OTHER_REASON_MAX_LENGTH = 500;

interface AddFormState {
  error: string | null;
}

const INITIAL_STATE: AddFormState = { error: null };

interface AddFormProps {
  onSubmit: (input: CreateCarePlanItemInput) => Promise<void>;
}

export function AddForm({ onSubmit }: AddFormProps) {
  const [type, setType] = useState<CarePlanItemType>("instruction");
  const [name, setName] = useState("");
  const [timing, setTiming] = useState<CarePlanTiming[]>(INITIAL_TIMING);
  const [refId, setRefId] = useState<string | null>(null);
  /** 選択した参照先マスタの単価。0 は有限値として保持する。 */
  const [refUnitPrice, setRefUnitPrice] = useState<number | null>(null);
  /** type=item の手入力「その他」モード。マスタ参照の代わりに理由と単価を直接入力する。 */
  const [manual, setManual] = useState(false);
  const [otherReason, setOtherReason] = useState("");
  const [manualUnitPrice, setManualUnitPrice] = useState("");

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

  const [state, formAction] = useActionState(
    async (_prevState: AddFormState, _formData: FormData): Promise<AddFormState> => {
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
        await onSubmit({
          type,
          name: trimmedName,
          timing,
          manual: true,
          other_reason: trimmedReason,
          unit_price: price,
          ...buildRefFields(type, null),
        });
      } else {
        if (needsRef && !refId) return { error: "参照するマスタを選択してください" };

        await onSubmit({
          type,
          name: trimmedName,
          timing,
          ...buildRefFields(type, refId),
          // マスタ参照を持つ type は選択マスタの単価を unit_price に転記する(BE は省略時 0 で保存)
          ...(refUnitPrice !== null ? { unit_price: refUnitPrice } : {}),
        });
      }

      setName("");
      setType("instruction");
      setTiming(INITIAL_TIMING);
      setRefId(null);
      setRefUnitPrice(null);
      setManual(false);
      setOtherReason("");
      setManualUnitPrice("");
      return { error: null };
    },
    INITIAL_STATE,
  );

  const canSubmit =
    !!name.trim() &&
    (isManualItem ? !!otherReason.trim() && manualUnitPrice.trim() !== "" : !needsRef || !!refId);

  return (
    <form action={formAction} className={`border-t ${C.borderLight} pt-3 mt-2`}>
      <p className={`text-xs font-medium ${C.text60} mb-2`}>新しいケアプラン項目を追加</p>
      <div className="flex flex-col gap-2">
        {state.error ? (
          <p role="alert" className={`text-xs ${C.textNotionRed}`}>
            {state.error}
          </p>
        ) : null}
        <div className="flex gap-2 items-center">
          <Select value={type} onValueChange={(v) => handleTypeChange(v as CarePlanItemType)}>
            <SelectTrigger className="w-28 h-11 text-xs" aria-label="ケアプラン項目種別">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{TYPE_SELECT_ITEMS}</SelectContent>
          </Select>
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            aria-label="ケアプラン項目名"
            className="h-11 text-sm flex-1"
            placeholder="名称を入力"
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
              className="h-11 text-sm"
              placeholder="単価（円）を入力"
            />
            <Input
              value={otherReason}
              onChange={(e) => setOtherReason(e.target.value)}
              maxLength={OTHER_REASON_MAX_LENGTH}
              aria-label="その他理由"
              className="h-11 text-sm"
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
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className={`text-xs ${C.text50} shrink-0`}>タイミング:</span>
            <div className="flex gap-2">
              {TIMING_OPTIONS.map((opt) => (
                <label
                  key={opt.value}
                  className="flex min-h-11 min-w-11 items-center gap-1 cursor-pointer"
                >
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
          <SubmitButton size="sm" disabled={!canSubmit} className="h-8 text-xs gap-1">
            <Plus className={ICON.action} />
            追加
          </SubmitButton>
        </div>
      </div>
    </form>
  );
}
