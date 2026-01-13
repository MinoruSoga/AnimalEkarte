import { useRef } from "react";
import { useDrag, useDrop } from "react-dnd";
import { Card, CardContent } from "../../../components/ui/card";
import { Badge } from "../../../components/ui/badge";
import { Clock, Dog, Stethoscope, Scissors, Calendar, AlertCircle, Syringe, Activity } from "lucide-react";
import { Appointment } from "../../../types";

interface AppointmentCardProps {
  appointment: Appointment;
  columnTitle: string;
  index: number;
  moveCard: (dragIndex: number, hoverIndex: number, sourceColumn: string, targetColumn: string) => boolean;
  onCardClick: (appointment: Appointment) => void;
}

export const AppointmentCard = ({ 
  appointment, 
  columnTitle, 
  index,
  moveCard,
  onCardClick
}: AppointmentCardProps) => {
  const ref = useRef<HTMLDivElement>(null);

  const [{ isDragging }, drag] = useDrag(() => ({
    type: 'appointment',
    item: { index, columnTitle, appointment },
    collect: (monitor) => ({
      isDragging: !!monitor.getItem() && monitor.getItem().appointment?.id === appointment.id,
    }),
  }));

  const [{ isOver }, drop] = useDrop(() => ({
    accept: 'appointment',
    hover: (item: any, monitor) => {
      if (!ref.current) return;

      const dragIndex = item.index;
      const hoverIndex = index;
      const sourceColumn = item.columnTitle;
      const targetColumn = columnTitle;

      if (dragIndex === hoverIndex && sourceColumn === targetColumn) {
        return;
      }

      const hoverBoundingRect = ref.current?.getBoundingClientRect();
      const hoverMiddleY = (hoverBoundingRect.bottom - hoverBoundingRect.top) / 2;
      const clientOffset = monitor.getClientOffset();
      const hoverClientY = clientOffset!.y - hoverBoundingRect.top;

      // Logic for determining insertion point
      let insertIndex = hoverIndex;

      if (sourceColumn === targetColumn) {
        // Standard sorting logic within same column
        if (dragIndex < hoverIndex && hoverClientY < hoverMiddleY) return;
        if (dragIndex > hoverIndex && hoverClientY > hoverMiddleY) return;
        
        // For same column, we swap, so target index is hoverIndex
        insertIndex = hoverIndex;
      } else {
        // Cross-column: allow inserting before or after the hovered card
        if (hoverClientY > hoverMiddleY) {
          insertIndex = hoverIndex + 1;
        }
      }

      if (moveCard(dragIndex, insertIndex, sourceColumn, targetColumn)) {
        item.index = insertIndex;
        item.columnTitle = targetColumn;
      }
    },
    collect: (monitor) => ({
      isOver: monitor.isOver(),
    }),
  }));

  drag(drop(ref));

  const getServiceIcon = (service: string) => {
    if (service.includes("トリミング")) return <Scissors className="size-3" />;
    if (service.includes("ワクチン")) return <Syringe className="size-3" />;
    if (service.includes("手術")) return <Activity className="size-3" />;
    if (service.includes("診療")) return <Stethoscope className="size-3" />;
    return <Stethoscope className="size-3" />;
  };

  return (
    <div 
      ref={ref} 
      style={{ opacity: isDragging ? 0.5 : 1 }} 
      className="cursor-grab active:cursor-grabbing group"
      onClick={() => onCardClick(appointment)}
    >
      <Card className={`w-full hover:bg-accent/30 transition-colors border border-border shadow-[0_1px_3px_rgba(0,0,0,0.05)] hover:shadow-[0_2px_6px_rgba(0,0,0,0.08)] ${isOver ? 'ring-1 ring-primary/20' : ''}`}>
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
