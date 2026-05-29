import { useCallback } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { Plus, Scissors } from "lucide-react";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { C, ICON } from "@/lib/design-tokens";
import { ResourceMasterTrimming } from "@/types/generated/models";
import { TrimmingDeleteDialogs } from "../components/TrimmingDeleteDialogs";
import { TrimmingSettingsSidePanels } from "../components/TrimmingSettingsSidePanels";
import { TrimmingSettingsTabs } from "../components/TrimmingSettingsTabs";
import type {
  CourseFormData,
  OptionFormData,
} from "../components/TrimmingSidePanelModel";
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
  toTrimmingTabValue,
} from "./TrimmingSettingsModel";

export function TrimmingSettings() {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterTrimming);
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = toTrimmingTabValue(searchParams.get("tab"));

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
            <TrimmingSettingsTabs
              activeTab={activeTab}
              onTabChange={handleTabChange}
              onCourseEditTargetChange={courseCrud.setEditTarget}
              onOptionEditTargetChange={optionCrud.setEditTarget}
              canEdit={canEdit}
            />
          </PageLayout>
        </div>

        <TrimmingSettingsSidePanels
          activeTab={activeTab}
          courseEditTarget={courseCrud.editTarget}
          coursePanelItem={courseCrud.panelItem}
          optionEditTarget={optionCrud.editTarget}
          optionPanelItem={optionCrud.panelItem}
          canDelete={canDelete}
          canEdit={canEdit}
          onCourseClose={courseCrud.handleClose}
          onCourseSave={courseSave.handleSave}
          onCourseDeleteRequest={courseCrud.setPendingDelete}
          onOptionClose={optionCrud.handleClose}
          onOptionSave={optionSave.handleSave}
          onOptionDeleteRequest={optionCrud.setPendingDelete}
          onDirtyChange={handleDirtyChange}
        />
      </div>

      <TrimmingDeleteDialogs
        pendingCourseDelete={courseCrud.pendingDelete}
        pendingOptionDelete={optionCrud.pendingDelete}
        onCourseDeleteCancel={courseCrud.handleDeleteCancel}
        onCourseDeleteConfirm={courseCrud.handleDeleteConfirm}
        onOptionDeleteCancel={optionCrud.handleDeleteCancel}
        onOptionDeleteConfirm={optionCrud.handleDeleteConfirm}
      />
    </>
  );
}
