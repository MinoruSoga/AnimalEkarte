import { useMemo, useCallback, memo } from "react";
import {
  CommandDialog,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
} from "@/components/ui/command";
import { Check } from "lucide-react";
import { C, ICON } from "@/lib/design-tokens";
import { useGetStaffs } from "@/hooks/use-staffs";
import { normalizedIncludes } from "@/lib/normalize-kana";

interface StaffSelectionModalProps {
  open: boolean;
  selectedStaffName: string;
  onSelect: (staffId: string, staffName: string) => void;
  onOpenChange: (open: boolean) => void;
}

export const StaffSelectionModal = memo(function StaffSelectionModal({
  open,
  selectedStaffName,
  onSelect,
  onOpenChange,
}: StaffSelectionModalProps) {
  const { data: staffs = [] } = useGetStaffs();

  const groupedStaffs = useMemo(() => {
    const active = staffs.filter((s) => s.isActive);
    const groups: Record<string, typeof active> = {};
    const occupationOrder: string[] = [];
    for (const s of active) {
      const key = s.occupationName ?? "未設定";
      if (!groups[key]) {
        groups[key] = [];
        occupationOrder.push(key);
      }
      groups[key].push(s);
    }
    return { groups, occupationOrder };
  }, [staffs]);

  const commandFilter = useCallback(
    (value: string, search: string) => (normalizedIncludes(value, search) ? 1 : 0),
    [],
  );

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
      filter={commandFilter}
      title="担当医を選択"
      description="担当するスタッフを検索・選択してください"
    >
      <CommandInput placeholder="スタッフ名で検索..." />

      <CommandList className="max-h-[500px]">
        <CommandEmpty className={`py-12 text-center text-sm ${C.text60}`}>
          該当するスタッフが見つかりません。
        </CommandEmpty>

        {groupedStaffs.occupationOrder.map((occupationKey) => {
          const items = groupedStaffs.groups[occupationKey];
          if (!items || items.length === 0) return null;

          return (
            <CommandGroup key={occupationKey} heading={occupationKey}>
              {items.map((staff) => {
                const isSelected = selectedStaffName === staff.name;
                return (
                  <CommandItem
                    key={staff.id}
                    value={`${staff.name} ${occupationKey}`}
                    onSelect={() => handleSelect(staff.id, staff.name)}
                    className={`cursor-pointer !py-2 ${isSelected ? C.bgPage : ""}`}
                  >
                    <div className="flex flex-1 items-center justify-between">
                      <span className={`font-medium ${C.text} text-sm`}>{staff.name}</span>
                      {isSelected ? <Check className={`${ICON.action} ${C.text}`} /> : null}
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
