import { useState } from "react";
import { C } from "@/lib/design-tokens";
import { PetDeceasedDialog } from "./PetDeceasedDialog";
import { PetDeceasedBanner } from "./PetDeceasedBanner";

interface PetDeceasedRecordButtonProps {
  petId: string;
  petName: string;
  petBreed?: string;
  petGender?: string;
  birthDate?: string;
  deceasedAt?: string | null;
  canEdit?: boolean;
}

export function calcAge(birthDate: string): string {
  const birth = new Date(birthDate);
  const now = new Date();
  let age = now.getFullYear() - birth.getFullYear();
  const hasBirthdayPassed =
    now.getMonth() > birth.getMonth() ||
    (now.getMonth() === birth.getMonth() && now.getDate() >= birth.getDate());
  if (!hasBirthdayPassed) {
    age -= 1;
  }
  return `${Math.max(0, age)}歳`;
}

export function PetDeceasedRecordButton({
  petId,
  petName,
  petBreed,
  petGender,
  birthDate,
  deceasedAt,
  canEdit = false,
}: PetDeceasedRecordButtonProps) {
  const [dialogOpen, setDialogOpen] = useState(false);

  const petAge = birthDate ? calcAge(birthDate) : undefined;

  if (deceasedAt) {
    return (
      <PetDeceasedBanner
        deceasedAt={deceasedAt}
        birthDate={birthDate}
        petId={petId}
        canEdit={canEdit}
      />
    );
  }

  return (
    <>
      <button
        type="button"
        onClick={() => setDialogOpen(true)}
        className={`text-sm ${C.text50} underline hover:no-underline ${C.hoverText} transition-colors`}
      >
        死亡を記録
      </button>

      {dialogOpen ? (
        <PetDeceasedDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          petId={petId}
          petName={petName}
          petBreed={petBreed}
          petGender={petGender}
          petAge={petAge}
        />
      ) : null}
    </>
  );
}
