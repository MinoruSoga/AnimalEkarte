import { useState } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { toast } from "sonner";
import { usePetInfo } from "@/hooks/use-pet";
import { useGetMedicalRecord } from "../api/get-medical-record";
import { useCreateMedicalRecord } from "../api/create-medical-record";
import { useUpdateMedicalRecord } from "../api/update-medical-record";
import type { CreateMedicalRecordRequest, UpdateMedicalRecordRequest } from "../api/types";
import type { TreatmentItem } from "../components/TreatmentTable";

export function useMedicalRecordForm(recordId?: string) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isNewRecord = !recordId;

  const [activeTab, setActiveTab] = useState("問診");
  const [treatmentPlanItems, setTreatmentPlanItems] = useState<TreatmentItem[]>([]);
  const [treatmentCompletedItems, setTreatmentCompletedItems] = useState<TreatmentItem[]>([]);

  // 編集モード: カルテからpetIdを取得
  const { data: existingRecord } = useGetMedicalRecord(recordId ?? "");

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
      };

      try {
        await createMutation.mutateAsync(req);
        toast.success("カルテを作成しました");
        setTimeout(() => {
          if (location.state?.from) {
            navigate(location.state.from);
          } else {
            navigate("/medical-records");
          }
        }, 800);
      } catch {
        toast.error("カルテの作成に失敗しました");
      }
    } else if (recordId) {
      const req: UpdateMedicalRecordRequest = {
        status: "作成中",
      };

      try {
        await updateMutation.mutateAsync({ id: recordId, req });
        toast.success("カルテを更新しました");
        setTimeout(() => {
          if (location.state?.from) {
            navigate(location.state.from);
          } else {
            navigate("/medical-records");
          }
        }, 800);
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
  };
}
