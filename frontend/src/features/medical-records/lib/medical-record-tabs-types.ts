import type { Pet } from "@/types";
import type { RecommendationReason } from "../constants/recommendation-reason";
import type { InterviewHistoryItem } from "../types";

export interface MedicalRecordTabsAreaProps {
  activeTab: string;
  mountedTabs: Set<string>;
  isNewRecord: boolean;
  recordId: string | undefined;
  selectedPet: Pet;
  chiefComplaint: string;
  chiefComplaintTypeId: number | null;
  treatmentPolicy: string;
  historyItems: InterviewHistoryItem[];
  physicalExam: string;
  plan: string;
  assessment: string;
  diagnosis1CategoryId: number | null;
  diagnosis1NameId: number | null;
  diagnosis2CategoryId: number | null;
  diagnosis2NameId: number | null;
  ownerDiscountRate: number;
  hasLineIntegration: boolean;
  recommendationReason: RecommendationReason | null;
  lstepStatus: "synced" | "not-linked" | "opt-out" | undefined;
  recordStatus: string;
  diagnosis1NameIdError: string | undefined;
  /** 確定済みまたは送信権限なしの編集ロック。読み取り専用パネルはロック境界の外に配置する。 */
  isLocked: boolean;
  /** P2-15: 拠点横断で開いたカルテの子リソース操作用。レコード自身の clinicId */
  recordClinicId?: string;
  /** EMR-356: 履歴取得中（「測定なし」と区別するため別状態として下ろす） */
  historyLoading?: boolean;
  /** EMR-356: 履歴取得エラー（部分データがあっても失敗を「記録なし」として隠さない） */
  historyError?: boolean;
  /** Guarded pagination stopped before all rows were verified. */
  historyIncomplete?: boolean;
  /** Selected chart time; absent means past chronology cannot be established. */
  recordBoundary?: Date;
  /**
   * EMR-356: 選択中カルテの petId。体重パネルは currentRecord.petId ===
   * selectedPet.id が確認できた時のみ書き込み可とする（切替途中の旧カルテ誤書防止）。
   */
  recordPetId?: string;
  onChiefComplaintChange: (value: string) => void;
  onChiefComplaintTypeIdChange: (id: number | null) => void;
  onTreatmentPolicyChange: (value: string) => void;
  onPhysicalExamChange: (value: string) => void;
  onPlanChange: (value: string) => void;
  onAssessmentChange: (value: string) => void;
  onDiagnosis1CategoryIdChange: (id: number | null) => void;
  onDiagnosis1NameIdChange: (id: number | null) => void;
  onDiagnosis2CategoryIdChange: (id: number | null) => void;
  onDiagnosis2NameIdChange: (id: number | null) => void;
  onRecommendationReasonChange: (value: RecommendationReason | null) => void;
  onRegisterEstimateSave: (fn: () => Promise<void>) => void;
  /** EMR-219: 問診履歴コピー適用時に元カルテの治療明細も複写する（recordId は複写元カルテ id） */
  onCopyRecordTreatments?: (sourceRecordId: string) => void;
}
