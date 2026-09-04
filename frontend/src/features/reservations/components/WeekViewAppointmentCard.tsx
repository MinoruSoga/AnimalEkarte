import { memo, useMemo, useRef } from "react";
import { format } from "date-fns";
import { motion } from "motion/react";

import { useReducedMotion } from "@/hooks/use-reduced-motion";
import { C } from "@/lib/design-tokens";
import { getReservationTypeColor, getReservationTypeName } from "@/lib/status-helpers";
import { DISPLAY_TIME_FORMAT } from "@/lib/format/date";
import type { Reservation } from "@/types";

import {
  HOUR_HEIGHT,
  STATUS_DOT_STYLE,
  WHILE_DRAG_FULL,
  WHILE_DRAG_REDUCED,
  type ReservationTypeColor,
} from "../lib/week-view-grid-constants";

interface AppointmentCardProps {
  appointment: Reservation;
  layoutStyle: { left: string; width: string };
  onClick: (reservation: Reservation) => void;
  onUpdate?: (reservation: Reservation, newStart: Date, newEnd: Date) => void;
  dynamicColorMap?: Map<string, ReservationTypeColor>;
}

interface FirstVisitBadgeProps {
  show: boolean;
  compact?: boolean;
}

function FirstVisitBadge({ show, compact = false }: FirstVisitBadgeProps) {
  if (!show) return null;

  return (
    <span
      className={`${C.bgRedLight} ${C.textNotionRed} ${
        compact ? "px-0.5 text-xs" : "px-1.5 text-sm"
      } rounded flex-shrink-0 no-underline`}
    >
      初
    </span>
  );
}

export const AppointmentCard = memo(function AppointmentCard({
  appointment,
  layoutStyle,
  onClick,
  onUpdate,
  dynamicColorMap,
}: AppointmentCardProps) {
  const startMinutes = appointment.start.getHours() * 60 + appointment.start.getMinutes();
  const endMinutes = appointment.end.getHours() * 60 + appointment.end.getMinutes();
  const durationMinutes = endMinutes - startMinutes;
  const height = Math.max((durationMinutes / 60) * HOUR_HEIGHT, 44);
  const top = (startMinutes / 60) * HOUR_HEIGHT;

  const dragConstraints = useMemo(
    () => ({
      top: -top,
      bottom: 24 * HOUR_HEIGHT - (top + height),
    }),
    [height, top],
  );

  const isDimmed = appointment.status === "completed" || appointment.status === "cancelled";
  const isCancelled = appointment.status === "cancelled";
  const isCompact = height <= 44;
  const isNarrow = !isCompact && height <= 85;
  const dotInfo = STATUS_DOT_STYLE[appointment.status];
  const isDragging = useRef(false);
  const reduced = useReducedMotion();
  const whileDragProps = reduced ? WHILE_DRAG_REDUCED : WHILE_DRAG_FULL;
  const colorStyle = getReservationTypeColor(appointment.type, dynamicColorMap);
  const isClassNameColor = typeof colorStyle === "string";
  const isFirstVisit = appointment.visitType === "first";
  const typeName = getReservationTypeName(appointment.type);
  const typeLabel =
    dynamicColorMap?.get(appointment.type)?.isInactive === true ? `${typeName}（無効）` : typeName;

  const tooltipText = useMemo(
    () =>
      [
        `${format(appointment.start, DISPLAY_TIME_FORMAT)}–${format(appointment.end, DISPLAY_TIME_FORMAT)}`,
        appointment.petName,
        appointment.ownerName,
        appointment.doctor,
        typeLabel,
      ]
        .filter(Boolean)
        .join(" / "),
    [appointment, typeLabel],
  );

  return (
    <motion.button
      type="button"
      className={`absolute min-h-11 rounded border hover:ring-1 ${C.ringPrimary20} transition-all cursor-grab active:cursor-grabbing z-10 overflow-hidden group touch-none
        ${isCompact ? "py-px px-1" : isNarrow ? "px-1 py-0.5" : "p-1"}
        ${isClassNameColor ? colorStyle : ""}
        ${isDimmed ? "opacity-60" : "opacity-100"}
        ${isCancelled ? `line-through ${C.decorationDanger50}` : ""}
      `}
      aria-label={`${format(appointment.start, DISPLAY_TIME_FORMAT)}〜${format(appointment.end, DISPLAY_TIME_FORMAT)} ${appointment.petName} ${appointment.ownerName} ${typeLabel}`}
      title={tooltipText}
      style={{
        top: `${top}px`,
        height: `${height}px`,
        left: layoutStyle.left,
        width: layoutStyle.width,
        ...(isClassNameColor ? {} : (colorStyle as React.CSSProperties)),
      }}
      onClick={(e) => {
        e.stopPropagation();
        if (!isDragging.current) {
          onClick(appointment);
        }
        isDragging.current = false;
      }}
      drag="y"
      dragMomentum={false}
      dragElastic={0}
      dragConstraints={dragConstraints}
      dragSnapToOrigin={!onUpdate}
      onDrag={() => {
        isDragging.current = true;
      }}
      onDragEnd={(_, info) => {
        const movedMinutes = (info.offset.y / HOUR_HEIGHT) * 60;
        const snappedMinutes = Math.round(movedMinutes / 15) * 15;
        if (snappedMinutes === 0 || !onUpdate) return;

        const newStart = new Date(appointment.start.getTime() + snappedMinutes * 60000);
        const newEnd = new Date(appointment.end.getTime() + snappedMinutes * 60000);
        onUpdate(appointment, newStart, newEnd);
      }}
      whileDrag={whileDragProps}
    >
      {dotInfo ? (
        <div
          className={`rounded-full ${dotInfo.color} border border-white absolute z-20 ${
            isCompact ? "w-1.5 h-1.5 top-0.5 right-0.5" : "w-2.5 h-2.5 top-1 right-1"
          }`}
          title={dotInfo.label}
        />
      ) : null}

      {isCompact ? (
        <div className="flex items-center gap-1 h-full pointer-events-none relative z-10 pr-2">
          <span className="font-bold text-xs whitespace-nowrap leading-none">
            {format(appointment.start, DISPLAY_TIME_FORMAT)}
          </span>
          <FirstVisitBadge show={isFirstVisit} compact />
          <span className="truncate text-xs font-medium leading-none">{appointment.petName}</span>
        </div>
      ) : null}

      {isNarrow ? (
        <div className="flex flex-col h-full pointer-events-none relative z-10">
          <div className="flex items-center gap-1 pr-3 leading-none">
            <span className="font-bold text-xs whitespace-nowrap">
              {format(appointment.start, DISPLAY_TIME_FORMAT)}
            </span>
            <FirstVisitBadge show={isFirstVisit} compact />
          </div>
          <div className="font-medium truncate text-xs leading-none mt-0.5">
            {appointment.petName}
          </div>
          <div className="truncate text-xs opacity-70 leading-none mt-0.5">
            {appointment.ownerName}
          </div>
        </div>
      ) : null}

      {!isCompact && !isNarrow ? (
        <div className="flex flex-col h-full pointer-events-none relative z-10">
          <div className="flex items-center gap-1 font-bold text-sm leading-none mb-1 pr-3">
            <span className="truncate">{format(appointment.start, DISPLAY_TIME_FORMAT)}</span>
            <FirstVisitBadge show={isFirstVisit} />
          </div>
          <div className="font-medium truncate text-sm leading-none">{appointment.petName}</div>
          {height > 36 ? (
            <div className="truncate text-sm opacity-70 leading-none mt-0.5">
              {appointment.ownerName}
            </div>
          ) : null}
          {height > 52 && appointment.doctor ? (
            <div className="truncate text-sm opacity-70 leading-none mt-0.5">
              {appointment.doctor}
            </div>
          ) : null}
          {height > 68 ? (
            <div className="truncate text-sm opacity-80 mt-auto leading-none pb-0.5">
              {getReservationTypeName(appointment.type)}
            </div>
          ) : null}
        </div>
      ) : null}
    </motion.button>
  );
});
