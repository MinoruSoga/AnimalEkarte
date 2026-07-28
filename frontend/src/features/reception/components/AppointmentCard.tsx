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
import Syringe from "lucide-react/dist/esm/icons/syringe";
import Activity from "lucide-react/dist/esm/icons/activity";
import FileText from "lucide-react/dist/esm/icons/file-text";
import CreditCard from "lucide-react/dist/esm/icons/credit-card";
import BedDouble from "lucide-react/dist/esm/icons/bed-double";

// Internal
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { C, ICON } from "@/lib/design-tokens";
import { getVisitTypeColor } from "@/constants/status-colors";
import {
  DangerLevelHigh,
  PetStatusDeceased,
} from "@/types/generated/models";

// Types
import type { ReceptionAppointment } from "../api/types";

interface ServiceIconProps {
  service: string;
  category?: string;
}

function isHospitalizationService(service: string): boolean {
  return service.includes("入院") || service.includes("ホテル");
}

function ServiceIcon({ service, category }: ServiceIconProps) {
  if (category === "trimming" || service.includes("トリミング")) return <Scissors className={ICON.xs} />;
  if (isHospitalizationService(service)) return <BedDouble className={ICON.xs} />;
  if (service.includes("ワクチン")) return <Syringe className={ICON.xs} />;
  if (service.includes("手術")) return <Activity className={ICON.xs} />;
  if (service.includes("診療")) return <Stethoscope className={ICON.xs} />;
  return <Stethoscope className={ICON.xs} />;
}

interface AppointmentCardProps {
  appointment: ReceptionAppointment;
  columnTitle: string;
  onCardClick: (appointment: ReceptionAppointment) => void;
  onRecordOpen?: (appointment: ReceptionAppointment, columnTitle: string) => void;
  isDragOverlay?: boolean;
}

export const AppointmentCard = memo(function AppointmentCard({
  appointment,
  columnTitle,
  onCardClick,
  onRecordOpen,
  isDragOverlay = false,
}: AppointmentCardProps) {
  const navigate = useNavigate();
  const isDeceased = appointment.petStatus === PetStatusDeceased;

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
    disabled: isDragOverlay || isDeceased,
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const isTrimming = appointment.reservationCategory === "trimming";
  const isHospitalization = isHospitalizationService(appointment.reservationType);
  const isMedical = !isTrimming && !isHospitalization;
  const isHighDanger = appointment.petDangerLevel === DangerLevelHigh;
  const visitColor = getVisitTypeColor(appointment.visitType);
  const canOpenRecordFromCard = isTrimming
    ? columnTitle === "受付済"
    : isMedical && (columnTitle === "受付済" || columnTitle === "診療中");

  const handleKarteClick = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    onRecordOpen?.(appointment, columnTitle);
    const params = new URLSearchParams();
    if (appointment.petId) params.set("petId", appointment.petId);
    if (appointment.id) params.set("appointmentId", appointment.id);
    if (appointment.visitDate) params.set("visitDate", appointment.visitDate);
    const query = params.toString();
    const basePath = isTrimming
      ? appointment.petId
        ? paths.trimming.new.getHref()
        : paths.trimming.selectPet.getHref()
      : appointment.petId
        ? paths.medicalRecords.new.getHref()
        : paths.medicalRecords.selectPet.getHref();

    navigate(
      query ? `${basePath}?${query}` : basePath,
      { state: { from: "/", appointmentId: appointment.id, visitDate: appointment.visitDate } },
    );
  }, [navigate, isTrimming, appointment, columnTitle, onRecordOpen]);

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
      aria-disabled={isDeceased}
      className="cursor-grab active:cursor-grabbing group touch-none"
      onClick={() => {
        if (!isDeceased) onCardClick(appointment);
      }}
      onKeyDown={(e) => {
        if (e.target !== e.currentTarget) return;
        if (!isDeceased && (e.key === "Enter" || e.key === " ")) {
          e.preventDefault();
          onCardClick(appointment);
        }
      }}
    >
      <Card className={`w-full ${C.hoverBgPage} transition-colors border ${C.borderLight} rounded-md`}>
        <CardContent className="p-[13px] space-y-[9px]">
          <div className="flex items-start justify-between gap-2">
            <div className={`flex items-center gap-1.5 ${C.text60} min-w-0`}>
              <Clock className={`${ICON.xs} flex-shrink-0`} />
              <span className="text-base font-medium font-mono">{appointment.time}</span>
            </div>
          </div>

          <div className="space-y-0.5">
            <p className="text-base font-semibold truncate leading-tight">{appointment.ownerName}</p>
            <div className={`flex items-center gap-1 ${C.text60}`}>
              <Dog className={`${ICON.xs} flex-shrink-0`} />
              <p className="text-base truncate">{appointment.petType} - {appointment.petName}</p>
              {isDeceased ? (
                <span className={`text-2xs font-semibold px-1.5 py-0.5 rounded ${C.bgDanger} ${C.textWhite} uppercase ml-1`}>
                  【死亡】
                </span>
              ) : null}
              {isHighDanger ? (
                <Popover>
                  <PopoverTrigger asChild>
                    <button
                      type="button"
                      aria-label={`${appointment.petName}の危険理由を表示`}
                      className={`inline-flex items-center rounded px-1.5 py-0.5 text-xs font-semibold ${C.bgDanger10} ${C.danger} ${C.borderDanger20} outline-none focus-visible:ring-2 ${C.focusRingAccent40}`}
                      onPointerDown={(event) => event.stopPropagation()}
                      onClick={(event) => event.stopPropagation()}
                    >
                      ⚠ 危険
                    </button>
                  </PopoverTrigger>
                  <PopoverContent
                    align="start"
                    aria-label={`${appointment.petName}の危険理由`}
                    onOpenAutoFocus={(event) => event.preventDefault()}
                    onPointerDown={(event) => event.stopPropagation()}
                    onClick={(event) => event.stopPropagation()}
                    className="w-64"
                  >
                    <p className={`text-sm font-semibold ${C.danger}`}>危険理由</p>
                    <p className={`mt-1 whitespace-pre-wrap break-words text-sm ${C.textInkSecondary}`}>
                      {appointment.petDangerReason?.trim() || "理由未登録"}
                    </p>
                  </PopoverContent>
                </Popover>
              ) : null}
            </div>
          </div>

          <div className="flex items-center flex-wrap gap-1 pt-0.5">
            <Badge
              variant="secondary"
              className={`text-sm px-[7.5px] h-[22px] ${visitColor.badgeBg} ${visitColor.badgeText} ${visitColor.badgeBorder}`}
            >
              {appointment.visitType}
            </Badge>
            <Badge variant="outline" className={`flex items-center gap-1 text-sm px-[7.5px] h-[22px] ${C.bgWhite}`}>
              <ServiceIcon service={appointment.reservationType} category={appointment.reservationCategory} />
              <span className="truncate max-w-[80px]">{appointment.reservationType}</span>
            </Badge>

            {/* BUG-037: 担当医バッジ — doctor が未設定でも「担当医未設定」として表示 */}
            <Badge variant="outline" className={`flex items-center gap-1 text-sm px-[7.5px] h-[22px] ${appointment.isDesignated ? `${C.bgDiscountLight} ${C.textDiscount} ${C.borderDiscount20}` : `${C.bgWhite} ${C.text60}`}`}>
              <Stethoscope className={`${ICON.xs} shrink-0`} />
              <span className="truncate max-w-[80px]">{appointment.doctor ?? "担当医未設定"}</span>
              {appointment.isDesignated ? <span className="text-2xs ml-0.5 font-semibold">指</span> : null}
            </Badge>
          </div>

          {/* ミニアクションボタン */}
          {isDeceased ? null : (
            <div className={`flex flex-wrap items-center gap-1 pt-0.5 border-t ${C.borderDivider}`}>
              {canOpenRecordFromCard ? (
                <button
                  type="button"
                  aria-label={isTrimming ? `${appointment.petName}のトリミング記録` : `${appointment.petName}のカルテ`}
                  className={`flex min-h-11 min-w-11 items-center justify-center gap-1 text-2xs ${C.textBrand} ${C.bgBrandLight30} border ${C.borderBrandLight} rounded px-1.5 ${C.hoverBgBrandLight60} transition-colors`}
                  onClick={handleKarteClick}
                >
                  {isTrimming ? <Scissors className={`${ICON.xs} shrink-0`} /> : <FileText className={`${ICON.xs} shrink-0`} />}
                  <span>{isTrimming ? "施術" : "カルテ"}</span>
                </button>
              ) : null}
              {columnTitle !== "診療中" ? (
                <button
                  type="button"
                  aria-label={`${appointment.petName}の会計`}
                  className={`flex min-h-11 min-w-11 items-center justify-center gap-1 text-2xs ${C.textStatusGreen} ${C.bgStatusGreen30} border ${C.borderStatusGreen} rounded px-1.5 ${C.hoverBgStatusGreenLight60} transition-colors`}
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
                  className={`flex min-h-11 min-w-11 items-center justify-center gap-1 text-2xs ${C.textStatusPurple} ${C.bgStatusPurple30} border ${C.borderStatusPurple} rounded px-1.5 ${C.hoverBgStatusPurpleLight60} transition-colors`}
                  onClick={handleHospitalizationClick}
                >
                  <BedDouble className={`${ICON.xs} shrink-0`} />
                  <span>入院</span>
                </button>
              ) : null}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
});
