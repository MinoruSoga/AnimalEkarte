import { UnifiedTabs, UnifiedTabsContent } from "@/components/shared/UnifiedTabs";
import type { DiagnosisName, DiagnosisType } from "../api/diagnosis";
import {
  DIAGNOSIS_TABS,
  type DiagnosisTabValue,
} from "../routes/DiagnosisSettingsModel";
import { DiagnosisNameTab, DiagnosisTypeTab } from "./DiagnosisTabs";

interface DiagnosisSettingsTabsProps {
  activeTab: DiagnosisTabValue;
  onTabChange: (tab: string) => void;
  onTypeEditTargetChange: (value: DiagnosisType | "new" | null) => void;
  onNameEditTargetChange: (value: DiagnosisName | "new" | null) => void;
  canEdit: boolean;
}

export function DiagnosisSettingsTabs({
  activeTab,
  onTabChange,
  onTypeEditTargetChange,
  onNameEditTargetChange,
  canEdit,
}: DiagnosisSettingsTabsProps) {
  return (
    <UnifiedTabs
      items={DIAGNOSIS_TABS}
      value={activeTab}
      onValueChange={onTabChange}
      className="flex flex-col gap-4"
    >
      <UnifiedTabsContent value="diagnosis_type" className="mt-4">
        <DiagnosisTypeTab
          onEditTargetChange={onTypeEditTargetChange}
          canEdit={canEdit}
        />
      </UnifiedTabsContent>
      <UnifiedTabsContent value="diagnosis_name" className="mt-4">
        <DiagnosisNameTab
          onEditTargetChange={onNameEditTargetChange}
          canEdit={canEdit}
        />
      </UnifiedTabsContent>
    </UnifiedTabs>
  );
}
