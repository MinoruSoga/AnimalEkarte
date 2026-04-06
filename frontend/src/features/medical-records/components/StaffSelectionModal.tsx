import React, { useMemo } from "react";
import { MasterSelectModal } from "@/components/shared/MasterSelectModal";
import { useGetStaffs } from "@/features/master";

interface StaffSelectionModalProps {
  open: boolean;
  selectedStaffName: string;
  onSelect: (staffId: string, staffName: string) => void;
  onOpenChange: (open: boolean) => void;
}

export const StaffSelectionModal = React.memo(function StaffSelectionModal({
  open,
  selectedStaffName,
  onSelect,
  onOpenChange,
}: StaffSelectionModalProps) {
  const { data: staffs = [] } = useGetStaffs();

  const activeStaffs = useMemo(
    () =>
      staffs
        .filter((s) => s.isActive)
        .map((s) => ({ id: s.id, name: s.name })),
    [staffs],
  );

  return (
    <MasterSelectModal
      open={open}
      onOpenChange={onOpenChange}
      title="担当医を選択"
      items={activeStaffs}
      selectedValue={selectedStaffName}
      matchBy="name"
      searchPlaceholder="スタッフ名で検索..."
      onSelect={(item) => onSelect(String(item.id), item.name)}
    />
  );
});
