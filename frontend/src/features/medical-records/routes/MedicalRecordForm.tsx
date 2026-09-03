import {
  memo,
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
} from "react";
import { useParams, useNavigate } from "react-router";

import { paths } from "@/config/paths";
import { handleApiError } from "@/lib/handle-api-error";
import { toast } from "sonner";
import { usePermission } from "@/hooks/use-permission";
import { useTitle } from "@/hooks/use-title";
import { useMedicalRecordForm } from "../hooks/use-medical-record-form";
import { useGetMedicalRecord } from "../api/get-medical-record";
import { useDeleteMedicalRecord } from "../api/delete-medical-record";
import { resolveMedicalRecordFormGate } from "./medical-record-form-model";
import { MedicalRecordFormStatusView } from "./MedicalRecordFormStatusPanels";
import { MedicalRecordFormReadyPage } from "./MedicalRecordFormReadyPanels";

export const MedicalRecordForm = memo(function MedicalRecordForm() {
  const { id: recordId } = useParams();
  const navigate = useNavigate();
  const form = useMedicalRecordForm(recordId);
  useTitle(recordId ? `カルテ編集 (#${recordId})` : "カルテ入力");

  const { canEdit, canCreate, canDelete } = usePermission("medical-records");
  const canSubmit = form.isNewRecord ? canCreate : canEdit;
  const canDeleteRef = useRef(canDelete);
  const selectedPetStatusRef = useRef(form.selectedPet?.status);
  useLayoutEffect(() => {
    canDeleteRef.current = canDelete;
  }, [canDelete]);
  useLayoutEffect(() => {
    selectedPetStatusRef.current = form.selectedPet?.status;
  }, [form.selectedPet?.status]);

  const { data: currentRecord } = useGetMedicalRecord(recordId ?? "");
  const recordClinicId = currentRecord?.clinicId;
  const { mutate: deleteRecord, isPending: isDeleting } = useDeleteMedicalRecord(recordClinicId);

  const handleDeleteConfirm = useCallback(() => {
    if (
      !recordId
      || canDeleteRef.current !== true
      || selectedPetStatusRef.current === "死亡"
    ) return;
    deleteRecord(recordId, {
      onSuccess: () => {
        toast.success("カルテを削除しました");
        navigate(paths.medicalRecords.getHref());
      },
      onError: (error) => {
        handleApiError(error, "カルテ削除");
      },
    });
  }, [recordId, deleteRecord, navigate]);

  useEffect(() => {
    if (form.shouldRedirectToSelectPet) {
      navigate(paths.medicalRecords.selectPet.getHref());
    }
  }, [form.shouldRedirectToSelectPet, navigate]);

  const gate = resolveMedicalRecordFormGate({
    isReadLoading: form.isReadLoading,
    notFound: form.notFound,
    isReadNotFound: form.isReadNotFound,
    isReadError: form.isReadError,
    retryRead: form.retryRead,
    isPetLoading: form.isPetLoading,
    isNewRecord: form.isNewRecord,
    hasSelectedPet: Boolean(form.selectedPet),
  });
  if (gate) {
    return <MedicalRecordFormStatusView gate={gate} onBack={form.handleBack} />;
  }

  const selectedPet = form.selectedPet;
  if (!selectedPet) return null;

  // BUG-002: 死亡ペットへの /medical-records/new?petId=… 直叩きは編集フォームを出さない。
  if (form.isNewRecord && selectedPet.status === "死亡") {
    return <MedicalRecordFormStatusView gate={{ kind: "deceased-new" }} onBack={form.handleBack} />;
  }

  return (
    <MedicalRecordFormReadyPage
      recordId={recordId}
      selectedPet={selectedPet}
      form={form}
      canEdit={canEdit}
      canSubmit={canSubmit === true}
      canDelete={canDelete}
      isDeleting={isDeleting}
      onDeleteConfirm={handleDeleteConfirm}
    />
  );
});
