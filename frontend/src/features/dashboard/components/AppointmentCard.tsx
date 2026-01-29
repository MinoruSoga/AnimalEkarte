// React/Framework

// External
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { Clock, Dog, Stethoscope, Scissors, Calendar, AlertCircle, Syringe, Activity } from "lucide-react";

// Internal
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

// Types
import type { Appointment } from "@/types";

interface AppointmentCardProps {
  appointment: Appointment;
  columnTitle: string;
  onCardClick: (appointment: Appointment) => void;
  isDragOverlay?: boolean;
}

export const AppointmentCard = ({
  appointment,
  columnTitle,
  onCardClick,
  isDragOverlay = false,
}: AppointmentCardProps) => {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: appointment.id,
    data: { columnTitle, appointment },
    disabled: isDragOverlay,
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const getServiceIcon = (service: string) => {
    if (service.includes("トリミング")) return <Scissors className="size-3" />;
    if (service.includes("ワクチン")) return <Syringe className="size-3" />;
    if (service.includes("手術")) return <Activity className="size-3" />;
    if (service.includes("診療")) return <Stethoscope className="size-3" />;
    return <Stethoscope className="size-3" />;
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      className="cursor-grab active:cursor-grabbing group touch-none"
      onClick={() => onCardClick(appointment)}
    >
      <Card className="w-full hover:bg-accent/30 transition-colors border border-border shadow-[0_1px_3px_rgba(0,0,0,0.05)] hover:shadow-[0_2px_6px_rgba(0,0,0,0.08)]">
        <CardContent className="p-3 space-y-2">
          <div className="flex items-start justify-between gap-2">
            <div className="flex items-center gap-1.5 text-muted-foreground min-w-0">
              <Clock className="size-3.5 flex-shrink-0" />
              <span className="text-sm font-medium font-mono">{appointment.time}</span>
            </div>
            {appointment.nextAppointment && (
              <Badge
                variant={appointment.nextAppointment === "精算未確認" ? "destructive" : "secondary"}
                className="text-xs px-1.5 h-5 flex items-center gap-0.5 flex-shrink-0"
              >
                {appointment.nextAppointment === "精算未確認" && <AlertCircle className="size-3" />}
                {appointment.nextAppointment === "次回予約済" && <Calendar className="size-3" />}
                {appointment.nextAppointment}
              </Badge>
            )}
          </div>

          <div className="space-y-0.5">
            <p className="text-sm font-semibold truncate leading-tight">{appointment.ownerName}</p>
            <div className="flex items-center gap-1 text-muted-foreground">
              <Dog className="size-3.5 flex-shrink-0" />
              <p className="text-sm truncate">{appointment.petType} - {appointment.petName}</p>
            </div>
          </div>

          <div className="flex items-center flex-wrap gap-1 pt-0.5">
            <Badge
              variant="secondary"
              className={`text-xs px-1.5 h-5 ${appointment.visitType === "初診" ? "bg-blue-50/60 text-blue-700/90 border-blue-200/50" : "bg-slate-50/60 text-slate-700/90 border-slate-200/50"}`}
            >
              {appointment.visitType}
            </Badge>
            <Badge variant="outline" className="flex items-center gap-1 text-xs px-1.5 h-5 bg-white">
              {getServiceIcon(appointment.serviceType)}
              <span className="truncate max-w-[80px]">{appointment.serviceType}</span>
            </Badge>

            {(appointment.doctor || appointment.isDesignated) && (
                 <Badge variant="outline" className={`flex items-center gap-1 text-xs px-1.5 h-5 ${appointment.isDesignated ? "bg-orange-50 text-orange-700 border-orange-200" : "bg-white text-gray-600"}`}>
                    <span className="truncate max-w-[80px]">{appointment.doctor || "指名あり"}</span>
                    {appointment.isDesignated && <span className="text-[10px] ml-0.5 font-bold">指</span>}
                 </Badge>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
};
