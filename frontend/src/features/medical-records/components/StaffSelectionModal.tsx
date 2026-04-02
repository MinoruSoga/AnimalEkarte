import React, { useMemo } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { useGetStaffs } from "@/features/master";
import { C } from "@/lib/design-tokens";

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
  const { data: staffs = [], isLoading } = useGetStaffs();

  const activeStaffs = useMemo(
    () => staffs.filter((s) => s.isActive),
    [staffs]
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle className={`text-base font-bold ${C.text}`}>
            医師を選択
          </DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-2 max-h-[400px] overflow-y-auto py-2">
          {isLoading ? (
            <div className={`text-center text-sm ${C.text60} py-4`}>
              読み込み中...
            </div>
          ) : activeStaffs.length === 0 ? (
            <div className={`text-center text-sm ${C.text60} py-4`}>
              医師情報がありません
            </div>
          ) : (
            activeStaffs.map((staff) => (
              <Button
                key={staff.id}
                variant="ghost"
                className={`justify-start h-10 text-sm font-normal ${
                  selectedStaffName === staff.name
                    ? `${C.bgMedicalBlue} text-white ${C.hoverBgMedicalBlue90} hover:text-white`
                    : `${C.text} ${C.hoverBgPage}`
                }`}
                onClick={() => {
                  onSelect(staff.id, staff.name);
                  onOpenChange(false);
                }}
              >
                {staff.name}
              </Button>
            ))
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
});
