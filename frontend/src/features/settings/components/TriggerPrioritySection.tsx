import { useState, useMemo, useActionState } from "react";
import { toast } from "sonner";
import { C, STYLE } from "@/lib/design-tokens";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { TriggerTypeLabels } from "@/config/lstep-trigger-types";
import {
  useGetTriggerPriorities,
  useUpdateTriggerPriorities,
} from "../hooks/use-trigger-priorities";
import type { TriggerPriorityItem } from "../hooks/use-trigger-priorities";

type TriggerPriorityFormState = { error?: string } | null;

function sortItems(items: TriggerPriorityItem[]): TriggerPriorityItem[] {
  return [...items].sort((a, b) => {
    if (a.priority !== b.priority) return a.priority - b.priority;
    const la = TriggerTypeLabels[a.trigger_type] ?? a.trigger_type;
    const lb = TriggerTypeLabels[b.trigger_type] ?? b.trigger_type;
    return la.localeCompare(lb, "ja");
  });
}

export function TriggerPrioritySection() {
  const { data, isLoading, isError } = useGetTriggerPriorities();
  const updateMutation = useUpdateTriggerPriorities();

  // draft: 編集中の状態。baseline: 最後にサーバーと同期した状態（保存後も更新）。
  // React Query の structuralSharing で data 参照が変わらない場合でも
  // baseline を独立管理することで isDirty を正確に追跡できる。
  const [draft, setDraft] = useState<TriggerPriorityItem[]>([]);
  const [baseline, setBaseline] = useState<TriggerPriorityItem[]>([]);

  // FE-RC-032: useEffect 同期の代わりに render 中の prev 比較で同期する
  // （OwnersList.tsx の searchParamsKey パターンと同型）。サーバーデータの
  // シグネチャが変わった時だけ setState し、それ以外は no-op のまま。
  const serverKey = useMemo(
    () => (data ? JSON.stringify(sortItems(data.items)) : ""),
    [data],
  );
  const [prevServerKey, setPrevServerKey] = useState(serverKey);
  if (serverKey !== prevServerKey) {
    setPrevServerKey(serverKey);
    if (data) {
      const sorted = sortItems(data.items);
      setBaseline(sorted);
      setDraft(sorted);
    }
  }

  const isDirty = useMemo(() => {
    const baseMap = new Map(baseline.map((i) => [i.trigger_type, i.priority]));
    return draft.some((d) => baseMap.get(d.trigger_type) !== d.priority);
  }, [draft, baseline]);

  const handlePriorityChange = (triggerType: string, value: string) => {
    const n = parseInt(value, 10);
    if (isNaN(n)) return;
    setDraft((prev) =>
      prev.map((item) =>
        item.trigger_type === triggerType ? { ...item, priority: n } : item,
      ),
    );
  };

  const [state, formAction] = useActionState<TriggerPriorityFormState, FormData>(
    async () => {
      if (draft.some((item) => item.priority < 1)) {
        return { error: "優先順位は1以上を指定してください" };
      }
      const submitted = sortItems([...draft]);
      try {
        await updateMutation.mutateAsync({ items: submitted });
        // baseline を送信値に更新してダーティ状態をリセット。
        // data 参照が変わらない場合でも UI は正確に clean 状態になる。
        setBaseline(submitted);
        setDraft(submitted);
        toast.success("配信優先順位を保存しました");
        return null;
      } catch {
        // FE-RC-005: API エラーは useUpdateTriggerPriorities の onError
        // （use-trigger-priorities.ts）が handleApiError 済み。ここでは再通知しない。
        return { error: "配信優先順位の保存に失敗しました" };
      }
    },
    null,
  );

  if (isLoading) {
    return (
      <div className={`${STYLE.formCard} max-w-2xl mt-6`}>
        <div className={`flex items-center justify-center py-8 text-sm ${C.text50}`}>
          読み込み中...
        </div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className={`${STYLE.formCard} max-w-2xl mt-6`}>
        <div className={`flex items-center justify-center py-8 text-sm ${C.danger}`}>
          読み込みに失敗しました
        </div>
      </div>
    );
  }

  return (
    <div className={`${STYLE.formCard} max-w-2xl mt-6`}>
      <h2 className={`text-base font-semibold ${C.text} mb-1`}>配信優先順位</h2>
      <p className={`text-sm ${C.text60} mb-4`}>
        数値が小さいほど優先。同値は同一優先階層として扱われます。
      </p>

      <form action={formAction} noValidate>
        <div className={`border ${C.borderLight} rounded-xs overflow-hidden`}>
          {draft.map((item, idx) => (
            <div
              key={item.trigger_type}
              className={`flex items-center gap-4 px-4 py-2.5 ${
                idx > 0 ? `border-t ${C.borderLight}` : ""
              }`}
            >
              <span className={`flex-1 text-sm ${C.text}`}>
                {TriggerTypeLabels[item.trigger_type] ?? item.trigger_type}
              </span>
              <input
                type="number"
                min={1}
                value={item.priority}
                onChange={(e) => handlePriorityChange(item.trigger_type, e.target.value)}
                aria-label={`${TriggerTypeLabels[item.trigger_type] ?? item.trigger_type} 優先順位`}
                className={`${STYLE.formInput} rounded-xs border px-2 w-20 text-right outline-none focus:ring-2 ${C.focusRingAccent30}`}
              />
            </div>
          ))}
        </div>

        {state?.error ? (
          <p className={`text-sm ${C.danger} mt-2`} role="alert">
            {state.error}
          </p>
        ) : null}

        <div className={`flex justify-end pt-4 border-t ${C.borderLight} mt-4`}>
          <SubmitButton disabled={!isDirty} loadingText="保存中..." className="h-10 text-sm px-4">
            保存
          </SubmitButton>
        </div>
      </form>
    </div>
  );
}
