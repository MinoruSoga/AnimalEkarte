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
  petStatus?: string;
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
  petStatus,
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

  if (petStatus === "死亡") {
    return (
      <p className={`text-sm ${C.danger}`}>
        生死データに不整合があります（死亡ステータス・死亡日時未登録）。修復は管理者に依頼してください
      </p>
    );
  }

  if (canEdit === true) {
    return (
      <>
        <button
          type="button"
          onClick={() => setDialogOpen(true)}
          className={`text-sm ${C.danger} underline hover:no-underline transition-colors`}
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
            canEdit={canEdit}
            onRecorded={onRecorded}
          />
        ) : null}
      </>
    );
  }

  return null;
}
