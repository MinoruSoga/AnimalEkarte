import { useCallback } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { Plus, Scissors } from "lucide-react";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { UnifiedTabs, UnifiedTabsContent } from "@/components/shared/UnifiedTabs";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { C, ICON } from "@/lib/design-tokens";
import { ResourceMasterTrimming } from "@/types/generated/models";
import { TrimmingCourseSidePanel } from "../components/TrimmingCourseSidePanel";
import { TrimmingOptionSidePanel } from "../components/TrimmingOptionSidePanel";
import type {
  CourseFormData,
  OptionFormData,
} from "../components/TrimmingSidePanelModel";
import { TrimmingCourseTab, TrimmingOptionTab } from "../components/TrimmingTabs";
import {
  useCreateTrimmingCourse,
  useCreateTrimmingOption,
  useDeleteTrimmingCourse,
  useDeleteTrimmingOption,
  useUpdateTrimmingCourse,
  useUpdateTrimmingOption,
  type TrimmingCourse,
  type TrimmingOption,
} from "../api/trimming";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import {
  buildTrimmingCourseCreateRequest,
  buildTrimmingCourseUpdateRequest,
  buildTrimmingOptionCreateRequest,
  buildTrimmingOptionUpdateRequest,
} from "./TrimmingSettingsModel";

const TABS = [
  { value: "course", label: "コース" },
  { value: "option", label: "オプション" },
] as const;

export function TrimmingSettings() {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterTrimming);
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = searchParams.get("tab") ?? "course";

  const createCourseMutation = useCreateTrimmingCourse();
  const updateCourseMutation = useUpdateTrimmingCourse();
  const deleteCourseMutation = useDeleteTrimmingCourse();
  const createOptionMutation = useCreateTrimmingOption();
  const updateOptionMutation = useUpdateTrimmingOption();
  const deleteOptionMutation = useDeleteTrimmingOption();

  const dirty = useSidePeekDirty();
  const handleDirtyChange = useCallback((isDirty: boolean) => {
    if (isDirty) dirty.markDirty();
    else dirty.markClean();
  }, [dirty]);

  const courseCrud = useMasterCRUD<TrimmingCourse>({
    data: undefined,
    deleteMutation: deleteCourseMutation,
    entityLabel: "トリミングコース",
    dirtyGuard: dirty,
  });

  const optionCrud = useMasterCRUD<TrimmingOption>({
    data: undefined,
    deleteMutation: deleteOptionMutation,
    entityLabel: "トリミングオプション",
    dirtyGuard: dirty,
  });

  const courseSetEditTarget = courseCrud.setEditTarget;
  const courseSetPendingDelete = courseCrud.setPendingDelete;
  const optionSetEditTarget = optionCrud.setEditTarget;
  const optionSetPendingDelete = optionCrud.setPendingDelete;

  const handleTabChange = useCallback((tab: string) => {
    if (!dirty.confirmDiscard()) return;
    setSearchParams({ tab });
    courseSetEditTarget(null);
    optionSetEditTarget(null);
    courseSetPendingDelete(null);
    optionSetPendingDelete(null);
  }, [setSearchParams, courseSetEditTarget, optionSetEditTarget, courseSetPendingDelete, optionSetPendingDelete, dirty]);

  const courseSave = useMasterSave({
    crud: courseCrud,
    createMutation: createCourseMutation,
    updateMutation: updateCourseMutation,
    validate: (data: CourseFormData) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildTrimmingCourseCreateRequest,
    toUpdateRequest: buildTrimmingCourseUpdateRequest,
  });

  const optionSave = useMasterSave({
    crud: optionCrud,
    createMutation: createOptionMutation,
    updateMutation: updateOptionMutation,
    validate: (data: OptionFormData) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildTrimmingOptionCreateRequest,
    toUpdateRequest: buildTrimmingOptionUpdateRequest,
  });

  return (
    <>
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <PageLayout
            title="トリミングマスタ"
            icon={<Scissors className={`${ICON.page} ${C.text}`} />}
            resource={ResourceMasterTrimming}
            onBack={() => navigate(paths.settings.getHref())}
            maxWidth="max-w-full"
            headerAction={
              canCreate ? (
                <PrimaryButton onClick={() => {
                  if (activeTab === "course") courseCrud.handleNew();
                  else optionCrud.handleNew();
                }}>
                  <Plus className={`mr-1.5 ${ICON.action}`} />
                  新規登録
                </PrimaryButton>
              ) : null
            }
          >
            <div className="flex flex-col gap-4">
              <UnifiedTabs
                items={TABS}
                value={activeTab}
                onValueChange={handleTabChange}
                className="flex flex-col gap-4"
              >
                <UnifiedTabsContent value="course" className="mt-4">
                  <TrimmingCourseTab
                    onEditTargetChange={courseCrud.setEditTarget}
                    canEdit={canEdit}
                  />
                </UnifiedTabsContent>
                <UnifiedTabsContent value="option" className="mt-4">
                  <TrimmingOptionTab
                    onEditTargetChange={optionCrud.setEditTarget}
                    canEdit={canEdit}
                  />
                </UnifiedTabsContent>
              </UnifiedTabs>
            </div>
          </PageLayout>
        </div>

        {activeTab === "course" && courseCrud.isEditing === true ? (
          <TrimmingCourseSidePanel
            key={courseCrud.panelItem ? String(courseCrud.panelItem.id) : "new-trimming-course"}
            item={courseCrud.panelItem}
            onClose={courseCrud.handleClose}
            onSave={courseSave.handleSave}
            onDeleteRequest={canDelete ? courseCrud.setPendingDelete : undefined}
            readOnly={!canEdit}
            onDirtyChange={handleDirtyChange}
          />
        ) : null}
        {activeTab === "option" && optionCrud.isEditing === true ? (
          <TrimmingOptionSidePanel
            key={optionCrud.panelItem ? String(optionCrud.panelItem.id) : "new-trimming-option"}
            item={optionCrud.panelItem}
            onClose={optionCrud.handleClose}
            onSave={optionSave.handleSave}
            onDeleteRequest={canDelete ? optionCrud.setPendingDelete : undefined}
            readOnly={!canEdit}
            onDirtyChange={handleDirtyChange}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={courseCrud.pendingDelete !== null}
        onClose={courseCrud.handleDeleteCancel}
        title="トリミングコースを削除しますか？"
        description={`「${courseCrud.pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={courseCrud.handleDeleteConfirm}
      />
      <ConfirmDialog
        open={optionCrud.pendingDelete !== null}
        onClose={optionCrud.handleDeleteCancel}
        title="トリミングオプションを削除しますか？"
        description={`「${optionCrud.pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={optionCrud.handleDeleteConfirm}
      />
    </>
  );
}
