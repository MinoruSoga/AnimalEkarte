import type { ReservationType } from "../api/reservation-types";
import type { ReservationTypeGroup } from "../api/reservation-type-groups";
import { GroupSidePanel, type GroupFormData } from "./ReservationTypeGroupSidePanel";
import { CategorySidePanel, type CategoryFormData } from "./ReservationTypeSidePanel";

type GroupOption = { id: string; name: string; color: string };

interface ReservationTypeSettingsSidePanelsProps {
  groupEditTarget: ReservationTypeGroup | "new" | null;
  groupPanelItem: ReservationTypeGroup | null;
  categoryEditTarget: ReservationType | "new" | null;
  categoryPanelItem: ReservationType | null;
  activeGroups: GroupOption[];
  categoryDefaultGroupId?: string;
  canDelete: boolean;
  canEdit: boolean;
  onGroupClose: () => void;
  onGroupSave: (data: GroupFormData) => Promise<boolean> | boolean;
  onGroupDeleteRequest: (item: ReservationTypeGroup) => void;
  onCategoryClose: () => void;
  onCategorySave: (data: CategoryFormData) => Promise<boolean> | boolean;
  onCategoryDeleteRequest: (item: ReservationType) => void;
  onDirtyChange: (dirty: boolean) => void;
}

export function ReservationTypeSettingsSidePanels({
  groupEditTarget,
  groupPanelItem,
  categoryEditTarget,
  categoryPanelItem,
  activeGroups,
  categoryDefaultGroupId,
  canDelete,
  canEdit,
  onGroupClose,
  onGroupSave,
  onGroupDeleteRequest,
  onCategoryClose,
  onCategorySave,
  onCategoryDeleteRequest,
  onDirtyChange,
}: ReservationTypeSettingsSidePanelsProps) {
  return (
    <>
      {groupEditTarget !== null ? (
        <GroupSidePanel
          key={groupPanelItem ? String(groupPanelItem.id) : "new-group"}
          item={groupPanelItem}
          onClose={onGroupClose}
          onSave={onGroupSave}
          onDeleteRequest={canDelete ? onGroupDeleteRequest : undefined}
          readOnly={!canEdit}
          onDirtyChange={onDirtyChange}
        />
      ) : null}
      {categoryEditTarget !== null ? (
        <CategorySidePanel
          key={categoryPanelItem ? String(categoryPanelItem.id) : "new-category"}
          item={categoryPanelItem}
          onClose={onCategoryClose}
          onSave={onCategorySave}
          onDeleteRequest={canDelete ? onCategoryDeleteRequest : undefined}
          readOnly={!canEdit}
          groups={activeGroups}
          defaultGroupId={categoryDefaultGroupId}
          onDirtyChange={onDirtyChange}
        />
      ) : null}
    </>
  );
}
