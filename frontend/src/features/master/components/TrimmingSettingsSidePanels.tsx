import type { TrimmingCourse, TrimmingOption } from "../api/trimming";
import type {
  CourseFormData,
  OptionFormData,
} from "./trimming-side-panel-model";
import type { TrimmingTabValue } from "../routes/trimming-settings-model";
import { TrimmingCourseSidePanel } from "./TrimmingCourseSidePanel";
import { TrimmingOptionSidePanel } from "./TrimmingOptionSidePanel";

interface TrimmingSettingsSidePanelsProps {
  activeTab: TrimmingTabValue;
  courseEditTarget: TrimmingCourse | "new" | null;
  coursePanelItem: TrimmingCourse | null;
  optionEditTarget: TrimmingOption | "new" | null;
  optionPanelItem: TrimmingOption | null;
  canDelete: boolean;
  canEdit: boolean;
  onCourseClose: () => void;
  onCourseSave: (data: CourseFormData) => Promise<boolean> | boolean;
  onCourseDeleteRequest: (item: TrimmingCourse) => void;
  onOptionClose: () => void;
  onOptionSave: (data: OptionFormData) => Promise<boolean> | boolean;
  onOptionDeleteRequest: (item: TrimmingOption) => void;
  onDirtyChange: (dirty: boolean) => void;
}

export function TrimmingSettingsSidePanels({
  activeTab,
  courseEditTarget,
  coursePanelItem,
  optionEditTarget,
  optionPanelItem,
  canDelete,
  canEdit,
  onCourseClose,
  onCourseSave,
  onCourseDeleteRequest,
  onOptionClose,
  onOptionSave,
  onOptionDeleteRequest,
  onDirtyChange,
}: TrimmingSettingsSidePanelsProps) {
  return (
    <>
      {activeTab === "course" && courseEditTarget !== null ? (
        <TrimmingCourseSidePanel
          key={coursePanelItem ? String(coursePanelItem.id) : "new-trimming-course"}
          item={coursePanelItem}
          onClose={onCourseClose}
          onSave={onCourseSave}
          onDeleteRequest={canDelete ? onCourseDeleteRequest : undefined}
          readOnly={!canEdit}
          onDirtyChange={onDirtyChange}
        />
      ) : null}
      {activeTab === "option" && optionEditTarget !== null ? (
        <TrimmingOptionSidePanel
          key={optionPanelItem ? String(optionPanelItem.id) : "new-trimming-option"}
          item={optionPanelItem}
          onClose={onOptionClose}
          onSave={onOptionSave}
          onDeleteRequest={canDelete ? onOptionDeleteRequest : undefined}
          readOnly={!canEdit}
          onDirtyChange={onDirtyChange}
        />
      ) : null}
    </>
  );
}
