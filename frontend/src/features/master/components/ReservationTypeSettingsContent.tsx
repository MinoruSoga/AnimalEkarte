import Plus from "lucide-react/dist/esm/icons/plus";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { C, ICON } from "@/lib/design-tokens";
import type { ReservationType } from "../api/reservation-types";
import type { ReservationTypeGroup } from "../api/reservation-type-groups";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import { ReservationTypeGroupedTable } from "./ReservationTypeGroupedTable";

interface ReservationTypeSettingsContentProps {
  groups: ReservationTypeGroup[];
  categories: ReservationType[];
  activeFilters: ActiveFilter[];
  onFilterChange: (filters: ActiveFilter[]) => void;
  searchTerm: string;
  onSearchChange: (term: string) => void;
  count: number;
  canCreate: boolean;
  canEdit: boolean;
  onCategoryEdit: (category: ReservationType) => void;
  onGroupEdit: (group: ReservationTypeGroup) => void;
  onCategoryAddInGroup: (groupId: string | undefined) => void;
  onGroupAdd: () => void;
}

export function ReservationTypeSettingsContent({
  groups,
  categories,
  activeFilters,
  onFilterChange,
  searchTerm,
  onSearchChange,
  count,
  canCreate,
  canEdit,
  onCategoryEdit,
  onGroupEdit,
  onCategoryAddInGroup,
  onGroupAdd,
}: ReservationTypeSettingsContentProps) {
  return (
    <div className="flex flex-col gap-4">
      <PropertyFilter
        properties={[MASTER_STATUS_FILTER]}
        activeFilters={activeFilters}
        onFilterChange={onFilterChange}
        searchTerm={searchTerm}
        onSearchChange={onSearchChange}
        searchPlaceholder="予約区分名で検索..."
        count={count}
      />
      <ReservationTypeGroupedTable
        groups={groups}
        categories={categories}
        onCategoryEdit={onCategoryEdit}
        onGroupEdit={onGroupEdit}
        onCategoryAddInGroup={onCategoryAddInGroup}
        canEdit={canEdit}
      />
      {canCreate ? (
        <button
          type="button"
          onClick={onGroupAdd}
          className={`flex min-h-11 min-w-11 items-center gap-1.5 text-sm ${C.text45} ${C.hoverText} ${C.hoverBgLight}
            px-2 py-1.5 rounded-xxs transition-colors w-fit`}
        >
          <Plus className={ICON.action} />
          グループを追加
        </button>
      ) : null}
    </div>
  );
}
