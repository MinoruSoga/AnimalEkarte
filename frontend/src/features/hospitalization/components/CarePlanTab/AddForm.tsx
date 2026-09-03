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
import { requiresRef, buildRefFields } from "./care-plan-item-model";

// Types
import type { CarePlanItemType, CarePlanTiming, CreateCarePlanItemInput } from "../../api/care-plan-items";

const INITIAL_TIMING: CarePlanTiming[] = ["morning"];

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

  const needsRef = requiresRef(type);

  const handleTypeChange = useCallback((next: CarePlanItemType) => {
    setType(next);
    setRefId(null);
  }, []);

  const handleTimingToggle = useCallback((t: CarePlanTiming) => {
    setTiming((prev) => (prev.includes(t) ? prev.filter((x) => x !== t) : [...prev, t]));
  }, []);

  const [state, formAction] = useActionState(
    async (_prevState: AddFormState, _formData: FormData): Promise<AddFormState> => {
      const trimmedName = name.trim();
      if (!trimmedName) return { error: "名称は必須です" };
      if (needsRef && !refId) return { error: "参照するマスタを選択してください" };

      await onSubmit({
        type,
        name: trimmedName,
        timing,
        ...buildRefFields(type, refId),
      });

      setName("");
      setType("instruction");
      setTiming(INITIAL_TIMING);
      setRefId(null);
      return { error: null };
    },
    INITIAL_STATE,
  );

  const canSubmit = !!name.trim() && (!needsRef || !!refId);

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
        {needsRef ? <CarePlanRefSelect type={type} value={refId} onChange={setRefId} /> : null}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className={`text-xs ${C.text50} shrink-0`}>タイミング:</span>
            <div className="flex gap-2">
              {TIMING_OPTIONS.map((opt) => (
                <label key={opt.value} className="flex min-h-11 min-w-11 items-center gap-1 cursor-pointer">
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
