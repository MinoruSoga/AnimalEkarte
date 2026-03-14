/**
 * OwnerFormPage - app層での feature 合成
 * owners feature と pets feature を app層でのみ合成し、
 * feature 間 import を排除する。
 */

// features（app層なので両 feature を import 可能）
import { OwnerForm } from "@/features/owners/routes/OwnerForm";
import { createPet, useCreatePet } from "@/features/pets/api/create-pet";
import { useUpdatePet } from "@/features/pets/api/update-pet";
import { useDeletePet } from "@/features/pets/api/delete-pet";
import type { PetMutations } from "@/types/pet";

export function OwnerFormPage() {
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

  return <OwnerForm petMutations={petMutations} />;
}
