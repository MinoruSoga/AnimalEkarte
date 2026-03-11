// React/Framework
import { useNavigate } from "react-router";

// External
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { Clock, Dog, Stethoscope, Scissors, Calendar, AlertCircle, Syringe, Activity, FileText, CreditCard, BedDouble } from "lucide-react";

// Internal
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { C } from "@/lib/design-tokens";

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
  const navigate = useNavigate();

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

  const isTrimming = appointment.serviceType.includes("トリミング");
  const isHospitalization = appointment.serviceType.includes("入院");

  const handleKarteClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (isTrimming) {
      navigate(appointment.petId ? `/trimming/new?petId=${appointment.petId}` : "/trimming/new", {
        state: { from: "/" },
      });
    } else {
      navigate(appointment.petId ? `/medical-records/new?petId=${appointment.petId}` : "/medical-records/select-pet", {
        state: { from: "/", appointmentId: appointment.id },
      });
    }
  };

  const handleAccountingClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    navigate(appointment.petId ? `/accounting/new?petId=${appointment.petId}` : "/accounting/new", {
      state: { from: "/", appointmentId: appointment.id },
    });
  };

  const handleHospitalizationClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    navigate(appointment.petId ? `/hospitalization/new?petId=${appointment.petId}` : "/hospitalization/new", {
      state: { from: "/" },
    });
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
      <Card className={`w-full ${C.hoverBgPage} transition-colors border ${C.borderLight}`}>
        <CardContent className="p-3 space-y-2">
          <div className="flex items-start justify-between gap-2">
            <div className={`flex items-center gap-1.5 ${C.text60} min-w-0`}>
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
            <div className={`flex items-center gap-1 ${C.text60}`}>
              <Dog className="size-3.5 flex-shrink-0" />
              <p className="text-sm truncate">{appointment.petType} - {appointment.petName}</p>
            </div>
          </div>

          <div className="flex items-center flex-wrap gap-1 pt-0.5">
            <Badge
              variant="secondary"
              className={`text-xs px-1.5 h-5 ${appointment.visitType === "初診" ? `bg-[#D3E5EF]/60 text-[#183B56]/90 border-[#B8D4E3]/50` : `bg-[#F7F6F3]/60 text-[#37352F]/90 border-[rgba(55,53,47,0.09)]/50`}`}
            >
              {appointment.visitType}
            </Badge>
            <Badge variant="outline" className="flex items-center gap-1 text-xs px-1.5 h-5 bg-white">
              {getServiceIcon(appointment.serviceType)}
              <span className="truncate max-w-[80px]">{appointment.serviceType}</span>
            </Badge>

            {(appointment.doctor || appointment.isDesignated) && (
                 <Badge variant="outline" className={`flex items-center gap-1 text-xs px-1.5 h-5 ${appointment.isDesignated ? `bg-[#FAEBDD] text-[#D9730D] border-[#D9730D]/20` : `bg-white text-[#37352F]/60`}`}>
                    <span className="truncate max-w-[80px]">{appointment.doctor || "指名あり"}</span>
                    {appointment.isDesignated && <span className="text-[10px] ml-0.5 font-bold">指</span>}
                 </Badge>
            )}
          </div>

          {/* ミニアクションボタン */}
          <div className="flex items-center gap-1 pt-0.5 border-t border-[rgba(55,53,47,0.06)]">
            <button
              type="button"
              aria-label={isTrimming ? `${appointment.petName}のトリミング記録` : `${appointment.petName}のカルテ`}
              className="flex items-center gap-1 text-[11px] text-[#2383E2] bg-[#D3E5EF]/30 border border-[#B8D4E3]/40 rounded px-1.5 py-0.5 hover:bg-[#D3E5EF]/60 transition-colors"
              onClick={handleKarteClick}
            >
              {isTrimming ? <Scissors className="size-3 shrink-0" /> : <FileText className="size-3 shrink-0" />}
              <span>{isTrimming ? "施術" : "カルテ"}</span>
            </button>
            {columnTitle !== "診療中" && (
              <button
                type="button"
                aria-label={`${appointment.petName}の会計`}
                className="flex items-center gap-1 text-[11px] text-[#0F7B6C] bg-[#DDEDEA]/30 border border-[#DDEDEA]/40 rounded px-1.5 py-0.5 hover:bg-[#DDEDEA]/60 transition-colors"
                onClick={handleAccountingClick}
              >
                <CreditCard className="size-3 shrink-0" />
                <span>会計</span>
              </button>
            )}
            {columnTitle !== "診療中" && isHospitalization && (
              <button
                type="button"
                aria-label={`${appointment.petName}の入院登録`}
                className="flex items-center gap-1 text-[11px] text-[#6940A5] bg-[#EEE0F7]/30 border border-[#6940A5]/20 rounded px-1.5 py-0.5 hover:bg-[#EEE0F7]/60 transition-colors"
                onClick={handleHospitalizationClick}
              >
                <BedDouble className="size-3 shrink-0" />
                <span>入院</span>
              </button>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
};
