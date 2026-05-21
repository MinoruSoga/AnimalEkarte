// React/Framework
import { memo, useRef, useCallback, useMemo } from "react";
import type { MouseEvent as ReactMouseEvent, TouchEvent as ReactTouchEvent } from "react";

// External
import { startOfWeek, addDays, isSameDay, format } from "date-fns";
import { ja } from "date-fns/locale";
import { motion } from "motion/react";

// Internal
import { getReservationTypeColor, getReservationTypeName } from "@/utils/status-helpers";
import { C, PALETTE } from "@/lib/design-tokens";
import { useReducedMotion } from "@/hooks/use-reduced-motion";

// Types
import type { Reservation, ReservationStatus } from "@/types";

// 15分予約が narrow モード（petName + ownerName 表示）に収まる最小高さ
// 15min = 40px (> isCompact 閾値 35px), 30min = 80px (full モード)
const HOUR_HEIGHT = 160;
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

interface ReservationTypeColor {
  style: React.CSSProperties;
  dotStyle: React.CSSProperties;
  hex: string;
}

interface WeekViewProps {
  currentDate: Date;
  appointments: Reservation[];
  onAppointmentClick: (reservation: Reservation) => void;
  onTimeSlotClick?: (date: Date) => void;
  onAppointmentUpdate?: (reservation: Reservation, newStart: Date, newEnd: Date) => void;
  /** Dynamic color map from reservationType master (name → ReservationTypeColor) */
  dynamicColorMap?: Map<string, ReservationTypeColor>;
  /** 表示モード: 5=週全体スクロール表示（デフォルト）, 7=全曜日を一画面表示 */
  days?: 5 | 7;
}

// Helper: Calculate event layout (overlapping)
const calculateEventLayout = (
  dayAppointments: Reservation[]
): Record<string, { left: string; width: string }> => {
  const sorted = [...dayAppointments].sort(
    (a, b) => a.start.getTime() - b.start.getTime()
  );
  const clusters: Reservation[][] = [];
  let currentCluster: Reservation[] = [];
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
};

// Sub-component: Time Sidebar
const TimeSidebar = memo(function TimeSidebar() {
  return (
    <div className={`w-12 flex-shrink-0 flex flex-col ${C.bgPage} border-r ${C.borderMedium} z-30 sticky left-0`}>
      {HOURS.map((hour) => (
        <div
          key={hour}
          className={`relative flex-shrink-0 text-xs ${C.text40} text-right pr-2 pt-0.5 leading-none`}
          style={{ height: `${HOUR_HEIGHT}px` }}
        >
          {hour}:00
          <div className={`absolute top-0 right-0 w-1.5 h-[1px] ${C.bgLight}`} />
          {/* :30 half-hour label */}
          <div
            className={`absolute right-0 pr-2 text-xs leading-none`}
            style={{ top: `${HOUR_HEIGHT / 2}px`, transform: "translateY(-50%)", opacity: 0.45 }}
          >
            :30
          </div>
          <div
            className={`absolute right-0 w-[5px] h-[1px] ${C.bgLight}`}
            style={{ top: `${HOUR_HEIGHT / 2}px` }}
          />
        </div>
      ))}
    </div>
  );
});

// Sub-component: Reservation Card
const AppointmentCard = memo(function AppointmentCard({
  appointment,
  layoutStyle,
  onClick,
  onUpdate,
  dynamicColorMap,
}: {
  appointment: Reservation;
  layoutStyle: { left: string; width: string };
  onClick: (reservation: Reservation) => void;
  onUpdate?: (reservation: Reservation, newStart: Date, newEnd: Date) => void;
  dynamicColorMap?: Map<string, ReservationTypeColor>;
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
  // 15min = 40px → compact (single-line), 30min = 80px → narrow (multi-line)
  const isCompact = height <= 44;
  const isNarrow = !isCompact && height <= 85;

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

  const colorStyle = getReservationTypeColor(appointment.type, dynamicColorMap);
  const isClassNameColor = typeof colorStyle === "string";

  return (
    <motion.div
      className={`absolute rounded border hover:ring-1 ${C.ringPrimary20} transition-all cursor-grab active:cursor-grabbing z-10 overflow-hidden group touch-none
        ${isCompact ? "py-px px-1" : isNarrow ? "px-1 py-0.5" : "p-1"}
        ${isClassNameColor ? colorStyle : ""}
        ${isDimmed ? "opacity-60" : "opacity-100"}
        ${isCancelled ? `line-through ${C.decorationDanger50}` : ""}
      `}
      role="button"
      aria-label={`${format(appointment.start, "H:mm")}〜${format(appointment.end, "H:mm")} ${appointment.petName} ${appointment.ownerName} ${getReservationTypeName(appointment.type)}`}
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
});

// Sub-component: Day Column
const DayColumn = memo(function DayColumn({
  date,
  appointments,
  onAppointmentClick,
  onTimeSlotClick,
  onAppointmentUpdate,
  dynamicColorMap,
  minWidth,
}: {
  date: Date;
  appointments: Reservation[];
  onAppointmentClick: (reservation: Reservation) => void;
  onTimeSlotClick?: (date: Date) => void;
  onAppointmentUpdate?: (reservation: Reservation, newStart: Date, newEnd: Date) => void;
  dynamicColorMap?: Map<string, ReservationTypeColor>;
  minWidth?: string;
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
      className={`flex-1 flex-shrink-0 border-r ${C.borderMedium} relative ${
        isToday ? C.bgAccentLight8 : ""
      }`}
      style={{ minWidth: minWidth ?? "20%" }}
      onClick={handleColumnClick}
      onTouchStart={handleTouchStart}
      onTouchEnd={handleTouchEnd}
    >
      {/* Grid Lines — :00 solid, :30 dashed, :15/:45 dotted */}
      <div className="absolute inset-0 flex flex-col pointer-events-none z-0">
        {HOURS.map((_, h) => (
          <div
            key={h}
            className="relative w-full flex-shrink-0"
            style={{ height: `${HOUR_HEIGHT}px` }}
          >
            {/* :15 dotted */}
            <div
              className={`absolute left-0 right-0 border-t border-dotted ${C.borderDivider}`}
              style={{ top: `${HOUR_HEIGHT / 4}px`, opacity: 0.3 }}
            />
            {/* :30 dashed */}
            <div
              className={`absolute left-0 right-0 border-t border-dashed ${C.borderDivider}`}
              style={{ top: `${HOUR_HEIGHT / 2}px`, opacity: 0.55 }}
            />
            {/* :45 dotted */}
            <div
              className={`absolute left-0 right-0 border-t border-dotted ${C.borderDivider}`}
              style={{ top: `${(HOUR_HEIGHT * 3) / 4}px`, opacity: 0.3 }}
            />
            {/* :00 solid hour line */}
            <div className={`absolute bottom-0 left-0 right-0 border-b ${C.borderDivider}`} />
          </div>
        ))}
      </div>

      {/* Current Time Indicator */}
      {isToday ? (
        <div
          className="absolute w-full border-t-2 z-20 pointer-events-none"
          style={{ borderColor: PALETTE.danger, top: `${currentTimeTop}px` }}
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
});

// Main Component
export const WeekView = memo(function WeekView({
  currentDate,
  appointments,
  onAppointmentClick,
  onTimeSlotClick,
  onAppointmentUpdate,
  dynamicColorMap,
  days = 5,
}: WeekViewProps) {
  // days=5: 全7曜日を20%幅でレンダリング（5列が表示域に収まり残りはスクロール）— 以前のデフォルト動作
  // days=7: 全7曜日を等幅でレンダリング（全曜日が一画面に収まる）
  const columnWidth = days === 7 ? `${(100 / 7).toFixed(4)}%` : "20%";

  const startDate = useMemo(
    () => startOfWeek(currentDate, { locale: ja }),
    [currentDate]
  );

  // Pre-compute appointments grouped by date key for O(1) per-day lookup
  const appointmentsByDay = useMemo(() => {
    const map = new Map<string, Reservation[]>();
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

  const headerDays = useMemo(
    () =>
      WEEK_DAYS.map((offset) => {
        const day = addDays(startDate, offset);
        const isToday = isSameDay(day, new Date());
        const count = appointmentsByDay.get(format(day, "yyyy-MM-dd"))?.length ?? 0;
        const dow = day.getDay(); // 0=Sun, 6=Sat
        return (
          <div
            key={offset}
            data-testid="week-header-day"
            className={`flex-1 flex-shrink-0 text-center py-2 border-r border-b ${C.borderMedium} ${C.bgWhite} ${
              isToday ? C.bgAccentLight30 : ""
            }`}
            style={{ minWidth: columnWidth }}
          >
            <div
              className={`text-sm mb-1 ${
                dow === 0
                  ? C.textNotionRed
                  : dow === 6
                  ? C.accent
                  : C.text60
              }`}
            >
              {format(day, "E", { locale: ja })}
            </div>
            <div className="flex items-center justify-center">
              <div
                className={`relative text-lg font-bold ${
                  isToday ? C.accent : C.text
                }`}
              >
                {format(day, "d")}
                {count > 0 ? (
                  <span
                    className={`absolute -right-7 bottom-0 text-xs whitespace-nowrap ${C.accent}`}
                  >
                    {count}件
                  </span>
                ) : null}
              </div>
            </div>
          </div>
        );
      }),
    [startDate, appointmentsByDay, columnWidth]
  );

  return (
    <div className={`flex-1 border ${C.borderMedium} rounded-lg ${C.bgWhite} overflow-auto relative`}>
      <div style={{ minWidth: "100%" }}>
        {/* Week Header */}
        <div className={`flex border-b ${C.borderMedium} sticky top-0 z-30 ${C.bgWhite}`}>
          <div className={`w-12 flex-shrink-0 ${C.bgPage} border-r border-b ${C.borderMedium} sticky left-0 z-40`} />
          <div className="flex flex-1">{headerDays}</div>
        </div>

        {/* Scrollable Grid */}
        <div className="flex relative">
          <TimeSidebar />

          {/* Day Columns */}
          <div className="flex relative flex-1">
            {WEEK_DAYS.map((offset) => {
              const day = addDays(startDate, offset);
              const dayKey = format(day, "yyyy-MM-dd");
              const dayAppointments = appointmentsByDay.get(dayKey) ?? [];

              return (
                <DayColumn
                  key={offset}
                  date={day}
                  appointments={dayAppointments}
                  onAppointmentClick={onAppointmentClick}
                  onTimeSlotClick={onTimeSlotClick}
                  onAppointmentUpdate={onAppointmentUpdate}
                  dynamicColorMap={dynamicColorMap}
                  minWidth={columnWidth}
                />
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
});
