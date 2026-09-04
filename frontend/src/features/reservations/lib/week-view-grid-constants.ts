import type React from "react";

import { C, STYLE } from "@/lib/design-tokens";
import type { Reservation, ReservationStatus } from "@/types";

// 15分枠を44pxにし、連続する最短予約でも実buttonのタッチ領域を重ねない。
export const HOUR_HEIGHT = 176;
export const HOURS = Array.from({ length: 24 }, (_, i) => i);
export const WEEK_DAYS = [0, 1, 2, 3, 4, 5, 6] as const;

export const WHILE_DRAG_FULL = {
  zIndex: 50,
  scale: 1.02,
  opacity: 0.9,
  boxShadow: STYLE.dragPreviewShadowLarge,
} as const;

export const WHILE_DRAG_REDUCED = { zIndex: 50 } as const;

export const STATUS_DOT_STYLE: Partial<
  Record<ReservationStatus, { color: string; label: string }>
> = {
  checked_in: { color: C.bgBrandDot, label: "受付済" },
  in_consultation: { color: C.bgStatusPurpleDot, label: "診療中" },
  accounting: { color: C.bgDiscount, label: "会計待ち" },
  completed: { color: C.bgStatusGrayMedium, label: "完了" },
  cancelled: { color: C.bgNotionRed, label: "キャンセル" },
};

export interface ReservationTypeColor {
  style: React.CSSProperties;
  dotStyle: React.CSSProperties;
  hex: string;
  isInactive?: boolean;
}

export type EventLayout = Record<string, { left: string; width: string }>;

export function calculateEventLayout(dayAppointments: Reservation[]): EventLayout {
  const sorted = [...dayAppointments].sort((a, b) => a.start.getTime() - b.start.getTime());
  const clusters: Reservation[][] = [];
  let currentCluster: Reservation[] = [];
  let clusterEnd = 0;

  sorted.forEach((ev) => {
    if (currentCluster.length === 0) {
      currentCluster.push(ev);
      clusterEnd = ev.end.getTime();
    } else if (ev.start.getTime() < clusterEnd) {
      currentCluster.push(ev);
      if (ev.end.getTime() > clusterEnd) clusterEnd = ev.end.getTime();
    } else {
      clusters.push(currentCluster);
      currentCluster = [ev];
      clusterEnd = ev.end.getTime();
    }
  });
  if (currentCluster.length > 0) clusters.push(currentCluster);

  const styles: EventLayout = {};

  clusters.forEach((cluster) => {
    const columns: Reservation[][] = [];
    const eventColIndex: Record<string, number> = {};

    cluster.forEach((ev) => {
      let placed = false;
      for (let i = 0; i < columns.length; i++) {
        const col = columns[i];
        const last = col[col.length - 1];
        if (last.end <= ev.start) {
          col.push(ev);
          eventColIndex[ev.id] = i;
          placed = true;
          break;
        }
      }
      if (!placed) {
        columns.push([ev]);
        eventColIndex[ev.id] = columns.length - 1;
      }
    });

    const columnWidth = 100 / columns.length;
    cluster.forEach((ev) => {
      const colIndex = eventColIndex[ev.id];
      styles[ev.id] = {
        left: `${colIndex * columnWidth}%`,
        width: `${columnWidth}%`,
      };
    });
  });

  return styles;
}
