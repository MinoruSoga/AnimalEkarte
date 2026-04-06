import React, { useMemo, useCallback } from "react";
import { CommandDialog, CommandInput, CommandList, CommandEmpty, CommandGroup, CommandItem } from "@/components/ui/command";
import { Check } from "lucide-react";
import { C, ICON } from "@/lib/design-tokens";
import { useGetStaffs, STAFF_ROLE_LABELS } from "@/features/master";
import type { StaffRoleValue } from "@/features/master";

interface StaffSelectionModalProps {
  open: boolean;
  selectedStaffName: string;
  onSelect: (staffId: string, staffName: string) => void;
  onOpenChange: (open: boolean) => void;
}

const ROLE_ORDER: StaffRoleValue[] = ["veterinarian", "nurse", "trimmer", "reception", "manager"];

export const StaffSelectionModal = React.memo(function StaffSelectionModal({
  open,
  selectedStaffName,
  onSelect,
  onOpenChange,
}: StaffSelectionModalProps) {
  const { data: staffs = [] } = useGetStaffs();

  const groupedStaffs = useMemo(() => {
    const active = staffs.filter((s) => s.isActive);
    const groups: Record<string, typeof active> = {};
    for (const s of active) {
      const role = s.staffRole;
      if (!groups[role]) groups[role] = [];
      groups[role].push(s);
    }
    return groups;
  }, [staffs]);

  const handleSelect = useCallback(
    (staffId: string, staffName: string) => {
      onSelect(staffId, staffName);
      onOpenChange(false);
    },
    [onSelect, onOpenChange],
  );

  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
      title="担当医を選択"
      description="担当するスタッフを検索・選択してください"
    >
      <CommandInput placeholder="スタッフ名で検索..." />

      <CommandList className="max-h-[500px]">
        <CommandEmpty className={`py-12 text-center text-sm ${C.text60}`}>
          該当するスタッフが見つかりません。
        </CommandEmpty>

        {ROLE_ORDER.map((role) => {
          const items = groupedStaffs[role];
          if (!items || items.length === 0) return null;

          return (
            <CommandGroup key={role} heading={STAFF_ROLE_LABELS[role]}>
              {items.map((staff) => {
                const isSelected = selectedStaffName === staff.name;
                return (
                  <CommandItem
                    key={staff.id}
                    value={`${staff.name} ${STAFF_ROLE_LABELS[staff.staffRole]}`}
                    onSelect={() => handleSelect(staff.id, staff.name)}
                    className={`cursor-pointer !py-2 ${isSelected ? C.bgPage : ""}`}
                  >
                    <div className="flex flex-1 items-center justify-between">
                      <span className={`font-medium ${C.text} text-sm`}>
                        {staff.name}
                      </span>
                      {isSelected ? (
                        <Check className={`${ICON.action} ${C.textPrimary}`} />
                      ) : null}
                    </div>
                  </CommandItem>
                );
              })}
            </CommandGroup>
          );
        })}
      </CommandList>
    </CommandDialog>
  );
});
