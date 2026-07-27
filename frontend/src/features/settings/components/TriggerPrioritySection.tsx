import { useState, useEffect, useMemo } from "react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { C, STYLE } from "@/lib/design-tokens";
import { Button } from "@/components/ui/button";
import { TriggerTypeLabels } from "@/config/lstep-trigger-types";
import {
  useGetTriggerPriorities,
  useUpdateTriggerPriorities,
} from "../hooks/use-trigger-priorities";
import type { TriggerPriorityItem } from "../hooks/use-trigger-priorities";

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

  // サーバーデータと baseline/draft を同期する。useEffect 内 setState は同期目的のため許容。
  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (data) {
      const sorted = sortItems(data.items);
      setBaseline(sorted);
      setDraft(sorted);
    }
  }, [data]);
  /* eslint-enable react-hooks/set-state-in-effect */

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

  const handleSave = async () => {
    if (draft.some((item) => item.priority < 1)) {
      toast.error("優先順位は1以上を指定してください");
      return;
    }
    const submitted = sortItems([...draft]);
    try {
      await updateMutation.mutateAsync({ items: submitted });
      // baseline を送信値に更新してダーティ状態をリセット。
      // data 参照が変わらない場合でも UI は正確に clean 状態になる。
      setBaseline(submitted);
      setDraft(submitted);
      toast.success("配信優先順位を保存しました");
    } catch (error) {
      handleApiError(error, "配信優先順位の保存");
    }
  };

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

      <div className={`flex justify-end pt-4 border-t ${C.borderLight} mt-4`}>
        <Button
          type="button"
          onClick={handleSave}
          disabled={!isDirty || updateMutation.isPending}
          className={`${C.bgBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} ${C.textOnBrand} rounded-full transition-colors shadow-none border-transparent h-10 text-sm px-4`}
        >
          {updateMutation.isPending ? "保存中..." : "保存"}
        </Button>
      </div>
    </div>
  );
}
