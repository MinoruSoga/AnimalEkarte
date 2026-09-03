import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import { UnifiedTabs, UnifiedTabsContent } from "@/components/shared/UnifiedTabs";
import { C, ICON } from "@/lib/design-tokens";
import { MasterTabPage } from "../components/MasterTabPage";
import { TreatmentPlanDeleteDialog } from "../components/TreatmentPlanDeleteDialog";
import { TreatmentPlanSidePanelHost } from "../components/TreatmentPlanSidePanelHost";
import { TreatmentPlanTabContent } from "../components/TreatmentPlanTabContent";
import { TREATMENT_PLAN_TABS, type TreatmentPlanTabValue } from "./treatment-plan-master-model";
import type { TreatmentItem } from "@/lib/transforms/treatment";
import type { Resource } from "@/types/generated/models";
import type { useTreatmentPlanMasterResources } from "../hooks/use-treatment-plan-master-resources";
import type { TreatmentFormData } from "../components/TreatmentItemSidePanel";

type TreatmentPlanResources = ReturnType<typeof useTreatmentPlanMasterResources>;

interface TreatmentPlanMasterViewProps {
  activeTab: TreatmentPlanTabValue;
  activeResource: Resource;
  canEdit: boolean;
  canEditCheckup: boolean;
  activeCanEdit: boolean;
  activeCanCreate: boolean;
  activeCanDelete: boolean;
  editTarget: TreatmentItem | "new" | null;
  pendingDelete: TreatmentItem | null;
  resources: TreatmentPlanResources;
  discardDialog: React.ReactNode;
  onNew: () => void;
  onClose: () => void;
  onSave: (data: TreatmentFormData) => Promise<boolean> | boolean;
  onDeleteRequest: () => void;
  onDirtyChange: (dirty: boolean) => void;
  onTabChange: (tab: string) => void;
  onEditTargetChange: (target: TreatmentItem | "new" | null) => void;
  onDeleteCancel: () => void;
  onDeleteConfirm: () => void;
}

export function TreatmentPlanMasterView({
  activeTab,
  activeResource,
  canEdit,
  canEditCheckup,
  activeCanEdit,
  activeCanCreate,
  activeCanDelete,
  editTarget,
  pendingDelete,
  resources,
  discardDialog,
  onNew,
  onClose,
  onSave,
  onDeleteRequest,
  onDirtyChange,
  onTabChange,
  onEditTargetChange,
  onDeleteCancel,
  onDeleteConfirm,
}: TreatmentPlanMasterViewProps) {
  return (
    <>
    <MasterTabPage
      title="診療項目マスタ"
      icon={<Stethoscope className={`${ICON.page} ${C.text}`} />}
      resource={activeResource}
      onNew={onNew}
      sidePanel={
        <TreatmentPlanSidePanelHost
          editTarget={editTarget}
          selectedItem={resources.selectedItem}
          parentCandidates={resources.parentCandidates}
          hasChildren={resources.hasChildren}
          canDelete={activeCanDelete}
          canCreate={activeCanCreate}
          canEdit={activeCanEdit}
          examinationType={resources.selectedExamination}
          showAnesthesia={activeTab === "procedure"}
          onClose={onClose}
          onSave={onSave}
          onDeleteRequest={onDeleteRequest}
          onDirtyChange={onDirtyChange}
        />
      }
      deleteDialogs={
        <TreatmentPlanDeleteDialog
          entityLabel={resources.tabConfigs[activeTab].entityLabel}
          pendingDelete={pendingDelete}
          onClose={onDeleteCancel}
          onConfirm={onDeleteConfirm}
        />
      }
    >
      <UnifiedTabs
        items={TREATMENT_PLAN_TABS}
        value={activeTab}
        onValueChange={onTabChange}
        className="flex flex-col gap-4"
      >
        {TREATMENT_PLAN_TABS.map((tab) => {
          const config = resources.tabConfigs[tab.value];
          return (
            <UnifiedTabsContent key={tab.value} value={tab.value} className="mt-4">
              <TreatmentPlanTabContent
                {...config}
                onEditTargetChange={onEditTargetChange}
                canEdit={tab.value === "checkup" ? canEditCheckup : canEdit}
              />
            </UnifiedTabsContent>
          );
        })}
      </UnifiedTabs>
    </MasterTabPage>
    {discardDialog}
    </>
  );
}
