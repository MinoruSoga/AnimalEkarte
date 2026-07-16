import { useState } from "react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { transformCreatePetRequest, transformUpdatePetRequest, PET_STATUS_REVERSE_MAP } from "@/lib/transforms/pet";
import type { Pet } from "@/types";
import type { PetMutations } from "@/types/pet";
import type { PetFormData } from "../types";

interface UsePetFormListStateArgs {
  id?: string;
  initialPets: PetFormData[];
  petMutations?: PetMutations;
}

export function usePetFormListState({ id, initialPets, petMutations }: UsePetFormListStateArgs) {
  const [pets, setPets] = useState<PetFormData[]>(initialPets);
  const [petModalOpen, setPetModalOpen] = useState(false);
  const [editingPet, setEditingPet] = useState<PetFormData | null>(null);

  const handleAddPet = () => {
    setEditingPet(null);
    setPetModalOpen(true);
  };

  const handleEditPet = (pet: PetFormData) => {
    setEditingPet(pet);
    setPetModalOpen(true);
  };

  const handleDeletePet = (petId: string) => {
    const target = pets.find(p => p.id === petId);
    if (target?.isPending) {
      setPets(prev => prev.filter(p => p.id !== petId));
      toast.success("ペットを削除しました");
      return;
    }

    petMutations?.deletePetMutate(petId, {
      onSuccess: () => {
        setPets(prev => prev.filter((pet) => pet.id !== petId));
        toast.success("ペットを削除しました");
      },
      onError: (error: unknown) => {
        handleApiError(error, "削除");
      },
    });
  };

  const handleSavePet = (petData: PetFormData) => {
    if (editingPet) {
      if (editingPet.isPending) {
        setPets(prev =>
          prev.map(p =>
            p.id === editingPet.id
              ? { ...petData, id: editingPet.id, isPending: true }
              : p
          )
        );
        return;
      }

      const updateRequest = transformUpdatePetRequest({
        petNumber: petData.petNumber,
        name: petData.petName,
        petNameKana: petData.petNameKana,
        animalSpeciesId: petData.animalSpeciesId,
        gender: petData.gender,
        birthDate: petData.birthDate,
        breed: petData.breed,
        color: petData.color,
        bloodType: petData.bloodType,
        microchipNumber: petData.microchipNumber,
        weight: petData.weight,
        food: petData.food,
        environment: petData.environment,
        neuteredDate: petData.neuteredDate,
        acquisitionType: petData.acquisitionType,
        dangerLevel: petData.dangerLevel,
        status: PET_STATUS_REVERSE_MAP[petData.status],
        insuranceId: petData.insuranceId,
        remarks: petData.remarks,
      });

      // 死亡→生存 の実際の遷移時のみ死亡記録を解除する。
      // editingPet.status は編集開始時にロードされた初期値（handleEditPet 以外で書き換わらない）。
      // 送信後の status が生存というだけでは発火させない（大半のペット編集は status を触らず生存のまま）。
      const isPetRevival = editingPet.status === "死亡" && petData.status === "生存";

      petMutations?.updatePetMutate(
        { id: editingPet.id, req: updateRequest },
        {
          onSuccess: () => {
            setPets(prev =>
              prev.map((pet) =>
                pet.id === editingPet.id ? { ...pet, ...petData } : pet
              )
            );
            toast.success("ペット情報を更新しました");
            // status PATCH が永続化された後にのみ死亡記録を解除する。
            if (isPetRevival) {
              petMutations?.revokePetDeathMutate(editingPet.id);
            }
          },
          onError: (error: unknown) => {
            handleApiError(error, "更新");
          },
        }
      );
    } else {
      if (!petData.animalSpeciesId) {
        return;
      }

      if (!id) {
        const tempId = `temp-${Date.now()}`;
        setPets(prev => [...prev, { ...petData, id: tempId, isPending: true }]);
        return;
      }

      const createRequest = transformCreatePetRequest({
        ownerId: id,
        name: petData.petName || "",
        animalSpeciesId: petData.animalSpeciesId,
        petNumber: petData.petNumber,
        petNameKana: petData.petNameKana,
        breed: petData.breed,
        color: petData.color,
        bloodType: petData.bloodType,
        microchipNumber: petData.microchipNumber,
        gender: petData.gender,
        birthDate: petData.birthDate,
        weight: petData.weight,
        food: petData.food,
        environment: petData.environment,
        neuteredDate: petData.neuteredDate,
        acquisitionType: petData.acquisitionType,
        dangerLevel: petData.dangerLevel,
        status: PET_STATUS_REVERSE_MAP[petData.status],
        insuranceId: petData.insuranceId,
        remarks: petData.remarks,
      });

      petMutations?.createPetMutate(createRequest, {
        onSuccess: (newPetData: Pet) => {
          const newPet: PetFormData = {
            id: newPetData.id,
            petNumber: newPetData.petNumber || "",
            petName: newPetData.name,
            petNameKana: newPetData.petNameKana || "",
            status: newPetData.status || "生存",
            species: newPetData.species,
            animalSpeciesId: newPetData.animalSpeciesId,
            gender: newPetData.gender || "",
            birthDate: newPetData.birthDate || "",
            color: newPetData.color || "",
            bloodType: newPetData.bloodType || "",
            microchipNumber: newPetData.microchipNumber || "",
            weight: newPetData.weight || "",
            food: newPetData.food || "",
            environment: newPetData.environment || "",
            neuteredDate: newPetData.neuteredDate || "",
            acquisitionType: (newPetData.acquisitionType as PetFormData["acquisitionType"]) || "購入",
            dangerLevel: (newPetData.dangerLevel as PetFormData["dangerLevel"]) || "低",
            remarks: newPetData.remarks || "",
            breed: newPetData.breed,
            insuranceId: newPetData.insuranceId,
            insuranceName: undefined,
            insuranceDetails: newPetData.insuranceDetails,
            deceasedAt: newPetData.deceasedAt,
          };
          setPets(prev => [...prev, newPet]);
          toast.success("ペットを追加しました");
        },
        onError: (error: unknown) => {
          handleApiError(error, "追加");
        },
      });
    }
  };

  return {
    pets,
    setPets,
    petModalOpen,
    setPetModalOpen,
    editingPet,
    handleAddPet,
    handleEditPet,
    handleDeletePet,
    handleSavePet,
  };
}
