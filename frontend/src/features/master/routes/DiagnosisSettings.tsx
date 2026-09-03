import { useCallback } from "react";
import { useSearchParams } from "react-router";
import ClipboardList from "lucide-react/dist/esm/icons/clipboard-list";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { C, ICON } from "@/lib/design-tokens";
import { ResourceMasterMedical } from "@/types/generated/models";
import { UnifiedTabs, UnifiedTabsContent } from "@/components/shared/UnifiedTabs";
import type {
  DiagnosisNameFormData,
  DiagnosisTypeFormData,
} from "../lib/diagnosis-side-panel-model";
import { DiagnosisDeleteDialogs } from "../components/DiagnosisDeleteDialogs";
import { DiagnosisSettingsSidePanels } from "../components/DiagnosisSettingsSidePanels";
import { DiagnosisNameTab, DiagnosisTypeTab } from "../components/DiagnosisTabs";
import { MasterTabPage } from "../components/MasterTabPage";
import {
  buildDiagnosisNameCreateRequest,
  buildDiagnosisNameUpdateRequest,
  buildDiagnosisTypeCreateRequest,
  buildDiagnosisTypeUpdateRequest,
  DIAGNOSIS_TABS,
  toDiagnosisTabValue,
} from "./diagnosis-settings-model";
import {
  useCreateDiagnosisName,
  useCreateDiagnosisType,
  useDeleteDiagnosisName,
  useDeleteDiagnosisType,
  useGetDiagnosisNames,
  useGetDiagnosisTypes,
  useUpdateDiagnosisName,
  useUpdateDiagnosisType,
  type DiagnosisName,
  type DiagnosisType,
} from "../api/diagnosis";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";

export function DiagnosisSettings() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterMedical);
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = toDiagnosisTabValue(searchParams.get("tab"));

  const { data: rawCategories } = useGetDiagnosisTypes();
  const { data: rawNames } = useGetDiagnosisNames();
  const createCategoryMutation = useCreateDiagnosisType();
  const updateCategoryMutation = useUpdateDiagnosisType();
  const deleteCategoryMutation = useDeleteDiagnosisType();
  const createNameMutation = useCreateDiagnosisName();
  const updateNameMutation = useUpdateDiagnosisName();
  const deleteNameMutation = useDeleteDiagnosisName();

  const dirty = useSidePeekDirty();
  const handleDirtyChange = useCallback(
    (isDirty: boolean) => {
      if (isDirty) dirty.markDirty();
      else dirty.markClean();
    },
    [dirty],
  );

  const catCrud = useMasterCRUD<DiagnosisType>({
    data: rawCategories,
    deleteMutation: deleteCategoryMutation,
    entityLabel: "診断カテゴリ",
    dirtyGuard: dirty,
    permissions: { canDelete },
  });

  const nameCrud = useMasterCRUD<DiagnosisName>({
    data: rawNames,
    deleteMutation: deleteNameMutation,
    entityLabel: "診断病名",
    dirtyGuard: dirty,
    permissions: { canDelete },
  });

  const catSetEditTarget = catCrud.setEditTarget;
  const nameSetEditTarget = nameCrud.setEditTarget;

  const handleTabChange = useCallback(
    (tab: string) => {
      dirty.runWithDiscardCheck(() => {
        setSearchParams({ tab });
        catSetEditTarget(null);
        nameSetEditTarget(null);
      });
    },
    [setSearchParams, catSetEditTarget, nameSetEditTarget, dirty],
  );

  const handleNew = useCallback(() => {
    if (activeTab === "diagnosis_type") catCrud.handleNew();
    else nameCrud.handleNew();
  }, [activeTab, catCrud, nameCrud]);

  const catSave = useMasterSave({
    crud: catCrud,
    createMutation: createCategoryMutation,
    updateMutation: updateCategoryMutation,
    validate: (data: DiagnosisTypeFormData) => (data.name.trim() ? null : "名称を入力してください"),
    toCreateRequest: buildDiagnosisTypeCreateRequest,
    toUpdateRequest: buildDiagnosisTypeUpdateRequest,
    permissions: { canCreate, canEdit },
  });

  const nameSave = useMasterSave({
    crud: nameCrud,
    createMutation: createNameMutation,
    updateMutation: updateNameMutation,
    validate: (data: DiagnosisNameFormData) => {
      if (!data.name.trim()) return "診断病名を入力してください";
      if (!data.diagnosisTypeId) return "カテゴリを選択してください";
      return null;
    },
    toCreateRequest: buildDiagnosisNameCreateRequest,
    toUpdateRequest: buildDiagnosisNameUpdateRequest,
    permissions: { canCreate, canEdit },
  });

  return (
    <>
      <MasterTabPage
        title="診断マスタ"
        icon={<ClipboardList className={`${ICON.page} ${C.text}`} />}
        resource={ResourceMasterMedical}
        onNew={handleNew}
        sidePanel={
          <DiagnosisSettingsSidePanels
            activeTab={activeTab}
            typeEditTarget={catCrud.editTarget}
            typePanelItem={catCrud.panelItem}
            nameEditTarget={nameCrud.editTarget}
            namePanelItem={nameCrud.panelItem}
            categories={rawCategories ?? []}
            canDelete={canDelete}
            canEdit={canEdit}
            onTypeClose={catCrud.handleClose}
            onTypeSave={catSave.handleSave}
            onTypeDeleteRequest={catCrud.setPendingDelete}
            onNameClose={nameCrud.handleClose}
            onNameSave={nameSave.handleSave}
            onNameDeleteRequest={nameCrud.setPendingDelete}
            onDirtyChange={handleDirtyChange}
          />
        }
        deleteDialogs={
          <DiagnosisDeleteDialogs
            pendingTypeDelete={catCrud.pendingDelete}
            pendingNameDelete={nameCrud.pendingDelete}
            onTypeDeleteCancel={catCrud.handleDeleteCancel}
            onTypeDeleteConfirm={catCrud.handleDeleteConfirm}
            onNameDeleteCancel={nameCrud.handleDeleteCancel}
            onNameDeleteConfirm={nameCrud.handleDeleteConfirm}
          />
        }
      >
        <UnifiedTabs
          items={DIAGNOSIS_TABS}
          value={activeTab}
          onValueChange={handleTabChange}
          className="flex flex-col gap-4"
        >
          <UnifiedTabsContent value="diagnosis_type" className="mt-4">
            <DiagnosisTypeTab onEditTargetChange={catCrud.setEditTarget} canEdit={canEdit} />
          </UnifiedTabsContent>
          <UnifiedTabsContent value="diagnosis_name" className="mt-4">
            <DiagnosisNameTab onEditTargetChange={nameCrud.setEditTarget} canEdit={canEdit} />
          </UnifiedTabsContent>
        </UnifiedTabs>
      </MasterTabPage>
      {dirty.discardDialog}
    </>
  );
}
