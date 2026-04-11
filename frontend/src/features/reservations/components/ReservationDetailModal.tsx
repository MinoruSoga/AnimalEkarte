import { memo } from "react";
import { ICON, C } from "@/lib/design-tokens";
import type { ReactNode } from "react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { Calendar, Clock, Stethoscope, FileText, Pencil, Scissors, Building2, FilePlus2, PawPrint, Tag } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import type { Appointment, ReservationStatus } from "../types";
import { RESERVATION_STATUS_VALUES } from "../types";
import { getReservationTypeName, getReservationStatusLabel } from "@/utils/status-helpers";
import { typedSetter } from "@/lib/type-utils";
import { useReservationTypeColorMap } from "@/features/master";
import { RESERVATION_STATUS_COLORS, getReservationStatusColor, getVisitTypeColor } from "@/utils/constants/status-colors";

interface ReservationDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  onEdit?: (appointment: Appointment) => void;
  onDelete?: (appointment: Appointment) => void;
  onCreateRecord?: (appointment: Appointment) => void;
  onStatusChange?: (appointment: Appointment, status: ReservationStatus) => void;
  appointment: Appointment | null;
}

// STATUS_OPTIONS は RESERVATION_STATUS_COLORS に集約済み（status-colors.ts）

interface ActionConfig {
  label: string;
  Icon: React.ComponentType<{ className?: string }>;
}

const ACTION_CONFIG_MAP: Record<string, ActionConfig> = {
  "トリミング": { label: "トリミング記録作成", Icon: Scissors },
  "入院": { label: "入院・ホテル登録", Icon: Building2 },
  "ホテル": { label: "入院・ホテル登録", Icon: Building2 },
};

const DEFAULT_ACTION_CONFIG: ActionConfig = {
  label: "カルテ作成",
  Icon: FilePlus2,
};

// rendering-hoist-jsx: 静的な定数から生成する JSX はモジュールレベルで巻き上げ
const RESERVATION_STATUS_SELECT_ITEMS = (Object.entries(RESERVATION_STATUS_COLORS) as [ReservationStatus, typeof RESERVATION_STATUS_COLORS[ReservationStatus]][]).map(([value, colors]) => (
  <SelectItem key={value} value={value} className="text-sm">
    <div className="flex items-center gap-2">
      <span className={`w-2 h-2 rounded-full ${colors.dot}`} />
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
  appointment,
}: ReservationDetailModalProps) {
  const { getColor } = useReservationTypeColorMap();

  if (!appointment) return null;

  const actionConfig = ACTION_CONFIG_MAP[appointment.type] ?? DEFAULT_ACTION_CONFIG;
  const visitAccent = getVisitTypeColor(appointment.visitType);
  const currentStatus = getReservationStatusColor(appointment.status);

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[420px] p-0 gap-0 overflow-hidden bg-white rounded-xl">
        {/* Accent Header */}
        <div className={`h-1.5 w-full ${visitAccent.dot}`} />

        <DialogHeader className="px-5 pt-4 pb-0 pr-12">
          <div className="flex items-center gap-2.5">
            <div className={`flex items-center gap-1.5 px-2 py-0.5 rounded-full text-sm ${visitAccent.bg} ${visitAccent.text} ${visitAccent.border} border`}>
              <span className={`w-1.5 h-1.5 rounded-full ${visitAccent.dot}`} />
              {appointment.visitType === "first" ? "初診" : "再診"}
            </div>
            <DialogTitle className={`text-sm ${C.text}`}>
              {getReservationTypeName(appointment.type)}
            </DialogTitle>
          </div>
          <DialogDescription className="sr-only">
            予約の詳細情報
          </DialogDescription>
        </DialogHeader>

        <div className="px-5 pt-3 pb-4 space-y-4">
          {/* Status Selector */}
          {onStatusChange ? (
            <div className={`flex items-center justify-between rounded-lg border px-3 py-2 ${currentStatus.bg} ${currentStatus.text} border-transparent`}>
              <div className="flex items-center gap-2 text-sm">
                <span className={`w-2 h-2 rounded-full ${currentStatus.dot}`} />
                <span>{getReservationStatusLabel(appointment.status)}</span>
              </div>
              <Select
                value={appointment.status}
                onValueChange={typedSetter(
                  (val: ReservationStatus) => onStatusChange(appointment, val),
                  [...RESERVATION_STATUS_VALUES]
                )}
              >
                <SelectTrigger className="h-7 w-auto gap-1 border-0 bg-white/60 hover:bg-white/80 text-sm px-2 shadow-none focus:ring-0">
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
                  {format(appointment.start, "yyyy年 M月 d日 (E)", { locale: ja })}
                </div>
                <div className={`flex items-center gap-1.5 text-sm ${C.text60} mt-0.5`}>
                  <Clock className={ICON.xs} />
                  {format(appointment.start, "H:mm")} – {format(appointment.end, "H:mm")}
                </div>
              </div>
            </div>
          </div>

          {/* Patient Info */}
          <div className="space-y-1">
            <div className="flex items-center gap-1.5 mb-2">
              <PawPrint className={`${ICON.xs} ${C.text40}`} />
              <span className={`text-sm ${C.text50} tracking-wide`}>患者情報</span>
            </div>
            <div className={`divide-y ${C.divideDividerFaint}`}>
              <InfoRow label="ペット名">
                <span className="font-medium">{appointment.petName}</span>
              </InfoRow>
              <InfoRow label="飼い主名">
                {appointment.ownerName}
              </InfoRow>
              {appointment.petId ? (
                <InfoRow label="カルテNo.">
                  <span className={`font-mono ${C.text70}`}>{appointment.petId}</span>
                </InfoRow>
              ) : null}
            </div>
          </div>

          {/* Medical Details */}
          <div className="space-y-1">
            <div className="flex items-center gap-1.5 mb-2">
              <Stethoscope className={`${ICON.xs} ${C.text40}`} />
              <span className={`text-sm ${C.text50} tracking-wide`}>診療詳細</span>
            </div>
            <div className={`divide-y ${C.divideDividerFaint}`}>
              <InfoRow label="担当医">
                <div className="flex items-center gap-1.5">
                  {appointment.doctor}
                  {appointment.isDesignated ? (
                    <Badge variant="outline" className={`text-[11px] h-5 px-1.5 ${C.bgNotice} ${C.textNotice} ${C.borderNotice}`}>
                      指名
                    </Badge>
                  ) : null}
                </div>
              </InfoRow>
              <InfoRow label="予約区分">
                <div className="flex items-center gap-1.5">
                  <span className="w-2 h-2 rounded-full shrink-0" style={appointment ? getColor(appointment.type).dotStyle : undefined} />
                  <Tag className={`${ICON.xs} ${C.text40}`} />
                  {getReservationTypeName(appointment.type)}
                </div>
              </InfoRow>
            </div>
          </div>

          {/* Notes */}
          {appointment.notes ? (
            <div className={`rounded-lg border ${C.borderNotice50} ${C.bgNotice40} p-3`}>
              <div className={`flex items-center gap-1.5 text-sm ${C.textNotice} mb-1.5`}>
                <FileText className={`${ICON.xs}`} />
                <span>メモ</span>
              </div>
              <p className={`text-sm ${C.text80} whitespace-pre-wrap leading-relaxed`}>{appointment.notes}</p>
            </div>
          ) : null}
        </div>

        <DialogFooter className={`px-5 py-3 ${C.bgSubtle} flex flex-row items-center border-t ${C.borderDivider}`}>
          <div className="flex-1">
            {onDelete ? (
              <DeleteIconButton onClick={() => onDelete(appointment)} />
            ) : null}
          </div>
          <div className="flex gap-2">
            {onEdit ? (
              <Button
                variant="outline"
                size="sm"
                onClick={() => onEdit(appointment)}
                className={`h-9 text-sm gap-1.5 ${C.borderMedium} bg-white ${C.text} ${C.hoverBgPage}`}
              >
                <Pencil className={`${ICON.xs}`} />
                編集
              </Button>
            ) : null}
            {onCreateRecord ? (
              <Button
                size="sm"
                className={`${C.bgAccent} ${C.textWhite} ${C.bgAccentHover} h-9 text-sm gap-1.5 shadow-sm`}
                onClick={() => onCreateRecord(appointment)}
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
