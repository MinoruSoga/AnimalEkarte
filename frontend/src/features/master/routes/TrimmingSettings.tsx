import { useCallback } from "react";
import { useSearchParams } from "react-router";
import { Scissors } from "lucide-react";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { C, ICON } from "@/lib/design-tokens";
import { ResourceMasterTrimming } from "@/types/generated/models";
import { UnifiedTabs, UnifiedTabsContent } from "@/components/shared/UnifiedTabs";
import { TrimmingDeleteDialogs } from "../components/TrimmingDeleteDialogs";
import { TrimmingSettingsSidePanels } from "../components/TrimmingSettingsSidePanels";
import { TrimmingCourseTab, TrimmingOptionTab } from "../components/TrimmingTabs";
import { MasterTabPage } from "../components/MasterTabPage";
import type {
  CourseFormData,
  OptionFormData,
} from "../components/trimming-side-panel-model";
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
  TRIMMING_TABS,
  toTrimmingTabValue,
} from "./trimming-settings-model";

export function TrimmingSettings() {
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
  }, [dirty.markDirty, dirty.markClean]);

  const courseCrud = useMasterCRUD<TrimmingCourse>({
    data: undefined,
    deleteMutation: deleteCourseMutation,
    entityLabel: "トリミングコース",
    dirtyGuard: dirty,
    permissions: { canDelete },
  });

  const optionCrud = useMasterCRUD<TrimmingOption>({
    data: undefined,
    deleteMutation: deleteOptionMutation,
    entityLabel: "トリミングオプション",
    dirtyGuard: dirty,
    permissions: { canDelete },
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

  const handleNew = useCallback(() => {
    if (activeTab === "course") courseCrud.handleNew();
    else optionCrud.handleNew();
  }, [activeTab, courseCrud, optionCrud]);

  const courseSave = useMasterSave({
    crud: courseCrud,
    createMutation: createCourseMutation,
    updateMutation: updateCourseMutation,
    validate: (data: CourseFormData) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildTrimmingCourseCreateRequest,
    toUpdateRequest: buildTrimmingCourseUpdateRequest,
    permissions: { canCreate, canEdit },
  });

  const optionSave = useMasterSave({
    crud: optionCrud,
    createMutation: createOptionMutation,
    updateMutation: updateOptionMutation,
    validate: (data: OptionFormData) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildTrimmingOptionCreateRequest,
    toUpdateRequest: buildTrimmingOptionUpdateRequest,
    permissions: { canCreate, canEdit },
  });

  const handleCourseSave = useCallback(async (data: CourseFormData) => {
    const ok = await courseSave.handleSave(data);
    if (ok) dirty.markClean();
    return ok;
  }, [courseSave.handleSave, dirty.markClean]);

  const handleOptionSave = useCallback(async (data: OptionFormData) => {
    const ok = await optionSave.handleSave(data);
    if (ok) dirty.markClean();
    return ok;
  }, [optionSave.handleSave, dirty.markClean]);

  return (
    <MasterTabPage
      title="トリミングマスタ"
      icon={<Scissors className={`${ICON.page} ${C.text}`} />}
      resource={ResourceMasterTrimming}
      onNew={handleNew}
      sidePanel={
        <TrimmingSettingsSidePanels
          activeTab={activeTab}
          courseEditTarget={courseCrud.editTarget}
          coursePanelItem={courseCrud.panelItem}
          optionEditTarget={optionCrud.editTarget}
          optionPanelItem={optionCrud.panelItem}
          canDelete={canDelete}
          canEdit={canEdit}
          onCourseClose={courseCrud.handleClose}
          onCourseSave={handleCourseSave}
          onCourseDeleteRequest={courseCrud.setPendingDelete}
          onOptionClose={optionCrud.handleClose}
          onOptionSave={handleOptionSave}
          onOptionDeleteRequest={optionCrud.setPendingDelete}
          onDirtyChange={handleDirtyChange}
        />
      }
      deleteDialogs={
        <TrimmingDeleteDialogs
          pendingCourseDelete={courseCrud.pendingDelete}
          pendingOptionDelete={optionCrud.pendingDelete}
          onCourseDeleteCancel={courseCrud.handleDeleteCancel}
          onCourseDeleteConfirm={courseCrud.handleDeleteConfirm}
          onOptionDeleteCancel={optionCrud.handleDeleteCancel}
          onOptionDeleteConfirm={optionCrud.handleDeleteConfirm}
        />
      }
    >
      <UnifiedTabs
        items={TRIMMING_TABS}
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
    </MasterTabPage>
  );
}
