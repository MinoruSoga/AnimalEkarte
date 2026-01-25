// React/Framework
import { useMemo } from "react";

// External
import { startOfWeek, addDays, format, isSameDay } from "date-fns";
import { ja } from "date-fns/locale";
import { motion } from "motion/react";

// Internal
import { getReservationTypeColor, getReservationTypeName } from "../../../lib/status-helpers";

// Types
import type { ReservationAppointment } from "../../../types";

// Reduced height for High Density UI
const HOUR_HEIGHT = 120;
const HOURS = Array.from({ length: 24 }, (_, i) => i);

interface WeekViewProps {
  currentDate: Date;
  appointments: ReservationAppointment[];
  onAppointmentClick: (appointment: ReservationAppointment) => void;
  onTimeSlotClick?: (date: Date) => void;
  onAppointmentUpdate?: (appointment: ReservationAppointment, newStart: Date, newEnd: Date) => void;
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
const TimeSidebar = () => {
  return (
    <div className="w-12 flex-shrink-0 flex flex-col bg-[#F7F6F3] border-r border-[rgba(55,53,47,0.16)] z-30 sticky left-0">
      {HOURS.map((hour) => (
        <div
          key={hour}
          className="text-sm text-[#37352F]/40 text-right pr-2 pt-1 relative leading-none"
          style={{ height: `${HOUR_HEIGHT}px` }}
        >
          {hour}:00
          <div className="absolute top-0 right-0 w-1.5 h-[1px] bg-[rgba(55,53,47,0.16)]" />
        </div>
      ))}
    </div>
  );
};

// Sub-component: Appointment Card
const AppointmentCard = ({
  appointment,
  layoutStyle,
  onClick,
  onUpdate,
}: {
  appointment: ReservationAppointment;
  layoutStyle: { left: string; width: string };
  onClick: (appointment: ReservationAppointment) => void;
  onUpdate?: (appointment: ReservationAppointment, newStart: Date, newEnd: Date) => void;
}) => {
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
  const maxBottom = 24 * HOUR_HEIGHT;
  const dragConstraints = {
      top: -top,
      bottom: maxBottom - (top + height)
  };

  const isDimmed = appointment.status === 'completed' || appointment.status === 'cancelled';
  const isCancelled = appointment.status === 'cancelled';

  // Status indicator dot
  const getStatusDot = () => {
      switch (appointment.status) {
          case 'checked_in': return <div className="w-2.5 h-2.5 rounded-full bg-blue-500 border border-white absolute top-1 right-1 shadow-sm z-20" title="受付済" />;
          case 'in_consultation': return <div className="w-2.5 h-2.5 rounded-full bg-purple-500 border border-white absolute top-1 right-1 shadow-sm z-20" title="診療中" />;
          case 'accounting': return <div className="w-2.5 h-2.5 rounded-full bg-orange-500 border border-white absolute top-1 right-1 shadow-sm z-20" title="会計待ち" />;
          case 'completed': return <div className="w-2.5 h-2.5 rounded-full bg-gray-400 border border-white absolute top-1 right-1 shadow-sm z-20" title="完了" />;
          case 'cancelled': return <div className="w-2.5 h-2.5 rounded-full bg-red-500 border border-white absolute top-1 right-1 shadow-sm z-20" title="キャンセル" />;
          default: return null;
      }
  };

  return (
    <motion.div
      className={`absolute rounded border p-1 text-sm shadow-sm hover:shadow-md transition-all cursor-grab active:cursor-grabbing z-10 overflow-hidden leading-tight group 
        ${getReservationTypeColor(appointment.type)}
        ${isDimmed ? "opacity-60" : "opacity-100"}
        ${isCancelled ? "line-through decoration-red-500/50" : ""}
      `}
      style={{
        top: `${top}px`,
        height: `${height}px`,
        left: layoutStyle.left,
        width: layoutStyle.width,
      }}
      onClick={(e) => {
        e.stopPropagation();
        onClick(appointment);
      }}
      drag="y"
      dragMomentum={false}
      dragElastic={0}
      dragConstraints={dragConstraints}
      onDragEnd={(_, info) => {
        // Calculate moved minutes based on Y offset
        // Round to nearest 15 minutes (15min = 30px height)
        const movedMinutes = (info.offset.y / HOUR_HEIGHT) * 60;
        const snappedMinutes = Math.round(movedMinutes / 15) * 15;
        
        if (snappedMinutes === 0 || !onUpdate) return;

        const newStart = new Date(appointment.start.getTime() + snappedMinutes * 60000);
        const newEnd = new Date(appointment.end.getTime() + snappedMinutes * 60000);
        
        onUpdate(appointment, newStart, newEnd);
      }}
      whileDrag={{ 
        zIndex: 50, 
        scale: 1.02, 
        opacity: 0.9,
        boxShadow: "0 10px 15px -3px rgba(0, 0, 0, 0.1)"
      }}
    >
      {getStatusDot()}
      
      <div className="flex flex-col h-full pointer-events-none relative z-10">
        <div className="flex items-center gap-1 font-bold text-sm leading-none mb-1 pr-3">
            <span className="truncate">
                {format(appointment.start, "H:mm")}
            </span>
            {appointment.visitType === "first" && (
                <span className="bg-red-100 text-red-700 px-1.5 rounded text-sm flex-shrink-0 no-underline">
                    初
                </span>
            )}
        </div>
        
        <div className="font-medium truncate text-sm leading-none">
            {appointment.petName}
        </div>
        
        {height > 40 && (
            <div className="truncate text-sm opacity-80 mt-auto leading-none pb-0.5">
                {getReservationTypeName(appointment.type)}
            </div>
        )}
      </div>
    </motion.div>
  );
};

// Sub-component: Day Column
const DayColumn = ({
  date,
  appointments,
  onAppointmentClick,
  onTimeSlotClick,
  onAppointmentUpdate,
}: {
  date: Date;
  appointments: ReservationAppointment[];
  onAppointmentClick: (appointment: ReservationAppointment) => void;
  onTimeSlotClick?: (date: Date) => void;
  onAppointmentUpdate?: (appointment: ReservationAppointment, newStart: Date, newEnd: Date) => void;
}) => {
  const layoutStyles = useMemo(
    () => calculateEventLayout(appointments),
    [appointments]
  );

  const isToday = isSameDay(date, new Date());

  // Current time indicator position
  const currentTimeTop =
    (new Date().getHours() * 60 + new Date().getMinutes()) / 60 * HOUR_HEIGHT;

  const handleColumnClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!onTimeSlotClick) return;

    const rect = e.currentTarget.getBoundingClientRect();
    const y = e.clientY - rect.top; // Y coordinate within the column

    // Calculate time
    const hoursFromStart = y / HOUR_HEIGHT;
    const totalMinutes = Math.floor(hoursFromStart * 60);
    
    // Create clicked date
    const clickedDate = new Date(date);
    clickedDate.setHours(0, 0, 0, 0); // Base 0:00
    clickedDate.setMinutes(totalMinutes);

    onTimeSlotClick(clickedDate);
  };

  return (
    <div
      className={`flex-1 flex-shrink-0 border-r border-[rgba(55,53,47,0.09)] relative ${
        isToday ? "bg-blue-50/10" : ""
      }`}
      style={{ minWidth: "20%" }}
      onClick={handleColumnClick}
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
      {isToday && (
        <div
          className="absolute w-full border-t-2 border-red-400 z-20 pointer-events-none"
          style={{ top: `${currentTimeTop}px` }}
        >
          <div className="absolute -left-1 -top-1.5 w-3 h-3 rounded-full bg-red-400" />
        </div>
      )}

      {/* Appointments */}
      {appointments.map((app) => (
        <AppointmentCard
          key={app.id}
          appointment={app}
          layoutStyle={layoutStyles[app.id] || { left: "0%", width: "100%" }}
          onClick={onAppointmentClick}
          onUpdate={onAppointmentUpdate}
        />
      ))}
    </div>
  );
};

// Main Component
export const WeekView = ({
  currentDate,
  appointments,
  onAppointmentClick,
  onTimeSlotClick,
  onAppointmentUpdate,
}: WeekViewProps) => {
  const startDate = startOfWeek(currentDate, { locale: ja });

  // Header days
  const headerDays = Array.from({ length: 7 }).map((_, i) => {
    const day = addDays(startDate, i);
    const isToday = isSameDay(day, new Date());
    return (
      <div
        key={i}
        className={`flex-1 flex-shrink-0 text-center py-2 border-r border-[rgba(55,53,47,0.16)] bg-white ${
          isToday ? "bg-blue-50/50" : ""
        }`}
        style={{ minWidth: "20%" }}
      >
        <div
          className={`text-sm mb-1 ${
            i === 0
              ? "text-red-500"
              : i === 6
              ? "text-blue-500"
              : "text-[#37352F]/60"
          }`}
        >
          {format(day, "E", { locale: ja })}
        </div>
        <div
          className={`text-lg font-bold ${
            isToday ? "text-blue-600" : "text-[#37352F]"
          }`}
        >
          {format(day, "d")}
        </div>
      </div>
    );
  });

  return (
    <div className="flex-1 border border-[rgba(55,53,47,0.16)] rounded-lg bg-white overflow-auto shadow-sm relative">
      <div style={{ minWidth: "100%" }}>
        {/* Week Header */}
        <div className="flex border-b border-[rgba(55,53,47,0.16)] sticky top-0 z-30 bg-white">
          <div className="w-12 flex-shrink-0 bg-[#F7F6F3] border-r border-[rgba(55,53,47,0.16)] sticky left-0 z-40" />
          <div className="flex flex-1">{headerDays}</div>
        </div>

        {/* Scrollable Grid */}
        <div className="flex relative">
          <TimeSidebar />

          {/* Day Columns */}
          <div className="flex relative flex-1">
            {Array.from({ length: 7 }).map((_, i) => {
              const day = addDays(startDate, i);
              const dayAppointments = appointments.filter((app) =>
                isSameDay(app.start, day)
              );

              return (
                <DayColumn
                  key={i}
                  date={day}
                  appointments={dayAppointments}
                  onAppointmentClick={onAppointmentClick}
                  onTimeSlotClick={onTimeSlotClick}
                  onAppointmentUpdate={onAppointmentUpdate}
                />
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
};
