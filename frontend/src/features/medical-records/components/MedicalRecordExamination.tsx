// React/Framework
import { memo, useDeferredValue, useMemo, useState, useCallback, Suspense } from "react";
import { useSearchParams } from "react-router";

// Internal
import { normalizedIncludes } from "@/lib/normalize-kana";
import { usePermission } from "@/hooks/use-permission";

// Relative
import { useGetRecordExaminations } from "../api/get-record-examinations";
import { MEDICAL_RECORD_FIXED_ACTION_LAYOUT } from "../constants/fixed-action-layout";
import { ExaminationFilter } from "./ExaminationFilter";
import { ExaminationGroup } from "./ExaminationGroup";
import { ExaminationImportDialog } from "./MedicalRecordLazyModals";
import {
  MEDICAL_RECORD_EXAM_ID_PARAM,
  isTargetExamGroup,
  orderExamGroupsForTarget,
} from "../lib/medical-record-examination-model";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { Plus } from "lucide-react";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { LabDeviceUnlinkedBanner } from "@/components/shared/LabDeviceUnlinkedBanner/LabDeviceUnlinkedBanner";
import { getLockedEmptyStateMessage } from "../lib/record-empty-state";

interface MedicalRecordExaminationProps {
  isNewRecord?: boolean;
  petId?: string;
  medicalRecordId?: string;
  /** EMR-412 K5: ロック中は理由つき空状態文言。既定 false（旧挙動）。 */
  isLocked?: boolean;
  /** EMR-412 K5: 確定済み判定（K5 文言の優先理由）。既定 false。 */
  isFinalized?: boolean;
}

export const MedicalRecordExamination = memo(function MedicalRecordExamination({
  isNewRecord = false,
  petId,
  medicalRecordId,
  isLocked = false,
  isFinalized = false,
}: MedicalRecordExaminationProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearch = useDeferredValue(searchTerm);
  const [dateStart, setDateStart] = useState("");
  const [dateEnd, setDateEnd] = useState("");
  const [isImportDialogOpen, setIsImportDialogOpen] = useState(false);
  const [searchParams] = useSearchParams();
  const examId = searchParams.get(MEDICAL_RECORD_EXAM_ID_PARAM);
  const { canCreate } = usePermission("medical-records");
  // EMR-412 K5: 作成権限なしもロック扱いにし、取り込み導線を出さない。
  const isImportLocked = isLocked || !canCreate;
  const canImport = !isNewRecord && !isImportLocked;

  const {
    data: examinationResult,
    isLoading,
    refetch,
  } = useGetRecordExaminations(
    isNewRecord ? undefined : petId,
    isNewRecord ? undefined : medicalRecordId,
  );
  const apiExamGroups = useMemo(() => examinationResult?.items ?? [], [examinationResult?.items]);
  const isTruncated = examinationResult?.isTruncated ?? false;

  const examGroups = useMemo(() => {
    const filtered = apiExamGroups.filter((g) =>
      deferredSearch ? g.items.some((item) => normalizedIncludes(item.name, deferredSearch)) : true,
    );
    return orderExamGroupsForTarget(filtered, examId);
  }, [apiExamGroups, deferredSearch, examId]);

  const handleImportClick = useCallback(() => {
    if (!canImport) return;
    setIsImportDialogOpen(true);
  }, [canImport]);

  const handleImported = useCallback(() => {
    void refetch();
  }, [refetch]);

  // EMR-412 K5: ロック中（確定済/編集権限なし）は理由で文言を分ける。
  const lockedEmptyMessage = getLockedEmptyStateMessage({ isFinalized, isLocked: isImportLocked });

  return (
    <div
      className={`flex flex-col gap-3 flex-1 min-h-0 overflow-y-auto relative ${MEDICAL_RECORD_FIXED_ACTION_LAYOUT.scrollBottomClearance} pr-1`}
    >
      {petId && !isNewRecord ? <LabDeviceUnlinkedBanner petId={petId} /> : null}
      {/* Search & Actions Header */}
      <ExaminationFilter
        searchTerm={deferredSearch}
        onSearchChange={setSearchTerm}
        dateStart={dateStart}
        onDateStartChange={setDateStart}
        dateEnd={dateEnd}
        onDateEndChange={setDateEnd}
        onImport={canImport ? handleImportClick : undefined}
      />

      {/* Results Title */}
      <div>
        <h2 className={`text-sm font-semibold ${C.text} pl-1`}>検査結果一覧</h2>
        {isTruncated ? (
          <p className={`mt-1 pl-1 text-xs ${C.text50}`} role="status">
            直近{HISTORY_FETCH_LIMIT.toLocaleString()}件を表示しています
          </p>
        ) : null}
      </div>

      {/* Exam Groups */}
      {isLoading ? (
        <div className={`flex items-center justify-center h-24 text-sm ${C.text60} pl-1`}>
          読み込み中...
        </div>
      ) : examGroups.length === 0 ? (
        <div
          className={`flex flex-col items-center justify-center gap-3 h-24 text-sm ${C.text60} pl-1`}
        >
          <p>
            {isNewRecord
              ? "検査記録はまだありません。カルテ保存後に検査を追加できます。"
              : (lockedEmptyMessage ??
                "検査記録がありません。「記録を追加」ボタンから追加してください。")}
          </p>
          {/* EMR-412 K4: 追加ボタンは STYLE.inlineAddBtn に統一。ロック中は定期健診と同じく出さない。 */}
          {canImport ? (
            <button type="button" className={STYLE.inlineAddBtn} onClick={handleImportClick}>
              <Plus className={ICON.xs} />
              <span>記録を追加</span>
            </button>
          ) : null}
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
      {canImport && isImportDialogOpen ? (
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
