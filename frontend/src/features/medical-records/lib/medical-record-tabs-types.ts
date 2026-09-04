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
  nextVisitDate: string;
  hasLineIntegration: boolean;
  recommendationReason: RecommendationReason | null;
  lstepStatus: "synced" | "not-linked" | "opt-out" | undefined;
  recordStatus: string;
  diagnosis1NameIdError: string | undefined;
  /** P2-15: 拠点横断で開いたカルテの子リソース操作用。レコード自身の clinicId */
  recordClinicId?: string;
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
  onNextVisitDateChange: (value: string) => void;
  onNextVisitDateValidChange: (valid: boolean) => void;
  onRecommendationReasonChange: (value: RecommendationReason | null) => void;
  onRegisterEstimateSave: (fn: () => Promise<void>) => void;
}
