// React/Framework
import { useCallback } from "react";
import { useNavigate } from "react-router";

// External
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import Clock from "lucide-react/dist/esm/icons/clock";
import Dog from "lucide-react/dist/esm/icons/dog";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import Scissors from "lucide-react/dist/esm/icons/scissors";
import Calendar from "lucide-react/dist/esm/icons/calendar";
import AlertCircle from "lucide-react/dist/esm/icons/alert-circle";
import Syringe from "lucide-react/dist/esm/icons/syringe";
import Activity from "lucide-react/dist/esm/icons/activity";
import FileText from "lucide-react/dist/esm/icons/file-text";
import CreditCard from "lucide-react/dist/esm/icons/credit-card";
import BedDouble from "lucide-react/dist/esm/icons/bed-double";

// Internal
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { C } from "@/lib/design-tokens";

// Types
import type { Appointment } from "@/types";

interface ServiceIconProps {
  service: string;
}

function ServiceIcon({ service }: ServiceIconProps) {
  if (service.includes("トリミング")) return <Scissors className="size-3" />;
  if (service.includes("ワクチン")) return <Syringe className="size-3" />;
  if (service.includes("手術")) return <Activity className="size-3" />;
  if (service.includes("診療")) return <Stethoscope className="size-3" />;
  return <Stethoscope className="size-3" />;
}

interface AppointmentCardProps {
  appointment: Appointment;
  columnTitle: string;
  onCardClick: (appointment: Appointment) => void;
  isDragOverlay?: boolean;
}

export function AppointmentCard({
  appointment,
  columnTitle,
  onCardClick,
  isDragOverlay = false,
}: AppointmentCardProps) {
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

  const isTrimming = appointment.serviceType.includes("トリミング");
  const isHospitalization = appointment.serviceType.includes("入院");

  const handleKarteClick = useCallback((e: React.MouseEvent) => {
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
  }, [navigate, isTrimming, appointment.petId, appointment.id]);

  const handleAccountingClick = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    navigate(appointment.petId ? `/accounting/new?petId=${appointment.petId}` : "/accounting/new", {
      state: { from: "/", appointmentId: appointment.id },
    });
  }, [navigate, appointment.petId, appointment.id]);

  const handleHospitalizationClick = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    navigate(appointment.petId ? `/hospitalization/new?petId=${appointment.petId}` : "/hospitalization/new", {
      state: { from: "/" },
    });
  }, [navigate, appointment.petId]);

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      className="cursor-grab active:cursor-grabbing group touch-none"
      onClick={() => onCardClick(appointment)}
    >
      <Card className={`w-full ${C.hoverBgPage} transition-colors border ${C.borderLight} rounded-[6px] shadow-[0px_1px_3px_0px_rgba(0,0,0,0.1),0px_1px_2px_0px_rgba(0,0,0,0.06)]`}>
        <CardContent className="p-[13px] space-y-[9px]">
          <div className="flex items-start justify-between gap-2">
            <div className={`flex items-center gap-1.5 ${C.text60} min-w-0`}>
              <Clock className="size-3.5 flex-shrink-0" />
              <span className="text-base font-medium font-mono tracking-[var(--tracking-notion)]">{appointment.time}</span>
            </div>
            {appointment.nextAppointment && (
              <Badge
                variant={appointment.nextAppointment === "精算未確認" ? "destructive" : "secondary"}
                className="text-sm px-[7.5px] h-[22px] flex items-center gap-0.5 flex-shrink-0 tracking-[var(--tracking-notion-sm)]"
              >
                {appointment.nextAppointment === "精算未確認" && <AlertCircle className="size-3" />}
                {appointment.nextAppointment === "次回予約済" && <Calendar className="size-3" />}
                {appointment.nextAppointment}
              </Badge>
            )}
          </div>

          <div className="space-y-0.5">
            <p className="text-base font-semibold truncate leading-tight tracking-[var(--tracking-notion)]">{appointment.ownerName}</p>
            <div className={`flex items-center gap-1 ${C.text60}`}>
              <Dog className="size-3.5 flex-shrink-0" />
              <p className="text-base truncate tracking-[var(--tracking-notion)]">{appointment.petType} - {appointment.petName}</p>
            </div>
          </div>

          <div className="flex items-center flex-wrap gap-1 pt-0.5">
            <Badge
              variant="secondary"
              className={`text-sm px-[7.5px] h-[22px] tracking-[var(--tracking-notion-sm)] ${appointment.visitType === "初診" ? `bg-[#D3E5EF]/60 text-[#183B56]/90 border-[#B8D4E3]/50` : `bg-[#F7F6F3]/60 text-[#37352F]/90 border-[rgba(55,53,47,0.09)]/50`}`}
            >
              {appointment.visitType}
            </Badge>
            <Badge variant="outline" className="flex items-center gap-1 text-sm px-[7.5px] h-[22px] bg-white tracking-[var(--tracking-notion-sm)]">
              <ServiceIcon service={appointment.serviceType} />
              <span className="truncate max-w-[80px]">{appointment.serviceType}</span>
            </Badge>

            {(appointment.doctor || appointment.isDesignated) && (
                 <Badge variant="outline" className={`flex items-center gap-1 text-sm px-[7.5px] h-[22px] tracking-[var(--tracking-notion-sm)] ${appointment.isDesignated ? `bg-[#FAEBDD] text-[#D9730D] border-[#D9730D]/20` : `bg-white text-[#37352F]/60`}`}>
                    <span className="truncate max-w-[80px]">{appointment.doctor || "指名あり"}</span>
                    {appointment.isDesignated && <span className="text-[10px] ml-0.5 font-bold tracking-[0.12em]">指</span>}
                 </Badge>
            )}
          </div>

          {/* ミニアクションボタン */}
          <div className="flex items-center gap-1 pt-0.5 border-t border-[rgba(55,53,47,0.06)]">
            <button
              type="button"
              aria-label={isTrimming ? `${appointment.petName}のトリミング記録` : `${appointment.petName}のカルテ`}
              className="flex items-center gap-1 text-[11px] tracking-[var(--tracking-notion-xs)] text-[#2383E2] bg-[#D3E5EF]/30 border border-[#B8D4E3]/40 rounded px-1.5 py-0.5 hover:bg-[#D3E5EF]/60 transition-colors"
              onClick={handleKarteClick}
            >
              {isTrimming ? <Scissors className="size-3 shrink-0" /> : <FileText className="size-3 shrink-0" />}
              <span>{isTrimming ? "施術" : "カルテ"}</span>
            </button>
            {columnTitle !== "診療中" && (
              <button
                type="button"
                aria-label={`${appointment.petName}の会計`}
                className="flex items-center gap-1 text-[11px] tracking-[var(--tracking-notion-xs)] text-[#0F7B6C] bg-[#DDEDEA]/30 border border-[#DDEDEA]/40 rounded px-1.5 py-0.5 hover:bg-[#DDEDEA]/60 transition-colors"
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
                className="flex items-center gap-1 text-[11px] tracking-[var(--tracking-notion-xs)] text-[#6940A5] bg-[#EEE0F7]/30 border border-[#6940A5]/20 rounded px-1.5 py-0.5 hover:bg-[#EEE0F7]/60 transition-colors"
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
}
