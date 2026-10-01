// React/Framework
import { memo, useCallback, useState } from "react";

// Internal
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";

// Relative
import { InterviewChiefComplaint } from "./InterviewChiefComplaint";
import { InterviewTreatmentPolicy } from "./InterviewTreatmentPolicy";
import { InterviewHistory } from "./InterviewHistory";
import {
  DEFAULT_CHIEF_COMPLAINT,
  DEFAULT_TREATMENT_POLICY,
} from "../hooks/use-medical-record-form-model";
import type { InterviewHistoryCopySource, InterviewHistoryItem } from "../types";

interface MedicalRecordInterviewProps {
  chiefComplaint: string;
  setChiefComplaint: (value: string) => void;
  chiefComplaintTypeId: number | null;
  setChiefComplaintTypeId: (id: number | null) => void;
  treatmentPolicy: string;
  setTreatmentPolicy: (value: string) => void;
  historyItems?: InterviewHistoryItem[];
  setHistoryItems?: (items: InterviewHistoryItem[]) => void;
  /** BUG-035 residual: 問診臨床欄を content attribute で固定 */
  isFinalized?: boolean;
  /** BUG-035: 確定済み/送信権限なし。編集2列のみ disabled fieldset に入れ、履歴検索はロック外に残す。 */
  isLocked?: boolean;
  /** EMR-219: 前回複写適用時に元カルテの治療明細行も複写する（複写元カルテ id を引く） */
  onCopyRecordTreatments?: (sourceRecordId: string) => void;
}

// rendering-hoist-jsx: テンプレート一覧は静的なのでモジュール定数に巻き上げ
const INTERVIEW_TEMPLATES: { label: string; text: string }[] = [
  { label: "定期検診", text: "# 定期検診\n特に異常なし。食欲・元気あり。" },
  { label: "ワクチン", text: "# 混合ワクチン接種\n体調良好。" },
  {
    label: "下痢・嘔吐",
    text: "# 消化器症状\n・嘔吐：あり（回数：　）\n・下痢：あり（性状：　）\n・食欲：なし",
  },
  { label: "皮膚", text: "# 皮膚症状\n・痒み：あり\n・発赤：あり\n・部位：" },
];

// S16 BUG-INTERVIEW-HISTORY-DEMO-ROWS: 履歴0件時にデモ行（id=1..3 →
// /medical-records/{1,2,3} への実リンク）を表示すると無関係なカルテへ誘導するため、
// API 結果が空/未ロードのときは必ず空状態を出す（InterviewHistory 側に EmptyState あり）。
const EMPTY_HISTORY_ITEMS: InterviewHistoryItem[] = [];

export const MedicalRecordInterview = memo(function MedicalRecordInterview({
  chiefComplaint,
  setChiefComplaint,
  chiefComplaintTypeId,
  setChiefComplaintTypeId,
  treatmentPolicy,
  setTreatmentPolicy,
  historyItems,
  isFinalized = false,
  isLocked = false,
  onCopyRecordTreatments,
}: MedicalRecordInterviewProps) {
  const handleInsertTemplate = useCallback(
    (text: string) => {
      setChiefComplaint(text);
    },
    [setChiefComplaint],
  );

  // EMR-182: 前回複写。copySource に存在する項目だけを現在の setter に流す（送信は行わない）。
  const [pendingCopy, setPendingCopy] = useState<InterviewHistoryCopySource | null>(null);

  const applyCopySource = useCallback(
    (source: InterviewHistoryCopySource) => {
      if (source.chiefComplaint !== undefined) setChiefComplaint(source.chiefComplaint);
      if (source.treatmentPolicy !== undefined) setTreatmentPolicy(source.treatmentPolicy);
      if (source.chiefComplaintTypeId !== undefined) {
        setChiefComplaintTypeId(source.chiefComplaintTypeId);
      }
      // EMR-219: 問診項目に加えて治療明細行も複写する。即時適用・確認ダイアログ適用の
      // 双方がこの一点を通るため、ここでのみ発火する。recordId が無い履歴行は明細複写なし。
      if (source.recordId) onCopyRecordTreatments?.(source.recordId);
    },
    [setChiefComplaint, setChiefComplaintTypeId, setTreatmentPolicy, onCopyRecordTreatments],
  );

  const handleCopyItem = useCallback(
    (item: InterviewHistoryItem) => {
      const source = item.copySource;
      if (!source) return;
      const untouched =
        chiefComplaint === DEFAULT_CHIEF_COMPLAINT &&
        treatmentPolicy === DEFAULT_TREATMENT_POLICY &&
        chiefComplaintTypeId === null;
      if (untouched) {
        applyCopySource(source);
      } else {
        setPendingCopy(source);
      }
    },
    [chiefComplaint, chiefComplaintTypeId, treatmentPolicy, applyCopySource],
  );

  const handleConfirmCopy = useCallback(() => {
    if (pendingCopy) applyCopySource(pendingCopy);
    setPendingCopy(null);
  }, [pendingCopy, applyCopySource]);

  const handleCloseCopyConfirm = useCallback(() => setPendingCopy(null), []);

  const resolvedHistoryItems = historyItems ?? EMPTY_HISTORY_ITEMS;

  return (
    <div className="grid grid-cols-1 lg:grid-cols-12 lg:grid-rows-1 gap-3 flex-1 min-h-0 h-full">
      <fieldset disabled={isLocked} className="contents">
        {/* Left Column: 主訴情報 (Chief Complaint) */}
        <InterviewChiefComplaint
          className="col-span-1 lg:col-span-3 h-full"
          chiefComplaint={chiefComplaint}
          setChiefComplaint={setChiefComplaint}
          chiefComplaintTypeId={chiefComplaintTypeId}
          setChiefComplaintTypeId={setChiefComplaintTypeId}
          templates={INTERVIEW_TEMPLATES}
          onInsertTemplate={handleInsertTemplate}
          isFinalized={isFinalized}
        />

        {/* Middle Column: 治療方針 (Treatment Policy) */}
        <InterviewTreatmentPolicy
          className="col-span-1 lg:col-span-4 h-full"
          treatmentPolicy={treatmentPolicy}
          setTreatmentPolicy={setTreatmentPolicy}
          isFinalized={isFinalized}
        />
      </fieldset>

      {/* Right Column: カルテ履歴 (Medical History) */}
      <InterviewHistory
        className="col-span-1 lg:col-span-5 h-full"
        historyItems={resolvedHistoryItems}
        isLocked={isLocked}
        onCopyItem={handleCopyItem}
      />

      <ConfirmDialog
        open={pendingCopy !== null}
        onClose={handleCloseCopyConfirm}
        onConfirm={handleConfirmCopy}
        title="過去の問診内容をコピーしますか？"
        description="主訴詳細・治療方針・主訴区分の現在の入力が上書きされ、前回の治療明細も追加されます。"
        confirmLabel="コピー"
      />
    </div>
  );
});
