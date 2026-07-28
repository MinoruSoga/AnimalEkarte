import { memo } from "react";
import { ICON, C, STYLE } from "@/lib/design-tokens";
import type { ReactNode } from "react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { Calendar, Clock, Stethoscope, FileText, Pencil, Scissors, Building2, FilePlus2, PawPrint, Tag, AlertTriangle } from "lucide-react";
import { useGetOwnerLineTags } from "@/hooks/use-owner-line-tags";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import type { Reservation, ReservationStatus } from "../types";
import { RESERVATION_STATUS_VALUES } from "../types";
import { getReservationTypeName, getReservationStatusLabel } from "@/lib/status-helpers";
import { DISPLAY_TIME_FORMAT } from "@/lib/format/date";
import { typedSetter } from "@/lib/type-utils";
import { useReservationTypeColorMap } from "@/hooks/use-reservation-type-color-map";
import { RESERVATION_STATUS_COLORS, getReservationStatusColor, getVisitTypeColor } from "@/constants/status-colors";

interface ReservationDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  onEdit?: (reservation: Reservation) => void;
  onDelete?: (reservation: Reservation) => void;
  onCreateRecord?: (reservation: Reservation) => void;
  onStatusChange?: (reservation: Reservation, status: ReservationStatus) => void;
  reservation: Reservation | null;
}

// STATUS_OPTIONS は RESERVATION_STATUS_COLORS に集約済み（status-colors.ts）

interface ActionConfig {
  label: string;
  Icon: React.ComponentType<{ className?: string }>;
}

const DEFAULT_ACTION_CONFIG: ActionConfig = {
  label: "カルテ作成",
  Icon: FilePlus2,
};

function isHospitalizationReservation(type: string): boolean {
  return type.includes("入院") || type.includes("ホテル");
}

function getActionConfig(reservation: Reservation): ActionConfig {
  if (reservation.category === "trimming") {
    return { label: "トリミング記録作成", Icon: Scissors };
  }
  if (isHospitalizationReservation(reservation.type)) {
    return { label: "入院・ホテル登録", Icon: Building2 };
  }
  return DEFAULT_ACTION_CONFIG;
}

// rendering-hoist-jsx: 静的な定数から生成する JSX はモジュールレベルで巻き上げ
const RESERVATION_STATUS_SELECT_ITEMS = (Object.entries(RESERVATION_STATUS_COLORS) as [ReservationStatus, typeof RESERVATION_STATUS_COLORS[ReservationStatus]][]).map(([value, colors]) => (
  <SelectItem key={value} value={value} className="text-sm">
    <div className="flex items-center gap-2">
      {/* BUG-323: Status Dot Icon Token 使用統一 */}
      <span className={`${ICON.dot} rounded-full ${colors.dot}`} />
      {colors.label}
    </div>
  </SelectItem>
));

// getVisitTypeAccent は getVisitTypeColor に集約済み（status-colors.ts）

function InfoRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex items-center justify-between py-2">
      <span className={`text-sm ${C.text50}`}>{label}</span>
      <span className={`text-sm ${C.text}`}>{children}</span>
    </div>
  );
}

export const ReservationDetailModal = memo(function ReservationDetailModal({
  isOpen,
  onClose,
  onEdit,
  onDelete,
  onCreateRecord,
  onStatusChange,
  reservation,
}: ReservationDetailModalProps) {
  const { getColor } = useReservationTypeColorMap();
  const { data: lineData } = useGetOwnerLineTags(reservation?.ownerId ?? "");

  if (!reservation) return null;

  const actionConfig = getActionConfig(reservation);
  const visitAccent = getVisitTypeColor(reservation.visitType);
  const currentStatus = getReservationStatusColor(reservation.status);

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className={`sm:max-w-[420px] p-0 gap-0 overflow-hidden ${C.bgWhite} rounded-xl`}>
        {/* Accent Header */}
        <div className={`h-1.5 w-full ${visitAccent.dot}`} />

        <DialogHeader className="px-6 pt-4 pb-0 pr-12">
          <div className="flex items-center gap-2.5">
            <div className={`flex items-center gap-1.5 px-2 py-0.5 rounded-full text-sm ${visitAccent.bg} ${visitAccent.text} ${visitAccent.border} border`}>
              <span className={`${ICON.dotSm} rounded-full ${visitAccent.dot}`} />
              {reservation.visitType === "first" ? "初診" : "再診"}
            </div>
            <DialogTitle className={`text-sm ${C.text}`}>
              {getReservationTypeName(reservation.type)}
            </DialogTitle>
          </div>
          <DialogDescription className="sr-only">
            予約の詳細情報
          </DialogDescription>
        </DialogHeader>

        <div className="px-6 pt-3 pb-4 space-y-4">
          {/* LINE Warning */}
          {lineData !== undefined && (!lineData.is_linked || lineData.lstep_opt_out) ? (
            <div className={`flex items-start gap-2 p-3 rounded-md ${C.bgWarning50} border ${C.borderWarning20} text-sm ${C.textWarning}`}>
              <AlertTriangle className="shrink-0 mt-0.5 w-4 h-4" />
              <span>
                {lineData.lstep_opt_out
                  ? "この飼い主はLINE配信停止中です。Lステップ同期対象外になります。"
                  : "この飼い主はLINE未連携です。Lステップ同期対象外になります。"}
              </span>
            </div>
          ) : null}

          {/* Status Selector */}
          {onStatusChange ? (
            <div className={`flex items-center justify-between rounded-lg border px-3 py-2 ${currentStatus.bg} ${currentStatus.text} border-transparent`}>
              <div className="flex items-center gap-2 text-sm">
                {/* BUG-323: Status Dot Icon Token 使用統一 */}
                <span className={`${ICON.dot} rounded-full ${currentStatus.dot}`} />
                <span>{getReservationStatusLabel(reservation.status)}</span>
              </div>
              <Select
                value={reservation.status}
                onValueChange={typedSetter(
                  (val: ReservationStatus) => onStatusChange(reservation, val),
                  [...RESERVATION_STATUS_VALUES]
                )}
              >
                <SelectTrigger className={`h-7 w-auto gap-1 border-0 ${C.bgWhite}/60 hover:${C.bgWhite}/80 text-sm px-2 shadow-none focus:ring-0`}>
                  <SelectValue placeholder="変更" />
                </SelectTrigger>
                <SelectContent>
                  {RESERVATION_STATUS_SELECT_ITEMS}
                </SelectContent>
              </Select>
            </div>
          ) : null}

          {/* Date & Time Card */}
          <div className={`rounded-lg border ${C.borderLight} ${C.bgSubtle} p-3`}>
            <div className="flex items-center gap-3">
              <div className={`flex items-center justify-center w-9 h-9 rounded-lg ${C.bgPrimary5}`}>
                <Calendar className={`${ICON.action} ${C.text60}`} />
              </div>
              <div className="flex-1">
                <div className={`text-sm ${C.text}`}>
                  {format(reservation.start, "yyyy年 M月 d日 (E)", { locale: ja })}
                </div>
                <div className={`flex items-center gap-1.5 text-sm ${C.text60} mt-0.5`}>
                  <Clock className={ICON.xs} />
                  {format(reservation.start, DISPLAY_TIME_FORMAT)} – {format(reservation.end, DISPLAY_TIME_FORMAT)}
                </div>
              </div>
            </div>
          </div>

          {/* Patient Info */}
          <div className="space-y-1">
            <div className="flex items-center gap-1.5 mb-2">
              <PawPrint className={`${ICON.xs} ${C.text40}`} />
              <span className={STYLE.sectionLabel}>患者情報</span>
            </div>
            <div className={`divide-y ${C.divideDividerFaint}`}>
              <InfoRow label="ペット名">
                <span className="font-medium">
                  {reservation.petName}
                  {reservation.petType ? (
                    <span className={`ml-1.5 text-xs ${C.text50}`}>({reservation.petType})</span>
                  ) : null}
                </span>
              </InfoRow>
              <InfoRow label="飼主名">
                {reservation.ownerName}
              </InfoRow>
              {reservation.petId ? (
                <InfoRow label="カルテNo.">
                  <span className={`font-mono ${C.text70}`}>{reservation.petId}</span>
                </InfoRow>
              ) : null}
            </div>
          </div>

          {/* Medical Details */}
          <div className="space-y-1">
            <div className="flex items-center gap-1.5 mb-2">
              <Stethoscope className={`${ICON.xs} ${C.text40}`} />
              <span className={STYLE.sectionLabel}>診療詳細</span>
            </div>
            <div className={`divide-y ${C.divideDividerFaint}`}>
              <InfoRow label="担当医">
                <div className="flex items-center gap-1.5">
                  {reservation.doctor}
                  {reservation.isDesignated ? (
                    <Badge variant="outline" className={`text-2xs h-5 px-1.5 ${C.bgNotice} ${C.textNotice} ${C.borderNotice}`}>
                      指名
                    </Badge>
                  ) : null}
                </div>
              </InfoRow>
              <InfoRow label="予約区分">
                <div className="flex items-center gap-1.5">
                  {/* BUG-323: Status Dot Icon Token 使用統一 */}
                  <span className={`${ICON.dot} rounded-full shrink-0`} style={reservation ? getColor(reservation.type).dotStyle : undefined} />
                  <Tag className={`${ICON.xs} ${C.text40}`} />
                  {getReservationTypeName(reservation.type)}
                </div>
              </InfoRow>
            </div>
          </div>

          {/* Notes */}
          {reservation.notes ? (
            <div className={`rounded-lg border ${C.borderNotice50} ${C.bgNotice40} p-3`}>
              <div className={`flex items-center gap-1.5 text-sm ${C.textNotice} mb-1.5`}>
                <FileText className={`${ICON.xs}`} />
                <span>メモ</span>
              </div>
              <p className={`text-sm ${C.text80} whitespace-pre-wrap leading-relaxed`}>{reservation.notes}</p>
            </div>
          ) : null}
        </div>

        <DialogFooter className={`px-6 py-3 ${C.bgSubtle} flex flex-row items-center border-t ${C.borderDivider}`}>
          <div className="flex-1">
            {onDelete ? (
              <DeleteIconButton onClick={() => onDelete(reservation)} />
            ) : null}
          </div>
          <div className="flex gap-2">
            {onEdit ? (
              <Button
                variant="outline"
                size="sm"
                onClick={() => onEdit(reservation)}
                className={`h-9 text-sm gap-1.5 ${C.borderMedium} ${C.bgWhite} ${C.text} ${C.hoverBgPage}`}
              >
                <Pencil className={`${ICON.xs}`} />
                編集
              </Button>
            ) : null}
            {onCreateRecord ? (
              <Button
                size="sm"
                className={`${C.bgBrand} ${C.textOnBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} h-9 text-sm gap-1.5 rounded-full shadow-none`}
                onClick={() => onCreateRecord(reservation)}
              >
                <actionConfig.Icon className={ICON.action} />
                {actionConfig.label}
              </Button>
            ) : null}
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
});
