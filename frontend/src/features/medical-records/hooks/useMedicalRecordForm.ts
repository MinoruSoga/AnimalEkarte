import { useState, useEffect } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { toast } from "sonner";
import { usePetInfo } from "@/hooks/use-pet";
import { useGetMedicalRecord } from "../api/get-medical-record";
import { useCreateMedicalRecord } from "../api/create-medical-record";
import { useUpdateMedicalRecord } from "../api/update-medical-record";
import type { CreateMedicalRecordRequest, UpdateMedicalRecordRequest } from "../api/types";
import type { TreatmentItem } from "../components/TreatmentTable";

const DEFAULT_CHIEF_COMPLAINT = "# どんな症状\n\n# どこが\n\n# いつから\n\n# その他・備考\n\n# フリースペース";
const DEFAULT_TREATMENT_POLICY = "# 治療方針";
const DEFAULT_PLAN = "# 治療方針";
const DEFAULT_ASSESSMENT = "# 診断詳細";

export function useMedicalRecordForm(recordId?: string) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isNewRecord = !recordId;

  const [activeTab, setActiveTab] = useState("問診");
  const [treatmentPlanItems, setTreatmentPlanItems] = useState<TreatmentItem[]>([]);
  const [treatmentCompletedItems, setTreatmentCompletedItems] = useState<TreatmentItem[]>([]);

  // 問診タブの状態
  const [chiefComplaint, setChiefComplaint] = useState(DEFAULT_CHIEF_COMPLAINT);
  const [treatmentPolicy, setTreatmentPolicy] = useState(DEFAULT_TREATMENT_POLICY);

  // 診察/治療プランタブの状態（SOAPS）
  const [plan, setPlan] = useState(DEFAULT_PLAN);
  const [assessment, setAssessment] = useState(DEFAULT_ASSESSMENT);

  // 編集モード: カルテからpetIdを取得
  const { data: existingRecord } = useGetMedicalRecord(recordId ?? "");

  // 既存カルテデータをフォームに反映
  useEffect(() => {
    if (!existingRecord) return;
    if (existingRecord.chiefComplaint) setChiefComplaint(existingRecord.chiefComplaint);
    if (existingRecord.plan) setPlan(existingRecord.plan);
    if (existingRecord.assessment) setAssessment(existingRecord.assessment);
    if (existingRecord.notes) setTreatmentPolicy(existingRecord.notes);
  }, [existingRecord]);

  // petIdを決定: 新規作成時はURLパラメータ、編集時はカルテのpetId
  const resolvedPetId = isNewRecord ? (petId ?? "") : (existingRecord?.petId ?? "");

  // Petデータを取得
  const { pet: selectedPet, isLoading: isPetLoading } = usePetInfo(resolvedPetId);

  const createMutation = useCreateMedicalRecord();
  const updateMutation = useUpdateMedicalRecord();

  const handleBack = () => {
    if (location.state?.from) {
      navigate(location.state.from);
      return;
    }

    if (!recordId) {
      navigate("/medical-records/select-pet");
    } else {
      navigate("/medical-records");
    }
  };

  const handleSave = async () => {
    if (!selectedPet) return;

    const today = new Date().toISOString().split("T")[0];

    if (isNewRecord) {
      const req: CreateMedicalRecordRequest = {
        pet_id: selectedPet.id,
        owner_id: selectedPet.ownerId,
        visit_date: today,
        visit_type: "再診",
        status: "作成中",
        chief_complaint: chiefComplaint !== DEFAULT_CHIEF_COMPLAINT ? chiefComplaint : undefined,
        plan: plan !== DEFAULT_PLAN ? plan : undefined,
        assessment: assessment !== DEFAULT_ASSESSMENT ? assessment : undefined,
        notes: treatmentPolicy !== DEFAULT_TREATMENT_POLICY ? treatmentPolicy : undefined,
      };

      try {
        await createMutation.mutateAsync(req);
        toast.success("カルテを作成しました");
        navigate(location.state?.from ?? "/medical-records");
      } catch {
        toast.error("カルテの作成に失敗しました");
      }
    } else if (recordId) {
      const req: UpdateMedicalRecordRequest = {
        status: "作成中",
        chief_complaint: chiefComplaint,
        plan,
        assessment,
        notes: treatmentPolicy,
      };

      try {
        await updateMutation.mutateAsync({ id: recordId, req });
        toast.success("カルテを更新しました");
        navigate(location.state?.from ?? "/medical-records");
      } catch {
        toast.error("カルテの更新に失敗しました");
      }
    }
  };

  // petIdがなく新規作成の場合はselect-petへリダイレクト（呼び出し元で判定）
  const shouldRedirectToSelectPet = isNewRecord && !petId;

  return {
    isNewRecord,
    activeTab,
    setActiveTab,
    selectedPet: selectedPet ?? null,
    isPetLoading,
    shouldRedirectToSelectPet,
    handleBack,
    handleSave,
    treatmentPlanItems,
    setTreatmentPlanItems,
    treatmentCompletedItems,
    setTreatmentCompletedItems,
    // 問診タブ
    chiefComplaint,
    setChiefComplaint,
    treatmentPolicy,
    setTreatmentPolicy,
    // 診察/治療プランタブ（SOAPS）
    plan,
    setPlan,
    assessment,
    setAssessment,
  };
}
