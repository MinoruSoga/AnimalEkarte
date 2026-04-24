/**
 * OwnerFormPage - app層での feature 合成
 * owners feature と pets feature と line-reservation feature を app層でのみ合成し、
 * feature 間 import を排除する。
 */
import { useParams } from "react-router";

// features（app層なので複数 feature を import 可能）
import { OwnerForm } from "@/features/owners";
import { createPet, useCreatePet, useUpdatePet, useDeletePet } from "@/features/pets";
import { LinkedLineCustomers } from "@/features/line-reservation";
import { useAuth } from "@/hooks/use-auth";
import type { PetMutations } from "@/types/pet";

export function OwnerFormPage() {
  const { id: ownerId } = useParams();
  const { user } = useAuth();
  const clinicId = user?.mainClinicId ?? null;

  const { mutate: createPetMutate } = useCreatePet();
  const { mutate: updatePetMutate } = useUpdatePet();
  const { mutate: deletePetMutate } = useDeletePet();

  const petMutations: PetMutations = {
    createPetFn: createPet,
    createPetMutate: (req, { onSuccess, onError }) =>
      createPetMutate(req, { onSuccess, onError }),
    updatePetMutate: (args, { onSuccess, onError }) =>
      updatePetMutate(args, { onSuccess, onError }),
    deletePetMutate: (id, { onSuccess, onError }) =>
      deletePetMutate(id, { onSuccess, onError }),
  };

  // LINE連携セクション（編集モード=ownerIdがある時のみ意味がある）
  const lineSection = ownerId ? (
    <LinkedLineCustomers clinicId={clinicId} ownerId={Number(ownerId)} />
  ) : null;

  return <OwnerForm petMutations={petMutations} lineSection={lineSection} />;
}
