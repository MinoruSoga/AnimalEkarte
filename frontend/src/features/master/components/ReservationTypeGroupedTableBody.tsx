import { Fragment } from "react";

import type { ReservationType } from "../api/reservation-types";
import type { ReservationTypeGroup } from "../api/reservation-type-groups";
import {
  ReservationTypeEmptyGroupRow,
  ReservationTypeGroupHeader,
  ReservationTypeRow,
  ReservationTypeUncategorizedHeader,
} from "./ReservationTypeGroupedTableRows";
import {
  UNCATEGORIZED_RESERVATION_TYPE_GROUP_ID,
  type ReservationTypesByGroup,
} from "../lib/reservation-type-grouped-table-model";

interface ReservationTypeGroupedTableBodyProps {
  groups: ReservationTypeGroup[];
  groupedCategories: ReservationTypesByGroup;
  collapsed: Set<string>;
  canEdit: boolean;
  onToggleCollapse: (id: string) => void;
  onCategoryEdit: (category: ReservationType) => void;
  onGroupEdit: (group: ReservationTypeGroup) => void;
  onCategoryAddInGroup: (groupId: string | undefined) => void;
}

export function ReservationTypeGroupedTableBody({
  groups,
  groupedCategories,
  collapsed,
  canEdit,
  onToggleCollapse,
  onCategoryEdit,
  onGroupEdit,
  onCategoryAddInGroup,
}: ReservationTypeGroupedTableBodyProps) {
  const uncategorizedCategories = groupedCategories.uncategorized;
  const uncategorizedCollapsed = collapsed.has(UNCATEGORIZED_RESERVATION_TYPE_GROUP_ID);

  return (
    <tbody>
      {groups.map((group) => {
        const groupCategories = groupedCategories.map.get(group.id) ?? [];
        const isCollapsed = collapsed.has(group.id);

        return (
          <Fragment key={group.id}>
            <ReservationTypeGroupHeader
              group={group}
              count={groupCategories.length}
              isCollapsed={isCollapsed}
              canEdit={canEdit}
              onToggle={() => onToggleCollapse(group.id)}
              onGroupEdit={() => onGroupEdit(group)}
              onCategoryAdd={() => onCategoryAddInGroup(group.id)}
            />
            {!isCollapsed ? (
              <ReservationTypeGroupCategories
                categories={groupCategories}
                canEdit={canEdit}
                onCategoryEdit={onCategoryEdit}
              />
            ) : null}
          </Fragment>
        );
      })}

      {uncategorizedCategories.length > 0 || groups.length === 0 ? (
        <Fragment key={UNCATEGORIZED_RESERVATION_TYPE_GROUP_ID}>
          <ReservationTypeUncategorizedHeader
            count={uncategorizedCategories.length}
            isCollapsed={uncategorizedCollapsed}
            canEdit={canEdit}
            onToggle={() => onToggleCollapse(UNCATEGORIZED_RESERVATION_TYPE_GROUP_ID)}
            onCategoryAdd={() => onCategoryAddInGroup(undefined)}
          />
          {!uncategorizedCollapsed ? (
            <ReservationTypeRows
              categories={uncategorizedCategories}
              canEdit={canEdit}
              onCategoryEdit={onCategoryEdit}
            />
          ) : null}
        </Fragment>
      ) : null}
    </tbody>
  );
}

interface ReservationTypeRowsProps {
  categories: ReservationType[];
  canEdit: boolean;
  onCategoryEdit: (category: ReservationType) => void;
}

function ReservationTypeGroupCategories({
  categories,
  canEdit,
  onCategoryEdit,
}: ReservationTypeRowsProps) {
  if (categories.length === 0) {
    return <ReservationTypeEmptyGroupRow />;
  }

  return (
    <ReservationTypeRows
      categories={categories}
      canEdit={canEdit}
      onCategoryEdit={onCategoryEdit}
    />
  );
}

function ReservationTypeRows({ categories, canEdit, onCategoryEdit }: ReservationTypeRowsProps) {
  return (
    <>
      {categories.map((category) => (
        <ReservationTypeRow
          key={category.id}
          category={category}
          canEdit={canEdit}
          onEdit={() => onCategoryEdit(category)}
        />
      ))}
    </>
  );
}
