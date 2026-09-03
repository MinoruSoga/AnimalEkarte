import { useCallback, useEffect, useRef } from "react";

import type { DragEndEvent } from "@dnd-kit/core";

import type { ColumnData } from "@/types";

type MoveCard = (
  hoverIndex: number,
  sourceColumn: string,
  targetColumn: string,
  cardId: string,
) => unknown;

function resolveTargetTitle(event: DragEndEvent): string {
  const { over } = event;
  return (
    ((over?.data?.current as Record<string, unknown> | undefined)?.columnTitle as string) ||
    (over?.id as string).replace("column-", "")
  );
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
    (event: DragEndEvent) => {
      const { active, over } = event;
      if (!over) return;

      const cardId = active.id as string;
      const targetTitle = resolveTargetTitle(event);
      const cols = columnsRef.current;
      const sourceColumn = cols.find((col) =>
        col.appointments.some((appointment) => appointment.id === cardId),
      );
      if (!sourceColumn) return;

      const targetCol = cols.find((col) => col.title === targetTitle);
      if (!targetCol) return;

      const dragIndex = sourceColumn.appointments.findIndex(
        (appointment) => appointment.id === cardId,
      );
      const hoverIndex = resolveHoverIndex(over.id as string, targetCol);
      // 同一カラム・同一位置への drop は no-op なのでスキップ（余分な再レンダーを避ける）
      if (sourceColumn.title === targetTitle && dragIndex === hoverIndex) return;

      moveCard(hoverIndex, sourceColumn.title, targetTitle, cardId);
    },
    [moveCard],
  );

  return {
    handleDragEnd: moveActiveCard,
  };
}
