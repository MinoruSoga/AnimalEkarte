import { memo, useCallback, useMemo } from "react";
import type { PermissionGroup } from "@/types/generated/models";
import { PALETTE } from "@/lib/design-tokens";

interface PermissionGroupSelectorProps {
  availableGroups: PermissionGroup[];
  selectedGroupIds: number[];
  onChange: (groupIds: number[]) => void;
}

export const PermissionGroupSelector = memo(function PermissionGroupSelector({
  availableGroups,
  selectedGroupIds,
  onChange,
}: PermissionGroupSelectorProps) {
  const handleToggle = useCallback(
    (groupId: number) => {
      const isSelected = selectedGroupIds.includes(groupId);
      onChange(
        isSelected
          ? selectedGroupIds.filter((id) => id !== groupId)
          : [...selectedGroupIds, groupId],
      );
    },
    [selectedGroupIds, onChange],
  );

  const groupButtons = useMemo(
    () =>
      availableGroups.map((group) => {
        const isSelected = selectedGroupIds.includes(group.id);
        return (
          <button
            key={group.id}
            type="button"
            onClick={() => handleToggle(group.id)}
            className="px-3 py-1.5 rounded-full text-sm font-medium border transition-colors"
            style={
              isSelected
                ? { backgroundColor: group.color ?? PALETTE.defaultGray, borderColor: group.color ?? PALETTE.defaultGray, color: PALETTE.white }
                : { backgroundColor: "transparent", borderColor: PALETTE.borderUnselected, color: "inherit" }
            }
          >
            {group.name}
          </button>
        );
      }),
    [availableGroups, selectedGroupIds, handleToggle],
  );

  return availableGroups.length === 0 ? (
    <p className="text-sm text-muted-foreground">
      権限グループが作成されていません。
      <br />
      先に「権限グループ管理」でグループを作成してください。
    </p>
  ) : (
    <div className="flex flex-wrap gap-2">{groupButtons}</div>
  );
});
