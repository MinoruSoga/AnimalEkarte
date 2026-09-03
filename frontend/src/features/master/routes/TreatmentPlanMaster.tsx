import { useState, useCallback, useLayoutEffect, useRef } from "react";
import { useSearchParams } from "react-router";
import { toast } from "sonner";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { handleApiError } from "@/lib/handle-api-error";
import { usePermission } from "@/hooks/use-permission";
import { toTreatmentPlanTabValue } from "./treatment-plan-master-model";
import type { TreatmentItem } from "@/lib/transforms/treatment";
import { ResourceCheckups, ResourceMasterMedical } from "@/types/generated/models";
import { useTreatmentPlanMasterResources } from "../hooks/use-treatment-plan-master-resources";
import { useTreatmentPlanMasterSaves } from "../hooks/use-treatment-plan-master-saves";
import { TreatmentPlanMasterView } from "./treatment-plan-master-view";

export function TreatmentPlanMaster() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterMedical);
  const {
    canCreate: canCreateCheckup,
    canEdit: canEditCheckup,
    canDelete: canDeleteCheckup,
  } = usePermission(ResourceCheckups);
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = toTreatmentPlanTabValue(searchParams.get("tab"));
  const activeResource = activeTab === "checkup" ? ResourceCheckups : ResourceMasterMedical;
  const activeCanEdit = activeTab === "checkup" ? canEditCheckup : canEdit;
  const activeCanCreate = activeTab === "checkup" ? canCreateCheckup : canCreate;
  const activeCanDelete = activeTab === "checkup" ? canDeleteCheckup : canDelete;
  const permissionsRef = useRef({ canDelete: activeCanDelete === true });
  useLayoutEffect(() => {
    permissionsRef.current = { canDelete: activeCanDelete === true };
  }, [activeCanDelete]);
  const [editTarget, setEditTarget] = useState<TreatmentItem | "new" | null>(null);
  const [pendingDelete, setPendingDelete] = useState<TreatmentItem | null>(null);

  const dirty = useSidePeekDirty();
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  const handleTabChange = useCallback((tab: string) => {
    dirty.runWithDiscardCheck(() => {
      setSearchParams({ tab });
      setEditTarget(null);
      setPendingDelete(null);
    });
  }, [setSearchParams, dirty]);

  const handleNew = useCallback(() => {
    dirty.runWithDiscardCheck(() => {
      setEditTarget("new");
    });
  }, [dirty]);

  const resources = useTreatmentPlanMasterResources({
    canEdit,
    canEditCheckup,
    activeTab,
    editTarget,
  });

  const startSaveTransition = useCallback((cb: () => void) => {
    cb();
  }, []);

  const { handleSave } = useTreatmentPlanMasterSaves({
    editTarget,
    setEditTarget,
    startSaveTransition,
    canCreate,
    canEdit,
    canCreateCheckup,
    canEditCheckup,
    activeTab,
    resources,
  });

  const handleClose = useCallback(() => {
    dirty.runWithDiscardCheck(() => {
      setEditTarget(null);
    });
  }, [dirty]);

  const setEditTargetGuarded = useCallback((target: TreatmentItem | "new" | null) => {
    dirty.runWithDiscardCheck(() => {
      setEditTarget(target);
    });
  }, [dirty]);

  const handleDeleteRequest = useCallback(() => {
    setPendingDelete(resources.selectedItem);
  }, [resources.selectedItem]);

  const handleDeleteCancel = useCallback(() => {
    setPendingDelete(null);
  }, []);

  const handleDeleteConfirm = useCallback(() => {
    if (!pendingDelete) return;
    const config = resources.tabConfigs[activeTab];
    if (permissionsRef.current.canDelete !== true) return;
    resources.deleteMutationByTab[activeTab].mutate(pendingDelete.id, {
      onSuccess: () => {
        setPendingDelete(null);
        setEditTarget(null);
        toast.success(`${config.entityLabel}を削除しました`);
      },
      onError: (error) => handleApiError(error, `${config.entityLabel}の削除`),
    });
  }, [activeTab, pendingDelete, resources]);

  return (
    <TreatmentPlanMasterView
      activeTab={activeTab}
      activeResource={activeResource}
      canEdit={canEdit}
      canEditCheckup={canEditCheckup}
      activeCanEdit={activeCanEdit}
      activeCanCreate={activeCanCreate}
      activeCanDelete={activeCanDelete}
      editTarget={editTarget}
      pendingDelete={pendingDelete}
      resources={resources}
      discardDialog={dirty.discardDialog}
      onNew={handleNew}
      onClose={handleClose}
      onSave={handleSave}
      onDeleteRequest={handleDeleteRequest}
      onDirtyChange={handleDirtyChange}
      onTabChange={handleTabChange}
      onEditTargetChange={setEditTargetGuarded}
      onDeleteCancel={handleDeleteCancel}
      onDeleteConfirm={handleDeleteConfirm}
    />
  );
}
