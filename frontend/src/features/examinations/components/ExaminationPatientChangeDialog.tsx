import { useCallback, useState } from "react";

import { PatientSelectionTable } from "@/components/shared/ReservationFormModal/PatientSelectionTable";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import type { Pet } from "@/types";

interface ExaminationPatientChangeDialogProps {
  selectedPet: Pet | undefined;
  onSelect: (pet: Pet) => void;
}

export function ExaminationPatientChangeDialog({
  selectedPet,
  onSelect,
}: ExaminationPatientChangeDialogProps) {
  const [open, setOpen] = useState(false);

  const handleSelect = useCallback(
    (pet: Pet) => {
      if (pet.status !== "生存") return;
      onSelect(pet);
      setOpen(false);
    },
    [onSelect],
  );

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button type="button" variant="outline" className="h-11 min-w-11 text-sm">
          患者を変更
        </Button>
      </DialogTrigger>
      <DialogContent className="flex h-[80vh] max-w-[calc(100%-2rem)] flex-col sm:max-w-6xl">
        <DialogHeader>
          <DialogTitle>検査対象の患者を変更</DialogTitle>
          <DialogDescription>
            初回確定前の検査だけ変更できます。死亡または状態不明の患者は選択できません。
          </DialogDescription>
        </DialogHeader>
        <div className="min-h-0 flex-1">
          <PatientSelectionTable
            selectedPets={selectedPet ? [selectedPet] : []}
            onSelect={handleSelect}
            includeDeceased
          />
        </div>
      </DialogContent>
    </Dialog>
  );
}
