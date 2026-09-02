// React/Framework
import { memo, useDeferredValue, useMemo, useState, useCallback, Suspense } from "react";
import { useSearchParams } from "react-router";

// Internal
import { normalizedIncludes } from "@/lib/normalize-kana";

// Relative
import { useGetRecordExaminations } from "../api/get-record-examinations";
import { ExaminationFilter } from "./ExaminationFilter";
import { ExaminationGroup } from "./ExaminationGroup";
import { ExaminationImportDialog } from "./MedicalRecordLazyModals";
import {
  MEDICAL_RECORD_EXAM_ID_PARAM,
  isTargetExamGroup,
  orderExamGroupsForTarget,
} from "./medical-record-examination-model";
import { C } from "@/lib/design-tokens";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { LabDeviceUnlinkedBanner } from "@/components/shared/LabDeviceUnlinkedBanner/LabDeviceUnlinkedBanner";

interface MedicalRecordExaminationProps {
  isNewRecord?: boolean;
  petId?: string;
  medicalRecordId?: string;
}

export const MedicalRecordExamination = memo(function MedicalRecordExamination({
  isNewRecord = false,
  petId,
  medicalRecordId,
}: MedicalRecordExaminationProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearch = useDeferredValue(searchTerm);
  const [dateStart, setDateStart] = useState("");
  const [dateEnd, setDateEnd] = useState("");
  const [isImportDialogOpen, setIsImportDialogOpen] = useState(false);
  const [searchParams] = useSearchParams();
  const examId = searchParams.get(MEDICAL_RECORD_EXAM_ID_PARAM);

  const { data: examinationResult, isLoading, refetch } = useGetRecordExaminations(
    isNewRecord ? undefined : petId,
    isNewRecord ? undefined : medicalRecordId,
  );
  const apiExamGroups = useMemo(
    () => examinationResult?.items ?? [],
    [examinationResult?.items],
  );
  const isTruncated = examinationResult?.isTruncated ?? false;

  const examGroups = useMemo(() => {
    const filtered = apiExamGroups.filter((g) =>
      deferredSearch
        ? g.items.some((item) => normalizedIncludes(item.name, deferredSearch))
        : true,
    );
    return orderExamGroupsForTarget(filtered, examId);
  }, [apiExamGroups, deferredSearch, examId]);

  const handleImportClick = useCallback(() => {
    setIsImportDialogOpen(true);
  }, []);

  const handleImported = useCallback(() => {
    void refetch();
  }, [refetch]);

  return (
    <div className="h-[calc(100vh-220px)] min-h-[500px] flex flex-col gap-3 overflow-y-auto pb-20 pr-1">
      {petId && !isNewRecord ? <LabDeviceUnlinkedBanner petId={petId} /> : null}
      {/* Search & Actions Header */}
      <ExaminationFilter
        searchTerm={deferredSearch}
        onSearchChange={setSearchTerm}
        dateStart={dateStart}
        onDateStartChange={setDateStart}
        dateEnd={dateEnd}
        onDateEndChange={setDateEnd}
        onImport={isNewRecord ? undefined : handleImportClick}
      />

      {/* Results Title */}
      <div>
        <h2 className={`text-sm font-bold ${C.text} pl-1`}>検査結果一覧</h2>
        {isTruncated ? (
          <p className={`mt-1 pl-1 text-xs ${C.text50}`} role="status">
            直近{HISTORY_FETCH_LIMIT}件を表示しています
          </p>
        ) : null}
      </div>

      {/* Exam Groups */}
      {isLoading ? (
        <div className={`flex items-center justify-center h-24 text-sm ${C.text40} pl-1`}>
          読み込み中...
        </div>
      ) : examGroups.length === 0 ? (
        <div className={`flex flex-col items-center justify-center gap-3 h-24 text-sm ${C.text40} pl-1`}>
          <p>
            {isNewRecord
              ? "検査記録はまだありません。カルテ保存後に検査を追加できます。"
              : "検査記録がありません。下の「記録を追加」ボタンから追加してください。"}
          </p>
          {isNewRecord ? null : (
            <button type="button" className="underline" onClick={handleImportClick}>
              記録を追加
            </button>
          )}
        </div>
      ) : null}
      <div className="flex flex-col gap-4 pl-1">
        {!isLoading
          ? examGroups.map((group) => (
              <ExaminationGroup
                key={group.id}
                group={group}
                petId={petId}
                highlighted={isTargetExamGroup(group.id, examId)}
              />
            ))
          : null}
      </div>

      {/* Examination Import Dialog */}
      {isImportDialogOpen ? (
        <Suspense fallback={null}>
          <ExaminationImportDialog
            open={isImportDialogOpen}
            onOpenChange={setIsImportDialogOpen}
            petId={petId}
            medicalRecordId={medicalRecordId}
            onImported={handleImported}
          />
        </Suspense>
      ) : null}
    </div>
  );
});
