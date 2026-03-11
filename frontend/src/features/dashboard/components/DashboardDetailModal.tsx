// React/Framework
import { useNavigate } from "react-router";

// External
import {
  Clock,
  User,
  Dog,
  Stethoscope,
  AlertCircle,
  FileText,
  CreditCard,
  Scissors,
  Trash2,
  Pencil,
  TestTube,
  BedDouble,
  ExternalLink,
} from "lucide-react";

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
import { C } from "@/lib/design-tokens";

// Types
import type { Appointment } from "@/types";

interface DashboardDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  appointment: Appointment | null;
  onConfirm?: () => void;
  onEdit?: (appointment: Appointment) => void;
  onCancel?: (appointment: Appointment) => void;
  currentStatus?: string;
}

const STATUS_COLOR: Record<string, string> = {
  "受付予約": `bg-[#D3E5EF] text-[#183B56] border-[#2383E2]/30`,
  "受付済":   `bg-[#DDEDEA] text-[#0F7B6C] border-[#DDEDEA]`,
  "診療中":   `bg-[#EEE0F7] text-[#6940A5] border-[#6940A5]/20`,
  "会計待ち": `bg-[#FFF3CD]/50 text-[#B58105] border-[#B58105]/20`,
  "会計済":   `bg-[#EAE9E5] text-[#37352F] border-[rgba(55,53,47,0.09)]`,
};

export const DashboardDetailModal = ({
  isOpen,
  onClose,
  appointment,
  onConfirm,
  onEdit,
  onCancel,
  currentStatus,
}: DashboardDetailModalProps) => {
  const navigate = useNavigate();

  if (!appointment) return null;

  const isTrimming = appointment.serviceType.includes("トリミング");
  const isHospitalization = appointment.serviceType.includes("入院");
  const isMedical = appointment.serviceType.includes("診療") || (!isTrimming && !isHospitalization);

  const navigateTo = (path: string, extraState?: Record<string, unknown>) => {
    navigate(path, { state: { from: "/", ...extraState } });
    onClose();
  };

  const handleCreateMedicalRecord = (tab?: string) => {
    const base = appointment.petId
      ? `/medical-records/new?petId=${appointment.petId}${tab ? `&tab=${tab}` : ""}`
      : "/medical-records/select-pet";
    navigateTo(base, { appointmentId: appointment.id });
  };

  const handleCreateTrimming = () =>
    navigateTo(appointment.petId ? `/trimming/new?petId=${appointment.petId}` : "/trimming/new");

  const handleCreateHospitalization = () =>
    navigateTo(appointment.petId ? `/hospitalization/new?petId=${appointment.petId}` : "/hospitalization/new");

  const handleCreateAccounting = () =>
    navigateTo(appointment.petId ? `/accounting/new?petId=${appointment.petId}` : "/accounting/new", {
      appointmentId: appointment.id,
    });

  const handleOpenOwnerDetail = () => {
    if (appointment.ownerId) {
      navigate(`/owners/${appointment.ownerId}`);
      onClose();
    } else if (appointment.petId) {
      navigate(`/pets/${appointment.petId}`);
      onClose();
    }
  };

  // 関連ページ: サービス種別に応じてカルテ/施術・会計・入院を表示
  const renderRelatedPages = () => (
    <div className="space-y-2">
      <h3 className="text-sm font-semibold text-[#37352F]/60 uppercase tracking-wider">関連ページ</h3>
      <div className="flex flex-wrap gap-2">
        {/* カルテ / 施術 */}
        <button
          type="button"
          className="flex items-center gap-1.5 text-sm text-[#2383E2] bg-[#D3E5EF]/40 hover:bg-[#D3E5EF] border border-[#2383E2]/30 rounded-md px-3 py-1.5 transition-colors group"
          onClick={() => {
            if (isTrimming) handleCreateTrimming();
            else handleCreateMedicalRecord();
          }}
        >
          {isTrimming ? <Scissors className="size-3.5" /> : <FileText className="size-3.5" />}
          <span>{isTrimming ? "施術" : "カルテ"}</span>
          <ExternalLink className="size-3 opacity-0 group-hover:opacity-100 transition-opacity" />
        </button>

        {/* 会計 */}
        <button
          type="button"
          className="flex items-center gap-1.5 text-sm text-[#0F7B6C] bg-[#DDEDEA]/40 hover:bg-[#DDEDEA] border border-[#DDEDEA] rounded-md px-3 py-1.5 transition-colors group"
          onClick={handleCreateAccounting}
        >
          <CreditCard className="size-3.5" />
          <span>会計</span>
          <ExternalLink className="size-3 opacity-0 group-hover:opacity-100 transition-opacity" />
        </button>

        {/* 入院 */}
        <button
          type="button"
          className="flex items-center gap-1.5 text-sm text-[#6940A5] bg-[#EEE0F7]/40 hover:bg-[#EEE0F7] border border-[#6940A5]/20 rounded-md px-3 py-1.5 transition-colors group"
          onClick={handleCreateHospitalization}
        >
          <BedDouble className="size-3.5" />
          <span>入院</span>
          <ExternalLink className="size-3 opacity-0 group-hover:opacity-100 transition-opacity" />
        </button>
      </div>
    </div>
  );

  const renderActions = () => {
    const ownerDetailBtn = (
      <Button variant="ghost" onClick={handleOpenOwnerDetail} className="h-10 text-sm text-[#37352F]">
        <User className="size-4" />
        飼主詳細
      </Button>
    );

    if (currentStatus === "受付予約") {
      return (
        <>
          {onCancel && (
            <Button variant="ghost" onClick={() => onCancel(appointment)} className={`h-10 text-sm ${C.danger} ${C.hoverBgDanger5}`}>
              <Trash2 className="size-4" />
              取消
            </Button>
          )}
          {onEdit && (
            <Button variant="outline" onClick={() => onEdit(appointment)} className="h-10 text-sm text-[#37352F] border-[rgba(55,53,47,0.16)]">
              <Pencil className="size-4" />
              編集
            </Button>
          )}
          {ownerDetailBtn}
          {onConfirm && (
            <Button onClick={onConfirm} className="h-10 text-sm bg-[#2383E2] hover:bg-[#1B6EC2] text-white rounded-[4px] transition-colors shadow-none border-transparent">
              受付済にする
            </Button>
          )}
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
                onClick={() => { if (onConfirm) onConfirm(); handleCreateMedicalRecord(); }}
                className="h-10 text-sm bg-[#2383E2] hover:bg-[#1B6EC2] text-white rounded-[4px] transition-colors shadow-none border-transparent"
              >
                <FileText className="size-4" />
                カルテ作成
              </Button>
              <span className="text-[10px] text-muted-foreground">※カルテ作成と同時に「診療中」へ移動します</span>
            </div>
          ) : (
            onConfirm && (
              <Button onClick={onConfirm} className="h-10 text-sm bg-[#2383E2] hover:bg-[#1B6EC2] text-white rounded-[4px] transition-colors shadow-none border-transparent">
                診察を開始する
              </Button>
            )
          )}
        </>
      );
    }

    if (currentStatus === "診療中") {
      return (
        <>
          {ownerDetailBtn}
          {onConfirm && (
            <Button variant="outline" onClick={onConfirm} className={`h-10 text-sm ${C.danger} ${C.borderDanger} ${C.hoverBgDanger5}`}>
              診察を終了する
            </Button>
          )}
          {isMedical && (
            <>
              <Button onClick={() => handleCreateMedicalRecord()} className="h-10 text-sm bg-[#2383E2] hover:bg-[#1B6EC2] text-white rounded-[4px] transition-colors shadow-none border-transparent">
                <FileText className="size-4" />
                カルテ入力
              </Button>
              <Button variant="outline" onClick={() => handleCreateMedicalRecord("test")} className={`h-10 text-sm ${C.text} ${C.borderMedium}`}>
                <TestTube className="size-4" />
                検査
              </Button>
            </>
          )}
          {isTrimming && (
            <Button onClick={handleCreateTrimming} className="h-10 text-sm bg-[#D9730D] hover:bg-[#D9730D]/90 text-white rounded-[4px] transition-colors shadow-none border-transparent">
              <Scissors className="size-4" />
              施術記録
            </Button>
          )}
        </>
      );
    }

    if (currentStatus === "会計待ち") {
      return (
        <>
          {ownerDetailBtn}
          <Button
            onClick={handleCreateAccounting}
            className="h-10 text-sm bg-[#2383E2] hover:bg-[#1B6EC2] text-white rounded-[4px] transition-colors shadow-none border-transparent"
          >
            <CreditCard className="size-4" />
            会計へ進む
          </Button>
        </>
      );
    }

    if (currentStatus === "会計済") {
      return (
        <>
          {ownerDetailBtn}
          {onConfirm && (
            <Button onClick={onConfirm} className="h-10 text-sm bg-[#2383E2] hover:bg-[#1B6EC2] text-white rounded-[4px] transition-colors shadow-none border-transparent">
              完了/リストから削除
            </Button>
          )}
        </>
      );
    }

    // Default Fallback
    return (
      <>
        {ownerDetailBtn}
        {isMedical && (
          <Button onClick={() => handleCreateMedicalRecord()} className="h-10 text-sm bg-[#2383E2] hover:bg-[#1B6EC2] text-white rounded-[4px] transition-colors shadow-none border-transparent">
            <FileText className="size-4" />
            カルテ確認
          </Button>
        )}
        {isTrimming && (
          <Button onClick={handleCreateTrimming} className="h-10 text-sm bg-[#2383E2] hover:bg-[#1B6EC2] text-white rounded-[4px] transition-colors shadow-none border-transparent">
            <Scissors className="size-4" />
            トリミング
          </Button>
        )}
        {isHospitalization && (
          <Button onClick={handleCreateHospitalization} className="h-10 text-sm bg-[#2383E2] hover:bg-[#1B6EC2] text-white rounded-[4px] transition-colors shadow-none border-transparent">
            <BedDouble className="size-4" />
            入院登録
          </Button>
        )}
        {onConfirm && (
          <Button onClick={onConfirm} className="h-10 text-sm bg-[#2383E2] hover:bg-[#1B6EC2] text-white rounded-[4px] transition-colors shadow-none border-transparent">
            ステータス変更
          </Button>
        )}
      </>
    );
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[480px] p-0 gap-0 overflow-hidden bg-white">
        {/* Header */}
        <DialogHeader className="p-5 pb-4 border-b border-[rgba(55,53,47,0.09)] pr-12">
          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-2">
              <span className={`flex h-8 w-8 items-center justify-center rounded-full text-sm font-bold ${
                appointment.visitType === "初診"
                  ? "bg-[#D3E5EF] text-[#2383E2]"
                  : "bg-[#EAE9E5] text-[#37352F]/60"
              }`}>
                {appointment.visitType === "初診" ? "初" : "再"}
              </span>
              <DialogTitle className="text-lg font-bold text-[#37352F]">
                {appointment.serviceType}
              </DialogTitle>
            </div>
            {currentStatus && (
              <Badge variant="outline" className={`${STATUS_COLOR[currentStatus] ?? "bg-gray-100 text-gray-600 border-gray-200"} px-3 py-1 text-sm font-medium border shrink-0`}>
                {currentStatus}
              </Badge>
            )}
          </div>
          <DialogDescription className="sr-only">予約の詳細情報</DialogDescription>
        </DialogHeader>

        {/* Body */}
        <div className="p-5 space-y-4 overflow-y-auto">
          {/* Time */}
          <div className="flex items-center gap-3 p-3 bg-[#F7F6F3] rounded-lg">
            <Clock className="size-5 text-[#37352F]/60 shrink-0" />
            <span className="font-mono text-xl font-medium text-[#37352F]">{appointment.time}</span>
            {appointment.nextAppointment && (
              <Badge
                variant={appointment.nextAppointment === "精算未確認" ? "destructive" : "secondary"}
                className="ml-auto flex items-center gap-1"
              >
                {appointment.nextAppointment === "精算未確認" && <AlertCircle className="size-3" />}
                {appointment.nextAppointment}
              </Badge>
            )}
          </div>

          {/* 患者情報 */}
          <div className="space-y-3">
            <h3 className="text-sm font-semibold text-[#37352F]/60 uppercase tracking-wider">患者情報</h3>
            <div className="flex items-center justify-between border-b border-[rgba(55,53,47,0.09)] pb-2">
              <div className="flex items-center gap-2 text-[#37352F]/60">
                <Dog className="size-4" />
                <span className="text-sm">ペット</span>
              </div>
              <div className="text-right">
                <div className="font-bold text-base">{appointment.petName}</div>
                <div className="text-sm text-[#37352F]/60">{appointment.petType}</div>
              </div>
            </div>
            <div className="flex items-center justify-between border-b border-[rgba(55,53,47,0.09)] pb-2">
              <div className="flex items-center gap-2 text-[#37352F]/60">
                <User className="size-4" />
                <span className="text-sm">飼い主</span>
              </div>
              <span className="font-medium text-[#37352F]">{appointment.ownerName}</span>
            </div>
          </div>

          {/* 診療詳細 */}
          <div className="space-y-3">
            <h3 className="text-sm font-semibold text-[#37352F]/60 uppercase tracking-wider">診療詳細</h3>
            <div className="flex items-center justify-between border-b border-[rgba(55,53,47,0.09)] pb-2">
              <div className="flex items-center gap-2 text-[#37352F]/60">
                <Stethoscope className="size-4" />
                <span className="text-sm">担当医</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="font-medium text-[#37352F]">{appointment.doctor || "未定"}</span>
                {appointment.isDesignated && (
                  <Badge variant="outline" className="text-xs h-6 bg-[#FAEBDD] text-[#D9730D] border-[#D9730D]/20">指名</Badge>
                )}
              </div>
            </div>
          </div>

          {/* 関連ページ */}
          {renderRelatedPages()}
        </div>

        {/* Footer */}
        <DialogFooter className="p-4 bg-[#F7F6F3]">
          <div className="flex flex-wrap gap-2 justify-end w-full">
            {renderActions()}
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
