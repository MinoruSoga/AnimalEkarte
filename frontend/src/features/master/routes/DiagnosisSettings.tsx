import { useCallback } from "react";
import { useNavigate, useSearchParams } from "react-router";
import ClipboardList from "lucide-react/dist/esm/icons/clipboard-list";
import Plus from "lucide-react/dist/esm/icons/plus";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { UnifiedTabs, UnifiedTabsContent } from "@/components/shared/UnifiedTabs";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { C, ICON } from "@/lib/design-tokens";
import { ResourceMasterMedical } from "@/types/generated/models";
import { DiagnosisNameSidePanel } from "../components/DiagnosisNameSidePanel";
import type {
  DiagnosisNameFormData,
  DiagnosisTypeFormData,
} from "../components/DiagnosisSidePanelModel";
import { DiagnosisTypeSidePanel } from "../components/DiagnosisTypeSidePanel";
import { DiagnosisNameTab, DiagnosisTypeTab } from "../components/DiagnosisTabs";
import {
  buildDiagnosisNameCreateRequest,
  buildDiagnosisNameUpdateRequest,
  buildDiagnosisTypeCreateRequest,
  buildDiagnosisTypeUpdateRequest,
} from "./DiagnosisSettingsModel";
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

const TABS = [
  { value: "diagnosis_type", label: "診断病名カテゴリ" },
  { value: "diagnosis_name", label: "診断病名" },
] as const;

export function DiagnosisSettings() {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterMedical);
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = searchParams.get("tab") ?? "diagnosis_type";

  const { data: rawCategories } = useGetDiagnosisTypes();
  const { data: rawNames } = useGetDiagnosisNames();
  const createCategoryMutation = useCreateDiagnosisType();
  const updateCategoryMutation = useUpdateDiagnosisType();
  const deleteCategoryMutation = useDeleteDiagnosisType();
  const createNameMutation = useCreateDiagnosisName();
  const updateNameMutation = useUpdateDiagnosisName();
  const deleteNameMutation = useDeleteDiagnosisName();

  const dirty = useSidePeekDirty();
  const handleDirtyChange = useCallback((isDirty: boolean) => {
    if (isDirty) dirty.markDirty();
    else dirty.markClean();
  }, [dirty]);

  const catCrud = useMasterCRUD<DiagnosisType>({
    data: rawCategories,
    deleteMutation: deleteCategoryMutation,
    entityLabel: "診断カテゴリ",
    dirtyGuard: dirty,
  });

  const nameCrud = useMasterCRUD<DiagnosisName>({
    data: rawNames,
    deleteMutation: deleteNameMutation,
    entityLabel: "診断病名",
    dirtyGuard: dirty,
  });

  const catSetEditTarget = catCrud.setEditTarget;
  const nameSetEditTarget = nameCrud.setEditTarget;

  const handleTabChange = useCallback((tab: string) => {
    if (!dirty.confirmDiscard()) return;
    setSearchParams({ tab });
    catSetEditTarget(null);
    nameSetEditTarget(null);
  }, [setSearchParams, catSetEditTarget, nameSetEditTarget, dirty]);

  const catSave = useMasterSave({
    crud: catCrud,
    createMutation: createCategoryMutation,
    updateMutation: updateCategoryMutation,
    validate: (data: DiagnosisTypeFormData) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildDiagnosisTypeCreateRequest,
    toUpdateRequest: buildDiagnosisTypeUpdateRequest,
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
  });

  return (
    <>
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <PageLayout
            title="診断病名マスタ"
            icon={<ClipboardList className={`${ICON.page} ${C.text}`} />}
            resource={ResourceMasterMedical}
            onBack={() => navigate(paths.settings.getHref())}
            maxWidth="max-w-full"
            headerAction={
              canCreate ? (
                <PrimaryButton onClick={() => {
                  if (activeTab === "diagnosis_type") catCrud.handleNew();
                  else nameCrud.handleNew();
                }}>
                  <Plus className={`mr-1.5 ${ICON.action}`} />
                  新規登録
                </PrimaryButton>
              ) : null
            }
          >
            <UnifiedTabs
              items={TABS}
              value={activeTab}
              onValueChange={handleTabChange}
              className="flex flex-col gap-4"
            >
              <UnifiedTabsContent value="diagnosis_type" className="mt-4">
                <DiagnosisTypeTab
                  onEditTargetChange={catCrud.setEditTarget}
                  canEdit={canEdit}
                />
              </UnifiedTabsContent>
              <UnifiedTabsContent value="diagnosis_name" className="mt-4">
                <DiagnosisNameTab
                  onEditTargetChange={nameCrud.setEditTarget}
                  canEdit={canEdit}
                />
              </UnifiedTabsContent>
            </UnifiedTabs>
          </PageLayout>
        </div>

        {activeTab === "diagnosis_type" && catCrud.isEditing ? (
          <DiagnosisTypeSidePanel
            key={catCrud.panelItem ? String(catCrud.panelItem.id) : "new-category"}
            item={catCrud.panelItem}
            onClose={catCrud.handleClose}
            onSave={catSave.handleSave}
            onDeleteRequest={canDelete ? catCrud.setPendingDelete : undefined}
            readOnly={!canEdit}
            onDirtyChange={handleDirtyChange}
          />
        ) : null}
        {activeTab === "diagnosis_name" && nameCrud.isEditing ? (
          <DiagnosisNameSidePanel
            key={nameCrud.panelItem ? String(nameCrud.panelItem.id) : "new-name"}
            item={nameCrud.panelItem}
            categories={rawCategories ?? []}
            onClose={nameCrud.handleClose}
            onSave={nameSave.handleSave}
            onDeleteRequest={canDelete ? nameCrud.setPendingDelete : undefined}
            readOnly={!canEdit}
            onDirtyChange={handleDirtyChange}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={catCrud.pendingDelete !== null}
        onClose={catCrud.handleDeleteCancel}
        title="診断カテゴリを削除しますか？"
        description={`「${catCrud.pendingDelete?.name}」を削除します。このカテゴリに属する診断名も影響を受けます。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={catCrud.handleDeleteConfirm}
      />
      <ConfirmDialog
        open={nameCrud.pendingDelete !== null}
        onClose={nameCrud.handleDeleteCancel}
        title="診断病名を削除しますか？"
        description={`「${nameCrud.pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={nameCrud.handleDeleteConfirm}
      />
    </>
  );
}
