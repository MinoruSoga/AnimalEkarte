import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router";
import { FileText, Trash2, MessageSquare, AlertCircle } from "lucide-react";

import { paths } from "@/config/paths";
import { useGetMasterItems } from "@/hooks/use-master-items";
import { useGetStaffs } from "@/hooks/use-staffs";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { useGetHospitalizations } from "../api/get-hospitalizations";
import {
  selectHospitalizationDoctorStaffs,
  toHospitalizationHistoryItems,
} from "./hospitalization-form-model";

import { Button } from "@/components/ui/button";
import { FormHeaderActions } from "@/components/shared/Form/FormHeaderActions";
import { formatDate } from "@/lib/format/date";
import { PatientInfoCard, formatPatientPetDetails } from "@/components/shared/PatientInfoCard";
import { PastRecordHistoryPanel } from "@/components/shared/PastRecordHistoryPanel";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import {
  MasterSelectModal,
  type MasterSelectItem,
} from "@/components/shared/MasterSelectModal";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { C, STYLE, ICON, LAYOUT } from "@/lib/design-tokens";
import { ResourceHospitalization } from "@/types/generated/models";
import type { MasterItem } from "@/types";
import type { HospitalizationTreatmentPlan } from "@/types";
import type { StaffItem } from "@/hooks/use-staffs";
import { HospitalizationBasicInfo } from "../components/HospitalizationBasicInfo";
import { HospitalizationNoteCard } from "../components/HospitalizationNoteCard";
import { HospitalizationTreatmentTable } from "../components/HospitalizationTreatmentTable";
import { HospitalizationCostSummary } from "../components/HospitalizationCostSummary";
import { H_STYLES } from "../lib/styles";
import type { HospitalizationFormData } from "../types";
import type { HospitalizationFormGate } from "./hospitalization-form-model";

// eslint-disable-next-line react-refresh/only-export-components -- 150行分割で page chrome hook を panels と同居
export function useHospitalizationFormChrome(input: {
  hospitalizationId: string | undefined;
  petId: string | null;
  locationFrom: string | undefined;
  isEdit: boolean;
  formData: HospitalizationFormData;
  formState: { success?: boolean; timestamp?: number; fieldErrors?: Record<string, string> };
  handleFormDataChangeRaw: (updates: Partial<HospitalizationFormData>) => void;
  calculateTotals: () => HospitalizationFormFieldsProps["totals"];
  selectedPet: HospitalizationFormFieldsProps["selectedPet"];
  canDelete: boolean | undefined;
  treatmentPlans: HospitalizationTreatmentPlan[];
}) {
  const navigate = useNavigate();
  const { data: cageItems } = useGetMasterItems("cage");
  const { data: staffs = [] } = useGetStaffs();
  const [staffModalOpen, setStaffModalOpen] = useState(false);
  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  useEffect(() => {
    if (input.formState.success) {
      markClean();
      navigate(input.locationFrom || paths.hospitalization.getHref());
    }
  }, [input.formState.success, input.formState.timestamp, navigate, markClean, input.locationFrom]);

  useEffect(() => {
    if (input.formState.fieldErrors && Object.keys(input.formState.fieldErrors).length > 0) {
      const firstErrorKey = Object.keys(input.formState.fieldErrors)[0];
      const element = document.getElementById(firstErrorKey);
      if (element) {
        element.focus();
        element.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }
  }, [input.formState.fieldErrors, input.formState.timestamp]);

  const totals = input.calculateTotals();
  const historyPetId = input.selectedPet?.id ?? input.petId ?? "";
  const { data: hospitalizationsResult, isLoading: isHistoryLoading } = useGetHospitalizations({
    petId: historyPetId || undefined,
    page: 1,
    limit: 100,
    statusFilter: "all",
  });
  const historyItems = useMemo(() => {
    if (!historyPetId) return [];
    return toHospitalizationHistoryItems(hospitalizationsResult?.data ?? []);
  }, [historyPetId, hospitalizationsResult?.data]);

  const handleBack = useCallback(() => {
    navigate(input.locationFrom || paths.hospitalization.getHref());
  }, [input.locationFrom, navigate]);

  const handleFormDataChangeRaw = input.handleFormDataChangeRaw;
  const handleFormChange = useCallback((updates: Partial<HospitalizationFormData>) => {
    markDirty();
    handleFormDataChangeRaw(updates);
  }, [markDirty, handleFormDataChangeRaw]);

  const doctorStaffItems = useMemo(
    () => selectHospitalizationDoctorStaffs(staffs, input.formData.doctorId),
    [staffs, input.formData.doctorId],
  );
  const handleSelectDoctor = useCallback((item: MasterSelectItem) => {
    handleFormChange({ doctorId: String(item.id), doctorName: item.name });
  }, [handleFormChange]);

  const hasChildTreatmentPlans = input.isEdit && input.treatmentPlans.length > 0;
  const canShowDelete = input.canDelete === true && !hasChildTreatmentPlans;

  useEffect(() => {
    if (!input.selectedPet && !input.isEdit && !input.petId) {
      navigate(paths.hospitalization.selectPet.getHref());
    }
  }, [input.selectedPet, input.isEdit, navigate, input.petId]);

  return {
    cageItems,
    staffModalOpen,
    setStaffModalOpen,
    isDirty,
    totals,
    isHistoryLoading,
    historyItems,
    handleBack,
    handleFormChange,
    doctorStaffItems,
    handleSelectDoctor,
    hasChildTreatmentPlans,
    canShowDelete,
  };
}

export function HospitalizationFormStatusView({
  gate,
  onBack,
}: {
  gate: HospitalizationFormGate | { kind: "new-deceased" };
  onBack: () => void;
}) {
  if (gate.kind === "new-pet-loading") return <LoadingFallback />;
  if (gate.kind === "new-no-pet") return null;

  if (gate.kind === "edit-loading") {
    return (
      <PageLayout
        title="入院"
        onBack={onBack}
        icon={<FileText className={`${ICON.page} ${C.text}`} />}
        resource={ResourceHospitalization}
        maxWidth={LAYOUT.pageContentMaxWidth.form}
      >
        <LoadingFallback />
      </PageLayout>
    );
  }
  if (gate.kind === "edit-not-found") {
    return (
      <PageLayout
        title="入院"
        onBack={onBack}
        icon={<FileText className={`${ICON.page} ${C.text}`} />}
        resource={ResourceHospitalization}
        maxWidth={LAYOUT.pageContentMaxWidth.form}
      >
        <ErrorFallback message="入院情報が見つかりません" />
      </PageLayout>
    );
  }
  if (gate.kind === "new-deceased") {
    return (
      <PageLayout
        title="入院登録"
        onBack={onBack}
        icon={<FileText className={`${ICON.page} ${C.text}`} />}
        resource={ResourceHospitalization}
        maxWidth={LAYOUT.pageContentMaxWidth.form}
      >
        <ErrorFallback message="死亡したペットは入院登録できません" />
      </PageLayout>
    );
  }
  return (
    <PageLayout
      title="入院"
      onBack={onBack}
      icon={<FileText className={`${ICON.page} ${C.text}`} />}
      resource={ResourceHospitalization}
      maxWidth={LAYOUT.pageContentMaxWidth.form}
    >
      <div className="space-y-3">
        <ErrorFallback message="入院情報の取得に失敗しました" />
        {gate.retryRead ? (
          <Button type="button" variant="outline" size="sm" onClick={gate.retryRead}>
            再試行
          </Button>
        ) : null}
      </div>
    </PageLayout>
  );
}

function HospitalizationFormHeaderExtra({
  hospitalizationId,
  canShowDelete,
  onOpenDetail,
  onOpenDeleteConfirm,
}: {
  hospitalizationId: string | undefined;
  canShowDelete: boolean;
  onOpenDetail: () => void;
  onOpenDeleteConfirm: () => void;
}) {
  return (
    <>
      {hospitalizationId ? (
        <Button
          variant="outline"
          type="button"
          className={`gap-2 h-10 text-sm px-4 ${C.text}`}
          onClick={onOpenDetail}
        >
          <FileText className={ICON.action} />
          デイリーカルテ
        </Button>
      ) : null}
      {hospitalizationId && canShowDelete ? (
        <Button
          variant="ghost"
          type="button"
          className={`${STYLE.btnDangerGhost} h-10 text-sm px-4`}
          onClick={onOpenDeleteConfirm}
        >
          <Trash2 className={`mr-1.5 ${ICON.action}`} />
          削除
        </Button>
      ) : null}
    </>
  );
}

interface HospitalizationPatient {
  id?: string;
  ownerName: string;
  name: string;
  petNumber?: string;
  weight?: string;
  species?: string;
  birthDate?: string;
  gender?: string;
  neuteredDate?: string;
  insuranceName?: string;
  insuranceDetails?: string;
  status?: string;
}

interface HospitalizationFormFieldsProps {
  selectedPet: HospitalizationPatient | undefined;
  formData: HospitalizationFormData;
  fieldErrors: Record<string, string> | undefined;
  cageItems: MasterItem[] | undefined;
  isEdit: boolean;
  canDelete: boolean | undefined;
  hasChildTreatmentPlans: boolean;
  treatmentPlans: HospitalizationTreatmentPlan[];
  totals: {
    subtotalBeforeDiscount: number;
    subtotalAfterDiscount: number;
    consumptionTax: number;
    total: number;
  };
  historyItems: { id: string; date: string; title: string; subtitle?: string }[];
  isHistoryLoading: boolean;
  canSubmit: boolean;
  onFormChange: (updates: Partial<HospitalizationFormData>) => void;
  onOpenStaffModal: () => void;
  onAddTreatmentPlan: () => void;
  onUpdateTreatmentPlan: (
    id: string,
    field: keyof HospitalizationTreatmentPlan,
    value: string | number | boolean,
  ) => void;
  onRemoveTreatmentPlan: (id: string) => void;
}

// FE-RC-085: ネスト三項を早期return関数へ分解。
function resolveNextVisitDate(formData: HospitalizationFormData): string | undefined {
  if (formData.nextVisit) return formatDate(formData.nextVisit);
  if (formData.endDate) return formatDate(formData.endDate);
  return undefined;
}

function resolveNextVisitContent(formData: HospitalizationFormData): string | undefined {
  if (formData.nextVisit) return "次回来院";
  if (formData.endDate) return "退院予定";
  return undefined;
}

function HospitalizationFormFields({
  selectedPet,
  formData,
  fieldErrors,
  cageItems,
  isEdit,
  canDelete,
  hasChildTreatmentPlans,
  treatmentPlans,
  totals,
  historyItems,
  isHistoryLoading,
  canSubmit,
  onFormChange,
  onOpenStaffModal,
  onAddTreatmentPlan,
  onUpdateTreatmentPlan,
  onRemoveTreatmentPlan,
}: HospitalizationFormFieldsProps) {
  return (
    <>
        {selectedPet ? (
            <PatientInfoCard
              ownerName={selectedPet.ownerName}
              petName={selectedPet.name}
              petNumber={selectedPet.petNumber || selectedPet.id || ""}
              weight={selectedPet.weight || "-"}
              staffName={formData.doctorName || "未設定"}
              staffLabel="担当医"
              staffButtonId="doctor_id"
              reservationType={formData.hospitalizationType}
              petDetails={formatPatientPetDetails({
                species: selectedPet.species,
                birthDate: selectedPet.birthDate,
                gender: selectedPet.gender,
                neuteredDate: selectedPet.neuteredDate,
              })}
              insuranceName={selectedPet.insuranceName}
              insuranceDetails={selectedPet.insuranceDetails}
              status={selectedPet.status === "死亡" ? "deceased" : "alive"}
              nextVisitDate={resolveNextVisitDate(formData)}
              nextVisitContent={resolveNextVisitContent(formData)}
              onStaffClick={canSubmit ? onOpenStaffModal : undefined}
            />
        ) : null}
        <FormFieldError message={fieldErrors?.pet} />

        <div className="grid grid-cols-1 gap-6 lg:grid-cols-5 mb-3">
          <div className="space-y-3 lg:col-span-3">
          <HospitalizationBasicInfo
            formData={formData}
            onChange={onFormChange}
            cageItems={cageItems ?? []}
            fieldErrors={{ cage_id: fieldErrors?.cage_id }}
          />

          <HospitalizationNoteCard
            id="owner_request"
            title="主訴"
            icon={MessageSquare}
            value={formData.ownerRequest}
            onChange={(val) => onFormChange({ ownerRequest: val })}
            placeholder="主訴を入力..."
          />

          <HospitalizationNoteCard
            id="staff_notes"
            title="スタッフへの連絡事項"
            icon={AlertCircle}
            value={formData.staffNotes}
            onChange={(val) => onFormChange({ staffNotes: val })}
            placeholder="連絡事項を入力..."
          />

        {isEdit ? (
          <p className={`mb-2 ${H_STYLES.text.sm} ${C.text60}`}>
            登録時の治療プランはスナップショットとして参照のみです。この画面では変更・削除できません。入院中の投薬・給餌などは入院詳細のケアプランで管理します。
          </p>
        ) : (
          <p className={`mb-2 ${H_STYLES.text.sm} ${C.text60}`}>
            治療内容・メモが入力された行のみ、入院登録時に治療プラン（登録時スナップショット）として保存されます。空行は保存されません。
          </p>
        )}
        {hasChildTreatmentPlans ? (
          <p className={`mb-2 ${H_STYLES.text.sm} ${C.text60}`} role="status">
            治療プランが紐付いているため、この入院は削除できません。
          </p>
        ) : null}
        <HospitalizationTreatmentTable
            treatmentPlans={treatmentPlans}
            onAdd={onAddTreatmentPlan}
            onUpdate={onUpdateTreatmentPlan}
            onRemove={canDelete ? onRemoveTreatmentPlan : undefined}
            readOnly={isEdit}
        />

        <p className={`mb-2 ${H_STYLES.text.sm} ${C.text60}`}>
          一括割引（%／円）はこの画面では利用できません。表示金額は治療プラン明細に基づく概算です。
        </p>
        <HospitalizationCostSummary totals={totals} />
          </div>
          <PastRecordHistoryPanel
            title="過去の入院履歴"
            searchPlaceholder="種別・ステータスで検索..."
            items={historyItems}
            isLoading={isHistoryLoading}
          />
        </div>
    </>
  );
}

interface HospitalizationFormBodyProps {
  hospitalizationId: string | undefined;
  canSubmit: boolean;
  canShowDelete: boolean;
  isDirty: boolean;
  isDeleteConfirmOpen: boolean;
  isDeletePending: boolean;
  staffModalOpen: boolean;
  doctorStaffItems: StaffItem[];
  formData: HospitalizationFormData;
  formAction: (payload: FormData) => void;
  fields: HospitalizationFormFieldsProps;
  onBack: () => void;
  onOpenDetail: () => void;
  onOpenDeleteConfirm: () => void;
  onCloseDeleteConfirm: () => void;
  onConfirmDelete: () => void;
  onStaffModalOpenChange: (open: boolean) => void;
  onSelectDoctor: (item: MasterSelectItem) => void;
}

export function HospitalizationFormBody({
  hospitalizationId,
  canSubmit,
  canShowDelete,
  isDirty,
  isDeleteConfirmOpen,
  isDeletePending,
  staffModalOpen,
  doctorStaffItems,
  formData,
  formAction,
  fields,
  onBack,
  onOpenDetail,
  onOpenDeleteConfirm,
  onCloseDeleteConfirm,
  onConfirmDelete,
  onStaffModalOpenChange,
  onSelectDoctor,
}: HospitalizationFormBodyProps) {
  return (
    <>
    <form action={formAction} className="h-full">
    <PageLayout
      title={hospitalizationId ? "入院編集" : "入院登録"}
      onBack={onBack}
      icon={<FileText className={`${ICON.page} ${C.text}`} />}
      resource={ResourceHospitalization}
      maxWidth={LAYOUT.pageContentMaxWidth.form}
      headerAction={
        <FormHeaderActions
          onCancel={onBack}
          submitLabel={canSubmit ? (hospitalizationId ? "更新" : "保存") : undefined}
          extra={
            <HospitalizationFormHeaderExtra
              hospitalizationId={hospitalizationId}
              canShowDelete={canShowDelete}
              onOpenDetail={onOpenDetail}
              onOpenDeleteConfirm={onOpenDeleteConfirm}
            />
          }
        />
      }
    >
        <NavigationBlocker when={isDirty} />
        <fieldset disabled={!canSubmit} className="border-0 p-0 m-0 min-w-0">
        <HospitalizationFormFields {...fields} canSubmit={canSubmit} />
        </fieldset>
    </PageLayout>
    </form>
    <ConfirmDialog
        open={isDeleteConfirmOpen}
        onClose={onCloseDeleteConfirm}
        title="入院を削除しますか？"
        description="この操作は取り消せません。"
        confirmLabel="削除"
        variant="destructive"
        onConfirm={onConfirmDelete}
        isPending={isDeletePending}
      />
    <MasterSelectModal
      open={staffModalOpen}
      onOpenChange={onStaffModalOpenChange}
      title="担当医を選択"
      items={doctorStaffItems}
      selectedValue={formData.doctorId}
      matchBy="id"
      searchPlaceholder="担当医を検索..."
      onSelect={onSelectDoctor}
    />
    </>
  );
}
