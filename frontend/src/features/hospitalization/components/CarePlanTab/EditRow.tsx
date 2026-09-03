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

  const needsRef = requiresRef(type);

  const handleTypeChange = useCallback((next: CarePlanItemType) => {
    setType(next);
    setRefId(null);
  }, []);

  const handleTimingToggle = useCallback((t: CarePlanTiming) => {
    setTiming((prev) => (prev.includes(t) ? prev.filter((x) => x !== t) : [...prev, t]));
  }, []);

  const [state, formAction, isSaving] = useActionState(
    async (_prevState: EditRowState, _formData: FormData): Promise<EditRowState> => {
      const trimmedName = name.trim();
      if (!trimmedName) return { error: "名称は必須です" };
      if (needsRef && !refId) return { error: "参照するマスタを選択してください" };

      await onSave({
        name: trimmedName,
        type,
        timing,
        ...buildRefFields(type, refId),
      });
      return { error: null };
    },
    INITIAL_STATE,
  );

  const canSave = !!name.trim() && (!needsRef || !!refId);

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
      {needsRef ? <CarePlanRefSelect type={type} value={refId} onChange={setRefId} /> : null}
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
