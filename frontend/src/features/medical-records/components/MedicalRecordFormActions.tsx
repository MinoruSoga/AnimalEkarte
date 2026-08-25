import { HeartPulse, Lock, Printer, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { C, ICON, Z_CLASS } from "@/lib/design-tokens";
import { formatJSTDate } from "@/lib/jst-date";
import { MedicalRecordPrintView } from "./MedicalRecordPrintView";
import type { Treatment } from "../types";

interface MedicalRecordFloatingActionsProps {
  activeTab: string;
  canDelete: boolean;
  canEdit: boolean;
  canSubmit: boolean;
  isNewRecord: boolean;
  isCreating: boolean;
  isSaving: boolean;
  isFinalized: boolean;
  onDeleteClick: () => void;
  onVitalsClick: () => void;
  onPrintClick: () => void;
  onFinalizeClick: () => void;
}

export function MedicalRecordFloatingActions({
  activeTab,
  canDelete,
  canEdit,
  canSubmit,
  isNewRecord,
  isCreating,
  isSaving,
  isFinalized,
  onDeleteClick,
  onVitalsClick,
  onPrintClick,
  onFinalizeClick,
}: MedicalRecordFloatingActionsProps) {
  if (activeTab === "会計(医師確認)") return null;

  return (
    <div className="fixed bottom-6 right-6 z-50 flex gap-2">
      {canDelete && !isNewRecord && activeTab === "問診" ? (
        <Button
          type="button"
          variant="ghost-danger"
          onClick={onDeleteClick}
          className={`border ${C.borderDanger} h-10 text-sm px-4`}
        >
          <Trash2 className={ICON.action} />
          削除
        </Button>
      ) : null}
      {activeTab !== "見積書" && canEdit ? (
        <Button
          type="button"
          variant="outline"
          onClick={onVitalsClick}
          disabled={isNewRecord || isFinalized}
          title={
            isNewRecord
              ? "カルテを保存してから利用できます"
              : isFinalized
                ? "確定済みカルテのためバイタルを追加できません"
                : undefined
          }
          className="h-10 text-sm px-4"
        >
          <HeartPulse className={ICON.action} />
          バイタル記録
        </Button>
      ) : null}
      {!isNewRecord ? (
        <Button
          type="button"
          variant="outline"
          onClick={onPrintClick}
          className="h-10 text-sm px-4"
        >
          <Printer className={ICON.action} />
          印刷
        </Button>
      ) : null}
      {/* SPEC-GAP: カルテ確定（Lock）。確定済みカルテは編集不可となり、以降の修正は追記(addendum)のみ */}
      {canEdit && !isNewRecord && !isFinalized ? (
        <Button
          type="button"
          variant="outline"
          onClick={onFinalizeClick}
          className="h-10 text-sm px-4"
        >
          <Lock className={ICON.action} />
          確定する
        </Button>
      ) : null}
      {/* BUG-035: 確定済みでは保存を出さない（disabled 表示より導線除去）。印刷・追記は残す。 */}
      {/* BUG-001: 予防接種の永続化は inner の接種記録追加。外側保存は偽陽性になるので出さない。 */}
      {canSubmit && !isFinalized && activeTab !== "予防接種" ? (
        <SubmitButton
          className="px-4"
          disabled={isCreating || isSaving}
        >
          {isCreating ? "カルテ作成中..." : "保存"}
        </SubmitButton>
      ) : null}
    </div>
  );
}

interface MedicalRecordDeleteDialogProps {
  open: boolean;
  petName: string;
  isDeleting: boolean;
  onClose: () => void;
  onConfirm: () => void;
}

export function MedicalRecordDeleteDialog({
  open,
  petName,
  isDeleting,
  onClose,
  onConfirm,
}: MedicalRecordDeleteDialogProps) {
  return (
    <ConfirmDialog
      open={open}
      onClose={onClose}
      onConfirm={onConfirm}
      title="カルテを削除しますか？"
      description={`${petName}のカルテデータを削除します。この操作は元に戻せません。`}
      confirmLabel={isDeleting ? "削除中..." : "削除する"}
      cancelLabel="キャンセル"
      variant="destructive"
    />
  );
}

interface MedicalRecordFinalizeDialogProps {
  open: boolean;
  isFinalizing: boolean;
  onClose: () => void;
  onConfirm: () => void;
}

// SPEC-GAP: カルテ確定の取り消し（unfinalize）API は存在せず、確定後の修正経路は
// 追記(addendum)のみ。ロック/Undo で代替できないため、削除ダイアログと同じ
// ConfirmDialog パターンで不可逆操作の唯一の停止手段とする。
export function MedicalRecordFinalizeDialog({
  open,
  isFinalizing,
  onClose,
  onConfirm,
}: MedicalRecordFinalizeDialogProps) {
  return (
    <ConfirmDialog
      open={open}
      onClose={onClose}
      onConfirm={onConfirm}
      title="カルテを確定しますか？"
      description="確定後はカルテを編集できなくなります。修正が必要な場合は訂正追記（addendum）を使用してください。この操作は元に戻せません。"
      confirmLabel={isFinalizing ? "確定中..." : "確定する"}
      cancelLabel="キャンセル"
      isPending={isFinalizing}
    />
  );
}

interface MedicalRecordPrintAreaProps {
  isPrinting: boolean;
  isNewRecord: boolean;
  recordId?: string;
  doctorName: string;
  recordDate?: string;
  pet: {
    name: string;
    species?: string;
    ownerName?: string;
  } | null;
  clinic?: {
    name?: string;
    address?: string;
    phoneNumber?: string;
  };
  chiefComplaint?: string;
  treatmentPolicy?: string;
  physicalExam?: string;
  diagnosisDetails?: string;
  treatments: Treatment[];
}

export function MedicalRecordPrintArea({
  isPrinting,
  isNewRecord,
  recordId,
  doctorName,
  recordDate,
  pet,
  clinic,
  chiefComplaint,
  treatmentPolicy,
  physicalExam,
  diagnosisDetails,
  treatments,
}: MedicalRecordPrintAreaProps) {
  if (!isPrinting || isNewRecord || !pet) return null;

  return (
    <div className={`hidden print:block fixed inset-0 ${C.bgWhite} ${Z_CLASS.overlay}`}>
      <style type="text/css" media="print">
        {`@page { size: A4 portrait; margin: 15mm; } body { margin: 0; -webkit-print-color-adjust: exact; }`}
      </style>
      <MedicalRecordPrintView
        recordNo={recordId}
        date={recordDate ?? formatJSTDate(new Date())}
        doctorName={doctorName}
        pet={pet}
        clinic={{
          name: clinic?.name,
          address: clinic?.address,
          phone: clinic?.phoneNumber,
        }}
        chiefComplaint={chiefComplaint}
        treatmentPolicy={treatmentPolicy}
        physicalExam={physicalExam}
        diagnosisDetails={diagnosisDetails}
        treatments={treatments}
      />
    </div>
  );
}
