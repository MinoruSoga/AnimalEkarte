// React/Framework
import { memo, useCallback } from "react";
import { useNavigate } from "react-router";

// External
import Clock from "lucide-react/dist/esm/icons/clock";
import User from "lucide-react/dist/esm/icons/user";
import Dog from "lucide-react/dist/esm/icons/dog";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import AlertCircle from "lucide-react/dist/esm/icons/alert-circle";
import FileText from "lucide-react/dist/esm/icons/file-text";
import CreditCard from "lucide-react/dist/esm/icons/credit-card";
import Scissors from "lucide-react/dist/esm/icons/scissors";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";
import Pencil from "lucide-react/dist/esm/icons/pencil";
import TestTube from "lucide-react/dist/esm/icons/test-tube";
import BedDouble from "lucide-react/dist/esm/icons/bed-double";
import ExternalLink from "lucide-react/dist/esm/icons/external-link";

// Internal
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import {
  DASHBOARD_STATUS_COLORS,
  DASHBOARD_STATUS_COLOR_FALLBACK,
} from "@/utils/constants/status-colors";

// Types
import type { Appointment } from "@/types";

// Module-level constants — avoid recreating compound class strings on every render
const SECTION_LABEL = `text-sm font-semibold ${C.text60} uppercase tracking-wider`;
const DIVIDER_ROW = `flex items-center justify-between border-b ${C.borderLight} pb-2`;
const ROW_ICON = `flex items-center gap-2 ${C.text60}`;

const RELATED_BTN_BASE = "flex items-center gap-1.5 text-sm border rounded-md px-3 py-1.5 transition-colors group";
const RELATED_BTN_KARTE      = `${RELATED_BTN_BASE} ${C.accent} bg-[#D3E5EF]/40 hover:bg-[#D3E5EF] ${C.borderAccentLight}`;
const RELATED_BTN_ACCOUNTING = `${RELATED_BTN_BASE} ${C.textStatusGreen} bg-[#DDEDEA]/40 hover:bg-[#DDEDEA] ${C.borderStatusGreen}`;
const RELATED_BTN_HOSPITAL   = `${RELATED_BTN_BASE} ${C.textStatusPurple} bg-[#EEE0F7]/40 hover:bg-[#EEE0F7] ${C.borderStatusPurple}`;


interface RelatedPagesProps {
  isTrimming: boolean;
  onCreateMedicalRecord: (tab?: string) => void;
  onCreateTrimming: () => void;
  onCreateAccounting: () => void;
  onCreateHospitalization: () => void;
}

function RelatedPages({
  isTrimming,
  onCreateMedicalRecord,
  onCreateTrimming,
  onCreateAccounting,
  onCreateHospitalization,
}: RelatedPagesProps) {
  return (
    <div className="space-y-2">
      <h3 className={SECTION_LABEL}>関連ページ</h3>
      <div className="flex flex-wrap gap-2">
        {/* カルテ / 施術 */}
        <button
          type="button"
          className={RELATED_BTN_KARTE}
          onClick={() => {
            if (isTrimming) onCreateTrimming();
            else onCreateMedicalRecord();
          }}
        >
          {isTrimming ? <Scissors className={`${ICON.xs}`} /> : <FileText className={`${ICON.xs}`} />}
          <span>{isTrimming ? "施術" : "カルテ"}</span>
          <ExternalLink className={`${ICON.xs} opacity-0 group-hover:opacity-100 transition-opacity`} />
        </button>

        {/* 会計 */}
        <button
          type="button"
          className={RELATED_BTN_ACCOUNTING}
          onClick={onCreateAccounting}
        >
          <CreditCard className={`${ICON.xs}`} />
          <span>会計</span>
          <ExternalLink className={`${ICON.xs} opacity-0 group-hover:opacity-100 transition-opacity`} />
        </button>

        {/* 入院 */}
        <button
          type="button"
          className={RELATED_BTN_HOSPITAL}
          onClick={onCreateHospitalization}
        >
          <BedDouble className={`${ICON.xs}`} />
          <span>入院</span>
          <ExternalLink className={`${ICON.xs} opacity-0 group-hover:opacity-100 transition-opacity`} />
        </button>
      </div>
    </div>
  );
}

interface ActionButtonsProps {
  currentStatus?: string;
  appointment: Appointment;
  isTrimming: boolean;
  isHospitalization: boolean;
  isMedical: boolean;
  onConfirm?: () => void;
  onEdit?: (appointment: Appointment) => void;
  onCancel?: (appointment: Appointment) => void;
  onOpenOwnerDetail: () => void;
  onCreateMedicalRecord: (tab?: string) => void;
  onCreateTrimming: () => void;
  onCreateAccounting: () => void;
  onCreateHospitalization: () => void;
}

function ActionButtons({
  currentStatus,
  appointment,
  isTrimming,
  isHospitalization,
  isMedical,
  onConfirm,
  onEdit,
  onCancel,
  onOpenOwnerDetail,
  onCreateMedicalRecord,
  onCreateTrimming,
  onCreateAccounting,
  onCreateHospitalization,
}: ActionButtonsProps) {
  const ownerDetailBtn = (
    <Button variant="ghost" onClick={onOpenOwnerDetail} className={`h-10 text-sm ${C.text}`}>
      <User className={ICON.action} />
      飼主詳細
    </Button>
  );

  if (currentStatus === "受付予約") {
    return (
      <>
        {onCancel ? (
          <Button variant="ghost" onClick={() => onCancel(appointment)} className={`h-10 text-sm ${C.danger} ${C.hoverBgDanger5}`}>
            <Trash2 className={ICON.action} />
            取消
          </Button>
        ) : null}
        {onEdit ? (
          <Button variant="outline" onClick={() => onEdit(appointment)} className={`h-10 text-sm ${C.text} ${C.borderMedium}`}>
            <Pencil className={ICON.action} />
            編集
          </Button>
        ) : null}
        {ownerDetailBtn}
        {onConfirm ? (
          <Button onClick={onConfirm} className={STYLE.confirmPrimary}>
            受付済にする
          </Button>
        ) : null}
      </>
    );
  }

  if (currentStatus === "受付済") {
    return (
      <>
        {ownerDetailBtn}
        {isMedical ? (
          <div className="flex flex-col items-end gap-1">
            <Button
              onClick={() => { if (onConfirm) onConfirm(); onCreateMedicalRecord(); }}
              className={STYLE.confirmPrimary}
            >
              <FileText className={ICON.action} />
              カルテ作成
            </Button>
            <span className="text-[10px] text-muted-foreground">※カルテ作成と同時に「診療中」へ移動します</span>
          </div>
        ) : (
          onConfirm ? (
            <Button onClick={onConfirm} className={STYLE.confirmPrimary}>
              診察を開始する
            </Button>
          ) : null
        )}
      </>
    );
  }

  if (currentStatus === "診療中") {
    return (
      <>
        {ownerDetailBtn}
        {onConfirm ? (
          <Button variant="outline" onClick={onConfirm} className={`h-10 text-sm ${C.danger} ${C.borderDanger} ${C.hoverBgDanger5}`}>
            診察を終了する
          </Button>
        ) : null}
        {isMedical ? (
          <>
            <Button onClick={() => onCreateMedicalRecord()} className={STYLE.confirmPrimary}>
              <FileText className={ICON.action} />
              カルテ入力
            </Button>
            <Button variant="outline" onClick={() => onCreateMedicalRecord("test")} className={`h-10 text-sm ${C.text} ${C.borderMedium}`}>
              <TestTube className={ICON.action} />
              検査
            </Button>
          </>
        ) : null}
        {isTrimming ? (
          <Button onClick={onCreateTrimming} className={`h-10 text-sm ${C.bgDiscount} ${C.bgDiscountHover} text-white rounded-[4px] transition-colors shadow-none border-transparent`}>
            <Scissors className={ICON.action} />
            施術記録
          </Button>
        ) : null}
      </>
    );
  }

  if (currentStatus === "会計待ち") {
    return (
      <>
        {ownerDetailBtn}
        <Button
          onClick={onCreateAccounting}
          className={STYLE.confirmPrimary}
        >
          <CreditCard className={ICON.action} />
          会計へ進む
        </Button>
      </>
    );
  }

  if (currentStatus === "会計済") {
    return (
      <>
        {ownerDetailBtn}
        {onConfirm ? (
          <Button onClick={onConfirm} className={STYLE.confirmPrimary}>
            完了/リストから削除
          </Button>
        ) : null}
      </>
    );
  }

  // Default Fallback
  return (
    <>
      {ownerDetailBtn}
      {isMedical ? (
        <Button onClick={() => onCreateMedicalRecord()} className={STYLE.confirmPrimary}>
          <FileText className={ICON.action} />
          カルテ確認
        </Button>
      ) : null}
      {isTrimming ? (
        <Button onClick={onCreateTrimming} className={STYLE.confirmPrimary}>
          <Scissors className={ICON.action} />
          トリミング
        </Button>
      ) : null}
      {isHospitalization ? (
        <Button onClick={onCreateHospitalization} className={STYLE.confirmPrimary}>
          <BedDouble className={ICON.action} />
          入院登録
        </Button>
      ) : null}
      {onConfirm ? (
        <Button onClick={onConfirm} className={STYLE.confirmPrimary}>
          ステータス変更
        </Button>
      ) : null}
    </>
  );
}

interface DashboardDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  appointment: Appointment | null;
  onConfirm?: () => void;
  onEdit?: (appointment: Appointment) => void;
  onCancel?: (appointment: Appointment) => void;
  currentStatus?: string;
}

export const DashboardDetailModal = memo(function DashboardDetailModal({
  isOpen,
  onClose,
  appointment,
  onConfirm,
  onEdit,
  onCancel,
  currentStatus,
}: DashboardDetailModalProps) {
  const navigate = useNavigate();

  // Extract primitives before hooks so useCallback deps stay stable
  const petId = appointment?.petId;
  const appointmentId = appointment?.id;
  const ownerId = appointment?.ownerId;

  const navigateAndClose = useCallback((path: string, extraState?: Record<string, unknown>) => {
    navigate(path, { state: { from: "/", ...extraState } });
    onClose();
  }, [navigate, onClose]);

  const handleCreateMedicalRecord = useCallback((tab?: string) => {
    const base = petId
      ? `/medical-records/new?petId=${petId}${tab ? `&tab=${tab}` : ""}`
      : "/medical-records/select-pet";
    navigateAndClose(base, { appointmentId });
  }, [petId, appointmentId, navigateAndClose]);

  const handleCreateTrimming = useCallback(() =>
    navigateAndClose(petId ? `/trimming/new?petId=${petId}` : "/trimming/new"),
  [petId, navigateAndClose]);

  const handleCreateHospitalization = useCallback(() =>
    navigateAndClose(petId ? `/hospitalization/new?petId=${petId}` : "/hospitalization/new"),
  [petId, navigateAndClose]);

  const handleCreateAccounting = useCallback(() =>
    navigateAndClose(petId ? `/accounting/new?petId=${petId}` : "/accounting/new", {
      appointmentId,
    }),
  [petId, appointmentId, navigateAndClose]);

  const handleOpenOwnerDetail = useCallback(() => {
    if (ownerId) {
      navigateAndClose(`/owners/${ownerId}`);
    } else if (petId) {
      navigateAndClose(`/pets/${petId}`);
    }
  }, [ownerId, petId, navigateAndClose]);

  if (!appointment) return null;

  const isTrimming = appointment.serviceType.includes("トリミング");
  const isHospitalization = appointment.serviceType.includes("入院");
  const isMedical = appointment.serviceType.includes("診療") || (!isTrimming && !isHospitalization);

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[480px] p-0 gap-0 overflow-hidden bg-white">
        {/* Header */}
        <DialogHeader className={`p-5 pb-4 border-b ${C.borderLight} pr-12`}>
          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-2">
              <span className={`flex h-8 w-8 items-center justify-center rounded-full text-sm font-bold ${
                appointment.visitType === "初診"
                  ? `${C.bgAccentLight} ${C.accent}`
                  : `${C.bgActive} ${C.text60}`
              }`}>
                {appointment.visitType === "初診" ? "初" : "再"}
              </span>
              <DialogTitle className={`text-lg font-bold ${C.text}`}>
                {appointment.serviceType}
              </DialogTitle>
            </div>
            {currentStatus ? (
              <Badge variant="outline" className={`${DASHBOARD_STATUS_COLORS[currentStatus] ?? DASHBOARD_STATUS_COLOR_FALLBACK} px-3 py-1 text-sm font-medium border shrink-0`}>
                {currentStatus}
              </Badge>
            ) : null}
          </div>
          <DialogDescription className="sr-only">予約の詳細情報</DialogDescription>
        </DialogHeader>

        {/* Body */}
        <div className="p-5 space-y-4 overflow-y-auto">
          {/* Time */}
          <div className={`flex items-center gap-3 p-3 ${C.bgPage} rounded-lg`}>
            <Clock className={`${ICON.page} ${C.text60} shrink-0`} />
            <span className={`font-mono text-xl font-medium ${C.text}`}>{appointment.time}</span>
            {appointment.nextAppointment ? (
              <Badge
                variant={appointment.nextAppointment === "精算未確認" ? "destructive" : "secondary"}
                className="ml-auto flex items-center gap-1"
              >
                {appointment.nextAppointment === "精算未確認" ? <AlertCircle className={ICON.xs} /> : null}
                {appointment.nextAppointment}
              </Badge>
            ) : null}
          </div>

          {/* 患者情報 */}
          <div className="space-y-3">
            <h3 className={SECTION_LABEL}>患者情報</h3>
            <div className={DIVIDER_ROW}>
              <div className={ROW_ICON}>
                <Dog className={ICON.action} />
                <span className="text-sm">ペット</span>
              </div>
              <div className="text-right">
                <div className="font-bold text-base">{appointment.petName}</div>
                <div className={`text-sm ${C.text60}`}>{appointment.petType}</div>
              </div>
            </div>
            <div className={DIVIDER_ROW}>
              <div className={ROW_ICON}>
                <User className={ICON.action} />
                <span className="text-sm">飼い主</span>
              </div>
              <span className={`font-medium ${C.text}`}>{appointment.ownerName}</span>
            </div>
          </div>

          {/* 診療詳細 */}
          <div className="space-y-3">
            <h3 className={SECTION_LABEL}>診療詳細</h3>
            <div className={DIVIDER_ROW}>
              <div className={ROW_ICON}>
                <Stethoscope className={ICON.action} />
                <span className="text-sm">担当医</span>
              </div>
              <div className="flex items-center gap-2">
                <span className={`font-medium ${C.text}`}>{appointment.doctor || "未定"}</span>
                {appointment.isDesignated ? (
                  <Badge variant="outline" className={`text-xs h-6 ${C.bgDiscountLight} ${C.textDiscount} ${C.borderDiscount20}`}>指名</Badge>
                ) : null}
              </div>
            </div>
          </div>

          {/* 関連ページ */}
          <RelatedPages
            isTrimming={isTrimming}
            onCreateMedicalRecord={handleCreateMedicalRecord}
            onCreateTrimming={handleCreateTrimming}
            onCreateAccounting={handleCreateAccounting}
            onCreateHospitalization={handleCreateHospitalization}
          />
        </div>

        {/* Footer */}
        <DialogFooter className={`p-4 ${C.bgPage}`}>
          <div className="flex flex-wrap gap-2 justify-end w-full">
            <ActionButtons
              currentStatus={currentStatus}
              appointment={appointment}
              isTrimming={isTrimming}
              isHospitalization={isHospitalization}
              isMedical={isMedical}
              onConfirm={onConfirm}
              onEdit={onEdit}
              onCancel={onCancel}
              onOpenOwnerDetail={handleOpenOwnerDetail}
              onCreateMedicalRecord={handleCreateMedicalRecord}
              onCreateTrimming={handleCreateTrimming}
              onCreateAccounting={handleCreateAccounting}
              onCreateHospitalization={handleCreateHospitalization}
            />
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
});
