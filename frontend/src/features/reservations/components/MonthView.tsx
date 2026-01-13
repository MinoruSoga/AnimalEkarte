import { startOfMonth, endOfMonth, startOfWeek, endOfWeek, format, isSameMonth, isSameDay, addDays } from "date-fns";
import { ja } from "date-fns/locale";
import { ReservationAppointment } from "../../../types";
import { getReservationTypeColor } from "../../../lib/status-helpers";

interface MonthViewProps {
  currentDate: Date;
  appointments: ReservationAppointment[];
  onAppointmentClick: (appointment: ReservationAppointment) => void;
}

export const MonthView = ({ currentDate, appointments, onAppointmentClick }: MonthViewProps) => {
  const monthStart = startOfMonth(currentDate);
  const monthEnd = endOfMonth(monthStart);
  const startDate = startOfWeek(monthStart, { locale: ja });
  const endDate = endOfWeek(monthEnd, { locale: ja });

  const dateFormat = "d";
  const rows = [];
  let days = [];
  let day = startDate;
  let formattedDate = "";

  // Header Row
  const daysOfWeek = ["日", "月", "火", "水", "木", "金", "土"];
  const headerRow = (
    <div className="grid grid-cols-7 border-b border-[rgba(55,53,47,0.16)] bg-[#F7F6F3]">
      {daysOfWeek.map((d, i) => (
        <div key={i} className={`py-3 text-sm font-bold text-center ${i === 0 ? "text-red-500" : i === 6 ? "text-blue-500" : "text-[#37352F]/60"}`}>
          {d}
        </div>
      ))}
    </div>
  );

  while (day <= endDate) {
    for (let i = 0; i < 7; i++) {
      formattedDate = format(day, dateFormat);
      const cloneDay = day;
      
      const dayAppointments = appointments.filter(app => isSameDay(app.start, cloneDay));

      days.push(
        <div
          key={day.toString()}
          className={`h-full min-h-[140px] bg-white border-b border-r border-[rgba(55,53,47,0.09)] p-2 transition-colors hover:bg-[#F7F6F3] cursor-pointer flex flex-col
            ${!isSameMonth(day, monthStart) ? "bg-[#F7F6F3]/30 text-[#37352F]/30" : "text-[#37352F]"}
            ${isSameDay(day, new Date()) ? "bg-blue-50/30" : ""}
          `}
        >
          <div className="flex justify-between items-start mb-2">
              <span className={`text-base font-bold size-7 flex items-center justify-center rounded-full ${isSameDay(day, new Date()) ? "bg-blue-600 text-white shadow-sm" : ""}`}>
                  {formattedDate}
              </span>
          </div>
          <div className="space-y-1.5 flex-1 overflow-hidden">
              {dayAppointments.slice(0, 4).map(app => (
                  <div 
                      key={app.id} 
                      className={`text-sm px-2 py-1.5 rounded border ${getReservationTypeColor(app.type)} truncate cursor-pointer hover:opacity-80 leading-tight`}
                      onClick={(e) => {
                          e.stopPropagation();
                          onAppointmentClick(app);
                      }}
                  >
                      <span className="font-bold mr-1.5">{format(app.start, "H:mm")}</span>
                      {app.petName}
                  </div>
              ))}
              {dayAppointments.length > 4 && (
                  <div className="text-sm text-[#37352F]/60 pl-1">
                      他 {dayAppointments.length - 4} 件
                  </div>
              )}
          </div>
        </div>
      );
      day = addDays(day, 1);
    }
    rows.push(
      <div className="grid grid-cols-7 flex-1" key={day.toString()}>
        {days}
      </div>
    );
    days = [];
  }

  return (
    <div className="flex flex-col h-full border-l border-t border-[rgba(55,53,47,0.16)] rounded-lg overflow-hidden bg-white shadow-sm">
      {headerRow}
      <div className="flex-1 flex flex-col">
        {rows}
      </div>
    </div>
  );
};
