import { UnifiedTabs, UnifiedTabsContent } from "@/components/shared/UnifiedTabs";
import type { TrimmingCourse, TrimmingOption } from "../api/trimming";
import {
  TRIMMING_TABS,
  type TrimmingTabValue,
} from "../routes/TrimmingSettingsModel";
import { TrimmingCourseTab, TrimmingOptionTab } from "./TrimmingTabs";

interface TrimmingSettingsTabsProps {
  activeTab: TrimmingTabValue;
  onTabChange: (tab: string) => void;
  onCourseEditTargetChange: (value: TrimmingCourse | "new" | null) => void;
  onOptionEditTargetChange: (value: TrimmingOption | "new" | null) => void;
  canEdit: boolean;
}

export function TrimmingSettingsTabs({
  activeTab,
  onTabChange,
  onCourseEditTargetChange,
  onOptionEditTargetChange,
  canEdit,
}: TrimmingSettingsTabsProps) {
  return (
    <UnifiedTabs
      items={TRIMMING_TABS}
      value={activeTab}
      onValueChange={onTabChange}
      className="flex flex-col gap-4"
    >
      <UnifiedTabsContent value="course" className="mt-4">
        <TrimmingCourseTab
          onEditTargetChange={onCourseEditTargetChange}
          canEdit={canEdit}
        />
      </UnifiedTabsContent>
      <UnifiedTabsContent value="option" className="mt-4">
        <TrimmingOptionTab
          onEditTargetChange={onOptionEditTargetChange}
          canEdit={canEdit}
        />
      </UnifiedTabsContent>
    </UnifiedTabs>
  );
}
