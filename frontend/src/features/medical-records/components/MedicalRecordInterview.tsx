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

const DEFAULT_HISTORY_ITEMS: InterviewHistoryItem[] = [
  {
    id: "1",
    date: "2022/10/10 (月)",
    author: "医者A",
    type: "再診",
    title: "消化器症状",
    content: "嘔吐2回、下痢なし。食欲低下。",
  },
  {
    id: "2",
    date: "2022/10/09 (日)",
    author: "医者B",
    type: "初診",
    title: "定期検診",
    content: "異常なし。体重3.5kg",
  },
  {
    id: "3",
    date: "2022/09/15 (木)",
    author: "医者A",
    type: "ワクチン",
    title: "混合ワクチン",
    content: "5種混合ワクチン接種。副反応なし。",
  },
];

export const MedicalRecordInterview = memo(function MedicalRecordInterview({
  chiefComplaint,
  setChiefComplaint,
  chiefComplaintTypeId,
  setChiefComplaintTypeId,
  treatmentPolicy,
  setTreatmentPolicy,
  historyItems,
  isFinalized = false,
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
    },
    [setChiefComplaint, setChiefComplaintTypeId, setTreatmentPolicy],
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

  const resolvedHistoryItems =
    historyItems && historyItems.length > 0 ? historyItems : DEFAULT_HISTORY_ITEMS;

  return (
    <div className="grid grid-cols-1 lg:grid-cols-12 lg:grid-rows-1 gap-3 flex-1 min-h-0 h-full">
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

      {/* Right Column: カルテ履歴 (Medical History) */}
      <InterviewHistory
        className="col-span-1 lg:col-span-5 h-full"
        historyItems={resolvedHistoryItems}
        onCopyItem={handleCopyItem}
      />

      <ConfirmDialog
        open={pendingCopy !== null}
        onClose={handleCloseCopyConfirm}
        onConfirm={handleConfirmCopy}
        title="過去の問診内容をコピーしますか？"
        description="主訴詳細・治療方針・主訴区分の現在の入力が上書きされます。"
        confirmLabel="コピー"
      />
    </div>
  );
});
