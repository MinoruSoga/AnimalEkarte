// React/Framework
import { useState, useMemo, useCallback, useLayoutEffect, useRef } from "react";
import { useSearchParams } from "react-router";
import { toast } from "sonner";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";

// Shared hooks
import { useMasterSave } from "../hooks/use-master-save";

// External
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";

// Tabs
import { UnifiedTabs, UnifiedTabsContent } from "@/components/shared/UnifiedTabs";

// Internal shared
import { C, ICON } from "@/lib/design-tokens";
import { handleApiError } from "@/lib/handle-api-error";
import { usePermission } from "@/hooks/use-permission";
import { MasterTabPage } from "../components/MasterTabPage";
import { TreatmentPlanDeleteDialog } from "../components/TreatmentPlanDeleteDialog";
import { TreatmentPlanSidePanelHost } from "../components/TreatmentPlanSidePanelHost";
import { TreatmentPlanTabContent } from "../components/TreatmentPlanTabContent";
import type { TreatmentFormData } from "../components/TreatmentItemSidePanel";
import {
  buildCheckupCreateRequest,
  buildCheckupUpdateRequest,
  buildTreatmentTabConfigs,
  buildConsultationCreateRequest,
  buildConsultationUpdateRequest,
  buildExaminationCreateRequest,
  buildExaminationUpdateRequest,
  buildProcedureCreateRequest,
  buildProcedureUpdateRequest,
  buildVaccineCreateRequest,
  buildVaccineUpdateRequest,
  TREATMENT_PLAN_TABS,
  toTreatmentPlanTabValue,
} from "./treatment-plan-master-model";

// API hooks
import { useGetAllConsultations, useCreateConsultation, useUpdateConsultation, useDeleteConsultation, useReorderConsultations } from "../api/consultations";
import { useGetAllExaminationTypes, useCreateExaminationType, useUpdateExaminationType, useDeleteExaminationType, useReorderExaminationTypes } from "../api/exam-types-master";
import { useGetAllProcedures, useCreateProcedure, useUpdateProcedure, useDeleteProcedure, useReorderProcedures } from "../api/procedures";
import { useGetAllVaccinesMaster, useCreateVaccineMaster, useUpdateVaccineMaster, useDeleteVaccineMaster, useReorderVaccinesMaster } from "../api/vaccines-master";
import { useGetAllCheckupTypes, useCreateCheckupType, useUpdateCheckupType, useDeleteCheckupType, useReorderCheckupTypes } from "../api/checkup-types";
import type { CreateConsultationRequest, UpdateConsultationRequest } from "@/types/treatment";
import type { CreateExaminationTypeRequest as CreateExaminationRequest, UpdateExaminationTypeRequest as UpdateExaminationRequest } from "@/types/treatment";
import type { CreateProcedureRequest, UpdateProcedureRequest } from "@/types/treatment";
import type { CreateVaccineRequest as CreateVaccineMasterRequest, UpdateVaccineRequest as UpdateVaccineMasterRequest } from "@/types/treatment";
import type { CreateCheckupTypeRequest, UpdateCheckupTypeRequest } from "@/types/treatment";

// Types
import type { TreatmentItem } from "@/lib/transforms/treatment";
import { ResourceCheckups, ResourceMasterMedical } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// TreatmentPlanMaster (main page)
// ─────────────────────────────────────────────────

export function TreatmentPlanMaster() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterMedical);
  const {
    canCreate: canCreateCheckup,
    canEdit: canEditCheckup,
    canDelete: canDeleteCheckup,
  } = usePermission(ResourceCheckups);
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = toTreatmentPlanTabValue(searchParams.get("tab"));
  const activeResource = activeTab === "checkup" ? ResourceCheckups : ResourceMasterMedical;
  const activeCanEdit = activeTab === "checkup" ? canEditCheckup : canEdit;
  const activeCanCreate = activeTab === "checkup" ? canCreateCheckup : canCreate;
  const activeCanDelete = activeTab === "checkup" ? canDeleteCheckup : canDelete;
  const permissionsRef = useRef({ canDelete: activeCanDelete === true });
  useLayoutEffect(() => {
    permissionsRef.current = { canDelete: activeCanDelete === true };
  }, [activeCanDelete]);
  const [editTarget, setEditTarget] = useState<TreatmentItem | "new" | null>(null);
  const [pendingDelete, setPendingDelete] = useState<TreatmentItem | null>(null);

  // BUG-380: タブ間共有の未保存ガード
  const dirty = useSidePeekDirty();
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  const handleTabChange = useCallback((tab: string) => {
    if (!dirty.confirmDiscard()) return;
    setSearchParams({ tab });
    setEditTarget(null);
    setPendingDelete(null);
  }, [setSearchParams, dirty]);

  const handleNew = useCallback(() => {
    if (!dirty.confirmDiscard()) return;
    setEditTarget("new");
  }, [dirty]);

  // ── Consultations ──────────────────────────────────
  const { data: consultationData } = useGetAllConsultations();
  const createConsultation = useCreateConsultation();
  const updateConsultation = useUpdateConsultation();
  const deleteConsultation = useDeleteConsultation();
  const reorderConsultations = useReorderConsultations();

  // ── Examination Types ──────────────────────────────
  const { data: examinationData } = useGetAllExaminationTypes();
  const createExamination = useCreateExaminationType();
  const updateExamination = useUpdateExaminationType();
  const deleteExamination = useDeleteExaminationType();
  const reorderExaminations = useReorderExaminationTypes();

  // ── Procedures ─────────────────────────────────────
  const { data: procedureData } = useGetAllProcedures();
  const createProcedure = useCreateProcedure();
  const updateProcedure = useUpdateProcedure();
  const deleteProcedure = useDeleteProcedure();
  const reorderProcedures = useReorderProcedures();

  // ── Vaccines ───────────────────────────────────────
  const { data: vaccineData } = useGetAllVaccinesMaster();
  const createVaccine = useCreateVaccineMaster();
  const updateVaccine = useUpdateVaccineMaster();
  const deleteVaccine = useDeleteVaccineMaster();
  const reorderVaccines = useReorderVaccinesMaster();

  // ── Checkup Types ──────────────────────────────────
  const { data: checkupData } = useGetAllCheckupTypes();
  const createCheckup = useCreateCheckupType();
  const updateCheckup = useUpdateCheckupType();
  const deleteCheckup = useDeleteCheckupType();
  const reorderCheckups = useReorderCheckupTypes();

  // ── Tab configs (simplified — data & metadata only) ────────────

  const tabConfigs = useMemo(() => buildTreatmentTabConfigs({
    consultationData,
    examinationData,
    procedureData,
    vaccineData,
    checkupData,
    onReorderConsultations: (ids) => {
      if (!canEdit) return;
      reorderConsultations.mutate({ ids });
    },
    onReorderExaminations: (ids) => {
      if (!canEdit) return;
      reorderExaminations.mutate({ ids });
    },
    onReorderProcedures: (ids) => {
      if (!canEdit) return;
      reorderProcedures.mutate({ ids });
    },
    onReorderVaccines: (ids) => {
      if (!canEdit) return;
      reorderVaccines.mutate({ ids });
    },
    onReorderCheckups: (ids) => {
      if (!canEditCheckup) return;
      reorderCheckups.mutate({ ids });
    },
  }), [
    canEdit, canEditCheckup,
    consultationData, reorderConsultations,
    examinationData, reorderExaminations,
    procedureData, reorderProcedures,
    vaccineData, reorderVaccines,
    checkupData, reorderCheckups,
  ]);

  const selectedItem = editTarget !== null && editTarget !== "new" ? editTarget : null;
  const selectedExamination = activeTab === "examination" && selectedItem !== null
    ? examinationData?.find((item) => item.id === selectedItem.id)
    : undefined;

  // 現在タブの data (stable reference - undefined を [] に正規化)
  const activeTabData = useMemo(
    () => tabConfigs[activeTab].data ?? [],
    [tabConfigs, activeTab],
  );

  // 親カテゴリ候補: 現在タブ内の root 行 (parentId なし) のうち isActive, かつ自分自身を除く
  const parentCandidates = useMemo(
    () =>
      activeTabData.filter(
        (item) => !item.parentId && item.isActive && item.id !== selectedItem?.id,
      ),
    [activeTabData, selectedItem?.id],
  );

  // 子を持つ root かどうか (true → parentId セレクタ非表示)
  const selectedItemId = selectedItem?.id;
  const hasChildren = useMemo(
    () =>
      selectedItemId != null &&
      activeTabData.some((item) => item.parentId === selectedItemId),
    [activeTabData, selectedItemId],
  );

  const startSaveTransition = useCallback((cb: () => void) => {
    cb();
  }, []);

  const handleClose = useCallback(() => {
    if (!dirty.confirmDiscard()) return;
    setEditTarget(null);
  }, [dirty]);

  // BUG-380: 子コンポーネント (TreatmentTabContent) が行クリック時に呼ぶ setEditTarget をガード
  const setEditTargetGuarded = useCallback((target: TreatmentItem | "new" | null) => {
    if (!dirty.confirmDiscard()) return;
    setEditTarget(target);
  }, [dirty]);

  const tabItems = TREATMENT_PLAN_TABS;

  const minimalCrud = useMemo(() => ({
    editTarget,
    setEditTarget,
    startSaveTransition,
  }), [editTarget, startSaveTransition]);

  // ── FR2: useMasterSave hooks (5 tabs) ──────────────────────
  const consultationSave = useMasterSave<TreatmentItem, TreatmentFormData, CreateConsultationRequest, UpdateConsultationRequest>({
    crud: minimalCrud,
    createMutation: createConsultation,
    updateMutation: updateConsultation,
    permissions: { canCreate, canEdit },
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildConsultationCreateRequest,
    toUpdateRequest: buildConsultationUpdateRequest,
  });

  const examinationSave = useMasterSave<TreatmentItem, TreatmentFormData, CreateExaminationRequest, UpdateExaminationRequest>({
    crud: minimalCrud,
    createMutation: createExamination,
    updateMutation: updateExamination,
    permissions: { canCreate, canEdit },
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildExaminationCreateRequest,
    toUpdateRequest: buildExaminationUpdateRequest,
  });

  const procedureSave = useMasterSave<TreatmentItem, TreatmentFormData, CreateProcedureRequest, UpdateProcedureRequest>({
    crud: minimalCrud,
    createMutation: createProcedure,
    updateMutation: updateProcedure,
    permissions: { canCreate, canEdit },
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildProcedureCreateRequest,
    toUpdateRequest: buildProcedureUpdateRequest,
  });

  const vaccineSave = useMasterSave<TreatmentItem, TreatmentFormData, CreateVaccineMasterRequest, UpdateVaccineMasterRequest>({
    crud: minimalCrud,
    createMutation: createVaccine,
    updateMutation: updateVaccine,
    permissions: { canCreate, canEdit },
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildVaccineCreateRequest,
    toUpdateRequest: buildVaccineUpdateRequest,
  });

  const checkupSave = useMasterSave<TreatmentItem, TreatmentFormData, CreateCheckupTypeRequest, UpdateCheckupTypeRequest>({
    crud: minimalCrud,
    createMutation: createCheckup,
    updateMutation: updateCheckup,
    permissions: { canCreate: canCreateCheckup, canEdit: canEditCheckup },
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildCheckupCreateRequest,
    toUpdateRequest: buildCheckupUpdateRequest,
  });

  // Map tab values to save hooks
  const saveHooksByTab = useMemo(() => ({
    consultation: consultationSave,
    examination: examinationSave,
    procedure: procedureSave,
    vaccine: vaccineSave,
    checkup: checkupSave,
  }), [consultationSave, examinationSave, procedureSave, vaccineSave, checkupSave]);

  const handleSave = useCallback((data: TreatmentFormData) => {
    saveHooksByTab[activeTab].handleSave(data);
  }, [activeTab, saveHooksByTab]);

  const handleDeleteRequest = useCallback(() => {
    setPendingDelete(selectedItem);
  }, [selectedItem]);

  const handleDeleteCancel = useCallback(() => {
    setPendingDelete(null);
  }, []);

  const deleteMutationByTab = useMemo(() => ({
    consultation: deleteConsultation,
    examination: deleteExamination,
    procedure: deleteProcedure,
    vaccine: deleteVaccine,
    checkup: deleteCheckup,
  }), [deleteConsultation, deleteExamination, deleteProcedure, deleteVaccine, deleteCheckup]);

  const handleDeleteConfirm = useCallback(() => {
    if (!pendingDelete) return;
    const config = tabConfigs[activeTab];
    if (permissionsRef.current.canDelete !== true) return;
    deleteMutationByTab[activeTab].mutate(pendingDelete.id, {
      onSuccess: () => {
        setPendingDelete(null);
        setEditTarget(null);
        toast.success(`${config.entityLabel}を削除しました`);
      },
      onError: (error) => handleApiError(error, `${config.entityLabel}の削除`),
    });
  }, [activeTab, deleteMutationByTab, pendingDelete, tabConfigs]);

  return (
    <MasterTabPage
      title="診療項目マスタ"
      icon={<Stethoscope className={`${ICON.page} ${C.text}`} />}
      resource={activeResource}
      onNew={handleNew}
      sidePanel={
        <TreatmentPlanSidePanelHost
          editTarget={editTarget}
          selectedItem={selectedItem}
          parentCandidates={parentCandidates}
          hasChildren={hasChildren}
          canDelete={activeCanDelete}
          canCreate={activeCanCreate}
          canEdit={activeCanEdit}
          examinationType={selectedExamination}
          showAnesthesia={activeTab === "procedure"}
          onClose={handleClose}
          onSave={handleSave}
          onDeleteRequest={handleDeleteRequest}
          onDirtyChange={handleDirtyChange}
        />
      }
      deleteDialogs={
        <TreatmentPlanDeleteDialog
          entityLabel={tabConfigs[activeTab].entityLabel}
          pendingDelete={pendingDelete}
          onClose={handleDeleteCancel}
          onConfirm={handleDeleteConfirm}
        />
      }
    >
      <UnifiedTabs
        items={tabItems}
        value={activeTab}
        onValueChange={handleTabChange}
        className="flex flex-col gap-4"
      >
        {TREATMENT_PLAN_TABS.map((tab) => {
          const config = tabConfigs[tab.value];
          return (
            <UnifiedTabsContent key={tab.value} value={tab.value} className="mt-4">
              <TreatmentPlanTabContent
                {...config}
                onEditTargetChange={setEditTargetGuarded}
                canEdit={tab.value === "checkup" ? canEditCheckup : canEdit}
              />
            </UnifiedTabsContent>
          );
        })}
      </UnifiedTabs>
    </MasterTabPage>
  );
}
