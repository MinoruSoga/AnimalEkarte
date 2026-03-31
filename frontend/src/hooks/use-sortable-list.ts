import { useState, useMemo, useCallback } from "react";
import {
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
  type DragStartEvent,
} from "@dnd-kit/core";
import { arrayMove, sortableKeyboardCoordinates } from "@dnd-kit/sortable";

interface UseSortableListOptions<T extends { id: string }> {
  items: T[];
  /** ドラッグ完了時の新順序 ID 配列を受け取る。API 呼び出しはここで行う */
  onReorder: (newIds: string[]) => void;
}

interface UseSortableListReturn<T> {
  orderedItems: T[];
  sensors: ReturnType<typeof useSensors>;
  activeId: string | null;
  handleDragStart: (event: DragStartEvent) => void;
  handleDragEnd: (event: DragEndEvent) => void;
  handleDragCancel: () => void;
  /** 外部から楽観的順序をリセットする（onReorder 成功後） */
  resetOrder: () => void;
}

export function useSortableList<T extends { id: string }>({
  items,
  onReorder,
}: UseSortableListOptions<T>): UseSortableListReturn<T> {
  const [overrideOrder, setOverrideOrder] = useState<string[]>([]);
  const [activeId, setActiveId] = useState<string | null>(null);

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );

  const orderedItems = useMemo(() => {
    if (overrideOrder.length === 0) return items;
    const idx = new Map(overrideOrder.map((id, i) => [id, i]));
    return [...items].sort((a, b) => (idx.get(a.id) ?? 0) - (idx.get(b.id) ?? 0));
  }, [items, overrideOrder]);

  const handleDragStart = useCallback((event: DragStartEvent) => {
    setActiveId(String(event.active.id));
  }, []);

  const handleDragCancel = useCallback(() => {
    setActiveId(null);
  }, []);

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      setActiveId(null);
      const { active, over } = event;
      if (!over || active.id === over.id) return;

      const activeItemId = String(active.id);
      const overItemId = String(over.id);

      const currentIds = items.map((m) => m.id);
      const newOrder = arrayMove(
        currentIds,
        currentIds.indexOf(activeItemId),
        currentIds.indexOf(overItemId),
      );

      setOverrideOrder(newOrder);
      onReorder(newOrder);
    },
    [items, onReorder],
  );

  const resetOrder = useCallback(() => {
    setOverrideOrder([]);
  }, []);

  return {
    orderedItems,
    sensors,
    activeId,
    handleDragStart,
    handleDragEnd,
    handleDragCancel,
    resetOrder,
  };
}
