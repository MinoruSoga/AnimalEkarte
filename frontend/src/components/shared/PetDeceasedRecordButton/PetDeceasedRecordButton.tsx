import { useState } from "react";
import { C } from "@/lib/design-tokens";
import { calcAgeAt } from "@/lib/calc-age";
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
  /** BUG-407: 死亡記録成功時に外側フォームのローカル状態へ同期するためのコールバック */
  onRecorded?: (result: { deceasedAt: string; deceasedReason?: string }) => void;
  /** BUG-407: 死亡記録解除成功時に外側フォームのローカル状態へ同期するためのコールバック */
  onRevoked?: () => void;
}

export function PetDeceasedRecordButton({
  petId,
  petName,
  petBreed,
  petGender,
  birthDate,
  deceasedAt,
  canEdit = false,
  onRecorded,
  onRevoked,
}: PetDeceasedRecordButtonProps) {
  const [dialogOpen, setDialogOpen] = useState(false);

  const petAge = birthDate ? `${calcAgeAt(new Date(), birthDate)}歳` : undefined;

  if (deceasedAt) {
    return (
      <PetDeceasedBanner
        deceasedAt={deceasedAt}
        birthDate={birthDate}
        petId={petId}
        canEdit={canEdit}
        onRevoked={onRevoked}
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
          onRecorded={onRecorded}
        />
      ) : null}
    </>
  );
}
