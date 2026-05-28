import { HeartPulse, Printer, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { MedicalRecordPrintView } from "./MedicalRecordPrintView";
import type { Treatment } from "../types";

interface MedicalRecordFloatingActionsProps {
  activeTab: string;
  canDelete: boolean;
  canEdit: boolean;
  canSubmit: boolean;
  isNewRecord: boolean;
  isCreating: boolean;
  onDeleteClick: () => void;
  onVitalsClick: () => void;
  onPrintClick: () => void;
}

export function MedicalRecordFloatingActions({
  activeTab,
  canDelete,
  canEdit,
  canSubmit,
  isNewRecord,
  isCreating,
  onDeleteClick,
  onVitalsClick,
  onPrintClick,
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
          disabled={isNewRecord}
          title={isNewRecord ? "カルテを保存してから利用できます" : undefined}
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
      {canSubmit ? (
        <SubmitButton
          className={`${STYLE.btnPrimary} px-5`}
          disabled={isCreating}
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

interface MedicalRecordPrintAreaProps {
  isPrinting: boolean;
  isNewRecord: boolean;
  recordId?: string;
  doctorName: string;
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
    <div className={`hidden print:block fixed inset-0 ${C.bgWhite} z-[9999]`}>
      <style type="text/css" media="print">
        {`@page { size: A4 portrait; margin: 15mm; } body { margin: 0; -webkit-print-color-adjust: exact; }`}
      </style>
      <MedicalRecordPrintView
        recordNo={recordId}
        date={new Date().toLocaleDateString("ja-JP")}
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
