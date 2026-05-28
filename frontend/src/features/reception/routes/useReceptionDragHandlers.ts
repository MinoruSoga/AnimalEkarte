import { useCallback, useEffect, useRef } from "react";

import type { DragEndEvent, DragOverEvent } from "@dnd-kit/core";

import type { ColumnData } from "@/types";

type MoveCard = (
  dragIndex: number,
  hoverIndex: number,
  sourceColumn: string,
  targetColumn: string,
  draggedCardId?: string,
) => unknown;

function resolveTargetTitle(event: DragOverEvent | DragEndEvent): string {
  const { over } = event;
  return ((over?.data?.current as Record<string, unknown> | undefined)?.columnTitle as string) || (over?.id as string).replace("column-", "");
}

function resolveHoverIndex(overId: string, targetCol: ColumnData): number {
  if (overId.startsWith("column-")) {
    return targetCol.appointments.length;
  }
  const hoverIndex = targetCol.appointments.findIndex((appointment) => appointment.id === overId);
  return hoverIndex === -1 ? targetCol.appointments.length : hoverIndex;
}

export function useReceptionDragHandlers(columns: ColumnData[], moveCard: MoveCard) {
  const columnsRef = useRef(columns);
  useEffect(() => {
    columnsRef.current = columns;
  }, [columns]);

  const moveActiveCard = useCallback(
    (event: DragOverEvent | DragEndEvent) => {
      const { active, over } = event;
      if (!over) return;

      const activeId = active.id as string;
      const targetTitle = resolveTargetTitle(event);
      const cols = columnsRef.current;
      const sourceColumn = cols.find((col) => col.appointments.some((appointment) => appointment.id === activeId));
      if (!sourceColumn) return;

      const targetCol = cols.find((col) => col.title === targetTitle);
      if (!targetCol) return;

      const dragIndex = sourceColumn.appointments.findIndex((appointment) => appointment.id === activeId);
      if (dragIndex === -1) return;

      const hoverIndex = resolveHoverIndex(over.id as string, targetCol);
      if (sourceColumn.title === targetTitle && dragIndex === hoverIndex) return;

      moveCard(dragIndex, hoverIndex, sourceColumn.title, targetTitle, activeId);
    },
    [moveCard],
  );

  return {
    handleDragEnd: moveActiveCard,
    handleDragOver: moveActiveCard,
  };
}
