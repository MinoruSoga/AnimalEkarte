// React/Framework
import { useRef, useCallback, useMemo } from "react";
import type { MouseEvent as ReactMouseEvent, TouchEvent as ReactTouchEvent } from "react";

// External
import { startOfWeek, addDays, isSameDay, format } from "date-fns";
import { ja } from "date-fns/locale";
import { motion } from "motion/react";

// Internal
import { getReservationTypeColor, getReservationTypeName } from "@/utils/status-helpers";
import { C, PALETTE } from "@/lib/design-tokens";
import { useReducedMotion } from "@/hooks/useReducedMotion";

// Types
import type { ReservationAppointment, ReservationStatus } from "@/types";

// Reduced height for High Density UI
const HOUR_HEIGHT = 120;
const HOURS = Array.from({ length: 24 }, (_, i) => i);

const WHILE_DRAG_FULL = {
  zIndex: 50,
  scale: 1.02,
  opacity: 0.9,
  boxShadow: "0 10px 30px rgba(0,0,0,0.15)",
} as const;

const WHILE_DRAG_REDUCED = { zIndex: 50 } as const;

const WEEK_DAYS = [0, 1, 2, 3, 4, 5, 6] as const;

const STATUS_DOT_STYLE: Partial<Record<ReservationStatus, { color: string; label: string }>> = {
  checked_in:      { color: C.bgAccent, label: "受付済" },
  in_consultation: { color: C.bgStatusPurpleDot, label: "診療中" },
  accounting:      { color: C.bgDiscount, label: "会計待ち" },
  completed:       { color: C.bgStatusGrayMedium, label: "完了" },
  cancelled:       { color: C.bgNotionRed, label: "キャンセル" },
};

interface WeekViewProps {
  currentDate: Date;
  appointments: ReservationAppointment[];
  onAppointmentClick: (appointment: ReservationAppointment) => void;
  onTimeSlotClick?: (date: Date) => void;
  onAppointmentUpdate?: (appointment: ReservationAppointment, newStart: Date, newEnd: Date) => void;
  /** Dynamic color map from serviceType master (name → "bg-xx text-xx border-xx") */
  dynamicColorMap?: Map<string, string>;
}

// Helper: Calculate event layout (overlapping)
const calculateEventLayout = (
  dayAppointments: ReservationAppointment[]
): Record<string, { left: string; width: string }> => {
  const sorted = [...dayAppointments].sort(
    (a, b) => a.start.getTime() - b.start.getTime()
  );
  const clusters: ReservationAppointment[][] = [];
  let currentCluster: ReservationAppointment[] = [];
  let clusterEnd = 0;

  sorted.forEach((ev) => {
    if (currentCluster.length === 0) {
      currentCluster.push(ev);
      clusterEnd = ev.end.getTime();
    } else {
      if (ev.start.getTime() < clusterEnd) {
        currentCluster.push(ev);
        if (ev.end.getTime() > clusterEnd) clusterEnd = ev.end.getTime();
      } else {
        clusters.push(currentCluster);
        currentCluster = [ev];
        clusterEnd = ev.end.getTime();
      }
    }
  });
  if (currentCluster.length > 0) clusters.push(currentCluster);

  const styles: Record<string, { left: string; width: string }> = {};

  clusters.forEach((cluster) => {
    const columns: ReservationAppointment[][] = [];
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

    const w = 100 / columns.length;
    cluster.forEach((ev) => {
      const colIndex = eventColIndex[ev.id];
      styles[ev.id] = {
        left: `${colIndex * w}%`,
        width: `${w}%`,
      };
    });
  });

  return styles;
};

// Sub-component: Time Sidebar
function TimeSidebar() {
  return (
    <div className={`w-12 flex-shrink-0 flex flex-col ${C.bgPage} border-r border-[rgba(55,53,47,0.16)] z-30 sticky left-0`}>
      {HOURS.map((hour) => (
        <div
          key={hour}
          className={`text-sm ${C.text40} text-right pr-2 pt-1 relative leading-none`}
          style={{ height: `${HOUR_HEIGHT}px` }}
        >
          {hour}:00
          <div className="absolute top-0 right-0 w-1.5 h-[1px]" style={{ backgroundColor: PALETTE.borderMedium }} />
        </div>
      ))}
    </div>
  );
}

// Sub-component: Appointment Card
function AppointmentCard({
  appointment,
  layoutStyle,
  onClick,
  onUpdate,
  dynamicColorMap,
}: {
  appointment: ReservationAppointment;
  layoutStyle: { left: string; width: string };
  onClick: (appointment: ReservationAppointment) => void;
  onUpdate?: (appointment: ReservationAppointment, newStart: Date, newEnd: Date) => void;
  dynamicColorMap?: Map<string, string>;
}) {
  const startHour = appointment.start.getHours();
  const startMin = appointment.start.getMinutes();
  const endHour = appointment.end.getHours();
  const endMin = appointment.end.getMinutes();

  const startMinutes = startHour * 60 + startMin;
  const durationMinutes = endHour * 60 + endMin - startMinutes;

  // Calculate height but ensure a minimum clickable area
  const height = Math.max((durationMinutes / 60) * HOUR_HEIGHT, 24);
  const top = (startMinutes / 60) * HOUR_HEIGHT;

  // Constraints for dragging
  const dragConstraints = useMemo(
    () => ({
      top: -top,
      bottom: 24 * HOUR_HEIGHT - (top + height),
    }),
    [top, height]
  );

  const isDimmed = appointment.status === "completed" || appointment.status === "cancelled";
  const isCancelled = appointment.status === "cancelled";

  // Layout mode based on available height
  const isCompact = height <= 35; // 15min
  const isNarrow = !isCompact && height <= 65; // 30min

  // Status indicator dot
  const dotInfo = STATUS_DOT_STYLE[appointment.status];

  // Tooltip
  const tooltipText = useMemo(
    () =>
      [
        `${format(appointment.start, "H:mm")}–${format(appointment.end, "H:mm")}`,
        appointment.petName,
        appointment.ownerName,
        appointment.doctor,
        getReservationTypeName(appointment.type),
      ]
        .filter(Boolean)
        .join(" / "),
    [appointment]
  );

  // Touch drag: distinguish taps from drags
  const isDragging = useRef(false);
  const reduced = useReducedMotion();

  const whileDragProps = reduced ? WHILE_DRAG_REDUCED : WHILE_DRAG_FULL;

  return (
    <motion.div
      className={`absolute rounded border hover:ring-1 ${C.ringPrimary20} transition-all cursor-grab active:cursor-grabbing z-10 overflow-hidden group touch-none
        ${isCompact ? "py-px px-1" : isNarrow ? "px-1 py-0.5" : "p-1"}
        ${getReservationTypeColor(appointment.type, dynamicColorMap)}
        ${isDimmed ? "opacity-60" : "opacity-100"}
        ${isCancelled ? "line-through decoration-red-500/50" : ""}
      `}
      role="button"
      aria-label={`${format(appointment.start, "H:mm")}〜${format(appointment.end, "H:mm")} ${appointment.petName} ${appointment.ownerName} ${getReservationTypeName(appointment.type)}`}
      title={tooltipText}
      style={{
        top: `${top}px`,
        height: `${height}px`,
        left: layoutStyle.left,
        width: layoutStyle.width,
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
      {/* Status dot */}
      {dotInfo ? (
        <div
          className={`rounded-full ${dotInfo.color} border border-white absolute shadow-sm z-20 ${
            isCompact
              ? "w-1.5 h-1.5 top-0.5 right-0.5"
              : "w-2.5 h-2.5 top-1 right-1"
          }`}
          title={dotInfo.label}
        />
      ) : null}

      {/* Compact: 15min — single horizontal line */}
      {isCompact ? (
        <div className="flex items-center gap-1 h-full pointer-events-none relative z-10 pr-2">
          <span className="font-bold text-xs whitespace-nowrap leading-none">
            {format(appointment.start, "H:mm")}
          </span>
          {appointment.visitType === "first" ? (
            <span className={`${C.bgRedLight} ${C.textNotionRed} px-0.5 rounded text-xs flex-shrink-0 no-underline leading-none`}>
              初
            </span>
          ) : null}
          <span className="truncate text-xs font-medium leading-none">
            {appointment.petName}
          </span>
        </div>
      ) : null}

      {/* Narrow: 30min — compact 3-line layout */}
      {isNarrow ? (
        <div className="flex flex-col h-full pointer-events-none relative z-10">
          <div className="flex items-center gap-1 pr-3 leading-none">
            <span className="font-bold text-xs whitespace-nowrap">
              {format(appointment.start, "H:mm")}
            </span>
            {appointment.visitType === "first" ? (
              <span className={`${C.bgRedLight} ${C.textNotionRed} px-0.5 rounded text-xs flex-shrink-0 no-underline`}>
                初
              </span>
            ) : null}
          </div>
          <div className="font-medium truncate text-xs leading-none mt-0.5">
            {appointment.petName}
          </div>
          <div className="truncate text-xs opacity-70 leading-none mt-0.5">
            {appointment.ownerName}
          </div>
        </div>
      ) : null}

      {/* Full: 60min+ — detailed multi-line layout */}
      {!isCompact && !isNarrow ? (
        <div className="flex flex-col h-full pointer-events-none relative z-10">
          <div className="flex items-center gap-1 font-bold text-sm leading-none mb-1 pr-3">
            <span className="truncate">
              {format(appointment.start, "H:mm")}
            </span>
            {appointment.visitType === "first" ? (
              <span className={`${C.bgRedLight} ${C.textNotionRed} px-1.5 rounded text-sm flex-shrink-0 no-underline`}>
                初
              </span>
            ) : null}
          </div>
          <div className="font-medium truncate text-sm leading-none">
            {appointment.petName}
          </div>
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
    </motion.div>
  );
}

// Sub-component: Day Column
function DayColumn({
  date,
  appointments,
  onAppointmentClick,
  onTimeSlotClick,
  onAppointmentUpdate,
  dynamicColorMap,
}: {
  date: Date;
  appointments: ReservationAppointment[];
  onAppointmentClick: (appointment: ReservationAppointment) => void;
  onTimeSlotClick?: (date: Date) => void;
  onAppointmentUpdate?: (appointment: ReservationAppointment, newStart: Date, newEnd: Date) => void;
  dynamicColorMap?: Map<string, string>;
}) {
  const layoutStyles = useMemo(
    () => calculateEventLayout(appointments),
    [appointments]
  );

  const isToday = isSameDay(date, new Date());

  // Current time indicator position
  const now = new Date();
  const currentTimeTop =
    ((now.getHours() * 60 + now.getMinutes()) / 60) * HOUR_HEIGHT;

  // Compute clicked time from a vertical coordinate
  const computeTimeFromY = useCallback(
    (y: number): Date => {
      const hoursFromStart = y / HOUR_HEIGHT;
      const totalMinutes = Math.round(Math.floor(hoursFromStart * 60) / 15) * 15;
      const clickedDate = new Date(date);
      clickedDate.setHours(0, 0, 0, 0);
      clickedDate.setMinutes(totalMinutes);
      return clickedDate;
    },
    [date]
  );

  const handleColumnClick = (e: ReactMouseEvent<HTMLDivElement>) => {
    if (!onTimeSlotClick) return;
    const rect = e.currentTarget.getBoundingClientRect();
    const y = e.clientY - rect.top;
    onTimeSlotClick(computeTimeFromY(y));
  };

  // Touch support: use touchEnd so it doesn't conflict with scroll
  const touchStartY = useRef<number | null>(null);
  const touchStartTime = useRef<number>(0);

  const handleTouchStart = useCallback((e: ReactTouchEvent<HTMLDivElement>) => {
    touchStartY.current = e.touches[0].clientY;
    touchStartTime.current = Date.now();
  }, []);

  const handleTouchEnd = useCallback(
    (e: ReactTouchEvent<HTMLDivElement>) => {
      if (!onTimeSlotClick || touchStartY.current === null) return;

      const touch = e.changedTouches[0];
      const dy = Math.abs(touch.clientY - touchStartY.current);
      const dt = Date.now() - touchStartTime.current;

      // Only treat as tap if small movement & short duration (not scroll)
      if (dy < 10 && dt < 300) {
        const rect = e.currentTarget.getBoundingClientRect();
        const y = touch.clientY - rect.top;
        onTimeSlotClick(computeTimeFromY(y));
      }

      touchStartY.current = null;
    },
    [onTimeSlotClick, computeTimeFromY]
  );

  return (
    <div
      className={`flex-1 flex-shrink-0 border-r border-[rgba(55,53,47,0.16)] relative ${
        isToday ? "bg-[#D3E5EF]/8" : ""
      }`}
      style={{ minWidth: "20%" }}
      onClick={handleColumnClick}
      onTouchStart={handleTouchStart}
      onTouchEnd={handleTouchEnd}
    >
      {/* Grid Lines */}
      <div className="absolute inset-0 flex flex-col pointer-events-none z-0">
        {HOURS.map((_, h) => (
          <div
            key={h}
            className="border-b border-[rgba(55,53,47,0.06)] w-full"
            style={{ height: `${HOUR_HEIGHT}px` }}
          />
        ))}
      </div>

      {/* Current Time Indicator */}
      {isToday ? (
        <div
          className={`absolute w-full border-t-2 border-red-400 z-20 pointer-events-none`}
          style={{ top: `${currentTimeTop}px` }}
        >
          <div className={`absolute -left-1 -top-1.5 w-3 h-3 rounded-full ${C.bgNotionRed}`} />
        </div>
      ) : null}

      {/* Appointments */}
      {appointments.map((app) => (
        <AppointmentCard
          key={app.id}
          appointment={app}
          layoutStyle={layoutStyles[app.id] || { left: "0%", width: "100%" }}
          onClick={onAppointmentClick}
          onUpdate={onAppointmentUpdate}
          dynamicColorMap={dynamicColorMap}
        />
      ))}
    </div>
  );
}

// Main Component
export function WeekView({
  currentDate,
  appointments,
  onAppointmentClick,
  onTimeSlotClick,
  onAppointmentUpdate,
  dynamicColorMap,
}: WeekViewProps) {
  const startDate = useMemo(
    () => startOfWeek(currentDate, { locale: ja }),
    [currentDate]
  );

  const headerDays = useMemo(
    () =>
      WEEK_DAYS.map((i) => {
        const day = addDays(startDate, i);
        const isToday = isSameDay(day, new Date());
        return (
          <div
            key={i}
            className={`flex-1 flex-shrink-0 text-center py-2 border-r border-b border-[rgba(55,53,47,0.16)] bg-white ${
              isToday ? "bg-[#D3E5EF]/30" : ""
            }`}
            style={{ minWidth: "20%" }}
          >
            <div
              className={`text-sm mb-1 ${
                i === 0
                  ? C.textNotionRed
                  : i === 6
                  ? C.accent
                  : C.text60
              }`}
            >
              {format(day, "E", { locale: ja })}
            </div>
            <div
              className={`text-lg font-bold ${
                isToday ? C.accent : C.text
              }`}
            >
              {format(day, "d")}
            </div>
          </div>
        );
      }),
    [startDate]
  );

  // Pre-compute appointments grouped by date key for O(1) per-day lookup
  const appointmentsByDay = useMemo(() => {
    const map = new Map<string, ReservationAppointment[]>();
    for (const app of appointments) {
      const key = format(app.start, "yyyy-MM-dd");
      const existing = map.get(key);
      if (existing) {
        existing.push(app);
      } else {
        map.set(key, [app]);
      }
    }
    return map;
  }, [appointments]);

  return (
    <div className={`flex-1 border border-[rgba(55,53,47,0.16)] rounded-lg bg-white overflow-auto relative`}>
      <div style={{ minWidth: "100%" }}>
        {/* Week Header */}
        <div className={`flex border-b border-[rgba(55,53,47,0.16)] sticky top-0 z-30 bg-white`}>
          <div className={`w-12 flex-shrink-0 ${C.bgPage} border-r border-b border-[rgba(55,53,47,0.16)] sticky left-0 z-40`} />
          <div className="flex flex-1">{headerDays}</div>
        </div>

        {/* Scrollable Grid */}
        <div className="flex relative">
          <TimeSidebar />

          {/* Day Columns */}
          <div className="flex relative flex-1">
            {WEEK_DAYS.map((i) => {
              const day = addDays(startDate, i);
              const dayKey = format(day, "yyyy-MM-dd");
              const dayAppointments = appointmentsByDay.get(dayKey) ?? [];

              return (
                <DayColumn
                  key={i}
                  date={day}
                  appointments={dayAppointments}
                  onAppointmentClick={onAppointmentClick}
                  onTimeSlotClick={onTimeSlotClick}
                  onAppointmentUpdate={onAppointmentUpdate}
                  dynamicColorMap={dynamicColorMap}
                />
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}
