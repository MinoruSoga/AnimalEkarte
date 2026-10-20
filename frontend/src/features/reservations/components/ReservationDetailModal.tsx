import { memo } from "react";
import { BADGE, ICON, C, STYLE } from "@/lib/design-tokens";
import type { ReactNode } from "react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import {
  Calendar,
  Clock,
  Stethoscope,
  FileText,
  Pencil,
  Scissors,
  Building2,
  FilePlus2,
  PawPrint,
  Tag,
  AlertTriangle,
  CalendarCheck,
} from "lucide-react";
import { useGetOwnerLineTags } from "@/hooks/use-owner-line-tags";
import { useGetReservation } from "@/hooks/use-get-reservation";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { BirthdayBadge } from "@/components/shared/BirthdayBadge";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import type { Reservation, ReservationStatus } from "../types";
import { RESERVATION_STATUS_VALUES } from "../types";
import { getReservationTypeName, getReservationStatusLabel } from "@/lib/status-helpers";
import { DISPLAY_TIME_FORMAT } from "@/lib/format/date";
import { typedSetter } from "@/lib/type-utils";
import { useReservationTypeColorMap } from "../hooks/use-reservation-type-color-map";
import {
  RESERVATION_STATUS_COLORS,
  getReservationStatusColor,
  getVisitTypeColor,
} from "@/constants/status-colors";
import {
  EXPIRED_TENTATIVE_LABEL,
  isExpiredTentativeReservation,
  isTentativeReservation,
  TENTATIVE_APPOINTMENT_LABEL,
} from "../lib/tentative";
import { RevisitIntakeSection } from "./RevisitIntakeSection";

interface ReservationDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  onEdit?: (reservation: Reservation) => void;
  onDelete?: (reservation: Reservation) => void;
  onCreateRecord?: (reservation: Reservation) => void;
  onStatusChange?: (reservation: Reservation, status: ReservationStatus) => void;
  /**
   * EMR-308: tentative（次回予定・未予約）行を「予約を確定」モーダルで開く。
   * 省略時は tentative 行に確定ボタンを出さない（fail-closed）。
   */
  onConfirmTentative?: (reservation: Reservation) => void;
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
const RESERVATION_STATUS_SELECT_ITEMS = (
  Object.entries(RESERVATION_STATUS_COLORS) as [
    ReservationStatus,
    (typeof RESERVATION_STATUS_COLORS)[ReservationStatus],
  ][]
).map(([value, colors]) => (
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
  onConfirmTentative,
  reservation,
}: ReservationDetailModalProps) {
  const { getColor } = useReservationTypeColorMap();
  // EMR-227/EMR-301: 共有 transform (src/lib/transforms/reservation.ts) が petStatus /
  // petDeceasedAt を additive 伝播する。wire ペットサマリの死亡OR契約
  // （status=deceased または deceased_at 非null）で判定する。
  // レビュー指摘: reservation prop はカードクリック時の snapshot で、死亡登録等の
  // mutation 後も古い sentinel を保持する。GET /reservations/:id（modal 表示に
  // 必要な reservations:view と同一権限）の最新応答を優先して snapshot を追従させる。
  // owners:view 依存の /pets/:id フェッチは再導入しない。
  const { data: freshReservation } = useGetReservation(
    reservation?.id ?? "",
    reservation?.clinicId,
  );
  const effective = freshReservation ?? reservation;
  // PR#614 再レビュー指摘2: LINE タグも表示・操作対象と同一 owner を見るため、
  // snapshot ではなく effective 予約の ownerId で引く。rules-of-hooks により
  // 早期 return より前で導出し、未確定時は ""（hook 側 enabled=false）。
  const { data: lineData } = useGetOwnerLineTags(effective?.ownerId ?? "", effective?.clinicId);

  if (!effective) return null;

  const isDeceasedPet = effective.petStatus === "deceased" || effective.petDeceasedAt != null;

  // EMR-308: tentative（次回予定・未予約）は実時間を持たない。status セレクトや通常編集ではなく
  // 「予約を確定」（confirm-tentative API）で昇格させるため専用 UI に切り替える。
  const isTentative = isTentativeReservation(effective);
  const isExpiredTentative = isExpiredTentativeReservation(effective);
  const actionConfig = getActionConfig(effective);
  const visitAccent = getVisitTypeColor(effective.visitType);
  const currentStatus = getReservationStatusColor(effective.status);
  const typeColor = getColor(effective.type);
  const typeLabel = typeColor.isInactive
    ? `${getReservationTypeName(effective.type)}（無効）`
    : getReservationTypeName(effective.type);

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent
        className={`sm:max-w-[420px] max-h-[85vh] overflow-y-auto p-0 gap-0 overflow-x-hidden ${C.bgWhite} rounded-xl`}
      >
        {/* Accent Header */}
        <div className={`h-1.5 w-full ${visitAccent.dot}`} />

        <DialogHeader className="px-6 pt-4 pb-0 pr-12">
          <div className="flex items-center gap-2.5">
            <div
              className={`flex items-center gap-1.5 px-2 py-0.5 rounded-full text-sm ${visitAccent.bg} ${visitAccent.text} ${visitAccent.border} border`}
            >
              <span className={`${ICON.dotSm} rounded-full ${visitAccent.dot}`} />
              {effective.visitType === "first" ? "初診" : "再診"}
            </div>
            <DialogTitle className={`text-sm ${C.text}`}>{typeLabel}</DialogTitle>
            {/* EMR-308: tentative は色だけでなく明示テキストで識別する */}
            {isTentative ? (
              <span
                className={`flex items-center gap-1 px-2 py-0.5 rounded-full text-2xs font-semibold ${C.bgWarning50} ${C.textWarning} ${C.borderWarning20} border`}
              >
                {isExpiredTentative ? EXPIRED_TENTATIVE_LABEL : TENTATIVE_APPOINTMENT_LABEL}
              </span>
            ) : null}
          </div>
          <DialogDescription className="sr-only">予約の詳細情報</DialogDescription>
        </DialogHeader>

        <div className="px-6 pt-3 pb-4 space-y-4">
          {/* LINE Warning */}
          {lineData !== undefined && (!lineData.is_linked || lineData.lstep_opt_out) ? (
            <div
              className={`flex items-start gap-2 p-3 rounded-md ${C.bgWarning50} border ${C.borderWarning20} text-sm ${C.textWarning}`}
            >
              <AlertTriangle className={`shrink-0 mt-0.5 ${ICON.sm}`} />
              <span>
                {lineData.lstep_opt_out
                  ? "この飼主はLINE配信停止中です。Lステップ同期対象外になります。"
                  : "この飼主はLINE未連携です。Lステップ同期対象外になります。"}
              </span>
            </div>
          ) : null}

          {/* Status Selector — tentative 行は通常 PATCH で状態を変えず confirm-tentative で昇格する */}
          {onStatusChange && !isTentative ? (
            <div
              className={`flex items-center justify-between rounded-lg border px-3 py-2 ${currentStatus.bg} ${currentStatus.text} border-transparent`}
            >
              <div className="flex items-center gap-2 text-sm">
                {/* BUG-323: Status Dot Icon Token 使用統一 */}
                <span className={`${ICON.dot} rounded-full ${currentStatus.dot}`} />
                <span>{getReservationStatusLabel(effective.status)}</span>
              </div>
              <Select
                value={effective.status}
                onValueChange={typedSetter(
                  (val: ReservationStatus) => onStatusChange(effective, val),
                  [...RESERVATION_STATUS_VALUES],
                )}
              >
                <SelectTrigger
                  className={`h-11 min-w-11 w-auto gap-1 border-0 ${C.bgWhite60} ${C.hoverBgWhite80} text-sm px-2 shadow-none`}
                >
                  <SelectValue placeholder="変更" />
                </SelectTrigger>
                <SelectContent>{RESERVATION_STATUS_SELECT_ITEMS}</SelectContent>
              </Select>
            </div>
          ) : null}

          {/* Date & Time Card */}
          <div className={`rounded-lg border ${C.borderLight} ${C.bgSubtle} p-3`}>
            <div className="flex items-center gap-3">
              <div
                className={`flex items-center justify-center w-9 h-9 rounded-lg ${C.bgPrimary5}`}
              >
                <Calendar className={`${ICON.action} ${C.text60}`} />
              </div>
              <div className="flex-1">
                <div className={`text-sm ${C.text}`}>
                  {format(effective.start, "yyyy年 M月 d日 (E)", { locale: ja })}
                </div>
                <div className={`flex items-center gap-1.5 text-sm ${C.text60} mt-0.5`}>
                  <Clock className={ICON.xs} />
                  {/* tentative の start/end は意図日 00:00–00:15 の sentinel。実時刻ではないため明示表示 */}
                  {isTentative
                    ? "時刻未設定（予定日のみ）"
                    : `${format(effective.start, DISPLAY_TIME_FORMAT)} – ${format(effective.end, DISPLAY_TIME_FORMAT)}`}
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
                <span className="font-medium inline-flex items-center gap-1.5">
                  {effective.petName}
                  {/* EMR-301: 誕生月 Cake バッジ（死亡ペット抑制） */}
                  {!isDeceasedPet ? (
                    <BirthdayBadge
                      birthDate={effective.petBirthDate}
                      subjectName={effective.petName}
                    />
                  ) : null}
                  {isDeceasedPet ? (
                    <span
                      className={`text-2xs font-semibold px-1.5 py-0.5 rounded ${C.bgDanger} ${C.textWhite} uppercase`}
                    >
                      【死亡】
                    </span>
                  ) : null}
                  {effective.petType ? (
                    <span className={`ml-1.5 text-xs ${C.text50}`}>({effective.petType})</span>
                  ) : null}
                </span>
              </InfoRow>
              <InfoRow label="飼主名">{effective.ownerName}</InfoRow>
              {effective.petId ? (
                <InfoRow label="カルテNo.">
                  <span className={`font-mono ${C.text70}`}>{effective.petId}</span>
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
                  {effective.doctor}
                  {effective.isDesignated ? (
                    <Badge variant="outline" className={`text-2xs h-5 px-1.5 ${BADGE.yellow}`}>
                      指名
                    </Badge>
                  ) : null}
                </div>
              </InfoRow>
              <InfoRow label="予約区分">
                <div className="flex items-center gap-1.5">
                  {/* BUG-323: Status Dot Icon Token 使用統一 */}
                  <span
                    className={`${ICON.dot} rounded-full shrink-0`}
                    style={typeColor.dotStyle}
                  />
                  <Tag className={`${ICON.xs} ${C.text40}`} />
                  {typeLabel}
                </div>
              </InfoRow>
            </div>
          </div>

          {/* EMR-475: 飼主由来の事前問診原文（read-only）。予約切替・閉鎖で unmount され query を破棄する。
              スタッフ補足は下の「予約内容」または Inquiry へ（原文とは別欄） */}
          <RevisitIntakeSection
            key={`${effective.id}|${effective.clinicId ?? ""}`}
            appointmentId={effective.id}
            clinicId={effective.clinicId}
            enabled={isOpen}
          />

          {/* Notes */}
          {effective.notes ? (
            <div className={`rounded-lg border ${C.borderLight} ${C.bgPage} p-3`}>
              <div className={`flex items-center gap-1.5 text-sm ${C.text60} mb-1.5`}>
                <FileText className={`${ICON.xs}`} />
                <span>予約内容</span>
              </div>
              <p className={`text-sm ${C.text80} whitespace-pre-wrap leading-relaxed`}>
                {effective.notes}
              </p>
            </div>
          ) : null}
        </div>

        <DialogFooter
          className={`px-6 py-3 ${C.bgSubtle} flex flex-row items-center border-t ${C.borderDivider}`}
        >
          <div className="flex-1">
            {/* tentative 行の終了は DELETE next-visit-intent（期限超過パネル「終了」）経路のみ。
                通常 DELETE で sentinel 行を消す導線は出さない */}
            {onDelete && !isTentative ? (
              <DeleteIconButton onClick={() => onDelete(effective)} />
            ) : null}
          </div>
          <div className="flex gap-2">
            {isTentative ? (
              // EMR-308: 確定は実予約化なので死亡ペットは出さない
              // （EMR-177 positive-match ゲートと同一契約 — pet 未取得時は BE に委譲）。
              onConfirmTentative && !isDeceasedPet ? (
                <Button
                  size="sm"
                  onClick={() => onConfirmTentative(effective)}
                  className={`h-9 text-sm gap-1.5 ${C.bgActionPrimarySolid} ${C.textOnActionPrimary} ${C.hoverBgActionPrimarySolid}`}
                >
                  <CalendarCheck className={ICON.xs} />
                  予約を確定
                </Button>
              ) : null
            ) : onEdit ? (
              <Button
                variant="outline"
                size="sm"
                onClick={() => onEdit(effective)}
                className={`h-9 text-sm gap-1.5 ${C.borderMedium} ${C.bgWhite} ${C.text} ${C.hoverBgPage}`}
              >
                <Pencil className={`${ICON.xs}`} />
                編集
              </Button>
            ) : null}
            {/* 死亡ペットは新規カルテ作成不可 (EMR-177 positive-match ゲート)。
                EMR-308: tentative 行も出さない — appointmentId 付き遷移は sentinel 行に
                直接紐づき昇格しない。当日 tentative の昇格は「予約を確定」または
                カルテ新規作成（record_shortcut → BE 同一行昇格）経路で行う。 */}
            {onCreateRecord && !isDeceasedPet && !isTentative ? (
              <Button
                size="sm"
                className={`${C.bgActionPrimarySolid} ${C.textOnActionPrimary} ${C.hoverBgActionPrimarySolid} ${C.hoverTextOnActionPrimary} ${C.activeBgActionPrimarySolid} ${C.activeTextOnActionPrimary} h-9 text-sm gap-1.5 rounded-full shadow-none`}
                onClick={() => onCreateRecord(effective)}
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
