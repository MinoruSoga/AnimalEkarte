// React/Framework
import { memo, useCallback } from "react";
import { useNavigate } from "react-router";

// Internal
import { paths } from "@/config/paths";

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
import { C, ICON } from "@/lib/design-tokens";
import { getVisitTypeColor } from "@/utils/constants/status-colors";

// Types
import type { Appointment } from "@/types";

interface ServiceIconProps {
  service: string;
}

function ServiceIcon({ service }: ServiceIconProps) {
  if (service.includes("トリミング")) return <Scissors className={ICON.xs} />;
  if (service.includes("ワクチン")) return <Syringe className={ICON.xs} />;
  if (service.includes("手術")) return <Activity className={ICON.xs} />;
  if (service.includes("診療")) return <Stethoscope className={ICON.xs} />;
  return <Stethoscope className={ICON.xs} />;
}

interface AppointmentCardProps {
  appointment: Appointment;
  columnTitle: string;
  onCardClick: (appointment: Appointment) => void;
  isDragOverlay?: boolean;
}

export const AppointmentCard = memo(function AppointmentCard({
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

  const isTrimming = appointment.reservationCategory.includes("トリミング");
  const isHospitalization = appointment.reservationCategory.includes("入院");
  const visitColor = getVisitTypeColor(appointment.visitType);

  const handleKarteClick = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    if (isTrimming) {
      navigate(
        appointment.petId
          ? `${paths.trimming.new.getHref()}?petId=${appointment.petId}`
          : paths.trimming.new.getHref(),
        { state: { from: "/" } },
      );
    } else {
      navigate(
        appointment.petId
          ? `${paths.medicalRecords.new.getHref()}?petId=${appointment.petId}`
          : paths.medicalRecords.selectPet.getHref(),
        { state: { from: "/", appointmentId: appointment.id } },
      );
    }
  }, [navigate, isTrimming, appointment.petId, appointment.id]);

  const handleAccountingClick = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    navigate(
      appointment.petId
        ? `${paths.accounting.new.getHref()}?petId=${appointment.petId}`
        : paths.accounting.new.getHref(),
      { state: { from: "/", appointmentId: appointment.id } },
    );
  }, [navigate, appointment.petId, appointment.id]);

  const handleHospitalizationClick = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    navigate(
      appointment.petId
        ? `${paths.hospitalization.new.getHref()}?petId=${appointment.petId}`
        : paths.hospitalization.new.getHref(),
      { state: { from: "/" } },
    );
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
              <Clock className={`${ICON.xs} flex-shrink-0`} />
              <span className="text-base font-medium font-mono tracking-[var(--tracking-notion)]">{appointment.time}</span>
            </div>
            {appointment.nextAppointment ? (
              <Badge
                variant={appointment.nextAppointment === "精算未確認" ? "destructive" : "secondary"}
                className="text-sm px-[7.5px] h-[22px] flex items-center gap-0.5 flex-shrink-0 tracking-[var(--tracking-notion-sm)]"
              >
                {appointment.nextAppointment === "精算未確認" ? <AlertCircle className={ICON.xs} /> : null}
                {appointment.nextAppointment === "次回予約済" ? <Calendar className={ICON.xs} /> : null}
                {appointment.nextAppointment}
              </Badge>
            ) : null}
          </div>

          <div className="space-y-0.5">
            <p className="text-base font-semibold truncate leading-tight tracking-[var(--tracking-notion)]">{appointment.ownerName}</p>
            <div className={`flex items-center gap-1 ${C.text60}`}>
              <Dog className={`${ICON.xs} flex-shrink-0`} />
              <p className="text-base truncate tracking-[var(--tracking-notion)]">{appointment.petType} - {appointment.petName}</p>
            </div>
          </div>

          <div className="flex items-center flex-wrap gap-1 pt-0.5">
            <Badge
              variant="secondary"
              className={`text-sm px-[7.5px] h-[22px] tracking-[var(--tracking-notion-sm)] ${visitColor.badgeBg} ${visitColor.badgeText} ${visitColor.badgeBorder}`}
            >
              {appointment.visitType}
            </Badge>
            <Badge variant="outline" className="flex items-center gap-1 text-sm px-[7.5px] h-[22px] bg-white tracking-[var(--tracking-notion-sm)]">
              <ServiceIcon service={appointment.reservationCategory} />
              <span className="truncate max-w-[80px]">{appointment.reservationCategory}</span>
            </Badge>

            {/* BUG-037: 担当医バッジ — doctor が未設定でも「担当医未設定」として表示 */}
            <Badge variant="outline" className={`flex items-center gap-1 text-sm px-[7.5px] h-[22px] tracking-[var(--tracking-notion-sm)] ${appointment.isDesignated ? `${C.bgDiscountLight} ${C.textDiscount} ${C.borderDiscount20}` : `bg-white ${C.text60}`}`}>
              <Stethoscope className={`${ICON.xs} shrink-0`} />
              <span className="truncate max-w-[80px]">{appointment.doctor ?? "担当医未設定"}</span>
              {appointment.isDesignated ? <span className="text-[10px] ml-0.5 font-bold tracking-[0.12em]">指</span> : null}
            </Badge>
          </div>

          {/* ミニアクションボタン */}
          <div className={`flex items-center gap-1 pt-0.5 border-t ${C.borderDivider}`}>
            <button
              type="button"
              aria-label={isTrimming ? `${appointment.petName}のトリミング記録` : `${appointment.petName}のカルテ`}
              className={`flex items-center gap-1 text-[11px] tracking-[var(--tracking-notion-xs)] ${C.accent} ${C.bgAccentLight30} border ${C.borderAccentBadge} rounded px-1.5 py-0.5 ${C.hoverBgAccentLight60} transition-colors`}
              onClick={handleKarteClick}
            >
              {isTrimming ? <Scissors className={`${ICON.xs} shrink-0`} /> : <FileText className={`${ICON.xs} shrink-0`} />}
              <span>{isTrimming ? "施術" : "カルテ"}</span>
            </button>
            {columnTitle !== "診療中" ? (
              <button
                type="button"
                aria-label={`${appointment.petName}の会計`}
                className={`flex items-center gap-1 text-[11px] tracking-[var(--tracking-notion-xs)] ${C.textStatusGreen} ${C.bgStatusGreen30} border ${C.borderStatusGreen} rounded px-1.5 py-0.5 ${C.hoverBgStatusGreenLight60} transition-colors`}
                onClick={handleAccountingClick}
              >
                <CreditCard className={`${ICON.xs} shrink-0`} />
                <span>会計</span>
              </button>
            ) : null}
            {columnTitle !== "診療中" && isHospitalization ? (
              <button
                type="button"
                aria-label={`${appointment.petName}の入院登録`}
                className={`flex items-center gap-1 text-[11px] tracking-[var(--tracking-notion-xs)] ${C.textStatusPurple} ${C.bgStatusPurple30} border ${C.borderStatusPurple} rounded px-1.5 py-0.5 ${C.hoverBgStatusPurpleLight60} transition-colors`}
                onClick={handleHospitalizationClick}
              >
                <BedDouble className={`${ICON.xs} shrink-0`} />
                <span>入院</span>
              </button>
            ) : null}
          </div>
        </CardContent>
      </Card>
    </div>
  );
});
