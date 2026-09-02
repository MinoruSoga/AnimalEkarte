import type { ColumnData } from "@/types";
import type { ReceptionAppointment } from "../api/types";

// --- 純粋な columns 変換ヘルパー（楽観的 UI 更新の DRY 化）---
// いずれも新しい columns を返し、対象が見つからなければ null を返す（呼び出し側で `?? prev`）。

/** 各カラムと appointments 配列を複製した浅いクローンを返す */
export function cloneColumns(columns: ColumnData[]): ColumnData[] {
  return columns.map((col) => ({ ...col, appointments: [...col.appointments] }));
}

/** id のカードを全カラムから除外した columns を返す */
export function removeCard(columns: ColumnData[], cardId: string): ColumnData[] | null {
  const next = cloneColumns(columns);
  for (const col of next) {
    const index = col.appointments.findIndex((a) => a.id === cardId);
    if (index > -1) {
      col.appointments.splice(index, 1);
      return next;
    }
  }
  return null;
}

/** id のカードへ updated のフィールドをマージした columns を返す */
export function mergeCard(columns: ColumnData[], updated: ReceptionAppointment): ColumnData[] | null {
  const next = cloneColumns(columns);
  for (const col of next) {
    const index = col.appointments.findIndex((a) => a.id === updated.id);
    if (index > -1) {
      col.appointments[index] = { ...col.appointments[index], ...updated };
      return next;
    }
  }
  return null;
}

/**
 * card を sourceTitle → targetTitle へ移した columns を返す。
 * target 内の挿入位置は resolveInsertIndex（card を抜いた後の target appointments を受け取る）で決定し、
 * 省略時は末尾。source/target カラムまたは card が見つからなければ null。
 */
export function relocateCard(
  columns: ColumnData[],
  cardId: string,
  sourceTitle: string,
  targetTitle: string,
  resolveInsertIndex?: (targetAppointments: ReceptionAppointment[]) => number,
): ColumnData[] | null {
  const next = cloneColumns(columns);
  const sourceCol = next.find((c) => c.title === sourceTitle);
  const targetCol = next.find((c) => c.title === targetTitle);
  if (!sourceCol || !targetCol) return null;

  const dragIndex = sourceCol.appointments.findIndex((a) => a.id === cardId);
  if (dragIndex === -1) return null;

  // 同一カラム移動では sourceCol === targetCol。splice 後の target を resolveInsertIndex が見る。
  const [card] = sourceCol.appointments.splice(dragIndex, 1);
  const insertIndex = resolveInsertIndex
    ? resolveInsertIndex(targetCol.appointments)
    : targetCol.appointments.length;
  targetCol.appointments.splice(insertIndex, 0, card);
  return next;
}

export function findAppointment(
  columns: ColumnData[],
  appointmentId: string,
): ReceptionAppointment | undefined {
  return columns
    .flatMap((column) => column.appointments)
    .find((appointment) => appointment.id === appointmentId);
}

export interface CardSnapshot {
  appointment: ReceptionAppointment;
  columnTitle: string;
  index: number;
}

export function captureCardSnapshot(
  columns: ColumnData[],
  appointmentId: string,
): CardSnapshot | null {
  for (const column of columns) {
    const index = column.appointments.findIndex(
      (appointment) => appointment.id === appointmentId,
    );
    if (index !== -1) {
      return {
        appointment: column.appointments[index],
        columnTitle: column.title,
        index,
      };
    }
  }
  return null;
}

/** 対象カードだけを操作前の位置と内容へ戻し、他カードの最新状態は維持する。 */
export function restoreCardSnapshot(columns: ColumnData[], snapshot: CardSnapshot): ColumnData[] {
  const next = removeCard(columns, snapshot.appointment.id) ?? cloneColumns(columns);
  const originalColumn = next.find((column) => column.title === snapshot.columnTitle);
  if (!originalColumn) return columns;

  originalColumn.appointments.splice(
    Math.min(snapshot.index, originalColumn.appointments.length),
    0,
    snapshot.appointment,
  );
  return next;
}
