// Internal
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

// Relative
import { VitalsTab } from "./VitalsTab/VitalsTab";

interface VitalsModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  medicalRecordId: string;
  /** P2-15: 拠点横断で開いたカルテの子リソース操作用。レコード自身の clinicId */
  recordClinicId?: string;
  isPetDeceased?: boolean;
}

export function VitalsModal({
  open,
  onOpenChange,
  medicalRecordId,
  recordClinicId,
  isPetDeceased = false,
}: VitalsModalProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>バイタル記録</DialogTitle>
          <DialogDescription>この診療記録に紐づくバイタルを確認・編集します。</DialogDescription>
        </DialogHeader>
        <VitalsTab
          medicalRecordId={medicalRecordId}
          recordClinicId={recordClinicId}
          isPetDeceased={isPetDeceased}
        />
      </DialogContent>
    </Dialog>
  );
}
