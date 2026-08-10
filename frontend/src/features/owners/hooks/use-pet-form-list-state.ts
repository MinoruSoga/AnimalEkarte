import { useCallback, useLayoutEffect, useRef, useState } from "react";
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
  permissions?: Readonly<PetMutationPermissions>;
}

interface PetMutationPermissions {
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
}

const DENIED_MUTATION_PERMISSIONS: Readonly<PetMutationPermissions> = {
  canCreate: false,
  canEdit: false,
  canDelete: false,
};

/** BUG-002: 専用 lifecycle mutation 成功後の外側一覧同期ペイロード */
type PetLifecycleChange = {
  petId: string;
  status: "死亡" | "生存";
  deceasedAt: string | null;
  deceasedReason?: string | null;
};

export function usePetFormListState({
  id,
  initialPets,
  petMutations,
  permissions = DENIED_MUTATION_PERMISSIONS,
}: UsePetFormListStateArgs) {
  const [pets, setPets] = useState<PetFormData[]>(initialPets);
  const [petModalOpen, setPetModalOpen] = useState(false);
  const [editingPet, setEditingPet] = useState<PetFormData | null>(null);
  const { canCreate, canEdit, canDelete } = permissions;
  const permissionsRef = useRef(permissions);
  const petsRef = useRef(pets);
  useLayoutEffect(() => {
    permissionsRef.current = { canCreate, canEdit, canDelete };
  }, [canCreate, canDelete, canEdit]);
  useLayoutEffect(() => {
    petsRef.current = pets;
  }, [pets]);

  // BUG-002: 死亡登録/解除は専用 API が完了済み。外側 pets と編集中 pet を ID 一致で不変更新する。
  // OwnerForm の editingPetRef.status ガードが stale にならないよう editingPet も同期する。
  const handlePetLifecycleChange = useCallback(
    ({ petId, status, deceasedAt, deceasedReason }: PetLifecycleChange) => {
      setPets((prev) => {
        // foreign/absent ID: 配列参照ごと no-op（map で新配列を作らない）
        if (!prev.some((pet) => pet.id === petId)) {
          return prev;
        }
        return prev.map((pet) =>
          pet.id === petId
            ? {
                ...pet,
                status,
                deceasedAt,
                deceasedReason: status === "生存" ? null : (deceasedReason ?? pet.deceasedReason),
              }
            : pet,
        );
      });
      setEditingPet((prev) =>
        prev?.id === petId
          ? {
              ...prev,
              status,
              deceasedAt,
              deceasedReason: status === "生存" ? null : (deceasedReason ?? prev.deceasedReason),
            }
          : prev,
      );
    },
    [],
  );

  const handleAddPet = () => {
    if (permissionsRef.current.canCreate !== true) return;
    setEditingPet(null);
    setPetModalOpen(true);
  };

  const handleEditPet = (pet: PetFormData) => {
    if (permissionsRef.current.canEdit !== true) return;
    setEditingPet(pet);
    setPetModalOpen(true);
  };

  const handleDeletePet = (petId: string) => {
    const target = petsRef.current.find(p => p.id === petId);
    if (
      target?.status === "死亡" ||
      permissionsRef.current.canDelete !== true
    ) {
      return;
    }
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
        if (permissionsRef.current.canEdit !== true) return;
        setPets(prev =>
          prev.map(p =>
            p.id === editingPet.id
              ? { ...petData, id: editingPet.id, isPending: true }
              : p
          )
        );
        return;
      }

      const currentPet = petsRef.current.find((pet) => pet.id === editingPet.id) ?? editingPet;
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
        dangerReason: petData.dangerReason,
        originalDangerReason: currentPet.dangerReason,
        // status は渡さない(BUG-415): transformUpdatePetRequest は status を無視する。
        insuranceId: petData.insuranceId,
        remarks: petData.remarks,
      });

      if (
        permissionsRef.current.canEdit !== true ||
        currentPet.status === "死亡"
      ) {
        return;
      }
      // BUG-415: 生死ステータスの変更は監査付きの死亡登録/取消エンドポイント
      // (PetCareSection → PetDeceasedRecordButton → useRecordPetDeath/useRevokePetDeath) に
      // 一本化済み。それらは status 変更を自身のミューテーションで即時完結させるため、
      // 汎用 Save (updatePetMutate) に「死亡→生存の遷移を検知して revokePetDeathMutate を
      // 補完発火する」旧ロジック(isPetRevival)は不要かつ二重発火の原因になるため削除した。
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
        if (permissionsRef.current.canCreate !== true) return;
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
        dangerReason: petData.dangerReason,
        status: PET_STATUS_REVERSE_MAP[petData.status],
        insuranceId: petData.insuranceId,
        remarks: petData.remarks,
      });

      if (permissionsRef.current.canCreate !== true) return;
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
            dangerReason: newPetData.dangerReason || "",
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
    handlePetLifecycleChange,
  };
}
