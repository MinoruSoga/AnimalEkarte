// React/Framework
import { memo, useState, useCallback, useTransition, useMemo } from "react";
import { toast } from "sonner";

// External
import { FlaskConical } from "lucide-react";

// Internal
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { handleApiError } from "@/lib/handle-api-error";
import { useGetExaminations } from "@/hooks/use-examinations";
import { useUpdateExamination } from "@/hooks/use-update-examination";
import {
  filterImportableExaminations,
  isExaminationImportable,
} from "./examination-import-candidates";

interface ExaminationImportDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  petId?: string;
  medicalRecordId?: string;
  onImported?: () => void;
}

export const ExaminationImportDialog = memo(function ExaminationImportDialog({
  open,
  onOpenChange,
  petId,
  medicalRecordId,
  onImported,
}: ExaminationImportDialogProps) {
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [isLinking, startLinkTransition] = useTransition();

  const { data: examinations = [], isLoading } = useGetExaminations(
    petId ? { petId } : undefined
  );

  // BUG-014: 確定済み・リビジョン済みなど取込不可の検査は候補から除外する
  const availableExams = useMemo(
    () => filterImportableExaminations(examinations, medicalRecordId),
    [examinations, medicalRecordId],
  );

  const { mutateAsync: updateExamination } = useUpdateExamination();

  const handleToggle = useCallback((id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }, []);

  const handleImport = useCallback(() => {
    if (!medicalRecordId || selectedIds.size === 0) return;
    const importableIds = Array.from(selectedIds).filter((id) => {
      const exam = examinations.find((e) => e.id === id);
      return exam != null && isExaminationImportable(exam, medicalRecordId);
    });
    if (importableIds.length === 0) {
      toast.error("取込可能な検査が選択されていません");
      setSelectedIds(new Set());
      return;
    }
    startLinkTransition(async () => {
      try {
        await Promise.all(
          importableIds.map((id) =>
            updateExamination({
              id,
              req: { medical_record_id: Number(medicalRecordId) },
            }),
          ),
        );
        toast.success(`${importableIds.length}件の検査記録を取り込みました`);
        setSelectedIds(new Set());
        onImported?.();
        onOpenChange(false);
      } catch (err) {
        handleApiError(err, "検査取り込み");
      }
    });
  }, [
    examinations,
    medicalRecordId,
    selectedIds,
    updateExamination,
    onImported,
    onOpenChange,
  ]);

  const handleClose = useCallback(() => {
    setSelectedIds(new Set());
    onOpenChange(false);
  }, [onOpenChange]);

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-xl max-h-[80vh] flex flex-col">
        <DialogHeader>
          <DialogTitle className={`flex items-center gap-2 text-base font-bold ${C.text}`}>
            <FlaskConical className={`${ICON.action} ${C.textMedicalBlue}`} />
            検査取り込み
          </DialogTitle>
          <DialogDescription>
            このペットの検査記録からカルテに取り込む検査を選択します。
          </DialogDescription>
        </DialogHeader>

        {/* Exam List */}
        <div className="flex-1 overflow-y-auto min-h-0 space-y-2 py-1">
          {isLoading ? (
            <div className={`text-sm ${C.text40} text-center py-8`}>読み込み中...</div>
          ) : availableExams.length === 0 ? (
            <div className={`text-sm ${C.text40} text-center py-8`}>取り込める検査記録がありません</div>
          ) : (
            availableExams.map((exam) => {
              const isSelected = selectedIds.has(exam.id);
              return (
                <button
                  key={exam.id}
                  type="button"
                  onClick={() => handleToggle(exam.id)}
                  className={`w-full text-left p-3 rounded-lg border transition-colors ${
                    isSelected
                      ? `${C.borderBlue400} ${C.bgStatusBlueLight}`
                      : `${C.borderMedium} ${C.bgWhite} ${STYLE.tableRowHover}`
                  }`}
                >
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex flex-col gap-0.5 min-w-0">
                      <span className={`text-sm font-medium ${C.text} truncate`}>
                        {exam.testType ?? "（種別未設定）"}
                      </span>
                      <span className={`text-xs ${C.text60}`}>
                        {exam.date ? exam.date : "日付未設定"}
                        {exam.doctor ? ` · ${exam.doctor}` : ""}
                      </span>
                    </div>
                    <div
                      className={`w-4 h-4 rounded border-2 flex-shrink-0 ${
                        isSelected ? `${C.bgStatusBlueDot} ${C.borderBlue500}` : `${C.borderGray300} ${C.bgWhite}`
                      }`}
                    />
                  </div>
                </button>
              );
            })
          )}
        </div>

        {/* Footer */}
        <div className={`flex justify-end gap-3 pt-2 border-t ${C.borderGray100}`}>
          <Button type="button" variant="outline" onClick={handleClose}>
            キャンセル
          </Button>
          <Button
            type="button"
            onClick={handleImport}
            disabled={selectedIds.size === 0 || isLinking}
            className={`${C.bgBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} ${C.textOnBrand} rounded-full border-transparent`}
          >
            {isLinking ? "取り込み中..." : `${selectedIds.size}件取り込む`}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
});
