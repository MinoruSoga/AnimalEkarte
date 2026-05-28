// React/Framework
import { useState, useMemo, useCallback } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { paths } from "@/config/paths";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";

// Shared hooks
import { useMasterSave } from "../hooks/use-master-save";
import type { UseMasterCRUDReturn } from "../hooks/use-master-crud";

// External
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import Plus from "lucide-react/dist/esm/icons/plus";

// Tabs
import { UnifiedTabs, UnifiedTabsContent } from "@/components/shared/UnifiedTabs";

// Internal shared
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { C, ICON } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";
import {
  TreatmentPlanTabContent,
  type TreatmentTabConfig,
} from "../components/TreatmentPlanTabContent";
import {
  TreatmentItemSidePanel,
  type TreatmentFormData,
} from "../components/TreatmentItemSidePanel";

// API hooks
import { useGetAllConsultations, useCreateConsultation, useUpdateConsultation, useReorderConsultations } from "../api/consultations";
import { useGetAllExaminationTypes, useCreateExaminationType, useUpdateExaminationType, useReorderExaminationTypes } from "../api/exam-types-master";
import { useGetAllProcedures, useCreateProcedure, useUpdateProcedure, useReorderProcedures } from "../api/procedures";
import { useGetAllVaccinesMaster, useCreateVaccineMaster, useUpdateVaccineMaster, useReorderVaccinesMaster } from "../api/vaccines-master";
import { useGetAllCheckupTypes, useCreateCheckupType, useUpdateCheckupType, useReorderCheckupTypes } from "../api/checkup-types";
import type { CreateConsultationRequest, UpdateConsultationRequest } from "@/types/treatment";
import type { CreateExaminationTypeRequest as CreateExaminationRequest, UpdateExaminationTypeRequest as UpdateExaminationRequest } from "@/types/treatment";
import type { CreateProcedureRequest, UpdateProcedureRequest } from "@/types/treatment";
import type { CreateVaccineRequest as CreateVaccineMasterRequest, UpdateVaccineRequest as UpdateVaccineMasterRequest } from "@/types/treatment";
import type { CreateCheckupTypeRequest, UpdateCheckupTypeRequest } from "@/types/treatment";

// Types
import type { TreatmentItem } from "@/lib/transforms/treatment";
import { ResourceMasterMedical } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

const TABS = [
  { value: "consultation", label: "診察" },
  { value: "examination", label: "検査" },
  { value: "procedure", label: "処置" },
  { value: "vaccine", label: "予防接種" },
  { value: "checkup", label: "定期健診" },
] as const;

// ─────────────────────────────────────────────────
// TreatmentPlanMaster (main page)
// ─────────────────────────────────────────────────

export function TreatmentPlanMaster() {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterMedical);
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = searchParams.get("tab") ?? "consultation";
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

  // ── Consultations ──────────────────────────────────
  const { data: consultationData } = useGetAllConsultations();
  const createConsultation = useCreateConsultation();
  const updateConsultation = useUpdateConsultation();
  const reorderConsultations = useReorderConsultations();

  // ── Examination Types ──────────────────────────────
  const { data: examinationData } = useGetAllExaminationTypes();
  const createExamination = useCreateExaminationType();
  const updateExamination = useUpdateExaminationType();
  const reorderExaminations = useReorderExaminationTypes();

  // ── Procedures ─────────────────────────────────────
  const { data: procedureData } = useGetAllProcedures();
  const createProcedure = useCreateProcedure();
  const updateProcedure = useUpdateProcedure();
  const reorderProcedures = useReorderProcedures();

  // ── Vaccines ───────────────────────────────────────
  const { data: vaccineData } = useGetAllVaccinesMaster();
  const createVaccine = useCreateVaccineMaster();
  const updateVaccine = useUpdateVaccineMaster();
  const reorderVaccines = useReorderVaccinesMaster();

  // ── Checkup Types ──────────────────────────────────
  const { data: checkupData } = useGetAllCheckupTypes();
  const createCheckup = useCreateCheckupType();
  const updateCheckup = useUpdateCheckupType();
  const reorderCheckups = useReorderCheckupTypes();

  // ── Tab configs (simplified — data & metadata only) ────────────

  const tabConfigs = useMemo<Record<string, Omit<TreatmentTabConfig, 'onCreate' | 'onUpdate' | 'onDelete'>>>(() => ({
    consultation: {
      data: consultationData,
      entityLabel: "診察",
      emptyMessage: "診察が登録されていません",
      searchPlaceholder: "診察名で検索...",
      onReorder: (ids) => reorderConsultations.mutate({ ids }),
    },
    examination: {
      data: examinationData,
      entityLabel: "検査",
      emptyMessage: "検査が登録されていません",
      searchPlaceholder: "検査名で検索...",
      onReorder: (ids) => reorderExaminations.mutate({ ids }),
    },
    procedure: {
      data: procedureData,
      entityLabel: "処置",
      emptyMessage: "処置が登録されていません",
      searchPlaceholder: "処置名で検索...",
      onReorder: (ids) => reorderProcedures.mutate({ ids }),
    },
    vaccine: {
      data: vaccineData,
      entityLabel: "予防接種",
      emptyMessage: "予防接種が登録されていません",
      searchPlaceholder: "予防接種名で検索...",
      onReorder: (ids) => reorderVaccines.mutate({ ids }),
    },
    checkup: {
      data: checkupData,
      entityLabel: "定期健診",
      emptyMessage: "定期健診が登録されていません",
      searchPlaceholder: "定期健診名で検索...",
      onReorder: (ids) => reorderCheckups.mutate({ ids }),
    },
  }), [
    consultationData, reorderConsultations,
    examinationData, reorderExaminations,
    procedureData, reorderProcedures,
    vaccineData, reorderVaccines,
    checkupData, reorderCheckups,
  ]);

  const selectedItem = editTarget !== null && editTarget !== "new" ? editTarget : null;

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

  const tabItems = TABS;

  // Minimal CRUD object for useMasterSave (FR2 only, skip editTarget management)
  // useMasterSave only accesses editTarget, handleClose, startSaveTransition — cast is safe
  const minimalCrud = useMemo(() => ({
    editTarget,
    handleClose,
    startSaveTransition,
  }) as UseMasterCRUDReturn<TreatmentItem>, [editTarget, handleClose, startSaveTransition]);

  // ── FR2: useMasterSave hooks (5 tabs) ──────────────────────
  const consultationSave = useMasterSave<TreatmentItem, TreatmentFormData, CreateConsultationRequest, UpdateConsultationRequest>({
    crud: minimalCrud,
    createMutation: createConsultation,
    updateMutation: updateConsultation,
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: (data) => ({
      name: data.name,
      price: data.price,
      description: data.description || undefined,
      is_active: data.isActive,
      tax_type: data.taxType,
      tax_rate: data.taxRate,
    }),
    toUpdateRequest: (data) => ({
      name: data.name,
      price: data.price,
      description: data.description || undefined,
      is_active: data.isActive,
      tax_type: data.taxType,
      tax_rate: data.taxRate,
    }),
  });

  const examinationSave = useMasterSave<TreatmentItem, TreatmentFormData, CreateExaminationRequest, UpdateExaminationRequest>({
    crud: minimalCrud,
    createMutation: createExamination,
    updateMutation: updateExamination,
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: (data) => ({
      name: data.name,
      price: data.price,
      description: data.description || undefined,
      is_active: data.isActive,
      is_non_insurance: data.isNonInsurance,
    }),
    toUpdateRequest: (data) => ({
      name: data.name,
      price: data.price,
      description: data.description || undefined,
      is_active: data.isActive,
      is_non_insurance: data.isNonInsurance,
    }),
  });

  const procedureSave = useMasterSave<TreatmentItem, TreatmentFormData, CreateProcedureRequest, UpdateProcedureRequest>({
    crud: minimalCrud,
    createMutation: createProcedure,
    updateMutation: updateProcedure,
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: (data) => ({
      name: data.name,
      price: data.price,
      description: data.description || undefined,
      is_active: data.isActive,
      tax_type: data.taxType,
      tax_rate: data.taxRate,
    }),
    toUpdateRequest: (data) => ({
      name: data.name,
      price: data.price,
      description: data.description || undefined,
      is_active: data.isActive,
      tax_type: data.taxType,
      tax_rate: data.taxRate,
    }),
  });

  const vaccineSave = useMasterSave<TreatmentItem, TreatmentFormData, CreateVaccineMasterRequest, UpdateVaccineMasterRequest>({
    crud: minimalCrud,
    createMutation: createVaccine,
    updateMutation: updateVaccine,
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: (data) => ({
      name: data.name,
      price: data.price,
      description: data.description || undefined,
      is_active: data.isActive,
    }),
    toUpdateRequest: (data) => ({
      name: data.name,
      price: data.price,
      description: data.description || undefined,
      is_active: data.isActive,
    }),
  });

  const checkupSave = useMasterSave<TreatmentItem, TreatmentFormData, CreateCheckupTypeRequest, UpdateCheckupTypeRequest>({
    crud: minimalCrud,
    createMutation: createCheckup,
    updateMutation: updateCheckup,
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: (data) => ({
      name: data.name,
      price: data.price,
      description: data.description || undefined,
      is_active: data.isActive,
    }),
    toUpdateRequest: (data) => ({
      name: data.name,
      price: data.price,
      description: data.description || undefined,
      is_active: data.isActive,
    }),
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
    const hook = saveHooksByTab[activeTab as keyof typeof saveHooksByTab];
    if (hook) {
      hook.handleSave(data);
    }
  }, [activeTab, saveHooksByTab]);

  const handleDeleteRequest = useCallback(() => {
    setPendingDelete(selectedItem);
  }, [selectedItem]);

  return (
    <>
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <PageLayout
            title="治療プランマスタ"
            icon={<Stethoscope className={`${ICON.page} ${C.text}`} />}
            resource={ResourceMasterMedical}
            onBack={() => navigate(paths.settings.getHref())}
            maxWidth="max-w-full"
            headerAction={
              canCreate ? (
                <PrimaryButton onClick={() => {
                  if (!dirty.confirmDiscard()) return;
                  setEditTarget("new");
                }}>
                  <Plus className={`mr-1.5 ${ICON.action}`} />
                  新規登録
                </PrimaryButton>
              ) : null
            }
          >
            <UnifiedTabs
              items={tabItems}
              value={activeTab}
              onValueChange={handleTabChange}
              className="flex flex-col gap-4"
            >
              {TABS.map((tab) => {
                const config = tabConfigs[tab.value];
                return (
                  <UnifiedTabsContent key={tab.value} value={tab.value} className="mt-4">
                    <TreatmentPlanTabContent
                      {...config}
                      onEditTargetChange={setEditTargetGuarded}
                      pendingDelete={pendingDelete}
                      onPendingDeleteChange={setPendingDelete}
                      canEdit={canEdit}
                    />
                  </UnifiedTabsContent>
                );
              })}
            </UnifiedTabs>
          </PageLayout>
        </div>

        {editTarget !== null ? (
          <TreatmentItemSidePanel
            key={selectedItem ? String(selectedItem.id) : "new-item"}
            item={selectedItem}
            onClose={handleClose}
            onSave={handleSave}
            onDeleteRequest={canDelete ? handleDeleteRequest : undefined}
            readOnly={!canEdit}
            onDirtyChange={handleDirtyChange}
          />
        ) : null}
      </div>
    </>
  );
}
